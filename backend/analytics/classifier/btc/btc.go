package btc

import (
	"backend/constants"
	"backend/db"
	"backend/db/analytics/classifier/btc"
	"backend/external"
	"context"
	"slices"
)

// NumWasabi2Denominations is the number of Wasabi 2.0 PrivateSend denominations
const NumWasabi2Denominations = 79

var denominationsTypesWasabi2 = [NumWasabi2Denominations]int64{5000, 6561, 8192, 10000, 13122, 16384, 19683, 20000,
	32768, 39366, 50000, 59049, 65536, 100000, 118098, 131072, 177147, 200000, 262144, 354294, 500000, 524288, 531441,
	1000000, 1048576, 1062882, 1594323, 2000000, 2097152, 3188646, 4194304, 4782969, 5000000, 8388608, 9565938,
	10000000, 14348907, 16777216, 20000000, 28697814, 33554432, 43046721, 50000000, 67108864, 86093442, 100000000,
	129140163, 134217728, 200000000, 258280326, 268435456, 387420489, 500000000, 536870912, 774840978, 1000000000,
	1073741824, 1162261467, 2000000000, 2147483648, 2324522934, 3486784401, 4294967296, 5000000000, 6973568802,
	8589934592, 10000000000, 10460353203, 17179869184, 20000000000, 20920706406, 31381059609, 34359738368, 50000000000,
	62762119218, 68719476736, 94143178827, 100000000000, 137438953472}

// Iterate returns
// - true when iterating should continue
// - false when not
func Iterate(ctx context.Context, c external.Database, from int64, to int64) (bool, error) {
	// get the transaction of the current block range
	transactions, err := db.GetTransactionsByBlock(ctx, c, from, to, true)
	if err != nil {
		return false, err
	}

	// step 1: classify all transactions of the current block locally based on their own properties
	mixingTransactions, err := classifyTransactions(transactions)
	if err != nil {
		return false, err
	}

	// step 2.1: store the privacy type of mixing transactions.
	if len(mixingTransactions) > 0 {
		if err = db.UpdateTransactions(ctx, c, mixingTransactions); err != nil {
			return false, err
		}
	}

	// step 2.2: set the privacy type of destination transactions by analyzing the connected transactions.
	if err = btc.ClassifyDestinationAndOriginsByBlock(ctx, c, from, to); err != nil {
		return false, err
	}

	return true, nil
}

// classifyTransactions detects mixing transactions and sets the privacy type appropriately
// The returned slice contains all classified transactions or nil if no privacy transactions have been found.
func classifyTransactions(transactions []db.Transaction) (mixing []db.Transaction, err error) {
	for _, transaction := range transactions {
		// only do classification for non-classified transactions
		if constants.IsValidDashTransactionType(transaction.Type) {
			continue
		}

		if isWasabi2Mixing(transaction) {
			mixing = append(mixing, newWasabi2MixingTransaction(transaction.UID))
			continue
		}
	}
	return
}

// isWasabi2Mixing checks if the transaction is a wasabi 2.0 mixing transaction
// credit to paper: "Heuristics for Detecting CoinJoin Transactions
// on the Bitcoin Blockchain" https://arxiv.org/abs/2311.12491
func isWasabi2Mixing(t db.Transaction) bool {
	// number of target inputs. Paper suggest 50, but practice shows that transactions with only 15 exist.
	const p = 15
	if len(t.Inputs) < p {
		return false
	}

	sigScripts := map[string]bool{}
	for _, o := range t.Outputs {
		sigScripts[o.KeyAsm] = true
	}
	// output scripts must be unique
	if len(sigScripts) != len(t.Outputs) {
		return false
	}

	// mininmum input output
	const vMin = int64(5000)
	if slices.ContainsFunc(t.Outputs, func(output db.Output) bool {
		return output.Amount == nil || *output.Amount < vMin
	}) {
		return false
	}

	denominationOut := countWasabi2Denominations(t.Outputs)
	var outputDenominationCount int
	for _, denomination := range denominationOut {
		outputDenominationCount += denomination
	}

	// number of participants
	const aMax = 10
	if outputDenominationCount < len(t.Inputs)/aMax {
		return false
	}

	// number of output denominations must be at least half of the number of outputs
	if outputDenominationCount < (len(t.Outputs)-1)/2 {
		return false
	}

	return true
}

// newWasabi2MixingTransaction returns a new wasabi 2.0 mixing transaction with the given type and uid.
func newWasabi2MixingTransaction(uid string) db.Transaction {
	return db.Transaction{UID: uid, Type: constants.TypeWasabi2Mixing}
}

// countWasabi2Denominations returns for each denomination how often it occurred in the given outputs
func countWasabi2Denominations(outputs []db.Output) [NumWasabi2Denominations]int {
	amounts := make([]int64, len(outputs))

	for i, o := range outputs {
		if o.Amount == nil {
			return [NumWasabi2Denominations]int{}
		}
		amounts[i] = *o.Amount
	}

	return CountAmountWasabi2Denominations(amounts)
}

// CountAmountWasabi2Denominations returns the number of occurrences of each denomination in the given amounts
func CountAmountWasabi2Denominations(amounts []int64) (denominations [NumWasabi2Denominations]int) {
	for _, amt := range amounts {
	inner:
		for i, v := range denominationsTypesWasabi2 {
			if amt == v {
				denominations[i]++
				break inner
			}
		}
	}

	return
}
