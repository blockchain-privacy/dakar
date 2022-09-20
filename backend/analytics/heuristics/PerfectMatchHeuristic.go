package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/clustering"
	"backend/db/analytics/heuristics"
	"backend/external"

	"fmt"
)

// perfectMatchHeuristic - see exec for description
type perfectMatchHeuristic struct {
	heuristicType        string
	parameterDescription string
	userUID              string
	excludeAddresses     bool
	excludeSpendingGaps  bool
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
		return errInvalidClusterTypes
	}

	h.clusterTypes = clusterTypes
	return nil
}

// getClusterTypes returns the cluster types this heuristic uses to cluster addresses
func (h *perfectMatchHeuristic) getClusterTypes() []clustering.ClusterType {
	return h.clusterTypes
}

// setExcludeAddresses sets whether certain addresses should be excluded from the lookups
func (h *perfectMatchHeuristic) setExcludeAddresses(excludeAddresses bool) {
	h.excludeAddresses = excludeAddresses
}

// getExcludeAddresses returns whether certain addresses should be excluded from the lookups
func (h *perfectMatchHeuristic) getExcludeAddresses() bool {
	return h.excludeAddresses
}

// setExcludeSpendingGaps sets whether mixing outputs with a spending gap should be traversed
func (h *perfectMatchHeuristic) setExcludeSpendingGaps(excludeSpendingGaps bool) {
	h.excludeSpendingGaps = excludeSpendingGaps
}

// getExcludeSpendingGaps returns whether mixing outputs with a spending gap should be traversed
func (h *perfectMatchHeuristic) getExcludeSpendingGaps() bool {
	return h.excludeSpendingGaps
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
//   - filter all origins of sources, which have denominations without a perfect match for the
//     denominations of the destination transaction
func (h perfectMatchHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	// origins hold all origins found bei either the parent heuristic
	// or the destination transaction specified by txHash
	origins := make(map[string]heuristics.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	var attributionMap map[heuristics.ClusterUID][]string
	{ // separate enclosure so the results slice can be garbage collected
		var results []heuristics.HeuristicTransaction
		parentHeuristicSet := isParentHeuristicSet(parentHeuristicUID)

		if parentHeuristicSet {
			// get origins from parent heuristic
			var err error
			results, attributionMap, err = heuristics.GetHeuristicResults(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, attributionMap, err = getDestinationTxOrigins(dgraph, g, txHash, h.userUID,
				h.clusterTypes, h.excludeAddresses, h.excludeSpendingGaps)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		sourceTransactionMap = addOriginsToMap(sourceTransactionMap, results)

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 {
		return nil, errNoOriginsAtStart
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

	return createHeuristicClusters(resultClusters, attributionMap), nil
}

// returns true if all denominations with the same amount of denom1 are contained in denom2
func isEqualDenomination(denom1 [db.NumDenominations]int, denom2 [db.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] != d {
			return false
		}
	}
	return true
}
