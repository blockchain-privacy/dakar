package heuristics

import (
	"backend/analytics/classifier/dash"
	"backend/analytics/graph"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"fmt"
	"github.com/qrest/gomisc/serror"
)

// denominationTypeHeuristic - see exec for description
type denominationTypeHeuristic struct {
	heuristicType string
	c             heuristics.Options
}

func newDenominationTypeHeuristic() heuristic {
	return &denominationTypeHeuristic{heuristicType: heuristicTypeDenominationType}
}

func (h *denominationTypeHeuristic) getType() string {
	return h.heuristicType
}

func (h *denominationTypeHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.c = c

	return nil
}

func (h *denominationTypeHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *denominationTypeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *denominationTypeHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Denomination type",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Filters all origins of clusters, which have denominations of " +
			"types which do not occur in the denominations of the " +
			"destination transaction. For example a destination " +
			"transaction spends 5 × 10.0001 and 10 × 1.00001. Now " +
			"all clusters are excluded which do not have these exact " +
			"two types of denominations.",
		AllowedParents: []string{heuristicTypeReverseLookup, heuristicTypeOneSource, heuristicTypeReverseAmount, heuristicTypePerfect},
	}
}

// denominationTypeHeuristic applies the following heuristic:
//   - filter all origins of sources, which have denominations of types which do not occur in the
//     denominations of the destination transaction
func (h *denominationTypeHeuristic) exec(ctx context.Context, dgraph external.Database, _ *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
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
		constants.TransactionTypesDash)
	if err != nil {
		return nil, err
	}

	// maps an address to its origin transactions
	sourceTransactionMap := mapClusterToTransactions(results)

	// origins hold all origins found bei either the parent heuristic
	// or the destination transaction specified by txHash
	origins := make(map[string]heuristics.HeuristicTransaction, len(results))
	// Convert from slice to Hash
	for _, r := range results {
		origins[r.UID] = r
	}

	if len(origins) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
	}

	transaction, err := heuristics.GetInputAmounts(ctx, dgraph, h.c.TransactionHash, constants.TransactionTypesDash)
	if err != nil {
		return nil, err
	}

	inputDenominationCounts := getDenominationCounts(transaction)

	originAmounts := buildSourceAmounts(origins)

	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for k, o := range originAmounts {
		if hasSameDenominationTypes(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], db.UIDNode{UID: tx.UID})
			}
		}
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}

// returns true if both destinationDenominations and originDenominations have the exact same types
func hasSameDenominationTypes(destinationDenominations [dash.NumDenominations]int,
	originDenominations [dash.NumDenominations]int) bool {
	for i, destinationDenomination := range destinationDenominations {
		if originDenominations[i] == destinationDenomination && destinationDenomination == 0 {
			continue
		}

		if (originDenominations[i] > 0 && destinationDenomination == 0) ||
			(originDenominations[i] == 0 && destinationDenomination > 0) {
			return false
		}
	}
	return true
}
