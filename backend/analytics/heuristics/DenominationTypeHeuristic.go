package heuristics

import (
	"backend/analytics"
	"backend/analytics/graph"
	"backend/db"
	"backend/db/analytics/heuristics"
	"backend/external"
	"fmt"
	"github.com/qrest/gomisc/serror"
)

// denominationTypeHeuristic - see exec for description
type denominationTypeHeuristic struct {
	heuristicType        string
	parameterDescription string
	c                    heuristics.Options
}

func newDenominationTypeHeuristic() heuristic {
	return &denominationTypeHeuristic{heuristicType: "denomination_type"}
}

func (h *denominationTypeHeuristic) getType() string {
	return h.heuristicType
}

func (h *denominationTypeHeuristic) getParameterString() string {
	return h.parameterDescription
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
		Title:    "Denomination Type",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "The denomination type heuristic filters " +
			"all origins of clusters, which have denominations of " +
			"types which do not occur in the denominations of the " +
			"destination transaction. For example a destination " +
			"transaction spends 5 × 10.0001 and 10 × 1.00001. Now " +
			"all clusters are excluded which do not have these exact " +
			"two types of denominations.",
	}
}

// denominationTypeHeuristic applies the following heuristic:
//   - filter all origins of sources, which have denominations of types which do not occur in the
//     denominations of the destination transaction
func (h *denominationTypeHeuristic) exec(dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	// origins hold all origins found bei either the parent heuristic
	// or the destination transaction specified by txHash
	origins := make(map[string]heuristics.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	var attributionMap map[heuristics.ClusterUID][]string

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	{ // separate enclosure so the results slice can be garbage collected
		var results []heuristics.HeuristicTransaction
		parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
		if err != nil {
			return nil, err
		}
		if parentHeuristicSet {
			// get origins from parent heuristic
			var err error
			results, attributionMap, err = heuristics.GetHeuristicTransactions(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			results, attributionMap, err = getDestinationTxOrigins(ctx, dgraph, g, h.c.TransactionHash, h.c)
			if err != nil {
				return nil, err
			}
		}

		sourceTransactionMap = addOriginsToMap(sourceTransactionMap, results)

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
	}

	transaction, err := heuristics.GetInputAmounts(dgraph, h.c.TransactionHash)
	if err != nil {
		return nil, err
	}

	inputDenominationCounts := getDenominationCounts(transaction)

	originAmounts := buildSourceAmounts(origins)

	resultClusters := make(map[heuristics.ClusterUID][]heuristics.HeuristicResult)
	for k, o := range originAmounts {
		if hasSameDenominationTypes(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				resultClusters[tx.Cluster] = append(resultClusters[tx.Cluster], heuristics.HeuristicResult{
					Origin: db.UIDNode{UID: tx.UID},
				})
			}
		}
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}

// returns true if both destinationDenominations and originDenominations have the exact same types
func hasSameDenominationTypes(destinationDenominations [analytics.NumDenominations]int,
	originDenominations [analytics.NumDenominations]int) bool {
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
