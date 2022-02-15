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

// denominationTypeHeuristic - see exec for description
type denominationTypeHeuristic struct {
	heuristicType        string
	parameterDescription string
	userUID              string
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
		return errorInvalidClusterTypes
	}

	h.clusterTypes = clusterTypes
	return nil
}

// setUserUID sets the UID of the user who created this heuristic
func (h *denominationTypeHeuristic) setUserUID(uid string) {
	h.userUID = uid
}

func (h denominationTypeHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
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
// - filter all origins of sources, which have denominations of types which do not occur in the
//		denominations of the destination transaction
func (h denominationTypeHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
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
		if hasSameDenominationTypes(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], heuristics.HeuristicResult{
					Origin: heuristics.DummyNode{UID: tx.UID},
				})
			}
		}
	}

	return createHeuristicClusters(resultClusters), nil
}

// returns true if both destinationDenominations and originDenominations have the exact same types
func hasSameDenominationTypes(destinationDenominations [dbop.NumDenominations]int,
	originDenominations [dbop.NumDenominations]int) bool {
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
