package heuristics

import (
	"backend/analytics"
	"backend/analytics/graph"
	"backend/db"
	"backend/db/analytics/heuristics"
	"backend/external"
	"fmt"
	"github.com/qrest/gomisc/serror"
)

// reverseAmountHeuristic - see exec for description
type reverseAmountHeuristic struct {
	heuristicType        string
	parameterDescription string
	c                    heuristics.Options
}

func newReverseAmountHeuristic() heuristic {
	return &reverseAmountHeuristic{heuristicType: "reverse_amount"}
}

func (h *reverseAmountHeuristic) getType() string {
	return h.heuristicType
}

func (h *reverseAmountHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *reverseAmountHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.c = c

	return nil
}

func (h *reverseAmountHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *reverseAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *reverseAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Reverse Amount",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Returns all origins of sources, which " +
			"have equal or more denominations to fund the " +
			"destination transaction.",
	}
}

// reverseAmountHeuristic applies the following heuristic:
// - filter all origins of sources, which do not have equal or more denominations to fund the destination transaction
func (h *reverseAmountHeuristic) exec(dgraph external.Database, g *graph.Wrapper, parentHeuristicUID string) (
	[]heuristics.HeuristicCluster, error) {
	// origins hold all origins found bei either the parent heuristic
	// or the destination transaction specified by txHash
	origins := make(map[string]heuristics.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	var attributionMap map[heuristics.ClusterUID][]string

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	{ // separate enclosure so the results slice can be garbage collected
		parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
		if err != nil {
			return nil, err
		}
		var results []heuristics.HeuristicTransaction
		if parentHeuristicSet {
			// get origins from parent heuristic
			var err error
			results, attributionMap, err = heuristics.GetHeuristicTransactions(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			results, attributionMap, err = getDestinationTxOrigins(ctx, dgraph, g, h.c.TransactionHash, h.c)
			if err != nil {
				return nil, err
			}
		}

		sourceTransactionMap = addOriginsToMap(sourceTransactionMap, results)

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
	}

	transaction, err := heuristics.GetInputAmounts(dgraph, h.c.TransactionHash)
	if err != nil {
		return nil, err
	}

	inputDenominationCounts := getDenominationCounts(transaction)

	originAmounts := buildSourceAmounts(origins)

	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for clusterID, denominationSlice := range originAmounts {
		if containsDenomination(inputDenominationCounts, denominationSlice) {
			// save all transaction uids of a particular cluster to the return set
			for _, tx := range sourceTransactionMap[clusterID] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], db.UIDNode{UID: tx.UID})
			}
		}
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}

// containsDenomination returns true if all denominations with at
// least the same amount of denom1 are contained in denom2
func containsDenomination(denom1 [analytics.NumDenominations]int, denom2 [analytics.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] < d {
			return false
		}
	}
	return true
}
