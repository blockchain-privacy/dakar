package heuristics

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	"backend/db/analytics/clustering"
	"backend/db/analytics/exclusion"
	"backend/db/analytics/heuristics"
	"backend/external"
	"strconv"
	"time"

	"fmt"
)

// forwardAmountHeuristic - see exec for description
type forwardAmountHeuristic struct {
	heuristicType        string
	parameterDescription string
	userUID              string
	lookForwardTime      time.Duration
	excludeAddresses     bool
	clusterTypes         []clustering.ClusterType
}

// newForwardAmountHeuristic constructs an forwardAmountHeuristic. hoursToLookForward in hours.
func newForwardAmountHeuristic(hoursToLookForward uint32, clusterTypes []clustering.ClusterType) *forwardAmountHeuristic {
	lForwardTime := time.Duration(hoursToLookForward) * time.Hour
	return &forwardAmountHeuristic{
		heuristicType:        "forward_amount",
		lookForwardTime:      lForwardTime,
		parameterDescription: lForwardTime.String(),
		clusterTypes:         clusterTypes,
	}
}

func (h *forwardAmountHeuristic) getType() string {
	return h.heuristicType
}

func (h *forwardAmountHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h *forwardAmountHeuristic) hasParameter() bool {
	return true
}

func (h *forwardAmountHeuristic) setParameter(p string) error {
	hoursToLookForward, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.parameterDescription = strconv.FormatUint(hoursToLookForward, 10)
	return nil
}

// setClusterTypes sets additional cluster types, which are used to execute the heuristic.
// Multi-input clusters are always used to execute the heuristic,
// any cluster type set here will be used additionally. If at least one cluster type is set,
// then the consolidation of the multi-input clusters and the additional clusters will be used.
func (h *forwardAmountHeuristic) setClusterTypes(clusterTypes []clustering.ClusterType) error {
	if !areClusterTypesValid(clusterTypes) {
		return errorInvalidClusterTypes
	}

	h.clusterTypes = clusterTypes
	return nil
}

// getClusterTypes returns the cluster types this heuristic uses to cluster addresses
func (h *forwardAmountHeuristic) getClusterTypes() []clustering.ClusterType {
	return h.clusterTypes
}

// setExcludeAddresses sets whether certain addresses should be excluded from the lookups
func (h *forwardAmountHeuristic) setExcludeAddresses(excludeAddresses bool) {
	h.excludeAddresses = excludeAddresses
}

// getExcludeAddresses returns whether certain addresses should be excluded from the lookups
func (h *forwardAmountHeuristic) getExcludeAddresses() bool {
	return h.excludeAddresses
}

// setUserUID sets the UID of the user who created this heuristic
func (h *forwardAmountHeuristic) setUserUID(uid string) {
	h.userUID = uid
}

func (h *forwardAmountHeuristic) String() string {
	return fmt.Sprintf("Type: %s", h.heuristicType)
}

func (h *forwardAmountHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Forward Amount",
		Type:     h.heuristicType,
		Category: heuristicCategoryForward,
		Description: "Returns all destination transactions " +
			"which can be fully funded by the origins of their source. " +
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

func (h *forwardAmountHeuristic) clone() heuristic {
	newHeuristic := *h
	return &newHeuristic
}

// forwardAmountHeuristic applies the following heuristic:
// - filters all destinations which can not be funded by the sources based on the denominations of the source
func (h *forwardAmountHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string, parentHeuristicUID string) (
	[]heuristics.HeuristicCluster, error) {
	// origins hold all origins found bei either the parent heuristic
	//or the destination transaction specified by txHash
	origins := make(map[string]heuristics.HeuristicTransaction)
	// maps a cluster to its origin transactions
	clusterOrigins := make(map[heuristics.ClusterUID]map[string]heuristics.HeuristicTransaction)
	// attributionMap maps a clusterUID to a slice of attribution UIDs
	var attributionMap map[heuristics.ClusterUID][]string
	{ // separate enclosure so the results slice can be garbage collected
		var results []heuristics.HeuristicTransaction
		if isParentHeuristicSet(parentHeuristicUID) {
			// get origins from parent heuristic
			var err error
			results, attributionMap, err = heuristics.GetHeuristicResults(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, attributionMap, err = getDestinationTxOriginsTimeLimited(dgraph, g, txHash, h.lookForwardTime,
				h.userUID, h.clusterTypes, h.excludeAddresses)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		var err error
		clusterOrigins, err = addOriginsToMap(clusterOrigins, results)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// Convert from slice to Hash
		for _, r := range results {
			origins[r.UID] = r
		}
	}

	if len(origins) == 0 || len(clusterOrigins) == 0 {
		return nil, errorNoOriginsAtStart
	}

	var exclusions []string
	if h.excludeAddresses {
		var err error
		exclusions, err = exclusion.GetAddressExclusionUIDs(dgraph, h.userUID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
	}

	type clusterDestination struct {
		cluster heuristics.ClusterUID
		txs     map[string]heuristics.HeuristicTransaction
	}

	var clusterDestinations []clusterDestination

	for c, txMap := range clusterOrigins {
		var txUIDs []string

		for k := range txMap {
			txUIDs = append(txUIDs, k)
		}
		destinations, err := getOriginDestinationsWithInputs(dgraph, g, txUIDs, h.lookForwardTime, exclusions)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		destinationMap := make(map[string]heuristics.HeuristicTransaction)
		for _, d := range destinations {
			destinationMap[d.UID] = d
		}

		clusterDestinations = append(clusterDestinations, clusterDestination{cluster: c, txs: destinationMap})
	}

	originAmounts := buildSourceAmounts(origins)

	resultClusters := make(map[heuristics.ClusterUID][]heuristics.HeuristicResult)
	for _, destinations := range clusterDestinations {
		var clusterFilteredDestinations []heuristics.DummyNode

		for _, tx := range destinations.txs {
			inputDenominationCounts := getDenominationCounts(tx)

			// check if the denominations of the destination transactions can be funded by the denomination of its cluster
			if containsDenomination(inputDenominationCounts, originAmounts[destinations.cluster]) {
				clusterFilteredDestinations = append(clusterFilteredDestinations, heuristics.DummyNode{UID: tx.UID})
			}
		}

		// get properties of a random origin of this cluster
		var originUID string
		var clusterID heuristics.ClusterUID
		for _, v := range clusterOrigins[destinations.cluster] {
			originUID = v.UID
			clusterID = v.Cluster
			break
		}

		resultClusters[clusterID] = append(resultClusters[clusterID], heuristics.HeuristicResult{
			Origin:       heuristics.DummyNode{UID: originUID},
			Destinations: clusterFilteredDestinations,
		})
	}

	return createHeuristicClusters(resultClusters, attributionMap), nil
}
