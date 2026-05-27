// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
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
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// wasabi2ReverseAmountHeuristic - see exec for description
type wasabi2ReverseAmountHeuristic struct {
	heuristicType string
	c             HeuristicOptions
}

func NewWasabi2ReverseAmountHeuristic() Heuristic {
	return &wasabi2ReverseAmountHeuristic{heuristicType: heuristicTypeWasabi2ReverseAmount}
}

func (h *wasabi2ReverseAmountHeuristic) GetType() string {
	return h.heuristicType
}

func (h *wasabi2ReverseAmountHeuristic) SetConfig(c HeuristicOptions) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.c = c

	return nil
}

func (h *wasabi2ReverseAmountHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *wasabi2ReverseAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *wasabi2ReverseAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Reverse amount",
		Type:        h.heuristicType,
		Category:    heuristicCategoryReverse,
		Description: "Returns all clusters which can fully fund the destination transaction via their origins.",
		AllowedParents: []string{heuristicTypeWasabi2ReverseLookupByDepth, heuristicTypeWasabi2ReverseLookupByTime,
			heuristicTypeWasabi2OneSourceByTime, heuristicTypeWasabi2OneSourceByDepth},
	}
}

// Exec of the wasabi2ReverseAmountHeuristic applies the following heuristic:
// - filter all origins of sources, which do not create enough output denominations to fund the destination transaction
func (h *wasabi2ReverseAmountHeuristic) Exec(ctx context.Context, dgraph external.Database, _ *graph.Wrapper,
	parentUID string, parentResults []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	var results []heuristics.HeuristicTransaction
	var err error
	if parentResults == nil {
		parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentUID)
		if err != nil {
			return nil, err
		}
		// heuristic is only allowed to be connected to another heuristic
		if !parentHeuristicSet {
			return nil, serror.New(errHeuristicNotValid)
		}

		// get origins from parent heuristic
		results, err = heuristics.GetHeuristicTransactions(ctx, dgraph, parentUID,
			constants.TypeWasabi2Mixing)
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return nil, serror.New(errNoOriginsAtStart)
		}
	} else {
		results, err = getHeuristicTransactions(ctx, dgraph, parentResults, constants.TypeWasabi2Mixing)
		if err != nil {
			return nil, err
		}
	}

	transaction, err := heuristics.GetInputAmounts(ctx, dgraph, h.c.TransactionHash, constants.TypeWasabi2Mixing)
	if err != nil {
		return nil, err
	}

	// the amount of mixed funds consumed by the destination transaction
	var destinationInputSum int64
	for _, input := range transaction.Outputs {
		destinationInputSum += input.Amount
	}

	// sanity check
	if destinationInputSum == 0 {
		return nil, serror.FromStrWithContext("destination transaction does not spend any mixed funds",
			"transaction", h.c.TransactionHash)
	}

	clusterTransactionMap := mapClusterToTransactions(results)
	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for _, clusterOrigins := range clusterTransactionMap {
		var clusterOutputAmount int64
		for _, origin := range clusterOrigins {
			for _, output := range origin.Outputs {
				clusterOutputAmount += output.Amount
			}
		}

		if clusterOutputAmount >= destinationInputSum {
			for _, origin := range clusterOrigins {
				resultClusters[origin.Cluster] = append(resultClusters[origin.Cluster], db.UIDNode{UID: origin.UID})
			}
		}
	}

	return createHeuristicClusters(resultClusters), nil
}
