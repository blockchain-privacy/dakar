// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"backend/analytics/graph"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/db/heuristics"
	"backend/external"
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// reverseLookupHeuristic - see exec for description
type reverseLookupHeuristic struct {
	heuristicType string
	c             HeuristicOptions
	lookBackTime  time.Duration
}

func NewReverseLookupHeuristic() Heuristic {
	return &reverseLookupHeuristic{heuristicType: heuristicTypeReverseLookup}
}

func (h *reverseLookupHeuristic) GetType() string {
	return h.heuristicType
}

func (h *reverseLookupHeuristic) SetConfig(c HeuristicOptions) error {
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

func (h *reverseLookupHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *reverseLookupHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *reverseLookupHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Reverse lookup",
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
		AllowedParents: constants.TransactionTypesDash,
	}
}

// Exec of the reverseLookupHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *reverseLookupHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentUID string,
	_ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return reverseLookup(ctx, dgraph, g, parentUID, h.lookBackTime, 0, h.c, constants.TypeDashMixing)
}

func reverseLookup(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string, lookBackTime time.Duration, depth int,
	options HeuristicOptions, mixingTransactionType string) ([]heuristics.HeuristicCluster, error) {
	if lookBackTime == 0 && depth == 0 {
		return nil, nil
	}

	if lookBackTime != 0 && depth != 0 {
		return nil, serror.FromStr("both depth and look back time are set")
	}

	parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
	if err != nil {
		return nil, err
	}
	// heuristic is only allowed to be connected to a transaction
	if parentHeuristicSet {
		return nil, serror.New(errHeuristicNotValid)
	}

	inputTransactions, err := getInputTransactions(ctx, dgraph, options.TransactionHash, mixingTransactionType)
	if err != nil {
		return nil, err
	}

	if len(inputTransactions) == 0 {
		// nothing to do
		return nil, nil
	}

	attributions, err := attribution.GetAttributionsPerCluster(ctx, dgraph, options.UserUID, options.ClusterTypes)
	if err != nil {
		return nil, err
	}

	allTimeLimitedOrigins := make(map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	attributionMap := make(map[heuristics.ClusterUID][]string)
	for _, it := range inputTransactions {
		timeLimitedOrigins, usedAttributions, err := getTimeLimitedOrigins(ctx, dgraph, g, it.UID,
			lookBackTime, depth, attributions, options, mixingTransactionType)
		if err != nil {
			return nil, err
		}
		if timeLimitedOrigins == nil {
			continue
		}
		// save all origins only once
		for _, t := range timeLimitedOrigins {
			allTimeLimitedOrigins[t.UID] = t
		}

		// merge the attribution maps
		for id, usedAttribution := range usedAttributions {
			attributionMap[id] = usedAttribution
		}
	}

	// do custom peelchain handling for whirlpool origins
	if mixingTransactionType == constants.TypeWhirlpoolMixing {
		if err = mergeOriginPeelchainCluster(g, allTimeLimitedOrigins); err != nil {
			return nil, err
		}
	}

	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for k, v := range allTimeLimitedOrigins {
		resultClusters[v.Cluster] = append(resultClusters[v.Cluster], db.UIDNode{UID: k})
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
