package graph

import (
	"backend/cmd/cliutil"

	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/dgo/v2"
)

// graphLoggerPrefix is the prefix which is printed for each log message of analyticsLogger
const graphLoggerPrefix = "\033[0;32mgraph\u001B[0m\t"

var graphLogger = log.New(log.Writer(), graphLoggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	graphLogger = log.New(out, graphLoggerPrefix, flag)

}

func info(v ...interface{}) {
	graphLogger.Println(v)
}

type Wrapper struct {
	// isLoading is true if the graph loading was started.
	// It stays true even if the graphs are finished loading to prevent loading more than once.
	isLoading bool

	// transactionGraphMutex acts as a mutex for transactionGraph
	transactionGraphMutex *sync.RWMutex
	transactionGraph      *ReversibleGraph

	// addressGraphMutex acts as a mutex for addressGraph
	addressGraphMutex *sync.RWMutex
	addressGraph      *UndirectedGraph
}

// NewWrapper constructs a new Wrapper
func NewWrapper() *Wrapper {
	return &Wrapper{transactionGraphMutex: new(sync.RWMutex), addressGraphMutex: new(sync.RWMutex)}
}

// IsTransactionGraphLoaded returns true if the transaction graph is loaded
func (w *Wrapper) IsTransactionGraphLoaded() bool {
	w.transactionGraphMutex.RLock()
	defer w.transactionGraphMutex.RUnlock()
	return w.transactionGraph != nil
}

// IsAddressGraphLoaded returns true if the address graph is loaded
func (w *Wrapper) IsAddressGraphLoaded() bool {
	w.addressGraphMutex.RLock()
	defer w.addressGraphMutex.RUnlock()
	return w.addressGraph != nil
}

// ReverseLookup performs a reverse lookup of the given uid.
func (w *Wrapper) ReverseLookup(uid string, maxLookBackTime time.Duration) (map[string]bool, map[string]bool, map[string]bool,
	error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, nil, nil, errors.New("transaction graph is not loaded yet")
	}
	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()
	return ReverseLookup(w.transactionGraph, uid, maxLookBackTime)
}

// ForwardLookup performs a forward lookup of the given uid.
func (w *Wrapper) ForwardLookup(uid string, targetUid string) (map[string]bool, map[string]bool, map[string]bool,
	error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, nil, nil, errors.New("transaction graph is not loaded yet")
	}
	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()
	return ForwardLookup(w.transactionGraph, uid, targetUid)
}

// GetClusters returns a mapping between address uids and ClusterId's
func (w *Wrapper) GetClusters(addressUids map[string]bool) (map[string]ClusterId, error) {
	if !w.IsAddressGraphLoaded() {
		return nil, errors.New("address graph is not loaded yet")
	}
	w.addressGraphMutex.Lock()
	defer w.addressGraphMutex.Unlock()
	return GetClusters(w.addressGraph, addressUids)
}

// GetInputTransactions returns the uids of all directly connected input transactions of the tx specified by uid
func (w *Wrapper) GetInputTransactions(uid string) ([]string, error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, errors.New("transaction graph is not loaded yet")
	}
	w.addressGraphMutex.Lock()
	defer w.addressGraphMutex.Unlock()
	return GetInputTransactions(w.transactionGraph, uid)
}

// LoadGraphs loads the transaction and address graph into the wrapper
func (w *Wrapper) LoadGraphs(dgraph *dgo.Dgraph) error {
	if w.isLoading {
		return errors.New("error can not load graph as it is already loaded or still loading")
	}

	txGraph, err := LoadTransactionGraph(dgraph)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	addressGraph, err := LoadAddressGraph(dgraph, txGraph)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	w.transactionGraphMutex.Lock()
	w.transactionGraph = txGraph
	w.transactionGraphMutex.Unlock()

	w.addressGraphMutex.Lock()
	w.addressGraph = addressGraph
	w.addressGraphMutex.Unlock()

	return nil
}
