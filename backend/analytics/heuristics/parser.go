package heuristics

import (
	"backend/db/analytics/clustering"
	"errors"
)

// ValidHeuristicTypes includes all heuristics which are possible to receive from the frontend.
// New heuristics must be added here
var ValidHeuristicTypes = []heuristic{
	newOneSourceHeuristic(0, nil),
	newReverseAmountHeuristic(nil),
	newPerfectMatchHeuristic(nil),
	newDenominationTypeHeuristic(nil),
	newReverseLookupHeuristic(0, nil),
	newForwardLookupHeuristic(0, nil),
	newForwardAmountHeuristic(0, nil),
	newSimpleForwardHeuristic(0, nil),
}

// typeMap K: heuristic types, v: heuristics
var typeMap = make(map[string]heuristic)

var errHeuristicNotValid = errors.New("error heuristics are not valid")

// areClusterTypesValid checks if the given clusterTypes are valid
func areClusterTypesValid(clusterTypes []clustering.ClusterType) bool {
	if len(clusterTypes) == 0 {
		return true
	}

	// for now only one additional cluster type exists
	return len(clusterTypes) == 1 && clusterTypes[0] == clustering.TypeCustom
}

func init() {
	for _, h := range ValidHeuristicTypes {
		typeMap[h.getType()] = h
	}
}
