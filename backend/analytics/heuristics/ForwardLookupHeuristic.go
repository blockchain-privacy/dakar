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

// forwardHeuristic - see exec for description
type forwardHeuristic struct {
	heuristicType        string
	parameterDescription string
	c                    heuristics.Options
	lookForwardTime      time.Duration
}

func newForwardLookupHeuristic() heuristic {
	return &forwardHeuristic{heuristicType: heuristicTypeForwardLookup}
}

func (h *forwardHeuristic) getType() string {
	return h.heuristicType
}

func (h *forwardHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *forwardHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	hoursToLookForward, err := strconv.ParseInt(c.Parameter, 10, 64)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.parameterDescription = strconv.FormatInt(hoursToLookForward, 10)
	h.c = c

	return nil
}

func (h *forwardHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *forwardHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *forwardHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Forward Lookup",
		Type:        h.heuristicType,
		Category:    heuristicCategoryForward,
		Description: "Traverses the transaction graph forward and collects destination transactions.",
		Parameter: &DescriptorParameter{
			DefaultValue: "48",
			Description:  "Look forward time in hours",
			Type:         "int",
		},
		AllowedParents: constants.TransactionTypesDash,
	}
}

// forwardLookupHeuristic applies the following heuristics:
//   - parent == transaction: By traversing the mixing graph forward limited by time,
//     find all destination transactions connected to this transaction.
//   - parent == heuristic: None, this is not allowed.
func (h *forwardHeuristic) exec(ctx context.Context, dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	if h.lookForwardTime == 0 {
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

	uid, err := db.GetTransactionUID(ctx, dgraph, h.c.TransactionHash)
	if err != nil {
		return nil, err
	}

	results, resultAttributionMap, err := heuristics.GetTransactionsWithOutputAmountAndCluster(ctx, dgraph,
		[]string{uid}, h.c.UserUID, h.c.ClusterTypes)
	if err != nil {
		return nil, err
	}

	if len(results) > 1 {
		return nil, serror.FromStr("received more than one transaction")
	}

	if len(results) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
	}

	var exclusions []string
	if h.c.ExcludeAddresses {
		var err error
		exclusions, err = exclusion.GetAddressExclusionUIDs(ctx, dgraph, h.c.UserUID)
		if err != nil {
			return nil, err
		}
	}

	uidMap, err := g.ForwardLookupByTime(uid, h.lookForwardTime, exclusions, h.c.ExcludeSpendingGaps)
	if err != nil {
		return nil, err
	}

	result := make([]db.UIDNode, 0, len(uidMap))
	for k := range uidMap {
		result = append(result, db.UIDNode{UID: k})
	}

	return createHeuristicClusters(map[heuristics.ClusterUID][]db.UIDNode{results[0].Cluster: result}, resultAttributionMap), nil
}
