package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	"backend/external"
	"strconv"
	"time"

	"fmt"
)

// forwardAmountHeuristic - see exec for description
type forwardAmountHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookForwardTime      time.Duration
}

// newForwardAmountHeuristic constructs an forwardAmountHeuristic. hoursToLookForward in hours.
func newForwardAmountHeuristic(hoursToLookForward uint32) *forwardAmountHeuristic {
	lForwardTime := time.Duration(hoursToLookForward) * time.Hour
	return &forwardAmountHeuristic{
		heuristicType:        "forward_amount",
		lookForwardTime:      lForwardTime,
		parameterDescription: lForwardTime.String(),
	}
}

func (h *forwardAmountHeuristic) getType() string {
	return h.heuristicType
}

func (h *forwardAmountHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *forwardAmountHeuristic) hasParameter() bool {
	return true
}

func (h *forwardAmountHeuristic) setParameter(p string) error {
	hoursToLookForward, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.parameterDescription = strconv.FormatUint(hoursToLookForward, 10)
	return nil
}

func (h *forwardAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h *forwardAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Forward Amount",
		Type:     h.heuristicType,
		Category: heuristicCategoryForward,
		Description: "Returns all destination transactions " +
			"which can be fully funded by the origins of their source. " +
			"If this heuristic " +
			"is placed at the root level a reverse lookup with the same " +
			"time as the forward lookup will be performed.",
		Parameter: &struct {
			DefaultValue string `json:"value,omitempty"`
			Description  string `json:"description,omitempty"`
			Type         string `json:"type,omitempty"`
		}{
			DefaultValue: "48",
			Description:  "Look forward time in hours",
			Type:         "int",
		},
	}
}

func (h *forwardAmountHeuristic) clone() heuristic {
	newHeuristic := *h
	return &newHeuristic
}

// forwardAmountHeuristic applies the following heuristic:
// - filters all destinations which can not be funded by the sources based on the denominations of the source
func (h *forwardAmountHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string, parentHeuristicUID string) (
	[]dbtxh.HeuristicResult, error) {
	// origins hold all origins found bei either the parent heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]dbtxh.HeuristicTransaction)
	// maps a cluster to its origin transactions
	clusterOrigins := make(map[dbtxh.ClusterUID]map[string]dbtxh.HeuristicTransaction)

	{ // separate enclosure so the results slice can be garbage collected
		var results []dbtxh.HeuristicTransaction
		if isParentHeuristicSet(parentHeuristicUID) {
			// get origins from parent heuristic
			var err error
			results, err = dbtxh.GetHeuristicResults(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, err = getDestinationTxOriginsTimeLimited(dgraph, g, txHash, h.lookForwardTime)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		var err error
		clusterOrigins, err = addOriginsToMap(clusterOrigins, results)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 || len(clusterOrigins) == 0 {
		return nil, errorNoOriginsAtStart
	}

	var clusterDestinations []struct {
		cluster dbtxh.ClusterUID
		txs     map[string]dbtxh.HeuristicTransaction
	}

	for c, txMap := range clusterOrigins {
		var txUIDs []string

		for k := range txMap {
			txUIDs = append(txUIDs, k)
		}
		destinations, err := getOriginDestinationsWithInputs(dgraph, g, txUIDs, h.lookForwardTime)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		destinationMap := make(map[string]dbtxh.HeuristicTransaction)
		for _, d := range destinations {
			destinationMap[d.UID] = d
		}

		clusterDestinations = append(clusterDestinations, struct {
			cluster dbtxh.ClusterUID
			txs     map[string]dbtxh.HeuristicTransaction
		}{cluster: c, txs: destinationMap})
	}

	originAmounts := buildSourceAmounts(origins)

	var filteredDestinations []dbtxh.HeuristicResult
	for _, destinations := range clusterDestinations {
		var clusterFilteredDestinations []dbtxh.DummyNode

		for _, tx := range destinations.txs {
			inputDenominationCounts := getDenominationCounts(tx)

			// check if the denominations of the destination transactions can be funded by the denomination of its cluster
			if containsDenomination(inputDenominationCounts, originAmounts[destinations.cluster]) {
				clusterFilteredDestinations = append(clusterFilteredDestinations, dbtxh.DummyNode{UID: tx.UID})
			}
		}

		// get a random origin of this cluster
		var originUID string
		for _, v := range clusterOrigins[destinations.cluster] {
			originUID = v.UID
		}

		filteredDestinations = append(filteredDestinations, dbtxh.HeuristicResult{
			Origin:       dbtxh.DummyNode{UID: originUID},
			Destinations: clusterFilteredDestinations,
		})
	}

	return filteredDestinations, nil
}
