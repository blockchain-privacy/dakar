package analytics

import (
	"context"
	"dashrpc/cmd/cliutil"
	dbaddr "dashrpc/db/address"
	dban "dashrpc/db/analytics"
	dbblk "dashrpc/db/block"
	dbstat "dashrpc/db/status"
	dbtx "dashrpc/db/transaction"
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

func StartPost(ctx context.Context, dgraph *dgo.Dgraph) error {
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

		updatedBlock, err := analyseBlock(dgraph, currentBlock, ctx.Done())
		if err != nil {
			if errors.Is(err, errorInterrupted) {
				analyzingInterrupted()
				return nil
			}

			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(updatedBlock.Transactions) > 0 {
			if err := dbblk.UpdateBlock(dgraph, updatedBlock); err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		if err := dbstat.SetLastAnalysedBlockId(dgraph, state.id); err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		state.id++
		counterAnalysedBlocks++

		if counterAnalysedBlocks%1000 == 0 {
			metric("moving avg 1000 blocks:", time.Since(timerGlobal).Milliseconds()/int64(counterAnalysedBlocks), "ms/block")
			timerGlobal = time.Now()
		}
	}

	return nil
}

func analyzingInterrupted() {
	info("Stopped")
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

func analyseBlock(dgraph *dgo.Dgraph, block dbblk.Block, interrupt <-chan struct{}) (updatedBlock dbblk.Block, err error) {
	updatedBlock.Uid = block.Uid

	for _, tx := range block.Transactions {
		select {
		case <-interrupt:
			err = errorInterrupted
			return
		default:
			// we do nothing
		}

		transaction, txErr := dbtx.GetTransaction(dgraph, tx.Hash)
		if txErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
		transaction, err = setPrivacyType(dgraph, transaction)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		if transaction.PrivacyType != "" {
			updatedTransaction := dbtx.Transaction{
				Uid:         transaction.Uid,
				PrivacyType: transaction.PrivacyType,
			}

			// find all potential origins for transaction
			if transaction.PrivacyType == dbtx.PrivacyDestination {
				info("Starting analyzing", transaction.Hash)
				start := time.Now()

				origins, originErr := dban.AnalyzeOrigins(dgraph, transaction.Hash)
				if originErr != nil {
					err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), originErr)
					return
				}

				for _, o := range origins {
					updatedTransaction.Origins = append(updatedTransaction.Origins, dbtx.Transaction{Uid: o})
				}
				t := time.Now()
				info("Finished analyzing", transaction.Hash, "Elapsed time:", t.Sub(start), "Origins:", len(origins))
			}

			updatedBlock.Transactions = append(updatedBlock.Transactions, updatedTransaction)
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
		// it is possible for addressses not to be found
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), addErr)
		return
	}

	if newTx.IsPrivacyOrigin(areALLAddressesEqual(addresses)) {
		newTx.SetPrivacyOrigin()
	}

	return
}

// returns true if all addresses are different
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

func GetAllPaths(dgraph *dgo.Dgraph, transactionsHash string) (err error) {
	transaction, err := dbtx.GetTransaction(dgraph, transactionsHash)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	origins, err := dban.GetOrigins(dgraph, transactionsHash)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for i, o := range origins {
		weight, weightErr := dban.GetShortestPathWeight(dgraph, transaction.Uid, o.Uid)
		if weightErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), weightErr)
			return
		}

		fmt.Println(weight)

		elements, pathError := dban.GetPaths(dgraph, transaction.Uid, o.Uid, 1, uint32(weight))
		if pathError != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), pathError)
			return
		}

		if len(elements) < 3 {
			fmt.Println("Found element smaller 3")
		}

		fmt.Println(i)
	}

	return
}
