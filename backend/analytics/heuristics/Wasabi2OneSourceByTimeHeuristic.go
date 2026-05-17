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

// wasabi2OneSourceByTimeHeuristic - see exec for description
type wasabi2OneSourceByTimeHeuristic struct {
	heuristicType string
	lookBackTime  time.Duration
	c             HeuristicOptions
}

func NewWasabi2OneSourceByTimeHeuristic() Heuristic {
	return &wasabi2OneSourceByTimeHeuristic{heuristicType: heuristicTypeWasabi2OneSourceByTime}
}

func (h *wasabi2OneSourceByTimeHeuristic) GetType() string {
	return h.heuristicType
}

func (h *wasabi2OneSourceByTimeHeuristic) SetConfig(c HeuristicOptions) error {
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

func (h *wasabi2OneSourceByTimeHeuristic) GetConfig() HeuristicOptions {
	return h.c
}

func (h *wasabi2OneSourceByTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Parameter: %v", h.heuristicType, h.c)
}

func (h *wasabi2OneSourceByTimeHeuristic) GetDescriptor() Descriptor {
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
		AllowedParents: constants.TransactionTypesWasabi2,
	}
}

// Exec of the wasabi2OneSourceByTimeHeuristic applies the following heuristics:
//   - filter all origins, which are not created in the time span defined by lookBackTime
//   - filter all origins of clusters, which do not have enough denominations to fund all of their respective
//     outputs of input transaction which are used as inputs in the destination transaction
//   - filter all origins of clusters, which do not occur in all sets of input transaction origins
func (h *wasabi2OneSourceByTimeHeuristic) Exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentUID string, _ []heuristics.HeuristicCluster) ([]heuristics.HeuristicCluster, error) {
	return wasabi2OneSource(ctx, dgraph, g, parentUID, h.lookBackTime, 0, h.c)
}

func wasabi2OneSource(ctx context.Context, dgraph external.Database, g *graph.Wrapper, parentHeuristicUID string,
	lookBackTime time.Duration, depth int, options HeuristicOptions) ([]heuristics.HeuristicCluster, error) {
	if lookBackTime == 0 && depth == 0 {
		return nil, nil
	}

	if lookBackTime != 0 && depth != 0 {
		return nil, serror.FromStr("both depth and look back time are set")
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
	inputTransactions, err := getInputTransactions(ctx, dgraph, options.TransactionHash, constants.TypeWasabi2Mixing)
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
	// contains all time limited origins per input transaction
	var allTxAndOrigins []txAndOrigins
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	attributionMap := make(map[heuristics.ClusterUID][]string)
	for _, it := range inputTransactions {
		timeLimitedOrigins, usedAttributions, err := getTimeLimitedOrigins(ctx, dgraph, g, it.UID,
			lookBackTime, depth, exclusions, attributions, options, constants.TypeWasabi2Mixing)
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

	// mRemovableClusters holds all clusters which can be removed,
	// due to not being able to fund all connected input transactions
	mRemovableClusters := make(map[heuristics.ClusterUID]bool)
	// clusters holds all clusters found in all input transactions
	clusters := make(map[heuristics.ClusterUID]bool)
	// for each input transaction to the destination transaction,
	// inputClusters holds one map with all its occurring clusters
	var inputClusters []map[heuristics.ClusterUID]bool

	for _, t := range allTxAndOrigins {
		var inputTxOutputSum int64
		for _, output := range t.inputTransaction.Outputs {
			// skip if output of was not used by destination transaction
			if output.InputTransaction != options.TransactionHash {
				continue
			}
			inputTxOutputSum += output.Amount
		}

		// add element inputClusters and set index of current element
		inputClusters = append(inputClusters, make(map[heuristics.ClusterUID]bool))
		icIndex := len(inputClusters) - 1

		// per input transaction, map clusters to their origins and
		// mark them for removal if they don't have enough funds for the input transaction
		clusterTransactionMap := mapClusterToTransactions(t.origins)
		for clusterUID, clusterOrigins := range clusterTransactionMap {
			var clusterOutputAmount int64
			for _, origin := range clusterOrigins {
				for _, output := range origin.Outputs {
					clusterOutputAmount += output.Amount
				}
			}

			clusters[clusterUID] = true
			inputClusters[icIndex][clusterUID] = true

			if clusterOutputAmount < inputTxOutputSum {
				mRemovableClusters[clusterUID] = true
			}
		}
	}

	// Remove clusters which do not have enough denominations to
	// fund all input transaction to which they are connected
	for k := range mRemovableClusters {
		delete(clusters, k)
	}

	// create cluster->origin map
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
