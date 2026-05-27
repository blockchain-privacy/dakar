// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"context"
	"fmt"

	"gitlab.com/blockchain-privacy/dakar/analytics/classifier/btc"
	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/cmd/cliutil"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/heuristics"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// whirlpoolReverseAmountHeuristic - see exec for description
type whirlpoolReverseAmountHeuristic struct {
	heuristicType string
	c             HeuristicOptions
}

func NewWhirlpoolReverseAmountHeuristic() Heuristic {
	return &whirlpoolReverseAmountHeuristic{heuristicType: heuristicTypeWhirlpoolReverseAmount}
}

func (h *whirlpoolReverseAmountHeuristic) GetType() string {
	return h.heuristicType
}

func (h *whirlpoolReverseAmountHeuristic) SetConfig(c HeuristicOptions) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.c = c

	return nil
}

func (h *whirlpoolReverseAmountHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *whirlpoolReverseAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *whirlpoolReverseAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Reverse amount",
		Type:        h.heuristicType,
		Category:    heuristicCategoryReverse,
		Description: "Returns all clusters which can fully fund the destination transaction via their origins.",
		AllowedParents: []string{heuristicTypeWhirlpoolReverseLookupByDepth, heuristicTypeWhirlpoolReverseLookupByTime,
			heuristicTypeWhirlpoolOneSourceByTime, heuristicTypeWhirlpoolOneSourceByDepth},
	}
}

// Exec of the whirlpoolReverseAmountHeuristic applies the following heuristic:
// - filter all origins of sources, which do not create enough output denominations to fund the destination transaction
func (h *whirlpoolReverseAmountHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
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
			constants.TypeWhirlpoolMixing)
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return nil, serror.New(errNoOriginsAtStart)
		}
	} else {
		results, err = getHeuristicTransactions(ctx, dgraph, parentResults, constants.TypeWhirlpoolMixing)
		if err != nil {
			return nil, err
		}
	}

	transaction, err := heuristics.GetInputAmounts(ctx, dgraph, h.c.TransactionHash, constants.TypeWhirlpoolMixing)
	if err != nil {
		return nil, err
	}

	// origins hold all origins found by the parent heuristic
	origins := make(map[string]heuristics.HeuristicTransaction, len(results))
	// Convert from slice to Hash
	for _, r := range results {
		origins[r.UID] = r
	}

	if err = mergeOriginPeelchainCluster(g, origins); err != nil {
		return nil, err
	}

	// apply new cluster to results
	for i := range results {
		results[i].Cluster = origins[results[i].UID].Cluster
	}

	// 0: exact denomination because the mixing transaction outputs do not carry a fee
	inputDenominationCounts := getWhirlpoolDenominationCounts(transaction, 0)
	// 100: because origin outputs also carry a fee
	originAmounts := buildWhirlpoolSourceAmounts(origins, 100)
	clusterTransactionMap := mapClusterToTransactions(results)
	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for clusterID, denominationSlice := range originAmounts {
		if containsWhirlpoolDenomination(inputDenominationCounts, denominationSlice) {
			// add all transaction uids of a particular cluster to the return set
			for _, tx := range clusterTransactionMap[clusterID] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], db.UIDNode{UID: tx.UID})
			}
		}
	}

	return createHeuristicClusters(resultClusters), nil
}

// containsWhirlpoolDenomination returns true if all denominations with at
// least the same amount of denom1 are contained in denom2
func containsWhirlpoolDenomination(denom1 [btc.NumWhirlpoolDenominations]int, denom2 [btc.NumWhirlpoolDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] < d {
			return false
		}
	}
	return true
}

// mergeOriginPeelchainCluster detects if the origins are connected via a
// peelchain and assigns the same cluster to all transactions part of a set.
func mergeOriginPeelchainCluster(g *graph.Wrapper, origins map[string]heuristics.HeuristicTransaction) error {
	partitions, err := g.PartitionNodesByDirectConnections(cliutil.GetMapKeys(origins))
	if err != nil {
		return err
	}

	for _, neighbours := range partitions {
		if len(neighbours) < 2 {
			// nothing to do
			continue
		}

		firstNeighbour, ok := origins[neighbours[0]]
		if !ok {
			return serror.FromStrWithContext("partitioned node not found", "node", neighbours[0])
		}

		clusterMergeList := map[string]bool{}
		for _, n := range neighbours {
			clusterMergeList[string(origins[n].Cluster)] = true
		}

		for k, v := range origins {
			if clusterMergeList[string(v.Cluster)] {
				// make all neighbours belong to the same cluster
				v.Cluster = firstNeighbour.Cluster
				origins[k] = v
			}
		}
	}

	return nil
}
