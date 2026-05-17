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

// wasabi2ReverseLookupByTimeHeuristic - see exec for description
type wasabi2ReverseLookupByTimeHeuristic struct {
	heuristicType string
	c             HeuristicOptions
	lookBackTime  time.Duration
}

func NewWasabi2ReverseLookupByTimeHeuristic() Heuristic {
	return &wasabi2ReverseLookupByTimeHeuristic{heuristicType: heuristicTypeWasabi2ReverseLookupByTime}
}

func (h *wasabi2ReverseLookupByTimeHeuristic) GetType() string {
	return h.heuristicType
}

func (h *wasabi2ReverseLookupByTimeHeuristic) SetConfig(c HeuristicOptions) error {
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

func (h *wasabi2ReverseLookupByTimeHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *wasabi2ReverseLookupByTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *wasabi2ReverseLookupByTimeHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Reverse lookup by time",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Starting from each connected mixing transaction, traverses the transaction " +
			"graph backwards for the given duration and returns all found origins.",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			MinimumValue: parameterMinDuration,
			MaximumValue: parameterMaxDuration,
			Description:  parameterDescriptionLookBack,
			Type:         parameterTypeInt,
		},
		AllowedParents: constants.TransactionTypesWasabi2,
	}
}

// Exec of the wasabi2ReverseLookupByTimeHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *wasabi2ReverseLookupByTimeHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return reverseLookup(ctx, dgraph, g, parentUID, h.lookBackTime, 0, h.c, constants.TypeWasabi2Mixing)
}
