package analytics

import (
	"backend/blockIterator"
	"backend/cmd/cliutil"
	dbstat "backend/db/status"
	dbtx "backend/db/transaction"
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
	if !a.config.IsClassifyingEnabled {
		return errors.New("classifying is disabled per configuration")
	}

	if err := dbstat.SetClassifying(a.db, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if err := setInitialClassifierId(a.db, a.config.ClassifierStartBlock); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	crawlerStatus, err := dbstat.GetCrawlerStatus(a.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	classifierStatus, err := dbstat.GetClassifierStatus(a.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if classifierStatus.LastClassifiedBlockId == nil {
		return errors.New("error last classified block is not set")
	}

	var state blockIterator.State

	state.Id = *classifierStatus.LastClassifiedBlockId + 1

	if crawlerStatus.LastBlockId == nil {
		// nothing crawled yet, so set Top to a lower number as Id
		state.Top = *classifierStatus.LastClassifiedBlockId
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

	transactions, err := dbtx.GetTransactionByBlock(a.db, a.state.Id)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	privacyTransactions, err := processPrivacyType(a.ctx, a.db, transactions)
	if err != nil {
		if errors.Is(err, errorInterrupted) {
			return false, nil
		}

		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if len(privacyTransactions) > 0 {
		info(len(privacyTransactions), "privacy transactions, block id:", a.state.Id)
		// update the block in the database
		// after that function call the privacy type of all transactions is set

		// todo refactor into something like ''UpdatePrivacyType''
		//if err := dbblk.UpdateBlock(a.db, updatedBlock); err != nil {
		//	return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		//}
	}

	if err := dbstat.SetLastClassifiedBlockId(a.db, a.state.Id); err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return true, nil
}

func (a *Classifier) PostExecution() error {
	if err := dbstat.SetClassifying(a.db, false); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// setInitialClassifierId sets the starting classifier block id to the
// value of startBlockClassifier if no value has been set yet
func setInitialClassifierId(dgraph *dgo.Dgraph, startBlockClassifier uint64) (err error) {
	status, err := dbstat.GetClassifierStatus(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if status.LastClassifiedBlockId == nil ||
		*status.LastClassifiedBlockId < startBlockClassifier {
		if err = dbstat.SetLastClassifiedBlockId(dgraph, startBlockClassifier); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}
	return
}
