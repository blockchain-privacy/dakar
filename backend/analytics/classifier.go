package analytics

import (
	"backend/blockIterator"
	"backend/cmd/cliutil"
	dbblk "backend/db/block"
	dbstat "backend/db/status"
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
)

type Classifier struct {
	config Config
	db     *dgo.Dgraph
	ctx    context.Context
	state  blockIterator.State
}

// NewClassifier creates a new Classifier object
func NewClassifier(ctx context.Context, dgraph *dgo.Dgraph, cfg Config) *Classifier {
	return &Classifier{
		config: cfg,
		db:     dgraph,
		ctx:    ctx,
	}
}

func (a *Classifier) Logger() *log.Logger {
	return analyticsLogger
}

func (a *Classifier) State() blockIterator.State {
	return a.state
}

func (a *Classifier) SetState(newState blockIterator.State) {
	a.state = newState
}

func (a *Classifier) Context() context.Context {
	return a.ctx
}

func (a *Classifier) Db() *dgo.Dgraph {
	return a.db
}

func (a *Classifier) IncrementState() {
	a.state.Id++
}

// Empty checks if there is work in the queue or more block above the current one
func (a *Classifier) Empty() bool {
	return a.state.Id > a.state.Top
}

func (a *Classifier) CalculateInitialState() error {
	if !a.config.IsAnalysingEnabled {
		return errors.New("analysing is disabled per configuration")
	}

	if err := dbstat.SetAnalyzing(a.db, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if err := setInitialAnalyserId(a.db, a.config.AnalyseStartBlock); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	crawlerStatus, err := dbstat.GetCrawlerStatus(a.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	analyzerStatus, err := dbstat.GetAnalyzerStatus(a.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if analyzerStatus.LastAnalysedBlockId == nil {
		return errors.New("error last analysed block is not set")
	}

	var state blockIterator.State

	state.Id = *analyzerStatus.LastAnalysedBlockId + 1

	if crawlerStatus.LastBlockId == nil {
		// nothing crawled yet, so set Top to a lower number as Id
		state.Top = *analyzerStatus.LastAnalysedBlockId
	} else if *crawlerStatus.LowestBlockId > state.Id {
		// happens the crawler is started with a high start block id in block range mode
		state.Id = *crawlerStatus.LowestBlockId
		state.Top = *crawlerStatus.LastBlockId
	} else {
		// this is the usual case: Set Top to the current last crawled block height
		state.Top = *crawlerStatus.LastBlockId
	}

	a.state = state

	return nil
}

func (a *Classifier) Iterate() (bool, error) {
	if a.Empty() {
		return false, errors.New("got empty state")
	}

	currentBlock, err := dbblk.GetBlockById(a.db, a.state.Id)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	updatedBlock, err := processPrivacyType(a.ctx, a.db, currentBlock)
	if err != nil {
		if errors.Is(err, errorInterrupted) {
			return false, nil
		}

		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if len(updatedBlock.Transactions) > 0 {
		// update the block in the database
		// after that function call the privacy type of all transactions is set
		if err := dbblk.UpdateBlock(a.db, updatedBlock); err != nil {
			return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
	}

	if err := dbstat.SetLastAnalysedBlockId(a.db, a.state.Id); err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return true, nil
}

func (a *Classifier) PostExecution() error {
	if err := dbstat.SetAnalyzing(a.db, false); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}
