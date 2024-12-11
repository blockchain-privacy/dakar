package heuristics

import (
	"backend/analytics/graph"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/exclusion"
	"backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"strconv"
	"time"
)

// oneSourceHeuristic - see exec for description
type oneSourceHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookBackTime         time.Duration
	c                    heuristics.Options
}

func newOneSourceHeuristic() heuristic {
	return &oneSourceHeuristic{heuristicType: heuristicTypeOneSource}
}

func (h *oneSourceHeuristic) getType() string {
	return h.heuristicType
}

func (h *oneSourceHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *oneSourceHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	duration, err := strconv.ParseInt(c.Parameter, 10, 64)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.lookBackTime = time.Duration(duration) * time.Hour
	h.parameterDescription = strconv.FormatInt(duration, 10)
	h.c = c

	return nil
}

func (h *oneSourceHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *oneSourceHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *oneSourceHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "One Source",
		Type:        h.heuristicType,
		Category:    heuristicCategoryReverse,
		Description: "Filters by time, direct input transaction amount filter and omni sources",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			Description:  "Look back time in hours",
			Type:         "int",
		},
		AllowedParents: constants.TransactionTypesDash,
	}
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
func (h *oneSourceHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentHeuristicUID string) (
	[]heuristics.HeuristicCluster, error) {
	if h.lookBackTime == 0 {
		return nil, nil
	}

	parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
	if err != nil {
		return nil, err
	}
	// heuristic is only allowed to be connected to a transaction
	if parentHeuristicSet {
		return nil, serror.New(errHeuristicNotValid)
	}

	// Get all transactions which are connected via the inputs of the destination
	// transaction specified by txHash.
	inputTransactions, err := getInputTransactions(ctx, dgraph, h.c.TransactionHash)
	if err != nil {
		return nil, err
	}

	if len(inputTransactions) == 0 {
		// nothing to do
		return nil, nil
	}

	// contains all time limited origins
	var allTimeLimitedOrigins []heuristics.HeuristicTransaction
	// contains all time limited origins per input transaction
	var allTxAndOrigins []txAndOrigins //nolint:prealloc
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	attributionMap := make(map[heuristics.ClusterUID][]string)

	var exclusions []string
	if h.c.ExcludeAddresses {
		exclusions, err = exclusion.GetAddressExclusionUIDs(ctx, dgraph, h.c.UserUID)
		if err != nil {
			return nil, err
		}
	}

	for _, it := range inputTransactions {
		timeLimitedOrigins, usedAttributions, err := getTimeLimitedOrigins(ctx, dgraph, g, it,
			h.lookBackTime, exclusions, h.c)
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

	// mRemovableSources holds all sources which can be removed,
	// due to not being able to fund all connected input transactions
	mRemovableSources := make(map[heuristics.ClusterUID]bool)
	// sources holds all sources found in all input transactions
	sources := make(map[heuristics.ClusterUID]bool)
	// save origins in global cluster->origin map
	sourceTransactionMap := addTransactionToCluster(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction{}, allTimeLimitedOrigins)
	// for each input transaction to the destination transaction,
	// inputSources holds one map with all its occurring sources
	var inputSources []map[heuristics.ClusterUID]bool //nolint:prealloc
	for _, t := range allTxAndOrigins {
		// get input denominations
		nDenominations, denominationIndex, getErr := getNumberOfDenominations(t.inputTransaction, h.c.TransactionHash)
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

	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for k := range remainingOrigins {
		v, ok := allOriginMap[k]
		if !ok {
			return nil, serror.FromStr("could not find origin in all origin map")
		}

		resultClusters[v.Cluster] = append(resultClusters[v.Cluster], db.UIDNode{UID: k})
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
