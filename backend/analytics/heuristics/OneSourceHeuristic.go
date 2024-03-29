package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db/analytics/clustering"
	"backend/db/analytics/exclusion"
	"backend/db/analytics/heuristics"
	"backend/external"
	"fmt"
	"strconv"
	"time"
)

// oneSourceHeuristic - see exec for description
type oneSourceHeuristic struct {
	heuristicType        string
	parameterDescription string
	userUID              string
	workspaceUID         string
	excludeAddresses     bool
	excludeSpendingGaps  bool
	lookBackTime         time.Duration
	clusterTypes         []clustering.ClusterType
}

// newOneSourceHeuristic constructs an oneSourceHeuristic. hoursToLookBack in hours
func newOneSourceHeuristic(hoursToLookBack uint32, clusterTypes []clustering.ClusterType) *oneSourceHeuristic {
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	return &oneSourceHeuristic{
		heuristicType:        "one_source",
		lookBackTime:         lBackTime,
		parameterDescription: strconv.FormatUint(uint64(hoursToLookBack), 10),
		clusterTypes:         clusterTypes,
	}
}

func (h oneSourceHeuristic) getType() string {
	return h.heuristicType
}

func (h oneSourceHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h oneSourceHeuristic) hasParameter() bool {
	return true
}

func (h *oneSourceHeuristic) setParameter(p string) error {
	hoursToLookBack, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return cliutil.NewStackError(err)
	}
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	h.lookBackTime = lBackTime
	h.parameterDescription = strconv.FormatUint(hoursToLookBack, 10)
	return nil
}

// setClusterTypes sets additional cluster types, which are used to execute the heuristic.
// Multi-input clusters are always used to execute the heuristic,
// any cluster type set here will be used additionally. If at least one cluster type is set,
// then the consolidation of the multi-input clusters and the additional clusters will be used.
func (h *oneSourceHeuristic) setClusterTypes(clusterTypes []clustering.ClusterType) error {
	if !areClusterTypesValid(clusterTypes) {
		return cliutil.NewStackError(errInvalidClusterTypes)
	}

	h.clusterTypes = clusterTypes
	return nil
}

// getClusterTypes returns the cluster types this heuristic uses to cluster addresses
func (h *oneSourceHeuristic) getClusterTypes() []clustering.ClusterType {
	return h.clusterTypes
}

// setExcludeAddresses sets whether certain addresses should be excluded from the lookups
func (h *oneSourceHeuristic) setExcludeAddresses(excludeAddresses bool) {
	h.excludeAddresses = excludeAddresses
}

// getExcludeAddresses returns whether certain addresses should be excluded from the lookups
func (h *oneSourceHeuristic) getExcludeAddresses() bool {
	return h.excludeAddresses
}

// setExcludeSpendingGaps sets whether mixing outputs with a spending gap should be traversed
func (h *oneSourceHeuristic) setExcludeSpendingGaps(excludeSpendingGaps bool) {
	h.excludeSpendingGaps = excludeSpendingGaps
}

// getExcludeSpendingGaps returns whether mixing outputs with a spending gap should be traversed
func (h *oneSourceHeuristic) getExcludeSpendingGaps() bool {
	return h.excludeSpendingGaps
}

// setUserUID sets the UID of the user who created this heuristic
func (h *oneSourceHeuristic) setUserUID(uid string) {
	h.userUID = uid
}

// setWorkspaceUID sets the UID of the workspace to which this heuristic belongs
func (h *oneSourceHeuristic) setWorkspaceUID(uid string) {
	h.workspaceUID = uid
}

// getWorkspaceUID returns the workspace UID this heuristic belongs to
func (h *oneSourceHeuristic) getWorkspaceUID() string {
	return h.workspaceUID
}

func (h oneSourceHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h oneSourceHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "One Source",
		Type:        h.heuristicType,
		Category:    heuristicCategoryReverse,
		Description: "Filters by time, direct input transaction amount filter and omni sources",
		Parameter: &struct {
			DefaultValue string `json:"value,omitempty"`
			Description  string `json:"description,omitempty"`
			Type         string `json:"type,omitempty"`
		}{
			DefaultValue: "48",
			Description:  "Look back time in hours",
			Type:         "int",
		},
	}
}

func (h oneSourceHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

type txAndOrigins struct {
	inputTransaction heuristics.HeuristicTransaction
	origins          []heuristics.HeuristicTransaction
}

// oneSourceHeuristic applies the following heuristics:
//   - filter all origins, which are not created in the time span defined by lookBackTime
//   - filter all origins of sources, which do not have enough denominations to fund all of their respective
//     outputs of input transaction which are used as inputs in the destination transaction
//   - filter all origins of sources, which do not occur in all sets of input transaction origins
//
// This heuristic does not use the results from its parent heuristic
func (h oneSourceHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string, _ string) (
	[]heuristics.HeuristicCluster, error) {
	// Get all transactions which are connected via the inputs of the destination
	// transaction specified by txHash. These transactions are called >>input transactions<<.
	inputTransactions, err := heuristics.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, err
	}

	// sources holds all sources found in all input transactions
	sources := make(map[heuristics.ClusterUID]bool)
	// mRemovableSources holds all sources which can be removed,
	// due to not being able to fund all connected input transactions
	mRemovableSources := make(map[heuristics.ClusterUID]bool)
	// maps a cluster to its origin transactions
	sourceTransactionMap := make(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction)
	// for each input transaction to the destination transaction,
	// inputSources holds one map with all its occurring sources
	var inputSources []map[heuristics.ClusterUID]bool //nolint:prealloc
	// contains all time limited origins
	var allTimeLimitedOrigins []heuristics.HeuristicTransaction
	// contains all time limited origins per input transaction
	var allTxAndOrigins []txAndOrigins //nolint:prealloc
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	attributionMap := make(map[heuristics.ClusterUID][]string)

	var exclusions []string

	if h.excludeAddresses {
		exclusions, err = exclusion.GetAddressExclusionUIDs(dgraph, h.userUID)
		if err != nil {
			return nil, err
		}
	}

	for _, it := range inputTransactions {
		timeLimitedOrigins, usedAttributions, err := getTimeLimitedOrigins(dgraph, g, it, h.lookBackTime, h.userUID,
			h.clusterTypes, exclusions, h.excludeSpendingGaps)
		if err != nil {
			return nil, err
		}

		if len(timeLimitedOrigins) == 0 {
			continue
		}

		// merge the attribution maps
		for id, attributions := range usedAttributions {
			attributionMap[id] = attributions
		}

		allTimeLimitedOrigins = append(allTimeLimitedOrigins, timeLimitedOrigins...)

		allTxAndOrigins = append(allTxAndOrigins, txAndOrigins{inputTransaction: it, origins: timeLimitedOrigins})
	}

	// save origins in global cluster->origin map
	sourceTransactionMap = addOriginsToMap(sourceTransactionMap, allTimeLimitedOrigins)

	for _, t := range allTxAndOrigins {
		// get input denominations
		nDenominations, denominationIndex, getErr := getNumberOfDenominations(t.inputTransaction, txHash)
		if getErr != nil {
			return nil, getErr
		}

		oSource := countClusterDenominations(t.origins, denominationIndex)

		// add element inputSources and set index of current element
		inputSources = append(inputSources, make(map[heuristics.ClusterUID]bool))
		iSSIndex := len(inputSources) - 1

		// Loop through all clusters of the current input transaction and mark
		// the clusters which do not have enough denominations to fund all outputs of
		// the input transaction which are used as inputs in the destination transaction
		for k, v := range oSource.clusters {
			sources[k] = true
			inputSources[iSSIndex][k] = true
			if v < nDenominations {
				mRemovableSources[k] = true
			}
		}
	}

	// Remove sources which do not have enough denominations to
	// fund all input transaction to which they are connected
	for k := range mRemovableSources {
		delete(sources, k)
	}

	// save all addresses (sources) which are not part of all input transactions
	var omniSources []heuristics.ClusterUID

	for k := range sources {
		found := true
		// check for each input transaction if the source k exists. If not set the flag to false
		for _, inputTransactionSource := range inputSources {
			if !inputTransactionSource[k] {
				found = false
				break
			}
		}

		if found {
			omniSources = append(omniSources, k)
		}
	}

	remainingOrigins := make(map[string]bool)
	// collect all transactions from omni sources
	for _, omniSource := range omniSources {
		omniOrigins := sourceTransactionMap[omniSource]
		for _, o := range omniOrigins {
			remainingOrigins[o.UID] = true
		}
	}

	allOriginMap := make(map[string]heuristics.HeuristicTransaction)
	for _, tx := range allTimeLimitedOrigins {
		allOriginMap[tx.UID] = tx
	}

	resultClusters := make(map[heuristics.ClusterUID][]heuristics.HeuristicResult)
	for k := range remainingOrigins {
		v, ok := allOriginMap[k]
		if !ok {
			return nil, cliutil.NewStackErrorStr("could not find origin in all origin map")
		}

		resultClusters[v.Cluster] = append(resultClusters[v.Cluster], heuristics.HeuristicResult{
			Origin: heuristics.DummyNode{UID: k},
		})
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
