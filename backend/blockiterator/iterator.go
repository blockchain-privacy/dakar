package blockiterator

import (
	"backend/cmd/cliutil"
	"context"
	"errors"
	"fmt"
	"log"
	"time"
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
	// NextBlock tries to increase the internal state to the next block. Returns false if this fails.
	// This will be called periodically when Empty returns true. Should return true if the state
	// transition was successful.
	NextBlock() (bool, error)
	// PostExecution is always executed, even if PreLoop or Loop fail.
	// This function should do operations like the setting the database status
	PostExecution() error
	IncrementState() error
	// Empty returns true if the BlockIterator has no more data to iterate on.
	// This happens if State.ID is higher than State.Top
	Empty() bool
	// CurrentBlock returns the height of the block which is currently processed
	CurrentBlock() uint64

	Logger() *log.Logger
	Context() context.Context
	Name() string
}

// State holds the current state of the processing loop
type State struct {
	// ID is the current block height
	ID uint64

	// Top is the highest block height, which was observed at one point
	Top uint64
}

func (s State) String() string {
	return fmt.Sprintf("ID: %d, Top: %d", s.ID, s.Top)
}

func info(iterator BlockIterator, v ...interface{}) {
	iterator.Logger().Println(append([]interface{}{iterator.Name()}, v...)...)
}

// StartIteration starts the iteration process
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

	info(iterator, "starting at:", iterator.CurrentBlock())

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
			isInterrupt, waitErr := waitForNextDBBlockID(iterator)
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
		if incErr := iterator.IncrementState(); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), incErr)
			return
		}

		// metrics
		numIteratedBlocks++
		if numIteratedBlocks%1000 == 0 {
			info(iterator, "avg 1000 blocks:", time.Since(timerGlobal).Milliseconds()/1000, "ms/block")
			timerGlobal = time.Now()
		}
	}
}

// waitForNextDBBlockID waits for the next block.
// if the interrupt receives a signal isInterrupt is true
// if the next block is available, currentBlock gets updated
func waitForNextDBBlockID(it BlockIterator) (isInterrupt bool, err error) {
	ctx := it.Context()
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			isInterrupt = true
			return
		case <-ticker.C:
			// if iterator state is not empty anymore
			if !it.Empty() {
				return
			}

			if ok, nextErr := it.NextBlock(); err != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), nextErr)
				return
			} else if ok {
				return
			}
		}
	}
}
