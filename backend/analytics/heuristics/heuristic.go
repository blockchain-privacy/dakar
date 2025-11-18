// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"backend/analytics/classifier/btc"
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
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

type HeuristicConstructor func() Heuristic

var ConstructorMap = make(map[string]HeuristicConstructor)

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
	// maximum duration in hours
	parameterMaxDuration = 90
	// minimum duration in hours
	parameterMinDuration = 1
	// maximum depth
	parameterMaxDepth = 5
	parameterMinDepth = 1

	parameterDescriptionLookForward = "Maximum duration (hours)"
	parameterDescriptionLookBack    = "Maximum duration (hours)"
	parameterDescriptionDepth       = "Maximum traversal depth"
	parameterTypeInt                = "int"
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
	heuristicTypeWasabi2OneSourceByTime      = "wasabi2_one_source_by_time"
	heuristicTypeWasabi2OneSourceByDepth     = "wasabi2_one_source_by_depth"
	heuristicTypeWasabi2ReverseAmount        = "wasabi2_reverse_amount"
	heuristicTypeWasabi2ForwardLookupByTime  = "wasabi2_forward_lookup_by_time"
	heuristicTypeWasabi2ForwardLookupByDepth = "wasabi2_forward_lookup_by_depth"
	// Whirlpool
	heuristicTypeWhirlpoolReverseLookupByTime  = "whirlpool_reverse_lookup_by_time"
	heuristicTypeWhirlpoolReverseLookupByDepth = "whirlpool_reverse_lookup_by_depth"
	heuristicTypeWhirlpoolOneSourceByTime      = "whirlpool_one_source_by_time"
	heuristicTypeWhirlpoolOneSourceByDepth     = "whirlpool_one_source_by_depth"
	heuristicTypeWhirlpoolReverseAmount        = "whirlpool_reverse_amount"
)

func init() {
	// validHeuristicTypes contains all heuristics which are possible to receive from the frontend.
	// New heuristics must be added here
	var validHeuristicTypes = []HeuristicConstructor{
		// Dash
		NewOneSourceHeuristic,
		NewReverseAmountHeuristic,
		NewPerfectMatchHeuristic,
		NewDenominationTypeHeuristic,
		NewReverseLookupHeuristic,
		NewForwardAmountHeuristic,
		NewForwardLookupHeuristic,
		// Wasabi 2.0
		NewWasabi2ReverseLookupByTimeHeuristic,
		NewWasabi2ReverseLookupByDepthHeuristic,
		NewWasabi2OneSourceByTimeHeuristic,
		NewWasabi2OneSourceByDepthHeuristic,
		NewWasabi2ReverseAmountHeuristic,
		NewWasabi2ForwardLookupByTimeHeuristic,
		NewWasabi2ForwardLookupByDepthHeuristic,
		// Whirlpool
		NewWhirlpoolReverseLookupByTimeHeuristic,
		NewWhirlpoolReverseLookupByDepthHeuristic,
		NewWhirlpoolOneSourceByTimeHeuristic,
		NewWhirlpoolOneSourceByDepthHeuristic,
		NewWhirlpoolReverseAmountHeuristic,
	}

	for _, h := range validHeuristicTypes {
		ConstructorMap[h().GetType()] = h
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
	// MaximumValue in case it is an integer
	MaximumValue int `json:"maximum,omitempty"`
	// MinimumValue in case it is an integer
	MinimumValue int    `json:"minimum,omitempty"`
	Description  string `json:"description,omitempty"`
	// Type must be one of the following values: 'int', 'string'
	Type string `json:"type,omitempty"`
}

type Descriptor struct {
	Title       string               `json:"title,omitempty"`
	Type        string               `json:"type,omitempty"`
	Description string               `json:"description,omitempty"`
	Category    string               `json:"category,omitempty"`
	Parameter   *DescriptorParameter `json:"parameter,omitempty"`
	// controls which types of parent are allowed for a heuristic. Possible values are transaction types and heuristic types.
	// currently not enforced in the backend.
	AllowedParents []string `json:"allowedParents,omitempty"`
}

type Heuristic interface {
	fmt.Stringer
	// Exec executes the heuristic and returns the altered set of origin uids.
	// If parentResults is unset, parentUID is used to access the results of the parent heuristic if applicable
	Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentUID string,
		parentResults []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error)
	// GetType returns the heuristic type
	GetType() string
	// SetConfig applies the provided configuration values
	SetConfig(heuristics.Options) error
	// GetConfig returns the configuration of the heuristic
	GetConfig() heuristics.Options
	// GetDescriptor returns the description of the heuristic and its expected parameter for the frontend
	GetDescriptor() Descriptor
}

// getNumberOfDashDenominations returns the number of Dash denominations. If destinationTransaction is set, it
// only counts outputs with input transactions equal to destinationTransaction. An error is returned
// if more than one type of denominations is found
func getNumberOfDashDenominations(it heuristics.HeuristicTransaction, destinationTransaction string) (nDenominations int,
	denomIndex int, err error) {
	numDenominations := getDashDenominationCountsWithFilter(it, destinationTransaction)

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

// getNumberOfWhirlpoolDenominations returns the number of Whirlpool denominations. If destinationTransaction is set, it
// only counts outputs with input transactions equal to destinationTransaction. An error is returned
// if more than one type of denominations is found
func getNumberOfWhirlpoolDenominations(it heuristics.HeuristicTransaction, destinationTransaction string) (nDenominations int,
	denomIndex int, err error) {
	numDenominations := getWhirlpoolDenominationCountsWithFilter(it, destinationTransaction)

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

// getDashDenominationCountsWithFilter gets the counts of each Dash denomination type.
// If filterTx is set, it only counts outputs with input transactions equal to filterTx.
func getDashDenominationCountsWithFilter(it heuristics.HeuristicTransaction, filterTx string) [dash.NumDenominations]int {
	var denominations []int64 //nolint:prealloc
	for _, output := range it.Outputs {
		if filterTx != "" && output.InputTransaction != filterTx {
			continue
		}
		denominations = append(denominations, output.Amount)
	}

	return dash.CountAmountDenominations(denominations)
}

// getWhirlpoolDenominationCountsWithFilter gets the counts of each Whirlpool denomination type.
// If filterTx is set, it only counts outputs with input transactions equal to filterTx.
func getWhirlpoolDenominationCountsWithFilter(it heuristics.HeuristicTransaction, filterTx string) [btc.NumWhirlpoolDenominations]int {
	var denominations []int64 //nolint:prealloc
	for _, output := range it.Outputs {
		if filterTx != "" && output.InputTransaction != filterTx {
			continue
		}
		denominations = append(denominations, output.Amount)
	}

	return btc.CountAmountWhirlpoolFuzzyDenominations(denominations, 0)
}

// If the given transaction hash belongs to a mixing transaction then it returns the transaction itself,
// otherwise it returns the input transactions of the transaction.
func getInputTransactions(ctx context.Context, c external.Database, txhash string,
	allowedTransactionType string) ([]heuristics.HeuristicTransaction, error) {
	transaction, err := db.GetTransaction(ctx, c, txhash)
	if err != nil {
		return nil, err
	}

	var inputTransactions []heuristics.HeuristicTransaction
	if constants.IsMixingTransaction(transaction.Type) {
		hs, err := heuristics.GetInputTransaction(ctx, c, txhash)
		if err != nil {
			return nil, err
		}
		inputTransactions = []heuristics.HeuristicTransaction{*hs}
	} else {
		hs, err := heuristics.GetInputTransactions(ctx, c, txhash, allowedTransactionType)
		if err != nil {
			return nil, err
		}
		inputTransactions = hs
	}

	return inputTransactions, nil
}

// gets the counts of each Dash denomination type
func getDashDenominationCounts(it heuristics.HeuristicTransaction) [dash.NumDenominations]int {
	denominations := make([]int64, len(it.Outputs))
	for i, output := range it.Outputs {
		denominations[i] = output.Amount
	}

	return dash.CountAmountDenominations(denominations)
}

// gets the counts of each Whirlpool denomination type
func getWhirlpoolDenominationCounts(it heuristics.HeuristicTransaction, minDiff int64) [btc.NumWhirlpoolDenominations]int {
	denominations := make([]int64, len(it.Outputs))
	for i, output := range it.Outputs {
		denominations[i] = output.Amount
	}

	return btc.CountAmountWhirlpoolFuzzyDenominations(denominations, minDiff)
}

type clusterDenominations struct {
	denominationIndex int
	// key: cluster id, value: number of denominations of type denominationIndex
	clusters map[heuristics.ClusterUID]int
}

// mapClusterToTransactions maps clusters to their transactions
func mapClusterToTransactions(origins []heuristics.HeuristicTransaction) map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction {
	sourceTransactionMap := map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction{}
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

// countClusterDashDenominations creates a map of clusters with the
// number of Dash denominations of the specified denomination type
func countClusterDashDenominations(origins []heuristics.HeuristicTransaction,
	denominationIndex int) (oSource clusterDenominations) {
	oSource.denominationIndex = denominationIndex
	oSource.clusters = make(map[heuristics.ClusterUID]int)
	for _, o := range origins {
		oSource.clusters[o.Cluster] += getDashDenominationCounts(o)[denominationIndex]
	}

	return
}

// countClusterWhirlpoolDenominations creates a map of clusters with the
// number of Whirlpool denominations of the specified denomination type
func countClusterWhirlpoolDenominations(origins []heuristics.HeuristicTransaction,
	denominationIndex int) (oSource clusterDenominations) {
	oSource.denominationIndex = denominationIndex
	oSource.clusters = make(map[heuristics.ClusterUID]int)
	for _, o := range origins {
		oSource.clusters[o.Cluster] += getWhirlpoolDenominationCounts(o, 100)[denominationIndex]
	}

	return
}

func buildDashSourceAmounts(origins map[string]heuristics.HeuristicTransaction) map[heuristics.ClusterUID][dash.NumDenominations]int {
	sourceAmounts := make(map[heuristics.ClusterUID][dash.NumDenominations]int)

	for _, o := range origins {
		denominationSlice := getDashDenominationCounts(o)
		for i := range denominationSlice {
			denominationSlice[i] += sourceAmounts[o.Cluster][i]
		}

		sourceAmounts[o.Cluster] = denominationSlice
	}
	return sourceAmounts
}

func buildWhirlpoolSourceAmounts(origins map[string]heuristics.HeuristicTransaction,
	minDiff int64) map[heuristics.ClusterUID][btc.NumWhirlpoolDenominations]int {
	sourceAmounts := make(map[heuristics.ClusterUID][btc.NumWhirlpoolDenominations]int)

	for _, o := range origins {
		denominationSlice := getWhirlpoolDenominationCounts(o, minDiff)
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
	transactionUID string, lookBackTime time.Duration, maxDepth int, exclusions []string,
	attributions map[string][]string, c heuristics.Options, allowedTransactionType string) (origins []heuristics.HeuristicTransaction,
	attributionMapping map[heuristics.ClusterUID][]string, err error) {
	// do reverse lookup
	endpoints, err := g.ReverseLookup(transactionUID, lookBackTime, maxDepth, exclusions, c.ExcludeSpendingGaps)
	if err != nil {
		return nil, nil, err
	}

	// get tx details for each uid
	return heuristics.GetTransactionsWithOutputAmountAndCluster(ctx, dgraph,
		cliutil.GetMapKeys(endpoints), c.UserUID, c.ClusterTypes, attributions, allowedTransactionType)
}

// getTimeLimitedDestinations returns all destinations of the given transaction.
// If lookBackTime is bigger than zero only destinations in the time range of
// tx.ts - lookBackTime will be returned.
func getTimeLimitedDestinations(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	transactionUID string, lookForwardTime time.Duration, maxDepth int, exclusions []string,
	attributions map[string][]string, c heuristics.Options, allowedTransactionType string) (origins []heuristics.HeuristicTransaction,
	attributionMapping map[heuristics.ClusterUID][]string, err error) {
	// do reverse lookup
	endpoints, err := g.ForwardLookup(transactionUID, lookForwardTime, maxDepth, exclusions, c.ExcludeSpendingGaps)
	if err != nil {
		return nil, nil, err
	}

	// get tx details for each uid
	return heuristics.GetTransactionsWithOutputAmountAndCluster(ctx, dgraph,
		cliutil.GetMapKeys(endpoints), c.UserUID, c.ClusterTypes, attributions, allowedTransactionType)
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
	thisHeuristic Heuristic
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

	if err = clonedHeuristic.SetConfig(c); err != nil {
		return
	}

	executor = Executor{
		thisHeuristic: clonedHeuristic,
		rootUID:       parentUID,
	}

	return
}

// IsConfigValid checks if the given config can be applied to its heuristic type and if the heuristic parameter is valid
func IsConfigValid(config heuristics.Options) error {
	constructor, ok := ConstructorMap[config.Type]
	if !ok {
		return serror.FromStrWithContext("invalid heuristic type", "type", config.Type)
	}

	clonedHeuristic := constructor()
	c := config

	descriptorParameter := clonedHeuristic.GetDescriptor().Parameter
	if descriptorParameter != nil && descriptorParameter.Type == parameterTypeInt {
		p, err := strconv.Atoi(config.Parameter)
		if err != nil {
			return serror.New(err)
		}

		if p < descriptorParameter.MinimumValue || p > descriptorParameter.MaximumValue {
			return serror.FromStrWithContext("invalid parameter value", "value", p)
		}
	}

	return clonedHeuristic.SetConfig(c)
}

// Run starts the execution of the given heuristic executor.
func (hx Executor) Run(ctx context.Context, dgraph external.Database, g *graph.Wrapper) ([]heuristics.HeuristicCluster, error) {
	heuristicClusters, err := hx.thisHeuristic.Exec(ctx, dgraph, g, hx.rootUID, nil)
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

// getHeuristicTransactions returns the provided transactions with a cluster UID and their output amounts
func getHeuristicTransactions(ctx context.Context, dgraph external.Database, clusters []heuristics.HeuristicCluster,
	allowedTransactionType string) ([]heuristics.HeuristicTransaction, error) {
	var txUIDs []string
	uidToCluster := map[string]heuristics.ClusterUID{}
	for i, cluster := range clusters {
		cUID := heuristics.ClusterUID(strconv.Itoa(i))
		for _, result := range cluster.Results {
			txUIDs = append(txUIDs, result.UID)
			uidToCluster[result.UID] = cUID
		}
	}

	transactions, err := heuristics.GetHeuristicTransactionsOutputs(ctx, dgraph, txUIDs, allowedTransactionType)
	if err != nil {
		return nil, err
	}

	for i, t := range transactions {
		transactions[i].Cluster = uidToCluster[t.UID]
	}

	return transactions, nil
}
