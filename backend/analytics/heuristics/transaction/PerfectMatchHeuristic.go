package transaction

import (
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	dbop "backend/db/output"

	"fmt"

	"github.com/dgraph-io/dgo/v2"
)

type PerfectMatchHeuristic struct {
	heuristicType        string
	parameterDescription string
}

// PerfectMatchHeuristic constructor
func NewPerfectMatchHeuristic() PerfectMatchHeuristic {
	return PerfectMatchHeuristic{
		heuristicType: "perfect_match",
	}
}

func (h PerfectMatchHeuristic) getType() string {
	return h.heuristicType
}

func (h PerfectMatchHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h PerfectMatchHeuristic) hasParameter() bool {
	return false
}

func (h PerfectMatchHeuristic) setParameter(p string) error {
	return nil
}

func (h PerfectMatchHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h PerfectMatchHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// PerfectMatchHeuristic applies the following heuristic:
// - filter all origins of sources, which have denominations without a perfect match for the
//		denominations of the destination transaction
func (h PerfectMatchHeuristic) exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error) {
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
		if isEqualDenomination(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				filteredOrigins = append(filteredOrigins, tx.Uid)
			}
		}
	}

	return filteredOrigins, nil
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
