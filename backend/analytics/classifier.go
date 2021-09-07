package analytics

import (
	"backend/blockiterator"
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db/analytics"
	op "backend/db/output"
	dbstat "backend/db/status"
	dbtx "backend/db/transaction"
	"backend/external"

	"context"
	"errors"
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

// Classifier implements BlockIterator which classifies the transactions of each traversed block
type Classifier struct {
	config       Config
	db           external.Database
	ctx          context.Context
	state        blockiterator.State
	blocks       prometheus.Counter
	transactions prometheus.Counter
	blockHeight  prometheus.Gauge
}

// NewClassifier creates a new Classifier object
func NewClassifier(ctx context.Context, dgraph external.Database, cfg Config) *Classifier {
	return &Classifier{
		config: cfg,
		db:     dgraph,
		ctx:    ctx,
		blocks: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_classifier_blocks_processed_total",
			Help: "The total number of blocks processed by the classifier",
		}),
		transactions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_classifier_transactions_processed_total",
			Help: "The total number of transactions processed by the classifier",
		}),
		blockHeight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "dakar_classifier_last_block",
			Help: "The last processed block by the classifier",
		}),
	}
}

// Name returns the name
func (c *Classifier) Name() string {
	return "classifier"
}

// Logger returns the Logger
func (c *Classifier) Logger() *log.Logger {
	return analyticsLogger
}

// Context returns the context
func (c *Classifier) Context() context.Context {
	return c.ctx
}

// IncrementState increments the state one block
func (c *Classifier) IncrementState() error {
	c.state.ID++
	return nil
}

// Empty checks if there are more blocks above the current one
func (c *Classifier) Empty() bool {
	return c.state.ID > c.state.Top
}

// CalculateInitialState calculates the state on which the iterator starts processing
func (c *Classifier) CalculateInitialState() error {
	if !c.config.IsClassifyingEnabled {
		return errors.New("classifying is disabled per configuration")
	}

	if err := dbstat.SetClassifying(c.db, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if err := setInitialClassifierID(c.db, c.config.ClassifierStartAfterBlock); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	crawlerStatus, err := dbstat.GetCrawlerStatus(c.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	classifierStatus, err := dbstat.GetClassifierStatus(c.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if classifierStatus.LastClassifiedBlockID == nil {
		return errors.New("error last classified block is not set")
	}

	var state blockiterator.State

	state.ID = *classifierStatus.LastClassifiedBlockID + 1

	if crawlerStatus.LastBlockID == nil {
		// nothing crawled yet, so set Top to a lower number as ID
		state.Top = *classifierStatus.LastClassifiedBlockID
	} else {
		// this is the usual case: Set Top to the current last crawled block height
		state.Top = *crawlerStatus.LastBlockID
	}

	c.state = state

	// state.ID - 1 because the ID is the next block
	c.blockHeight.Set(float64(c.state.ID - 1))

	return nil
}

// getUids return uid slice
func getUids(txs []dbtx.Transaction) []string {
	var uids []string
	for _, t := range txs {
		uids = append(uids, t.UID)
	}
	return uids
}

// getConnectedCollaterals returns two sets of collateral transactions which are connected to the given transaction set.
func getConnectedCollaterals(dgraph external.Database, potentialCollateralTransactions []dbtx.Transaction,
	blockHeight uint64) (originCC []dbtx.Transaction, originCP []dbtx.Transaction, err error) {
	for len(potentialCollateralTransactions) > 0 {
		mixing, cc, cp, getErr := classifyTransactions(dgraph, potentialCollateralTransactions)
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
		potentialCollateralTransactions, dbErr = analytics.GetCollateralInputTransactions(dgraph, txUids, blockHeight)
		if dbErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), dbErr)
			return
		}
	}

	return
}

// NextBlock tries to increase the internal state to the next block
func (c *Classifier) NextBlock() (bool, error) {
	status, err := dbstat.GetCrawlerStatus(c.db)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	} else if status.LastBlockID == nil {
		return false, errors.New("last crawled block is not set")
	}

	if c.state.ID <= *status.LastBlockID {
		c.state.Top = *status.LastBlockID
		return true, nil
	}

	return false, nil
}

// CurrentBlock returns the height of the block which is currently classified
func (c *Classifier) CurrentBlock() uint64 {
	return c.state.ID
}

// Iterate classifies all transactions of the current block based
// on their own properties (number of outputs/inputs, amounts, fee, etc...)
// and how they are connected to other transactions.
func (c *Classifier) Iterate() (bool, error) {
	if c.Empty() {
		return false, errors.New("got empty state")
	}

	// get the transaction of the current block height
	transactions, err := dbtx.GetTransactionByBlock(c.db, c.state.ID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// step 1: classify all transactions of the current block locally based on their own properties
	mixingTransactions, ccTransactions, cpTransactions, err := classifyTransactions(c.db, transactions)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// the classifications of step 1 are in some cases only indications of the classifications.
	// step 2: either insert the classified directly into the db or only if they are connected
	// to a certain type of transactions

	// step 2.1: set the privacy type of mixing transactions.
	if len(mixingTransactions) > 0 {
		if updateErr := dbtx.UpdateTransactions(c.db, mixingTransactions); updateErr != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), updateErr)
		}
	}

	// step 2.2.1: set the privacy type of destination transactions by analyzing the connected transactions.
	// Origins are only returned in this step and not set directly, if the number of potentialCollateralTransactions
	// is bigger zero. This is so the classification is resilient against sudden shutdowns. If the origins were
	// set directly, the iteration after a fault would not find any potentialCollateralTransactions. Thus, the
	// origins are set in step 2.2.2
	potentialCollateralTransactions, foundOrigins,
		classErr := analytics.ClassifyDestinationAndOriginsByBlock(c.db, c.state.ID)
	if classErr != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), classErr)
	}

	// if no potentialCollateralTransactions were found, then the origins are already set
	if len(potentialCollateralTransactions) > 0 {
		// step 2.2.2: if potential collateral transaction (connected to origin transactions) have
		// been found they are getting classified, before appending them to the set of transactions
		// which is getting inserted into the db
		originCC, originCP, err := getConnectedCollaterals(c.db, potentialCollateralTransactions, c.state.ID)
		if err != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		var updatedTransactions []dbtx.Transaction

		for _, o := range foundOrigins {
			updatedTransactions = append(updatedTransactions, newOriginTransaction(o.UID))
		}

		updatedTransactions = append(updatedTransactions, originCC...)
		updatedTransactions = append(updatedTransactions, originCP...)

		if len(updatedTransactions) > 0 {
			if updateErr := dbtx.UpdateTransactions(c.db, updatedTransactions); updateErr != nil {
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
			numInserted, ccErr = analytics.SetCollateralCreation(c.db, getUids(ccTransactions))
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
			numInserted, cpErr = analytics.SetCollateralPayment(c.db, getUids(cpTransactions))
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
	if statusErr := dbstat.SetLastClassifiedBlockID(c.db, c.state.ID); statusErr != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), statusErr)
	}

	c.blocks.Inc()
	c.transactions.Add(float64(len(mixingTransactions)))
	c.blockHeight.Set(float64(c.state.ID))

	return true, nil
}

// PostExecution sets the classifier status activity flag to false
func (c *Classifier) PostExecution() error {
	if err := dbstat.SetClassifying(c.db, false); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// setInitialClassifierID sets the starting classifier block id to the
// value of startBlockClassifier if no value has been set yet
func setInitialClassifierID(dgraph external.Database, startBlockClassifier uint64) (err error) {
	status, err := dbstat.GetClassifierStatus(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if status.LastClassifiedBlockID == nil ||
		*status.LastClassifiedBlockID < startBlockClassifier {
		if err = dbstat.SetLastClassifiedBlockID(dgraph, startBlockClassifier); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}
	return
}

// isCollateralCreation checks if the transactions is a collateral creation transaction
func isCollateralCreation(dgraph external.Database, t dbtx.Transaction) (bool, error) {
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

	inputCount, outputCount, err := dbtx.GetOutputAddressCounts(dgraph, t.UID)
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
	pt := constants.PrivacyCollateralCreation
	return dbtx.Transaction{UID: uid, PrivacyType: &pt}
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
	pt := constants.PrivacyCollateralPayment
	return dbtx.Transaction{UID: uid, PrivacyType: &pt}
}

// isMixing checks if the transactions is a mixing transaction
// -1: not a mixing transaction
// 0-4: denomination type
func isMixing(t dbtx.Transaction) int {
	// At least 3 clients per mixing transaction -> more than 2 inputs/outputs
	// Maximal 9 inputs per client and a maximum of 20 clients in one mixing transaction -> 180 inputs/outputs
	if *t.Fee != 0 || len(t.Inputs) < 3 || len(t.Inputs) != len(t.Outputs) || len(t.Inputs) > 180 {
		return -1
	}

	denominationIn := op.CountOutputDenominations(t.Inputs)
	denominationOut := op.CountOutputDenominations(t.Outputs)
	denominationIndex := -1
	for i := range denominationIn {
		// inputs and outputs should have the same amount of each denomination type
		if denominationIn[i] != denominationOut[i] {
			return -1
		}

		if denominationIn[i] > 0 {
			// there is more than one denomination type
			if denominationIndex >= 0 {
				return -1
			}
			// the number of denominations should be the same as the inputs/outputs
			if denominationIn[i] != len(t.Inputs) {
				return -1
			}
			denominationIndex = i
		}
	}

	return denominationIndex
}

// newMixingTransaction returns a new mixing transaction with the given type and uid.
// bit must be a value between 0 and 4
func newMixingTransaction(uid string, bit int) dbtx.Transaction {
	pt := constants.MixingTypes[bit]
	return dbtx.Transaction{UID: uid, PrivacyType: &pt}
}

// newOriginTransaction returns a new origin transaction with the given uid
func newOriginTransaction(uid string) dbtx.Transaction {
	pt := constants.PrivacyOrigin
	return dbtx.Transaction{UID: uid, PrivacyType: &pt}
}

// hasValidPrivacyType check is the transaction has a valid privacy type
func hasValidPrivacyType(tx dbtx.Transaction) bool {
	t := tx.PrivacyType
	return t != nil && *t <= constants.PrivacyCollateralPaymentLast
}

// classifyTransactions detects mixing and collateral creation transactions and sets the privacy type appropriately
// The returned slice contains all classified transactions or nil if no privacy transactions have been found.
func classifyTransactions(dgraph external.Database, transactions []dbtx.Transaction) (mixing []dbtx.Transaction,
	cc []dbtx.Transaction, cp []dbtx.Transaction, err error) {
	for _, transaction := range transactions {
		// only do classification for non-classified transactions
		if hasValidPrivacyType(transaction) {
			continue
		}

		if dIndex := isMixing(transaction); dIndex >= 0 {
			mixing = append(mixing, newMixingTransaction(transaction.UID, dIndex))
			continue
		}

		if isCollateralPayment(transaction) {
			cp = append(cp, newCollateralPaymentTransaction(transaction.UID))
			continue
		}

		isCC, collateralErr := isCollateralCreation(dgraph, transaction)
		if collateralErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), collateralErr)
		}

		if isCC {
			cc = append(cc, newCollateralCreationTransaction(transaction.UID))
			continue
		}
	}
	return
}
