// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"backend/blockiterator"
	"backend/db"
	"backend/db/analytics"
	"backend/db/status"
	"backend/external"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "graph"}, v...)...)
}

// Wrapper is wrapper for in-memory graphs
type Wrapper struct {
	context      context.Context
	db           external.Database
	state        blockiterator.State
	blocks       prometheus.Counter
	transactions prometheus.Counter
	blockHeight  prometheus.Gauge

	// isLoading is true if the graph loading was started.
	// It stays true even if the graphs are finished loading to prevent loading more than once.
	isLoading bool

	// transactionGraphMutex acts as a mutex for transactionGraph
	transactionGraphMutex *sync.RWMutex
	transactionGraph      *ReversibleGraph
}

// NewWrapper constructs a new Wrapper
func NewWrapper(ctx context.Context, dgraph external.Database) *Wrapper {
	return &Wrapper{
		context:               ctx,
		transactionGraphMutex: new(sync.RWMutex),
		db:                    dgraph,
	}
}

func (w *Wrapper) RegisterMetrics(req prometheus.Registerer) {
	w.blocks = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_graph_blocks_processed_total",
		Help: "The total number of blocks processed by the graph wrapper",
	})
	req.MustRegister(w.blocks)

	w.transactions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_graph_transactions_processed_total",
		Help: "The total number of transactions processed by the graph wrapper",
	})
	req.MustRegister(w.transactions)

	w.blockHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dakar_graph_last_block",
		Help: "The last processed block by the graph wrapper",
	})
	req.MustRegister(w.blockHeight)
}

// SetMaxBlocks is not supported, so do nothing
func (w *Wrapper) SetMaxBlocks(int64) {}

// IsTransactionGraphLoaded returns true if the transaction graph is loaded
func (w *Wrapper) IsTransactionGraphLoaded() bool {
	w.transactionGraphMutex.RLock()
	defer w.transactionGraphMutex.RUnlock()

	return w.transactionGraph != nil
}

// ReverseLookup performs a reverse lookup of the given uid.
func (w *Wrapper) ReverseLookup(uid string, maxLookBackTime time.Duration, maxDepth int,
	addressExclusions []string, excludeSpendingGaps bool) (map[string]bool, error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, serror.FromStr("transaction graph is not loaded yet")
	}

	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()

	return ReverseLookup(w.transactionGraph, uid, maxLookBackTime, maxDepth, addressExclusions, excludeSpendingGaps)
}

// ForwardLookup performs a forward lookup of the given uid.
func (w *Wrapper) ForwardLookup(uid string, maxLookForwardTime time.Duration, maxDepth int,
	addressExclusions []string, excludeSpendingGaps bool) (map[string]bool, error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, serror.FromStr("transaction graph is not loaded yet")
	}

	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()

	return ForwardLookup(w.transactionGraph, uid, maxLookForwardTime, maxDepth,
		addressExclusions, excludeSpendingGaps)
}

// PartitionNodesByDirectConnections partitions the given nodes into groups which can reach each other by direct links
func (w *Wrapper) PartitionNodesByDirectConnections(nodes []string) ([][]string, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	nodeSet := map[int64]bool{}
	for _, n := range nodes {
		i, err := ToInteger(n)
		if err != nil {
			return nil, err
		}
		nodeSet[i] = true
	}

	// need to write look, because the graph becomes undirected for a short time
	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()

	var partitions [][]string
	for n := range nodeSet {
		neighbours, err := getConnectedNodes(w.transactionGraph, n, nodeSet)
		if err != nil {
			return nil, err
		}

		// neighbours always contains at least one node (n), so no need to check for length
		part := make([]string, len(neighbours))
		for i, neighbour := range neighbours {
			part[i] = ToHex(neighbour)
			delete(nodeSet, neighbour)
		}

		partitions = append(partitions, part)
	}
	return partitions, nil
}

// SpendingFingerprint returns a list of transaction uids which have a similar spending pattern
// and the number of mixing sessions of this transactions
func (w *Wrapper) SpendingFingerprint(uid string) ([]FingerPrint, int, error) {
	if !w.IsTransactionGraphLoaded() {
		return nil, -1, serror.FromStr("transaction graph is not loaded yet")
	}

	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()

	return SpendingFingerprint(w.transactionGraph, uid, 30)
}

// LoadGraphs loads the transaction graph into the wrapper
func (w *Wrapper) LoadGraphs(config Config) error {
	if w.isLoading {
		return serror.FromStr("can not load graph as it is already loaded or still loading")
	}

	if w.db == nil {
		return serror.FromStr("database is not set")
	}

	w.isLoading = true

	ctx, cancel := db.GetLongTaskContext()
	defer cancel()

	classifierStatus, err := status.GetClassifierStatus(ctx, w.db)
	if err != nil {
		if errors.Is(err, status.ErrStatusNotFound) {
			info("Classifier status is not set. Classify at least one block before starting to load graphs.")
			return nil
		}
		return err
	}

	if classifierStatus.LastClassifiedBlockID == nil {
		// there are no classifications yet -> do not try to load graph
		return nil
	}

	w.state.ID = *classifierStatus.LastClassifiedBlockID + 1
	w.state.Top = *classifierStatus.LastClassifiedBlockID

	// state.ID - 1 because the ID is the next block
	w.blockHeight.Set(float64(w.state.ID - 1))

	numTxToLoad := 0

	if graphLimit, ok := os.LookupEnv("DEV_GRAPH_LIMIT"); ok {
		numGraphLimit, err := strconv.Atoi(graphLimit)
		if err != nil {
			return serror.FromFormat("DEV_GRAPH_LIMIT is not a number: %w", err)
		}
		if numGraphLimit < 0 {
			info("DEV_GRAPH_LIMIT environment variable is negative. Exiting ...")
			return serror.FromStr("negative DEV_GRAPH_LIMIT environment variable")
		} else if numGraphLimit == 0 {
			info("DEV_GRAPH_LIMIT environment variable is set to zero. Ignoring ...")
		} else {
			info(fmt.Sprintf("DEV_GRAPH_LIMIT environment variable is set. Limiting in-memory mixing graph to %d transactions", numGraphLimit))
			numTxToLoad = numGraphLimit
		}
	}

	txGraph, err := LoadTransactionGraph(ctx, config, w.db, numTxToLoad)
	if err != nil {
		if errors.Is(err, ErrDBContainsNoClassifiedTransactions) {
			return nil
		}
		return err
	}

	w.SetGraph(txGraph)

	return nil
}

func (w *Wrapper) SetGraph(txGraph *ReversibleGraph) {
	w.transactionGraphMutex.Lock()
	w.transactionGraph = txGraph
	w.transactionGraphMutex.Unlock()
}

// ------------ Block Iterator interface methods ------------

func (w *Wrapper) Props() blockiterator.Properties {
	return blockiterator.Properties{
		Name:                        "graph wrapper",
		Context:                     w.context,
		CurrentBlock:                w.state.ID,
		ProcessedBlockCount:         1,
		SupportsMultiBlockIteration: false,
	}
}

// CalculateInitialState is a dummy which only checks if LoadGraphs
// has been executed. LoadGraphs set the initial state.
func (w *Wrapper) CalculateInitialState(context.Context) error {
	// check if state was set by LoadGraphs
	if !w.isLoading {
		return serror.FromStr("graphs were not loaded before iteration started")
	}

	return nil
}

// Next tries to increase the internal state to the next block
func (w *Wrapper) Next(ctx context.Context) (bool, error) {
	if w.db == nil {
		return false, serror.FromStr("database handle is not set")
	}

	classifierStatus, err := status.GetClassifierStatus(ctx, w.db)
	if err != nil || classifierStatus.LastClassifiedBlockID == nil {
		return false, err
	}

	return w.state.ID <= *classifierStatus.LastClassifiedBlockID, nil
}

// PostExecution does nothing
func (w *Wrapper) PostExecution(context.Context) error {
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
// destination transactions of the current block into the in-memory graph
func (w *Wrapper) Iterate(ctx context.Context) (bool, error) {
	connectedNodes, singleNodes, err := analytics.GetPrivacyTransactionsByBlock(ctx, w.db, w.state.ID)
	if err != nil {
		return false, err
	}

	if len(connectedNodes) == 0 && len(singleNodes) == 0 {
		// nothing to do
		return true, nil
	}

	if len(connectedNodes) == 0 || len(singleNodes) == 0 {
		// something is wrong
		return false, serror.FromStr("count of single or connected nodes is zero")
	}

	w.transactionGraphMutex.Lock()
	defer w.transactionGraphMutex.Unlock()

	if graphErr := upsertSingleNodes(w.transactionGraph, singleNodes); graphErr != nil {
		return false, graphErr
	}

	if graphErr := addEdges(w.transactionGraph, connectedNodes); graphErr != nil {
		return false, graphErr
	}

	w.blocks.Inc()
	w.transactions.Add(float64(len(connectedNodes) + len(singleNodes)))
	w.blockHeight.Set(float64(w.state.ID))

	return true, nil
}
