package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	dbop "backend/db/output"
	"backend/external"

	"fmt"
)

// AmountHeuristic - see exec for description
type AmountHeuristic struct {
	heuristicType        string
	parameterDescription string
}

// NewAmountHeuristic constructs an AmountHeuristic
func NewAmountHeuristic() AmountHeuristic {
	return AmountHeuristic{
		heuristicType: "global_amount",
	}
}

func (h AmountHeuristic) getType() string {
	return h.heuristicType
}

func (h AmountHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h AmountHeuristic) hasParameter() bool {
	return false
}

func (h AmountHeuristic) setParameter(_ string) error {
	return nil
}

func (h AmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h AmountHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// AmountHeuristic applies the following heuristic:
// - filter all origins of sources, which do not have equal or more denominations to fund the destination transaction
func (h AmountHeuristic) exec(dgraph *external.GraphDB, g *graph.Wrapper, txHash string, parentHeuristicUID string) ([]string, error) {
	// origins holds all origins found bei either the parent heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]dbtxh.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[graph.ClusterID]map[string]dbtxh.HeuristicTransaction)
	var clusters map[string]graph.ClusterID
	{ // separate enclosure so the results slice can be garbage collected
		var results []dbtxh.HeuristicTransaction
		parentHeuristicSet := isParentHeuristicSet(parentHeuristicUID)

		if parentHeuristicSet {
			// get origins from parent heuristic
			var err error
			results, err = dbtxh.GetHeuristicResults(dgraph, parentHeuristicUID)
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
		sourceTransactionMap, clusters, err = addOriginsToMap(g, sourceTransactionMap, results)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 {
		return nil, ErrorNoOriginsAtStart
	}

	transaction, err := dbtxh.GetInputAmounts(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	inputDenominationCounts := getDenominationCounts(transaction)

	originAmounts := buildSourceAmounts(origins, clusters)

	var filteredOrigins []string
	for k, o := range originAmounts {
		if containsDenomination(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				filteredOrigins = append(filteredOrigins, tx.UID)
			}
		}
	}

	return filteredOrigins, nil
}

// returns true if all denominations with at least the same amount of denom1 are contained in denom2
func containsDenomination(denom1 [dbop.NumDenominations]int, denom2 [dbop.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] < d {
			return false
		}
	}
	return true
}
