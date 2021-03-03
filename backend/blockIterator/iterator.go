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
	GetInitialState() (State, error)
	Iterate(State) (bool, error)
	// PostLoop is always executed, even if PreLoop or Loop fail.
	// This function should do operations like the setting the database status
	PostExecution() error
	Logger() *log.Logger
	Context() context.Context
	Db() *dgo.Dgraph
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

func StartIteration(iterator BlockIterator) (err error) {
	l := iterator.Logger()

	if l == nil {
		return errors.New("logger is nil")
	}

	defer func() {
		l.Println("iterator stopped")
		// if the call to PostExecution results in an error, then only set the
		// error if the error is currently nil
		postErr := iterator.PostExecution()
		if err == nil && postErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), postErr)
		}
	}()

	l.Println("doing pre loop")
	initialState, initErr := iterator.GetInitialState()
	if initErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), initErr)
		return
	}

	numIteratedBlocks := 0
	timerGlobal := time.Now()
	ctx := iterator.Context()

	l.Println("doing loop")
	nextState := initialState

	for nextState.Top != 0 && nextState.Id != 0 {
		select {
		case <-ctx.Done():
			return
		default:
			// we do nothing
		}

		// copy state
		state := nextState

		if state.Id > state.Top {
			l.Println("Waiting for next block", state)

			updatedState, isInterrupt, waitErr := waitForNextDbBlockId(iterator.Context(),
				iterator.Db(), state)
			if waitErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), waitErr)
				return
			}
			state = updatedState

			if isInterrupt {
				return
			}

			l.Println("Found next block. New state:", state)
		}

		ok, iterateErr := iterator.Iterate(state)
		if iterateErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), iterateErr)
			return
		}

		// stop execution
		if !ok {
			return
		}

		numIteratedBlocks++

		if numIteratedBlocks%100 == 0 {
			l.Println("avg 100 blocks:", time.Since(timerGlobal).Milliseconds()/100, "ms/block")
			timerGlobal = time.Now()
		}

		nextState.Top = state.Top
		nextState.Id = state.Id + 1
	}

	return
}

// waitForNextDbBlockId waits for the next block.
// if the interrupt receives a signal isInterrupt is true
// if the next block is available, currentBlock gets updated
func waitForNextDbBlockId(ctx context.Context, dgraph *dgo.Dgraph,
	currentState State) (nextState State, isInterrupt bool, err error) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			isInterrupt = true
			return
		case <-ticker.C:
			status, statusError := dbstat.GetCrawlerStatus(dgraph)
			if statusError != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), statusError)
				return
			}

			if status.LastBlockId != nil && *status.LastBlockId >= currentState.Id {
				if *status.LowestBlockId > currentState.Id {
					currentState.Id = *status.LowestBlockId
				}

				nextState = currentState
				nextState.Top = *status.LastBlockId
				return
			}
		}
	}
}
