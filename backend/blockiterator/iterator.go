package blockiterator

import (
	"context"
	"fmt"
	"github.com/qrest/gomisc/serror"
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
	// Props returns the properties of the iterator
	Props() Properties
}

type Properties struct {
	// Name of the iterator
	Name    string
	Context context.Context
	// Logger used for this iterator
	Logger *slog.Logger
	// CurrentBlock is the block height which is currently processed
	CurrentBlock uint64
	// ProcessedBlockCount is the number of blocks which have been processed by the last Iterate call
	ProcessedBlockCount uint64
}

// State holds the current state of the processing loop
type State struct {
	// ID is the current block height
	ID uint64

	// Top is the highest block height, which was observed at some point
	Top uint64
}

func (s State) String() string {
	return fmt.Sprintf("ID: %d, Top: %d", s.ID, s.Top)
}

func info(iterator BlockIterator, msg string, v ...interface{}) {
	props := iterator.Props()
	props.Logger.Info(msg, append([]interface{}{"block_iterator_name", props.Name}, v...)...)
}

// StartIteration starts the iteration process
func StartIteration(iterator BlockIterator) (err error) {
	props := iterator.Props()
	if l := props.Logger; l == nil {
		return serror.FromStr(props.Name + " logger is nil")
	}

	defer func() {
		info(iterator, "iterator stopped")
		// if the call to PostExecution results in an error, then only set the
		// error if the error is currently nil
		postErr := iterator.PostExecution()
		if err == nil && postErr != nil {
			err = postErr
		}
	}()

	if initErr := iterator.CalculateInitialState(); initErr != nil {
		err = initErr
		return
	}

	info(iterator, fmt.Sprintf("starting at: %d", iterator.Props().CurrentBlock))

	lastMetricPrintBlockID := uint64(0)
	numIteratedBlocks := uint64(0)
	timerGlobal := time.Now()
	ctx := props.Context

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
				err = waitErr
				return
			}

			if isInterrupt {
				return
			}
		}

		ok, iterateErr := iterator.Iterate()
		if iterateErr != nil {
			err = iterateErr
			return
		}

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
		blocksSinceLastPrint := int64(numIteratedBlocks - lastMetricPrintBlockID)
		if blocksSinceLastPrint >= 1000 {
			info(iterator, fmt.Sprintf("avg %d blocks: %v ms/block", blocksSinceLastPrint,
				time.Since(timerGlobal).Milliseconds()/blocksSinceLastPrint))
			timerGlobal = time.Now()
			lastMetricPrintBlockID = numIteratedBlocks
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

			if ok, err := it.NextBlock(); err != nil {
				return false, err
			} else if ok {
				return false, nil
			}
		}
	}
}
