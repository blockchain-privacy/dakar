package heuristics

import (
	"backend/analytics/graph"
	"backend/constants"
	"backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"strconv"
)

// forwardHeuristic - see exec for description
type wasabi2ForwardLookupByDepthHeuristic struct {
	heuristicType string
	c             heuristics.Options
	depth         int
}

func newwasabi2ForwardLookupByDepthHeuristic() heuristic {
	return &wasabi2ForwardLookupByDepthHeuristic{heuristicType: heuristicTypeWasabi2ForwardLookupByDepth}
}

func (h *wasabi2ForwardLookupByDepthHeuristic) getType() string {
	return h.heuristicType
}

func (h *wasabi2ForwardLookupByDepthHeuristic) setConfig(c heuristics.Options) error {
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

func (h *wasabi2ForwardLookupByDepthHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *wasabi2ForwardLookupByDepthHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *wasabi2ForwardLookupByDepthHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Forward lookup by depth",
		Type:        h.heuristicType,
		Category:    heuristicCategoryForward,
		Description: "Traverses the transaction graph forward until given depth is reached and collects destination transactions.",
		Parameter: &DescriptorParameter{
			DefaultValue: "2",
			MinimumValue: parameterMinDepth,
			MaximumValue: parameterMaxDepth,
			Description:  parameterDescriptionDepth,
			Type:         parameterTypeInt,
		},
		AllowedParents: constants.TransactionTypesWasabi2,
	}
}

// wasabi2ForwardLookupByDepthHeuristic finds all destination transactions connected the given
// transaction by traversing the mixing graph forward limited by depth.
func (h *wasabi2ForwardLookupByDepthHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	return forwardLookup(ctx, dgraph, g, parentHeuristicUID, 0, h.depth, h.c)
}
