package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db/analytics/heuristics"
	dbop "backend/db/output"
	"backend/external"

	"fmt"
)

// denominationTypeHeuristic - see exec for description
type denominationTypeHeuristic struct {
	heuristicType        string
	parameterDescription string
}

// newDenominationTypeHeuristic constructs a denominationTypeHeuristic
func newDenominationTypeHeuristic() denominationTypeHeuristic {
	return denominationTypeHeuristic{
		heuristicType: "denomination_type",
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
	parentHeuristicUID string) ([]heuristics.HeuristicResult, error) {
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
			results, err = getDestinationTxOrigins(dgraph, g, txHash)
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

	var filteredOrigins []heuristics.HeuristicResult
	for k, o := range originAmounts {
		if hasSameDenominationTypes(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				filteredOrigins = append(filteredOrigins, heuristics.HeuristicResult{Origin: heuristics.DummyNode{UID: tx.UID}})
			}
		}
	}

	return filteredOrigins, nil
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
