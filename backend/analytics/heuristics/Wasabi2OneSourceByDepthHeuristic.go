// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"backend/analytics/graph"
	"backend/constants"
	"backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"fmt"
	"strconv"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// wasabi2OneSourceByDepthHeuristic - see exec for description
type wasabi2OneSourceByDepthHeuristic struct {
	heuristicType string
	depth         int
	c             heuristics.Options
}

func NewWasabi2OneSourceByDepthHeuristic() Heuristic {
	return &wasabi2OneSourceByDepthHeuristic{heuristicType: heuristicTypeWasabi2OneSourceByDepth}
}

func (h *wasabi2OneSourceByDepthHeuristic) GetType() string {
	return h.heuristicType
}

func (h *wasabi2OneSourceByDepthHeuristic) SetConfig(c heuristics.Options) error {
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

func (h *wasabi2OneSourceByDepthHeuristic) GetConfig() heuristics.Options {
	return h.c
}

func (h *wasabi2OneSourceByDepthHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *wasabi2OneSourceByDepthHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "One source by depth",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Destination transactions spend outputs of their connected input mixing transactions. " +
			"Each input mixing transaction is connected to a mixing sub graph. Starting from each connected " +
			"mixing transaction, this heuristic traverses the transaction graph backwards for the given " +
			"duration and excludes all clusters which can't fund every mixing sub graph (due to lack of funds " +
			"or due to having no connection to them).",
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

// Exec of the wasabi2OneSourceByDepthHeuristic applies the following heuristics:
//   - filter all origins, which are not created in the time span defined by depth
//   - filter all origins of clusters, which do not have enough denominations to fund all of their respective
//     outputs of input transaction which are used as inputs in the destination transaction
//   - filter all origins of clusters, which do not occur in all sets of input transaction origins
func (h *wasabi2OneSourceByDepthHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return wasabi2OneSource(ctx, dgraph, g, parentUID, 0, h.depth, h.c)
}
