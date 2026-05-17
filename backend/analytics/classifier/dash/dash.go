// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package dash

import (
	"backend/constants"
	"backend/db"
	"backend/db/analytics/classifier/dash"
	"backend/external"
	"context"
	"slices"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// ------------------------- Private Send Example Graph -------------------------
//
// Time ─────────────────────────────────────────────────────────────────────────►
//
//                                                 ┌──────────┐   ┌─────────┐
//                                                 │C Creation├───┤C Payment│
//                                                 └───┬──────┘   └─────────┘
//                                                     │
//        ┌──────────┐ ┌─────────┐ ┌─────────┐ ┌───────┴──┐  ┌─────────┐
//      ┌─┤C Creation├─┤C Payment├─┤C Payment│ │C Creation├──┤C Payment│
//      │ └──────────┘ └─────────┘ └─────────┘ └────┬─────┘  └─────────┘
//      │                                           │
//  ┌───┴──┐        ┌──────┐      ┌──────┐       ┌──┴───┐          ┌───────────┐
//  │Origin├───┬────┤Mixing├──┬───┤Mixing├───┬───┤Mixing├──────────┤Destination│
//  └──────┘   │    └──────┘  │   └──────┘   │   └──────┘          └───────────┘
//             │              │              │
//             │              │              │
//             │    ┌──────┐  │   ┌──────┐   │   ┌──────┐
//             └────┤Mixing├──┼───┤Mixing├───┼───┤Mixing├──────┐
//                  └──────┘  │   └──────┘   │   └──────┘      │
//                            │              │                 │
//                            │              │                 │
//  ┌──────┐        ┌──────┐  │   ┌──────┐   │   ┌──────┐      │   ┌───────────┐
//  │Origin├────────┤Mixing├──┴───┤Mixing├───┴───┤Mixing├──────┴───┤Destination│
//  └──┬───┘        └──────┘      └───┬──┘       └──────┘          └───────────┘
//     │                              │
//   ┌─┴───────┐ ┌─────────┐          │ ┌──────────┐ ┌─────────┐
//   │C Payment├─┤C Payment│          └─┤C Creation├─┤C Payment│
//   └─────────┘ └─────────┘            └──────────┘ └─────────┘

const (
	// minCollateral is 1/10 of the smallest denomination: round(100001/10).
	minCollateral = 10000
	// OldMinCollateral is the minimum collateral before the 5th denomination
	// was added in protocol version 70213 it was round(1000010/10): 100000
	// OldMinCollateral = 100000
	// maxCollateral is the maximum allowed collateral
	maxCollateral = 40000 // 4*minCollateral
	// oldMaxCollateral is to old collateral
	oldMaxCollateral = 400000 // 4*OldMinCollateral
	// NumDenominations is the number of Dash PrivateSend denominations
	NumDenominations = 5
)

var denominationsTypes = [NumDenominations]int64{1000010000, 100001000, 10000100, 1000010, 100001}

// Iterate returns
// - true when iterating should continue
// - false when not
func Iterate(ctx context.Context, c external.Database, from int64, to int64) (bool, error) {
	// get the transaction of the current block range
	transactions, err := db.GetTransactionsByBlock(ctx, c, from, to, nil)
	if err != nil {
		return false, err
	}

	// step 1: classify all transactions of the current block locally based on their own properties
	mixingTransactions, ccTransactions, cpTransactions, err := classifyTransactions(ctx, c, transactions)
	if err != nil {
		return false, err
	}

	// the classifications of step 1 are in some cases only indications of the true classifications.
	// step 2: either insert the classified directly into the db or only if they are connected
	// to a certain type of transactions

	// step 2.1: store the transaction type of mixing transactions.
	if len(mixingTransactions) > 0 {
		if updateErr := db.UpdateTransactions(ctx, c, mixingTransactions); updateErr != nil {
			return false, updateErr
		}
	}

	// step 2.2.1: set the transaction type of destination transactions by analyzing the connected transactions.
	// Origins are only returned in this step and not set directly, if the number of potentialCollateralTransactions
	// is bigger than zero. This is so the classification is resilient against sudden shutdowns. If the origins were
	// set directly, the iteration after a fault would not find any potentialCollateralTransactions. Thus, the
	// origins are set in step 2.2.2
	potentialCollateralTransactions, foundOrigins,
		classErr := dash.ClassifyDestinationAndOriginsByBlock(ctx, c, from, to)
	if classErr != nil {
		return false, classErr
	}

	// if no potentialCollateralTransactions were found, then the origins are already set
	if len(potentialCollateralTransactions) > 0 {
		// step 2.2.2: if potential collateral transaction (connected to origin transactions) have
		// been found they are getting classified, before appending them to the set of transactions
		// which is getting inserted into the db
		originCC, originCP, err := getConnectedCollaterals(ctx, c, potentialCollateralTransactions, to)
		if err != nil {
			return false, err
		}

		updatedTransactions := make([]db.Transaction, len(foundOrigins))
		for i, o := range foundOrigins {
			updatedTransactions[i] = newOriginTransaction(o.UID)
		}

		updatedTransactions = slices.Concat(updatedTransactions, originCC, originCP)

		if len(updatedTransactions) > 0 {
			if updateErr := db.UpdateTransactions(ctx, c, updatedTransactions); updateErr != nil {
				return false, updateErr
			}
		}
	}

	// step 2.3: set collateral creation type
	if len(ccTransactions) > 0 {
		var insertedSum = 0
		var numInserted = 1
		var ccErr error

		// need to set type multiple times for the same block as transactions
		// could be connected to transactions in the same block
		for numInserted > 0 {
			numInserted, ccErr = dash.SetCollateralCreation(ctx, c, getUids(ccTransactions))
			if ccErr != nil {
				return false, ccErr
			}

			insertedSum += numInserted
			// all inserted -> no need for a second round
			if insertedSum == len(ccTransactions) {
				break
			}
		}
	}

	// step 2.4: set collateral payment type
	if len(cpTransactions) > 0 {
		var insertedSum = 0
		var numInserted = 1
		var cpErr error

		// need to set type multiple times for the same block as transactions
		// could be connected to transactions in the same block
		for numInserted > 0 {
			numInserted, cpErr = dash.SetCollateralPayment(ctx, c, getUids(cpTransactions))
			if cpErr != nil {
				return false, cpErr
			}

			insertedSum += numInserted
			// all inserted -> no need for a second round
			if insertedSum == len(cpTransactions) {
				break
			}
		}
	}

	return true, nil
}

// CountAmountDenominations returns the number of occurrences of each denomination in the given amounts
func CountAmountDenominations(amounts []int64) (denominations [NumDenominations]int) {
	for _, amt := range amounts {
	inner:
		for i, v := range denominationsTypes {
			if amt == v {
				denominations[i]++
				break inner
			}
		}
	}

	return
}

// countOutputDenominations returns for each denomination how often it occurred in the given outputs
func countOutputDenominations(outputs []db.Output) [NumDenominations]int {
	amounts := make([]int64, len(outputs))

	for i, o := range outputs {
		if o.Amount == nil {
			return [NumDenominations]int{}
		}
		amounts[i] = *o.Amount
	}

	return CountAmountDenominations(amounts)
}

// getUids return uid slice
func getUids(txs []db.Transaction) []string {
	uids := make([]string, len(txs))
	for i, t := range txs {
		uids[i] = t.UID
	}
	return uids
}

// getConnectedCollaterals returns a set of collateral creation and a set of
// collateral payment transactions, which are connected to the given transaction set.
func getConnectedCollaterals(ctx context.Context, dgraph external.Database, potentialCollateralTransactions []db.Transaction, blockHeight int64) (originCC []db.Transaction, originCP []db.Transaction, err error) {
	for len(potentialCollateralTransactions) > 0 {
		mixing, cc, cp, getErr := classifyTransactions(ctx, dgraph, potentialCollateralTransactions)
		if getErr != nil {
			err = getErr
			return
		}

		// no mixing transaction should be recognized in this step
		if len(mixing) > 0 {
			err = serror.FromStr("mixing transaction occurred after secondary classification loop")
			return
		}

		// nothing to do?
		if len(cc)+len(cp) == 0 {
			break
		}

		// append new cc and cp transactions to set which gets inserted into the db later
		originCC = append(originCC, cc...)
		originCP = append(originCP, cp...)

		// extract all uids from the transactions
		txUids := getUids(append(cc, cp...))

		var dbErr error
		potentialCollateralTransactions, dbErr = dash.GetCollateralInputTransactions(ctx, dgraph, txUids, blockHeight)
		if dbErr != nil {
			err = dbErr
			return
		}
	}

	return
}

// isCollateralCreation checks if the transactions is a collateral creation transaction
func isCollateralCreation(ctx context.Context, dgraph external.Database, t db.Transaction) (bool, error) {
	if *t.Fee == 0 || len(t.Inputs) < 1 || len(t.Outputs) != 2 {
		return false, nil
	}

	outputSum := *t.Outputs[0].Amount + *t.Outputs[1].Amount
	// must have at least enough to pay maxCollateral
	if outputSum < maxCollateral {
		return false, nil
	}

	// check if both outputs do not fulfill the minimum collateral amount
	if *t.Outputs[0].Amount < minCollateral && *t.Outputs[1].Amount < minCollateral {
		return false, nil
	}

	// one output must be smaller or equal to the old maximum collateral
	if *t.Outputs[0].Amount > oldMaxCollateral*2 && *t.Outputs[1].Amount > oldMaxCollateral*2 {
		return false, nil
	}

	inputCount, outputCount, err := db.GetInputOutputAddressCounts(ctx, dgraph, t.UID)
	if err != nil {
		return false, err
	}

	// inputs must be from the same address and outputs must go to different addresses
	if inputCount != 1 || outputCount == 1 {
		return false, nil
	}

	return true, nil
}

// newCollateralPaymentTransaction returns a new collateral creation transaction with the given uid
func newCollateralCreationTransaction(uid string) db.Transaction {
	return db.Transaction{UID: uid, Type: constants.TypeDashCC}
}

// isCollateralPayment checks if the transactions is a collateral payment transaction
func isCollateralPayment(t db.Transaction) bool {
	if *t.Fee == 0 || len(t.Inputs) != 1 || len(t.Outputs) != 1 {
		return false
	}

	// must be able to pay at least the minimum fee
	if *t.Inputs[0].Amount < minCollateral || *t.Fee < minCollateral {
		return false
	}

	// if the fee or amount is too big it is not a collateral payment
	if *t.Fee > oldMaxCollateral*2 || *t.Inputs[0].Amount > oldMaxCollateral*2 {
		return false
	}

	return true
}

// newCollateralPaymentTransaction returns a new collateral payment transaction with the given uid
func newCollateralPaymentTransaction(uid string) db.Transaction {
	return db.Transaction{UID: uid, Type: constants.TypeDashCP}
}

// isMixing checks if the transaction is a mixing transaction
func isMixing(t db.Transaction) bool {
	// At least 3 clients per mixing transaction -> more than 2 inputs/outputs
	// Maximal 9 inputs per client and a maximum of 20 clients in one mixing transaction -> 180 inputs/outputs
	if *t.Fee != 0 || len(t.Inputs) < 3 || len(t.Inputs) != len(t.Outputs) || len(t.Inputs) > 180 {
		return false
	}

	denominationIn := countOutputDenominations(t.Inputs)
	denominationOut := countOutputDenominations(t.Outputs)
	foundDenominations := false

	for i := range denominationIn {
		// inputs and outputs should have the same amount of each denomination type
		if denominationIn[i] != denominationOut[i] {
			return false
		}

		if denominationIn[i] > 0 {
			// there is more than one denomination type
			if foundDenominations {
				return false
			}
			// the number of denominations should be the same as the inputs/outputs
			if denominationIn[i] != len(t.Inputs) {
				return false
			}
			foundDenominations = true
		}
	}

	return foundDenominations
}

// newMixingTransaction returns a new mixing transaction with the given type and uid.
func newMixingTransaction(uid string) db.Transaction {
	return db.Transaction{UID: uid, Type: constants.TypeDashMixing}
}

// newOriginTransaction returns a new origin transaction with the given uid
func newOriginTransaction(uid string) db.Transaction {
	return db.Transaction{UID: uid, Type: constants.TypeDashOrigin}
}

// classifyTransactions detects mixing and collateral creation transactions and sets the transaction type appropriately.
// The returned slice contains all classified transactions or nil if no classified transactions have been found.
func classifyTransactions(ctx context.Context, dgraph external.Database,
	transactions []db.Transaction) (mixing []db.Transaction,
	cc []db.Transaction, cp []db.Transaction, err error) {
	for _, transaction := range transactions {
		// only do classification for non-classified transactions
		if transaction.Type != "" {
			continue
		}

		if isMixing(transaction) {
			mixing = append(mixing, newMixingTransaction(transaction.UID))
			continue
		}

		if isCollateralPayment(transaction) {
			cp = append(cp, newCollateralPaymentTransaction(transaction.UID))
			continue
		}

		isCC, collateralErr := isCollateralCreation(ctx, dgraph, transaction)
		if collateralErr != nil {
			err = collateralErr
			return
		}

		if isCC {
			cc = append(cc, newCollateralCreationTransaction(transaction.UID))
			continue
		}
	}
	return
}
