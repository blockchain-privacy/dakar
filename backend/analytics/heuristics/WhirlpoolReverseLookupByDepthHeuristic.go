// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"context"
	"fmt"
	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db/heuristics"
	"gitlab.com/blockchain-privacy/dakar/external"
	"strconv"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// whirlpoolReverseLookupByDepthHeuristic - see exec for description
type whirlpoolReverseLookupByDepthHeuristic struct {
	heuristicType string
	c             HeuristicOptions
	depth         int
}

func NewWhirlpoolReverseLookupByDepthHeuristic() Heuristic {
	return &whirlpoolReverseLookupByDepthHeuristic{heuristicType: heuristicTypeWhirlpoolReverseLookupByDepth}
}

func (h *whirlpoolReverseLookupByDepthHeuristic) GetType() string {
	return h.heuristicType
}

func (h *whirlpoolReverseLookupByDepthHeuristic) SetConfig(c HeuristicOptions) error {
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

func (h *whirlpoolReverseLookupByDepthHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *whirlpoolReverseLookupByDepthHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *whirlpoolReverseLookupByDepthHeuristic) GetDescriptor() Descriptor {
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
		AllowedParents: constants.TransactionTypesWhirlpool,
	}
}

// Exec of the whirlpoolReverseLookupByDepthHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *whirlpoolReverseLookupByDepthHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return reverseLookup(ctx, dgraph, g, parentUID, 0, h.depth, h.c, constants.TypeWhirlpoolMixing)
}
