package analytics

import (
	"backend/blockIterator"
	"backend/cmd/cliutil"
	dban "backend/db/analytics"
	dbstat "backend/db/status"
	dbtx "backend/db/transaction"
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"io"
	"log"
	"sync"
	"time"
)

const (
	// analyticsLoggerPrefix is the prefix which is printed for each log message of analyticsLogger
	analyticsLoggerPrefix = "\033[0;32manalyse\u001B[0m\t"

	// mutationBatchSize is the maximum size of origin batches created by the reverseLookup
	mutationBatchSize = 1000
)

var errorInterrupted = errors.New("interrupted")

var analyticsLogger = log.New(log.Writer(), analyticsLoggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	analyticsLogger = log.New(out, analyticsLoggerPrefix, flag)
}

func info(v ...interface{}) {
	analyticsLogger.Println(v)
}

// setInitialAnalyserId sets the starting analyser block id to the
// value of startBlockAnalyser if no value has been set yet
func setInitialAnalyserId(dgraph *dgo.Dgraph, startBlockAnalyser uint64) (err error) {
	status, err := dbstat.GetAnalyzerStatus(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if status.LastAnalysedBlockId == nil ||
		*status.LastAnalysedBlockId < startBlockAnalyser {
		if err = dbstat.SetLastAnalysedBlockId(dgraph, startBlockAnalyser); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}
	return
}

type Analyzer struct {
	config   Config
	db       *dgo.Dgraph
	ctx      context.Context
	state    blockIterator.State
	queue    chan string
	mapLock  sync.Mutex
	queueMap map[string]bool
}

// NewAnalyzer creates a new Analyzer object
func NewAnalyzer(ctx context.Context, dgraph *dgo.Dgraph, cfg Config) *Analyzer {
	return &Analyzer{
		config:   cfg,
		db:       dgraph,
		ctx:      ctx,
		queue:    make(chan string, 5),
		queueMap: make(map[string]bool),
	}
}

func (a *Analyzer) Name() string {
	return "analyzer"
}

func (a *Analyzer) Logger() *log.Logger {
	return analyticsLogger
}

func (a *Analyzer) State() blockIterator.State {
	return a.state
}

func (a *Analyzer) SetState(newState blockIterator.State) {
	a.state = newState
}

func (a *Analyzer) Context() context.Context {
	return a.ctx
}

func (a *Analyzer) Db() *dgo.Dgraph {
	return a.db
}

func (a *Analyzer) IncrementState() {
	a.state.Id++
}

// Empty checks if there is work in the queue or more block above the current one
func (a *Analyzer) Empty() bool {
	// len(channel) == 0 is bad practice to determine if the channel is empty,
	// when the channel is being read by multiple goroutines. In this case
	// the channel is only being read by one goroutine so this is fine.
	return len(a.queue) == 0 && a.state.Id > a.state.Top
}

func (a *Analyzer) CalculateInitialState() error {
	if !a.config.IsAnalysingEnabled {
		return errors.New("analysing is disabled per configuration")
	}

	if err := dbstat.SetAnalyzing(a.db, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if err := setInitialAnalyserId(a.db, a.config.AnalyseStartBlock); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	crawlerStatus, err := dbstat.GetCrawlerStatus(a.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	analyzerStatus, err := dbstat.GetAnalyzerStatus(a.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if analyzerStatus.LastAnalysedBlockId == nil {
		return errors.New("error last analysed block is not set")
	}

	var state blockIterator.State

	state.Id = *analyzerStatus.LastAnalysedBlockId + 1

	if crawlerStatus.LastBlockId == nil {
		// nothing crawled yet, so set Top to a lower number as Id
		state.Top = *analyzerStatus.LastAnalysedBlockId
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

func (a *Analyzer) AddToQueue(txHash string) bool {
	// todo return a status instead to differentiate
	//  between queue full and tx already in queue
	// todo do some verification that tx is actually a destination transaction
	a.mapLock.Lock()
	defer a.mapLock.Unlock()
	if _, ok := a.queueMap[txHash]; ok {
		return false
	}

	select {
	case a.queue <- txHash:
		a.queueMap[txHash] = true
		return true
	default:
		return false
	}
}

func (a *Analyzer) Iterate() (bool, error) {
	if a.Empty() {
		return false, errors.New("got empty state")
	}

	// check queue state
	select {
	case destinationTransaction, ok := <-a.queue:
		if ok {
			info("processing", destinationTransaction)
			lookup, err := transactionReverseLookup(a.ctx, a.db, destinationTransaction)
			if err != nil {
				return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
			info("Transaction", destinationTransaction, "origin count", lookup)
		} else {
			info("queue closed!")
		}
	default:
		info("no value in queue -> doing block iteration")
	}

	//currentBlock, err := dbblk.GetBlockById(a.db, a.state.Id)
	//if err != nil {
	//	return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	//}

	//updatedBlock, err := getMixingTransactions(a.ctx, a.db, currentBlock)
	//if err != nil {
	//	if errors.Is(err, errorInterrupted) {
	//		return false, nil
	//	}
	//
	//	return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	//}
	//
	//wasInterrupted := false
	//if len(updatedBlock.Transactions) > 0 {
	//	// update the block in the database
	//	// after that function call the privacy type of all transactions is set
	//	if err := dbblk.UpdateBlock(a.db, updatedBlock); err != nil {
	//		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	//	}
	//
	//	originCount, err := blockReverseLookup(a.ctx, a.db, updatedBlock.Uid)
	//	if err != nil {
	//		if errors.Is(err, errorInterrupted) {
	//			wasInterrupted = true
	//		} else {
	//			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	//		}
	//	}
	//	if originCount > 0 {
	//		info("Block", *currentBlock.Id, "origin count", originCount)
	//	}
	//}

	// only set last analysed flag if processes before were not interrupted
	//if wasInterrupted {
	//	return false, nil
	//} else {
	//	if err := dbstat.SetLastAnalysedBlockId(a.db, a.state.Id); err != nil {
	//		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	//	}
	//}

	return true, nil
}

func (a *Analyzer) PostExecution() error {
	if err := dbstat.SetAnalyzing(a.db, false); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// reverseLookup performs for all destinationInputTransactions a reverse lookup.
// The returned integer is the number of origins inserted. It is returned regardless of an error.
// reverseLookup process:
// 1. Starting from a transaction traverse all connected mixing transactions
// 2. Find all origin transaction which are directly connected to each mixing transaction and the
//    initial transaction
// 3. If the resulting number of origins is less than dban.SameRequestMutationLimit set the origins to the
//    initial transaction in the same query
// 4. If the resulting number of origins is bigger or equal to dban.SameRequestMutationLimit  set the origins
//    in batches of mutationBatchSize.
func reverseLookup(ctx context.Context, dgraph *dgo.Dgraph, destinationInputTransactions []string) (int64, error) {
	var insertedOrigins int64

	for _, t := range destinationInputTransactions {
		select {
		case <-ctx.Done():
			info("Stopping reverseLookup ...")
			return insertedOrigins, errorInterrupted
		default:
			// we do nothing
		}

		// get origins
		timeNow := time.Now()
		origins, err := dban.AnalyzeOrigins(dgraph, t)
		if err != nil {
			return insertedOrigins, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		queryTime := time.Since(timeNow)
		var mutationTime time.Duration

		// only set origins if not already done by previous step
		if len(origins) >= dban.SameRequestMutationLimit {
			// set origins
			timeNow = time.Now()
			isDone := false
			for i := 0; i < len(origins); i += mutationBatchSize {
				batch := origins[i:min(i+mutationBatchSize, len(origins))]

				// set flag to mark transaction as fully analysed
				if i+mutationBatchSize >= len(origins) {
					isDone = true
				}

				if err := dban.SetOrigins(dgraph, t, batch, isDone); err != nil {
					return insertedOrigins, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
				}
			}

			// metrics
			mutationTime = time.Since(timeNow)
		}

		info("analyzing", t, "origin count:", len(origins), "query time:", queryTime,
			"mutation time:", mutationTime, "full time:", queryTime+mutationTime)

		insertedOrigins += int64(len(origins))
	}

	return insertedOrigins, nil
}

// blockReverseLookup performs a reverse lookup for all input transactions of destination transactions included in the block
// The returned integer is the number of origins inserted. It is returned regardless of an error.
func blockReverseLookup(ctx context.Context, dgraph *dgo.Dgraph, blockUid string) (int64, error) {
	inputTransactions, err := dban.GetNotAnalyzedInputTransactionsPerBlock(dgraph, blockUid)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	num, err := reverseLookup(ctx, dgraph, inputTransactions)
	if err != nil {
		return num, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)

	}

	return num, nil
}

// transactionReverseLookup performs a reverse lookup for all input transactions of  the given destination transaction
// The returned integer is the number of origins inserted. It is returned regardless of an error.
// todo: currently unused - do not remove, this function is needed for the ad-hoc reverse lookup initiated by heuristic executors
func transactionReverseLookup(ctx context.Context, dgraph *dgo.Dgraph, destinationTransactionUid string) (int64, error) {
	inputTransactions, err := dban.GetNotAnalyzedInputTransactionsPerTx(dgraph, destinationTransactionUid)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	num, err := reverseLookup(ctx, dgraph, inputTransactions)
	if err != nil {
		return num, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return num, nil
}

// min returns the smaller of the two values
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// getPrivacyTransactions detects mixing and collateral creation transactions and sets the privacy type appropriately
// The returned slice contains all classified transactions or nil if no privacy transactions have been found.
func getPrivacyTransactions(dgraph *dgo.Dgraph, transactions []dbtx.Transaction) (mixing []dbtx.Transaction,
	cc []dbtx.Transaction, cp []dbtx.Transaction, err error) {
	for _, transaction := range transactions {
		if isMixing(transaction) {
			mixing = append(mixing, newMixingTransaction(transaction.Uid))
			continue
		}

		if isCollateralPayment(transaction) {
			cp = append(cp, newCollateralPaymentTransaction(transaction.Uid))
			continue
		}

		// todo this fails sometime if the origin transaction has not been classified yet.
		// this is the case if the collateral creation transaction occurs before the
		// first mixing transaction connected to the origin tx. Solution:
		// either do not do the lookup and "trust" the classification based on available data
		// or classify origins beforehand only on available data (like before)
		isCC, collateralErr := isCollateralCreation(dgraph, transaction)
		if collateralErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), collateralErr)
		}

		if isCC {
			cc = append(cc, newCollateralCreationTransaction(transaction.Uid))
			continue
		}
	}
	return
}
