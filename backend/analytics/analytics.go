package analytics

import (
	"backend/cmd/cliutil"
	dbaddr "backend/db/address"
	dban "backend/db/analytics"
	dbblk "backend/db/block"
	dbstat "backend/db/status"
	dbtx "backend/db/transaction"
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"io"
	"log"
	"time"
)

const (
	// analyticsLoggerPrefix is the prefix which is printed for each log message of analyticsLogger
	analyticsLoggerPrefix = "\033[0;32manalyse\u001B[0m\t"
	// metricsLoggerPrefix is the prefix which is printed for each log message of metricLogger
	metricsLoggerPrefix = "metric\t"

	// mutationBatchSize is the maximum size of origin batches created by the reverseLookup
	mutationBatchSize = 1000
)

var errorInterrupted = errors.New("interrupted")

var analyticsLogger = log.New(log.Writer(), analyticsLoggerPrefix, log.Flags())
var metricLogger = log.New(log.Writer(), metricsLoggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	analyticsLogger = log.New(out, analyticsLoggerPrefix, flag)
	metricLogger = log.New(out, metricsLoggerPrefix, flag)
}

func info(v ...interface{}) {
	analyticsLogger.Println(v)
}

func metric(v ...interface{}) {
	metricLogger.Println(v)
}

// holds the current state of the analyzing processing loop
type analyzerProcessingState struct {
	// current block id
	id uint64

	// highest block id
	top uint64
}

func (a analyzerProcessingState) String() string {
	return fmt.Sprintf("Id: %d, Top: %d", a.id, a.top)
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

// StartAnalysis starts the analysis of transactions in the database.
// It entails setting the privacy type for each transaction and performing
// a reverse transactions lookup to find all origins of destination transactions.
func StartAnalysis(ctx context.Context, dgraph *dgo.Dgraph, config Config) error {
	if !config.IsAnalysingEnabled {
		return errors.New("analysing is disabled per configuration")
	}

	if err := dbstat.SetAnalyzing(dgraph, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	defer func() {
		if err := dbstat.SetAnalyzing(dgraph, false); err != nil {
			info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
			return
		}
	}()

	if err := setInitialAnalyserId(dgraph, config.AnalyseStartBlock); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	crawlerStatus, err := dbstat.GetCrawlerStatus(dgraph)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	analyzerStatus, err := dbstat.GetAnalyzerStatus(dgraph)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	var state analyzerProcessingState

	if analyzerStatus.LastAnalysedBlockId == nil {
		return errors.New("error last analysed block is not set")
	}

	state.id = *analyzerStatus.LastAnalysedBlockId + 1

	if crawlerStatus.LastBlockId == nil {
		state.top = *analyzerStatus.LastAnalysedBlockId
	} else if *crawlerStatus.LowestBlockId > state.id {
		state.id = *crawlerStatus.LowestBlockId
		state.top = *crawlerStatus.LastBlockId
	} else {
		state.top = *crawlerStatus.LastBlockId
	}

	info("Starting process")

	counterAnalysedBlocks := 0
	timerGlobal := time.Now()
mainLoop:
	for {
		select {
		case <-ctx.Done():
			analyzingInterrupted()
			break mainLoop
		default:
			// we do nothing
		}

		// update top state
		if state.id > state.top {
			info("Waiting for next block", state)
			var isInterrupt bool
			// can not used short hand declaration, because it would mask currentBlock in the outer scope
			state, isInterrupt, err = waitForNextDbBlockId(dgraph, ctx.Done(), state)
			if err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}

			if isInterrupt {
				break mainLoop
			}

			info("Found next block. New state:", state)
		}

		currentBlock, err := dbblk.GetBlockById(dgraph, state.id)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		updatedBlock, err := processPrivacyType(dgraph, currentBlock, ctx.Done())
		if err != nil {
			if errors.Is(err, errorInterrupted) {
				analyzingInterrupted()
				return nil
			}

			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		wasInterrupted := false
		if len(updatedBlock.Transactions) > 0 {
			// update the block in the database
			// after that function call the privacy type of all transactions is set
			if err := dbblk.UpdateBlock(dgraph, updatedBlock); err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}

			originCount, err := blockReverseLookup(ctx, dgraph, updatedBlock.Uid)
			if err != nil {
				if errors.Is(err, errorInterrupted) {
					wasInterrupted = true
				} else {
					return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
				}
			}
			if originCount > 0 {
				info("Block", *currentBlock.Id, "origin count", originCount)
			}
		}

		// only set last analysed flag if processes before were not interrupted
		if !wasInterrupted {
			if err := dbstat.SetLastAnalysedBlockId(dgraph, state.id); err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		state.id++
		counterAnalysedBlocks++

		if counterAnalysedBlocks%100 == 0 {
			metric("avg 100 blocks:", time.Since(timerGlobal).Milliseconds()/100, "ms/block")
			timerGlobal = time.Now()
		}
	}

	return nil
}

// analyzingInterrupted should be called when the analyser has been interrupted.
func analyzingInterrupted() {
	info("Stopped")
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

		// only set origins if not already done by previous function
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

// returns the smaller of the two values
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// wait for the next block
// if the interrupt receives a signal isInterrupt is true
// if the next block is available, currentBlock gets updated
func waitForNextDbBlockId(dgraph *dgo.Dgraph, interrupt <-chan struct{},
	currentState analyzerProcessingState) (nextState analyzerProcessingState, isInterrupt bool, err error) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-interrupt:
			analyzingInterrupted()
			isInterrupt = true
			return
		case <-ticker.C:
			status, statusError := dbstat.GetCrawlerStatus(dgraph)
			if statusError != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), statusError)
				return
			}

			if status.LastBlockId != nil && *status.LastBlockId >= currentState.id {
				if *status.LowestBlockId > currentState.id {
					currentState.id = *status.LowestBlockId
				}

				nextState = currentState
				nextState.top = *status.LastBlockId
				return
			}
		}
	}
}

// Sets the privacy type for all transaction in block. The resulting updateBlock only has PrivateSend transactions
func processPrivacyType(dgraph *dgo.Dgraph, block dbblk.Block, interrupt <-chan struct{}) (
	updatedBlock dbblk.Block, err error) {
	updatedBlock.Uid = block.Uid

	for _, tx := range block.Transactions {
		select {
		case <-interrupt:
			err = errorInterrupted
			return
		default:
			// we do nothing
		}

		// get transaction data
		transaction, txErr := dbtx.GetTransaction(dgraph, tx.Hash, block.Hash)
		if txErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		// set the appropriate privacy type
		transaction, err = setPrivacyType(dgraph, transaction)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		// append the transaction to the change set in case it is a PrivateSend transaction
		if transaction.PrivacyType != "" {
			updatedBlock.Transactions = append(updatedBlock.Transactions, dbtx.Transaction{
				Uid:         transaction.Uid,
				PrivacyType: transaction.PrivacyType,
			})
		}
	}
	return
}

// sets the privacy type of the transaction
func setPrivacyType(dgraph *dgo.Dgraph, tx dbtx.Transaction) (newTx dbtx.Transaction, err error) {
	newTx = tx
	if newTx.IsMixing() {
		newTx.SetMixing()
		return
	}

	if newTx.IsPrivacyDestination() {
		newTx.SetPrivacyDestination()
		return
	}

	addresses, addErr := dbaddr.GetInputAddressesOfTransaction(dgraph, tx.Uid)
	if addErr != nil && !errors.Is(addErr, dbaddr.ErrorAddressNotFound) {
		// If the crawler is executed in range mode,
		// it is possible for addresses not to be found
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), addErr)
		return
	}

	if newTx.IsPrivacyOrigin(areALLAddressesEqual(addresses)) {
		newTx.SetPrivacyOrigin()
	}

	return
}

// returns true if all addresses are equal
func areALLAddressesEqual(addresses []dbaddr.Address) bool {
	if len(addresses) < 2 {
		return true
	}

	hashes := make(map[string]bool)

	for _, a := range addresses {
		hashes[a.Hash] = true
		if len(hashes) > 1 {
			return false
		}
	}

	return true
}
