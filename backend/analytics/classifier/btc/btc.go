// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

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

const NumWhirlpoolDenominations = 4

var denominationsTypesWasabi2 = [NumWasabi2Denominations]int64{5000, 6561, 8192, 10000, 13122, 16384, 19683, 20000,
	32768, 39366, 50000, 59049, 65536, 100000, 118098, 131072, 177147, 200000, 262144, 354294, 500000, 524288, 531441,
	1000000, 1048576, 1062882, 1594323, 2000000, 2097152, 3188646, 4194304, 4782969, 5000000, 8388608, 9565938,
	10000000, 14348907, 16777216, 20000000, 28697814, 33554432, 43046721, 50000000, 67108864, 86093442, 100000000,
	129140163, 134217728, 200000000, 258280326, 268435456, 387420489, 500000000, 536870912, 774840978, 1000000000,
	1073741824, 1162261467, 2000000000, 2147483648, 2324522934, 3486784401, 4294967296, 5000000000, 6973568802,
	8589934592, 10000000000, 10460353203, 17179869184, 20000000000, 20920706406, 31381059609, 34359738368, 50000000000,
	62762119218, 68719476736, 94143178827, 100000000000, 137438953472}

var denominationTypesWhirlpool = [NumWhirlpoolDenominations]int64{100000, 1000000, 5000000, 50000000}

// Iterate returns
// - true when iterating should continue
// - false when not
func Iterate(ctx context.Context, c external.Database, from int64, to int64) (bool, error) {
	// get the transaction of the current block range
	transactions, err := db.GetTransactionsByBlock(ctx, c, from, to, nil)
	if err != nil {
		return false, err
	}

	// step 1.1: classify all transactions of the current block locally based on their own properties
	wasabiMixing, whirlpoolMixingUIDs, err := classifyTransactions(ctx, c, transactions)
	if err != nil {
		return false, err
	}

	// step 1.2: update transactions classified as wasabi 2.0 mixing transactions.
	if len(wasabiMixing) > 0 {
		if err = db.UpdateTransactions(ctx, c, wasabiMixing); err != nil {
			return false, err
		}
	}

	if len(whirlpoolMixingUIDs) > 0 {
		// step 2.1: get transactions which provide inputs to the potential mixing transactions.
		// Note: this may include already classified transactions
		potWhirlpoolOrigins, originsToMixingMap, err := btc.GetPotentialWhirlpoolMixingTransactions(ctx, c, whirlpoolMixingUIDs)
		if err != nil {
			return false, err
		}

		// step 2.2: whirlpool mixing transactions must be connected to at least one whirlpool origin transaction.
		// Whirlpool origin transactions have a high chance to be misclassified,
		// therefore classify both mixing and origin transactions only if they are connected to each other.
		if len(potWhirlpoolOrigins) > 0 {
			// - get all unclassified transactions which are connected to the potential whirlpool mixing transactions
			// - from the results, check if any transaction can be classified as a whirlpool origin transaction
			// - persist classifications of all origin-mixing pairs
			classifiedTransactions := classifyWhirlpoolOriginTransactions(potWhirlpoolOrigins, originsToMixingMap)
			if len(classifiedTransactions) > 0 {
				if err = db.UpdateTransactions(ctx, c, classifiedTransactions); err != nil {
					return false, err
				}
			}
		}
	}

	// step 3: set the transaction type of
	// - wasabi 2.0 origin transactions
	// - wasabi 2.0 destination transactions
	// - whirlpool destination transactions
	if err = btc.ClassifyDestinationAndOriginsByBlock(ctx, c, from, to); err != nil {
		return false, err
	}

	return true, nil
}

// classifyWhirlpoolOriginTransactions classifies the given origin transactions
// and return them with their connected mixing transactions
func classifyWhirlpoolOriginTransactions(origins []db.Transaction, originToMixingMap map[string][]string) []db.Transaction {
	var classifiedTransactions []db.Transaction
	confirmedMixingTransactions := map[string]bool{}
	for _, whirlpoolOrigin := range origins {
		if isWhirlpoolOrigin(whirlpoolOrigin) {
			classifiedTransactions = append(classifiedTransactions, db.Transaction{UID: whirlpoolOrigin.UID,
				Type: constants.TypeWhirlpoolOrigin})
			mixingTxs := originToMixingMap[whirlpoolOrigin.UID]
			// make sure mixing uids are unique
			for _, m := range mixingTxs {
				confirmedMixingTransactions[m] = true
			}
		}
	}

	// add mixing transactions to set of transactions which are going to be persisted
	for m := range confirmedMixingTransactions {
		classifiedTransactions = append(classifiedTransactions, db.Transaction{UID: m, Type: constants.TypeWhirlpoolMixing})
	}

	return classifiedTransactions
}

// classifyTransactions detects mixing transactions and sets the transaction type appropriately
// The returned slice contains all classified transactions or nil if no classified transactions have been found.
func classifyTransactions(ctx context.Context, c external.Database,
	transactions []db.Transaction) (wasabi2Mixing []db.Transaction, whirlpoolMixingUIDs []string, err error) {
	for _, transaction := range transactions {
		// only do classification for non-classified transactions
		if transaction.Type != "" {
			continue
		}

		if isWasabi2Mixing(ctx, c, transaction) {
			wasabi2Mixing = append(wasabi2Mixing, db.Transaction{UID: transaction.UID, Type: constants.TypeWasabi2Mixing})
			continue
		}

		if isWhirlpoolMixing(transaction) {
			whirlpoolMixingUIDs = append(whirlpoolMixingUIDs, transaction.UID)
			continue
		}
	}
	return
}

// isWasabi2MixingProperties checks if the transaction is a wasabi 2.0 mixing transaction
// based on the direct properties of the transaction
func isWasabi2MixingProperties(t db.Transaction) bool {
	// paper suggest a minimum of 50 inputs, but data shows that transactions with only 15 exist.
	// also set minimum of outputs to 10.
	if len(t.Inputs) < 15 || len(t.Outputs) < 10 {
		return false
	}

	// minimum input output
	const vMin = int64(5000)
	if slices.ContainsFunc(t.Inputs, func(output db.Output) bool {
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
	if float64(outputDenominationCount) < float64(len(t.Inputs))/aMax {
		return false
	}

	// number of output denominations must be at least half of the number of outputs
	if float64(outputDenominationCount) < float64(len(t.Outputs)-1)/2 {
		return false
	}

	// exclude if transaction only contains common denominations (multiple of 5000)
	for i, d := range denominationOut {
		if d > 0 && denominationsTypesWasabi2[i]%5000 != 0 {
			return true
		}
	}

	return false
}

// isWasabi2Mixing checks if the transaction is a wasabi 2.0 mixing transaction
// credit to paper: "Heuristics for Detecting CoinJoin Transactions
// on the Bitcoin Blockchain" https://arxiv.org/abs/2311.12491
func isWasabi2Mixing(ctx context.Context, c external.Database, t db.Transaction) bool {
	if !isWasabi2MixingProperties(t) {
		return false
	}

	// The paper states that each output script should be unique.
	// Instead, we check that each output address is unique, which is a stronger assumption.
	outputCount, err := db.GetOutputAddressCounts(ctx, c, t.UID)
	if err != nil {
		return false
	}

	// output addresses must be unique
	return outputCount == len(t.Outputs)
}

// isWhirlpoolMixing checks if the transaction is a whirlpool mixing transaction
// credit to paper: "Heuristics for Detecting CoinJoin Transactions
// on the Bitcoin Blockchain" https://arxiv.org/abs/2311.12491
func isWhirlpoolMixing(t db.Transaction) bool {
	// "Surge Cycles" increased max number of outputs to 8
	// source: https://medium.com/samourai-wallet/introducing-whirlpool-surge-cycles-b5b484a1670f
	const minOutputs = 5
	const maxOutputs = 8
	if len(t.Inputs) != len(t.Outputs) || len(t.Inputs) < minOutputs || len(t.Inputs) > maxOutputs {
		return false
	}

	numOutputs := len(t.Inputs)

	denominationOut := countWhirlpoolDenominations(t.Outputs)
	denominationIndex := -1
	for i, outputDenominationCount := range denominationOut {
		// there must be only one denomination type, which must be all 5 outputs
		if outputDenominationCount == numOutputs {
			denominationIndex = i
			break
		}
	}

	// no denomination has the required number of occurrences
	if denominationIndex == -1 {
		return false
	}

	// there must be at least one non-denomination input
	denominationIn := countWhirlpoolDenominations(t.Inputs)
	//nolint:gosec
	if denominationIn[denominationIndex] == numOutputs {
		return false
	}

	for _, input := range t.Inputs {
		if input.Amount == nil {
			return false
		}

		if !isAmountWhirlpoolDenominationPlusE(*input.Amount, denominationTypesWhirlpool[denominationIndex], 0) {
			return false
		}
	}

	return true
}

// isAmountWhirlpoolDenominationPlusE returns true if it is close to the given whirlpool denomination.
// set minDiff to > 0 if the amount must not be equal to the denomination
func isAmountWhirlpoolDenominationPlusE(amount int64, denomination int64, minDiff int64) bool {
	diff := amount - denomination
	// diff must be between 0 and 100000 and not be higher than the denomination itself
	return diff <= 100000 && diff >= minDiff && diff <= denomination
}

// isWhirlpoolOrigin checks if the transaction is a whirlpool origin transaction
// credit to paper: "Heuristics for Detecting CoinJoin Transactions
// on the Bitcoin Blockchain" https://arxiv.org/abs/2311.12491
func isWhirlpoolOrigin(t db.Transaction) bool {
	if len(t.Inputs) == 0 || len(t.Outputs) < 3 {
		return false
	}

	amountCounts := map[int64]int{}
	hasOutputWithoutAmount := false
	for _, output := range t.Outputs {
		if output.Amount == nil || *output.Amount == 0 {
			if hasOutputWithoutAmount {
				// can not have more than one output without amount
				return false
			}

			hasOutputWithoutAmount = true
			continue
		}

		// 100 satoshi is the minimum fee per denomination
		if isAnyWhirlpoolDenominationPlusE(*output.Amount, 100) {
			amountCounts[*output.Amount] = amountCounts[*output.Amount] + 1
		}
	}

	// does not have zero amount or no amount close to a denomination
	if !hasOutputWithoutAmount || len(amountCounts) == 0 {
		return false
	}

	// find most frequent amount
	var highestCount int
	var highestCountAmount int64
	for amount, amountCount := range amountCounts {
		if amountCount > highestCount {
			highestCount = amountCount
			highestCountAmount = amount
		} else if amountCount == highestCount && amount > highestCountAmount {
			highestCountAmount = amount
		}
	}

	// 3: fee, change  and data output, change might not always be present
	return highestCount >= len(t.Outputs)-3
}

func isAnyWhirlpoolDenominationPlusE(amount int64, minDiff int64) bool {
	for _, denomination := range denominationTypesWhirlpool {
		if isAmountWhirlpoolDenominationPlusE(amount, denomination, minDiff) {
			return true
		}
	}
	return false
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

// countWhirlpoolDenominations returns for each denomination how often it occurred in the given outputs
func countWhirlpoolDenominations(outputs []db.Output) [NumWhirlpoolDenominations]int {
	amounts := make([]int64, len(outputs))

	for i, o := range outputs {
		if o.Amount == nil {
			return [NumWhirlpoolDenominations]int{}
		}
		amounts[i] = *o.Amount
	}
	return CountAmountWhirlpoolDenominations(amounts)
}

// CountAmountWhirlpoolFuzzyDenominations returns the number of occurrences of each denomination in the given amounts
func CountAmountWhirlpoolFuzzyDenominations(amounts []int64, minDiff int64) (denominations [NumWhirlpoolDenominations]int) {
	for _, amt := range amounts {
	inner:
		for i, v := range denominationTypesWhirlpool {
			if isAmountWhirlpoolDenominationPlusE(amt, v, minDiff) {
				denominations[i]++
				break inner
			}
		}
	}

	return
}

// CountAmountWhirlpoolDenominations returns the number of occurrences of each denomination in the given amounts
func CountAmountWhirlpoolDenominations(amounts []int64) (denominations [NumWhirlpoolDenominations]int) {
	for _, amt := range amounts {
	inner:
		for i, v := range denominationTypesWhirlpool {
			if v == amt {
				denominations[i]++
				break inner
			}
		}
	}

	return
}
