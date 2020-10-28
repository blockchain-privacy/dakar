package transaction

import (
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
)

type DenominationTypeHeuristic struct {
	heuristicType        string
	parameterDescription string
}

// DenominationTypeHeuristic constructor
func NewDenominationTypeHeuristic() DenominationTypeHeuristic {
	return DenominationTypeHeuristic{
		heuristicType: "denomination_type",
	}
}

func (h DenominationTypeHeuristic) getType() string {
	return h.heuristicType
}

func (h DenominationTypeHeuristic) getParameter() string {
	return h.parameterDescription
}

// DenominationTypeHeuristic applies the following heuristic:
// - filter all origins of sources, which have denominations of types which do not occur in the
//		denominations of the destination transaction
func (h DenominationTypeHeuristic) exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error) {
	// origins holds all origins found bei either the parent heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]dbtxh.HeuristicTransaction)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[string]map[string]dbtxh.HeuristicTransaction)

	{ // separate enclosure so the results slice can be garbage collected
		var results []dbtxh.HeuristicTransaction
		parentHeuristicSet := isParentHeuristicSet(parentHeuristicUid)

		if parentHeuristicSet {
			// get origins from parent heuristic
			var err error
			results, err = dbtxh.GetHeuristicResults(dgraph, parentHeuristicUid)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, err = dbtxh.GetOrigins(dgraph, txHash)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		sourceTransactionMap = addOriginsToMap(sourceTransactionMap, results)

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.Uid] = r
		}
	}

	if len(origins) == 0 {
		return nil, errors.New(fmt.Sprintln("found no origins:", parentHeuristicUid))
	}

	transaction, err := dbtxh.GetInputAmounts(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	inputDenominationCounts := getDenominationCounts(transaction)

	log.Println("Destination transaction denomination counts:", inputDenominationCounts)

	originAmounts := buildSourceAmounts(origins)

	log.Println("Sources found:", len(originAmounts), "Origins found:", len(origins))
	var filteredOrigins []string
	var atLeastOmniSource []string
	for k, o := range originAmounts {
		if hasSameDenominationTypes(inputDenominationCounts, o) {

			atLeastOmniSource = append(atLeastOmniSource, k)

			for _, tx := range sourceTransactionMap[k] {
				filteredOrigins = append(filteredOrigins, tx.Uid)
			}

		}
	}

	log.Println("Remaining sources after filter:", len(atLeastOmniSource),
		"Remaining origins:", len(filteredOrigins))

	return filteredOrigins, nil
}
