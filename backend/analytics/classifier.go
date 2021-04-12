package analytics

import (
	"backend/blockIterator"
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db/analytics"
	op "backend/db/output"
	dbstat "backend/db/status"
	dbtx "backend/db/transaction"
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
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

type Classifier struct {
	config Config
	db     *dgo.Dgraph
	ctx    context.Context
	state  blockIterator.State
}

// NewClassifier creates a new Classifier object
func NewClassifier(ctx context.Context, dgraph *dgo.Dgraph, cfg Config) *Classifier {
	return &Classifier{
		config: cfg,
		db:     dgraph,
		ctx:    ctx,
	}
}

func (a *Classifier) Name() string {
	return "classifier"
}

func (a *Classifier) Logger() *log.Logger {
	return analyticsLogger
}

func (a *Classifier) State() blockIterator.State {
	return a.state
}

func (a *Classifier) SetState(newState blockIterator.State) {
	a.state = newState
}

func (a *Classifier) Context() context.Context {
	return a.ctx
}

func (a *Classifier) Db() *dgo.Dgraph {
	return a.db
}

func (a *Classifier) IncrementState() {
	a.state.Id++
}

// Empty checks if there is work in the queue or more block above the current one
func (a *Classifier) Empty() bool {
	return a.state.Id > a.state.Top
}

func (a *Classifier) CalculateInitialState() error {
	if !a.config.IsClassifyingEnabled {
		return errors.New("classifying is disabled per configuration")
	}

	if err := dbstat.SetClassifying(a.db, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if err := setInitialClassifierId(a.db, a.config.ClassifierStartBlock); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	crawlerStatus, err := dbstat.GetCrawlerStatus(a.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	classifierStatus, err := dbstat.GetClassifierStatus(a.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if classifierStatus.LastClassifiedBlockId == nil {
		return errors.New("error last classified block is not set")
	}

	var state blockIterator.State

	state.Id = *classifierStatus.LastClassifiedBlockId + 1

	if crawlerStatus.LastBlockId == nil {
		// nothing crawled yet, so set Top to a lower number as Id
		state.Top = *classifierStatus.LastClassifiedBlockId
	} else if *crawlerStatus.LowestBlockId > state.Id {
		// happens the crawler is started with a high start block id in block range mode
		state.Id = *crawlerStatus.LowestBlockId
		state.Top = *crawlerStatus.LastBlockId
	} else {
		// this is the usual case: Set Top to the current last crawled block height
		state.Top = *crawlerStatus.LastBlockId
	}

	a.state = state

	return nil
}

// getUids return uid slice
func getUids(txs []dbtx.Transaction) []string {
	var uids []string
	for _, t := range txs {
		uids = append(uids, t.Uid)
	}
	return uids
}

// getConnectedCollaterals returns two sets of collateral transactions which are connected to the given transaction set.
func getConnectedCollaterals(dgraph *dgo.Dgraph, potentialCollateralTransactions []dbtx.Transaction,
	blockHeight uint64) (originCC []dbtx.Transaction, originCP []dbtx.Transaction, err error) {
	for len(potentialCollateralTransactions) > 0 {
		mixing, cc, cp, getErr := getPrivacyTransactions(dgraph, potentialCollateralTransactions)
		if getErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), getErr)
			return
		}

		// no mixing transaction should be recognized in this step
		if len(mixing) > 0 {
			err = errors.New("error mixing transaction after secondary classification loop")
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
		potentialCollateralTransactions, dbErr = dbtx.GetInputTransactions(dgraph, txUids, blockHeight)
		if dbErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), dbErr)
			return
		}
	}

	return
}

// Iterate does the classification for all transactions of the current block. Transactions are
// classified based on their own properties (number of outputs/inputs, amounts, fee, etc...)
// and how they are connected to other transactions.
func (a *Classifier) Iterate() (bool, error) {
	if a.Empty() {
		return false, errors.New("got empty state")
	}

	// get the transaction of the current block height
	transactions, err := dbtx.GetTransactionByBlock(a.db, a.state.Id)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// step 1: classify all transactions of the current block locally based on their own properties
	mixingTransactions, ccTransactions, cpTransactions, err := getPrivacyTransactions(a.db, transactions)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// the classifications of step 1 are in some cases only indications of the classifications.
	// step 2: either insert the classified directly into the db or only if they are connected
	// to a certain type of transactions

	// step 2.1: set the privacy type of mixing transactions.
	if len(mixingTransactions) > 0 {
		if updateErr := dbtx.UpdateTransactions(a.db, mixingTransactions); updateErr != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), updateErr)
		}
	}

	// step 2.2.1: set the privacy type of destination transactions by analyzing the connected transactions.
	// Origins are only returned in this step and not set directly, if the number of potentialCollateralTransactions
	// is bigger zero. This is so the classification is resilient against sudden shutdowns. If the origins were
	// set directly, the iteration after a fault would not find any potentialCollateralTransactions. Thus the
	// origins are set in step 2.2.2
	potentialCollateralTransactions, foundOrigins, classErr := analytics.DoClassification(a.db, a.state.Id)
	if classErr != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), classErr)
	}

	// if no potentialCollateralTransactions were found, then the origins are already set
	if len(potentialCollateralTransactions) > 0 {
		// step 2.2.2: if potential collateral transaction (connected to origin transactions) have
		// been found they are getting classified, before appending them to the set of transactions
		// which is getting inserted into the db
		originCC, originCP, err := getConnectedCollaterals(a.db, potentialCollateralTransactions, a.state.Id)
		if err != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		var updatedTransactions []dbtx.Transaction

		for _, o := range foundOrigins {
			updatedTransactions = append(updatedTransactions, newOriginTransaction(o.Uid))
		}

		updatedTransactions = append(updatedTransactions, originCC...)
		updatedTransactions = append(updatedTransactions, originCP...)

		if len(updatedTransactions) > 0 {
			if updateErr := dbtx.UpdateTransactions(a.db, updatedTransactions); updateErr != nil {
				return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), updateErr)
			}
		}
	}

	// step 2.3: set collateral creation type
	if len(ccTransactions) > 0 {
		var insertedSum uint64
		var numInserted uint64 = 1
		var ccErr error

		// need to set type multiple times for the same block as transactions
		// could be connected to transactions in the same block
		for numInserted > 0 {
			numInserted, ccErr = analytics.SetCollateralCreation(a.db, getUids(ccTransactions))
			if ccErr != nil {
				return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ccErr)
			}

			insertedSum += numInserted
			// all inserted -> no need for a second round
			if insertedSum == uint64(len(ccTransactions)) {
				break
			}
		}
	}

	// step 2.4: set collateral creation type
	if len(cpTransactions) > 0 {
		var insertedSum uint64
		var numInserted uint64 = 1
		var cpErr error

		// need to set type multiple times for the same block as transactions
		// could be connected to transactions in the same block
		for numInserted > 0 {
			numInserted, cpErr = analytics.SetCollateralPayment(a.db, getUids(cpTransactions))
			if cpErr != nil {
				return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), cpErr)
			}

			insertedSum += numInserted
			// all inserted -> no need for a second round
			if insertedSum == uint64(len(cpTransactions)) {
				break
			}
		}
	}

	// set the last classified block
	if statusErr := dbstat.SetLastClassifiedBlockId(a.db, a.state.Id); statusErr != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), statusErr)
	}

	return true, nil
}

func (a *Classifier) PostExecution() error {
	if err := dbstat.SetClassifying(a.db, false); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// setInitialClassifierId sets the starting classifier block id to the
// value of startBlockClassifier if no value has been set yet
func setInitialClassifierId(dgraph *dgo.Dgraph, startBlockClassifier uint64) (err error) {
	status, err := dbstat.GetClassifierStatus(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if status.LastClassifiedBlockId == nil ||
		*status.LastClassifiedBlockId < startBlockClassifier {
		if err = dbstat.SetLastClassifiedBlockId(dgraph, startBlockClassifier); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}
	return
}

// isCollateralCreation checks if the transactions is a collateral creation transaction
func isCollateralCreation(dgraph *dgo.Dgraph, t dbtx.Transaction) (bool, error) {
	if *t.Fee == 0 || len(t.Inputs) < 1 || len(t.Outputs) != 2 {
		return false, nil
	}

	// must have at least enough to pay MaxCollateral
	if *t.Outputs[0].Amount+*t.Outputs[1].Amount < op.MaxCollateral {
		return false, nil
	}

	// check if both outputs do not fulfill the minimum collateral amount
	if *t.Outputs[0].Amount < op.MinCollateral && *t.Outputs[1].Amount < op.MinCollateral {
		return false, nil
	}

	inputCount, outputCount, err := dbtx.GetOutputAddressCounts(dgraph, t.Uid)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// inputs must be from the same address and outputs must go to different addresses
	if inputCount != 1 || outputCount == 1 {
		return false, nil
	}

	return true, nil
}

// newCollateralPaymentTransaction returns a new collateral creation transaction with the given uid
func newCollateralCreationTransaction(uid string) dbtx.Transaction {
	return dbtx.Transaction{Uid: uid, PrivacyType: constants.PrivacyCollateralCreation}
}

// isCollateralPayment checks if the transactions is a collateral payment transaction
func isCollateralPayment(t dbtx.Transaction) bool {
	if *t.Fee == 0 || len(t.Inputs) != 1 || len(t.Outputs) != 1 {
		return false
	}

	// must be able to pay at least the minimum fee
	if *t.Inputs[0].Amount < op.MinCollateral || *t.Fee < op.MinCollateral {
		return false
	}

	// if the fee or amount is too big it is not a collateral payment
	if *t.Fee > op.OldMaxCollateral*2 || *t.Inputs[0].Amount > op.OldMaxCollateral*2 {
		return false
	}

	return true
}

// newCollateralPaymentTransaction returns a new collateral payment transaction with the given uid
func newCollateralPaymentTransaction(uid string) dbtx.Transaction {
	return dbtx.Transaction{Uid: uid, PrivacyType: constants.PrivacyCollateralPayment}
}

// isMixing checks if the transactions is a mixing transaction
func isMixing(t dbtx.Transaction) bool {
	// At least 3 clients per mixing transaction -> >2 inputs/outputs
	// Maximal 9 inputs per client and a maximum of 20 clients in one mixing transaction -> 180 inputs/outputs
	if *t.Fee != 0 || len(t.Inputs) < 3 || len(t.Outputs) < 3 ||
		len(t.Inputs) != len(t.Outputs) || len(t.Inputs) > 180 {
		return false
	}

	denomIn := op.CountOutputDenominations(t.Inputs)
	denomOut := op.CountOutputDenominations(t.Outputs)

	sum := 0
	for _, v := range denomIn {
		sum += v
	}
	if sum == 0 {
		return false
	}
	sum = 0
	for _, v := range denomOut {
		sum += v
	}
	if sum == 0 {
		return false
	}
	for i := range denomIn {
		if denomIn[i] != denomOut[i] {
			return false
		}
	}
	return true
}

// newMixingTransaction returns a new mixing transaction with the given uid
func newMixingTransaction(uid string) dbtx.Transaction {
	return dbtx.Transaction{Uid: uid, PrivacyType: constants.PrivacyMixing}
}

// newOriginTransaction returns a new origin transaction with the given uid
func newOriginTransaction(uid string) dbtx.Transaction {
	return dbtx.Transaction{Uid: uid, PrivacyType: constants.PrivacyOrigin}
}
