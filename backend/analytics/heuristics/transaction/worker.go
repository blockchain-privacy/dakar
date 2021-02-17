package transaction

import (
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"

	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/dgo/v2"
)

// copyOnModify is true when existing heuristic trees should be copied before modification
const copyOnModify = false

type HeuristicQueueStatus int

const (
	// StatusHeuristicAdded is set if the heuristic has been successfully added
	StatusHeuristicAdded = HeuristicQueueStatus(iota)
	// StatusHeuristicDuplicate is set if the heuristic is already in the work queue
	StatusHeuristicDuplicate
	// StatusHeuristicNotInQueue is set if the heuristic is not in the work queue
	StatusHeuristicNotInQueue
	// StatusHeuristicInQueue is set if the heuristic is in the work queue
	StatusHeuristicInQueue
	// StatusHeuristicProcessing is set if the heuristic is currently being processed
	StatusHeuristicProcessing

	// loggerPrefix is the prefix which is printed for each log message
	loggerPrefix = "\033[0;34mhworker\u001B[0m\t"
)

var thisLogger = log.New(log.Writer(), loggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	thisLogger = log.New(out, loggerPrefix, flag)
}

func info(v ...interface{}) {
	thisLogger.Println(v)
}

type workKey struct {
	txhash  string
	userUid string
}

type Work struct {
	// executors contains the HeuristicExecutor trees
	executors []HeuristicExecutor
	// removableHeuristics contains the uids of all heuristics are ready for deletion
	removableHeuristics []string
	// treeRoots contains the uids of the roots of all trees modified by the heuristics in executors
	// needed for copying modified heuristic trees
	treeRoots []string
}

type Worker struct {
	currentWorkItem workKey
	executionMap    map[workKey]Work
	mutex           *sync.Mutex
}

func NewWorker() Worker {
	var mLock sync.Mutex
	return Worker{executionMap: make(map[workKey]Work), mutex: &mLock}
}

func (w *Worker) AddWork(transactionHash string, userUid string, work Work) bool {
	key := workKey{
		txhash:  transactionHash,
		userUid: userUid,
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if _, exists := w.executionMap[key]; exists {
		return false
	}

	w.executionMap[key] = work

	return true
}

// IsInQueue returns true if the given transaction hash and user id is in the work queue
func (w *Worker) IsInQueue(tx string, userUid string) bool {
	key := workKey{
		txhash:  tx,
		userUid: userUid,
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()
	_, ok := w.executionMap[key]
	return ok
}

// GetStatus returns the current execution status of the given transaction hash and user id
func (w *Worker) GetStatus(tx string, userUid string) HeuristicQueueStatus {
	key := workKey{
		txhash:  tx,
		userUid: userUid,
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()
	_, ok := w.executionMap[key]

	if !ok {
		return StatusHeuristicNotInQueue
	}

	if w.currentWorkItem == key {
		return StatusHeuristicProcessing
	}

	return StatusHeuristicInQueue
}

func (w *Worker) StartWorking(ctx context.Context, dgraph *dgo.Dgraph) {
	go w.work(ctx, dgraph)
}

func stoppingWorker() {
	info("stopping ...")
}

func (w *Worker) work(ctx context.Context, dgraph *dgo.Dgraph) {

	var work Work
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

mainLoop:
	for {
		select {
		case <-ctx.Done():
			stoppingWorker()
			break mainLoop
		case <-ticker.C:
			// get work for this cycle
			w.mutex.Lock()
			for k, v := range w.executionMap {
				work = v
				w.currentWorkItem = k
				break
			}
			w.mutex.Unlock()

			// do we have something to do?
			if len(work.executors) > 0 || len(work.removableHeuristics) > 0 {
				info("processing work package")
				// copy tree
				wasCopyingErrorFree := true
				if copyOnModify {
					for _, root := range work.treeRoots {
						if err := dbtxh.CopyHeuristicTree(dgraph, root); err != nil {
							info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
							wasCopyingErrorFree = false
							break
						}
					}
				}

				if wasCopyingErrorFree {
					// delete changed or removable heuristics
					if err := dbtxh.DeleteHeuristics(dgraph, work.removableHeuristics, w.currentWorkItem.userUid); err != nil {
						info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
						// no return/break because we want keep working even if we are failing
						// no continue because we still need to do the deletion of this (faulty) job and reset the memory
					} else {
						// if no error occurred -> execute the new heuristics
						for _, e := range work.executors {
							if err = e.RunSynchronous(dgraph, w.currentWorkItem.txhash, "",
								w.currentWorkItem.userUid); err != nil {
								info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
							}
						}
					}
				}
				info("processing work done")
			}

			w.mutex.Lock()
			delete(w.executionMap, w.currentWorkItem)
			w.currentWorkItem = workKey{}
			w.mutex.Unlock()

			// reset memory
			work = Work{}
		}
	}
}
