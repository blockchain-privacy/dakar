package heuristics

import (
	"backend/analytics/graph"
	"backend/db"
	"backend/db/analytics/exclusion"
	"backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"github.com/qrest/gomisc/serror"

	"fmt"
	"strconv"
	"time"
)

// reverseLookupHeuristic - see exec for description
type reverseLookupHeuristic struct {
	heuristicType        string
	parameterDescription string
	c                    heuristics.Options
	lookBackTime         time.Duration
}

func newReverseLookupHeuristic() heuristic {
	return &reverseLookupHeuristic{heuristicType: heuristicTypeReverseLookup}
}

func (h *reverseLookupHeuristic) getType() string {
	return h.heuristicType
}

func (h *reverseLookupHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *reverseLookupHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	duration, err := strconv.ParseInt(c.Parameter, 10, 64)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.lookBackTime = time.Duration(duration) * time.Hour
	h.parameterDescription = strconv.FormatInt(duration, 10)
	h.c = c

	return nil
}

func (h *reverseLookupHeuristic) getConfig() heuristics.Options {
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
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			Description:  "Look back time in hours",
			Type:         "int",
		},
		AllowedParents: []string{parentTypeTransaction},
	}
}

// reverseLookupHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *reverseLookupHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	if h.lookBackTime == 0 {
		return nil, nil
	}

	parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
	if err != nil {
		return nil, err
	}
	// heuristic is only allowed to be connected to a transaction
	if parentHeuristicSet {
		return nil, serror.New(errHeuristicNotValid)
	}

	inputTransactions, err := getInputTransactions(ctx, dgraph, h.c.TransactionHash)
	if err != nil {
		return nil, err
	}

	var exclusions []string
	if h.c.ExcludeAddresses {
		exclusions, err = exclusion.GetAddressExclusionUIDs(ctx, dgraph, h.c.UserUID)
		if err != nil {
			return nil, err
		}
	}

	allTimeLimitedOrigins := make(map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	attributionMap := make(map[heuristics.ClusterUID][]string)
	for _, it := range inputTransactions {
		timeLimitedOrigins, usedAttributions, err := getTimeLimitedOrigins(ctx, dgraph, g, it,
			h.lookBackTime, exclusions, h.c)
		if err != nil {
			return nil, err
		}
		if timeLimitedOrigins == nil {
			continue
		}
		// save all origins only once
		for _, t := range timeLimitedOrigins {
			allTimeLimitedOrigins[t.UID] = t
		}

		// merge the attribution maps
		for id, attributions := range usedAttributions {
			attributionMap[id] = attributions
		}
	}

	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for k, v := range allTimeLimitedOrigins {
		resultClusters[v.Cluster] = append(resultClusters[v.Cluster], db.UIDNode{UID: k})
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
