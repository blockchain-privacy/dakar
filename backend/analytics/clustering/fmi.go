// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package clustering

import (
	"backend/blockiterator"
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/clustering"
	dbstat "backend/db/status"
	"backend/external"
	"context"
	"slices"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// maximum number of addresses per cluster. Cluster with a size of maxClusterSize
// won't be added addresses. Also,transactions will be ignored if they would merge
// multiple clusters with an accumulated number of over maxClusterSize.
const maxClusterSize = 50_000

// FlatMultiInput implements BlockIterator which creates clusters via the multi-input heuristic
type FlatMultiInput struct {
	config Config
	db     external.Database
	ctx    context.Context
	state  blockiterator.State

	// how many blocks are processed in one iteration at maximum
	maxBlocks int64
	// number of blocks which have been processed by the last Iterate call
	blocksProcessed int64

	blocks         prometheus.Counter
	transactions   prometheus.Counter
	mergedClusters prometheus.Counter
	newAddresses   prometheus.Counter
	blockHeight    prometheus.Gauge
}

type Config struct {
	// if a transaction has more than the specified number of inputs,
	// it will be considered a CoinJoin transaction. If set to zero, this filter will be ignored.
	excludeInputCountThreshold int
	// if a transaction has more than the specified number of outputs,
	// it will be considered a CoinJoin transaction. If set to zero, this filter will be ignored.
	excludeOutputCountThreshold int
}

func NewDashConfig() Config {
	return Config{excludeInputCountThreshold: 0, excludeOutputCountThreshold: 0}
}

func NewBTCConfig() Config {
	return Config{excludeInputCountThreshold: 10, excludeOutputCountThreshold: 10}
}

// NewFlatMultiInput creates a new flat multi-input clustering object
func NewFlatMultiInput(ctx context.Context, dgraph external.Database, cfg Config) *FlatMultiInput {
	return &FlatMultiInput{
		config:    cfg,
		db:        dgraph,
		ctx:       ctx,
		maxBlocks: 1,
	}
}

func (m *FlatMultiInput) RegisterMetrics(req prometheus.Registerer) {
	m.blocks = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_clustering_fmi_blocks_processed_total",
		Help: "The total number of blocks processed by the FMI clustering process",
	})
	req.MustRegister(m.blocks)
	m.transactions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_clustering_fmi_transactions_processed_total",
		Help: "The total number of transactions processed by the FMI clustering process",
	})
	req.MustRegister(m.transactions)
	m.mergedClusters = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_clustering_fmi_clusters_merged_total",
		Help: "The total number of clusters merged by the FMI clustering process",
	})
	req.MustRegister(m.mergedClusters)
	m.newAddresses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dakar_clustering_fmi_new_addresses_total",
		Help: "The total number of new addresses added to clusters by the FMI clustering process",
	})
	req.MustRegister(m.newAddresses)
	m.blockHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dakar_clustering_fmi_last_block",
		Help: "The last processed block by the FMI clustering process",
	})
	req.MustRegister(m.blockHeight)
}

// CalculateInitialState calculates the state on which the iterator starts processing
func (m *FlatMultiInput) CalculateInitialState(ctx context.Context) error {
	if err := dbstat.SetClusteringFMI(ctx, m.db, true); err != nil {
		return err
	}

	if err := setInitialFMIClusteringID(ctx, m.db); err != nil {
		return err
	}

	classifierStatus, err := dbstat.GetClassifierStatus(ctx, m.db)
	if err != nil {
		return err
	}

	clusteringStatus, err := dbstat.GetClusteringFMIStatus(ctx, m.db)
	if err != nil {
		return err
	}

	if clusteringStatus.LastClusteredBlockID == nil {
		return serror.FromStr("last FMI clustered block is not set")
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

type newCluster struct {
	changeTransaction string
	mergeList         []clustering.Cluster
	addresses         map[string]bool
}

// processAsNonMultiInput treats each address on its own, meaning the multi-input/common spend heuristic is not applied.
func processAsNonMultiInput(clusterMergeMap map[string]*newCluster, addressMergeMap map[string]*newCluster,
	clusterStore map[string]clustering.Cluster, txUID string, addresses []clustering.AddressWithCluster) {
	for _, addr := range addresses {
		addressesWithoutCluster := make(map[string]bool)
		existingClusters := make(map[string]bool)
		if addr.Cluster != nil {
			transactionCluster := addr.Cluster

			existingClusters[transactionCluster.UID] = true

			clusterStore[transactionCluster.UID] = clustering.Cluster{
				UID:          transactionCluster.UID,
				AddressCount: &transactionCluster.AddressCount,
			}
		} else {
			addressesWithoutCluster[addr.UID] = true
		}

		// consider each output address on its own, as opposed to the common spend case on the input side
		addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore,
			txUID, existingClusters, addressesWithoutCluster)
	}
}

// processAsMultiInput assumes each address is created by the same entity, meaning the multi-input/common spend heuristic is applied.
func processAsMultiInput(clusterMergeMap map[string]*newCluster, addressMergeMap map[string]*newCluster,
	clusterStore map[string]clustering.Cluster, txUID string, addresses []clustering.AddressWithCluster) {
	addressesWithoutCluster := make(map[string]bool)
	existingClusters := make(map[string]bool)

	for _, addr := range addresses {
		if addr.Cluster != nil {
			transactionCluster := addr.Cluster

			// don't add addresses to clusters with a size of over 50 000 addresses
			if transactionCluster.AddressCount > maxClusterSize {
				continue
			}

			existingClusters[transactionCluster.UID] = true

			clusterStore[transactionCluster.UID] = clustering.Cluster{
				UID:          transactionCluster.UID,
				AddressCount: &transactionCluster.AddressCount,
			}
		} else {
			addressesWithoutCluster[addr.UID] = true
		}
	}

	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore,
		txUID, existingClusters, addressesWithoutCluster)
}

// isGenericCoinJoin returns true if the given transaction is considered a CoinJoin transaction based on the provided configuration
func isGenericCoinJoin(t clustering.TransactionWithInputOutputAddressCluster, c Config) bool {
	if c.excludeInputCountThreshold > 0 || c.excludeOutputCountThreshold > 0 {
		if c.excludeInputCountThreshold > 0 && len(t.InputAddresses) < c.excludeInputCountThreshold {
			return false
		}

		if c.excludeOutputCountThreshold > 0 && len(t.OutputAddresses) < c.excludeOutputCountThreshold {
			return false
		}
		return true
	}

	return false
}

// Iterate clusters all addresses of the current block based on the multi-input heuristic
func (m *FlatMultiInput) Iterate(ctx context.Context) (bool, error) {
	if m.maxBlocks == 0 {
		return false, serror.FromStr("max blocks must be higher than zero")
	}

	if m.Empty() {
		return false, serror.FromStr("received empty state")
	}

	// state.ID is a new block already, therefore maxBlocks has to be reduced by 1
	toBlockID := min(m.state.Top, m.state.ID+m.maxBlocks-1)
	// get the transaction of the current block height
	transactions, err := clustering.GetAddressesByBlock(ctx, m.db, m.state.ID, toBlockID, clustering.TypeFMI)
	if err != nil {
		return false, err
	}

	if len(transactions) > 0 {
		clusterStore := make(map[string]clustering.Cluster)

		// maps a cluster uid to its merge list
		clusterMergeMap := make(map[string]*newCluster)
		// maps an address uid to its merge list
		addressMergeMap := make(map[string]*newCluster)

		for _, tx := range transactions {
			// tx inputs
			if len(tx.InputAddresses) > 0 {
				if constants.IsMixingTransaction(tx.Type) || isGenericCoinJoin(tx, m.config) {
					// treat inputs of mixing transactions not with the multi-input heuristic
					processAsNonMultiInput(clusterMergeMap, addressMergeMap, clusterStore, tx.UID, tx.InputAddresses)
				} else {
					processAsMultiInput(clusterMergeMap, addressMergeMap, clusterStore, tx.UID, tx.InputAddresses)
				}
			}

			// tx outputs
			processAsNonMultiInput(clusterMergeMap, addressMergeMap, clusterStore, tx.UID, tx.OutputAddresses)
		}

		processedClusters := make(map[*newCluster]bool)
		var operations []clustering.DBOperation
		var clusterIndex int

		operations, err = buildDBOperation(processedClusters, addressMergeMap, clusterIndex)
		if err != nil {
			return false, err
		}

		// increase index
		clusterIndex += len(operations)

		clusters, err := buildDBOperation(processedClusters, clusterMergeMap, clusterIndex)
		if err != nil {
			return false, err
		}
		operations = append(operations, clusters...)

		// insert new clusters
		if len(operations) > 0 {
			// ProcessClusterOperations uses a long-running transaction,
			// therefore transaction retrying has to be handled manually
			if err = db.WithRetry(func() error {
				return clustering.ProcessClusterOperations(ctx, m.db, operations)
			}, 5); err != nil {
				return false, err
			}

			countMergedClusters, countNewAddresses := calculateMetrics(operations)

			// update metrics
			m.mergedClusters.Add(float64(countMergedClusters))
			m.newAddresses.Add(float64(countNewAddresses))
			m.transactions.Add(float64(len(transactions)))
		}
	}

	// set the last clustered block
	if err := dbstat.SetLastClusteredFMIBlockID(ctx, m.db, toBlockID); err != nil {
		return false, err
	}

	m.blocksProcessed = toBlockID - m.state.ID + 1
	m.blocks.Add(float64(m.blocksProcessed))
	m.blockHeight.Set(float64(toBlockID))

	return true, nil
}

func (m *FlatMultiInput) Props() blockiterator.Properties {
	return blockiterator.Properties{
		Name:                        "flat multi-input clustering",
		Context:                     m.ctx,
		CurrentBlock:                m.state.ID,
		ProcessedBlockCount:         m.blocksProcessed,
		SupportsMultiBlockIteration: true,
	}
}

// Next tries to increase the internal state to the next block
func (m *FlatMultiInput) Next(ctx context.Context) (bool, error) {
	status, err := dbstat.GetClassifierStatus(ctx, m.db)
	if err != nil {
		return false, err
	} else if status.LastClassifiedBlockID == nil {
		return false, serror.FromStr("last classified block is not set")
	}

	if m.state.ID <= *status.LastClassifiedBlockID {
		m.state.Top = *status.LastClassifiedBlockID
		return true, nil
	}

	return false, nil
}

func (m *FlatMultiInput) PostExecution(ctx context.Context) error {
	return dbstat.SetClusteringFMI(ctx, m.db, false)
}

func (m *FlatMultiInput) SetMaxBlocks(blockCount int64) {
	m.maxBlocks = blockCount
}

func (m *FlatMultiInput) IncrementState() error {
	m.state.ID += m.blocksProcessed
	return nil
}

// Empty checks if there are more blocks above the current one
func (m *FlatMultiInput) Empty() bool {
	return m.state.ID > m.state.Top
}

// setInitialFMIClusteringID sets the starting FMI clustering block id to 0 if no value has been set yet
func setInitialFMIClusteringID(ctx context.Context, dgraph external.Database) error {
	status, err := dbstat.GetClusteringFMIStatus(ctx, dgraph)
	if err != nil {
		return err
	}

	if status.LastClusteredBlockID == nil {
		if err = dbstat.SetLastClusteredFMIBlockID(ctx, dgraph, 0); err != nil {
			return err
		}
	}
	return nil
}

// addClustersToMergeList adds newClusters and newAddresses to clusterMergeMap and
// addressMergeMap with information from clusterStore
func addClustersToMergeList(clusterMergeMap map[string]*newCluster, addressMergeMap map[string]*newCluster,
	clusterStore map[string]clustering.Cluster, tx string, newClusters map[string]bool, newAddresses map[string]bool) {
	if len(newClusters) == 0 && len(newAddresses) == 0 {
		return
	}

	// find cluster merge lists which have elements of newClusters
	foundListMap := make(map[*newCluster]bool)

	if len(clusterMergeMap) > 0 {
		for k := range newClusters {
			if v, ok := clusterMergeMap[k]; ok {
				foundListMap[v] = true
			}
		}
	}

	if len(addressMergeMap) > 0 {
		for k := range newAddresses {
			if v, ok := addressMergeMap[k]; ok {
				foundListMap[v] = true
			}
		}
	}

	// if all clusters do not exist yet in the map -> add all of them as a new merge list
	if len(foundListMap) == 0 {
		var mergeList []clustering.Cluster

		for k := range newClusters {
			mergeList = append(mergeList, clusterStore[k])
		}

		nc := newCluster{
			changeTransaction: tx,
			mergeList:         mergeList,
			addresses:         newAddresses,
		}

		for k := range newClusters {
			clusterMergeMap[k] = &nc
		}

		for k := range newAddresses {
			addressMergeMap[k] = &nc
		}

		return
	}

	var mergeListPtr *newCluster
	var createdNewList bool
	// either create new merge list or find existing one
	if len(foundListMap) > 1 {
		var mergeList []clustering.Cluster
		addressList := make(map[string]bool)
		for list := range foundListMap {
			mergeList = append(mergeList, list.mergeList...)
			for a := range list.addresses {
				addressList[a] = true
			}
		}

		createdNewList = true
		mergeListPtr = &newCluster{
			mergeList: mergeList,
			addresses: addressList,
		}
	} else {
		// map is only one element big
		mergeListPtr = cliutil.GetOneKey(foundListMap)
	}

	mergeListPtr.changeTransaction = tx

	// new addresses to newCluster
	for a := range newAddresses {
		mergeListPtr.addresses[a] = true
	}

	// find new clusters by querying the clusterMergeMap and append them to the mergeList
	for k := range newClusters {
		if _, ok := clusterMergeMap[k]; !ok {
			mergeListPtr.mergeList = append(mergeListPtr.mergeList, clusterStore[k])
		}
	}

	// if a new merge list was created, set the references to it for all items,
	// otherwise only for the new items
	if createdNewList {
		for _, cluster := range mergeListPtr.mergeList {
			clusterMergeMap[cluster.UID] = mergeListPtr
		}

		for a := range mergeListPtr.addresses {
			addressMergeMap[a] = mergeListPtr
		}
	} else {
		for cluster := range newClusters {
			clusterMergeMap[cluster] = mergeListPtr
		}

		for address := range newAddresses {
			addressMergeMap[address] = mergeListPtr
		}
	}
}

// buildDBOperation creates database operations based on the given items. Items contained
// in processedClusters are not processed. The clusterIndex controls the temporary UID (_:0x123) of the new clusters.
func buildDBOperation(processedClusters map[*newCluster]bool, items map[string]*newCluster,
	clusterIndex int) ([]clustering.DBOperation, error) {
	var operations []clustering.DBOperation //nolint:prealloc

	for _, i := range items {
		if processedClusters[i] {
			continue
		}

		processedClusters[i] = true

		if len(i.mergeList) == 0 && len(i.addresses) == 0 {
			return nil, serror.FromStr("no clusters and addresses")
		}

		clusterIndex++
		var cluster *clustering.Cluster
		var addressCount int
		var oldClusters []string

		if len(i.mergeList) > 0 {
			for _, c := range i.mergeList {
				if c.AddressCount == nil {
					return nil, serror.FromFormat("address count is not set for cluster %s", c.UID)
				}
			}

			slices.SortFunc(i.mergeList, func(a, b clustering.Cluster) int {
				return *a.AddressCount - *b.AddressCount
			})

			// find the largest cluster, so we have to move the least amount of addresses
			var largestClusterUID string
			var clusterAddressCount int
			// at which point to exclude clusters from the mergelist, because otherwise the cluster size gets to high
			stopIndex := -1
			for y, c := range i.mergeList {
				if clusterAddressCount+*c.AddressCount > maxClusterSize {
					break
				}

				stopIndex = y
				clusterAddressCount += *c.AddressCount
			}

			if stopIndex >= 0 {
				largestClusterUID = i.mergeList[stopIndex].UID
				addressCount += clusterAddressCount
				for y, c := range i.mergeList {
					if y == stopIndex {
						// reached end of cluster which should be merged
						break
					}

					oldClusters = append(oldClusters, c.UID)
				}

				cluster = clustering.NewFMIClusterByUID(largestClusterUID)
			}
		}

		// if no clusters have been merged, create a new cluster
		if cluster == nil {
			cluster = clustering.NewFMICluster(clusterIndex)
		}

		cluster.Transaction.UID = i.changeTransaction

		// add addresses
		addressCount += len(i.addresses)
		cluster.AddressCount = &addressCount
		for address := range i.addresses {
			cluster.Addresses = append(cluster.Addresses, db.UIDNode{UID: address})
		}
		operations = append(operations, clustering.DBOperation{
			NewCluster:  *cluster,
			OldClusters: oldClusters,
		})
	}

	return operations, nil
}

func calculateMetrics(operations []clustering.DBOperation) (newClusterCount int, newAddressCount int) {
	for _, op := range operations {
		newAddressCount += len(op.NewCluster.Addresses)

		if op.OldClusters != nil {
			newClusterCount += len(op.OldClusters)
		}
	}

	return
}
