package processor

import (
	"backend/blockiterator"
	"backend/cmd/cliutil"
	dbstat "backend/db/status"
	"backend/external"
	"backend/jsonrpc"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
		blocks: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_crawler_blocks_processed_total",
			Help: "The total number of blocks processed by the crawler",
		}),
		transactions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_crawler_transactions_processed_total",
			Help: "The total number of transactions processed by the crawler",
		}),
		blockHeight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "dakar_crawler_last_block",
			Help: "The last processed block by the crawler",
		}),
	}
}

func (c *Crawler) Props() blockiterator.Properties {
	return blockiterator.Properties{
		Name:                "processor",
		Context:             c.ctx,
		Logger:              thisLogger,
		CurrentBlock:        c.state.id,
		ProcessedBlockCount: 1,
	}
}

// IncrementState increments the state one block
func (c *Crawler) IncrementState() error {
	if c.currentBlock == nil {
		return cliutil.NewStackErrorStr("currentBlock is nil")
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
func (c *Crawler) CalculateInitialState() error {
	if err := dbstat.SetCrawling(c.db, true); err != nil {
		return err
	}

	state, err := getInitialState(c.db, c.rpc)
	if err != nil {
		return err
	}

	c.state = state

	c.blockHeight.Set(float64(state.id))
	c.state.incremented = true
	info(fmt.Sprintf("Loading UTXOs of last %d blocks ...", c.initialBlockCacheSize))
	c.cache, err = newUTXOCache(c.db, int64(state.id), c.initialBlockCacheSize)
	if err != nil {
		return err
	}

	info(fmt.Sprintf("Loaded %d UTXOs", c.cache.getOutputCounts()))

	return nil
}

// PostExecution sets the crawler status activity flag to false
func (c *Crawler) PostExecution() error {
	return dbstat.SetCrawling(c.db, false)
}

// NextBlock tries to increase the internal state to the next block
func (c *Crawler) NextBlock() (bool, error) {
	if !c.state.incremented {
		// state is on next block
		block, err := c.rpc.GetBlockVerbose(c.state.hash)
		if err != nil {
			return false, cliutil.NewStackError(err)
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
			return false, cliutil.NewStackError(getErr)
		}
		c.currentBlock = currentBlock
		c.state.top = numBlocks
		return true, nil
	}

	return false, nil
}

// Iterate processes the current block. The block, its transactions,
// its outputs/inputs and all associated addresses are written to the database.
func (c *Crawler) Iterate() (bool, error) {
	if c.Empty() {
		return false, cliutil.NewStackErrorStr("got empty state")
	}

	var err error
	// get block from RPC-Client
	c.currentBlock, err = c.rpc.GetBlockVerbose(c.state.hash)
	if err != nil {
		return false, cliutil.NewStackError(err)
	}

	// do the actual processing and aggregate the resulting metrics
	if rBlockCounter, rTransactionCounter, processErr := processRound(c.db, c.rpc, c.state, c.currentBlock,
		c.config, c.cache); processErr == nil {
		c.blocks.Add(float64(rBlockCounter))
		c.transactions.Add(float64(rTransactionCounter))
		c.blockHeight.Set(float64(c.state.id))
	} else {
		return false, processErr
	}

	return true, nil
}
