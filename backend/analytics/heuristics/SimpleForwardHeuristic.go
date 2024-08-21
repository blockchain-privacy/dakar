package heuristics

import (
	"backend/analytics/graph"
	"backend/db"
	"backend/db/analytics/exclusion"
	"backend/db/analytics/heuristics"
	"backend/external"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"strconv"
	"time"
)

// simpleForwardHeuristic - see exec for description
type simpleForwardHeuristic struct {
	heuristicType        string
	parameterDescription string
	c                    heuristics.Options
	lookForwardTime      time.Duration
}

func newSimpleForwardHeuristic() heuristic {
	return &simpleForwardHeuristic{heuristicType: "simple_forward_lookup"}
}

func (h *simpleForwardHeuristic) getType() string {
	return h.heuristicType
}

func (h *simpleForwardHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *simpleForwardHeuristic) setConfig(c heuristics.Options) error {
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

func (h *simpleForwardHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *simpleForwardHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *simpleForwardHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Simple Forward Lookup",
		Type:     h.heuristicType,
		Category: heuristicCategoryForward,
		Description: "Performs a forward lookup for the " +
			"provided origin transaction. " +
			"If this heuristic is placed at the root level " +
			"a reverse lookup with the same " +
			"time as the forward lookup will be performed.",
		Parameter: &struct {
			DefaultValue string `json:"value,omitempty"`
			Description  string `json:"description,omitempty"`
			Type         string `json:"type,omitempty"`
		}{
			DefaultValue: "48",
			Description:  "Look forward time in hours",
			Type:         "int",
		},
	}
}

// forwardLookupHeuristic applies the following heuristics:
//   - parent == transaction: By traversing the mixing graph forward limited by time,
//     find all destination transactions connected to this transaction.
//   - parent == heuristic: By traversing the mixing graph forward limited by time,
//     find all destination transactions connected to the results of the parent heuristic.
func (h *simpleForwardHeuristic) exec(dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	if h.lookForwardTime == 0 {
		return nil, nil
	}
	var parentResults []heuristics.HeuristicTransaction
	// resultAttributionMap maps a clusterUID to a slice of attribution UIDs
	var resultAttributionMap map[heuristics.ClusterUID][]string

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
	if err != nil {
		return nil, err
	}

	if parentHeuristicSet {
		// get transactions from parent heuristic
		var err error
		parentResults, resultAttributionMap, err = heuristics.GetHeuristicTransactions(dgraph, parentHeuristicUID)
		if err != nil {
			return nil, err
		}
	} else {
		uid, err := db.GetTransactionUID(ctx, dgraph, h.c.TransactionHash)
		if err != nil {
			return nil, err
		}

		// get tx details for each uid
		parentResults, resultAttributionMap, err = heuristics.GetTransactionsWithOutputAmountAndCluster(dgraph,
			[]string{uid}, h.c.UserUID, h.c.ClusterTypes)
		if err != nil {
			return nil, err
		}

		if len(parentResults) > 1 {
			return nil, serror.FromFormat("received wrong amount of transactions: %d", len(parentResults))
		}
	}

	if len(parentResults) == 0 {
		return nil, serror.New(errNoOriginsAtStart)
	}

	var exclusions []string
	if h.c.ExcludeAddresses {
		var err error
		exclusions, err = exclusion.GetAddressExclusionUIDs(dgraph, h.c.UserUID)
		if err != nil {
			return nil, err
		}
	}

	resultClusters := make(map[heuristics.ClusterUID][]db.UIDNode, len(parentResults))

	for _, parentResult := range parentResults {
		uidMap, err := getOriginDestinationTimeLimited(g, []string{parentResult.UID}, h.lookForwardTime,
			exclusions, h.c.ExcludeSpendingGaps)
		if err != nil {
			return nil, err
		}

		result := make([]db.UIDNode, 0, len(uidMap))
		for k := range uidMap {
			result = append(result, db.UIDNode{UID: k})
		}

		resultClusters[parentResult.Cluster] = result
	}

	return createHeuristicClusters(resultClusters, resultAttributionMap), nil
}
