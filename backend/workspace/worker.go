// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"backend/analytics/graph"
	"backend/db"
	"backend/db/workspace"
	"backend/external"
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "workspace"}, v...)...)
}

func warn(err error, v ...any) {
	serror.Log(slog.Default(), err, v...)
}

// Work is an interface to pass a package of work to a Worker, which will process it eventually.
type Work interface {
	// Run processes the Work package. It receives the database and the graph wrapper.
	Run(context.Context, *Mutex, external.Database, *graph.Wrapper) error
}

type workItem struct {
	work Work
	done chan struct{}
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

	// workQueue is a channel which receives workItem. Items are run via workers.
	// This method of adding Work should only be used for ephemeral items, because if the server is shutdown the queue is lost.
	// For Work packages which should be persisted add them as selectors, that are picked up via GetWork().
	workQueue chan workItem

	// workerCount is the number of workers that work on workQueue
	workerCount int

	// disableDatabaseWork disables the worker which periodically retrieves selectors for the database.
	// Intended for testing.
	disableDatabaseWork bool
}

// NewWorker constructs a new Worker
func NewWorker(m *Mutex, c external.Database, g *graph.Wrapper, workerCount int) *Worker {
	return &Worker{
		graphWrapper:   g,
		db:             c,
		loopInterval:   time.Second * 5,
		workspaceMutex: m,
		workQueue:      make(chan workItem, 20),
		workerCount:    workerCount,
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
	wg := sync.WaitGroup{}
	for range w.workerCount {
		wg.Go(func() { w.startInternalWorker(ctx) })
	}

	if !w.disableDatabaseWork {
		ticker := time.Tick(w.loopInterval)
	mainLoop:
		for {
			select {
			case <-ctx.Done():
				info("stopping Work")
				break mainLoop
			case <-ticker:
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
					w.doWork(work)
				}
			}
		}
	}

	wg.Wait()
}

// AddWork returns true and a channel that returns, if the given Work item was added to the queue.
func (w *Worker) AddWork(ctx context.Context, work Work) (chan struct{}, bool) {
	ch := make(chan struct{}, 1)
	select {
	case w.workQueue <- workItem{work: work, done: ch}:
		return ch, true
	case <-ctx.Done():
	}

	return nil, false
}

// startInternalWorker starts a worker that runs work from the workQueue
func (w *Worker) startInternalWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-w.workQueue:
			w.jobsAdded.Add(1)
			w.doWork(item.work)
			item.done <- struct{}{}
		}
	}
}

// doWork runs a Work item
func (w *Worker) doWork(work Work) {
	workContext, cancel := db.GetTaskContext()
	if err := work.Run(workContext, w.workspaceMutex, w.db, w.graphWrapper); err != nil {
		warn(err)
		w.jobsError.Inc()
	}
	cancel()

	w.jobsCompleted.Inc()
}

// GetWork checks the database for not yet executed selectors, and constructs Work if any were found.
func GetWork(ctx context.Context, c external.Database) ([]Work, error) {
	timeoutContext, cancel := db.AddShortTaskContext(ctx)
	defer cancel()

	selectorItems, err := workspace.GetWaitingSelectors(timeoutContext, c, 20)
	if err != nil {
		return nil, err
	}

	workItems := make([]Work, len(selectorItems))
	for i, item := range selectorItems {
		switch item.SelectorType {
		case workspace.TypeTxProp:
			workItems[i], err = NewTxPropWork(item)
			if err != nil {
				return nil, err
			}
		case workspace.TypeTxGraph:
			workItems[i], err = NewTxGraphWork(item)
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
