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

// oneSourceHeuristic - see exec for description
type oneSourceHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookBackTime         time.Duration
}

// newOneSourceHeuristic constructs an oneSourceHeuristic. hoursToLookBack in hours
func newOneSourceHeuristic(hoursToLookBack uint32) *oneSourceHeuristic {
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	return &oneSourceHeuristic{
		heuristicType:        "one_source",
		lookBackTime:         lBackTime,
		parameterDescription: strconv.FormatUint(uint64(hoursToLookBack), 10),
	}
}

func (h oneSourceHeuristic) getType() string {
	return h.heuristicType
}

func (h oneSourceHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h oneSourceHeuristic) hasParameter() bool {
	return true
}

func (h *oneSourceHeuristic) setParameter(p string) error {
	hoursToLookBack, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	h.lookBackTime = lBackTime
	h.parameterDescription = strconv.FormatUint(hoursToLookBack, 10)
	return nil
}

func (h oneSourceHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h oneSourceHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "One Source",
		Type:        h.heuristicType,
		Category:    heuristicCategoryReverse,
		Description: "Filters by time, direct input transaction amount filter and omni sources",
		Parameter: &struct {
			DefaultValue string `json:"value,omitempty"`
			Description  string `json:"description,omitempty"`
			Type         string `json:"type,omitempty"`
		}{
			DefaultValue: "48",
			Description:  "Look back time in hours",
			Type:         "int",
		},
	}
}

func (h oneSourceHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

type txAndOrigins struct {
	inputTransaction dbtxh.HeuristicTransaction
	origins          []dbtxh.HeuristicTransaction
}

// oneSourceHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
// - filter all origins of sources, which do not have enough denominations to fund all of their respective
//		outputs of input transaction which are used as inputs in the destination transaction
// - filter all origins of sources, which do not occur in all sets of input transaction origins
// This heuristic does not use the results from its parent heuristic
func (h oneSourceHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string, _ string) (
	[]dbtxh.HeuristicResult, error) {
	// Get all transactions which are connected via the inputs of the destination
	// transaction specified by txHash. These transactions are called >>input transactions<<.
	inputTransactions, err := dbtxh.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// sources holds all sources found in all input transactions
	sources := make(map[dbtxh.ClusterUID]bool)
	// mRemovableSources holds all sources which can be removed,
	// due to not being able to fund all connected input transactions
	mRemovableSources := make(map[dbtxh.ClusterUID]bool)
	// maps a cluster to its origin transactions
	sourceTransactionMap := make(map[dbtxh.ClusterUID]map[string]dbtxh.HeuristicTransaction)
	// for each input transaction to the destination transaction,
	// inputSources holds one map with all its occurring sources
	var inputSources []map[dbtxh.ClusterUID]bool
	// contains all time limited origins
	var allTimeLimitedOrigins []dbtxh.HeuristicTransaction
	// contains all time limited origins per input transaction
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

	// save origins in global cluster->origin map
	sourceTransactionMap, err = addOriginsToMap(sourceTransactionMap, allTimeLimitedOrigins)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	for _, t := range allTxAndOrigins {
		// get input denominations
		nDenominations, denominationIndex, getErr := getNumberOfDenominations(t.inputTransaction, txHash)
		if getErr != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), getErr)
		}

		oSource, buildErr := countClusterDenominations(t.origins, denominationIndex)
		if buildErr != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), buildErr)
		}

		// add element inputSources and set index of current element
		inputSources = append(inputSources, make(map[dbtxh.ClusterUID]bool))
		iSSIndex := len(inputSources) - 1

		// Loop through all clusters of the current input transaction and mark
		// the clusters which do not have enough denominations to fund all outputs of
		// the input transaction which are used as inputs in the destination transaction
		for k, v := range oSource.clusters {
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
	var omniSources []dbtxh.ClusterUID
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

	var ret []dbtxh.HeuristicResult
	for k := range remainingOrigins {
		ret = append(ret, dbtxh.HeuristicResult{
			Origin: dbtxh.DummyNode{UID: k},
		})
	}

	return ret, nil
}
