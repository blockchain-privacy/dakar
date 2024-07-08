package heuristics

import (
	"backend/analytics"
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"backend/db/analytics/exclusion"
	"backend/db/analytics/heuristics"
	"backend/external"
	"errors"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"time"
)

// errNoOriginsAtStart defines an error which should be used when no origins are available
var errNoOriginsAtStart = errors.New("no origins can be fetched")

var errInvalidClusterTypes = errors.New("cluster types are not valid")

const (
	// heuristicCategoryReverse defines a category string for the frontend to order the heuristic
	heuristicCategoryReverse = "Reverse"
	// heuristicCategoryForward defines a category string for the frontend to order the heuristic
	heuristicCategoryForward = "Forward"
)

type Descriptor struct {
	Title       string `json:"title,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	// pointer so Parameter does not appear in JSON if not set
	Parameter *struct {
		DefaultValue string `json:"value,omitempty"`
		Description  string `json:"description,omitempty"`
		// Type must be one of the following values: 'int', 'string'
		Type string `json:"type,omitempty"`
	} `json:"parameter,omitempty"`
}

type heuristic interface {
	// exec executes the heuristic and returns the altered set of origin uids
	exec(dgraph external.Database, g *graph.Wrapper, txHash string,
		parentHeuristicUID string) ([]heuristics.HeuristicCluster, error)
	// getType returns the heuristic type
	getType() string
	// getParameterString returns the used parameter for this heuristic as a string
	getParameterString() string
	// hasParameter returns true if this heuristic has a parameter
	hasParameter() bool
	// setParameter sets the parameter
	setParameter(string) error
	// setClusterTypes sets the cluster types, which are used to cluster the results of the heuristic.
	// If cluster types are set to nil, the result will not be clustered.
	// If multiple cluster types are set, then the consolidation of these clusters will be used.
	setClusterTypes([]clustering.ClusterType) error
	// getClusterTypes returns the cluster types this heuristic uses to cluster addresses
	getClusterTypes() []clustering.ClusterType
	// setExcludeAddresses sets whether certain addresses should be excluded from the lookups
	setExcludeAddresses(bool)
	// getExcludeAddresses returns whether certain addresses should be excluded from the lookups
	getExcludeAddresses() bool
	// setExcludeSpendingGaps sets whether mixing outputs with a spending gap should be traversed
	setExcludeSpendingGaps(bool)
	// getExcludeSpendingGaps returns whether mixing outputs with a spending gap should be traversed
	getExcludeSpendingGaps() bool
	// setUserUID sets the UID of the user who created this heuristic
	setUserUID(string)
	// String returns the heuristic in string format
	String() string
	// GetDescriptor returns description of the heuristic and its expected parameter for the frontend
	GetDescriptor() Descriptor
	// clone copies an instance of this interface. This method is needed because
	// instances of interfaces can not be easily copied-by-value.
	// More information: https://stackoverflow.com/questions/37851500/how-to-copy-an-interface-value-in-go
	clone() heuristic
}

// getNumberOfDenominations returns the number of denominations. If destinationTransaction is set, it
// only counts outputs with input transactions equal to destinationTransaction. An error is returned
// if more than one type of denominations is found
func getNumberOfDenominations(it heuristics.HeuristicTransaction, destinationTransaction string) (nDenominations int,
	denomIndex int, err error) {
	numDenominations := getDenominationCountsWithFilter(it, destinationTransaction)

	found := false
	for i, nd := range numDenominations {
		if nd > 0 {
			if found {
				err = serror.FromFormat("found more than one denomination type in input transaction %s for destination tx %s",
					it, destinationTransaction)
				return
			}
			denomIndex = i
			found = true
		}
	}
	nDenominations = numDenominations[denomIndex]
	return
}

// getDenominationCountsWithFilter gets the counts of each denomination type.
// If filterTx is set, it only counts outputs with input transactions equal to filterTx.
func getDenominationCountsWithFilter(it heuristics.HeuristicTransaction, filterTx string) [analytics.NumDenominations]int {
	var denominations []int64 //nolint:prealloc
	for _, output := range it.Outputs {
		if filterTx != "" && output.InputTransaction != filterTx {
			continue
		}
		denominations = append(denominations, output.Amount)
	}

	return analytics.CountAmountDenominations(denominations)
}

// gets the counts of each denomination type
func getDenominationCounts(it heuristics.HeuristicTransaction) [analytics.NumDenominations]int {
	denominations := make([]int64, len(it.Outputs))
	for i, output := range it.Outputs {
		denominations[i] = output.Amount
	}

	return analytics.CountAmountDenominations(denominations)
}

type clusterDenominations struct {
	denominationIndex int
	// key: cluster id, value: number of denominations of type denominationIndex
	clusters map[heuristics.ClusterUID]int
}

// addOriginsToMap adds all origins to their respective source in sourceTransactionMap.
// The returned map contains the provided origins
func addOriginsToMap(sourceTransactionMap map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction,
	origins []heuristics.HeuristicTransaction) map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction {
	for _, o := range origins {
		// add transaction to sourceTransactionMap
		transactions := sourceTransactionMap[o.Cluster]

		if len(transactions) == 0 {
			transactions = make(map[string]heuristics.HeuristicTransaction)
		}

		transactions[o.UID] = o
		sourceTransactionMap[o.Cluster] = transactions
	}

	return sourceTransactionMap
}

// countClusterDenominations creates a map of clusters with the
// number of denominations of the specified denomination type
func countClusterDenominations(origins []heuristics.HeuristicTransaction,
	denominationIndex int) (oSource clusterDenominations) {
	oSource.denominationIndex = denominationIndex
	oSource.clusters = make(map[heuristics.ClusterUID]int)
	for _, o := range origins {
		nDenominations := getDenominationCounts(o)[denominationIndex]
		oSource.clusters[o.Cluster] += nDenominations
	}

	return
}

func buildSourceAmounts(origins map[string]heuristics.HeuristicTransaction) map[heuristics.ClusterUID][analytics.NumDenominations]int {
	sourceAmounts := make(map[heuristics.ClusterUID][analytics.NumDenominations]int)

	for _, o := range origins {
		denominationSlice := getDenominationCounts(o)
		for i := range denominationSlice {
			denominationSlice[i] += sourceAmounts[o.Cluster][i]
		}

		sourceAmounts[o.Cluster] = denominationSlice
	}
	return sourceAmounts
}

// getTimeLimitedOrigins returns all origins of the given transaction.
// If lookBackTime is bigger than zero only origins in the time range of
// tx.ts - lookBackTime will be returned.
func getTimeLimitedOrigins(dgraph external.Database, g *graph.Wrapper, tx heuristics.HeuristicTransaction,
	lookBackTime time.Duration, userUID string, clusterTypes []clustering.ClusterType, exclusions []string,
	excludeSpendingGaps bool) (
	origins []heuristics.HeuristicTransaction, attributionMapping map[heuristics.ClusterUID][]string, err error) {
	// do reverse lookup
	endpoints, err := g.ReverseLookup(tx.UID, lookBackTime, exclusions, excludeSpendingGaps)
	if err != nil {
		return nil, nil, err
	}

	// get tx details for each uid
	return heuristics.GetTransactionsWithOutputAmountAndCluster(dgraph, cliutil.GetMapKeys(endpoints),
		userUID, clusterTypes)
}

// getDestinationTxOrigins returns all origins of the given
// transaction, limited to a look back time of 90 days.
func getDestinationTxOrigins(dgraph external.Database, g *graph.Wrapper, txHash string, userUID string,
	requestedClusterTypes []clustering.ClusterType, excludeAddresses bool,
	excludeSpendingGaps bool) ([]heuristics.HeuristicTransaction,
	map[heuristics.ClusterUID][]string, error) {
	origins, attributionMapping, err := getDestinationTxOriginsTimeLimited(dgraph, g, txHash, time.Hour*24*90,
		userUID, requestedClusterTypes, excludeAddresses, excludeSpendingGaps)
	if err != nil {
		return nil, nil, err
	}
	return origins, attributionMapping, nil
}

// getDestinationTxOriginsTimeLimited returns all origins of the given
// transaction, for the given time limit.
func getDestinationTxOriginsTimeLimited(dgraph external.Database, g *graph.Wrapper, txHash string, dur time.Duration,
	userUID string, requestedClusterTypes []clustering.ClusterType, excludeAddresses bool, excludeSpendingGaps bool) (
	origins []heuristics.HeuristicTransaction, attributionMapping map[heuristics.ClusterUID][]string, err error) {
	// get uid for txhash
	uid, err := db.GetTransactionUID(dgraph, txHash)
	if err != nil {
		return nil, nil, err
	}

	inputTransactions, err := g.GetInputTransactions(uid)
	if err != nil {
		return nil, nil, err
	}

	var exclusions []string
	if excludeAddresses {
		exclusions, err = exclusion.GetAddressExclusionUIDs(dgraph, userUID)
		if err != nil {
			return nil, nil, err
		}
	}

	uidMap := make(map[string]bool)
	// do reverse lookup for all input transactions
	for _, it := range inputTransactions {
		endpoints, lookupErr := g.ReverseLookup(it, dur, exclusions, excludeSpendingGaps)
		if lookupErr != nil {
			return nil, nil, lookupErr
		}

		for k := range endpoints {
			uidMap[k] = true
		}
	}

	// get tx details for each uid
	origins, attributionMapping, err = heuristics.GetTransactionsWithOutputAmountAndCluster(dgraph, cliutil.GetMapKeys(uidMap),
		userUID, requestedClusterTypes)
	if err != nil {
		return nil, nil, err
	}

	return
}

// getOriginDestinationTimeLimited returns UID map of all destinations of the given origin UIDs
func getOriginDestinationTimeLimited(g *graph.Wrapper, originUIDs []string,
	dur time.Duration, exclusions []string, excludeSpendingGaps bool) (map[string]bool, error) {
	uidMap := make(map[string]bool)
	// do forward lookup for all origin transactions
	for _, it := range originUIDs {
		endpoints, lookupErr := g.ForwardLookupByTime(it, dur, exclusions, excludeSpendingGaps)
		if lookupErr != nil {
			return nil, lookupErr
		}

		for k := range endpoints {
			uidMap[k] = true
		}
	}

	return uidMap, nil
}

// getOriginDestinationsWithInputs returns all destinations
// of the given transactions limited by time. Each transaction contains its inputs.
func getOriginDestinationsWithInputs(dgraph external.Database, g *graph.Wrapper,
	originUIDs []string, dur time.Duration, exclusions []string,
	excludeSpendingGaps bool) (origins []heuristics.HeuristicTransaction, err error) {
	uidMap, err := getOriginDestinationTimeLimited(g, originUIDs, dur, exclusions, excludeSpendingGaps)
	if err != nil {
		return nil, err
	}

	// get tx details for each uid
	return heuristics.GetTransactionsWithInputAmount(dgraph, cliutil.GetMapKeys(uidMap))
}

func isParentHeuristicSet(parentHeuristicUID string) bool {
	return parentHeuristicUID != ""
}

// Executor holds information for executing on heuristic and its children
type Executor struct {
	rootUID         string
	thisHeuristic   heuristic
	transactionHash string
}

// ConstructExecutors creates executors based on heuristics
func ConstructExecutors(heuristicRequest heuristics.DatabaseHeuristicRequest, userUID string) (executor Executor, err error) {
	if heuristicRequest.TransactionHash == "" {
		err = serror.FromFormat("transaction of heuristic request is empty: %v", heuristicRequest)
		return
	}

	modelHeuristic, ok := typeMap[heuristicRequest.Type]
	if !ok || (modelHeuristic.hasParameter() && len(heuristicRequest.Parameter) == 0) {
		err = serror.New(errHeuristicNotValid)
		return
	}

	// copy parameters from heuristic request into newly created heuristic
	clonedHeuristic := modelHeuristic.clone()
	if clonedHeuristic.hasParameter() {
		err = clonedHeuristic.setParameter(heuristicRequest.Parameter)
		if err != nil {
			return
		}
	}

	if err = clonedHeuristic.setClusterTypes(heuristicRequest.ClusterTypes); err != nil {
		return
	}

	clonedHeuristic.setUserUID(userUID)
	clonedHeuristic.setExcludeAddresses(heuristicRequest.ExcludeAddresses)
	clonedHeuristic.setExcludeSpendingGaps(heuristicRequest.ExcludeSpendingGaps)

	executor = Executor{
		thisHeuristic:   clonedHeuristic,
		rootUID:         heuristicRequest.ParentHeuristicUID,
		transactionHash: heuristicRequest.TransactionHash,
	}

	return
}

// Run start the execution of the given heuristic executor. If parentHeuristicUID is not
// set (e.g. "") than the Executor.rootUID is used
func (hx Executor) Run(dgraph external.Database, g *graph.Wrapper) (*heuristics.Heuristic, error) {
	newHeuristic, err := exec(dgraph, g, hx.transactionHash, hx.rootUID, hx.thisHeuristic)
	if err != nil {
		return nil, fmt.Errorf("heuristic type: %s, parameter: %s, %w",
			hx.thisHeuristic.getType(), hx.thisHeuristic.getParameterString(), err)
	}

	return newHeuristic, nil
}

// exec executes the heuristic on the transaction specified by txHash for the given userUID
func exec(dgraph external.Database, g *graph.Wrapper, txHash string, parentHeuristicUID string,
	h heuristic) (newHeuristic *heuristics.Heuristic, err error) {
	heuristicClusters, err := h.exec(dgraph, g, txHash, parentHeuristicUID)
	if err != nil && !errors.Is(err, errNoOriginsAtStart) {
		return
	}

	// set DType
	for i := range heuristicClusters {
		heuristicClusters[i].SetDType()
		for y := range heuristicClusters[i].Results {
			heuristicClusters[i].Results[y].SetDType()
		}
	}

	// only set parent heuristic if uid is provided
	var pHeuristic []heuristics.Heuristic
	if parentHeuristicUID != "" {
		pHeuristic = []heuristics.Heuristic{{UID: parentHeuristicUID}}
	}

	clusterTypes := make([]string, len(h.getClusterTypes()))
	for i, cType := range h.getClusterTypes() {
		clusterTypes[i] = string(cType)
	}

	shouldExcludeAddresses := h.getExcludeAddresses()
	shouldExcludeSpendingGaps := h.getExcludeSpendingGaps()

	newHeuristic = &heuristics.Heuristic{
		HeuristicType:       h.getType(),
		ClusterTypes:        clusterTypes,
		ExcludeAddresses:    &shouldExcludeAddresses,
		ExcludeSpendingGaps: &shouldExcludeSpendingGaps,
		Clusters:            heuristicClusters,
		Parameter:           h.getParameterString(),
		ParentHeuristic:     pHeuristic,
		TxHash:              txHash,
		Timestamp:           time.Now().UTC().Format(time.RFC3339),
	}

	return
}

// createHeuristicClusters converts the given map into HeuristicCluster's
func createHeuristicClusters(clusterMap map[heuristics.ClusterUID][]heuristics.HeuristicResult,
	attributionMap map[heuristics.ClusterUID][]string) []heuristics.HeuristicCluster {
	resultCluster := make([]heuristics.HeuristicCluster, 0, len(clusterMap))
	for clusterID, results := range clusterMap {
		var attributions []attribution.Attribution
		if attributionMap != nil {
			if attrs, ok := attributionMap[clusterID]; ok {
				for _, a := range attrs {
					attributions = append(attributions, attribution.Attribution{UID: a})
				}
			}
		}

		resultCluster = append(resultCluster, heuristics.HeuristicCluster{
			Results:      results,
			Attributions: attributions,
		})
	}

	return resultCluster
}
