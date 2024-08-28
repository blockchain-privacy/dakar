package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/exclusion"
	"backend/db/analytics/heuristics"
	"backend/external"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"strconv"
	"time"
)

// forwardAmountHeuristic - see exec for description
type forwardAmountHeuristic struct {
	c                    heuristics.Options
	heuristicType        string
	parameterDescription string
	lookForwardTime      time.Duration
}

func newForwardAmountHeuristic() heuristic {
	return &forwardAmountHeuristic{heuristicType: "forward_amount"}
}

func (h *forwardAmountHeuristic) getType() string {
	return h.heuristicType
}

func (h *forwardAmountHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *forwardAmountHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	hoursToLookForward, err := strconv.ParseInt(c.Parameter, 10, 64)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.parameterDescription = strconv.FormatInt(hoursToLookForward, 10)
	h.c = c

	return nil
}

func (h *forwardAmountHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *forwardAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
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

// forwardAmountHeuristic applies the following heuristic:
// - filters all destinations which can not be funded by the sources based on the denominations of the source
func (h *forwardAmountHeuristic) exec(dgraph external.Database, g *graph.Wrapper, parentHeuristicUID string) (
	[]heuristics.HeuristicCluster, error) {
	if h.lookForwardTime == 0 {
		return nil, nil
	}
	// origins hold all origins found bei either the parent heuristic
	// or the destination transaction specified by txHash
	origins := make(map[string]heuristics.HeuristicTransaction)
	// maps a cluster to its origin transactions
	clusterOrigins := make(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	var attributionMap map[heuristics.ClusterUID][]string

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	{ // separate enclosure so the results slice can be garbage collected
		var results []heuristics.HeuristicTransaction
		parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
		if err != nil {
			return nil, err
		}

		if parentHeuristicSet {
			// get origins from parent heuristic
			var err error
			results, attributionMap, err = heuristics.GetHeuristicTransactions(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			results, attributionMap, err = getDestinationTxOriginsTimeLimited(ctx, dgraph, g,
				h.c.TransactionHash, h.lookForwardTime, h.c)
			if err != nil {
				return nil, err
			}
		}

		clusterOrigins = addOriginsToMap(clusterOrigins, results)

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 || len(clusterOrigins) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
	}

	var exclusions []string
	if h.c.ExcludeAddresses {
		var err error
		exclusions, err = exclusion.GetAddressExclusionUIDs(dgraph, h.c.UserUID)
		if err != nil {
			return nil, err
		}
	}

	type clusterDestination struct {
		cluster heuristics.ClusterUID
		txs     map[string]heuristics.HeuristicTransaction
	}

	clusterDestinations := make([]clusterDestination, 0, len(clusterOrigins))
	for c, txMap := range clusterOrigins {
		destinations, err := getOriginDestinationsWithInputs(dgraph, g, cliutil.GetMapKeys(txMap), h.lookForwardTime,
			exclusions, h.c.ExcludeSpendingGaps)
		if err != nil {
			return nil, err
		}

		destinationMap := make(map[string]heuristics.HeuristicTransaction)
		for _, d := range destinations {
			destinationMap[d.UID] = d
		}

		clusterDestinations = append(clusterDestinations, clusterDestination{cluster: c, txs: destinationMap})
	}

	originAmounts := buildSourceAmounts(origins)

	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for _, destinations := range clusterDestinations {
		var clusterFilteredDestinations []db.UIDNode
		for _, tx := range destinations.txs {
			inputDenominationCounts := getDenominationCounts(tx)

			// check if the denominations of the destination transactions can be funded by the denomination of its cluster
			if containsDenomination(inputDenominationCounts, originAmounts[destinations.cluster]) {
				clusterFilteredDestinations = append(clusterFilteredDestinations, db.UIDNode{UID: tx.UID})
			}
		}

		// get cluster ID of a random origin of this cluster
		var clusterID heuristics.ClusterUID
		for _, v := range clusterOrigins[destinations.cluster] {
			clusterID = v.Cluster
			break
		}

		resultClusters[clusterID] = clusterFilteredDestinations
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
