package transaction

import (
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"time"
)

type TimeAmountConstraintHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookBackTime         time.Duration
}

// TimeAmountConstraintHeuristic constructor
// lookBackTime in hours
func NewTimeAmountConstraintHeuristic(hoursToLookBack time.Duration) TimeAmountConstraintHeuristic {
	lBackTime := hoursToLookBack * time.Hour
	return TimeAmountConstraintHeuristic{
		heuristicType:        "timeamountconstraint",
		lookBackTime:         lBackTime,
		parameterDescription: lBackTime.String(),
	}
}

func (b TimeAmountConstraintHeuristic) getType() string {
	return b.heuristicType
}

func (b TimeAmountConstraintHeuristic) getParameter() string {
	return b.parameterDescription
}

// time limitation
func (b TimeAmountConstraintHeuristic) exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error) {
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

	inputAmountMap := make(map[string]int)
	originMap := make(map[string]bool)

	mRemovableSupersources := make(map[string]bool)
	var inputSuperSources []map[string]bool

	// maps a address to its origin transactions
	sourceTransactionMap := make(map[string]map[string]dbtxh.HeuristicTransaction)
	for _, o := range origins {
		originMap[o] = true
	}

	for _, it := range inputTransactions {
		// get input denominations
		nDenominations, denominationIndex, err := getNumberOfDenominationsWithFilter(it, txHash)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		inputAmountMap[it.Uid] = nDenominations

		timeLimitedOrigins, err := getTimeLimitedOrigins(dgraph, it, b.lookBackTime)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// save in global address->origin map
		for _, t := range timeLimitedOrigins {
			// add transaction to sourceTransactionMap
			transactions := sourceTransactionMap[t.Address]

			if len(transactions) == 0 {
				transactions = make(map[string]dbtxh.HeuristicTransaction)
			}

			transactions[t.Uid] = t
			sourceTransactionMap[t.Address] = transactions
		}

		// find super sources
		sSource, err := buildSuperSourcesWithAmount(timeLimitedOrigins, denominationIndex)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// add new element
		inputSuperSources = append(inputSuperSources, make(map[string]bool))
		iSSIndex := len(inputSuperSources) - 1

		for k, v := range sSource.sources {
			inputSuperSources[iSSIndex][k] = true
			if v < inputAmountMap[it.Uid] {
				mRemovableSupersources[k] = true
			}
		}
	}

	log.Println("global super sources", len(sourceTransactionMap), "removable super sources", len(mRemovableSupersources))

	// remove super sources
	superSourceCounter := 0
	filteredOriginMap := make(map[string]bool)
	for k, origins := range sourceTransactionMap {
		if mRemovableSupersources[k] {
			continue
		}
		superSourceCounter++
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
