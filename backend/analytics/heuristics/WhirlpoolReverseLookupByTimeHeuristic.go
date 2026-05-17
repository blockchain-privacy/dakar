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

// whirlpoolReverseLookupByTimeHeuristic - see exec for description
type whirlpoolReverseLookupByTimeHeuristic struct {
	heuristicType string
	c             HeuristicOptions
	lookBackTime  time.Duration
}

func NewWhirlpoolReverseLookupByTimeHeuristic() Heuristic {
	return &whirlpoolReverseLookupByTimeHeuristic{heuristicType: heuristicTypeWhirlpoolReverseLookupByTime}
}

func (h *whirlpoolReverseLookupByTimeHeuristic) GetType() string {
	return h.heuristicType
}

func (h *whirlpoolReverseLookupByTimeHeuristic) SetConfig(c HeuristicOptions) error {
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

func (h *whirlpoolReverseLookupByTimeHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *whirlpoolReverseLookupByTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *whirlpoolReverseLookupByTimeHeuristic) GetDescriptor() Descriptor {
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
		AllowedParents: constants.TransactionTypesWhirlpool,
	}
}

// Exec of the whirlpoolReverseLookupByTimeHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *whirlpoolReverseLookupByTimeHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return reverseLookup(ctx, dgraph, g, parentUID, h.lookBackTime, 0, h.c, constants.TypeWhirlpoolMixing)
}
