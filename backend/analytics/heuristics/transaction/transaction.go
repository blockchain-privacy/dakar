package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	dbop "backend/db/output"
	"backend/db/transaction"
	"backend/external"

	"errors"
	"fmt"
	"time"
)

var (
	// ErrorNoOriginsAtStart defines an error which should be used when no origins are available
	ErrorNoOriginsAtStart = errors.New("no origins can be fetched")
)

type heuristic interface {
	// exec executes the heuristic and returns the altered set of origin uids
	exec(dgraph external.Database, g *graph.Wrapper, txHash string, parentHeuristicUID string) ([]string, error)
	// getType returns the heuristic type
	getType() string
	// getParameterString returns the used parameter for this heuristic as a string
	getParameterString() string
	// hasParameter returns true if this heuristic has a parameter
	hasParameter() bool
	// setParameter sets the parameter
	setParameter(string) error
	// String returns the heuristic in string format
	String() string
	// clone clones an instance of this interface. This method is needed because
	// instances of interfaces can not be easily copied-by-value.
	// More information: https://stackoverflow.com/questions/37851500/how-to-copy-an-interface-value-in-go
	clone() heuristic
}

// getNumberOfDenominations returns the number of denominations. If destinationTransaction is set, it
// only counts outputs with input transactions equal to destinationTransaction. An error is returned
// if more than one type of denominations is found
func getNumberOfDenominations(it dbtxh.HeuristicTransaction, destinationTransaction string) (nDenominations int,
	denomIndex int, err error) {
	numDenominations := getDenominationCountsWithFilter(it, destinationTransaction)

	found := false
	for i, nd := range numDenominations {
		if nd > 0 {
			if found {
				err = fmt.Errorf("found more than one denomination type in input transaction %s for destination tx %s",
					it, destinationTransaction)
				return
			}
			denomIndex = i
			found = true
		}
	}
	nDenominations = numDenominations[denomIndex]
	return
}

// getDenominationCountsWithFilter gets the counts of each denomination type.
// If filterTx is set, it only counts outputs with input transactions equal to filterTx.
func getDenominationCountsWithFilter(it dbtxh.HeuristicTransaction, filterTx string) [dbop.NumDenominations]int {
	var denominations []int64
	for _, output := range it.Outputs {
		if filterTx != "" && output.InputTransaction != filterTx {
			continue
		}
		denominations = append(denominations, output.Amount)
	}

	return dbop.CountAmountDenominations(denominations)
}

// gets the counts of each denomination type
func getDenominationCounts(it dbtxh.HeuristicTransaction) [dbop.NumDenominations]int {
	var denominations []int64
	for _, output := range it.Outputs {
		denominations = append(denominations, output.Amount)
	}

	return dbop.CountAmountDenominations(denominations)
}

type originSource struct {
	denominationIndex int
	// key: cluster id, value: number of denominations of type denominationIndex
	sources map[graph.ClusterID]int
}

// addOriginsToMap adds all origins to their respective source in sourceTransactionMap.
// The returned map contains the provided origins
func addOriginsToMap(g *graph.Wrapper, sourceTransactionMap map[graph.ClusterID]map[string]dbtxh.HeuristicTransaction,
	origins []dbtxh.HeuristicTransaction) (map[graph.ClusterID]map[string]dbtxh.HeuristicTransaction,
	map[string]graph.ClusterID, error) {

	var allAddresses []string

	for _, o := range origins {
		// add to map
		allAddresses = append(allAddresses, o.Addresses...)
	}

	clusters, err := g.GetClusters(allAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	for _, o := range origins {
		clusterID := clusters[o.Addresses[0]]

		// add transaction to sourceTransactionMap
		transactions := sourceTransactionMap[clusterID]

		if len(transactions) == 0 {
			transactions = make(map[string]dbtxh.HeuristicTransaction)
		}

		transactions[o.UID] = o
		sourceTransactionMap[clusterID] = transactions
	}

	return sourceTransactionMap, clusters, nil
}

// buildSourcesWithAmount creates an array of sources with the
// number of denominations of the specified denomination type
func buildSourcesWithAmount(origins []dbtxh.HeuristicTransaction, denominationIndex int,
	clusters map[string]graph.ClusterID) (oSource originSource, err error) {
	oSource.denominationIndex = denominationIndex
	oSource.sources = make(map[graph.ClusterID]int)
	for _, o := range origins {
		nDenominations := getDenominationCounts(o)[denominationIndex]
		oSource.sources[clusters[o.Addresses[0]]] += nDenominations
	}

	return
}

func buildSourceAmounts(origins map[string]dbtxh.HeuristicTransaction,
	clusters map[string]graph.ClusterID) map[graph.ClusterID][dbop.NumDenominations]int {
	sourceAmounts := make(map[graph.ClusterID][dbop.NumDenominations]int)

	for _, o := range origins {
		clusterID := clusters[o.Addresses[0]]
		denominationSlice := getDenominationCounts(o)
		for i := range denominationSlice {
			denominationSlice[i] += sourceAmounts[clusterID][i]
		}

		sourceAmounts[clusterID] = denominationSlice
	}
	return sourceAmounts
}

// mapToSlice returns a slice containing all keys of the given map.
func mapToSlice(m map[string]bool) (uids []string) {
	for k := range m {
		uids = append(uids, k)
	}

	return
}

// getTimeLimitedOrigins returns all origins of the given transaction.
// If lookBackTime is bigger than zero only origins in the time range of
// tx.ts - lookBackTime will be returned.
func getTimeLimitedOrigins(dgraph external.Database, g *graph.Wrapper, tx dbtxh.HeuristicTransaction,
	lookBackTime time.Duration) (origins []dbtxh.HeuristicTransaction, err error) {
	// do reverse lookup
	endpoints, err := g.ReverseLookup(tx.UID, lookBackTime)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// get tx details for each uid
	origins, err = dbtxh.GetTransactionsWithOutputAmountAndInputAddresses(dgraph, mapToSlice(endpoints))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return
}

// getDestinationTxOrigins returns all origins of the given
// transaction, limited to a look back time of 90 days.
func getDestinationTxOrigins(dgraph external.Database, g *graph.Wrapper,
	txHash string) ([]dbtxh.HeuristicTransaction, error) {
	origins, err := getDestinationTxOriginsTimeLimited(dgraph, g, txHash, time.Hour*24*90)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	return origins, nil
}

// getDestinationTxOriginsTimeLimited returns all origins of the given
// transaction, for the given time limit.
func getDestinationTxOriginsTimeLimited(dgraph external.Database, g *graph.Wrapper,
	txHash string, dur time.Duration) (origins []dbtxh.HeuristicTransaction, err error) {
	// get uid for txhash
	uid, err := transaction.GetTransactionUID(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	inputTransactions, err := g.GetInputTransactions(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	uidMap := make(map[string]bool)
	// do reverse lookup for all input transactions
	for _, it := range inputTransactions {
		endpoints, err := g.ReverseLookup(it, dur)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		for k := range endpoints {
			uidMap[k] = true
		}
	}

	// store all uids from the map into a slice
	var uids []string
	for k := range uidMap {
		uids = append(uids, k)
	}

	// get tx details for each uid
	origins, err = dbtxh.GetTransactionsWithOutputAmountAndInputAddresses(dgraph, uids)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return
}

func isParentHeuristicSet(parentHeuristicUID string) bool {
	return parentHeuristicUID != ""
}

// HeuristicExecutor holds information for executing on heuristic and its children
type HeuristicExecutor struct {
	RootUID        string
	ThisHeuristic  heuristic
	NextHeuristics []HeuristicExecutor
}

// BuildExecutor is a convenience function for building heuristic executors
func BuildExecutor(thisHeuristic heuristic, nextHeuristics ...HeuristicExecutor) HeuristicExecutor {
	return HeuristicExecutor{
		ThisHeuristic:  thisHeuristic,
		NextHeuristics: nextHeuristics,
	}
}

// Run runs the given heuristic executor. The executor runs initial heuristic and
// triggers the Run function of all NextHeuristics. If parentHeuristicUID is not
// set (e.g. "") than the HeuristicExecutor.RootUID is used
func (hx HeuristicExecutor) Run(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string, userUID string) error {
	thisRootUID := hx.RootUID
	if parentHeuristicUID != "" {
		thisRootUID = parentHeuristicUID
	}

	newUID, err := Exec(dgraph, g, txHash, thisRootUID, hx.ThisHeuristic, userUID)
	if err != nil {
		// two fmt.Errorf so the error gets wrapped
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(),
			fmt.Errorf("heuristic type: %s, parameter: %s, %s",
				hx.ThisHeuristic.getType(), hx.ThisHeuristic.getParameterString(), err))
	}

	for _, executor := range hx.NextHeuristics {
		if runErr := executor.Run(dgraph, g, txHash, newUID, userUID); runErr != nil {
			// two fmt.Errorf so the error gets wrapped
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(),
				fmt.Errorf("heuristic type: %s, parameter: %s, %s",
					executor.ThisHeuristic.getType(), executor.ThisHeuristic.getParameterString(), runErr))
		}

	}

	var returnError error

	return returnError
}

// Exec executes the heuristic on the transaction specified by txHash for the given userUID
func Exec(dgraph external.Database, g *graph.Wrapper, txHash string, parentHeuristicUID string, h heuristic,
	userUID string) (thisUID string, err error) {
	originUIDs, err := h.exec(dgraph, g, txHash, parentHeuristicUID)
	if err != nil && !errors.Is(err, ErrorNoOriginsAtStart) {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// do not upsert heuristic for now
	var dummyOrigins []dbtxh.DummyOrigin

	for _, o := range originUIDs {
		dummyOrigins = append(dummyOrigins, dbtxh.DummyOrigin{UID: o})
	}

	// only set parent heuristic if uid is provided
	var pHeuristic []dbtxh.Heuristic
	if parentHeuristicUID != "" {
		pHeuristic = []dbtxh.Heuristic{{UID: parentHeuristicUID}}
	}

	thisUID, err = dbtxh.InsertHeuristic(dgraph, dbtxh.Heuristic{
		HeuristicType:   h.getType(),
		Origins:         dummyOrigins,
		Parameter:       h.getParameterString(),
		ParentHeuristic: pHeuristic,
		TxHash:          txHash,
	}, userUID)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}
