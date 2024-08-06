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

// forwardLookupHeuristic - see exec for description
type forwardLookupHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookForwardTime      time.Duration
	c                    heuristics.Options
}

func newForwardLookupHeuristic() heuristic {
	return &forwardLookupHeuristic{heuristicType: "forward_lookup"}
}

func (h *forwardLookupHeuristic) getType() string {
	return h.heuristicType
}

func (h *forwardLookupHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *forwardLookupHeuristic) setConfig(c heuristics.Options) error {
	if c.TransactionHash == "" {
		return serror.FromStrWithContext("transaction hash not set", "config", c)
	}

	hoursToLookForward, err := strconv.ParseUint(c.Parameter, 10, 32)
	if err != nil {
		return serror.New(err)
	}

	if !areClusterTypesValid(c.ClusterTypes) {
		return serror.New(errInvalidClusterTypes)
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.parameterDescription = strconv.FormatUint(hoursToLookForward, 10)
	h.c = c

	return nil
}

func (h *forwardLookupHeuristic) getConfig() heuristics.Options {
	return h.c
}

func (h *forwardLookupHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %v", h.heuristicType, h.c)
}

func (h *forwardLookupHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Forward Lookup",
		Type:     h.heuristicType,
		Category: heuristicCategoryForward,
		Description: "Performs a forward lookup for each origin " +
			"transaction of the parent heuristic. " +
			"If this heuristic " +
			"is placed at the root level a reverse lookup with the same " +
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
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h *forwardLookupHeuristic) exec(dgraph external.Database, g *graph.Wrapper,
	parentHeuristicUID string) ([]heuristics.HeuristicCluster, error) {
	var results []heuristics.HeuristicTransaction
	// resultAttributionMap maps a clusterUID to a slice of attribution UIDs
	var resultAttributionMap map[heuristics.ClusterUID][]string

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	{ // separate enclosure so the results slice can be garbage collected
		parentHeuristicSet, err := isParentAHeuristic(ctx, dgraph, parentHeuristicUID)
		if err != nil {
			return nil, err
		}

		if parentHeuristicSet {
			// get origins from parent heuristic
			var err error
			results, resultAttributionMap, err = heuristics.GetHeuristicTransactions(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			results, resultAttributionMap, err = getDestinationTxOriginsTimeLimited(ctx, dgraph, g,
				h.c.TransactionHash, h.lookForwardTime, h.c)
			if err != nil {
				return nil, err
			}
		}

		if len(results) == 0 {
			return nil, serror.New(errNoOriginsAtStart)
		}
	}

	var exclusions []string
	if h.c.ExcludeAddresses {
		var err error
		exclusions, err = exclusion.GetAddressExclusionUIDs(dgraph, h.c.UserUID)
		if err != nil {
			return nil, err
		}
	}

	resultClusters := make(map[heuristics.ClusterUID][]heuristics.HeuristicResult)
	for _, o := range results {
		uidMap, err := getOriginDestinationTimeLimited(g, []string{o.UID}, h.lookForwardTime,
			exclusions, h.c.ExcludeSpendingGaps)
		if err != nil {
			return nil, err
		}

		result := heuristics.HeuristicResult{
			Origin: db.UIDNode{UID: o.UID},
		}

		for k := range uidMap {
			result.Destinations = append(result.Destinations, db.UIDNode{UID: k})
		}

		resultClusters[o.Cluster] = append(resultClusters[o.Cluster], result)
	}

	return createHeuristicClusters(resultClusters, resultAttributionMap), nil
}
