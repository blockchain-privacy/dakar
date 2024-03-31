package heuristics

import (
	"backend/cmd/cliutil"
	"backend/db/analytics/clustering"
	"backend/db/analytics/heuristics"
	"errors"
)

// ValidHeuristicTypes includes all heuristics which are possible to receive from the frontend.
// New heuristics must be added here
var ValidHeuristicTypes = []heuristic{newOneSourceHeuristic(0, nil), newReverseAmountHeuristic(nil),
	newPerfectMatchHeuristic(nil), newDenominationTypeHeuristic(nil), newReverseLookupHeuristic(0, nil),
	newForwardLookupHeuristic(0, nil), newForwardAmountHeuristic(0, nil)}

// typeMap K: heuristic types, v: heuristics
var typeMap = make(map[string]heuristic)

// errors for this file
var (
	errHeuristicTypeNotFound = errors.New("error heuristic type not found")
	errHeuristicNotValid     = errors.New("error heuristics are not valid")
)

type heuristicTreeElement struct {
	parentHeuristicUID string
	heuristic          heuristic
}

func buildHeuristicTreeElement(hMap map[string]heuristic, h heuristics.DatabaseHeuristicRequest,
	userUID string) (element heuristicTreeElement, err error) {
	// create new heuristic
	if modelHeuristic, ok := hMap[h.Type]; ok {
		newHeuristic := modelHeuristic.clone()
		if newHeuristic.hasParameter() {
			err = newHeuristic.setParameter(h.Parameter)
			if err != nil {
				return
			}
		}

		err = newHeuristic.setClusterTypes(h.ClusterTypes)
		if err != nil {
			return
		}

		newHeuristic.setUserUID(userUID)
		newHeuristic.setExcludeAddresses(h.ExcludeAddresses)
		newHeuristic.setExcludeSpendingGaps(h.ExcludeSpendingGaps)
		newHeuristic.setWorkspaceUID(h.WorkspaceUID)

		element = heuristicTreeElement{
			parentHeuristicUID: h.ParentHeuristicUID,
			heuristic:          newHeuristic,
		}
	} else {
		err = cliutil.NewStackError(errHeuristicTypeNotFound)
		return
	}

	return
}

// areClusterTypesValid checks if the given clusterTypes are valid
func areClusterTypesValid(clusterTypes []clustering.ClusterType) bool {
	if len(clusterTypes) == 0 {
		return true
	}

	// for now only one additional cluster type exists
	return len(clusterTypes) == 1 && clusterTypes[0] == clustering.TypeCustom
}
