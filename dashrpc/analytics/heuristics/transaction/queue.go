package transaction

import (
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
)

var (
	errHeuristicTypeNotFound = errors.New("error heuristic type not found")
	errHeuristicNotValid     = errors.New("heuristics are not valid")
)

var validHeuristics = []heuristic{NewOneSourceHeuristic(0), NewAmountHeuristic(),
	NewPerfectMatchHeuristic(), NewDenominationTypeHeuristic()}

// isValid checks if the given heuristics are all valid
func isValid(hMap map[string]heuristic, heuristics []dbtxh.FrontendHeuristic) bool {
	if len(heuristics) == 0 {
		return false
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
		if heuristic, ok := hMap[h.Type]; !ok || (heuristic.hasParameter() && len(h.Parameter) == 0) {
			return false
		}
	}

	return true
}

func buildHeuristics(hMap map[string]heuristic, heuristics []dbtxh.FrontendHeuristic) (builtHeuristics []heuristic,
	err error) {
	for _, h := range heuristics {
		if newHeuristic, ok := hMap[h.Type]; ok {
			if newHeuristic.hasParameter() {
				err = newHeuristic.setParameter(h.Parameter)
				if err != nil {
					err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
					return
				}
			}

			builtHeuristics = append(builtHeuristics, newHeuristic)

		} else {
			err = errHeuristicTypeNotFound
			return
		}
	}

	return
}

func DoExecution(dgraph *dgo.Dgraph, heuristics []dbtxh.FrontendHeuristic) error {
	heuristicMap := make(map[string]heuristic)

	for _, h := range validHeuristics {
		heuristicMap[h.getType()] = h
	}

	if !isValid(heuristicMap, heuristics) {
		return errHeuristicNotValid
	}

	newHeuristics, err := buildHeuristics(heuristicMap, heuristics)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	fmt.Println(newHeuristics)

	return nil
}
