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

// copyOnModify is true when existing heuristic trees should be copied before modification
const copyOnModify = false

type Work struct {
	// executors contains the HeuristicExecutor trees
	executors []HeuristicExecutor
	// removableHeuristics contains the uids of all heuristics are ready for deletion
	removableHeuristics []string
	// treeRoots contains the uids of the roots of all trees modified by the heuristics in executors
	treeRoots []string
}

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
			// get work for this cycle
			w.mapLock.Lock()
			for k, v := range w.executionMap {
				work = v
				currentTransactionHash = k
				break
			}
			w.mapLock.Unlock()

			// do we have something to do?
			if len(work.executors) > 0 || len(work.removableHeuristics) > 0 {
				log.Print("processing work package")
				// copy tree
				wasCopyingErrorFree := true
				if copyOnModify {
					for _, root := range work.treeRoots {
						if err := dbtxh.CopyHeuristicTree(dgraph, root); err != nil {
							log.Println(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
							wasCopyingErrorFree = false
							break
						}
					}
				}

				if wasCopyingErrorFree {
					// delete changed or removable heuristics
					if err := dbtxh.DeleteHeuristics(dgraph, work.removableHeuristics); err != nil {
						log.Println(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
						// no return/break because we want keep working even if we are failing
						// no continue because we still need to do the deletion of this (faulty) job and reset the memory
					} else {
						// if no error occurred -> execute the new heuristics
						for _, e := range work.executors {
							if err = e.RunSynchronous(dgraph, currentTransactionHash, ""); err != nil {
								log.Println(fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
							}
						}
					}
				}
				log.Print("processing work done")
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
