// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"backend/analytics/graph"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/db/analytics/exclusion"
	"backend/db/heuristics"
	"backend/external"
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// whirlpoolOneSourceByTimeHeuristic - see exec for description
type whirlpoolOneSourceByTimeHeuristic struct {
	heuristicType string
	lookBackTime  time.Duration
	c             HeuristicOptions
}

func NewWhirlpoolOneSourceByTimeHeuristic() Heuristic {
	return &whirlpoolOneSourceByTimeHeuristic{heuristicType: heuristicTypeWhirlpoolOneSourceByTime}
}

func (h *whirlpoolOneSourceByTimeHeuristic) GetType() string {
	return h.heuristicType
}

func (h *whirlpoolOneSourceByTimeHeuristic) SetConfig(c HeuristicOptions) error {
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
	h.c = c

	return nil
}

func (h *whirlpoolOneSourceByTimeHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *whirlpoolOneSourceByTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *whirlpoolOneSourceByTimeHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "One source by time",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Destination transactions spend outputs of their connected input mixing transactions. " +
			"Each input mixing transaction is connected to a mixing sub graph. Starting from each connected " +
			"mixing transaction, this heuristic traverses the transaction graph backwards until the given depth " +
			"is reached and excludes all clusters which can't fund every mixing sub graph (due to lack of funds " +
			"or due to having no connection to them).",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			MinimumValue: parameterMinDuration,
			MaximumValue: parameterMaxDuration,
			Description:  parameterDescriptionLookBack,
			Type:         parameterTypeInt,
		},
		AllowedParents: constants.TransactionTypesWhirlpool,
	}
}

// Exec of the whirlpoolOneSourceByTimeHeuristic applies the following heuristics:
//   - filter all origins, which are not created in the time span defined by lookBackTime
//   - filter all origins of clusters, which do not have enough denominations to fund all of their respective
//     outputs of input transaction which are used as inputs in the destination transaction
//   - filter all origins of clusters, which do not occur in all sets of input transaction origins
func (h *whirlpoolOneSourceByTimeHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return whirlpoolOnceSource(ctx, dgraph, g, parentUID, h.lookBackTime, 0, h.c)
}

func whirlpoolOnceSource(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentHeuristicUID string,
	lookBackTime time.Duration, depth int, options HeuristicOptions) ([]heuristics.HeuristicCluster, error) {
	if lookBackTime == 0 {
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
	inputTransactions, err := getInputTransactions(ctx, dgraph, options.TransactionHash, constants.TypeWhirlpoolMixing)
	if err != nil {
		return nil, err
	}

	if len(inputTransactions) == 0 {
		// nothing to do
		return nil, nil
	}

	var exclusions []string
	if options.ExcludeAddresses {
		exclusions, err = exclusion.GetAddressExclusionUIDs(ctx, dgraph, options.UserUID)
		if err != nil {
			return nil, err
		}
	}

	attributions, err := attribution.GetAttributionsPerCluster(ctx, dgraph, options.UserUID, options.ClusterTypes)
	if err != nil {
		return nil, err
	}

	// contains all time limited origins
	var allTimeLimitedOrigins []heuristics.HeuristicTransaction
	allTimeLimitedOriginsMap := map[string]heuristics.HeuristicTransaction{}
	// contains all time limited origins per input transaction
	var allTxAndOrigins []txAndOrigins
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	attributionMap := make(map[heuristics.ClusterUID][]string)
	for _, it := range inputTransactions {
		timeLimitedOrigins, usedAttributions, err := getTimeLimitedOrigins(ctx, dgraph, g, it.UID,
			lookBackTime, depth, exclusions, attributions, options, constants.TypeWhirlpoolMixing)
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
		for _, t := range timeLimitedOrigins {
			allTimeLimitedOriginsMap[t.UID] = t
		}

		allTxAndOrigins = append(allTxAndOrigins, txAndOrigins{inputTransaction: it, origins: timeLimitedOrigins})
	}

	if err = mergeOriginPeelchainCluster(g, allTimeLimitedOriginsMap); err != nil {
		return nil, err
	}

	// apply updated clusters from peelchains
	for i := range allTimeLimitedOrigins {
		allTimeLimitedOrigins[i].Cluster = allTimeLimitedOriginsMap[allTimeLimitedOrigins[i].UID].Cluster
	}

	// apply updated clusters from peelchains
	for i := range allTxAndOrigins {
		for y := range allTxAndOrigins[i].origins {
			allTxAndOrigins[i].origins[y].Cluster = allTimeLimitedOriginsMap[allTxAndOrigins[i].origins[y].UID].Cluster
		}
	}

	// mRemovableClusters holds all clusters which can be removed,
	// due to not being able to fund all connected input transactions
	mRemovableClusters := make(map[heuristics.ClusterUID]bool)
	// clusters holds all clusters found in all input transactions
	clusters := make(map[heuristics.ClusterUID]bool)
	// for each input transaction to the destination transaction,
	// inputClusters holds one map with all its occurring clusters
	var inputClusters []map[heuristics.ClusterUID]bool
	for _, t := range allTxAndOrigins {
		// get input denominations
		nDenominations, denominationIndex, getErr := getNumberOfWhirlpoolDenominations(t.inputTransaction, options.TransactionHash)
		if getErr != nil {
			return nil, getErr
		}

		oSource := countClusterWhirlpoolDenominations(t.origins, denominationIndex)

		// add element inputSources and set index of current element
		inputClusters = append(inputClusters, make(map[heuristics.ClusterUID]bool))
		icIndex := len(inputClusters) - 1

		// Loop through all clusters of the current input transaction and mark
		// the clusters which do not have enough denominations to fund all outputs of
		// the input transaction which are used as inputs in the destination transaction
		for k, v := range oSource.clusters {
			clusters[k] = true
			inputClusters[icIndex][k] = true
			if v < nDenominations {
				mRemovableClusters[k] = true
			}
		}
	}

	// Remove sources which do not have enough denominations to
	// fund all input transaction to which they are connected
	for k := range mRemovableClusters {
		delete(clusters, k)
	}

	clusterTransactionMap := mapClusterToTransactions(allTimeLimitedOrigins)
	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode)
	for k := range clusters {
		doesClusterOccurInAllMixingSubgraphs := true
		// check for each input transaction if cluster k exists. If not set the flag to false
		for _, inputTransactionSource := range inputClusters {
			if !inputTransactionSource[k] {
				doesClusterOccurInAllMixingSubgraphs = false
				break
			}
		}

		if doesClusterOccurInAllMixingSubgraphs {
			omniOrigins := clusterTransactionMap[k]
			for _, o := range omniOrigins {
				resultClusters[o.Cluster] = append(resultClusters[o.Cluster], db.UIDNode{UID: o.UID})
			}
		}
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
