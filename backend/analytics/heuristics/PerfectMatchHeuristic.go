package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db/analytics/clustering"
	"backend/db/analytics/heuristics"
	dbop "backend/db/output"
	"backend/external"

	"fmt"
)

// perfectMatchHeuristic - see exec for description
type perfectMatchHeuristic struct {
	heuristicType        string
	parameterDescription string
	userUID              string
	clusterTypes         []clustering.ClusterType
}

// newPerfectMatchHeuristic constructs a perfectMatchHeuristic
func newPerfectMatchHeuristic(clusterTypes []clustering.ClusterType) *perfectMatchHeuristic {
	return &perfectMatchHeuristic{
		heuristicType: "perfect_match",
		clusterTypes:  clusterTypes,
	}
}

func (h perfectMatchHeuristic) getType() string {
	return h.heuristicType
}

func (h perfectMatchHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h perfectMatchHeuristic) hasParameter() bool {
	return false
}

func (h perfectMatchHeuristic) setParameter(_ string) error {
	return nil
}

// setClusterTypes sets additional cluster types, which are used to execute the heuristic.
// Multi-input clusters are always used to execute the heuristic,
// any cluster type set here will be used additionally. If at least one cluster type is set,
// then the consolidation of the multi-input clusters and the additional clusters will be used.
func (h *perfectMatchHeuristic) setClusterTypes(clusterTypes []clustering.ClusterType) error {
	if !areClusterTypesValid(clusterTypes) {
		return errorInvalidClusterTypes
	}

	h.clusterTypes = clusterTypes
	return nil
}

// setUserUID sets the UID of the user who created this heuristic
func (h *perfectMatchHeuristic) setUserUID(uid string) {
	h.userUID = uid
}

func (h perfectMatchHeuristic) GetDescriptor() Descriptor {
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

func (h perfectMatchHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h perfectMatchHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// perfectMatchHeuristic applies the following heuristic:
// - filter all origins of sources, which have denominations without a perfect match for the
//		denominations of the destination transaction
func (h perfectMatchHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	// origins hold all origins found bei either the parent heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]heuristics.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction)
	{ // separate enclosure so the results slice can be garbage collected
		var results []heuristics.HeuristicTransaction
		parentHeuristicSet := isParentHeuristicSet(parentHeuristicUID)

		if parentHeuristicSet {
			// get origins from parent heuristic
			var err error
			results, err = heuristics.GetHeuristicResults(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, err = getDestinationTxOrigins(dgraph, g, txHash, h.userUID, h.clusterTypes)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}
		var err error
		sourceTransactionMap, err = addOriginsToMap(sourceTransactionMap, results)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 {
		return nil, errorNoOriginsAtStart
	}

	transaction, err := heuristics.GetInputAmounts(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	inputDenominationCounts := getDenominationCounts(transaction)

	originAmounts := buildSourceAmounts(origins)

	resultClusters := make(map[heuristics.ClusterUID][]heuristics.HeuristicResult)
	for k, o := range originAmounts {
		if isEqualDenomination(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], heuristics.HeuristicResult{
					Origin: heuristics.DummyNode{UID: tx.UID},
				})
			}
		}
	}

	return createHeuristicClusters(resultClusters), nil
}

// returns true if all denominations with the same amount of denom1 are contained in denom2
func isEqualDenomination(denom1 [dbop.NumDenominations]int, denom2 [dbop.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] != d {
			return false
		}
	}
	return true
}
