package heuristics

import (
	"backend/analytics/classifier/dash"
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"backend/db/analytics/heuristics"
	"backend/db/workspace"
	"backend/external"
	"context"
	"errors"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"time"
)

type heuristicConstructor func() heuristic

var ConstructorMap = make(map[string]heuristicConstructor)

var (
	errHeuristicNotValid = errors.New("error heuristics are not valid")
	// errNoOriginsAtStart defines an error which should be used when no origins are available
	errNoOriginsAtStart    = errors.New("no origins could be fetched")
	errInvalidClusterTypes = errors.New("cluster types are not valid")
)

const (
	// heuristicCategoryReverse defines a category string for the frontend to order the heuristic
	heuristicCategoryReverse = "Reverse"
	// heuristicCategoryForward defines a category string for the frontend to order the heuristic
	heuristicCategoryForward = "Forward"
)

const (
	// Dash
	heuristicTypeReverseLookup    = "reverse_lookup"
	heuristicTypeOneSource        = "one_source"
	heuristicTypeReverseAmount    = "reverse_amount"
	heuristicTypePerfect          = "perfect_match"
	heuristicTypeDenominationType = "denomination_type"
	heuristicTypeForwardAmount    = "forward_amount"
	heuristicTypeForwardLookup    = "forward_lookup"
	// Wasabi 2.0
	heuristicTypeWasabi2ReverseLookupByTime  = "wasabi2_reverse_lookup_by_time"
	heuristicTypeWasabi2ReverseLookupByDepth = "wasabi2_reverse_lookup_by_depth"
)

func init() {
	// validHeuristicTypes contains all heuristics which are possible to receive from the frontend.
	// New heuristics must be added here
	var validHeuristicTypes = []heuristicConstructor{
		// Dash
		newOneSourceHeuristic,
		newReverseAmountHeuristic,
		newPerfectMatchHeuristic,
		newDenominationTypeHeuristic,
		newReverseLookupHeuristic,
		newForwardAmountHeuristic,
		newForwardLookupHeuristic,
		// Wasabi 2.0
		newWasabi2ReverseLookupByTimeHeuristic,
		newWasabi2ReverseLookupByDepthHeuristic,
	}

	for _, h := range validHeuristicTypes {
		ConstructorMap[h().getType()] = h
	}
}

// areClusterTypesValid checks if the given clusterTypes are valid
func areClusterTypesValid(clusterTypes []clustering.ClusterType) bool {
	if len(clusterTypes) == 0 {
		return true
	}

	// for now only one additional cluster type exists
	return len(clusterTypes) == 1 && clusterTypes[0] == clustering.TypeCustom
}

type DescriptorParameter struct {
	DefaultValue string `json:"value,omitempty"`
	Description  string `json:"description,omitempty"`
	// Type must be one of the following values: 'int', 'string'
	Type string `json:"type,omitempty"`
}

type Descriptor struct {
	Title       string `json:"title,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	// pointer so Parameter does not appear in JSON if not set
	Parameter *DescriptorParameter `json:"parameter,omitempty"`
	// controls which types of parent are allowed for a heuristic. Possible types are transaction types and heuristic types.
	// currently not enforced in the backend.
	AllowedParents []string `json:"allowedParents,omitempty"`
}

type heuristic interface {
	fmt.Stringer
	// exec executes the heuristic and returns the altered set of origin uids
	exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
		parentHeuristicUID string) ([]heuristics.HeuristicCluster, error)
	// getType returns the heuristic type
	getType() string
	// setConfig applies the provided configuration values
	setConfig(heuristics.Options) error
	// getConfig returns the configuration of the heuristic
	getConfig() heuristics.Options
	// GetDescriptor returns description of the heuristic and its expected parameter for the frontend
	GetDescriptor() Descriptor
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
func getDenominationCountsWithFilter(it heuristics.HeuristicTransaction, filterTx string) [dash.NumDenominations]int {
	var denominations []int64 //nolint:prealloc
	for _, output := range it.Outputs {
		if filterTx != "" && output.InputTransaction != filterTx {
			continue
		}
		denominations = append(denominations, output.Amount)
	}

	return dash.CountAmountDenominations(denominations)
}

// If the given transaction hash belongs to a mixing transaction then it returns the transaction itself,
// otherwise it return the input transactions of the transaction.
func getInputTransactions(ctx context.Context, c external.Database, txhash string) ([]heuristics.HeuristicTransaction, error) {
	transaction, err := db.GetTransaction(ctx, c, txhash)
	if err != nil {
		return nil, err
	}

	var inputTransactions []heuristics.HeuristicTransaction
	if transaction.Type == constants.TypeDashMixing {
		hs, err := heuristics.GetInputTransaction(ctx, c, txhash)
		if err != nil {
			return nil, err
		}
		inputTransactions = []heuristics.HeuristicTransaction{*hs}
	} else {
		hs, err := heuristics.GetInputTransactions(ctx, c, txhash)
		if err != nil {
			return nil, err
		}
		inputTransactions = hs
	}

	return inputTransactions, nil
}

// gets the counts of each denomination type
func getDenominationCounts(it heuristics.HeuristicTransaction) [dash.NumDenominations]int {
	denominations := make([]int64, len(it.Outputs))
	for i, output := range it.Outputs {
		denominations[i] = output.Amount
	}

	return dash.CountAmountDenominations(denominations)
}

type clusterDenominations struct {
	denominationIndex int
	// key: cluster id, value: number of denominations of type denominationIndex
	clusters map[heuristics.ClusterUID]int
}

// addTransactionToCluster adds the given transactions to its cluster
func addTransactionToCluster(sourceTransactionMap map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction,
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

func buildSourceAmounts(origins map[string]heuristics.HeuristicTransaction) map[heuristics.ClusterUID][dash.NumDenominations]int {
	sourceAmounts := make(map[heuristics.ClusterUID][dash.NumDenominations]int)

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
func getTimeLimitedOrigins(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	tx heuristics.HeuristicTransaction, lookBackTime time.Duration, maxDepth int, exclusions []string, c heuristics.Options) (
	origins []heuristics.HeuristicTransaction, attributionMapping map[heuristics.ClusterUID][]string, err error) {
	// do reverse lookup
	endpoints, err := g.ReverseLookup(tx.UID, lookBackTime, maxDepth, exclusions, c.ExcludeSpendingGaps)
	if err != nil {
		return nil, nil, err
	}

	// get tx details for each uid
	return heuristics.GetTransactionsWithOutputAmountAndCluster(ctx, dgraph,
		cliutil.GetMapKeys(endpoints), c.UserUID, c.ClusterTypes)
}

func isParentAHeuristic(ctx context.Context, c external.Database, parentUID string) (bool, error) {
	parentType, err := db.GetTypeByUID(ctx, c, parentUID)
	if err != nil {
		return false, err
	}

	return parentType == workspace.SelectorDType, nil
}

// Executor holds information for executing on heuristic and its children
type Executor struct {
	rootUID       string
	thisHeuristic heuristic
}

// ConstructExecutors creates executors based on heuristics
func ConstructExecutors(config heuristics.Options, userUID string, parentUID string) (executor Executor, err error) {
	constructor, ok := ConstructorMap[config.Type]
	if !ok {
		err = serror.New(errHeuristicNotValid)
		return
	}

	// copy parameters from heuristic request into newly created heuristic
	clonedHeuristic := constructor()
	c := config
	c.UserUID = userUID
	if err = clonedHeuristic.setConfig(c); err != nil {
		return
	}

	executor = Executor{
		thisHeuristic: clonedHeuristic,
		rootUID:       parentUID,
	}

	return
}

// Run starts the execution of the given heuristic executor.
func (hx Executor) Run(ctx context.Context, dgraph external.Database, g *graph.Wrapper) ([]heuristics.HeuristicCluster, error) {
	heuristicClusters, err := hx.thisHeuristic.exec(ctx, dgraph, g, hx.rootUID)
	if err != nil && !errors.Is(err, errNoOriginsAtStart) {
		return nil, err
	}

	// set DType
	for i := range heuristicClusters {
		heuristicClusters[i].SetDType()
	}

	return heuristicClusters, nil
}

// createHeuristicClusters converts the given map into HeuristicCluster's
func createHeuristicClusters(clusterMap map[heuristics.ClusterUID][]db.UIDNode,
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
