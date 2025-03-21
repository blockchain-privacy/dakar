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
)

// whirlpool2ReverseLookupByDepthHeuristic - see exec for description
type whirlpool2ReverseLookupByDepthHeuristic struct {
	heuristicType string
	c             heuristics.Options
	depth         int
}

func newWhirlpool2ReverseLookupByDepthHeuristic() heuristic {
	return &whirlpool2ReverseLookupByDepthHeuristic{heuristicType: heuristicTypeWhirlpoolReverseLookupByDepth}
}

func (h *whirlpool2ReverseLookupByDepthHeuristic) getType() string {
	return h.heuristicType
}

func (h *whirlpool2ReverseLookupByDepthHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	duration, err := strconv.Atoi(c.Parameter)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.depth = duration
	h.c = c

	return nil
}

func (h *whirlpool2ReverseLookupByDepthHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *whirlpool2ReverseLookupByDepthHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *whirlpool2ReverseLookupByDepthHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Reverse lookup by depth",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Starting from each connected mixing transaciton, traverses the transaction " +
			"graph backwards until the given depth is reached and returns all found origins.",
		Parameter: &DescriptorParameter{
			DefaultValue: "2",
			MinimumValue: parameterMinDepth,
			MaximumValue: parameterMaxDepth,
			Description:  parameterDescriptionDepth,
			Type:         parameterTypeInt,
		},
		AllowedParents: constants.TransactionTypesWhirlpool,
	}
}

// whirlpool2ReverseLookupByDepthHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *whirlpool2ReverseLookupByDepthHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	return reverseLookup(ctx, dgraph, g, parentHeuristicUID, 0, h.depth, h.c, constants.TypeWhirlpoolMixing)
}
