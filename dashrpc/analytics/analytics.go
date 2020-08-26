package analytics

import (
	"context"
	"dashrpc/cmd/cliutil"
	dbblk "dashrpc/db/block"
	dbstat "dashrpc/db/status"
	dbtx "dashrpc/db/transaction"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"time"
)

func info(v ...interface{}) {
	log.SetPrefix("\033[0;32manalyse\u001B[0m\t")
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

	crawlerStatus, err := dbstat.GetCrawlerStatus(dgraph)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	analyzerStatus, err := dbstat.GetAnalyzerStatus(dgraph)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	var state analyzerProcessingState

	if crawlerStatus.LastBlockId == nil {
		// no crawling has happened yet
		info("No blocks in database. Waiting until crawling starts.")

		// wait until crawler is active
		var isInterrupt bool
		state, isInterrupt, err = waitForNextDbBlockId(dgraph, ctx.Done(), state)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if isInterrupt {
			return nil
		}
	} else if analyzerStatus.LastAnalysedBlockId == nil && crawlerStatus.LowestBlockId != nil {
		state.id = *crawlerStatus.LowestBlockId
		state.top = *crawlerStatus.LastBlockId
	} else {
		if *analyzerStatus.LastAnalysedBlockId > *crawlerStatus.LastBlockId {
			return errors.New("inconsistent state of status")
		}

		state.id = *analyzerStatus.LastAnalysedBlockId + 1
		state.top = *crawlerStatus.LastBlockId
	}

	info("Starting process")

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

			info("Found next block", state)
		}

		currentBlock, err := dbblk.GetBlockById(dgraph, state.id)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// todo move to new function
		// todo use dbaddr.GetInputAddressesOfTransaction() for determining the privacy type correctly
		// todo make function modular enough to do mulitiple analytic steps
		// todo also make analytics pluggable
		updatedBlock := dbblk.Block{Uid: currentBlock.Uid}
		for _, tx := range currentBlock.Transactions {
			transaction, err := dbtx.GetTransaction(dgraph, tx.Hash)
			if err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
			transaction.SetPrivacyType()

			if transaction.PrivacyType != "" {
				updatedTransaction := dbtx.Transaction{
					Uid:         transaction.Uid,
					PrivacyType: transaction.PrivacyType,
				}
				updatedBlock.Transactions = append(updatedBlock.Transactions, updatedTransaction)
			}
		}

		if len(updatedBlock.Transactions) > 0 {
			if err := dbblk.UpdateBlock(dgraph, updatedBlock); err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
			info("updated block!")
		}

		if err := dbstat.SetLastAnalysedBlockId(dgraph, state.id); err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		state.id++
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

			if currentState.id == 0 && status.LowestBlockId != nil {
				nextState.id = *status.LowestBlockId
				nextState.top = *status.LastBlockId
				return
			} else if status.LastBlockId != nil {
				if currentState.id < *status.LastBlockId {
					nextState.id = currentState.id + 1
					nextState.top = *status.LastBlockId
					return
				}
			}
		}
	}
}
