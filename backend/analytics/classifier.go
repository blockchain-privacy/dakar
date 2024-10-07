package analytics

import (
	"backend/blockiterator"
	"backend/constants"
	"backend/db"
	"backend/db/analytics"
	dbstat "backend/db/status"
	"backend/external"
	"context"
	"github.com/qrest/gomisc/serror"
	"log"
	"slices"

	"github.com/prometheus/client_golang/prometheus"
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

// NumDenominations is the number of Dash PrivateSend denominations
const NumDenominations = 5

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
)

var denominationsTypes = [NumDenominations]int64{1000010000, 100001000, 10000100, 1000010, 100001}

// countOutputDenominations returns for each denomination how often it occurred in the given outputs
func countOutputDenominations(outputs []db.Output) [NumDenominations]int {
	amounts := make([]int64, len(outputs))

	for i, o := range outputs {
		if o.Amount == nil {
			log.Println("error amount not set")
			return [NumDenominations]int{}
		}
		amounts[i] = *o.Amount
	}

	return CountAmountDenominations(amounts)
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

// Classifier implements BlockIterator which classifies the transactions of each traversed block
type Classifier struct {
	config Config
	db     external.Database
	ctx    context.Context
	state  blockiterator.State

	// how many blocks are processed in one interation at maximum
	maxBlocks int64
	// number of blocks which have been processed by the last Iterate call
	blocksProcessed int64

	blocks       prometheus.Counter
	transactions prometheus.Counter
	blockHeight  prometheus.Gauge
}

// NewClassifier creates a new Classifier object
func NewClassifier(ctx context.Context, dgraph external.Database, cfg Config) *Classifier {
	return &Classifier{
		config:    cfg,
		db:        dgraph,
		ctx:       ctx,
		maxBlocks: 1,
	}
}

func (c *Classifier) RegisterMetrics(req prometheus.Registerer) {
	c.blocks = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_classifier_blocks_processed_total",
		Help: "The total number of blocks processed by the classifier",
	})
	req.MustRegister(c.blocks)
	c.transactions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_classifier_transactions_processed_total",
		Help: "The total number of transactions processed by the classifier",
	})
	req.MustRegister(c.transactions)
	c.blockHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dakar_classifier_last_block",
		Help: "The last processed block by the classifier",
	})
	req.MustRegister(c.blockHeight)
}

func (c *Classifier) Props() blockiterator.Properties {
	return blockiterator.Properties{
		Name:                        "classifier",
		Context:                     c.ctx,
		Logger:                      analyticsLogger,
		CurrentBlock:                c.state.ID,
		ProcessedBlockCount:         c.blocksProcessed,
		SupportsMultiBlockIteration: true,
	}
}

func (c *Classifier) SetMaxBlocks(max int64) {
	c.maxBlocks = max
}

// IncrementState increments the state one block
func (c *Classifier) IncrementState() error {
	c.state.ID += c.blocksProcessed
	return nil
}

// Empty checks if there are more blocks above the current one
func (c *Classifier) Empty() bool {
	return c.state.ID > c.state.Top
}

// CalculateInitialState calculates the state on which the iterator starts processing
func (c *Classifier) CalculateInitialState(ctx context.Context) error {
	if !c.config.IsClassifyingEnabled {
		return serror.FromStr("classifying is disabled per configuration")
	}

	if err := dbstat.SetClassifying(ctx, c.db, true); err != nil {
		return err
	}

	if err := setInitialClassifierID(ctx, c.db); err != nil {
		return err
	}

	crawlerStatus, err := dbstat.GetCrawlerStatus(ctx, c.db)
	if err != nil {
		return err
	}

	classifierStatus, err := dbstat.GetClassifierStatus(ctx, c.db)
	if err != nil {
		return err
	}

	if classifierStatus.LastClassifiedBlockID == nil {
		return serror.FromStr("last classified block is not set")
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
func getUids(txs []db.Transaction) []string {
	uids := make([]string, len(txs))
	for i, t := range txs {
		uids[i] = t.UID
	}
	return uids
}

// getConnectedCollaterals returns a set of collateral creation and a set of
// collateral payment transactions, which are connected to the given transaction set.
func getConnectedCollaterals(dgraph external.Database, potentialCollateralTransactions []db.Transaction,
	blockHeight int64) (originCC []db.Transaction, originCP []db.Transaction, err error) {
	for len(potentialCollateralTransactions) > 0 {
		mixing, cc, cp, getErr := classifyTransactions(dgraph, potentialCollateralTransactions)
		if getErr != nil {
			err = getErr
			return
		}

		// no mixing transaction should be recognized in this step
		if len(mixing) > 0 {
			err = serror.FromStr("mixing transaction occured after secondary classification loop")
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
			err = dbErr
			return
		}
	}

	return
}

// Next tries to increase the internal state to the next block.
// Returns true if the top block id was changed
func (c *Classifier) Next(ctx context.Context) (bool, error) {
	status, err := dbstat.GetCrawlerStatus(ctx, c.db)
	if err != nil {
		return false, err
	} else if status.LastBlockID == nil {
		return false, serror.FromStr("last crawled block is not set")
	}

	if c.state.ID <= *status.LastBlockID {
		c.state.Top = *status.LastBlockID
		return true, nil
	}

	return false, nil
}

// Iterate classifies all transactions of the current block based
// on their own properties (number of outputs/inputs, amounts, fee, etc...)
// and how they are connected to other transactions.
func (c *Classifier) Iterate(ctx context.Context) (bool, error) {
	if c.maxBlocks == 0 {
		return false, serror.FromStr("max blocks must be higher than zero")
	}

	if c.Empty() {
		return false, serror.FromStr("got empty state")
	}

	// state.ID is a new block already, therefore maxBlocks has to be reduced by 1
	toBlockID := min(c.state.Top, c.state.ID+c.maxBlocks-1)

	// get the transaction of the current block height
	transactions, err := db.GetTransactionsByBlock(c.db, c.state.ID, toBlockID)
	if err != nil {
		return false, err
	}

	// step 1: classify all transactions of the current block locally based on their own properties
	mixingTransactions, ccTransactions, cpTransactions, err := classifyTransactions(c.db, transactions)
	if err != nil {
		return false, err
	}

	// the classifications of step 1 are in some cases only indications of the true classifications.
	// step 2: either insert the classified directly into the db or only if they are connected
	// to a certain type of transactions

	// step 2.1: store the privacy type of mixing transactions.
	if len(mixingTransactions) > 0 {
		if updateErr := db.UpdateTransactions(c.db, mixingTransactions); updateErr != nil {
			return false, updateErr
		}
	}

	// step 2.2.1: set the privacy type of destination transactions by analyzing the connected transactions.
	// Origins are only returned in this step and not set directly, if the number of potentialCollateralTransactions
	// is bigger than zero. This is so the classification is resilient against sudden shutdowns. If the origins were
	// set directly, the iteration after a fault would not find any potentialCollateralTransactions. Thus, the
	// origins are set in step 2.2.2
	potentialCollateralTransactions, foundOrigins,
		classErr := analytics.ClassifyDestinationAndOriginsByBlock(c.db, c.state.ID, toBlockID)
	if classErr != nil {
		return false, classErr
	}

	// if no potentialCollateralTransactions were found, then the origins are already set
	if len(potentialCollateralTransactions) > 0 {
		// step 2.2.2: if potential collateral transaction (connected to origin transactions) have
		// been found they are getting classified, before appending them to the set of transactions
		// which is getting inserted into the db
		originCC, originCP, err := getConnectedCollaterals(c.db, potentialCollateralTransactions, toBlockID)
		if err != nil {
			return false, err
		}

		updatedTransactions := make([]db.Transaction, len(foundOrigins))
		for i, o := range foundOrigins {
			updatedTransactions[i] = newOriginTransaction(o.UID)
		}

		updatedTransactions = slices.Concat(updatedTransactions, originCC, originCP)

		if len(updatedTransactions) > 0 {
			if updateErr := db.UpdateTransactions(c.db, updatedTransactions); updateErr != nil {
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
			numInserted, ccErr = analytics.SetCollateralCreation(c.db, getUids(ccTransactions))
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
			numInserted, cpErr = analytics.SetCollateralPayment(c.db, getUids(cpTransactions))
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

	// set the last classified block
	if statusErr := dbstat.SetLastClassifiedBlockID(ctx, c.db, toBlockID); statusErr != nil {
		return false, statusErr
	}

	c.blocksProcessed = toBlockID - c.state.ID + 1
	c.blocks.Add(float64(c.blocksProcessed))
	c.transactions.Add(float64(len(mixingTransactions)))
	c.blockHeight.Set(float64(toBlockID))

	return true, nil
}

// PostExecution sets the classifier status activity flag to false
func (c *Classifier) PostExecution(ctx context.Context) error {
	return dbstat.SetClassifying(ctx, c.db, false)
}

// setInitialClassifierID sets the starting classifier block id to the
// value of startBlockClassifier if no value has been set yet
func setInitialClassifierID(ctx context.Context, dgraph external.Database) (err error) {
	status, err := dbstat.GetClassifierStatus(ctx, dgraph)
	if err != nil {
		return
	}

	if status.LastClassifiedBlockID == nil {
		if err = dbstat.SetLastClassifiedBlockID(ctx, dgraph, 0); err != nil {
			return
		}
	}
	return
}

// isCollateralCreation checks if the transactions is a collateral creation transaction
func isCollateralCreation(dgraph external.Database, t db.Transaction) (bool, error) {
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

	inputCount, outputCount, err := db.GetOutputAddressCounts(dgraph, t.UID)
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
	return db.Transaction{UID: uid, Type: constants.TypeCC}
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
	return db.Transaction{UID: uid, Type: constants.TypeCP}
}

// isMixing checks if the transactions is a mixing transaction
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
	return db.Transaction{UID: uid, Type: constants.TypeMixing}
}

// newOriginTransaction returns a new origin transaction with the given uid
func newOriginTransaction(uid string) db.Transaction {
	return db.Transaction{UID: uid, Type: constants.TypeOrigin}
}

// hasValidTransactionType check is the transaction has a valid privacy type
func hasValidTransactionType(tx db.Transaction) bool {
	return constants.IsValidTransactionType(tx.Type)
}

// classifyTransactions detects mixing and collateral creation transactions and sets the privacy type appropriately
// The returned slice contains all classified transactions or nil if no privacy transactions have been found.
func classifyTransactions(dgraph external.Database, transactions []db.Transaction) (mixing []db.Transaction,
	cc []db.Transaction, cp []db.Transaction, err error) {
	for _, transaction := range transactions {
		// only do classification for non-classified transactions
		if hasValidTransactionType(transaction) {
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

		isCC, collateralErr := isCollateralCreation(dgraph, transaction)
		if collateralErr != nil {
			err = collateralErr
		}

		if isCC {
			cc = append(cc, newCollateralCreationTransaction(transaction.UID))
			continue
		}
	}
	return
}
