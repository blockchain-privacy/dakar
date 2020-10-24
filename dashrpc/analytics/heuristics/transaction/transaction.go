package transaction

import (
	"dashrpc/cmd/cliutil"
	dban "dashrpc/db/analytics"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	dbop "dashrpc/db/output"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"time"
)

// nLockback is the number of hours which limits the maximal lookback on origins
const nLockback = 5 * 24 * time.Hour

type heuristic interface {
	// exec executes the heuristic and returns the altered set of origin uids
	exec(dgraph *dgo.Dgraph, txHash string, origins []string) ([]string, error)
	// getType returns the heuristic type
	getType() string
}

type TimeConstraintHeuristic struct {
	heuristicType string
}

// TimeConstraintHeuristic constructor
func NewTimeConstraintHeuristic() TimeConstraintHeuristic {
	return TimeConstraintHeuristic{
		heuristicType: "timeconstraint",
	}
}

// Returns the number of denominations.
// An error is returned if more than one type of denominations is found
func getNumberOfDenominations(it dbtxh.HeuristicTransaction) (nDenominations int, denomIndex int, err error) {
	numDenominations := getDenominationCounts(it)

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

// gets the counts of each denomination type
func getDenominationCounts(it dbtxh.HeuristicTransaction) [dbop.NumDenominations]int {
	var denominations []int64
	for _, output := range it.Outputs {
		denominations = append(denominations, output.Amount)
	}

	return dbop.CountAmountDenominations(denominations)
}

type superSource struct {
	denominationIndex int
	// key: address hash, value: number of denominations of type denominationIndex
	sources map[string]int
}

// todo: add description
func buildSuperSources(origins []dbtxh.HeuristicTransaction, denominationIndex int) (superSource superSource, err error) {
	superSource.denominationIndex = denominationIndex
	superSource.sources = make(map[string]int)
	for _, o := range origins {
		nDenominations := getDenominationCounts(o)[denominationIndex]
		superSource.sources[o.Address] += nDenominations
	}

	return
}

func buildSuperSourceAmounts(origins map[string]dbtxh.HeuristicTransaction) map[string][dbop.NumDenominations]int {
	superSourceAmounts := make(map[string][dbop.NumDenominations]int)

	for _, o := range origins {
		denominationSlice := getDenominationCounts(o)
		for i := range denominationSlice {
			denominationSlice[i] += superSourceAmounts[o.Address][i]
		}
		superSourceAmounts[o.Address] = denominationSlice
	}
	return superSourceAmounts
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

// does nothing so far
func (b TimeConstraintHeuristic) exec(dgraph *dgo.Dgraph, txHash string, origins []string) ([]string, error) {

	transaction, err := dbtxh.GetInputAmounts(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// todo: use later
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

	var inputSuperSources []map[string]bool

	for _, it := range inputTransactions {
		// get input denominations
		nDenominations, denominationIndex, err := getNumberOfDenominations(it)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		inputAmountMap[it.Uid] = nDenominations

		// get time limited origins
		t, err := time.Parse(time.RFC3339, it.Timestamp)
		if err != nil {
			return nil, err
		}

		t = t.Add(-1 * nLockback)
		timeLimitedOrigins, err := dbtxh.GetOriginsByDate(dgraph, it.Uid, t.Format(time.RFC3339))
		if err != nil {
			log.Println("continued")
			continue
			//return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		for _, o := range timeLimitedOrigins {
			allOrigins[o.Uid] = o
			allUids[o.Uid] = true
		}

		// find super sources
		sSource, err := buildSuperSources(timeLimitedOrigins, denominationIndex)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		inputSuperSources = append(inputSuperSources, make(map[string]bool))
		iSSIndex := len(inputSuperSources) - 1
		var removableSuperSources int
		for k, v := range sSource.sources {
			superSources[k] = true
			inputSuperSources[iSSIndex][k] = true
			if v < inputAmountMap[it.Uid] {
				mRemovableSupersources[k] = true
				removableSuperSources++
			}
		}

		log.Println("time limited origins:", len(timeLimitedOrigins), "number of supersources", len(sSource.sources), "removable:", removableSuperSources)

		//log.Println(it.Uid, it.Timestamp, len(timeLimitedOrigins))
	}

	originAmounts := buildSuperSourceAmounts(allOrigins)

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
			} else {
				atLeastOmniSource = append(atLeastOmniSource, o)
			}
		}

	}
	log.Println("Remaining omni sources after <at least> denomination amount filter:", len(atLeastOmniSource))
	log.Println("Remaining omni sources after <same type> denomination amount filter:", len(sameTypeOmniSource))
	log.Println("Remaining omni sources after <exact> denomination amount filter:", len(exactOmniSource))

	for _, o := range sameTypeOmniSource {
		log.Println(o)
	}

	var filteredOrigins []string
	for k := range allUids {
		filteredOrigins = append(filteredOrigins, k)
	}

	return filteredOrigins, nil
}

func (b TimeConstraintHeuristic) getType() string {
	return b.heuristicType
}

// Execute the heuristic on the transaction specified by txHash
func Exec(dgraph *dgo.Dgraph, txHash string, h heuristic) error {
	// todo remove
	log.Println("Starting heuristic", h.getType(), "for tx", txHash)

	origins, err := dban.GetOrigins(dgraph, txHash)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// todo remove
	log.Println("Original origin count:", len(origins))

	originUids, err := h.exec(dgraph, txHash, origins)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// todo remove
	log.Println("After heuristic origin count:", len(originUids))

	// do not upsert heuristic for now
	//var dummyOrigins []dbtxh.DummyOrigin
	//
	//for _, o := range originUids {
	//	dummyOrigins = append(dummyOrigins, dbtxh.DummyOrigin{Uid: o})
	//}

	//if err := dbtxh.UpsertHeuristic(dgraph, dbtxh.Heuristic{
	//	HeuristicType: h.getType(),
	//	Origins:       dummyOrigins,
	//	TxHash:        txHash,
	//}); err != nil {
	//	return err
	//}

	return nil
}
