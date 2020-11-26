package transaction

import (
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	dbop "dashrpc/db/output"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
)

type AmountHeuristic struct {
	heuristicType        string
	parameterDescription string
}

// AmountHeuristic constructor
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

func (h AmountHeuristic) setParameter(p string) error {
	return nil
}

func (h AmountHeuristic) String() string {
	return h.heuristicType
}

// AmountHeuristic applies the following heuristic:
// - filter all origins of sources, which do not have equal or more denominations to fund the destination transaction
func (h AmountHeuristic) exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error) {
	// origins holds all origins found bei either the parent heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]dbtxh.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[string]map[string]dbtxh.HeuristicTransaction)

	{ // separate enclosure so the results slice can be garbage collected
		var results []dbtxh.HeuristicTransaction
		parentHeuristicSet := isParentHeuristicSet(parentHeuristicUid)

		if parentHeuristicSet {
			// get origins from parent heuristic
			var err error
			results, err = dbtxh.GetHeuristicResults(dgraph, parentHeuristicUid)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, err = dbtxh.GetOrigins(dgraph, txHash)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		sourceTransactionMap = addOriginsToMap(sourceTransactionMap, results)

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.Uid] = r
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

	originAmounts := buildSourceAmounts(origins)

	var filteredOrigins []string
	for k, o := range originAmounts {
		if containsDenomination(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				filteredOrigins = append(filteredOrigins, tx.Uid)
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
