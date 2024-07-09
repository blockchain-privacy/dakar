package heuristics

import (
	"backend/analytics/graph"
	"backend/db/analytics/exclusion"
	"backend/db/analytics/heuristics"
	"backend/external"
	"github.com/qrest/gomisc/serror"

	"fmt"
	"strconv"
	"time"
)

// reverseLookupHeuristic - see exec for description
type reverseLookupHeuristic struct {
	heuristicType        string
	parameterDescription string
	c                    heuristics.Config
	lookBackTime         time.Duration
}

func newReverseLookupHeuristic() heuristic {
	return &reverseLookupHeuristic{heuristicType: "reverse_lookup"}
}

func (h *reverseLookupHeuristic) getType() string {
	return h.heuristicType
}

func (h *reverseLookupHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *reverseLookupHeuristic) setConfig(c heuristics.Config) error {
	duration, err := strconv.ParseUint(c.Parameter, 10, 32)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.lookBackTime = time.Duration(duration) * time.Hour
	h.parameterDescription = strconv.FormatUint(duration, 10)
	h.c = c

	return nil
}

func (h *reverseLookupHeuristic) getConfig() heuristics.Config {
	return h.c
}

func (h *reverseLookupHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *reverseLookupHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Reverse Lookup",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Performs a reverse lookup for the given duration and returns " +
			"all found origins. If this heuristic has a parent heuristic, only origins " +
			"which also occur in the parent heuristic will be returned. ",
		Parameter: &struct {
			DefaultValue string `json:"value,omitempty"`
			Description  string `json:"description,omitempty"`
			Type         string `json:"type,omitempty"`
		}{
			DefaultValue: "48",
			Description:  "Look back time in hours",
			Type:         "int",
		},
	}
}

// reverseLookupHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *reverseLookupHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	// holds all origins from either the parent heuristic or the associated destination transaction
	originLimit := make(map[string]bool)
	// parentAttributionMap maps a clusterUID to a slice of attribution UIDs
	var parentAttributionMap map[heuristics.ClusterUID][]string
	parentHeuristicSet := isParentHeuristicSet(parentHeuristicUID)
	if parentHeuristicSet {
		// get origins from parent heuristic
		parentHeuristicResults, attrMap, err := heuristics.GetHeuristicResults(dgraph, parentHeuristicUID)
		if err != nil {
			return nil, err
		}

		if parentHeuristicResults == nil {
			return nil, serror.New(errNoOriginsAtStart)
		}

		parentAttributionMap = attrMap
		for _, hr := range parentHeuristicResults {
			originLimit[hr.UID] = true
		}
	}

	// gather input information
	inputTransactions, err := heuristics.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, err
	}

	var exclusions []string
	if h.c.ExcludeAddresses {
		exclusions, err = exclusion.GetAddressExclusionUIDs(dgraph, h.c.UserUID)
		if err != nil {
			return nil, err
		}
	}

	allTimeLimitedOrigins := make(map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	attributionMap := make(map[heuristics.ClusterUID][]string)
	for _, it := range inputTransactions {
		timeLimitedOrigins, usedAttributions, err := getTimeLimitedOrigins(dgraph, g, it,
			h.lookBackTime, exclusions, h.c)
		if err != nil {
			return nil, err
		}
		if timeLimitedOrigins == nil {
			continue
		}
		// save all origins only once
		for _, t := range timeLimitedOrigins {
			// only save if the uid also exists in the parent origin set
			if parentHeuristicSet && !originLimit[t.UID] {
				continue
			}
			allTimeLimitedOrigins[t.UID] = t
		}

		if parentHeuristicSet {
			// no need to merge the attributions if they are not used
			continue
		}

		// merge the attribution maps
		for id, attributions := range usedAttributions {
			attributionMap[id] = attributions
		}
	}

	resultClusters := make(map[heuristics.ClusterUID][]heuristics.HeuristicResult)
	for k, v := range allTimeLimitedOrigins {
		resultClusters[v.Cluster] = append(resultClusters[v.Cluster], heuristics.HeuristicResult{
			Origin: heuristics.DummyNode{UID: k},
		})
	}

	if parentHeuristicSet {
		return createHeuristicClusters(resultClusters, parentAttributionMap), nil
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
