package graph

import (
	"backend/blockIterator"
	"backend/cmd/cliutil"
	"backend/db/status"
	"context"

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
	context context.Context
	db      *dgo.Dgraph
	state   blockIterator.State

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
func NewWrapper(ctx context.Context, dgraph *dgo.Dgraph) *Wrapper {
	return &Wrapper{context: ctx, transactionGraphMutex: new(sync.RWMutex), db: dgraph,
		addressGraphMutex: new(sync.RWMutex)}
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
func (w *Wrapper) LoadGraphs() error {
	if w.isLoading {
		return errors.New("error can not load graph as it is already loaded or still loading")
	}

	w.isLoading = true

	classifierStatus, err := status.GetClassifierStatus(w.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if *classifierStatus.IsClassifying {
		return errors.New("error can not load graphs in memory if classifier is active")
	}

	if classifierStatus.LastClassifiedBlockId == nil {
		// there are no classifications yet -> do not try to load graph
		return nil
	}

	w.state.Id = *classifierStatus.LastClassifiedBlockId
	w.state.Top = *classifierStatus.LastClassifiedBlockId

	txGraph, err := LoadTransactionGraph(w.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	addressGraph, err := LoadAddressGraph(w.db, txGraph)
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

// ------------ Block Iterator interface methods ------------

func (w *Wrapper) Logger() *log.Logger {
	return graphLogger
}

func (w *Wrapper) Context() context.Context {
	return w.context
}

func (w *Wrapper) Db() *dgo.Dgraph {
	return w.db
}

func (w *Wrapper) State() blockIterator.State {
	return w.state
}
func (w *Wrapper) Name() string {
	return "graph wrapper"
}

func (w *Wrapper) CalculateInitialState() error {
	// check if state was set by LoadGraphs
	if !w.isLoading {
		return errors.New("error graphs were not loaded before iteration started")
	}
	return nil
}

// GetHighestAvailableBlock returns the highest classified block
func (w *Wrapper) GetHighestAvailableBlock() (uint64, error) {
	classifierStatus, err := status.GetClassifierStatus(w.db)
	if err != nil || classifierStatus.LastClassifiedBlockId == nil {
		return 0, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return *classifierStatus.LastClassifiedBlockId, nil
}

func (w *Wrapper) PostExecution() error {
	// nothing to do
	return nil
}

func (w *Wrapper) IncrementState() {
	w.state.Id++
}

func (w *Wrapper) SetState(newState blockIterator.State) {
	w.state = newState
}

// Empty checks if there are more blocks above the current one
func (w *Wrapper) Empty() bool {
	return w.state.Id > w.state.Top
}

// Iterate loads the mixing transactions and all connected origin and
// destination transactions of the current block into the in-memory graphs
func (w *Wrapper) Iterate() (bool, error) {
	// todo implement
	// get per block: all mixing transactions, all destination transactions,
	// and all transactions which are connected to them
	// todo also check if starting state is correct +/-1
	return false, nil
}
