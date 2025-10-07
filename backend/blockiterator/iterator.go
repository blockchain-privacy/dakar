// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package blockiterator

import (
	"backend/db"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// BlockIterator defines the basic structure of a process which
// iterates sequentially over a set of blocks:
//  1. do pre loop operations like getting the start id
//  2. do the sequential processing of blocks
//  3. do post loop operations in the case of a
//     failure or if the process finished due to termination
type BlockIterator interface {
	// CalculateInitialState calculates the initial state of the BlockIterator
	CalculateInitialState(context.Context) error
	// Iterate does one execution loop
	// false -> stop execution
	Iterate(context.Context) (bool, error)
	// Next tries to increase the internal state to the next block. Returns false if this fails.
	// This will be called periodically when Empty returns true. Should return true if the state
	// transition was successful.
	Next(context.Context) (bool, error)
	// PostExecution before the block iterator stops.
	// This function should do operations like the setting the database status
	PostExecution(context.Context) error
	IncrementState() error
	// Empty returns true if the BlockIterator has no more data to iterate on.
	// This happens if State.ID is higher than State.Top
	Empty() bool
	// Props returns the properties of the iterator
	Props() Properties
	// SetMaxBlocks sets the number of blocks each iteration processes.
	SetMaxBlocks(int64)
}

type Properties struct {
	// Name of the iterator
	Name    string
	Context context.Context
	// CurrentBlock refers to the block height which is currently being processed
	CurrentBlock int64
	// ProcessedBlockCount is the total number of blocks that were processed during the last Iterate call
	ProcessedBlockCount int64
	// SupportsMultiBlockIteration returns true if the iterator is capable of iterating over multiple blocks
	SupportsMultiBlockIteration bool
}

// State holds the current state of the processing loop
type State struct {
	// ID is the current block height
	ID int64

	// Top is the highest block height, which was observed at some point
	Top int64
}

// scaleBlocksPerIteration determines the number of blocks to be processed in each iteration.
// The scaling is based on whether the target duration is less than or greater than the iteration duration.
// The upper limit is capped at 200 blocks.
func scaleBlocksPerIteration(target time.Duration, iterationDuration time.Duration, blockCount int64) int64 {
	if target <= 0 {
		return 1
	}

	upperLimit := time.Duration(float64(target) * 1.1)
	if iterationDuration > upperLimit {
		return max(1, int64(float64(blockCount)*0.75))
	}

	lowerLimit := time.Duration(float64(target) * 0.9)
	if iterationDuration < lowerLimit {
		return min(max(blockCount+1, int64(float64(blockCount)*1.1)), 200)
	}

	return blockCount
}

func (s State) String() string {
	return fmt.Sprintf("ID: %d, Top: %d", s.ID, s.Top)
}

func info(iterator BlockIterator, msg string, v ...interface{}) {
	slog.Info(msg, append([]interface{}{"block_iterator_name", iterator.Props().Name}, v...)...)
}

// StartIteration starts the iteration process. Set targetIterationDuration to the duration each iteration should scale to.
func StartIteration(iterator BlockIterator, targetIterationDuration time.Duration, postIterationHook func()) (err error) {
	defer func() {
		info(iterator, "iterator stopped", "current block", iterator.Props().CurrentBlock)

		// separate context because iterator context is cancelled at this point
		ctx, cancel := db.GetShortTaskContext()
		defer cancel()
		// if the call to PostExecution results in an error, then only set the
		// error if the error is currently nil
		postErr := iterator.PostExecution(ctx)
		if err == nil && postErr != nil {
			err = postErr
		}
	}()

	props := iterator.Props()
	if initErr := iterator.CalculateInitialState(props.Context); initErr != nil {
		// only return an error if context was not cancelled
		if !errors.Is(props.Context.Err(), context.Canceled) {
			err = initErr
		}

		return
	}

	info(iterator, "iterator started", "current block", iterator.Props().CurrentBlock)

	lastMetricPrintBlockID := int64(0)
	numIteratedBlocks := int64(0)
	timerGlobal := time.Now()

	for {
		select {
		case <-props.Context.Done():
			return
		default:
			// we do nothing
		}

		// check if we need to wait
		if iterator.Empty() {
			isInterrupt, waitErr := waitForNextDBBlockID(iterator)
			if waitErr != nil {
				err = waitErr
				return
			}

			if isInterrupt {
				return
			}
		}

		ctx, cancel := db.GetLongTaskContext()
		now := time.Now()
		ok, iterateErr := iterator.Iterate(ctx)
		// can't use defer in loop and context is only needed of Iterate()
		cancel()
		if iterateErr != nil {
			err = iterateErr
			return
		}
		iterationDuration := time.Since(now)

		// stop execution
		if !ok {
			return
		}

		// set next state
		if incErr := iterator.IncrementState(); incErr != nil {
			err = incErr
			return
		}

		// metrics
		numIteratedBlocks += iterator.Props().ProcessedBlockCount
		blocksSinceLastPrint := numIteratedBlocks - lastMetricPrintBlockID
		if blocksSinceLastPrint >= 1000 {
			info(iterator, fmt.Sprintf("avg %d blocks: %v ms/block", blocksSinceLastPrint,
				time.Since(timerGlobal).Milliseconds()/blocksSinceLastPrint))
			timerGlobal = time.Now()
			lastMetricPrintBlockID = numIteratedBlocks
		}

		if postIterationHook != nil {
			postIterationHook()
		}

		if iterator.Props().SupportsMultiBlockIteration {
			iterator.SetMaxBlocks(scaleBlocksPerIteration(targetIterationDuration,
				iterationDuration, iterator.Props().ProcessedBlockCount))
		}
	}
}

// waitForNextDBBlockID waits for the next block.
// if an interrupt was received, returns true.
func waitForNextDBBlockID(it BlockIterator) (bool, error) {
	ctx := it.Props().Context
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return true, nil
		case <-ticker.C:
			// if iterator state is not empty anymore
			if !it.Empty() {
				return false, nil
			}

			if ok, err := it.Next(ctx); err != nil {
				return false, err
			} else if ok {
				return false, nil
			}
		}
	}
}
