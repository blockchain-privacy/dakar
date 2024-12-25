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

// wasabi2ReverseLookupByTimeHeuristic - see exec for description
type wasabi2ReverseLookupByTimeHeuristic struct {
	heuristicType string
	c             heuristics.Options
	lookBackTime  time.Duration
}

func newWasabi2ReverseLookupByTimeHeuristic() heuristic {
	return &wasabi2ReverseLookupByTimeHeuristic{heuristicType: heuristicTypeWasabi2ReverseLookupByTime}
}

func (h *wasabi2ReverseLookupByTimeHeuristic) getType() string {
	return h.heuristicType
}

func (h *wasabi2ReverseLookupByTimeHeuristic) setConfig(c heuristics.Options) error {
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

func (h *wasabi2ReverseLookupByTimeHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *wasabi2ReverseLookupByTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *wasabi2ReverseLookupByTimeHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Reverse lookup by time",
		Type:        h.heuristicType,
		Category:    heuristicCategoryReverse,
		Description: "Traverses the transaction graph backwards for the given duration and returns all found origins.",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			Description:  "Look back time in hours",
			Type:         "int",
		},
		AllowedParents: constants.TransactionTypesWasabi2,
	}
}

// wasabi2ReverseLookupByTimeHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *wasabi2ReverseLookupByTimeHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	return reverseLookupByTime(ctx, dgraph, g, parentHeuristicUID, h.lookBackTime, 0, h.c)
}
