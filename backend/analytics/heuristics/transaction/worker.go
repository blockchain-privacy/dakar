package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"

	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/dgo/v210"
)

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
	// StatusHeuristicWorkerNotReady is set if the heuristic worker is not ready yet
	StatusHeuristicWorkerNotReady

	// loggerPrefix is the prefix which is printed for each log message
	loggerPrefix = "\033[0;34mhworker\u001B[0m\t"
)

var thisLogger = log.New(log.Writer(), loggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	thisLogger = log.New(out, loggerPrefix, flag)
}

func info(v ...interface{}) {
	thisLogger.Println(v...)
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
	// cancel stops the go routine started by Start
	cancel context.CancelFunc

	// activeMutex acts as a mutex for active and cancel
	activeMutex *sync.RWMutex
	active      bool

	// mapMutex acts as a mutex for executionMap and currentWorkItem
	mapMutex        *sync.RWMutex
	currentWorkItem workKey
	executionMap    map[workKey]Work

	// graphWrapper gives access to graph functions
	graphWrapper *graph.Wrapper
}

// NewWorker constructs a new Worker
func NewWorker(gWrapper *graph.Wrapper) *Worker {
	return &Worker{executionMap: make(map[workKey]Work), mapMutex: new(sync.RWMutex), activeMutex: new(sync.RWMutex),
		graphWrapper: gWrapper}
}

// Start starts the worker. To stop the worker cancel the context or call Stop.
// Returns false if the worker was already started.
func (w *Worker) Start(ctx context.Context, dgraph *dgo.Dgraph) bool {
	w.activeMutex.Lock()
	defer w.activeMutex.Unlock()
	if !w.active {
		w.active = true
		var cancelContext context.Context
		cancelContext, w.cancel = context.WithCancel(ctx)
		go w.work(cancelContext, dgraph)
		return true
	}
	return false
}

// Stop stops the worker.
func (w *Worker) Stop() {
	w.activeMutex.Lock()
	defer w.activeMutex.Unlock()
	if !w.active {
		return
	}

	w.cancel()
	w.active = false
}

func (w *Worker) AddWork(transactionHash string, userUid string, work Work) bool {
	key := workKey{
		txhash:  transactionHash,
		userUid: userUid,
	}

	w.mapMutex.Lock()
	defer w.mapMutex.Unlock()

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

	w.mapMutex.RLock()
	defer w.mapMutex.RUnlock()
	_, ok := w.executionMap[key]
	return ok
}

// IsReady returns true the worker is ready to work
func (w *Worker) IsReady() bool {
	return w.graphWrapper.IsAddressGraphLoaded() && w.graphWrapper.IsTransactionGraphLoaded()
}

// GetStatus returns the current execution status of the given transaction hash and user id
func (w *Worker) GetStatus(tx string, userUid string) HeuristicQueueStatus {
	if !w.IsReady() {
		return StatusHeuristicWorkerNotReady
	}

	key := workKey{
		txhash:  tx,
		userUid: userUid,
	}

	w.mapMutex.RLock()
	defer w.mapMutex.RUnlock()
	_, ok := w.executionMap[key]

	if !ok {
		return StatusHeuristicNotInQueue
	}

	if w.currentWorkItem == key {
		return StatusHeuristicProcessing
	}

	return StatusHeuristicInQueue
}

func stoppingWork() {
	info("stopping work ...")
}

// work periodically checks for new Work to be executed
func (w *Worker) work(ctx context.Context, dgraph *dgo.Dgraph) {

	var work Work
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

mainLoop:
	for {
		select {
		case <-ctx.Done():
			stoppingWork()
			break mainLoop
		case <-ticker.C:
			// check if transaction graph is ready
			if !w.graphWrapper.IsTransactionGraphLoaded() {
				continue
			}

			// get work for this cycle
			w.mapMutex.RLock()
			for k, v := range w.executionMap {
				work = v
				w.currentWorkItem = k
				break
			}
			w.mapMutex.RUnlock()

			// do we have something to do?
			if len(work.executors) > 0 || len(work.removableHeuristics) > 0 {
				info("processing work package")

				// delete changed or removable heuristics
				if err := dbtxh.DeleteUserHeuristics(dgraph, work.removableHeuristics, w.currentWorkItem.userUid); err != nil {
					info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
					// no return/break because we want keep working even if we are failing
					// no continue because we still need to do the deletion of this (faulty) job and reset the memory
				} else {
					// if no error occurred -> execute the new heuristics
					for _, e := range work.executors {
						if err = e.RunSynchronous(dgraph, w.graphWrapper, w.currentWorkItem.txhash, "",
							w.currentWorkItem.userUid); err != nil {
							info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
						}
					}
				}

				info("processing work done")
			}

			w.mapMutex.Lock()
			delete(w.executionMap, w.currentWorkItem)
			w.currentWorkItem = workKey{}
			w.mapMutex.Unlock()

			// reset memory
			work = Work{}
		}
	}
}

func (w *Worker) ReverseLookup(uid string, maxLookBackTime time.Duration) (map[string]bool, error) {
	return w.graphWrapper.ReverseLookup(uid, maxLookBackTime)
}

func (w *Worker) ForwardLookup(uid string, targetUid string) (map[string]bool, error) {
	return w.graphWrapper.ForwardLookup(uid, targetUid)
}

func (w *Worker) ForwardLookupByTime(uid string, maxLookForwardTime time.Duration) (map[string]bool, error) {
	return w.graphWrapper.ForwardLookupByTime(uid, maxLookForwardTime)
}

func (w *Worker) GetCluster(addressUid string) ([]string, error) {
	return w.graphWrapper.GetCluster(addressUid)
}
