// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package processor

import (
	"backend/blockiterator"
	dbstat "backend/db/status"
	"backend/external"
	"backend/jsonrpc"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// Crawler implements BlockIterator which processes the transactions of each traversed block
type Crawler struct {
	config       Config
	db           external.Database
	rpc          external.RPCClient
	ctx          context.Context
	state        crawlerState
	blocks       prometheus.Counter
	transactions prometheus.Counter
	blockHeight  prometheus.Gauge

	currentBlock *jsonrpc.GetBlockVerboseResult

	initialBlockCacheSize int64
	cache                 *outputCache
}

// NewCrawler creates a new Crawler object
func NewCrawler(ctx context.Context, database external.Database,
	rpc external.RPCClient, initialBlockCacheSize int64, cfg Config) *Crawler {
	return &Crawler{
		config:                cfg,
		db:                    database,
		rpc:                   rpc,
		ctx:                   ctx,
		initialBlockCacheSize: initialBlockCacheSize,
	}
}

func (c *Crawler) RegisterMetrics(req prometheus.Registerer) {
	c.blocks = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_crawler_blocks_processed_total",
		Help: "The total number of blocks processed by the crawler",
	})
	req.MustRegister(c.blocks)
	c.transactions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_crawler_transactions_processed_total",
		Help: "The total number of transactions processed by the crawler",
	})
	req.MustRegister(c.transactions)
	c.blockHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dakar_crawler_last_block",
		Help: "The last processed block by the crawler",
	})
	req.MustRegister(c.blockHeight)
}

func (c *Crawler) Props() blockiterator.Properties {
	return blockiterator.Properties{
		Name:                        "processor",
		Context:                     c.ctx,
		CurrentBlock:                c.state.id,
		ProcessedBlockCount:         1,
		SupportsMultiBlockIteration: false,
	}
}

// SetMaxBlocks is not supported for crawler, so do nothing
func (c *Crawler) SetMaxBlocks(int64) {}

// IncrementState increments the state one block
func (c *Crawler) IncrementState() error {
	if c.currentBlock == nil {
		return serror.FromStr("currentBlock is nil")
	}

	return c.state.increment(c.currentBlock.NextHash)
}

// Empty returns true if the BlockIterator has no more data to iterate on.
// This happens if State.ID is higher than State.Top
func (c *Crawler) Empty() bool {
	return c.currentBlock == nil ||
		(c.currentBlock != nil && c.currentBlock.NextHash == "") ||
		c.state.top < c.state.id+c.config.ForkRangeLimit
}

// CalculateInitialState calculates the state on which the iterator starts processing
func (c *Crawler) CalculateInitialState(ctx context.Context) error {
	if err := dbstat.SetCrawling(ctx, c.db, true); err != nil {
		return err
	}

	state, err := getInitialState(ctx, c.db, c.rpc)
	if err != nil {
		return err
	}

	c.state = state

	c.blockHeight.Set(float64(state.id))
	c.state.incremented = true
	info(fmt.Sprintf("Loading UTXOs of last %d blocks ...", c.initialBlockCacheSize))
	c.cache, err = newUTXOCache(ctx, c.db, state.id, c.initialBlockCacheSize)
	if err != nil {
		return err
	}

	info(fmt.Sprintf("Loaded %d UTXOs", c.cache.getOutputCounts()))

	return nil
}

// PostExecution sets the crawler status activity flag to false
func (c *Crawler) PostExecution(ctx context.Context) error {
	return dbstat.SetCrawling(ctx, c.db, false)
}

// Next tries to increase the internal state to the next block
func (c *Crawler) Next(context.Context) (bool, error) {
	if !c.state.incremented {
		// state is on next block
		block, err := c.rpc.GetBlockVerbose(c.state.hash)
		if err != nil {
			return false, err
		}

		if block.NextHash == "" {
			return false, nil
		}

		if incErr := c.state.increment(block.NextHash); incErr != nil {
			return false, incErr
		}
	}

	numBlocks, err := getRPCNumberOfBlocks(c.rpc)
	if err != nil {
		return false, err
	}

	if c.state.id <= numBlocks-c.config.ForkRangeLimit {
		currentBlock, getErr := c.rpc.GetBlockVerbose(c.state.hash)
		if getErr != nil {
			return false, getErr
		}
		c.currentBlock = currentBlock
		c.state.top = numBlocks
		return true, nil
	}

	return false, nil
}

// Iterate processes the current block. The block, its transactions,
// its outputs/inputs and all associated addresses are written to the database.
func (c *Crawler) Iterate(ctx context.Context) (bool, error) {
	if c.Empty() {
		return false, serror.FromStr("got empty state")
	}

	// get block from RPC-Client
	blk, err := c.rpc.GetBlockVerbose(c.state.hash)
	if err != nil {
		return false, err
	}
	c.currentBlock = blk

	txMap, err := createTransactionMap(c.rpc, blk.Tx)
	if err != nil {
		return false, err
	}

	// do the actual processing and aggregate the resulting metrics
	rBlockCounter, rTransactionCounter, err := processRound(ctx, c.db, c.state, c.currentBlock, txMap, c.config, c.cache)
	if err != nil {
		return false, err
	}

	c.blocks.Add(float64(rBlockCounter))
	c.transactions.Add(float64(rTransactionCounter))
	c.blockHeight.Set(float64(c.state.id))

	return true, nil
}

// createTransactionMap gets all requested transactions from the RPCClient
// and organizes them in a map indexed by the transaction hash
func createTransactionMap(client external.RPCClient, transactions []string) (map[string]jsonrpc.TxRawResult, error) {
	rawTransactions, err := client.GetRawTransactionVerboseBatch(transactions)
	if err != nil {
		return nil, err
	}
	txs := make(map[string]jsonrpc.TxRawResult, len(rawTransactions))
	for _, rawTransaction := range rawTransactions {
		if rawTransaction == nil {
			return nil, serror.FromFormat("raw transaction is nil. Request: %v", transactions)
		}

		txs[rawTransaction.Txid] = *rawTransaction
	}

	return txs, nil
}
