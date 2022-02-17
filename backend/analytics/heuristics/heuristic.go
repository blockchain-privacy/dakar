package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"backend/db/analytics/heuristics"
	dbop "backend/db/output"
	"backend/db/transaction"
	"backend/external"

	"errors"
	"fmt"
	"time"
)

// errorNoOriginsAtStart defines an error which should be used when no origins are available
var errorNoOriginsAtStart = errors.New("no origins can be fetched")

//
var errorInvalidClusterTypes = errors.New("cluster types are not valid")

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
	exec(dgraph external.Database, g *graph.Wrapper, txHash string, parentHeuristicUID string) ([]heuristics.HeuristicCluster, error)
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
				err = fmt.Errorf("found more than one denomination type in input transaction %s for destination tx %s",
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
func getDenominationCountsWithFilter(it heuristics.HeuristicTransaction, filterTx string) [dbop.NumDenominations]int {
	var denominations []int64
	for _, output := range it.Outputs {
		if filterTx != "" && output.InputTransaction != filterTx {
			continue
		}
		denominations = append(denominations, output.Amount)
	}

	return dbop.CountAmountDenominations(denominations)
}

// gets the counts of each denomination type
func getDenominationCounts(it heuristics.HeuristicTransaction) [dbop.NumDenominations]int {
	var denominations []int64
	for _, output := range it.Outputs {
		denominations = append(denominations, output.Amount)
	}

	return dbop.CountAmountDenominations(denominations)
}

type clusterDenominations struct {
	denominationIndex int
	// key: cluster id, value: number of denominations of type denominationIndex
	clusters map[heuristics.ClusterUID]int
}

// addOriginsToMap adds all origins to their respective source in sourceTransactionMap.
// The returned map contains the provided origins
func addOriginsToMap(sourceTransactionMap map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction,
	origins []heuristics.HeuristicTransaction) (map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction, error) {

	for _, o := range origins {

		// add transaction to sourceTransactionMap
		transactions := sourceTransactionMap[o.Cluster]

		if len(transactions) == 0 {
			transactions = make(map[string]heuristics.HeuristicTransaction)
		}

		transactions[o.UID] = o
		sourceTransactionMap[o.Cluster] = transactions
	}

	return sourceTransactionMap, nil
}

// countClusterDenominations creates a map of clusters with the
// number of denominations of the specified denomination type
func countClusterDenominations(origins []heuristics.HeuristicTransaction, denominationIndex int) (
	oSource clusterDenominations, err error) {
	oSource.denominationIndex = denominationIndex
	oSource.clusters = make(map[heuristics.ClusterUID]int)
	for _, o := range origins {
		nDenominations := getDenominationCounts(o)[denominationIndex]
		oSource.clusters[o.Cluster] += nDenominations
	}

	return
}

func buildSourceAmounts(origins map[string]heuristics.HeuristicTransaction) map[heuristics.ClusterUID][dbop.NumDenominations]int {
	sourceAmounts := make(map[heuristics.ClusterUID][dbop.NumDenominations]int)

	for _, o := range origins {
		denominationSlice := getDenominationCounts(o)
		for i := range denominationSlice {
			denominationSlice[i] += sourceAmounts[o.Cluster][i]
		}

		sourceAmounts[o.Cluster] = denominationSlice
	}
	return sourceAmounts
}

// getKeySlice returns a slice containing all keys of the given map
func getKeySlice(m map[string]bool) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}

	return
}

// getTimeLimitedOrigins returns all origins of the given transaction.
// If lookBackTime is bigger than zero only origins in the time range of
// tx.ts - lookBackTime will be returned.
func getTimeLimitedOrigins(dgraph external.Database, g *graph.Wrapper, tx heuristics.HeuristicTransaction,
	lookBackTime time.Duration, userUID string, clusterTypes []clustering.ClusterType) (
	origins []heuristics.HeuristicTransaction, attributionMapping map[heuristics.ClusterUID][]string, err error) {
	// do reverse lookup
	endpoints, err := g.ReverseLookup(tx.UID, lookBackTime)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// get tx details for each uid
	origins, attributionMapping, err = heuristics.GetTransactionsWithOutputAmountAndCluster(dgraph, getKeySlice(endpoints),
		userUID, clusterTypes)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return
}

// getDestinationTxOrigins returns all origins of the given
// transaction, limited to a look back time of 90 days.
func getDestinationTxOrigins(dgraph external.Database, g *graph.Wrapper, txHash string, userUID string,
	requestedClusterTypes []clustering.ClusterType) ([]heuristics.HeuristicTransaction,
	map[heuristics.ClusterUID][]string, error) {
	origins, attributionMapping, err := getDestinationTxOriginsTimeLimited(dgraph, g, txHash, time.Hour*24*90,
		userUID, requestedClusterTypes)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	return origins, attributionMapping, nil
}

// getDestinationTxOriginsTimeLimited returns all origins of the given
// transaction, for the given time limit.
func getDestinationTxOriginsTimeLimited(dgraph external.Database, g *graph.Wrapper, txHash string,
	dur time.Duration, userUID string, requestedClusterTypes []clustering.ClusterType) (
	origins []heuristics.HeuristicTransaction, attributionMapping map[heuristics.ClusterUID][]string, err error) {
	// get uid for txhash
	uid, err := transaction.GetTransactionUID(dgraph, txHash)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	inputTransactions, err := g.GetInputTransactions(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	uidMap := make(map[string]bool)
	// do reverse lookup for all input transactions
	for _, it := range inputTransactions {
		endpoints, lookupErr := g.ReverseLookup(it, dur)
		if lookupErr != nil {
			return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), lookupErr)
		}

		for k := range endpoints {
			uidMap[k] = true
		}
	}

	// get tx details for each uid
	origins, attributionMapping, err = heuristics.GetTransactionsWithOutputAmountAndCluster(dgraph, getKeySlice(uidMap),
		userUID, requestedClusterTypes)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return
}

// getOriginDestinationTimeLimited returns UID map of all destinations of the given origin UIDs
func getOriginDestinationTimeLimited(g *graph.Wrapper, originUIDs []string,
	dur time.Duration) (map[string]bool, error) {
	uidMap := make(map[string]bool)
	// do forward lookup for all origin transactions
	for _, it := range originUIDs {
		endpoints, lookupErr := g.ForwardLookupByTime(it, dur)
		if lookupErr != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), lookupErr)
		}

		for k := range endpoints {
			uidMap[k] = true
		}
	}

	return uidMap, nil
}

// getOriginDestinationsWithOutputs returns all destinations
// of the given transactions limited by time. Each transaction contains its outputs.
func getOriginDestinationsWithOutputs(dgraph external.Database, g *graph.Wrapper, originUIDs []string,
	dur time.Duration, userUID string, requestedClusterTypes []clustering.ClusterType) (
	origins []heuristics.HeuristicTransaction, attributionMapping map[heuristics.ClusterUID][]string, err error) {
	uidMap, err := getOriginDestinationTimeLimited(g, originUIDs, dur)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// get tx details for each uid
	origins, attributionMapping, err = heuristics.GetTransactionsWithOutputAmountAndCluster(dgraph,
		getKeySlice(uidMap), userUID, requestedClusterTypes)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return
}

// getOriginDestinationsWithInputs returns all destinations
// of the given transactions limited by time. Each transaction contains its inputs.
func getOriginDestinationsWithInputs(dgraph external.Database, g *graph.Wrapper,
	originUIDs []string, dur time.Duration) (origins []heuristics.HeuristicTransaction, err error) {
	uidMap, err := getOriginDestinationTimeLimited(g, originUIDs, dur)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// get tx details for each uid
	origins, err = heuristics.GetTransactionsWithInputAmount(dgraph, getKeySlice(uidMap))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return
}

func isParentHeuristicSet(parentHeuristicUID string) bool {
	return parentHeuristicUID != ""
}

// heuristicExecutor holds information for executing on heuristic and its children
type heuristicExecutor struct {
	rootUID        string
	thisHeuristic  heuristic
	nextHeuristics []heuristicExecutor
}

// buildExecutor is a convenience function for building heuristic executors
func buildExecutor(thisHeuristic heuristic, nextHeuristics ...heuristicExecutor) heuristicExecutor {
	return heuristicExecutor{
		thisHeuristic:  thisHeuristic,
		nextHeuristics: nextHeuristics,
	}
}

// run start the execution of the given heuristic executor. The executor runs initial heuristic and
// triggers the run function of all nextHeuristics. If parentHeuristicUID is not
// set (e.g. "") than the heuristicExecutor.rootUID is used
func (hx heuristicExecutor) run(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string, userUID string) error {
	thisRootUID := hx.rootUID
	if parentHeuristicUID != "" {
		thisRootUID = parentHeuristicUID
	}

	newUID, err := exec(dgraph, g, txHash, thisRootUID, hx.thisHeuristic, userUID)
	if err != nil {
		// two fmt.Errorf so the error gets wrapped
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(),
			fmt.Errorf("heuristic type: %s, parameter: %s, %s",
				hx.thisHeuristic.getType(), hx.thisHeuristic.getParameterString(), err))
	}

	for _, executor := range hx.nextHeuristics {
		if runErr := executor.run(dgraph, g, txHash, newUID, userUID); runErr != nil {
			// two fmt.Errorf so the error gets wrapped
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(),
				fmt.Errorf("heuristic type: %s, parameter: %s, %s",
					executor.thisHeuristic.getType(), executor.thisHeuristic.getParameterString(), runErr))
		}
	}

	var returnError error

	return returnError
}

// exec executes the heuristic on the transaction specified by txHash for the given userUID
func exec(dgraph external.Database, g *graph.Wrapper, txHash string, parentHeuristicUID string, h heuristic,
	userUID string) (thisUID string, err error) {
	heuristicClusters, err := h.exec(dgraph, g, txHash, parentHeuristicUID)
	if err != nil && !errors.Is(err, errorNoOriginsAtStart) {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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

	thisUID, err = heuristics.InsertHeuristic(dgraph, heuristics.Heuristic{
		HeuristicType:   h.getType(),
		Clusters:        heuristicClusters,
		Parameter:       h.getParameterString(),
		ParentHeuristic: pHeuristic,
		TxHash:          txHash,
	}, userUID)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// createHeuristicClusters converts the given map into HeuristicCluster's
func createHeuristicClusters(clusterMap map[heuristics.ClusterUID][]heuristics.HeuristicResult,
	attributionMap map[heuristics.ClusterUID][]string) []heuristics.HeuristicCluster {
	var resultCluster []heuristics.HeuristicCluster
	for clusterID, results := range clusterMap {
		var attributions []attribution.Attribution
		if attributionMap != nil {
			if attrs, ok := attributionMap[clusterID]; ok {
				for _, a := range attrs {
					attributions = append(attributions, attribution.Attribution{Uid: a})
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
