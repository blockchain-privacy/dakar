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

// ForwardAmountHeuristic - see exec for description
type ForwardAmountHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookForwardTime      time.Duration
}

// NewForwardAmountHeuristic constructs an ForwardAmountHeuristic. hoursToLookForward in hours.
func NewForwardAmountHeuristic(hoursToLookForward uint32) *ForwardAmountHeuristic {
	lForwardTime := time.Duration(hoursToLookForward) * time.Hour
	return &ForwardAmountHeuristic{
		heuristicType:        "forward_amount",
		lookForwardTime:      lForwardTime,
		parameterDescription: lForwardTime.String(),
	}
}

func (h *ForwardAmountHeuristic) getType() string {
	return h.heuristicType
}

func (h *ForwardAmountHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *ForwardAmountHeuristic) hasParameter() bool {
	return true
}

func (h *ForwardAmountHeuristic) setParameter(p string) error {
	hoursToLookForward, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.parameterDescription = strconv.FormatUint(hoursToLookForward, 10)
	return nil
}

func (h *ForwardAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h *ForwardAmountHeuristic) GetDescriptor() Descriptor {
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

func (h *ForwardAmountHeuristic) clone() Heuristic {
	newHeuristic := *h
	return &newHeuristic
}

// ForwardAmountHeuristic applies the following Heuristic:
// - filters all destinations which can not be funded by the sources based on the denominations of the source
func (h *ForwardAmountHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string, parentHeuristicUID string) (
	[]dbtxh.HeuristicResult, error) {
	// origins holds all origins found bei either the parent Heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]dbtxh.HeuristicTransaction)
	// maps a cluster to its origin transactions
	clusterOrigins := make(map[graph.ClusterID]map[string]dbtxh.HeuristicTransaction)
	var clusters map[string]graph.ClusterID
	{ // separate enclosure so the results slice can be garbage collected
		var results []dbtxh.HeuristicTransaction
		if isParentHeuristicSet(parentHeuristicUID) {
			// get origins from parent Heuristic
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
		clusterOrigins, clusters, err = addOriginsToMap(g, clusterOrigins, results)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 || len(clusters) == 0 || len(clusterOrigins) == 0 {
		return nil, ErrorNoOriginsAtStart
	}

	var clusterDestinations []struct {
		cluster graph.ClusterID
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
			cluster graph.ClusterID
			txs     map[string]dbtxh.HeuristicTransaction
		}{cluster: c, txs: destinationMap})
	}

	originAmounts := buildSourceAmounts(origins, clusters)

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
