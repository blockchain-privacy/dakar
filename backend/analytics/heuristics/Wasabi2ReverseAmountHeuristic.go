package heuristics

import (
	"backend/analytics/graph"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"fmt"
	"github.com/qrest/gomisc/serror"
)

// wasabi2ReverseAmountHeuristic - see exec for description
type wasabi2ReverseAmountHeuristic struct {
	heuristicType string
	c             heuristics.Options
}

func newWasabi2ReverseAmountHeuristic() heuristic {
	return &wasabi2ReverseAmountHeuristic{heuristicType: heuristicTypeWasabi2ReverseAmount}
}

func (h *wasabi2ReverseAmountHeuristic) getType() string {
	return h.heuristicType
}

func (h *wasabi2ReverseAmountHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.c = c

	return nil
}

func (h *wasabi2ReverseAmountHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *wasabi2ReverseAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
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

// wasabi2ReverseAmountHeuristic applies the following heuristic:
// - filter all origins of sources, which do not create enough output denominations to fund the destination transaction
func (h *wasabi2ReverseAmountHeuristic) exec(ctx context.Context, dgraph external.Database, _ *graph.Wrapper, parentHeuristicUID string) (
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
		constants.TransactionTypesWasabi2)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
	}

	transaction, err := heuristics.GetInputAmounts(ctx, dgraph, h.c.TransactionHash, constants.TransactionTypesWasabi2)
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

	// maps a cluster to its origin transactions
	clusterTransactionMap := addTransactionToCluster(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction{}, results)
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

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
