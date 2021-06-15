package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	"backend/external"

	"fmt"
	"strconv"
	"time"
)

// OneSourceHeuristic - see exec for description
type OneSourceHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookBackTime         time.Duration
}

// NewOneSourceHeuristic constructs an OneSourceHeuristic. hoursToLookBack in hours
func NewOneSourceHeuristic(hoursToLookBack uint32) *OneSourceHeuristic {
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	return &OneSourceHeuristic{
		heuristicType:        "one_source",
		lookBackTime:         lBackTime,
		parameterDescription: strconv.FormatUint(uint64(hoursToLookBack), 10),
	}
}

func (h OneSourceHeuristic) getType() string {
	return h.heuristicType
}

func (h OneSourceHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h OneSourceHeuristic) hasParameter() bool {
	return true
}

func (h *OneSourceHeuristic) setParameter(p string) error {
	hoursToLookBack, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	h.lookBackTime = lBackTime
	h.parameterDescription = strconv.FormatUint(hoursToLookBack, 10)
	return nil
}

func (h OneSourceHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h OneSourceHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

type txAndOrigins struct {
	inputTransaction dbtxh.HeuristicTransaction
	origins          []dbtxh.HeuristicTransaction
}

// OneSourceHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
// - filter all origins of sources, which do not have enough denominations to fund all of their respective
//		outputs of input transaction which are used as inputs in the destination transaction
// - filter all origins of sources, which do not occur in all sets of input transaction origins
// This heuristic does not use the results from its parent heuristic
func (h OneSourceHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string, _ string) ([]string, error) {
	// Get all transactions which are connected via the inputs of the destination
	// transaction specified by txHash. These transactions are called >>input transactions<<.
	inputTransactions, err := dbtxh.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// sources holds all sources found in all input transactions
	sources := make(map[graph.ClusterID]bool)
	// mRemovableSources holds all sources which can be removed,
	// due to not being able to fund all connected input transactions
	mRemovableSources := make(map[graph.ClusterID]bool)
	// maps an address to its origin transactions
	sourceTransactionMap := make(map[graph.ClusterID]map[string]dbtxh.HeuristicTransaction)
	// for each input transaction to the destination transaction,
	// inputSources holds one map with all its occurring sources
	var inputSources []map[graph.ClusterID]bool

	var allTimeLimitedOrigins []dbtxh.HeuristicTransaction

	// holds all
	var allTxAndOrigins []txAndOrigins
	for _, it := range inputTransactions {
		timeLimitedOrigins, err := getTimeLimitedOrigins(dgraph, g, it, h.lookBackTime)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(timeLimitedOrigins) == 0 {
			continue
		}

		allTimeLimitedOrigins = append(allTimeLimitedOrigins, timeLimitedOrigins...)

		allTxAndOrigins = append(allTxAndOrigins, txAndOrigins{inputTransaction: it, origins: timeLimitedOrigins})
	}

	// maps address uids to cluster id
	var clusters map[string]graph.ClusterID
	// save origins in global address->origin map
	sourceTransactionMap, clusters, err = addOriginsToMap(g, sourceTransactionMap, allTimeLimitedOrigins)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	for _, t := range allTxAndOrigins {
		// get input denominations
		nDenominations, denominationIndex, getErr := getNumberOfDenominations(t.inputTransaction, txHash)
		if getErr != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), getErr)
		}

		oSource, buildErr := buildSourcesWithAmount(t.origins, denominationIndex, clusters)
		if buildErr != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), buildErr)
		}

		// add element inputSources and set index of current element
		inputSources = append(inputSources, make(map[graph.ClusterID]bool))
		iSSIndex := len(inputSources) - 1

		// Loop through all sources of the current input transaction and mark
		// the sources which do not have enough denominations to fund all outputs of
		// the input transaction which are used as input in the destination transaction
		for k, v := range oSource.sources {
			sources[k] = true
			inputSources[iSSIndex][k] = true
			if v < nDenominations {
				mRemovableSources[k] = true
			}
		}
	}

	// Remove sources which do not have enough denominations to
	// fund all input transaction to which they are connected
	for k := range mRemovableSources {
		delete(sources, k)
	}

	// save all addresses (sources) which are not part of all input transactions
	var omniSources []graph.ClusterID
	for k := range sources {

		found := true
		// check for each input transaction if the source k exists. If not set the flag to false
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

	remainingOrigins := make(map[string]bool)
	// collect all transactions from omni sources
	for _, omniSource := range omniSources {
		omniOrigins := sourceTransactionMap[omniSource]
		for _, o := range omniOrigins {
			remainingOrigins[o.UID] = true
		}
	}

	var filteredOrigins []string
	for k := range remainingOrigins {
		filteredOrigins = append(filteredOrigins, k)
	}

	return filteredOrigins, nil
}
