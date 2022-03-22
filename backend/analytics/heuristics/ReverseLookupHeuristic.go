package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db/analytics/clustering"
	"backend/db/analytics/exclusion"
	"backend/db/analytics/heuristics"
	"backend/external"

	"fmt"
	"strconv"
	"time"
)

// reverseLookupHeuristic - see exec for description
type reverseLookupHeuristic struct {
	heuristicType        string
	parameterDescription string
	userUID              string
	excludeAddresses     bool
	lookBackTime         time.Duration
	clusterTypes         []clustering.ClusterType
}

// newReverseLookupHeuristic constructs a reverseLookupHeuristic. hoursToLookBack in hours.
func newReverseLookupHeuristic(hoursToLookBack uint32,
	clusteringMethod []clustering.ClusterType) *reverseLookupHeuristic {
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	return &reverseLookupHeuristic{
		heuristicType:        "reverse_lookup",
		lookBackTime:         lBackTime,
		parameterDescription: lBackTime.String(),
		clusterTypes:         clusteringMethod,
	}
}

func (h reverseLookupHeuristic) getType() string {
	return h.heuristicType
}

func (h reverseLookupHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h reverseLookupHeuristic) hasParameter() bool {
	return true
}

func (h *reverseLookupHeuristic) setParameter(p string) error {
	hoursToLookBack, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	h.lookBackTime = lBackTime
	h.parameterDescription = strconv.FormatUint(hoursToLookBack, 10)
	return nil
}

// setClusterTypes sets additional cluster types, which are used to execute the heuristic.
// Multi-input clusters are always used to execute the heuristic,
// any cluster type set here will be used additionally. If at least one cluster type is set,
// then the consolidation of the multi-input clusters and the additional clusters will be used.
func (h *reverseLookupHeuristic) setClusterTypes(clusterTypes []clustering.ClusterType) error {
	if !areClusterTypesValid(clusterTypes) {
		return errInvalidClusterTypes
	}

	h.clusterTypes = clusterTypes
	return nil
}

// getClusterTypes returns the cluster types this heuristic uses to cluster addresses
func (h *reverseLookupHeuristic) getClusterTypes() []clustering.ClusterType {
	return h.clusterTypes
}

// setExcludeAddresses sets whether certain addresses should be excluded from the lookups
func (h *reverseLookupHeuristic) setExcludeAddresses(excludeAddresses bool) {
	h.excludeAddresses = excludeAddresses
}

// getExcludeAddresses returns whether certain addresses should be excluded from the lookups
func (h *reverseLookupHeuristic) getExcludeAddresses() bool {
	return h.excludeAddresses
}

// setUserUID sets the UID of the user who created this heuristic
func (h *reverseLookupHeuristic) setUserUID(uid string) {
	h.userUID = uid
}

func (h reverseLookupHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h reverseLookupHeuristic) GetDescriptor() Descriptor {
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

func (h reverseLookupHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// reverseLookupHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h reverseLookupHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
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
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if parentHeuristicResults == nil {
			return nil, errNoOriginsAtStart
		}

		parentAttributionMap = attrMap
		for _, hr := range parentHeuristicResults {
			originLimit[hr.UID] = true
		}
	}

	// gather input information
	inputTransactions, err := heuristics.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	var exclusions []string
	if h.excludeAddresses {
		exclusions, err = exclusion.GetAddressExclusionUIDs(dgraph, h.userUID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
	}

	allTimeLimitedOrigins := make(map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	attributionMap := make(map[heuristics.ClusterUID][]string)
	for _, it := range inputTransactions {
		timeLimitedOrigins, usedAttributions, err := getTimeLimitedOrigins(dgraph, g, it, h.lookBackTime, h.userUID,
			h.clusterTypes, exclusions)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
	} else {
		return createHeuristicClusters(resultClusters, attributionMap), nil
	}
}
