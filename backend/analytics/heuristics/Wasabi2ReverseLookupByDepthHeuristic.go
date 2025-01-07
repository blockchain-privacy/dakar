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

// wasabi2ReverseLookupByDepthHeuristic - see exec for description
type wasabi2ReverseLookupByDepthHeuristic struct {
	heuristicType string
	c             heuristics.Options
	depth         int
}

func newWasabi2ReverseLookupByDepthHeuristic() heuristic {
	return &wasabi2ReverseLookupByDepthHeuristic{heuristicType: heuristicTypeWasabi2ReverseLookupByDepth}
}

func (h *wasabi2ReverseLookupByDepthHeuristic) getType() string {
	return h.heuristicType
}

func (h *wasabi2ReverseLookupByDepthHeuristic) setConfig(c heuristics.Options) error {
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

func (h *wasabi2ReverseLookupByDepthHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *wasabi2ReverseLookupByDepthHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *wasabi2ReverseLookupByDepthHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Reverse lookup by depth",
		Type:        h.heuristicType,
		Category:    heuristicCategoryReverse,
		Description: "Traverses the transaction graph backwards until the given depth is reached and returns all found origins.",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			Description:  "Maximum traversal depth",
			Type:         "int",
		},
		AllowedParents: constants.TransactionTypesWasabi2,
	}
}

// wasabi2ReverseLookupByDepthHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *wasabi2ReverseLookupByDepthHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	return reverseLookupByTime(ctx, dgraph, g, parentHeuristicUID, 0, h.depth, h.c)
}
