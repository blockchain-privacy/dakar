package heuristics

import (
	"backend/analytics"
	"backend/analytics/graph"
	"backend/db/analytics/clustering"
	"backend/db/analytics/heuristics"
	"backend/external"
	"github.com/qrest/gomisc/serror"
)

// denominationTypeHeuristic - see exec for description
type denominationTypeHeuristic struct {
	heuristicType        string
	parameterDescription string
	userUID              string
	excludeAddresses     bool
	excludeSpendingGaps  bool
	clusterTypes         []clustering.ClusterType
}

// newDenominationTypeHeuristic constructs a denominationTypeHeuristic
func newDenominationTypeHeuristic(clusterTypes []clustering.ClusterType) *denominationTypeHeuristic {
	return &denominationTypeHeuristic{
		heuristicType: "denomination_type",
		clusterTypes:  clusterTypes,
	}
}

func (h denominationTypeHeuristic) getType() string {
	return h.heuristicType
}

func (h denominationTypeHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h denominationTypeHeuristic) hasParameter() bool {
	return false
}

func (h denominationTypeHeuristic) setParameter(_ string) error {
	return nil
}

// setClusterTypes sets additional cluster types, which are used to execute the heuristic.
// Multi-input clusters are always used to execute the heuristic,
// any cluster type set here will be used additionally. If at least one cluster type is set,
// then the consolidation of the multi-input clusters and the additional clusters will be used.
func (h *denominationTypeHeuristic) setClusterTypes(clusterTypes []clustering.ClusterType) error {
	if !areClusterTypesValid(clusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.clusterTypes = clusterTypes
	return nil
}

// getClusterTypes returns the cluster types this heuristic uses to cluster addresses
func (h *denominationTypeHeuristic) getClusterTypes() []clustering.ClusterType {
	return h.clusterTypes
}

// setExcludeAddresses sets whether certain addresses should be excluded from the lookups
func (h *denominationTypeHeuristic) setExcludeAddresses(excludeAddresses bool) {
	h.excludeAddresses = excludeAddresses
}

// getExcludeAddresses returns whether certain addresses should be excluded from the lookups
func (h *denominationTypeHeuristic) getExcludeAddresses() bool {
	return h.excludeAddresses
}

// setExcludeSpendingGaps sets whether mixing outputs with a spending gap should be traversed
func (h *denominationTypeHeuristic) setExcludeSpendingGaps(excludeSpendingGaps bool) {
	h.excludeSpendingGaps = excludeSpendingGaps
}

// getExcludeSpendingGaps returns whether mixing outputs with a spending gap should be traversed
func (h *denominationTypeHeuristic) getExcludeSpendingGaps() bool {
	return h.excludeSpendingGaps
}

// setUserUID sets the UID of the user who created this heuristic
func (h *denominationTypeHeuristic) setUserUID(uid string) {
	h.userUID = uid
}

func (h denominationTypeHeuristic) String() string {
	return "Type: " + h.heuristicType
}

func (h denominationTypeHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Denomination Type",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "The denomination type heuristic filters " +
			"all origins of clusters, which have denominations of " +
			"types which do not occur in the denominations of the " +
			"destination transaction. For example a destination " +
			"transaction spends 5 × 10.0001 and 10 × 1.00001. Now " +
			"all clusters are excluded which do not have these exact " +
			"two types of denominations.",
	}
}

func (h denominationTypeHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// denominationTypeHeuristic applies the following heuristic:
//   - filter all origins of sources, which have denominations of types which do not occur in the
//     denominations of the destination transaction
func (h denominationTypeHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
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
				return nil, err
			}
		} else {
			var err error
			results, attributionMap, err = getDestinationTxOrigins(dgraph, g, txHash, h.userUID,
				h.clusterTypes, h.excludeAddresses, h.excludeSpendingGaps)
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

	transaction, err := heuristics.GetInputAmounts(dgraph, txHash)
	if err != nil {
		return nil, err
	}

	inputDenominationCounts := getDenominationCounts(transaction)

	originAmounts := buildSourceAmounts(origins)

	resultClusters := make(map[heuristics.ClusterUID][]heuristics.HeuristicResult)
	for k, o := range originAmounts {
		if hasSameDenominationTypes(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], heuristics.HeuristicResult{
					Origin: heuristics.DummyNode{UID: tx.UID},
				})
			}
		}
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}

// returns true if both destinationDenominations and originDenominations have the exact same types
func hasSameDenominationTypes(destinationDenominations [analytics.NumDenominations]int,
	originDenominations [analytics.NumDenominations]int) bool {
	for i, destinationDenomination := range destinationDenominations {
		if originDenominations[i] == destinationDenomination && destinationDenomination == 0 {
			continue
		}

		if (originDenominations[i] > 0 && destinationDenomination == 0) ||
			(originDenominations[i] == 0 && destinationDenomination > 0) {
			return false
		}
	}
	return true
}
