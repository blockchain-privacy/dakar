// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"context"
	"fmt"

	"gitlab.com/blockchain-privacy/dakar/analytics/classifier/dash"
	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/heuristics"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// perfectMatchHeuristic - see exec for description
type perfectMatchHeuristic struct {
	heuristicType string
	c             HeuristicOptions
}

func NewPerfectMatchHeuristic() Heuristic {
	return &perfectMatchHeuristic{heuristicType: heuristicTypePerfect}
}

func (h *perfectMatchHeuristic) GetType() string {
	return h.heuristicType
}

func (h *perfectMatchHeuristic) SetConfig(c HeuristicOptions) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.c = c

	return nil
}

func (h *perfectMatchHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *perfectMatchHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Perfect match",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Excludes all origins of clusters, which have denominations without " +
			"a perfect match for the denominations of the destination transaction.",
		AllowedParents: []string{heuristicTypeReverseLookup, heuristicTypeOneSource, heuristicTypeReverseAmount, heuristicTypeDenominationType},
	}
}

func (h *perfectMatchHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

// Exec of the perfectMatchHeuristic applies the following heuristic:
//   - filter all origins of sources, which have denominations without a perfect match for the
//     denominations of the destination transaction
func (h *perfectMatchHeuristic) Exec(ctx context.Context, dgraph external.Database, _ *graph.Wrapper, parentUID string,
	_ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
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
	results, attributionMap, err := heuristics.GetHeuristicTransactions(ctx, dgraph, parentUID,
		constants.TypeDashMixing)
	if err != nil {
		return nil, err
	}

	clusterTransactionMap := mapClusterToTransactions(results)

	// origins hold all origins found by the parent heuristic
	origins := make(map[string]heuristics.HeuristicTransaction, len(results))
	// Convert from slice to Hash
	for _, r := range results {
		origins[r.UID] = r
	}

	if len(origins) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
	}

	transaction, err := heuristics.GetInputAmounts(ctx, dgraph, h.c.TransactionHash, constants.TypeDashMixing)
	if err != nil {
		return nil, err
	}

	inputDenominationCounts := getDashDenominationCounts(transaction)
	originAmounts := buildDashSourceAmounts(origins)
	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for k, o := range originAmounts {
		if isEqualDenomination(inputDenominationCounts, o) {
			for _, tx := range clusterTransactionMap[k] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], db.UIDNode{UID: tx.UID})
			}
		}
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}

// returns true if all denominations with the same amount of denom1 are contained in denom2
func isEqualDenomination(denom1 [dash.NumDenominations]int, denom2 [dash.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] != d {
			return false
		}
	}
	return true
}
