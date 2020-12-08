package transaction

import (
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	dbop "dashrpc/db/output"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"sync"
	"time"
)

var (
	ErrorNoOriginsAtStart = errors.New("no origins can be fetched")
)

type heuristic interface {
	// exec executes the heuristic and returns the altered set of origin uids
	exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error)
	// getType returns the heuristic type
	getType() string
	// getParameterString returns the used parameter for this heuristic as a string
	getParameterString() string
	// hasParameter returns true if this heuristic has a parameter
	hasParameter() bool
	// setParameter sets the parameter
	setParameter(string) error
	// String returns the heurostic in string format
	String() string
}

// Returns the number of denominations.
// If filterTx is set, it only counts outputs with input transactions equal to filterTx
// An error is returned if more than one type of denominations is found
func getNumberOfDenominations(it dbtxh.HeuristicTransaction, destinationTransaction string) (nDenominations int, denomIndex int, err error) {
	numDenominations := getDenominationCountsWithFilter(it, destinationTransaction)

	found := false
	for i, nd := range numDenominations {
		if nd > 0 {
			if found {
				err = errors.New("found more than one type of denominations in input transaction")
				return
			}
			denomIndex = i
			found = true
		}
	}
	nDenominations = numDenominations[denomIndex]
	return
}

// gets the counts of each denomination type.
// If filterTx is set, it only counts outputs with input transactions equal to filterTx
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
	// key: address hash, value: number of denominations of type denominationIndex
	sources map[string]int
}

// returns a map of sources
func buildSources(origins []dbtxh.HeuristicTransaction) map[string]bool {
	sources := make(map[string]bool)
	for _, o := range origins {
		sources[o.Address] = true
	}

	return sources
}

// addOriginsToMap adds all origins to their respective source in sourceTransactionMap.
// The returned map contains the provided origins
func addOriginsToMap(sourceTransactionMap map[string]map[string]dbtxh.HeuristicTransaction,
	origins []dbtxh.HeuristicTransaction) map[string]map[string]dbtxh.HeuristicTransaction {

	for _, o := range origins {
		// add transaction to sourceTransactionMap
		transactions := sourceTransactionMap[o.Address]

		if len(transactions) == 0 {
			transactions = make(map[string]dbtxh.HeuristicTransaction)
		}

		transactions[o.Uid] = o
		sourceTransactionMap[o.Address] = transactions
	}

	return sourceTransactionMap
}

// buildSourcesWithAmount creates an array of sources with the number of denominations of the specified denomination type
func buildSourcesWithAmount(origins []dbtxh.HeuristicTransaction, denominationIndex int) (oSource originSource, err error) {
	oSource.denominationIndex = denominationIndex
	oSource.sources = make(map[string]int)
	for _, o := range origins {
		nDenominations := getDenominationCounts(o)[denominationIndex]
		oSource.sources[o.Address] += nDenominations
	}

	return
}

func buildSourceAmounts(origins map[string]dbtxh.HeuristicTransaction) map[string][dbop.NumDenominations]int {
	sourceAmounts := make(map[string][dbop.NumDenominations]int)

	for _, o := range origins {
		denominationSlice := getDenominationCounts(o)
		for i := range denominationSlice {
			denominationSlice[i] += sourceAmounts[o.Address][i]
		}
		sourceAmounts[o.Address] = denominationSlice
	}
	return sourceAmounts
}

func getTimeLimitedOrigins(dgraph *dgo.Dgraph, it dbtxh.HeuristicTransaction,
	lookbackTime time.Duration) (origins []dbtxh.HeuristicTransaction, err error) {
	// calculate look back time
	t, err := time.Parse(time.RFC3339, it.Timestamp)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	// get time limited origins
	t = t.Add(-1 * lookbackTime)
	origins, err = dbtxh.GetOriginsByDate(dgraph, it.Uid, t.Format(time.RFC3339))
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	return
}

func isParentHeuristicSet(parentHeuristicUid string) bool {
	return parentHeuristicUid != ""
}

type HeuristicExecutor struct {
	RootUid        string
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

// RunAsync runs the given heuristic executor. The executor runs initial heuristic and
// triggers the Run function of all NextHeuristics.
// WARNING: This will throw errors in certain cases while upserting a heuristic. This is the result
// of mutations of the same object. The upsert is built in a way, that in case of a failure this the mutation
// is done again. This way the result is inserted, despite the thrown error. Thus, this method achieves its goal.
// For a cleaner version, function use RunSynchronous
func (hx HeuristicExecutor) RunAsync(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) error {
	newUid, err := Exec(dgraph, txHash, parentHeuristicUid, hx.ThisHeuristic)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(),
			fmt.Errorf("heuristic type: %s, parameter: %s, %s",
				hx.ThisHeuristic.getType(), hx.ThisHeuristic.getParameterString(), err))
	}
	errChannel := make(chan error, len(hx.NextHeuristics))

	waitGroup := sync.WaitGroup{}
	for _, executor := range hx.NextHeuristics {
		waitGroup.Add(1)
		go func(e HeuristicExecutor, wg *sync.WaitGroup, eCH chan<- error) {
			defer wg.Done()
			if err := e.RunAsync(dgraph, txHash, newUid); err != nil {
				errChannel <- err
				return
			}
			errChannel <- nil
		}(executor, &waitGroup, errChannel)
	}

	waitGroup.Wait()
	var returnError error
	close(errChannel)
	for errs := range errChannel {
		if errs != nil {
			if returnError != nil {
				returnError = fmt.Errorf("%s, next error: %s", returnError, errs)
			} else {
				returnError = errs
			}
		}
	}

	return returnError
}

// RunSynchronous runs the given heuristic executor. The executor runs initial heuristic and
// triggers the RunSynchronous function of all NextHeuristics. If parentHeuristicUid is not
// set (e.g. "") than the HeuristicExecutor.RootUid is used
func (hx HeuristicExecutor) RunSynchronous(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) error {
	thisRootUid := hx.RootUid
	if parentHeuristicUid != "" {
		thisRootUid = parentHeuristicUid
	}

	newUid, err := Exec(dgraph, txHash, thisRootUid, hx.ThisHeuristic)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(),
			fmt.Errorf("heuristic type: %s, parameter: %s, %s",
				hx.ThisHeuristic.getType(), hx.ThisHeuristic.getParameterString(), err))
	}

	for _, executor := range hx.NextHeuristics {
		if err := executor.RunSynchronous(dgraph, txHash, newUid); err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(),
				fmt.Errorf("heuristic type: %s, parameter: %s, %s",
					executor.ThisHeuristic.getType(), executor.ThisHeuristic.getParameterString(), err))
		}

	}

	var returnError error

	return returnError
}

// fetchHeuristicSources returns the sources based on the results of the specified heuristic
func fetchHeuristicSources(dgraph *dgo.Dgraph, heuristicUid string) (sources map[string]bool, err error) {
	results, err := dbtxh.GetHeuristicResults(dgraph, heuristicUid)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	sources = buildSources(results)
	return
}

// Execute the heuristic on the transaction specified by txHash
func Exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string, h heuristic) (thisUid string, err error) {
	heuristicString := fmt.Sprintf("heuristic <%s> for tx %s", h.getType(), txHash)

	// todo remove
	log.Println(heuristicString, "starting")

	originUids, err := h.exec(dgraph, txHash, parentHeuristicUid)
	if err != nil && !errors.Is(err, ErrorNoOriginsAtStart) {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// do not upsert heuristic for now
	var dummyOrigins []dbtxh.DummyOrigin

	for _, o := range originUids {
		dummyOrigins = append(dummyOrigins, dbtxh.DummyOrigin{Uid: o})
	}

	// only set parent heuristic if uid is provided
	var pHeuristic []dbtxh.Heuristic
	if parentHeuristicUid != "" {
		pHeuristic = []dbtxh.Heuristic{{Uid: parentHeuristicUid}}
	}

	thisUid, err = dbtxh.InsertHeuristic(dgraph, dbtxh.Heuristic{
		HeuristicType:   h.getType(),
		Origins:         dummyOrigins,
		Parameter:       h.getParameterString(),
		ParentHeuristic: pHeuristic,
		TxHash:          txHash,
	})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// todo remove
	sources, err := fetchHeuristicSources(dgraph, thisUid)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// todo remove
	log.Println(heuristicString, "After heuristic origin count:", len(originUids), "sources count:", len(sources))
	//for k := range sources {
	//	log.Println(k)
	//}
	return
}
