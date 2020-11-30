package transaction

import (
	"context"
	"dashrpc/cmd/cliutil"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"sync"
	"time"
)

type Worker struct {
	executionMap map[string][]HeuristicExecutor
	mapLock      *sync.Mutex
}

func (w *Worker) AddWork(transactionHash string, executors []HeuristicExecutor) bool {
	w.mapLock.Lock()
	defer w.mapLock.Unlock()

	if _, exists := w.executionMap[transactionHash]; exists {
		return false
	}

	w.executionMap[transactionHash] = executors

	for k, v := range w.executionMap {
		log.Println(k, v)
	}
	return true
}

func NewWorker() Worker {
	var mLock sync.Mutex
	return Worker{executionMap: make(map[string][]HeuristicExecutor), mapLock: &mLock}
}

func (w *Worker) StartWorking(ctx context.Context, dgraph *dgo.Dgraph) {
	go w.work(ctx, dgraph)
}

func stoppingWorker() {
	log.Println("stopping worker")
}

func (w *Worker) work(ctx context.Context, dgraph *dgo.Dgraph) {

	var work []HeuristicExecutor
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

			if len(work) == 0 {
				log.Println("Found nothing. Sleeping ...")
				continue
			}

			for _, e := range work {
				if runErr := e.RunSynchronous(dgraph, currentTransactionHash, ""); runErr != nil {
					log.Println(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), runErr))
				}
			}

			w.mapLock.Lock()
			delete(w.executionMap, currentTransactionHash)
			w.mapLock.Unlock()

			// reset memory
			currentTransactionHash = ""
			work = nil
		}
	}
}
