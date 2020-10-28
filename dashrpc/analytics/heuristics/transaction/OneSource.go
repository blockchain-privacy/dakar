package transaction

import (
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"time"
)

type OneSourceHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookBackTime         time.Duration
}

// OneSourceHeuristic constructor
// lookBackTime in hours
func NewOneSourceHeuristic(hoursToLookBack time.Duration) OneSourceHeuristic {
	lBackTime := hoursToLookBack * time.Hour
	return OneSourceHeuristic{
		heuristicType:        "one_source",
		lookBackTime:         lBackTime,
		parameterDescription: lBackTime.String(),
	}
}

func (h OneSourceHeuristic) getType() string {
	return h.heuristicType
}

func (h OneSourceHeuristic) getParameter() string {
	return h.parameterDescription
}

// time limitation
func (h OneSourceHeuristic) exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error) {
	transaction, err := dbtxh.GetInputAmounts(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	inputDenominationCounts := getDenominationCounts(transaction)

	log.Println("Destination transaction denomination counts:", inputDenominationCounts)

	inputTransactions, err := dbtxh.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	inputAmountMap := make(map[string]int)

	sources := make(map[string]bool)
	mRemovableSources := make(map[string]bool)

	// maps an address to its origin transactions
	sourceTransactionMap := make(map[string]map[string]dbtxh.HeuristicTransaction)

	var inputSources []map[string]bool
	for _, it := range inputTransactions {
		// get input denominations
		nDenominations, denominationIndex, err := getNumberOfDenominations(it, txHash)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		inputAmountMap[it.Uid] = nDenominations

		timeLimitedOrigins, err := getTimeLimitedOrigins(dgraph, it, h.lookBackTime)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// save origins in global address->origin map
		sourceTransactionMap = addOriginsToMap(sourceTransactionMap, timeLimitedOrigins)

		// find super sources
		oSource, err := buildSourcesWithAmount(timeLimitedOrigins, denominationIndex)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		inputSources = append(inputSources, make(map[string]bool))
		iSSIndex := len(inputSources) - 1
		for k, v := range oSource.sources {
			sources[k] = true
			inputSources[iSSIndex][k] = true
			if v < inputAmountMap[it.Uid] {
				mRemovableSources[k] = true
			}
		}
	}

	log.Println("sources:", len(sources), "-- removable sources:", len(mRemovableSources))

	// remove sources
	for k := range mRemovableSources {
		delete(sources, k)
	}

	log.Println("remaining sources", len(sources))

	// save all addresses (sources) which are part of all input transactions
	var omniSources []string
	for k := range sources {

		found := true
		for _, inputTransactionSource := range inputSources {
			if !inputTransactionSource[k] {
				found = false
				break
			}
		}

		if found {
			omniSources = append(omniSources, k)
		}
	}

	log.Println("Found", len(omniSources), "omni sources")

	remainingOrigins := make(map[string]bool)
	// collect all transactions from omni sources
	for _, omniSource := range omniSources {
		omniOrigins := sourceTransactionMap[omniSource]
		for _, o := range omniOrigins {
			remainingOrigins[o.Uid] = true
		}
	}

	var filteredOrigins []string
	for k := range remainingOrigins {
		filteredOrigins = append(filteredOrigins, k)
	}

	return filteredOrigins, nil
}
