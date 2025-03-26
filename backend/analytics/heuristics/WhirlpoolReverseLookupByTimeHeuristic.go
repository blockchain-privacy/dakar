package heuristics

import (
	"backend/analytics/graph"
	"backend/constants"
	"backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"github.com/qrest/gomisc/serror"

	"fmt"
	"strconv"
	"time"
)

// whirlpoolReverseLookupByTimeHeuristic - see exec for description
type whirlpoolReverseLookupByTimeHeuristic struct {
	heuristicType string
	c             heuristics.Options
	lookBackTime  time.Duration
}

func newWhirlpoolReverseLookupByTimeHeuristic() heuristic {
	return &whirlpoolReverseLookupByTimeHeuristic{heuristicType: heuristicTypeWhirlpoolReverseLookupByTime}
}

func (h *whirlpoolReverseLookupByTimeHeuristic) getType() string {
	return h.heuristicType
}

func (h *whirlpoolReverseLookupByTimeHeuristic) setConfig(c heuristics.Options) error {
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

func (h *whirlpoolReverseLookupByTimeHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *whirlpoolReverseLookupByTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *whirlpoolReverseLookupByTimeHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Reverse lookup by time",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Starting from each connected mixing transaciton, traverses the transaction " +
			"graph backwards for the given duration and returns all found origins.",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			MinimumValue: parameterMinDuration,
			MaximumValue: parameterMaxDuration,
			Description:  parameterDescriptionLookBack,
			Type:         parameterTypeInt,
		},
		AllowedParents: constants.TransactionTypesWhirlpool,
	}
}

// whirlpoolReverseLookupByTimeHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *whirlpoolReverseLookupByTimeHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	return reverseLookup(ctx, dgraph, g, parentHeuristicUID, h.lookBackTime, 0, h.c, constants.TypeWhirlpoolMixing)
}
