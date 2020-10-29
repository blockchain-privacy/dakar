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

type heuristic interface {
	// exec executes the heuristic and returns the altered set of origin uids
	exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error)
	// getType returns the heuristic type
	getType() string
	// getParameter returns the used paramter for this heuristic
	getParameter() string
}

type InputAmountHeuristic struct {
	heuristicType string
}

// NewInputAmountHeuristic constructor
func NewInputAmountHeuristic() InputAmountHeuristic {
	return InputAmountHeuristic{
		heuristicType: "inputamount",
	}
}

func (i InputAmountHeuristic) getType() string {
	return i.heuristicType
}

func (i InputAmountHeuristic) getParameter() string {
	return ""
}

// filter by amount of input transactions
func (i InputAmountHeuristic) exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error) {
	return nil, nil
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

// returns a map of super sources
func buildSuperSources(origins []dbtxh.HeuristicTransaction) map[string]bool {
	superSources := make(map[string]bool)
	for _, o := range origins {
		superSources[o.Address] = true
	}

	return superSources
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

// returns true if all denominations with at least the same amount of denom1 are contained in denom2
func containsDenomination(denom1 [dbop.NumDenominations]int, denom2 [dbop.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] < d {
			return false
		}
	}
	return true
}

// returns true if all denominations with the same amount of denom1 are contained in denom2
func isEqualDenomination(denom1 [dbop.NumDenominations]int, denom2 [dbop.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] != d {
			return false
		}
	}
	return true
}

// returns true if denom1 has only denominations for the same types as denom2
func hasSameDenominationTypes(denom1 [dbop.NumDenominations]int, denom2 [dbop.NumDenominations]int) bool {
	for i, d := range denom1 {
		if denom2[i] == d && d == 0 {
			continue
		}

		if (denom2[i] > 0 && d == 0) || (denom2[i] == 0 && d > 0) {
			return false
		}
	}
	return true
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

type AllHeuristics struct {
	heuristicType        string
	parameterDescription string
	lookBackTime         time.Duration
}

// NewAllHeuristics constructor
// lookBackTime in hours
func NewAllHeuristics(hoursToLookBack uint32) AllHeuristics {
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	return AllHeuristics{
		heuristicType:        "all",
		lookBackTime:         lBackTime,
		parameterDescription: lBackTime.String(),
	}
}

func (b AllHeuristics) getType() string {
	return b.heuristicType
}

func (b AllHeuristics) getParameter() string {
	return b.parameterDescription
}

// todo remove
// does it all
func (b AllHeuristics) exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error) {

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
	allUids := make(map[string]bool)

	superSources := make(map[string]bool)
	mRemovableSupersources := make(map[string]bool)
	allOrigins := make(map[string]dbtxh.HeuristicTransaction)

	sourceTransactionMap := make(map[string]map[string]dbtxh.HeuristicTransaction)

	var inputSuperSources []map[string]bool

	for _, it := range inputTransactions {
		// get input denominations
		nDenominations, denominationIndex, err := getNumberOfDenominations(it, txHash)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		inputAmountMap[it.Uid] = nDenominations

		timeLimitedOrigins, err := getTimeLimitedOrigins(dgraph, it, b.lookBackTime)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		for _, o := range timeLimitedOrigins {
			allOrigins[o.Uid] = o
			allUids[o.Uid] = true

			// add transaction to sourceTransactionMap
			transactions := sourceTransactionMap[o.Address]

			if len(transactions) == 0 {
				transactions = make(map[string]dbtxh.HeuristicTransaction)
			}

			transactions[o.Uid] = o
			sourceTransactionMap[o.Address] = transactions
		}

		// find super sources
		sSource, err := buildSourcesWithAmount(timeLimitedOrigins, denominationIndex)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		inputSuperSources = append(inputSuperSources, make(map[string]bool))
		iSSIndex := len(inputSuperSources) - 1
		for k, v := range sSource.sources {
			superSources[k] = true
			inputSuperSources[iSSIndex][k] = true
			if v < inputAmountMap[it.Uid] {
				mRemovableSupersources[k] = true
			}
		}
	}

	originAmounts := buildSourceAmounts(allOrigins)

	log.Println("global super sources", len(superSources), "removable super sources", len(mRemovableSupersources))

	// remove super sources
	for k := range mRemovableSupersources {
		delete(superSources, k)
	}

	// save all addresses (super sources) which are part of all input transactions
	var omniSource []string
	for k := range superSources {

		found := true
		for _, inputTransactionSource := range inputSuperSources {
			if !inputTransactionSource[k] {
				found = false
				break
			}
		}

		if found {
			omniSource = append(omniSource, k)
		}
	}

	log.Println("Found", len(omniSource), "omni sources")

	var atLeastOmniSource []string
	var sameTypeOmniSource []string
	var exactOmniSource []string
	for _, o := range omniSource {
		denominations := originAmounts[o]
		if containsDenomination(inputDenominationCounts, denominations) {
			if isEqualDenomination(inputDenominationCounts, denominations) {
				exactOmniSource = append(exactOmniSource, o)
			} else if hasSameDenominationTypes(inputDenominationCounts, denominations) {
				sameTypeOmniSource = append(sameTypeOmniSource, o)
			}
			atLeastOmniSource = append(atLeastOmniSource, o)
		}
	}

	log.Println("Remaining omni sources after <at least> denomination filter:", len(atLeastOmniSource))
	log.Println("Remaining omni sources after <at least> + <same type> denomination filter:", len(sameTypeOmniSource))
	log.Println("Remaining omni sources after <at least> + <exact> denomination filter:", len(exactOmniSource))

	for _, o := range sameTypeOmniSource {
		log.Println(o, len(sourceTransactionMap[o]), originAmounts[o])
	}

	var filteredOrigins []string
	for k := range allUids {
		filteredOrigins = append(filteredOrigins, k)
	}

	return filteredOrigins, nil
}

func isParentHeuristicSet(parentHeuristicUid string) bool {
	return parentHeuristicUid != ""
}

type HeuristicExecutor struct {
	ThisHeuristic  heuristic
	NextHeuristics []HeuristicExecutor
}

// BuildExecutor is a convenience function for build heuristic executors
func BuildExecutor(thisHeuristic heuristic, nextHeuristics ...HeuristicExecutor) HeuristicExecutor {
	return HeuristicExecutor{
		ThisHeuristic:  thisHeuristic,
		NextHeuristics: nextHeuristics,
	}
}

// Run runs the given heuristic executor. The executor runs initial heuristic and
// triggers the Run function of the NextHeuristics
func (hx HeuristicExecutor) Run(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) error {
	newUid, err := Exec(dgraph, txHash, parentHeuristicUid, hx.ThisHeuristic)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	errChannel := make(chan error, len(hx.NextHeuristics))

	waitGroup := sync.WaitGroup{}
	for _, executor := range hx.NextHeuristics {
		waitGroup.Add(1)
		go func(e HeuristicExecutor, wg *sync.WaitGroup, eCH chan<- error) {
			defer wg.Done()
			if err := e.Run(dgraph, txHash, newUid); err != nil {
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
				returnError = fmt.Errorf("%s, %s", returnError, errs)
			} else {
				returnError = errs
			}
		}
	}

	return returnError
}

// Execute the heuristic on the transaction specified by txHash
func Exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string, h heuristic) (thisUid string, err error) {
	// todo remove
	log.Println("Starting heuristic <", h.getType(), "> for tx", txHash)

	originUids, err := h.exec(dgraph, txHash, parentHeuristicUid)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// todo remove
	log.Println("After heuristic origin count:", len(originUids))

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

	thisUid, err = dbtxh.UpsertHeuristic(dgraph, dbtxh.Heuristic{
		HeuristicType:   h.getType(),
		Origins:         dummyOrigins,
		Parameter:       h.getParameter(),
		ParentHeuristic: pHeuristic,
		TxHash:          txHash,
	})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}
