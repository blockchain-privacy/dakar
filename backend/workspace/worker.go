// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"context"
	"errors"
	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/workspace"
	"gitlab.com/blockchain-privacy/dakar/external"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	done chan error
}

// Worker works on the data defined in Work
type Worker struct {
	jobsAdded     prometheus.Counter
	jobsError     prometheus.Counter
	jobsCompleted prometheus.Counter

	// runDuration is the duration it takes to run a Work item
	runDuration prometheus.Histogram

	graphWrapper *graph.Wrapper
	db           external.Database

	// loopInterval is the time waited between checking if new work is available
	loopInterval time.Duration

	workspaceMutex *Mutex

	// workQueue is a channel which receives workItem. Items are run via workers.
	workQueue chan workItem
	// triggerSearch is a channel which when sent to, triggers a new database search for selectors.
	triggerSearch chan struct{}

	// workerCount is the number of workers that work on workQueue.
	// If set to zero, Start() will spawn GOMAXPROCS/2 workers.
	workerCount int

	// waitForInMemoryGraph makes Start() wait until the in-memory graph is loaded, before starting the workers.
	// It also enables the selector search worker. If disabled, workers might panic if the graph wrapper is not set.
	// Intended for testing.
	waitForInMemoryGraph bool
}

// NewWorker constructs a new Worker
func NewWorker(m *Mutex, c external.Database, g *graph.Wrapper) *Worker {
	return &Worker{
		graphWrapper:         g,
		db:                   c,
		loopInterval:         time.Second * 5,
		workspaceMutex:       m,
		workQueue:            make(chan workItem, 30), // max items getWork() returns is 20, so set it slightly higher
		triggerSearch:        make(chan struct{}),
		waitForInMemoryGraph: true,
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
	w.runDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dakar_worker_run_duration",
		Help:    "The duration it takes to run a work item",
		Buckets: []float64{0.5, 1, 2, 5, 10, 20, 30},
	})
	req.MustRegister(w.runDuration)
}

// SetLoopInterval sets the amount of time to wait between each work loop.
// A call has no effect if the worker is already active.
func (w *Worker) SetLoopInterval(loopInterval time.Duration) {
	w.loopInterval = loopInterval
}

// SetWaitForInMemoryGraph sets if Start() should wait until the in-memory graph is loaded, before starting the workers.
func (w *Worker) SetWaitForInMemoryGraph(flag bool) {
	w.waitForInMemoryGraph = flag
}

// SetWorkerCount sets the number of workers. Must be called before Start() to have an effect.
// If workerCount is set to zero, Start() will spawn GOMAXPROCS/2 workers.
func (w *Worker) SetWorkerCount(workerCount int) {
	w.workerCount = workerCount
}

// TriggerSearch tries to trigger a search for selectors. If successful, this skips the selector search timer.
// This is a no-op, if another trigger is already in progress.
func (w *Worker) TriggerSearch() {
	select {
	case w.triggerSearch <- struct{}{}:
	default:
	}
}

// waitForGraph returns if the graph is loaded
func (w *Worker) waitForGraph(ctx context.Context) bool {
	ticker := time.Tick(w.loopInterval)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			// check if transaction graph is ready
			if w.graphWrapper.IsTransactionGraphLoaded() {
				return true
			}
		}
	}
}

// Start starts the worker. To stop the worker cancel the context.
func (w *Worker) Start(ctx context.Context) {
	wg := sync.WaitGroup{}
	if w.waitForInMemoryGraph {
		w.waitForGraph(ctx)
		wg.Go(func() { w.findWorkInDatabase(ctx) })
	}

	workerCount := max(1, runtime.GOMAXPROCS(0)/2)
	if w.workerCount > 0 {
		workerCount = w.workerCount
	}

	info("starting workers", "count", workerCount)

	for range workerCount {
		wg.Go(func() { w.startWorker(ctx) })
	}

	wg.Wait()
}

// AddWork a non-nil channel, if the given Work item was queued. The channel returns when the Work has been run.
func (w *Worker) AddWork(ctx context.Context, work Work) chan error {
	ch := make(chan error, 1)
	select {
	case w.workQueue <- workItem{work: work, done: ch}:
		w.jobsAdded.Add(1)
		return ch
	case <-ctx.Done():
	}

	return nil
}

// startWorker starts a worker that runs work from the workQueue
func (w *Worker) startWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-w.workQueue:
			item.done <- w.doWork(ctx, item.work)
		}
	}
}

// doWork runs a Work item
func (w *Worker) doWork(ctx context.Context, work Work) error {
	timer := prometheus.NewTimer(w.runDuration)
	defer timer.ObserveDuration()
	workContext, cancel := db.AddTaskContext(ctx)
	defer cancel()
	if err := work.Run(workContext, w.workspaceMutex, w.db, w.graphWrapper); err != nil {
		warn(err)
		w.jobsError.Inc()
		return err
	}

	w.jobsCompleted.Inc()
	return nil
}

// findWorkInDatabase loops:
// - searches database for waiting selectors
// - adds them to the worker queue
// - waits for them to be finished
func (w *Worker) findWorkInDatabase(ctx context.Context) {
	ticker := time.Tick(w.loopInterval)

	for {
		select {
		case <-ctx.Done():
			info("stopping Work")
			return
		case <-w.triggerSearch:
		case <-ticker:
		}

		items, err := getWork(ctx, w.db)
		if err != nil {
			// print error only, if errors was not due to the context being cancelled
			if !errors.Is(err, context.Canceled) && status.Code(err) != codes.Canceled {
				warn(err)
			}

			continue
		}
		var waitingChannels []chan error
		// queue work
		for _, work := range items {
			if ch := w.AddWork(ctx, work); ch != nil {
				waitingChannels = append(waitingChannels, ch)
			}
		}

		// wait for all work to be finished before starting the next loop
		for _, ch := range waitingChannels {
			select {
			case <-ctx.Done():
				info("stopping Work")
				return
			case <-ch:
				// if there is an error, ignore it. In this case the warning was already printed in the worker
			}
		}
	}
}

// getWork checks the database for not yet executed selectors, and constructs Work if any were found.
func getWork(ctx context.Context, c external.Database) ([]Work, error) {
	timeoutContext, cancel := db.AddShortTaskContext(ctx)
	defer cancel()

	selectorItems, err := workspace.GetWaitingSelectors(timeoutContext, c, 20)
	if err != nil {
		return nil, err
	}

	workItems := make([]Work, len(selectorItems))
	for i, item := range selectorItems {
		switch item.SelectorType {
		case constants.TypeTxProp:
			workItems[i], err = NewTxPropWork(item)
			if err != nil {
				return nil, err
			}
		case constants.TypeTxGraph:
			workItems[i], err = NewTxGraphWork(item)
			if err != nil {
				return nil, err
			}
		case constants.TypeHeuristic:
			workItems[i], err = NewHeuristicWork(item)
			if err != nil {
				return nil, err
			}
		}
	}

	return workItems, nil
}
