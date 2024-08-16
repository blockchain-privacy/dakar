package workspace

import (
	"backend/analytics/graph"
	"backend/db/workspace"
	"backend/external"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qrest/gomisc/serror"
	"log/slog"
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
	Run(*Mutex, external.Database, *graph.Wrapper) error
}

// Worker works on the data defined in Work
type Worker struct {
	jobsAdded     prometheus.Counter
	jobsError     prometheus.Counter
	jobsCompleted prometheus.Counter

	graphWrapper *graph.Wrapper
	db           external.Database

	// loopInterval is the time waited between checking if new work is available
	loopInterval time.Duration

	workspaceMutex *Mutex
}

// NewWorker constructs a new Worker
func NewWorker(m *Mutex, c external.Database, g *graph.Wrapper) *Worker {
	return &Worker{
		graphWrapper:   g,
		db:             c,
		loopInterval:   time.Second * 5,
		workspaceMutex: m,
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

// Start starts the worker. To stop the worker cancel the context.
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.loopInterval)
	defer ticker.Stop()
mainLoop:
	for {
		select {
		case <-ctx.Done():
			info("stopping Work")
			break mainLoop
		case <-ticker.C:
			// check if transaction graph is ready
			if !w.graphWrapper.IsTransactionGraphLoaded() {
				continue
			}

			items, err := GetWork(ctx, w.db)
			if err != nil {
				warn(err)
				continue
			}

			w.jobsAdded.Add(float64(len(items)))

			for _, work := range items {
				if err := work.Run(w.workspaceMutex, w.db, w.graphWrapper); err != nil {
					warn(err)
					w.jobsError.Inc()
				}
				w.jobsCompleted.Inc()
			}
		}
	}
}

// GetWork checks the database for not yet executed selectors, and constructs Work if any were found.
func GetWork(ctx context.Context, c external.Database) ([]Work, error) {
	timeoutContext, cancel := context.WithTimeout(ctx, time.Minute*2)
	defer cancel()

	selectorItems, err := workspace.GetWaitingSelectors(timeoutContext, c, 20)
	if err != nil {
		return nil, err
	}

	workItems := make([]Work, len(selectorItems))
	for i, item := range selectorItems {
		switch item.SelectorType {
		case workspace.TypeTransactionProperties:
			workItems[i], err = NewSelectorWork(item)
			if err != nil {
				return nil, err
			}
		case workspace.TypeHeuristic:
			workItems[i], err = NewHeuristicWork(item)
			if err != nil {
				return nil, err
			}
		}
	}

	return workItems, nil
}
