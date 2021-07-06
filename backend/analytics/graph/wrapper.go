package graph

import (
	"backend/blockiterator"
	"backend/cmd/cliutil"
	"backend/db/analytics"
	"backend/db/status"
	"backend/external"

	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// graphLoggerPrefix is the prefix which is printed for each log message of analyticsLogger
const graphLoggerPrefix = "\033[0;32mgraph\u001B[0m\t"

var graphLogger = log.New(log.Writer(), graphLoggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	graphLogger = log.New(out, graphLoggerPrefix, flag)
}

func info(v ...interface{}) {
	graphLogger.Println(v...)
}

// Wrapper is wrapper for in-memory graphs
type Wrapper struct {
	context      context.Context
	db           external.Database
	state        blockiterator.State
	blocks       prometheus.Counter
	transactions prometheus.Counter
	newUids      prometheus.Counter
	blockHeight  prometheus.Gauge

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
func NewWrapper(ctx context.Context, dgraph external.Database) *Wrapper {
	return &Wrapper{
		context:               ctx,
		transactionGraphMutex: new(sync.RWMutex),
		db:                    dgraph,
		addressGraphMutex:     new(sync.RWMutex),
		blocks: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_graph_blocks_processed_total",
			Help: "The total number of blocks processed by the graph wrapper",
		}),
		transactions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_graph_transactions_processed_total",
			Help: "The total number of transactions processed by the graph wrapper",
		}),
		newUids: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_graph_newuids_processed_total",
			Help: "The total number of new uids processed by the graph wrapper",
		}),
		blockHeight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "dakar_graph_last_block",
			Help: "The last processed block by the graph wrapper",
		}),
	}
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
func (w *Wrapper) ReverseLookup(uid string, maxLookBackTime time.Duration) (map[string]bool, error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, errors.New("transaction graph is not loaded yet")
	}
	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()

	results, err := ReverseLookup(w.transactionGraph, uid, maxLookBackTime)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	return results, nil
}

// ForwardLookup performs a forward lookup of the given uid.
func (w *Wrapper) ForwardLookup(uid string, targetUID string) (map[string]bool, error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, errors.New("transaction graph is not loaded yet")
	}
	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()

	results, err := ForwardLookup(w.transactionGraph, uid, targetUID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return results, nil
}

// ForwardLookupByTime performs a forward lookup of the given uid.
func (w *Wrapper) ForwardLookupByTime(uid string, maxLookForwardTime time.Duration) (map[string]bool, error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, errors.New("transaction graph is not loaded yet")
	}
	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()

	results, err := ForwardLookupByTime(w.transactionGraph, uid, maxLookForwardTime)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return results, nil
}

// GetClusters returns a mapping between address uids and ClusterID's
func (w *Wrapper) GetClusters(addressUids []string) (map[string]ClusterID, error) {
	if !w.IsAddressGraphLoaded() {
		return nil, errors.New("address graph is not loaded yet")
	}
	w.addressGraphMutex.Lock()
	defer w.addressGraphMutex.Unlock()

	results, err := GetClusters(w.addressGraph, addressUids)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return results, nil
}

// GetCluster returns the cluster of the given address
func (w *Wrapper) GetCluster(addressUID string) ([]string, error) {
	if !w.IsAddressGraphLoaded() {
		return nil, errors.New("address graph is not loaded yet")
	}
	w.addressGraphMutex.Lock()
	defer w.addressGraphMutex.Unlock()

	results, err := GetCluster(w.addressGraph, addressUID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return results, nil
}

// GetInputTransactions returns the uids of all directly connected input transactions of the tx specified by uid
func (w *Wrapper) GetInputTransactions(uid string) ([]string, error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, errors.New("transaction graph is not loaded yet")
	}
	w.addressGraphMutex.Lock()
	defer w.addressGraphMutex.Unlock()

	results, err := GetInputTransactions(w.transactionGraph, uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return results, err
}

// LoadGraphs loads the transaction and address graph into the wrapper
func (w *Wrapper) LoadGraphs() error {
	if w.isLoading {
		return errors.New("error can not load graph as it is already loaded or still loading")
	}

	w.isLoading = true

	classifierStatus, err := status.GetClassifierStatus(w.db)
	if err != nil {
		if errors.Is(err, status.ErrorStatusNotFound) {
			info("Classifier status is not set. Classify at least one block before starting to load graphs.")
			return nil
		}
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if classifierStatus.LastClassifiedBlockID == nil {
		// there are no classifications yet -> do not try to load graph
		return nil
	}

	w.state.ID = *classifierStatus.LastClassifiedBlockID + 1
	w.state.Top = *classifierStatus.LastClassifiedBlockID

	// state.ID - 1 because the ID is the next block
	w.blockHeight.Set(float64(w.state.ID - 1))

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

// Logger returns the Logger
func (w *Wrapper) Logger() *log.Logger {
	return graphLogger
}

// Context returns the context
func (w *Wrapper) Context() context.Context {
	return w.context
}

// Name returns the name
func (w *Wrapper) Name() string {
	return "graph wrapper"
}

// CalculateInitialState calculates the state on which the iterator starts processing
func (w *Wrapper) CalculateInitialState() error {
	// check if state was set by LoadGraphs
	if !w.isLoading {
		return errors.New("error graphs were not loaded before iteration started")
	}
	return nil
}

// NextBlock tries to increase the internal state to the next block
func (w *Wrapper) NextBlock() (bool, error) {
	classifierStatus, err := status.GetClassifierStatus(w.db)
	if err != nil || classifierStatus.LastClassifiedBlockID == nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return w.state.ID <= *classifierStatus.LastClassifiedBlockID, nil
}

// CurrentBlock returns the height of the block which is currently being loaded
func (w *Wrapper) CurrentBlock() uint64 {
	return w.state.ID
}

// PostExecution does nothing
func (w *Wrapper) PostExecution() error {
	// nothing to do
	return nil
}

// IncrementState increments the state one block
func (w *Wrapper) IncrementState() error {
	w.state.ID++
	return nil
}

// Empty checks if there are more blocks above the current one
func (w *Wrapper) Empty() bool {
	return w.state.ID > w.state.Top
}

// Iterate loads the mixing transactions and all connected origin and
// destination transactions of the current block into the in-memory graphs
func (w *Wrapper) Iterate() (bool, error) {
	connectedNodes, singleNodes, err := analytics.GetPrivacyTransactionsByBlock(w.db, w.state.ID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if len(connectedNodes) == 0 && len(singleNodes) == 0 {
		// nothing to do
		return true, nil
	}

	if len(connectedNodes) == 0 || len(singleNodes) == 0 {
		// something is wrong
		return false, errors.New("error count of single or connected nodes is zero")
	}

	w.addressGraphMutex.Lock()
	w.transactionGraphMutex.Lock()
	defer w.addressGraphMutex.Unlock()
	defer w.transactionGraphMutex.Unlock()

	if graphErr := upsertSingleNodes(w.transactionGraph, singleNodes); graphErr != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), graphErr)
	}

	if graphErr := addEdges(w.transactionGraph, connectedNodes); graphErr != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), graphErr)
	}

	nodeUidsToLoad, err := filterOrigins(w.transactionGraph, w.addressGraph, singleNodes)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if len(nodeUidsToLoad) > 0 {
		originNodes, dbErr := analytics.GetInputAddresses(w.db, nodeUidsToLoad)
		if dbErr != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), dbErr)
		}

		if len(originNodes) != len(nodeUidsToLoad) {
			return false, errors.New("error number of requested origin nodes does not match number of received nodes")
		}

		if addErr := addAddressEdges(w.addressGraph, originNodes); addErr != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), addErr)
		}
	}

	w.blocks.Inc()
	w.transactions.Add(float64(len(connectedNodes) + len(singleNodes)))
	w.newUids.Add(float64(len(nodeUidsToLoad)))
	w.blockHeight.Set(float64(w.state.ID))

	return true, nil
}

// filterOrigins returns the nodes which are not already in the address graph and
// which are endpoints in the transaction graph
func filterOrigins(txGraph *ReversibleGraph, addrGraph *UndirectedGraph, nodes []analytics.Node) ([]string, error) {
	var originUids []string
	for _, node := range nodes {
		id, err := toInteger(node.UID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// check if node is already in address graph
		if addrGraph.Node(id) != nil {
			continue
		}

		// check if node is an endpoint
		if isEndpoint(txGraph, id) {
			originUids = append(originUids, node.UID)
		}
	}

	return originUids, nil
}
