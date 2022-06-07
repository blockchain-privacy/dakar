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

// reverseAmountHeuristic - see exec for description
type reverseAmountHeuristic struct {
	heuristicType        string
	parameterDescription string
	userUID              string
	excludeAddresses     bool
	clusterTypes         []clustering.ClusterType
}

// newReverseAmountHeuristic constructs an reverseAmountHeuristic
func newReverseAmountHeuristic(clusterTypes []clustering.ClusterType) *reverseAmountHeuristic {
	return &reverseAmountHeuristic{
		heuristicType: "reverse_amount",
		clusterTypes:  clusterTypes,
	}
}

func (h reverseAmountHeuristic) getType() string {
	return h.heuristicType
}

func (h reverseAmountHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h reverseAmountHeuristic) hasParameter() bool {
	return false
}

func (h reverseAmountHeuristic) setParameter(_ string) error {
	return nil
}

// setClusterTypes sets additional cluster types, which are used to execute the heuristic.
// Multi-input clusters are always used to execute the heuristic,
// any cluster type set here will be used additionally. If at least one cluster type is set,
// then the consolidation of the multi-input clusters and the additional clusters will be used.
func (h *reverseAmountHeuristic) setClusterTypes(clusterTypes []clustering.ClusterType) error {
	if !areClusterTypesValid(clusterTypes) {
		return errInvalidClusterTypes
	}

	h.clusterTypes = clusterTypes
	return nil
}

// getClusterTypes returns the cluster types this heuristic uses to cluster addresses
func (h *reverseAmountHeuristic) getClusterTypes() []clustering.ClusterType {
	return h.clusterTypes
}

// setExcludeAddresses sets whether certain addresses should be excluded from the lookups
func (h *reverseAmountHeuristic) setExcludeAddresses(excludeAddresses bool) {
	h.excludeAddresses = excludeAddresses
}

// getExcludeAddresses returns whether certain addresses should be excluded from the lookups
func (h *reverseAmountHeuristic) getExcludeAddresses() bool {
	return h.excludeAddresses
}

// setUserUID sets the UID of the user who created this heuristic
func (h *reverseAmountHeuristic) setUserUID(uid string) {
	h.userUID = uid
}

func (h reverseAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h reverseAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Reverse Amount",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Returns all origins of sources, which " +
			"have equal or more denominations to fund the " +
			"destination transaction.",
	}
}

func (h reverseAmountHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// reverseAmountHeuristic applies the following heuristic:
// - filter all origins of sources, which do not have equal or more denominations to fund the destination transaction
func (h reverseAmountHeuristic) exec(dgraph external.Database, g *graph.Wrapper,
	txHash string, parentHeuristicUID string) (
	[]heuristics.HeuristicCluster, error) {
	// origins hold all origins found bei either the parent heuristic
	// or the destination transaction specified by txHash
	origins := make(map[string]heuristics.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	var attributionMap map[heuristics.ClusterUID][]string
	{ // separate enclosure so the results slice can be garbage collected
		var results []heuristics.HeuristicTransaction
		if isParentHeuristicSet(parentHeuristicUID) {
			// get origins from parent heuristic
			var err error
			results, attributionMap, err = heuristics.GetHeuristicResults(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, attributionMap, err = getDestinationTxOrigins(dgraph, g, txHash, h.userUID,
				h.clusterTypes, h.excludeAddresses)
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
	for clusterID, denominationSlice := range originAmounts {
		if containsDenomination(inputDenominationCounts, denominationSlice) {
			// save all transaction uids of a particular cluster to the return set
			for _, tx := range sourceTransactionMap[clusterID] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], heuristics.HeuristicResult{
					Origin: heuristics.DummyNode{UID: tx.UID},
				})
			}
		}
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}

// containsDenomination returns true if all denominations with at
// least the same amount of denom1 are contained in denom2
func containsDenomination(denom1 [dbop.NumDenominations]int, denom2 [dbop.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] < d {
			return false
		}
	}
	return true
}
