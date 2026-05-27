// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/cmd/cliutil"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/heuristics"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// forwardAmountHeuristic - see exec for description
type forwardAmountHeuristic struct {
	c               HeuristicOptions
	heuristicType   string
	lookForwardTime time.Duration
}

func NewForwardAmountHeuristic() Heuristic {
	return &forwardAmountHeuristic{heuristicType: heuristicTypeForwardAmount}
}

func (h *forwardAmountHeuristic) GetType() string {
	return h.heuristicType
}

func (h *forwardAmountHeuristic) SetConfig(c HeuristicOptions) error {
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

func (h *forwardAmountHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *forwardAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *forwardAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Forward amount",
		Type:        h.heuristicType,
		Category:    heuristicCategoryForward,
		Description: "Returns all destination transactions which can be fully funded by the origins of their source.",
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

// Exec of the forwardAmountHeuristic applies the following heuristic:
// - filters all destinations which can not be funded by the sources based on the denominations of the source
func (h *forwardAmountHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	if h.lookForwardTime == 0 {
		return nil, nil
	}

	parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentUID)
	if err != nil {
		return nil, err
	}
	// heuristic is only allowed to be connected to a transaction
	if parentHeuristicSet {
		return nil, serror.New(errHeuristicNotValid)
	}

	uid, err := db.GetTransactionUID(ctx, dgraph, h.c.TransactionHash)
	if err != nil {
		return nil, err
	}

	// attributionMap maps a clusterUID to a slice of attribution UIDs
	results, attributionMap, err := heuristics.GetTransactionsWithOutputAmountAndCluster(ctx, dgraph,
		[]string{uid}, h.c.UserUID, h.c.ClusterTypes, nil, constants.TypeDashMixing)
	if err != nil {
		return nil, err
	}

	if len(results) > 1 {
		return nil, serror.FromStr("received more than one transaction")
	}
	if len(results) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
	}

	destinationUIDsMap, err := g.ForwardLookup(uid, h.lookForwardTime, 0, h.c.ExcludeSpendingGaps)
	if err != nil {
		return nil, err
	}

	// get tx details for each uid
	destinations, err := heuristics.GetTransactionsWithInputAmount(ctx, dgraph, cliutil.GetMapKeys(destinationUIDsMap))
	if err != nil {
		return nil, err
	}

	originOutputDenominations := buildDashSourceAmounts(map[string]heuristics.HeuristicTransaction{uid: results[0]})
	var clusterFilteredDestinations []db.UIDNode
	for _, tx := range destinations {
		// check if the denominations of the destination transactions can be funded by the denomination of its cluster
		if containsDashDenomination(getDashDenominationCounts(tx), originOutputDenominations[results[0].Cluster]) {
			clusterFilteredDestinations = append(clusterFilteredDestinations, db.UIDNode{UID: tx.UID})
		}
	}

	resultClusters := map[heuristics.ClusterUID][]db.UIDNode{}
	if len(clusterFilteredDestinations) > 0 {
		resultClusters[results[0].Cluster] = clusterFilteredDestinations
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
