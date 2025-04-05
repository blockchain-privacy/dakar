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

// whirlpoolOneSourceByDepthHeuristic - see exec for description
type whirlpoolOneSourceByDepthHeuristic struct {
	heuristicType string
	depth         int
	c             heuristics.Options
}

func NewWhirlpoolOneSourceByDepthHeuristic() Heuristic {
	return &whirlpoolOneSourceByDepthHeuristic{heuristicType: heuristicTypeWhirlpoolOneSourceByDepth}
}

func (h *whirlpoolOneSourceByDepthHeuristic) GetType() string {
	return h.heuristicType
}

func (h *whirlpoolOneSourceByDepthHeuristic) SetConfig(c heuristics.Options) error {
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

func (h *whirlpoolOneSourceByDepthHeuristic) GetConfig() heuristics.Options {
	return h.c
}

func (h *whirlpoolOneSourceByDepthHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *whirlpoolOneSourceByDepthHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "One source by depth",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Destination transactions spend outputs of their connected input mixing transactions. " +
			"Each input mixing transaction is connected to a mixing sub graph. Starting from each connected " +
			"mixing transacion, this heuristic traverses the transaction graph backwards for the given " +
			"duration and excludes all clusters which can't fund every mixing sub graph (due to lack of funds " +
			"or due to having no connection to them).",
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

// Exec of the whirlpoolOneSourceByDepthHeuristic applies the following heuristics:
//   - filter all origins, which are not created in the time span defined by depth
//   - filter all origins of clusters, which do not have enough denominations to fund all of their respective
//     outputs of input transaction which are used as inputs in the destination transaction
//   - filter all origins of clusters, which do not occur in all sets of input transaction origins
func (h *whirlpoolOneSourceByDepthHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentHeuristicUID string) (
	[]heuristics.HeuristicCluster, error) {
	return whirlpoolOnceSource(ctx, dgraph, g, parentHeuristicUID, 0, h.depth, h.c)
}
