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

// perfectMatchHeuristic - see exec for description
type perfectMatchHeuristic struct {
	heuristicType        string
	parameterDescription string
	c                    heuristics.Options
}

func newPerfectMatchHeuristic() heuristic {
	return &perfectMatchHeuristic{heuristicType: "perfect_match"}
}

func (h *perfectMatchHeuristic) getType() string {
	return h.heuristicType
}

func (h *perfectMatchHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *perfectMatchHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.c = c

	return nil
}

func (h *perfectMatchHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *perfectMatchHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Perfect Match",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "The perfect match heuristic filters all " +
			"origins of sources, which have denominations " +
			"without a perfect match for the denominations of " +
			"the destination transaction.",
	}
}

func (h *perfectMatchHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

// perfectMatchHeuristic applies the following heuristic:
//   - filter all origins of sources, which have denominations without a perfect match for the
//     denominations of the destination transaction
func (h *perfectMatchHeuristic) exec(dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
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
	for k, o := range originAmounts {
		if isEqualDenomination(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], db.UIDNode{UID: tx.UID})
			}
		}
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}

// returns true if all denominations with the same amount of denom1 are contained in denom2
func isEqualDenomination(denom1 [analytics.NumDenominations]int, denom2 [analytics.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] != d {
			return false
		}
	}
	return true
}
