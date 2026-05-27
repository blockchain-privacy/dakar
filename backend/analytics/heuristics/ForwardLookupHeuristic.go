// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"context"
	"fmt"
	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/heuristics"
	"gitlab.com/blockchain-privacy/dakar/external"
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// forwardHeuristic - see exec for description
type forwardHeuristic struct {
	heuristicType   string
	c               HeuristicOptions
	lookForwardTime time.Duration
}

func NewForwardLookupHeuristic() Heuristic {
	return &forwardHeuristic{heuristicType: heuristicTypeForwardLookup}
}

func (h *forwardHeuristic) GetType() string {
	return h.heuristicType
}

func (h *forwardHeuristic) SetConfig(c HeuristicOptions) error {
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

func (h *forwardHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *forwardHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *forwardHeuristic) GetDescriptor() Descriptor {
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
		AllowedParents: constants.TransactionTypesDash,
	}
}

// Exec of the forwardLookupHeuristic finds all destination transactions connected the given
// transaction by traversing the mixing graph forward limited by time.
func (h *forwardHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return forwardLookup(ctx, dgraph, g, parentUID, h.lookForwardTime, 0, h.c)
}

func forwardLookup(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentHeuristicUID string,
	lookForwardTime time.Duration, depth int, options HeuristicOptions) ([]heuristics.HeuristicCluster, error) {
	if lookForwardTime == 0 && depth == 0 {
		return nil, nil
	}

	parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
	if err != nil {
		return nil, err
	}
	// heuristic is only allowed to be connected to a transaction
	if parentHeuristicSet {
		return nil, serror.New(errHeuristicNotValid)
	}

	uid, err := db.GetTransactionUID(ctx, dgraph, options.TransactionHash)
	if err != nil {
		return nil, err
	}

	destinations, _, err := getTimeLimitedDestinations(ctx, dgraph, g, uid,
		lookForwardTime, depth, nil, options, constants.TypeDashMixing)
	if err != nil {
		return nil, err
	}

	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for _, dst := range destinations {
		resultClusters[dst.Cluster] = append(resultClusters[dst.Cluster], db.UIDNode{UID: dst.UID})
	}

	return createHeuristicClusters(resultClusters, nil), nil
}
