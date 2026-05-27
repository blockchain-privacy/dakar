// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"context"
	"fmt"
	"strconv"

	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db/heuristics"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// wasabi2ReverseLookupByDepthHeuristic - see exec for description
type wasabi2ReverseLookupByDepthHeuristic struct {
	heuristicType string
	c             HeuristicOptions
	depth         int
}

func NewWasabi2ReverseLookupByDepthHeuristic() Heuristic {
	return &wasabi2ReverseLookupByDepthHeuristic{heuristicType: heuristicTypeWasabi2ReverseLookupByDepth}
}

func (h *wasabi2ReverseLookupByDepthHeuristic) GetType() string {
	return h.heuristicType
}

func (h *wasabi2ReverseLookupByDepthHeuristic) SetConfig(c HeuristicOptions) error {
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

func (h *wasabi2ReverseLookupByDepthHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *wasabi2ReverseLookupByDepthHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *wasabi2ReverseLookupByDepthHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Reverse lookup by depth",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Starting from each connected mixing transaction, traverses the transaction " +
			"graph backwards until the given depth is reached and returns all found origins.",
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

// Exec of the wasabi2ReverseLookupByDepthHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *wasabi2ReverseLookupByDepthHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return reverseLookup(ctx, dgraph, g, parentUID, 0, h.depth, h.c, constants.TypeWasabi2Mixing)
}
