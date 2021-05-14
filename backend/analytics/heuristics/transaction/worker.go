package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"

	"context"
	"errors"
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
	// cancel stops the go routine started by Start
	cancel context.CancelFunc

	// activeMutex acts as a mutex for active and cancel
	activeMutex *sync.RWMutex
	active      bool

	// mapMutex acts as a mutex for executionMap and currentWorkItem
	mapMutex        *sync.RWMutex
	currentWorkItem workKey
	executionMap    map[workKey]Work

	// graphMutex acts as a mutex for graph
	graphMutex *sync.RWMutex
	graph      *graph.ReversibleGraph
}

// NewWorker constructs a new Worker
func NewWorker() Worker {
	return Worker{executionMap: make(map[workKey]Work), mapMutex: new(sync.RWMutex),
		activeMutex: new(sync.RWMutex), graphMutex: new(sync.RWMutex)}
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
		go w.loadGraph(dgraph)
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

// IsActive returns true if the worker is active
func (w *Worker) IsActive() bool {
	w.activeMutex.RLock()
	defer w.activeMutex.RUnlock()
	return w.active
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

// GetStatus returns the current execution status of the given transaction hash and user id
func (w *Worker) GetStatus(tx string, userUid string) HeuristicQueueStatus {
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
			// check if graph is ready
			if !w.IsGraphLoaded() {
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
					if err := dbtxh.DeleteUserHeuristics(dgraph, work.removableHeuristics, w.currentWorkItem.userUid); err != nil {
						info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
						// no return/break because we want keep working even if we are failing
						// no continue because we still need to do the deletion of this (faulty) job and reset the memory
					} else {
						// if no error occurred -> execute the new heuristics
						for _, e := range work.executors {
							w.graphMutex.RLock()
							if err = e.RunSynchronous(dgraph, w.graph, w.currentWorkItem.txhash, "",
								w.currentWorkItem.userUid); err != nil {
								info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
							}
							w.graphMutex.RUnlock()
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

// IsGraphLoaded returns true if the graph is loaded
func (w *Worker) IsGraphLoaded() bool {
	w.graphMutex.RLock()
	defer w.graphMutex.RUnlock()
	return w.graph != nil
}

func (w *Worker) loadGraph(dgraph *dgo.Dgraph) {
	if w.IsGraphLoaded() {
		return
	}

	newGraph, err := graph.LoadTransactionGraph(dgraph)
	if err != nil {
		info("graph failed to load", err)
		return
	}
	w.graphMutex.Lock()
	w.graph = newGraph
	w.graphMutex.Unlock()
}

// ReverseLookup performs a reverse lookup of the given uid.
func (w *Worker) ReverseLookup(uid string, maxLookBackTime time.Duration) (map[string]bool, map[string]bool, map[string]bool,
	error) {
	if !w.IsGraphLoaded() {
		return nil, nil, nil, errors.New("graph is not loaded yet")
	}
	w.graphMutex.Lock()
	defer w.graphMutex.Unlock()
	return graph.ReverseLookup(w.graph, uid, maxLookBackTime)
}

// ForwardLookup performs a forward lookup of the given uid.
func (w *Worker) ForwardLookup(uid string, targetUid string) (map[string]bool, map[string]bool, map[string]bool,
	error) {
	if !w.IsGraphLoaded() {
		return nil, nil, nil, errors.New("graph is not loaded yet")
	}
	w.graphMutex.Lock()
	defer w.graphMutex.Unlock()
	return graph.ForwardLookup(w.graph, uid, targetUid)
}
