package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	dbop "backend/db/output"
	"backend/external"

	"fmt"
)

// ReverseAmountHeuristic - see exec for description
type ReverseAmountHeuristic struct {
	heuristicType        string
	parameterDescription string
}

// NewReverseAmountHeuristic constructs an ReverseAmountHeuristic
func NewReverseAmountHeuristic() ReverseAmountHeuristic {
	return ReverseAmountHeuristic{
		heuristicType: "reverse_amount",
	}
}

func (h ReverseAmountHeuristic) getType() string {
	return h.heuristicType
}

func (h ReverseAmountHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h ReverseAmountHeuristic) hasParameter() bool {
	return false
}

func (h ReverseAmountHeuristic) setParameter(_ string) error {
	return nil
}

func (h ReverseAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h ReverseAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title: "Reverse Amount",
		Type:  h.heuristicType,
		Description: "Returns all origins of sources, which " +
			"have equal or more denominations to fund the " +
			"destination transaction.",
	}
}

func (h ReverseAmountHeuristic) clone() Heuristic {
	newHeuristic := h
	return &newHeuristic
}

// ReverseAmountHeuristic applies the following Heuristic:
// - filter all origins of sources, which do not have equal or more denominations to fund the destination transaction
func (h ReverseAmountHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string, parentHeuristicUID string) (
	[]dbtxh.HeuristicResult, error) {
	// origins holds all origins found bei either the parent Heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]dbtxh.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[graph.ClusterID]map[string]dbtxh.HeuristicTransaction)
	var clusters map[string]graph.ClusterID
	{ // separate enclosure so the results slice can be garbage collected
		var results []dbtxh.HeuristicTransaction
		if isParentHeuristicSet(parentHeuristicUID) {
			// get origins from parent Heuristic
			var err error
			results, err = dbtxh.GetHeuristicResults(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, err = getDestinationTxOrigins(dgraph, g, txHash)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		var err error
		sourceTransactionMap, clusters, err = addOriginsToMap(g, sourceTransactionMap, results)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 {
		return nil, ErrorNoOriginsAtStart
	}

	transaction, err := dbtxh.GetInputAmounts(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	inputDenominationCounts := getDenominationCounts(transaction)

	originAmounts := buildSourceAmounts(origins, clusters)

	var filteredOrigins []dbtxh.HeuristicResult
	for clusterID, denominationSlice := range originAmounts {
		if containsDenomination(inputDenominationCounts, denominationSlice) {
			// save all transaction uids of a particular cluster to the return set
			for _, tx := range sourceTransactionMap[clusterID] {

				filteredOrigins = append(filteredOrigins, dbtxh.HeuristicResult{Origin: dbtxh.DummyNode{UID: tx.UID}})
			}
		}
	}

	return filteredOrigins, nil
}

// containsDenomination returns true if all denominations with at least the same amount of denom1 are contained in denom2
func containsDenomination(denom1 [dbop.NumDenominations]int, denom2 [dbop.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] < d {
			return false
		}
	}
	return true
}
