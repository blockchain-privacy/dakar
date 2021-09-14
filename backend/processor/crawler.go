package processor

import (
	"backend/cmd/cliutil"
	dbstat "backend/db/status"
	"backend/external"
	"context"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
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

	isDatabaseEmpty bool
	currentBlock    *btcjson.GetBlockVerboseResult

	cache *utxoCache
}

// NewCrawler creates a new Crawler object
func NewCrawler(ctx context.Context, database external.Database, rpc external.RPCClient, cfg Config) *Crawler {
	return &Crawler{
		config: cfg,
		db:     database,
		rpc:    rpc,
		ctx:    ctx,
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

// Name returns the name
func (c *Crawler) Name() string {
	return "processor"
}

// Logger returns the Logger
func (c *Crawler) Logger() *log.Logger {
	return thisLogger
}

// Context returns the context
func (c *Crawler) Context() context.Context {
	return c.ctx
}

// Db returns the database access
func (c *Crawler) Db() external.Database {
	return c.db
}

// IncrementState increments the state one block
func (c *Crawler) IncrementState() error {
	if err := c.state.increment(c.currentBlock.NextHash); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	return nil
}

// Empty returns true if the BlockIterator has no more data to iterate on.
// This happens if State.ID is higher than State.Top
func (c *Crawler) Empty() bool {
	return (c.currentBlock != nil && c.currentBlock.NextHash == "") || c.state.top < c.state.id+c.config.ForkRangeLimit
}

// CalculateInitialState calculates the state on which the iterator starts processing
func (c *Crawler) CalculateInitialState() error {
	if err := dbstat.SetCrawling(c.db, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	state, err := getInitialState(c.db, c.rpc)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	status, err := dbstat.GetCrawlerStatus(c.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	c.state = state

	if status.LastBlockID == nil {
		c.isDatabaseEmpty = true
	}

	c.blockHeight.Set(float64(state.id))
	c.state.incremented = true

	info("Loading UTXOs of last", initialLoadSize, "blocks ...")
	c.cache, err = newCache(c.db, int64(state.id))
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	info("Loaded", c.cache.getOutputCounts(), "UTXOs")

	return nil
}

// PostExecution sets the crawler status activity flag to false
func (c *Crawler) PostExecution() error {
	if err := dbstat.SetCrawling(c.db, false); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// CurrentBlock returns the height of the block which is currently crawled
func (c *Crawler) CurrentBlock() uint64 {
	return c.state.id
}

// NextBlock tries to increase the internal state to the next block
func (c *Crawler) NextBlock() (bool, error) {
	if !c.state.incremented {
		// state is on next block
		block, err := c.rpc.GetBlockVerbose(c.state.chainHash)
		if err != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if block.NextHash == "" {
			return false, nil
		}

		if incErr := c.state.increment(block.NextHash); incErr != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), incErr)
		}
	}

	numBlocks, err := getRPCNumberOfBlocks(c.rpc)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if c.state.id <= numBlocks-c.config.ForkRangeLimit {
		currentBlock, getErr := c.rpc.GetBlockVerbose(c.state.chainHash)
		if getErr != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), getErr)
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
		return false, errors.New("got empty state")
	}

	var err error
	// get block from RPC-Client
	c.currentBlock, err = c.rpc.GetBlockVerbose(c.state.chainHash)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// do the actual processing and aggregate the resulting metrics
	if rBlockCounter, rTransactionCounter, processErr := processRound(c.db, c.rpc, c.state, c.currentBlock,
		c.isDatabaseEmpty, c.config, c.cache); processErr == nil {
		c.isDatabaseEmpty = false

		c.blocks.Add(float64(rBlockCounter))
		c.transactions.Add(float64(rTransactionCounter))
		c.blockHeight.Set(float64(c.state.id))
	} else {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), processErr)
	}

	return true, nil
}
