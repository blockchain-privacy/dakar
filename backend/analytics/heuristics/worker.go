package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics"
	"backend/external"

	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HeuristicQueueStatus is an enum which holds the status of the heuristic queue
type HeuristicQueueStatus int

const (
	// StatusHeuristicAdded is set if the heuristic has been successfully added
	StatusHeuristicAdded = HeuristicQueueStatus(iota)
	// StatusHeuristicDuplicate is set if the heuristic is already in the Work queue
	StatusHeuristicDuplicate
	// StatusHeuristicNotInQueue is set if the heuristic is not in the Work queue
	StatusHeuristicNotInQueue
	// StatusHeuristicInQueue is set if the heuristic is in the Work queue
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
	userUID string
}

// Work holds all Work related data for the Worker
type Work struct {
	// executors contains the heuristicExecutor trees
	executors []heuristicExecutor
	// removableHeuristics contains the uids of all heuristics are ready for deletion
	removableHeuristics []string
}

// Worker works on the data defined in Work
type Worker struct {
	jobsAdded      prometheus.Counter
	jobsCompleted  prometheus.Counter
	forwardLookups prometheus.Counter
	reverseLookups prometheus.Counter

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
	return &Worker{
		jobsAdded: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_heuristic_jobs_added_total",
			Help: "The total number of jobs added to the heuristic worker",
		}),
		jobsCompleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_heuristic_jobs_completed_total",
			Help: "The total number of jobs completed by the heuristic worker",
		}),
		reverseLookups: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_heuristic_reverse_lookups_total",
			Help: "The total number of reverse lookups executed by the heuristic worker",
		}),
		forwardLookups: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_heuristic_forward_lookups_total",
			Help: "The total number of forward lookups executed by the heuristic worker",
		}),
		executionMap: make(map[workKey]Work),
		mapMutex:     new(sync.RWMutex),
		activeMutex:  new(sync.RWMutex),
		graphWrapper: gWrapper,
	}
}

// Start starts the worker. To stop the worker cancel the context or call Stop.
// Returns false if the worker was already started.
func (w *Worker) Start(ctx context.Context, dgraph external.Database) bool {
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

// AddWork adds a Work item. Returns false if the transactionHash and userUID combination already exists in the queue.
func (w *Worker) AddWork(transactionHash string, userUID string, work Work) bool {
	key := workKey{
		txhash:  transactionHash,
		userUID: userUID,
	}

	w.mapMutex.Lock()
	defer w.mapMutex.Unlock()

	if _, exists := w.executionMap[key]; exists {
		return false
	}

	w.executionMap[key] = work

	w.jobsAdded.Inc()

	return true
}

// IsInQueue returns true if the given transaction hash and user id is in the Work queue
func (w *Worker) IsInQueue(tx string, userUID string) bool {
	key := workKey{
		txhash:  tx,
		userUID: userUID,
	}

	w.mapMutex.RLock()
	defer w.mapMutex.RUnlock()
	_, ok := w.executionMap[key]
	return ok
}

// IsReady returns true the worker is ready to Work
func (w *Worker) IsReady() bool {
	return w.graphWrapper.IsTransactionGraphLoaded()
}

// GetStatus returns the current execution status of the given transaction hash and user id
func (w *Worker) GetStatus(tx string, userUID string) HeuristicQueueStatus {
	if !w.IsReady() {
		return StatusHeuristicWorkerNotReady
	}

	key := workKey{
		txhash:  tx,
		userUID: userUID,
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
	info("stopping Work ...")
}

// Work periodically checks for new Work to be executed
func (w *Worker) work(ctx context.Context, dgraph external.Database) {

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

			// get Work for this cycle
			w.mapMutex.RLock()
			for k, v := range w.executionMap {
				work = v
				w.currentWorkItem = k
				break
			}
			w.mapMutex.RUnlock()

			// do we have something to do?
			if len(work.executors) > 0 || len(work.removableHeuristics) > 0 {
				info("processing Work package")

				// delete changed or removable heuristics
				if err := dbtxh.DeleteUserHeuristics(dgraph, work.removableHeuristics, w.currentWorkItem.userUID); err != nil {
					info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
					// no return/break because we want to keep working even if we are failing
					// no continue because we still need to do the deletion of this (faulty) job and reset the memory
				} else {
					// if no error occurred -> execute the new heuristics
					for _, e := range work.executors {
						if err = e.run(dgraph, w.graphWrapper, w.currentWorkItem.txhash, "",
							w.currentWorkItem.userUID); err != nil {
							info(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
						}
					}
				}
				w.jobsCompleted.Inc()
				info("processing Work done")
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

// ReverseLookup performs a reverse lookup for the given uid. It looks back at most maxLookBackTime
func (w *Worker) ReverseLookup(uid string, maxLookBackTime time.Duration) (map[string]bool, error) {
	w.reverseLookups.Inc()
	return w.graphWrapper.ReverseLookup(uid, maxLookBackTime, nil)
}

// ForwardLookup performs a forward lookup for the given uid. It looks forward at most maxLookForwardTime
func (w *Worker) ForwardLookup(uid string, maxLookForwardTime time.Duration) (map[string]bool, error) {
	w.forwardLookups.Inc()
	return w.graphWrapper.ForwardLookupByTime(uid, maxLookForwardTime, nil)
}
