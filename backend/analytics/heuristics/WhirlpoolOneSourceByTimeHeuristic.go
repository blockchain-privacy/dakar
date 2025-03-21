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
	"time"
)

// whirlpoolOneSourceByTimeHeuristic - see exec for description
type whirlpoolOneSourceByTimeHeuristic struct {
	heuristicType string
	lookBackTime  time.Duration
	c             heuristics.Options
}

func newWhirlpoolOneSourceByTimeHeuristic() heuristic {
	return &whirlpoolOneSourceByTimeHeuristic{heuristicType: heuristicTypeWhirlpoolOneSourceByTime}
}

func (h *whirlpoolOneSourceByTimeHeuristic) getType() string {
	return h.heuristicType
}

func (h *whirlpoolOneSourceByTimeHeuristic) setConfig(c heuristics.Options) error {
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

func (h *whirlpoolOneSourceByTimeHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *whirlpoolOneSourceByTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *whirlpoolOneSourceByTimeHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "One source by time",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Destination transactions spend outputs of their connected input mixing transactions. " +
			"Each input mixing transaction is connected to a mixing sub graph. Starting from each connected " +
			"mixing transacion, this heuristic traverses the transaction graph backwards until the given depth " +
			"is reached and excludes all clusters which can't fund every mixing sub graph (due to lack of funds " +
			"or due to having no connection to them).",
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

// whirlpoolOneSourceByTimeHeuristic applies the following heuristics:
//   - filter all origins, which are not created in the time span defined by lookBackTime
//   - filter all origins of clusters, which do not have enough denominations to fund all of their respective
//     outputs of input transaction which are used as inputs in the destination transaction
//   - filter all origins of clusters, which do not occur in all sets of input transaction origins
func (h *whirlpoolOneSourceByTimeHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentHeuristicUID string) (
	[]heuristics.HeuristicCluster, error) {
	return oneSource(ctx, dgraph, g, parentHeuristicUID, h.lookBackTime, 0, h.c)
}
