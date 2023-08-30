package clustering

import (
	"backend/blockiterator"
	"backend/cmd/cliutil"
	"backend/db/analytics/clustering"
	dbstat "backend/db/status"
	"backend/external"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"strconv"
)

// HierarchicalMultiInput implements BlockIterator which creates clusters via the multi-input heuristic
type HierarchicalMultiInput struct {
	db    external.Database
	ctx   context.Context
	state blockiterator.State

	blocks         prometheus.Counter
	transactions   prometheus.Counter
	mergedClusters prometheus.Counter
	newAddresses   prometheus.Counter
	blockHeight    prometheus.Gauge
}

// NewHierarchicalMultiInput creates a new hierarchical multi-input clustering object
func NewHierarchicalMultiInput(ctx context.Context, dgraph external.Database) *HierarchicalMultiInput {
	return &HierarchicalMultiInput{
		db:  dgraph,
		ctx: ctx,
		blocks: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_hmi_blocks_processed_total",
			Help: "The total number of blocks processed by the HMI clustering process",
		}),
		transactions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_hmi_transactions_processed_total",
			Help: "The total number of transactions processed by the HMI clustering process",
		}),
		mergedClusters: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_hmi_clusters_merged_total",
			Help: "The total number of clusters merged by the HMI clustering process",
		}),
		newAddresses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dakar_clustering_hmi_new_addresses_total",
			Help: "The total number of new addresses added to clusters by the HMI clustering process",
		}),
		blockHeight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "dakar_clustering_hmi_last_block",
			Help: "The last processed block by the HMI clustering process",
		}),
	}
}

// CalculateInitialState calculates the state on which the iterator starts processing
func (m *HierarchicalMultiInput) CalculateInitialState() error {
	if err := dbstat.SetClusteringHMI(m.db, true); err != nil {
		return err
	}

	if err := setInitialHMIClusteringID(m.db); err != nil {
		return err
	}

	classifierStatus, err := dbstat.GetClassifierStatus(m.db)
	if err != nil {
		return err
	}

	clusteringStatus, err := dbstat.GetClusteringHMIStatus(m.db)
	if err != nil {
		return err
	}

	if clusteringStatus.LastClusteredBlockID == nil {
		return cliutil.NewStackErrorStr("error last HMI clustered block is not set")
	}

	var state blockiterator.State

	state.ID = *clusteringStatus.LastClusteredBlockID + 1

	if classifierStatus.LastClassifiedBlockID == nil {
		// nothing classified yet, so set Top to a lower number as ID
		state.Top = *clusteringStatus.LastClusteredBlockID
	} else {
		// this is the usual case: Set Top to the current last classified block height
		state.Top = *classifierStatus.LastClassifiedBlockID
	}

	m.state = state

	// state.ID - 1 because the ID is the next block
	m.blockHeight.Set(float64(m.state.ID - 1))

	return nil
}

// Iterate clusters all addresses of the current block based on the multi-input heuristic
func (m *HierarchicalMultiInput) Iterate() (bool, error) {
	if m.Empty() {
		return false, cliutil.NewStackErrorStr("got empty state")
	}

	// get the transaction of the current block height
	transactions, err := clustering.GetInputAddressesByBlock(m.db, m.state.ID, clustering.TypeHMI)
	if err != nil {
		return false, err
	}

	var countMergedClusters int
	var countNewAddresses int

	// addressToClusterRoot maps an address UID to newly created cluster_id ("_:<cluster-id>") or a root cluster.
	// This is needed to create cluster relations between new clusters in the same block.
	addressToClusterRoot := make(map[string]string)

	childClusterToClusterRoot := make(map[string]string)

	if len(transactions) > 0 {
		var clusterIndex int
		var newClusters []clustering.Cluster
		clusterMap := make(map[string]clustering.Cluster)
		for _, tx := range transactions {
			// at least two addresses are needed to cluster
			if len(tx.Addresses) < 2 {
				continue
			}

			addressesWithoutCluster := make(map[string]bool)
			existingClusters := make(map[string]bool)

			for _, addr := range tx.Addresses {
				if len(addr.Clusters) > 0 {
					if len(addr.Clusters) != 1 {
						return false, cliutil.NewStackErrorf("found more than one multi-input cluster attached to address %v", addr)
					}

					transactionCluster := addr.Clusters[0]

					clusterMap[transactionCluster.UID] = clustering.Cluster{
						UID:          transactionCluster.UID,
						AddressCount: &transactionCluster.AddressCount,
					}

					if transactionCluster.Parents == nil {
						if localRoot := getClusterRootByCluster(childClusterToClusterRoot, transactionCluster.UID); localRoot != "" {
							// this happens if db-root cluster was found for which a new (local) cluster exists
							childClusterToClusterRoot[transactionCluster.UID] = localRoot
							existingClusters[localRoot] = true
						} else {
							existingClusters[transactionCluster.UID] = true
						}
					} else if r := getClusterRootByCluster(childClusterToClusterRoot, transactionCluster.UID); r != "" {
						// this is the case if for the cluster a known root cluster exists
						existingClusters[r] = true
					} else {
						root, dbErr := clustering.GetHierarchicalClusterRoot(m.db, transactionCluster.UID)
						if dbErr != nil {
							return false, fmt.Errorf("block %d cluster uid %s: %w",
								m.state.ID, transactionCluster.UID, dbErr)
						}

						clusterMap[root.UID] = clustering.Cluster{
							UID:          root.UID,
							AddressCount: &root.AddressCount,
						}

						if localRoot := getClusterRootByCluster(childClusterToClusterRoot, root.UID); localRoot != "" {
							// this happens if db-root cluster was found for which a new (local) cluster exists
							childClusterToClusterRoot[transactionCluster.UID] = localRoot
							existingClusters[localRoot] = true
						} else {
							childClusterToClusterRoot[transactionCluster.UID] = root.UID
							existingClusters[root.UID] = true
						}
					}
				} else if r := getClusterRootByCluster(addressToClusterRoot, addr.UID); r != "" {
					// this is the case if the address has no cluster attached
					// (db-state) but a local (not upserted) cluster was created
					existingClusters[r] = true
				} else {
					addressesWithoutCluster[addr.UID] = true
				}
			}

			if len(addressesWithoutCluster) == 0 && len(existingClusters) == 0 {
				// this should never happen
				return false, cliutil.NewStackErrorStr("Transaction " + tx.UID +
					" at block " + strconv.FormatUint(m.state.ID, 10) + " has invalid data")
			}

			if (len(existingClusters) == 0 && len(addressesWithoutCluster) < 2) ||
				(len(existingClusters) == 1 && len(addressesWithoutCluster) == 0) {
				// if transaction has zero clusters and less than two 2 addresses -> continue
				// if transaction has only one cluster and no new addresses -> continue
				continue
			}

			// create new cluster
			clusterIndex++
			cluster := clustering.NewHMICluster(clusterIndex, tx.UID)

			var addressCount int

			// calculate metrics
			countNewAddresses += len(addressesWithoutCluster)
			countMergedClusters += len(existingClusters)

			// add addresses
			addressCount += len(addressesWithoutCluster)
			for address := range addressesWithoutCluster {
				cluster.Addresses = append(cluster.Addresses, clustering.HollowAddress{UID: address})
			}

			// set the new cluster root for all addresses in the transaction
			for _, addr := range tx.Addresses {
				addressToClusterRoot[addr.UID] = cluster.UID
			}

			// add child clusters
			for c := range existingClusters {
				cluster.Children = append(cluster.Children, clustering.SubCluster{UID: c})
				// accumulate address counts from existing clusters
				if existingCluster, ok := clusterMap[c]; ok {
					addressCount += *existingCluster.AddressCount
				}

				childClusterToClusterRoot[c] = cluster.UID
				// the local cluster to cluster mapping has to be added to the address map so cluster connections can be followed
				addressToClusterRoot[c] = cluster.UID
			}

			for _, addr := range tx.Addresses {
				if addr.Clusters != nil {
					childClusterToClusterRoot[addr.Clusters[0].UID] = cluster.UID
				}
			}

			cluster.AddressCount = &addressCount

			newClusters = append(newClusters, cluster)
			clusterMap[cluster.UID] = cluster
		}

		// insert new clusters
		if len(newClusters) > 0 {
			if validationErr := validateClusters(newClusters); validationErr != nil {
				return false, fmt.Errorf("block id %d: %w", m.state.ID, validationErr)
			}

			clusterErr := clustering.AddClusters(m.db, newClusters, true)
			if clusterErr != nil {
				return false, clusterErr
			}
		}

		// update metrics
		m.mergedClusters.Add(float64(countMergedClusters))
		m.newAddresses.Add(float64(countNewAddresses))
		m.transactions.Add(float64(len(transactions)))
	}

	// set the last classified block
	if statusErr := dbstat.SetLastClusteredHMIBlockID(m.db, m.state.ID); statusErr != nil {
		return false, statusErr
	}

	m.blocks.Inc()
	m.blockHeight.Set(float64(m.state.ID))

	return true, nil
}

// NextBlock tries to increase the internal state to the next block
func (m *HierarchicalMultiInput) NextBlock() (bool, error) {
	status, err := dbstat.GetClassifierStatus(m.db)
	if err != nil {
		return false, err
	} else if status.LastClassifiedBlockID == nil {
		return false, cliutil.NewStackErrorStr("last classified block is not set")
	}

	if m.state.ID <= *status.LastClassifiedBlockID {
		m.state.Top = *status.LastClassifiedBlockID
		return true, nil
	}

	return false, nil
}

func (m *HierarchicalMultiInput) PostExecution() error {
	return dbstat.SetClusteringHMI(m.db, false)
}

func (m *HierarchicalMultiInput) IncrementState() error {
	m.state.ID++
	return nil
}

// Empty checks if there are more blocks above the current one
func (m *HierarchicalMultiInput) Empty() bool {
	return m.state.ID > m.state.Top
}

// CurrentBlock returns the height of the block which is getting clustered
func (m *HierarchicalMultiInput) CurrentBlock() uint64 {
	return m.state.ID
}

// Logger returns the Logger
func (m *HierarchicalMultiInput) Logger() *log.Logger {
	return clusteringLogger
}

// Context returns the context
func (m *HierarchicalMultiInput) Context() context.Context {
	return m.ctx
}

// Name returns the name
func (m *HierarchicalMultiInput) Name() string {
	return "Hierarchical Multi-Input Clustering"
}

// setInitialHMIClusteringID sets the starting HMI clustering block id to 0 if no value has been set yet
func setInitialHMIClusteringID(dgraph external.Database) error {
	status, err := dbstat.GetClusteringHMIStatus(dgraph)
	if err != nil {
		return err
	}

	if status.LastClusteredBlockID == nil {
		if err = dbstat.SetLastClusteredHMIBlockID(dgraph, 0); err != nil {
			return err
		}
	}
	return nil
}

// getClusterRootByCluster returns the rootUID of UID. This is done by following the
// relations given in clusterMapping. If an empty string is returned no rootUID exists.
func getClusterRootByCluster(clusterMapping map[string]string, uid string) string {
	if len(clusterMapping) == 0 {
		return ""
	}

	var rootUID string
	tmpUID := uid
	hops := 0

	for tmpUID != "" {
		var ok bool
		if tmpUID, ok = clusterMapping[tmpUID]; ok {
			rootUID = tmpUID
			hops++
		}
	}

	// add relation if multi hop was performed to get better performance in subsequent queries
	if hops > 1 {
		clusterMapping[uid] = rootUID
	}

	return rootUID
}

// validateClusters checks if the given clusters share clusters or addresses
func validateClusters(clusters []clustering.Cluster) error {
	clusterUIDs := make(map[string]bool)
	addressUIDs := make(map[string]bool)
	for _, cluster := range clusters {
		if len(cluster.Children) == 0 && len(cluster.Addresses) == 0 {
			return cliutil.NewStackErrorf("cluster %s has no addresses and no children", cluster.UID)
		}

		for _, child := range cluster.Children {
			if clusterUIDs[child.UID] {
				return cliutil.NewStackErrorf("cluster %s has multiple parents", child.UID)
			}
			clusterUIDs[child.UID] = true
		}

		for _, addr := range cluster.Addresses {
			if addressUIDs[addr.UID] {
				return cliutil.NewStackErrorf("address %s has multiple parents", addr.UID)
			}
			addressUIDs[addr.UID] = true
		}
	}

	return nil
}
