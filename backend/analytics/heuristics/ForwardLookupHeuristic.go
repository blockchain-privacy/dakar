package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db/analytics/clustering"
	"backend/db/analytics/heuristics"
	"backend/external"

	"fmt"
	"strconv"
	"time"
)

// forwardLookupHeuristic - see exec for description
type forwardLookupHeuristic struct {
	heuristicType        string
	parameterDescription string
	userUID              string
	lookForwardTime      time.Duration
	clusterTypes         []clustering.ClusterType
}

// newForwardLookupHeuristic constructs a forwardLookupHeuristic. hoursToLookForward in hours.
func newForwardLookupHeuristic(hoursToLookForward uint32, clusterTypes []clustering.ClusterType) *forwardLookupHeuristic {
	lForwardTime := time.Duration(hoursToLookForward) * time.Hour
	return &forwardLookupHeuristic{
		heuristicType:        "forward_lookup",
		lookForwardTime:      lForwardTime,
		parameterDescription: lForwardTime.String(),
		clusterTypes:         clusterTypes,
	}
}

func (h forwardLookupHeuristic) getType() string {
	return h.heuristicType
}

func (h forwardLookupHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h forwardLookupHeuristic) hasParameter() bool {
	return true
}

func (h *forwardLookupHeuristic) setParameter(p string) error {
	hoursToLookForward, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.parameterDescription = strconv.FormatUint(hoursToLookForward, 10)
	return nil
}

// setClusterTypes sets additional cluster types, which are used to execute the heuristic.
// Multi-input clusters are always used to execute the heuristic,
// any cluster type set here will be used additionally. If at least one cluster type is set,
// then the consolidation of the multi-input clusters and the additional clusters will be used.
func (h *forwardLookupHeuristic) setClusterTypes(clusterTypes []clustering.ClusterType) error {
	if !areClusterTypesValid(clusterTypes) {
		return errorInvalidClusterTypes
	}

	h.clusterTypes = clusterTypes
	return nil
}

// setUserUID sets the UID of the user who created this heuristic
func (h *forwardLookupHeuristic) setUserUID(uid string) {
	h.userUID = uid
}

func (h forwardLookupHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h forwardLookupHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Forward Lookup",
		Type:     h.heuristicType,
		Category: heuristicCategoryForward,
		Description: "Performs a forward lookup for each origin " +
			"transaction of the parent heuristic. " +
			"If this heuristic " +
			"is placed at the root level a reverse lookup with the same " +
			"time as the forward lookup will be performed.",
		Parameter: &struct {
			DefaultValue string `json:"value,omitempty"`
			Description  string `json:"description,omitempty"`
			Type         string `json:"type,omitempty"`
		}{
			DefaultValue: "48",
			Description:  "Look forward time in hours",
			Type:         "int",
		},
	}
}

func (h forwardLookupHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// forwardLookupHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h forwardLookupHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	var results []heuristics.HeuristicTransaction
	// resultAttributionMap maps a clusterUID to a slice of attribution UIDs
	resultAttributionMap := make(map[heuristics.ClusterUID][]string)
	{ // separate enclosure so the results slice can be garbage collected

		if isParentHeuristicSet(parentHeuristicUID) {
			// get origins from parent heuristic
			var err error
			results, err = heuristics.GetHeuristicResults(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, resultAttributionMap, err = getDestinationTxOriginsTimeLimited(dgraph, g, txHash, h.lookForwardTime, h.userUID, h.clusterTypes)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		if len(results) == 0 {
			return nil, errorNoOriginsAtStart
		}
	}

	// todo remove
	info(resultAttributionMap)

	resultClusters := make(map[heuristics.ClusterUID][]heuristics.HeuristicResult)
	// resultAttributionMap maps a clusterUID to a slice of attribution UIDs
	destinationAttributionMap := make(map[heuristics.ClusterUID][]string)
	for _, o := range results {
		destinations, dstAttributionMap, err := getOriginDestinationsWithOutputs(dgraph, g, []string{o.UID}, h.lookForwardTime, h.userUID, h.clusterTypes)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		result := heuristics.HeuristicResult{
			Origin: heuristics.DummyNode{UID: o.UID},
		}

		for _, v := range destinations {
			result.Destinations = append(result.Destinations, heuristics.DummyNode{UID: v.UID})
		}

		resultClusters[o.Cluster] = append(resultClusters[o.Cluster], result)

		// merge the attribution maps
		for id, attributions := range dstAttributionMap {
			destinationAttributionMap[id] = attributions
		}

	}

	return createHeuristicClusters(resultClusters), nil
}
