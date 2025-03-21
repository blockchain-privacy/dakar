package heuristics

import (
	"backend/analytics/classifier/btc"
	"backend/analytics/graph"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"fmt"
	"github.com/qrest/gomisc/serror"
)

// whirlpool2ReverseAmountHeuristic - see exec for description
type whirlpool2ReverseAmountHeuristic struct {
	heuristicType string
	c             heuristics.Options
}

func newWhirlpoolReverseAmountHeuristic() heuristic {
	return &whirlpool2ReverseAmountHeuristic{heuristicType: heuristicTypeWhirlpoolReverseAmount}
}

func (h *whirlpool2ReverseAmountHeuristic) getType() string {
	return h.heuristicType
}

func (h *whirlpool2ReverseAmountHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.c = c

	return nil
}

func (h *whirlpool2ReverseAmountHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *whirlpool2ReverseAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *whirlpool2ReverseAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Reverse amount",
		Type:        h.heuristicType,
		Category:    heuristicCategoryReverse,
		Description: "Returns all clusters which can fully fund the destination transaction via their origins.",
		AllowedParents: []string{heuristicTypeWhirlpoolReverseLookupByDepth, heuristicTypeWhirlpoolReverseLookupByTime,
			heuristicTypeWhirlpoolOneSourceByTime, heuristicTypeWhirlpoolOneSourceByDepth},
	}
}

// whirlpool2ReverseAmountHeuristic applies the following heuristic:
// - filter all origins of sources, which do not create enough output denominations to fund the destination transaction
func (h *whirlpool2ReverseAmountHeuristic) exec(ctx context.Context, dgraph external.Database, _ *graph.Wrapper, parentHeuristicUID string) (
	[]heuristics.HeuristicCluster, error) {
	parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
	if err != nil {
		return nil, err
	}
	// heuristic is only allowed to be connected to another heuristic
	if !parentHeuristicSet {
		return nil, serror.New(errHeuristicNotValid)
	}

	// get origins from parent heuristic
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	results, attributionMap, err := heuristics.GetHeuristicTransactions(ctx, dgraph, parentHeuristicUID,
		constants.TypeWhirlpoolMixing)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
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

	// 0: exact denomiatino because the mixing transaction outputs do not carry a fee
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

	return createHeuristicClusters(resultClusters, attributionMap), nil
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
