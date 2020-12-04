package transaction

import (
	"context"
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"sync"
	"time"
)

type Worker struct {
	executionMap map[string]Work
	mapLock      *sync.Mutex
}

func (w *Worker) AddWork(transactionHash string, work Work) bool {
	w.mapLock.Lock()
	defer w.mapLock.Unlock()

	if _, exists := w.executionMap[transactionHash]; exists {
		return false
	}

	w.executionMap[transactionHash] = work

	for k, v := range w.executionMap {
		log.Println(k, v)
	}
	return true
}

func NewWorker() Worker {
	var mLock sync.Mutex
	return Worker{executionMap: make(map[string]Work), mapLock: &mLock}
}

func (w *Worker) StartWorking(ctx context.Context, dgraph *dgo.Dgraph) {
	go w.work(ctx, dgraph)
}

func stoppingWorker() {
	log.Println("stopping worker")
}

func (w *Worker) work(ctx context.Context, dgraph *dgo.Dgraph) {

	var work Work
	var currentTransactionHash string
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

mainLoop:
	for {
		select {
		case <-ctx.Done():
			stoppingWorker()
			break mainLoop
		case <-ticker.C:
			w.mapLock.Lock()
			for k, v := range w.executionMap {
				work = v
				currentTransactionHash = k
				break
			}
			w.mapLock.Unlock()

			if len(work.executors) == 0 && len(work.removableHeuristics) == 0 {
				continue
			}

			if err := dbtxh.DeleteHeuristics(dgraph, work.removableHeuristics); err != nil {
				log.Println(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
			} else {
				for _, e := range work.executors {
					if err = e.RunSynchronous(dgraph, currentTransactionHash, ""); err != nil {
						log.Println(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
					}
				}
			}

			w.mapLock.Lock()
			delete(w.executionMap, currentTransactionHash)
			w.mapLock.Unlock()

			// reset memory
			currentTransactionHash = ""
			work = Work{}
		}
	}
}
