package worker

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/external"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"log/slog"
	"math/rand/v2"

	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var thisLogger *slog.Logger

// InitLogger creates new loggers with the given parameters.
func InitLogger() {
	thisLogger = slog.With(slog.String("module", "worker"))
}

func info(msg string, v ...any) {
	thisLogger.Info(msg, v...)
}

func warn(err error, v ...any) {
	cliutil.LogError(thisLogger, err, v...)
}

type workKey struct {
	// todo rethink limits of queue per user
	userUID string
	// workID is used to identify the work package per user
	workID int
}

func (w workKey) toString() string {
	return fmt.Sprintf("%s|%d", w.userUID, w.workID)
}

// Work is an interface to pass a package of work to a Worker, which will process it eventually.
type Work interface {
	// Run processes the a Work package. If available, it returns the resulting UID of the stored data.
	Run(external.Database, *graph.Wrapper) (string, error)
}

// Worker works on the data defined in Work
type Worker struct {
	jobsAdded     prometheus.Counter
	jobsCompleted prometheus.Counter

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

	// workLog is used to determine if a workID exist at all in the current server lifecycle
	workLog      *ristretto.Cache
	finishedWork *ristretto.Cache
}

// NewWorker constructs a new Worker
func NewWorker(gWrapper *graph.Wrapper) (*Worker, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 5000, // number of keys to track frequency of
		MaxCost:     500,  // number of mappings
		BufferItems: 64,   // number of keys per Get buffer
	})
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}
	workLog, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 5000, // number of keys to track frequency of
		MaxCost:     500,  // number of mappings
		BufferItems: 64,   // number of keys per Get buffer
	})
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	return &Worker{
		jobsAdded: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_worker_jobs_added_total",
			Help: "The total number of jobs added to the worker",
		}),
		jobsCompleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_worker_jobs_completed_total",
			Help: "The total number of jobs completed by the worker",
		}),
		executionMap: make(map[workKey]Work),
		mapMutex:     new(sync.RWMutex),
		activeMutex:  new(sync.RWMutex),
		graphWrapper: gWrapper,
		workLog:      workLog,
		finishedWork: cache,
	}, nil
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

// Stop stops the worker. The worker can also be stopped by cancelling the passed context to Start
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
func (w *Worker) AddWork(userUID string, work Work) int {
	w.mapMutex.Lock()
	defer w.mapMutex.Unlock()

	key := workKey{userUID: userUID}
	for {
		// no secure random number is needed here
		key.workID = rand.Int() //nolint:gosec
		if _, exists := w.executionMap[key]; !exists {
			break
		}
	}

	w.executionMap[key] = work
	w.jobsAdded.Inc()
	w.workLog.SetWithTTL(key.toString(), 1, 1, 12*time.Hour)

	return key.workID
}

// GetFinishedDatabaseUID returns the database UID if the Work package specified by its id is finished executing.
// Returns an empty string, if the work package is not found in the list of finished Work packages.
func (w *Worker) GetFinishedDatabaseUID(workID int, userUID string) (string, error) {
	databaseUIDInterface, ok := w.finishedWork.Get(workKey{workID: workID, userUID: userUID}.toString())
	if !ok {
		return "", nil
	}

	databaseUID, ok := databaseUIDInterface.(string)
	if !ok {
		return "", cliutil.NewStackErrorStr("not able to convert cache item to string")
	}

	return databaseUID, nil
}

// DoesWorkExist returns true if the given workID exists in the worklog
func (w *Worker) DoesWorkExist(workID int, userUID string) bool {
	_, ok := w.workLog.Get(workKey{workID: workID, userUID: userUID}.toString())
	return ok
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
			if len(w.executionMap) == 0 {
				w.mapMutex.RUnlock()
				continue
			}
			w.currentWorkItem, work = cliutil.GetOneItem(w.executionMap)
			w.mapMutex.RUnlock()

			databaseUID, err := work.Run(dgraph, w.graphWrapper)
			if err != nil {
				warn(err)
			}

			w.jobsCompleted.Inc()

			w.mapMutex.Lock()
			w.finishedWork.SetWithTTL(w.currentWorkItem.toString(), databaseUID, 1, 12*time.Hour)

			delete(w.executionMap, w.currentWorkItem)
			w.currentWorkItem = workKey{}
			w.mapMutex.Unlock()
		}
	}
}
