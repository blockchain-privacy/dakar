package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	dbop "backend/db/output"
	"backend/external"

	"fmt"
)

// DenominationTypeHeuristic - see exec for description
type DenominationTypeHeuristic struct {
	heuristicType        string
	parameterDescription string
}

// NewDenominationTypeHeuristic constructs a DenominationTypeHeuristic
func NewDenominationTypeHeuristic() DenominationTypeHeuristic {
	return DenominationTypeHeuristic{
		heuristicType: "denomination_type",
	}
}

func (h DenominationTypeHeuristic) getType() string {
	return h.heuristicType
}

func (h DenominationTypeHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h DenominationTypeHeuristic) hasParameter() bool {
	return false
}

func (h DenominationTypeHeuristic) setParameter(_ string) error {
	return nil
}

func (h DenominationTypeHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h DenominationTypeHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// DenominationTypeHeuristic applies the following heuristic:
// - filter all origins of sources, which have denominations of types which do not occur in the
//		denominations of the destination transaction
func (h DenominationTypeHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string) ([]string, error) {
	// origins holds all origins found bei either the parent heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]dbtxh.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[graph.ClusterID]map[string]dbtxh.HeuristicTransaction)
	var clusters map[string]graph.ClusterID
	{ // separate enclosure so the results slice can be garbage collected
		var results []dbtxh.HeuristicTransaction
		parentHeuristicSet := isParentHeuristicSet(parentHeuristicUID)

		if parentHeuristicSet {
			// get origins from parent heuristic
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

	var filteredOrigins []string
	for k, o := range originAmounts {
		if hasSameDenominationTypes(inputDenominationCounts, o) {
			for _, tx := range sourceTransactionMap[k] {
				filteredOrigins = append(filteredOrigins, tx.UID)
			}
		}
	}

	return filteredOrigins, nil
}

// returns true if both destinationDenominations and originDenominations have the exact same types
func hasSameDenominationTypes(destinationDenominations [dbop.NumDenominations]int,
	originDenominations [dbop.NumDenominations]int) bool {
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
