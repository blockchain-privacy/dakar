package workspace

import (
	"backend/analytics/graph"
	"backend/external"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qrest/gomisc/serror"
	"log/slog"
	"sync"
	"time"
)

var thisLogger *slog.Logger

// InitLogger creates new loggers with the given parameters.
func InitLogger() {
	thisLogger = slog.With(slog.String("module", "workspace"))
}

func info(msg string, v ...any) {
	thisLogger.Info(msg, v...)
}

func warn(err error, v ...any) {
	serror.Log(thisLogger, err, v...)
}

// Work is an interface to pass a package of work to a Worker, which will process it eventually.
type Work interface {
	// Run processes the a Work package. It receives the database and the graph wrapper.
	Run(external.Database, *graph.Wrapper) error
}

// Worker works on the data defined in Work
type Worker struct {
	jobsAdded     prometheus.Counter
	jobsError     prometheus.Counter
	jobsCompleted prometheus.Counter

	// cancel stops the go routine started by Start
	cancel context.CancelFunc

	// activeMutex acts as a mutex for active and cancel
	activeMutex *sync.RWMutex
	active      bool

	graphWrapper *graph.Wrapper
	db           external.Database

	// loopInterval is the time waited between checking if new work is available
	loopInterval time.Duration
}

// NewWorker constructs a new Worker
func NewWorker(c external.Database, g *graph.Wrapper) *Worker {
	return &Worker{
		activeMutex:  new(sync.RWMutex),
		graphWrapper: g,
		db:           c,
		loopInterval: time.Second * 5,
	}
}

func (w *Worker) RegisterMetrics(req prometheus.Registerer) {
	w.jobsAdded = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_worker_jobs_added_total",
		Help: "The total number of jobs added to the worker",
	})
	req.MustRegister(w.jobsAdded)
	w.jobsError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_worker_jobs_error_total",
		Help: "The total number of jobs which returned an error",
	})
	req.MustRegister(w.jobsError)
	w.jobsCompleted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_worker_jobs_completed_total",
		Help: "The total number of jobs completed by the worker",
	})
	req.MustRegister(w.jobsCompleted)
}

// SetLoopInterval sets the amount of time to wait between each work loop.
// A call has no effect if the worker is already active.
func (w *Worker) SetLoopInterval(loopInterval time.Duration) {
	w.loopInterval = loopInterval
}

// Start starts the worker. To stop the worker cancel the context or call Stop.
// Returns false if the worker was already started.
func (w *Worker) Start(ctx context.Context) bool {
	w.activeMutex.Lock()
	defer w.activeMutex.Unlock()
	if !w.active {
		w.active = true
		var cancelContext context.Context
		cancelContext, w.cancel = context.WithCancel(ctx)
		go w.work(cancelContext)
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

// IsActive returns true if the worker is active
func (w *Worker) IsActive() bool {
	w.activeMutex.RLock()
	defer w.activeMutex.RUnlock()
	return w.active
}

// work periodically checks for new Work to be executed
func (w *Worker) work(ctx context.Context) {
	ticker := time.NewTicker(w.loopInterval)
	defer ticker.Stop()
mainLoop:
	for {
		select {
		case <-ctx.Done():
			info("stopping Work")
			// if worker was cancelled by context, it still needs to be set as not active
			w.Stop()
			break mainLoop
		case <-ticker.C:
			// check if transaction graph is ready
			if !w.graphWrapper.IsTransactionGraphLoaded() {
				continue
			}

			items, err := CreateWork(w.db)
			if err != nil {
				warn(err)
				continue
			}

			w.jobsAdded.Add(float64(len(items)))

			for _, work := range items {
				if err := work.Run(w.db, w.graphWrapper); err != nil {
					warn(err)
					w.jobsError.Inc()
				}
				w.jobsCompleted.Inc()
			}
		}
	}
}
