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
	"time"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// forwardHeuristic - see exec for description
type wasabi2ForwardLookupByTimeHeuristic struct {
	heuristicType   string
	c               HeuristicOptions
	lookForwardTime time.Duration
}

func NewWasabi2ForwardLookupByTimeHeuristic() Heuristic {
	return &wasabi2ForwardLookupByTimeHeuristic{heuristicType: heuristicTypeWasabi2ForwardLookupByTime}
}

func (h *wasabi2ForwardLookupByTimeHeuristic) GetType() string {
	return h.heuristicType
}

func (h *wasabi2ForwardLookupByTimeHeuristic) SetConfig(c HeuristicOptions) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	hoursToLookForward, err := strconv.ParseInt(c.Parameter, 10, 64)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.c = c

	return nil
}

func (h *wasabi2ForwardLookupByTimeHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *wasabi2ForwardLookupByTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *wasabi2ForwardLookupByTimeHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Forward lookup by time",
		Type:        h.heuristicType,
		Category:    heuristicCategoryForward,
		Description: "Traverses the transaction graph forward for the given duration and collects destination transactions.",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			MinimumValue: parameterMinDuration,
			MaximumValue: parameterMaxDuration,
			Description:  parameterDescriptionLookForward,
			Type:         parameterTypeInt,
		},
		AllowedParents: constants.TransactionTypesWasabi2,
	}
}

// Exec of the wasabi2ForwardLookupByTimeHeuristic finds all destination transactions connected the given
// transaction by traversing the mixing graph forward limited by time.
func (h *wasabi2ForwardLookupByTimeHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentUID string,
	_ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return forwardLookup(ctx, dgraph, g, parentUID, h.lookForwardTime, 0, h.c)
}
