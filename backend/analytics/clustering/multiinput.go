package clustering

import (
	"backend/blockiterator"
	"backend/cmd/cliutil"
	dbstat "backend/db/status"
	"backend/external"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MultiInput implements BlockIterator which creates cluster via the multi-input heuristic
type MultiInput struct {
	db           external.Database
	ctx          context.Context
	state        blockiterator.State
	blocks       prometheus.Counter
	transactions prometheus.Counter
	blockHeight  prometheus.Gauge
}

// NewMultiInput creates a new Classifier object
func NewMultiInput(ctx context.Context, dgraph external.Database) *MultiInput {
	return &MultiInput{
		db:  dgraph,
		ctx: ctx,
		blocks: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_multi_input_blocks_processed_total",
			Help: "The total number of blocks processed by the multi-input clustering process",
		}),
		transactions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_multi_input_transactions_processed_total",
			Help: "The total number of transactions processed by the multi-input clustering process",
		}),
		blockHeight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "dakar_clustering_multi_input_last_block",
			Help: "The last processed block by the multi-input clustering process",
		}),
	}
}

func (m *MultiInput) CalculateInitialState() error {
	if err := dbstat.SetClusteringMultiInput(m.db, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if err := setInitialClusteringID(m.db); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	classifierStatus, err := dbstat.GetClassifierStatus(m.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	clusteringStatus, err := dbstat.GetClusteringMultiInputStatus(m.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if clusteringStatus.LastClusteredBlockID == nil {
		return errors.New("error last multi-input clustered block is not set")
	}

	var state blockiterator.State

	state.ID = *clusteringStatus.LastClusteredBlockID + 1

	if classifierStatus.LastClassifiedBlockID == nil {
		// nothing classified yet, so set Top to a lower number as ID
		state.Top = *clusteringStatus.LastClusteredBlockID
	} else {
		// this is the usual case: Set Top to the current last classified block height
		state.Top = *clusteringStatus.LastClusteredBlockID
	}

	m.state = state

	// state.ID - 1 because the ID is the next block
	m.blockHeight.Set(float64(m.state.ID - 1))

	return nil
}

func (m *MultiInput) Iterate() (bool, error) {
	panic("implement me")
}

// NextBlock tries to increase the internal state to the next block
func (m *MultiInput) NextBlock() (bool, error) {
	status, err := dbstat.GetClassifierStatus(m.db)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	} else if status.LastClassifiedBlockID == nil {
		return false, errors.New("last classified block is not set")
	}

	if m.state.ID <= *status.LastClassifiedBlockID {
		m.state.Top = *status.LastClassifiedBlockID
		return true, nil
	}

	return false, nil
}

func (m *MultiInput) PostExecution() error {
	if err := dbstat.SetClusteringMultiInput(m.db, false); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

func (m *MultiInput) IncrementState() error {
	m.state.ID++
	return nil
}

// Empty checks if there are more blocks above the current one
func (m *MultiInput) Empty() bool {
	return m.state.ID > m.state.Top
}

// CurrentBlock returns the height of the block which is getting clustered
func (m *MultiInput) CurrentBlock() uint64 {
	return m.state.ID
}

// Logger returns the Logger
func (m *MultiInput) Logger() *log.Logger {
	return clusteringLogger
}

// Context returns the context
func (m *MultiInput) Context() context.Context {
	return m.ctx
}

// Name returns the name
func (m *MultiInput) Name() string {
	return "Multi-Input Clustering"
}

// setInitialClusteringID sets the starting clustering block id to 0 if no value has been set yet
func setInitialClusteringID(dgraph external.Database) (err error) {
	status, err := dbstat.GetClusteringMultiInputStatus(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if status.LastClusteredBlockID == nil {
		if err = dbstat.SetLastClusteredBlockID(dgraph, 0); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}
	return
}
