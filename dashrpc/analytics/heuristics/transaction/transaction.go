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
const nLockback = 1 * time.Hour

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
func getNumberOfDenominations(it dbtxh.Transaction) (nDenominations int, denomIndex int, err error) {
	var denominations []int64
	for _, output := range it.Outputs {
		denominations = append(denominations, output.Amount)
	}

	numDenominations := dbop.CountAmountDenominations(denominations)

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

type superSource struct {
	denomination int64
	sources      []struct {
		addressHash    string
		nDenominations int
	}
}

// todo: add description
func buildSuperSources(origins []dbtxh.Transaction, denomination int64) (superSource superSource, err error) {
	superSource.denomination = denomination

	return
}

// does nothing so far
func (b TimeConstraintHeuristic) exec(dgraph *dgo.Dgraph, txHash string, origins []string) ([]string, error) {
	inputTransactions, err := dbtxh.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, err
	}

	inputAmountMap := make(map[string]int)
	allUids := make(map[string]bool)

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
			return nil, err
		}

		for _, o := range timeLimitedOrigins {
			allUids[o.Uid] = true
		}

		// find super sources
		sSource, err := buildSuperSources(timeLimitedOrigins, dbop.DenominationsTypes[denominationIndex])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		log.Println(sSource)

		//log.Println(it.Uid, it.Timestamp, len(timeLimitedOrigins))
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
		return err
	}

	// todo remove
	log.Println("Original origin count:", len(origins))

	originUids, err := h.exec(dgraph, txHash, origins)
	if err != nil {
		return err
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
