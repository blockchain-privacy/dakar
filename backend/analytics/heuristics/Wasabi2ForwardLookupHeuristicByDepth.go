// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"backend/analytics/graph"
	"backend/constants"
	"backend/db/heuristics"
	"backend/external"
	"context"
	"fmt"
	"strconv"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// forwardHeuristic - see exec for description
type wasabi2ForwardLookupByDepthHeuristic struct {
	heuristicType string
	c             HeuristicOptions
	depth         int
}

func NewWasabi2ForwardLookupByDepthHeuristic() Heuristic {
	return &wasabi2ForwardLookupByDepthHeuristic{heuristicType: heuristicTypeWasabi2ForwardLookupByDepth}
}

func (h *wasabi2ForwardLookupByDepthHeuristic) GetType() string {
	return h.heuristicType
}

func (h *wasabi2ForwardLookupByDepthHeuristic) SetConfig(c HeuristicOptions) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	d, err := strconv.Atoi(c.Parameter)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.depth = d
	h.c = c

	return nil
}

func (h *wasabi2ForwardLookupByDepthHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *wasabi2ForwardLookupByDepthHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
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

// Exec of the wasabi2ForwardLookupByDepthHeuristic finds all destination transactions connected the given
// transaction by traversing the mixing graph forward limited by depth.
func (h *wasabi2ForwardLookupByDepthHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return forwardLookup(ctx, dgraph, g, parentUID, 0, h.depth, h.c)
}
