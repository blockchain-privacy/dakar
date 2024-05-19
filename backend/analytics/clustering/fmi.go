package clustering

import (
	"backend/blockiterator"
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db/analytics/clustering"
	dbstat "backend/db/status"
	"backend/external"
	"log/slog"

	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// FlatMultiInput implements BlockIterator which creates clusters via the multi-input heuristic
type FlatMultiInput struct {
	db    external.Database
	ctx   context.Context
	state blockiterator.State

	// how many blocks are processed in one interation at maximum
	maxBlocks uint64

	// number of blocks which have been processed by the last Iterate call
	blocksProcessed uint64

	blocks         prometheus.Counter
	transactions   prometheus.Counter
	mergedClusters prometheus.Counter
	newAddresses   prometheus.Counter
	blockHeight    prometheus.Gauge
}

// NewFlatMultiInput creates a new flat multi-input clustering object
func NewFlatMultiInput(ctx context.Context, dgraph external.Database) *FlatMultiInput {
	return &FlatMultiInput{
		db:        dgraph,
		ctx:       ctx,
		maxBlocks: 100,
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
		return err
	}

	if err := setInitialFMIClusteringID(m.db); err != nil {
		return err
	}

	classifierStatus, err := dbstat.GetClassifierStatus(m.db)
	if err != nil {
		return err
	}

	clusteringStatus, err := dbstat.GetClusteringFMIStatus(m.db)
	if err != nil {
		return err
	}

	if clusteringStatus.LastClusteredBlockID == nil {
		return cliutil.NewStackErrorStr("last FMI clustered block is not set")
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

// Iterate clusters all addresses of the current block based on the multi-input heuristic
func (m *FlatMultiInput) Iterate() (bool, error) {
	if m.Empty() {
		return false, cliutil.NewStackErrorStr("got empty state")
	}

	// state.ID is a new block already, therefore maxBlocks has to be reduced by 1
	toBlockID := min(m.state.Top, m.state.ID+m.maxBlocks-1)
	// get the transaction of the current block height
	transactions, err := clustering.GetAddressesByBlock(m.db, m.state.ID, toBlockID, clustering.TypeFMI)
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
				if tx.PrivacyType != nil && constants.PrivacyType(*tx.PrivacyType).IsMixing() {
					// treat inputs of mixing transations not with the multi-input heuristic
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

		clusters, clusterErr := buildDBOperation(processedClusters, clusterMergeMap, clusterIndex)
		if err != nil {
			return false, clusterErr
		}
		operations = append(operations, clusters...)

		// insert new clusters
		if len(operations) > 0 {
			opErr := clustering.ProcessClusterOperations(m.db, operations)
			if opErr != nil {
				return false, opErr
			}

			countMergedClusters, countNewAddresses := calculateMetrics(operations)

			// update metrics
			m.mergedClusters.Add(float64(countMergedClusters))
			m.newAddresses.Add(float64(countNewAddresses))
			m.transactions.Add(float64(len(transactions)))
		}
	}

	// set the last clustered block
	if statusErr := dbstat.SetLastClusteredFMIBlockID(m.db, toBlockID); statusErr != nil {
		return false, statusErr
	}

	m.blocksProcessed = toBlockID - m.state.ID + 1
	m.blocks.Add(float64(m.blocksProcessed))
	m.blockHeight.Set(float64(toBlockID))

	return true, nil
}

// ProcessedBlockCount returns the number of blocks processed by a Iterate call
func (m *FlatMultiInput) ProcessedBlockCount() uint64 {
	return m.blocksProcessed
}

// NextBlock tries to increase the internal state to the next block
func (m *FlatMultiInput) NextBlock() (bool, error) {
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

func (m *FlatMultiInput) PostExecution() error {
	return dbstat.SetClusteringFMI(m.db, false)
}

func (m *FlatMultiInput) IncrementState() error {
	m.state.ID += m.blocksProcessed
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
func (m *FlatMultiInput) Logger() *slog.Logger {
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
func setInitialFMIClusteringID(dgraph external.Database) error {
	status, err := dbstat.GetClusteringFMIStatus(dgraph)
	if err != nil {
		return err
	}

	if status.LastClusteredBlockID == nil {
		if err = dbstat.SetLastClusteredFMIBlockID(dgraph, 0); err != nil {
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
			return nil, cliutil.NewStackErrorStr("no clusters and addresses")
		}

		clusterIndex++
		var cluster clustering.Cluster
		var addressCount int
		var oldClusters []string

		if len(i.mergeList) > 0 {
			// find the largest cluster, so we have to move the least amount of addresses
			var largestClusterUID string
			var largestAddressesCount int
			for _, c := range i.mergeList {
				if c.AddressCount == nil {
					return nil, cliutil.NewStackErrorf("address count is not set for cluster %s", c.UID)
				}
				addressCount += *c.AddressCount
				if *c.AddressCount > largestAddressesCount {
					largestClusterUID = c.UID
					largestAddressesCount = *c.AddressCount
				}
			}

			for _, c := range i.mergeList {
				if c.UID != largestClusterUID {
					oldClusters = append(oldClusters, c.UID)
				}
			}

			cluster = clustering.NewFMIClusterByUID(largestClusterUID)
		} else {
			cluster = clustering.NewFMICluster(clusterIndex)
		}

		cluster.Transaction.UID = i.changeTransaction

		// add addresses
		addressCount += len(i.addresses)
		cluster.AddressCount = &addressCount
		for address := range i.addresses {
			cluster.Addresses = append(cluster.Addresses, clustering.HollowAddress{UID: address})
		}
		operations = append(operations, clustering.DBOperation{
			NewCluster:  cluster,
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
