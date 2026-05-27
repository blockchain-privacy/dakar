// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package classifier

import (
	"context"
	"gitlab.com/blockchain-privacy/dakar/analytics/classifier/btc"
	"gitlab.com/blockchain-privacy/dakar/analytics/classifier/dash"
	"gitlab.com/blockchain-privacy/dakar/blockiterator"
	dbstat "gitlab.com/blockchain-privacy/dakar/db/status"
	"gitlab.com/blockchain-privacy/dakar/external"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

type iteratorFunction func(ctx context.Context, c external.Database, from int64, to int64) (bool, error)

type Config struct {
	iterFunc       iteratorFunction
	minBlockHeight int64
}

func NewDashConfig() Config {
	return Config{iterFunc: dash.Iterate, minBlockHeight: 0}
}

func NewBTCConfig() Config {
	return Config{
		iterFunc: btc.Iterate,
		// 740000 -> Jun 9, 2022 shortly before Wasabi 2.0 was released on Jun 15, 2022
		// 520000 -> Apr 26, 2018 shortly before Whirlpool repo was created on Apr 28, 2018
		minBlockHeight: 520000,
	}
}

// Classifier implements BlockIterator which classifies the transactions of each traversed block
type Classifier struct {
	config Config
	db     external.Database
	ctx    context.Context
	state  blockiterator.State

	// how many blocks are processed in one iteration at maximum
	maxBlocks int64
	// number of blocks which have been processed by the last Iterate call
	blocksProcessed int64

	blocks      prometheus.Counter
	blockHeight prometheus.Gauge
}

// NewClassifier creates a new Classifier object
func NewClassifier(ctx context.Context, dgraph external.Database, cfg Config) *Classifier {
	return &Classifier{
		config:    cfg,
		db:        dgraph,
		ctx:       ctx,
		maxBlocks: 1,
	}
}

func (c *Classifier) RegisterMetrics(req prometheus.Registerer) {
	c.blocks = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_classifier_blocks_processed_total",
		Help: "The total number of blocks processed by the classifier",
	})
	req.MustRegister(c.blocks)
	c.blockHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dakar_classifier_last_block",
		Help: "The last processed block by the classifier",
	})
	req.MustRegister(c.blockHeight)
}

func (c *Classifier) Props() blockiterator.Properties {
	return blockiterator.Properties{
		Name:                        "classifier",
		Context:                     c.ctx,
		CurrentBlock:                c.state.ID,
		ProcessedBlockCount:         c.blocksProcessed,
		SupportsMultiBlockIteration: true,
	}
}

func (c *Classifier) SetMaxBlocks(maxBlocks int64) {
	c.maxBlocks = maxBlocks
}

// IncrementState increments the state one block
func (c *Classifier) IncrementState() error {
	c.state.ID += c.blocksProcessed
	return nil
}

// Empty checks if there are more blocks above the current one
func (c *Classifier) Empty() bool {
	return c.state.ID > c.state.Top
}

// CalculateInitialState calculates the state on which the iterator starts processing
func (c *Classifier) CalculateInitialState(ctx context.Context) error {
	if err := dbstat.SetClassifying(ctx, c.db, true); err != nil {
		return err
	}

	if err := setInitialClassifierID(ctx, c.db, c.config.minBlockHeight); err != nil {
		return err
	}

	crawlerStatus, err := dbstat.GetCrawlerStatus(ctx, c.db)
	if err != nil {
		return err
	}

	classifierStatus, err := dbstat.GetClassifierStatus(ctx, c.db)
	if err != nil {
		return err
	}

	if classifierStatus.LastClassifiedBlockID == nil {
		return serror.FromStr("last classified block is not set")
	}

	var state blockiterator.State

	state.ID = *classifierStatus.LastClassifiedBlockID + 1

	if crawlerStatus.LastBlockID == nil {
		// nothing crawled yet, so set Top to a lower number as ID
		state.Top = *classifierStatus.LastClassifiedBlockID
	} else {
		// this is the usual case: Set Top to the current last crawled block height
		state.Top = *crawlerStatus.LastBlockID
	}

	c.state = state

	// state.ID - 1 because the ID is the next block
	c.blockHeight.Set(float64(c.state.ID - 1))

	return nil
}

// Next tries to increase the internal state to the next block.
// Returns true if the top block id was changed
func (c *Classifier) Next(ctx context.Context) (bool, error) {
	status, err := dbstat.GetCrawlerStatus(ctx, c.db)
	if err != nil {
		return false, err
	}

	if status.LastBlockID != nil && c.state.ID <= *status.LastBlockID {
		c.state.Top = *status.LastBlockID
		return true, nil
	}

	return false, nil
}

// Iterate classifies all transactions of the current block based
// on their own properties (number of outputs/inputs, amounts, fee, etc...)
// and how they are connected to other transactions.
func (c *Classifier) Iterate(ctx context.Context) (bool, error) {
	if c.config.iterFunc == nil {
		return false, serror.FromStr("iterator function is nil")
	}

	if c.maxBlocks == 0 {
		return false, serror.FromStr("max blocks must be higher than zero")
	}

	if c.Empty() {
		return false, serror.FromStr("got empty state")
	}

	// state.ID is a new block already, therefore maxBlocks has to be reduced by 1
	toBlockID := min(c.state.Top, c.state.ID+c.maxBlocks-1)

	if ok, err := c.config.iterFunc(ctx, c.db, c.state.ID, toBlockID); err != nil || !ok {
		return ok, err
	}

	// set the last classified block
	if statusErr := dbstat.SetLastClassifiedBlockID(ctx, c.db, toBlockID); statusErr != nil {
		return false, statusErr
	}

	c.blocksProcessed = toBlockID - c.state.ID + 1
	c.blocks.Add(float64(c.blocksProcessed))
	c.blockHeight.Set(float64(toBlockID))

	return true, nil
}

// PostExecution sets the classifier status activity flag to false
func (c *Classifier) PostExecution(ctx context.Context) error {
	return dbstat.SetClassifying(ctx, c.db, false)
}

// setInitialClassifierID sets the starting classifier block id to the
// value of startBlockClassifier if no value has been set yet
func setInitialClassifierID(ctx context.Context, dgraph external.Database, minBlockHeight int64) (err error) {
	status, err := dbstat.GetClassifierStatus(ctx, dgraph)
	if err != nil {
		return
	}

	if status.LastClassifiedBlockID == nil || *status.LastClassifiedBlockID < minBlockHeight {
		if err = dbstat.SetLastClassifiedBlockID(ctx, dgraph, minBlockHeight); err != nil {
			return
		}
	}
	return
}
