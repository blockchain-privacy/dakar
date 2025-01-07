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

// wasabi2OneSourceByDepthHeuristic - see exec for description
type wasabi2OneSourceByDepthHeuristic struct {
	heuristicType string
	depth         int
	c             heuristics.Options
}

func newWasabi2OneSourceByDepthHeuristic() heuristic {
	return &wasabi2OneSourceByDepthHeuristic{heuristicType: heuristicTypeWasabi2OneSourceByDepth}
}

func (h *wasabi2OneSourceByDepthHeuristic) getType() string {
	return h.heuristicType
}

func (h *wasabi2OneSourceByDepthHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	depth, err := strconv.Atoi(c.Parameter)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.depth = depth
	h.c = c

	return nil
}

func (h *wasabi2OneSourceByDepthHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *wasabi2OneSourceByDepthHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *wasabi2OneSourceByDepthHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "One source by depth",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Destination transactions spend outputs of their connected input mixing transactions. " +
			"Each input mixing transaction is connected to a mixing sub graph. " +
			"This heuristic excludes all clusters which can't fund every mixing sub " +
			"graph (due to lack of funds or du to having not connection to them).",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			Description:  "Depth",
			Type:         "int",
		},
		AllowedParents: constants.TransactionTypesWasabi2,
	}
}

// wasabi2OneSourceByDepthHeuristic applies the following heuristics:
//   - filter all origins, which are not created in the time span defined by depth
//   - filter all origins of clusters, which do not have enough denominations to fund all of their respective
//     outputs of input transaction which are used as inputs in the destination transaction
//   - filter all origins of clusters, which do not occur in all sets of input transaction origins
func (h *wasabi2OneSourceByDepthHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentHeuristicUID string) (
	[]heuristics.HeuristicCluster, error) {
	return oneSourceByTime(ctx, dgraph, g, parentHeuristicUID, 0, h.depth, h.c)
}
