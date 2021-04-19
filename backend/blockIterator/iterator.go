package blockIterator

import (
	"backend/cmd/cliutil"
	dbstat "backend/db/status"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/dgo/v2"
)

// BlockIterator defines the basic structure of a process which
// iterates sequentially over a set of blocks:
// 1. do pre loop operations like getting the start id
// 2. do the sequential processing of blocks
// 3. do post loop operations in the case of a
//    failure or if the process finished due to termination
type BlockIterator interface {
	// CalculateInitialState calculates the initial state of the BlockIterator
	CalculateInitialState() error
	// Iterate does one execution loop
	// false -> stop execution
	Iterate() (bool, error)
	// GetHighestAvailableBlock returns the highest block id/height which the
	// iterator can process at the time of the call
	GetHighestAvailableBlock() (uint64, error)
	// PostExecution is always executed, even if PreLoop or Loop fail.
	// This function should do operations like the setting the database status
	PostExecution() error
	IncrementState()
	SetState(State)

	// Empty returns true if the BlockIterator has no more data to iterate on.
	// This happens if State.ID is higher than State.Top
	Empty() bool

	Logger() *log.Logger
	Context() context.Context
	Db() *dgo.Dgraph
	State() State
	Name() string
}

// State holds the current state of the analyzing processing loop
type State struct {
	// Id is the current block height
	Id uint64

	// Top is the highest block height, which was observed at one point
	Top uint64
}

func (s State) String() string {
	return fmt.Sprintf("Id: %d, Top: %d", s.Id, s.Top)
}

func info(iterator BlockIterator, v ...interface{}) {
	iterator.Logger().Println(append([]interface{}{iterator.Name()}, v...))
}

func StartIteration(iterator BlockIterator) (err error) {
	if l := iterator.Logger(); l == nil {
		return errors.New(iterator.Name() + " logger is nil")
	}

	defer func() {
		info(iterator, "iterator stopped")
		// if the call to PostExecution results in an error, then only set the
		// error if the error is currently nil
		postErr := iterator.PostExecution()
		if err == nil && postErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), postErr)
		}
	}()

	if initErr := iterator.CalculateInitialState(); initErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), initErr)
		return
	}

	info(iterator, "starting at:", iterator.State().Id)

	numIteratedBlocks := 0
	timerGlobal := time.Now()
	ctx := iterator.Context()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// we do nothing
		}

		// check if we need to wait
		if iterator.Empty() {
			isInterrupt, waitErr := waitForNextDbBlockId(iterator)
			if waitErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), waitErr)
				return
			}

			if isInterrupt {
				return
			}
		}

		ok, iterateErr := iterator.Iterate()
		if iterateErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), iterateErr)
			return
		}

		// stop execution
		if !ok {
			return
		}

		// set next state
		iterator.IncrementState()

		// metrics
		numIteratedBlocks++
		if numIteratedBlocks%1000 == 0 {
			info(iterator, "avg 1000 blocks:", time.Since(timerGlobal).Milliseconds()/1000, "ms/block")
			timerGlobal = time.Now()
		}
	}
}

// waitForNextDbBlockId waits for the next block.
// if the interrupt receives a signal isInterrupt is true
// if the next block is available, currentBlock gets updated
func waitForNextDbBlockId(it BlockIterator) (isInterrupt bool, err error) {
	ctx := it.Context()
	dgraph := it.Db()

	currentState := it.State()
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			isInterrupt = true
			return
		case <-ticker.C:
			// if iterator state is not empty anymore crawler status
			if !it.Empty() {
				return
			}

			status, statusError := dbstat.GetCrawlerStatus(dgraph)
			if statusError != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), statusError)
				return
			}

			highestBlock, highErr := it.GetHighestAvailableBlock()
			if highErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), highErr)
				return
			}

			if highestBlock >= currentState.Id {
				if *status.LowestBlockId > currentState.Id {
					currentState.Id = *status.LowestBlockId
				}

				it.SetState(State{currentState.Id, highestBlock})

				return
			}
		}
	}
}
