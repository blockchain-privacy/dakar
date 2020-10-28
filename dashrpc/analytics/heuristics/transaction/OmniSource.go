package transaction

import (
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
)

type OmniSourceHeuristic struct {
	heuristicType        string
	parameterDescription string
}

// OmniSourceHeuristic constructor
// lookBackTime in hours
func NewOmniSourceHeuristic() OmniSourceHeuristic {
	return OmniSourceHeuristic{
		heuristicType: "omnisource",
	}
}

func (b OmniSourceHeuristic) getType() string {
	return b.heuristicType
}

func (b OmniSourceHeuristic) getParameter() string {
	return b.parameterDescription
}

// time limitation
func (b OmniSourceHeuristic) exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error) {
	var origins []string
	parentHeuristicSet := isParentHeuristicSet(parentHeuristicUid)
	if parentHeuristicSet {
		// get origins from parent heuristic
		parentHeuristic, err := dbtxh.GetHeuristic(dgraph, parentHeuristicUid)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(parentHeuristic.Origins) == 0 {
			return nil, errors.New(fmt.Sprintln("found no origins found for parent heuristic:", parentHeuristicUid))
		}

		for _, o := range parentHeuristic.Origins {
			origins = append(origins, o.Uid)
		}
	}

	// gather input information
	inputTransactions, err := dbtxh.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	originMap := make(map[string]bool)

	// inputSuperSources holds a map per input transaction.
	// Each map holds a all super sources of the associated input transaction
	var inputSuperSources []map[string]bool

	// superSources holds all superSources found in all input transactions
	superSources := make(map[string]bool)

	// maps a address to its origin transactions
	sourceTransactionMap := make(map[string]map[string]dbtxh.HeuristicTransaction)
	for _, o := range origins {
		originMap[o] = true
	}

	for _, it := range inputTransactions {
		inputOrigins, err := dbtxh.GetOriginsByUid(dgraph, it.Uid)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// save in global address->origin map
		for _, t := range inputOrigins {
			// add transaction to sourceTransactionMap
			transactions := sourceTransactionMap[t.Address]

			if len(transactions) == 0 {
				transactions = make(map[string]dbtxh.HeuristicTransaction)
			}

			transactions[t.Uid] = t
			sourceTransactionMap[t.Address] = transactions
		}

		// get super sources
		sSource := buildSuperSources(inputOrigins)

		// add new element
		inputSuperSources = append(inputSuperSources, make(map[string]bool))
		iSSIndex := len(inputSuperSources) - 1

		for k := range sSource {
			superSources[k] = true
			inputSuperSources[iSSIndex][k] = true
		}
	}

	// save all addresses (super sources) which are part of all input transactions
	omniSource := make(map[string]bool)
	for k := range superSources {

		found := true
		for _, inputTransactionSource := range inputSuperSources {
			if !inputTransactionSource[k] {
				found = false
				break
			}
		}

		if found {
			omniSource[k] = true
		}
	}

	// filter out origins which are not part of supersources
	superSourceCounter := 0
	filteredOriginMap := make(map[string]bool)
	for k, origins := range sourceTransactionMap {
		if !omniSource[k] {
			continue
		}
		superSourceCounter++

		// TODO IS WRONG, PROBABLY HAVE TO DO ALL THE STEPS UNTIL THE FILTERS IN THE SAME STEP :(
		for _, o := range origins {
			if parentHeuristicSet && !originMap[o.Uid] {
				continue
			}
			filteredOriginMap[o.Uid] = true
		}
	}

	log.Println("Found supersources:", superSourceCounter)

	// convert map to string slice
	var filteredOrigins []string
	for k := range filteredOriginMap {
		filteredOrigins = append(filteredOrigins, k)
	}

	return filteredOrigins, nil
}
