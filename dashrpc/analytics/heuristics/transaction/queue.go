package transaction

import (
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"errors"
	"github.com/dgraph-io/dgo/v2"
)

var validHeuristics = []heuristic{NewOneSourceHeuristic(0), NewAmountHeuristic(),
	NewPerfectMatchHeuristic(), NewDenominationTypeHeuristic()}

func isValid(heuristics []dbtxh.FrontendHeuristic) bool {
	if len(heuristics) == 0 {
		return false
	}

	typeToParameter := make(map[string]heuristic)

	for _, h := range validHeuristics {
		typeToParameter[h.getType()] = h
	}

	for _, h := range heuristics {
		// more than one parent is not allowed
		if len(h.ParentHeuristic) > 1 {
			return false
		}

		// at least one must be set
		if len(h.ParentHeuristic) == 0 && len(h.ChildHeuristics) == 0 {
			return false
		}

		// type must by in valid set; parameter must be set if the map has a parameter
		if heuristic, ok := typeToParameter[h.Type]; !ok || (heuristic.hasParameter() && len(h.Parameter) == 0) {
			return false
		}
	}

	return true
}

func DoExecution(dgraph *dgo.Dgraph, heuristics []dbtxh.FrontendHeuristic) error {
	if !isValid(heuristics) {
		return errors.New("heuristics are not valid")
	}

	return nil
}
