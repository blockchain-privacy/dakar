package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	dbop "backend/db/output"

	"fmt"

	"github.com/dgraph-io/dgo/v210"
)

type DenominationTypeHeuristic struct {
	heuristicType        string
	parameterDescription string
}

// NewDenominationTypeHeuristic constructs a DenominationTypeHeuristic
func NewDenominationTypeHeuristic() DenominationTypeHeuristic {
	return DenominationTypeHeuristic{
		heuristicType: "denomination_type",
	}
}

func (h DenominationTypeHeuristic) getType() string {
	return h.heuristicType
}

func (h DenominationTypeHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h DenominationTypeHeuristic) hasParameter() bool {
	return false
}

func (h DenominationTypeHeuristic) setParameter(_ string) error {
	return nil
}

func (h DenominationTypeHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h DenominationTypeHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// DenominationTypeHeuristic applies the following heuristic:
// - filter all origins of sources, which have denominations of types which do not occur in the
//		denominations of the destination transaction
func (h DenominationTypeHeuristic) exec(dgraph *dgo.Dgraph, g *graph.Wrapper, txHash string,
	parentHeuristicUid string) ([]string, error) {
	// origins holds all origins found bei either the parent heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]dbtxh.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[graph.ClusterId]map[string]dbtxh.HeuristicTransaction)
	var clusters map[string]graph.ClusterId
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

	originAmounts := buildSourceAmounts(origins, clusters)

	var filteredOrigins []string
	for k, o := range originAmounts {
		if hasSameDenominationTypes(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				filteredOrigins = append(filteredOrigins, tx.Uid)
			}
		}
	}

	return filteredOrigins, nil
}

// returns true if denom1 has only denominations for the same types as denom2
func hasSameDenominationTypes(denom1 [dbop.NumDenominations]int, denom2 [dbop.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] == d && d == 0 {
			continue
		}

		if (denom2[i] > 0 && d == 0) || (denom2[i] == 0 && d > 0) {
			return false
		}
	}
	return true
}
