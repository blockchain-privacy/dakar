// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"backend/analytics/classifier/dash"
	"backend/analytics/graph"
	"backend/constants"
	"backend/db"
	"backend/db/heuristics"
	"backend/external"
	"context"
	"fmt"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// reverseAmountHeuristic - see exec for description
type reverseAmountHeuristic struct {
	heuristicType string
	c             HeuristicOptions
}

func NewReverseAmountHeuristic() Heuristic {
	return &reverseAmountHeuristic{heuristicType: heuristicTypeReverseAmount}
}

func (h *reverseAmountHeuristic) GetType() string {
	return h.heuristicType
}

func (h *reverseAmountHeuristic) SetConfig(c HeuristicOptions) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.c = c

	return nil
}

func (h *reverseAmountHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *reverseAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *reverseAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:          "Reverse amount",
		Type:           h.heuristicType,
		Category:       heuristicCategoryReverse,
		Description:    "Returns all clusters which can fully fund the destination transaction via the denominations of their origins.",
		AllowedParents: []string{heuristicTypeReverseLookup, heuristicTypeOneSource, heuristicTypeDenominationType, heuristicTypePerfect},
	}
}

// Exec of the reverseAmountHeuristic applies the following heuristic:
// - filter all origins of sources, which do not create enough output denominations to fund the destination transaction
func (h *reverseAmountHeuristic) Exec(ctx context.Context, dgraph external.Database, _ *graph.Wrapper, parentUID string,
	parentResults []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	var results []heuristics.HeuristicTransaction
	var attributionMap map[heuristics.ClusterUID][]string
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
		// attributionMap maps a clusterUID to a slice of attribution UIDs
		results, attributionMap, err = heuristics.GetHeuristicTransactions(ctx, dgraph, parentUID,
			constants.TypeDashMixing)
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return nil, serror.New(errNoOriginsAtStart)
		}
	} else {
		results, err = getHeuristicTransactions(ctx, dgraph, parentResults, constants.TypeDashMixing)
		if err != nil {
			return nil, err
		}
	}

	transaction, err := heuristics.GetInputAmounts(ctx, dgraph, h.c.TransactionHash, constants.TypeDashMixing)
	if err != nil {
		return nil, err
	}

	// origins hold all origins found by the parent heuristic
	origins := make(map[string]heuristics.HeuristicTransaction, len(results))
	// Convert from slice to Hash
	for _, r := range results {
		origins[r.UID] = r
	}

	inputDenominationCounts := getDashDenominationCounts(transaction)
	originAmounts := buildDashSourceAmounts(origins)
	clusterTransactionMap := mapClusterToTransactions(results)
	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for clusterID, denominationSlice := range originAmounts {
		if containsDashDenomination(inputDenominationCounts, denominationSlice) {
			// add all transaction uids of a particular cluster to the return set
			for _, tx := range clusterTransactionMap[clusterID] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], db.UIDNode{UID: tx.UID})
			}
		}
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}

// containsDashDenomination returns true if all denominations with at
// least the same amount of denom1 are contained in denom2
func containsDashDenomination(denom1 [dash.NumDenominations]int, denom2 [dash.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] < d {
			return false
		}
	}
	return true
}
