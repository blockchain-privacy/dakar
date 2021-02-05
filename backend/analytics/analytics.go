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
	"log"
	"time"
)

// block id after which we start analysing. found empirically.
const analyseStartBlock = 206940

var errorInterrupted = errors.New("interrupted")

func info(v ...interface{}) {
	log.SetPrefix("\033[0;32manalyse\u001B[0m\t")
	log.Println(v)
	log.SetPrefix("")
}

func metric(v ...interface{}) {
	log.SetPrefix("metric\t")
	log.Println(v)
	log.SetPrefix("")
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

func setInitialAnalyserId(dgraph *dgo.Dgraph) (err error) {
	status, err := dbstat.GetAnalyzerStatus(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if status.LastAnalysedBlockId == nil ||
		*status.LastAnalysedBlockId < analyseStartBlock {
		if err = dbstat.SetLastAnalysedBlockId(dgraph, analyseStartBlock); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}
	return
}

// StartAnalysis starts the analysis of transactions in the database.
// It entails setting the privacy type for each transaction and performing
// a reverse transactions lookup to find all origins of destination transactions.
func StartAnalysis(ctx context.Context, dgraph *dgo.Dgraph) error {
	if err := dbstat.SetAnalyzing(dgraph, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	defer func() {
		if err := dbstat.SetAnalyzing(dgraph, false); err != nil {
			info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
			return
		}
	}()

	if err := setInitialAnalyserId(dgraph); err != nil {
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

		if len(updatedBlock.Transactions) > 0 {
			// update the block in the database
			// after that function call the privacy type of all transactions is set
			if err := dbblk.UpdateBlock(dgraph, updatedBlock); err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}

			if err := blockReverseLookup(dgraph, updatedBlock.Uid); err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		if err := dbstat.SetLastAnalysedBlockId(dgraph, state.id); err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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

// reverseLookup performs for all destinationInputTransactions a reverse lookup
func reverseLookup(dgraph *dgo.Dgraph, destinationInputTransactions []string) error {
	const mutationBatchSize = 1000
	for _, t := range destinationInputTransactions {
		info("reverse lookup for", t)
		timeNow := time.Now()
		origins, err := dban.AnalyzeOrigins(dgraph, t)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		info("origin count:", len(origins), "time:", time.Since(timeNow))

		isDone := false
		for i := 0; i < len(origins); i += mutationBatchSize {
			batch := origins[i:min(i+mutationBatchSize, len(origins))]
			info("setting origins:", len(batch))

			// set flag to mark transaction as fully analysed
			if i+mutationBatchSize >= len(origins) {
				isDone = true
			}

			if err := dban.SetOrigins(dgraph, t, batch, isDone); err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}
	}

	return nil
}

// blockReverseLookup performs a reverse lookup for all input transactions of destination transactions included in the block
func blockReverseLookup(dgraph *dgo.Dgraph, blockUid string) error {
	inputTransactions, err := dban.GetNotAnalyzedInputTransactionsPerBlock(dgraph, blockUid)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if err := reverseLookup(dgraph, inputTransactions); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// transactionReverseLookup performs a reverse lookup for all input transactions of  the given destination transaction
// todo: currently unused - do not remove, this function is needed for the ad-hoc reverse lookup initiated by heuristic executors
func transactionReverseLookup(dgraph *dgo.Dgraph, destinationTransactionUid string) error {
	inputTransactions, err := dban.GetNotAnalyzedInputTransactionsPerTx(dgraph, destinationTransactionUid)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if err := reverseLookup(dgraph, inputTransactions); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
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
		transaction, txErr := dbtx.GetTransaction(dgraph, tx.Hash)
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
