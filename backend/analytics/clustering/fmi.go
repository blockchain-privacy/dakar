package clustering

import (
	"backend/blockiterator"
	"backend/cmd/cliutil"
	"backend/db/analytics/clustering"
	dbstat "backend/db/status"
	"backend/external"
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
)

// FlatMultiInput implements BlockIterator which creates clusters via the multi-input heuristic
type FlatMultiInput struct {
	db    external.Database
	ctx   context.Context
	state blockiterator.State

	blocks         prometheus.Counter
	transactions   prometheus.Counter
	mergedClusters prometheus.Counter
	newAddresses   prometheus.Counter
	blockHeight    prometheus.Gauge
}

// NewFlatMultiInput creates a new Classifier object
func NewFlatMultiInput(ctx context.Context, dgraph external.Database) *FlatMultiInput {
	return &FlatMultiInput{
		db:  dgraph,
		ctx: ctx,
		blocks: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_fmi_blocks_processed_total",
			Help: "The total number of blocks processed by the FMI clustering process",
		}),
		transactions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_fmi_transactions_processed_total",
			Help: "The total number of transactions processed by the FMI clustering process",
		}),
		mergedClusters: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_fmi_clusters_merged_total",
			Help: "The total number of clusters merged by the FMI clustering process",
		}),
		newAddresses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_fmi_new_addresses_total",
			Help: "The total number of new addresses added to clusters by the FMI clustering process",
		}),
		blockHeight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "dakar_clustering_fmi_last_block",
			Help: "The last processed block by the FMI clustering process",
		}),
	}
}

// CalculateInitialState calculates the state on which the iterator starts processing
func (m *FlatMultiInput) CalculateInitialState() error {
	if err := dbstat.SetClusteringFMI(m.db, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if err := setInitialFMIClusteringID(m.db); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	classifierStatus, err := dbstat.GetClassifierStatus(m.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	clusteringStatus, err := dbstat.GetClusteringFMIStatus(m.db)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if clusteringStatus.LastClusteredBlockID == nil {
		return errors.New("error last FMI clustered block is not set")
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

// Iterate clusters all addresses of the current block based on the multi-input heuristic
func (m *FlatMultiInput) Iterate() (bool, error) {
	if m.Empty() {
		return false, errors.New("got empty state")
	}

	// get the transaction of the current block height
	transactions, err := clustering.GetInputAddressesByBlock(m.db, m.state.ID, clustering.TypeFMI)
	if err != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	//var countMergedClusters int
	//var countNewAddresses int
	//
	//var clusterIndex int

	if len(transactions) > 0 {
		//	var newClusters []clustering.Cluster
		//	clusterMap := make(map[string]clustering.Cluster)
		//	for _, tx := range transactions {
		//		// at least two addresses are needed to cluster
		//		if len(tx.Addresses) < 2 {
		//			continue
		//		}
		//
		//	}
		//	// update metrics
		//	m.mergedClusters.Add(float64(countMergedClusters))
		//	m.newAddresses.Add(float64(countNewAddresses))
		//	m.transactions.Add(float64(len(transactions)))
	}

	// set the last classified block
	if statusErr := dbstat.SetLastClusteredFMIBlockID(m.db, m.state.ID); statusErr != nil {
		return false, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), statusErr)
	}

	m.blocks.Inc()
	m.blockHeight.Set(float64(m.state.ID))

	return true, nil
}

// NextBlock tries to increase the internal state to the next block
func (m *FlatMultiInput) NextBlock() (bool, error) {
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

func (m *FlatMultiInput) PostExecution() error {
	if err := dbstat.SetClusteringFMI(m.db, false); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

func (m *FlatMultiInput) IncrementState() error {
	m.state.ID++
	return nil
}

// Empty checks if there are more blocks above the current one
func (m *FlatMultiInput) Empty() bool {
	return m.state.ID > m.state.Top
}

// CurrentBlock returns the height of the block which is getting clustered
func (m *FlatMultiInput) CurrentBlock() uint64 {
	return m.state.ID
}

// Logger returns the Logger
func (m *FlatMultiInput) Logger() *log.Logger {
	return clusteringLogger
}

// Context returns the context
func (m *FlatMultiInput) Context() context.Context {
	return m.ctx
}

// Name returns the name
func (m *FlatMultiInput) Name() string {
	return "Flat Multi-Input Clustering"
}

// setInitialFMIClusteringID sets the starting FMI clustering block id to 0 if no value has been set yet
func setInitialFMIClusteringID(dgraph external.Database) (err error) {
	status, err := dbstat.GetClusteringFMIStatus(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if status.LastClusteredBlockID == nil {
		if err = dbstat.SetLastClusteredFMIBlockID(dgraph, 0); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}
	return
}
