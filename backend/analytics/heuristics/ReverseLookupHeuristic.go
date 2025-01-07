package heuristics

import (
	"backend/analytics/graph"
	"backend/constants"
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
	heuristicType string
	c             heuristics.Options
	lookBackTime  time.Duration
}

func newReverseLookupHeuristic() heuristic {
	return &reverseLookupHeuristic{heuristicType: heuristicTypeReverseLookup}
}

func (h *reverseLookupHeuristic) getType() string {
	return h.heuristicType
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
		Title:       "Reverse lookup",
		Type:        h.heuristicType,
		Category:    heuristicCategoryReverse,
		Description: "Traverses the transaction graph backwards for the given duration and returns all found origins.",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			Description:  "Look back time in hours",
			Type:         "int",
		},
		AllowedParents: constants.TransactionTypesDash,
	}
}

// reverseLookupHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *reverseLookupHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	return reverseLookupByTime(ctx, dgraph, g, parentHeuristicUID, h.lookBackTime, 0, h.c, constants.TypeDashMixing)
}

func reverseLookupByTime(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string, lookBackTime time.Duration, depth int,
	options heuristics.Options, mixingTransactionType string) ([]heuristics.HeuristicCluster, error) {
	if lookBackTime == 0 && depth == 0 {
		return nil, nil
	}

	if lookBackTime != 0 && depth != 0 {
		return nil, serror.FromStr("both depth and look back time are set")
	}

	parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
	if err != nil {
		return nil, err
	}
	// heuristic is only allowed to be connected to a transaction
	if parentHeuristicSet {
		return nil, serror.New(errHeuristicNotValid)
	}

	inputTransactions, err := getInputTransactions(ctx, dgraph, options.TransactionHash, mixingTransactionType)
	if err != nil {
		return nil, err
	}

	if len(inputTransactions) == 0 {
		// nothing to do
		return nil, nil
	}

	var exclusions []string
	if options.ExcludeAddresses {
		exclusions, err = exclusion.GetAddressExclusionUIDs(ctx, dgraph, options.UserUID)
		if err != nil {
			return nil, err
		}
	}

	allTimeLimitedOrigins := make(map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	attributionMap := make(map[heuristics.ClusterUID][]string)
	for _, it := range inputTransactions {
		timeLimitedOrigins, usedAttributions, err := getTimeLimitedOrigins(ctx, dgraph, g, it,
			lookBackTime, depth, exclusions, options)
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
