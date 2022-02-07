package transaction

import (
	"backend/analytics/graph"
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	"backend/external"

	"fmt"
	"strconv"
	"time"
)

// ForwardLookupHeuristic - see exec for description
type ForwardLookupHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookForwardTime      time.Duration
}

// NewForwardLookupHeuristic constructs a ForwardLookupHeuristic. hoursToLookForward in hours.
func NewForwardLookupHeuristic(hoursToLookForward uint32) *ForwardLookupHeuristic {
	lForwardTime := time.Duration(hoursToLookForward) * time.Hour
	return &ForwardLookupHeuristic{
		heuristicType:        "forward_lookup",
		lookForwardTime:      lForwardTime,
		parameterDescription: lForwardTime.String(),
	}
}

func (h ForwardLookupHeuristic) getType() string {
	return h.heuristicType
}

func (h ForwardLookupHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h ForwardLookupHeuristic) hasParameter() bool {
	return true
}

func (h *ForwardLookupHeuristic) setParameter(p string) error {
	hoursToLookForward, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.parameterDescription = strconv.FormatUint(hoursToLookForward, 10)
	return nil
}

func (h ForwardLookupHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h ForwardLookupHeuristic) GetDescriptor() Descriptor {
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

func (h ForwardLookupHeuristic) clone() Heuristic {
	newHeuristic := h
	return &newHeuristic
}

// ForwardLookupHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h ForwardLookupHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string) ([]dbtxh.HeuristicResult, error) {

	var hResult []dbtxh.HeuristicResult
	{ // separate enclosure so the results slice can be garbage collected
		var results []dbtxh.HeuristicTransaction

		if isParentHeuristicSet(parentHeuristicUID) {
			// get origins from parent heuristic
			var err error
			results, err = dbtxh.GetHeuristicResults(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, err = getDestinationTxOriginsTimeLimited(dgraph, g, txHash, h.lookForwardTime)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		if len(results) == 0 {
			return nil, errorNoOriginsAtStart
		}

		for _, r := range results {
			hResult = append(hResult, dbtxh.HeuristicResult{Origin: dbtxh.DummyNode{UID: r.UID}})
		}
	}

	for i, o := range hResult {
		destinations, err := getOriginDestinationsWithOutputs(dgraph, g, []string{o.Origin.UID}, h.lookForwardTime)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		for _, v := range destinations {
			hResult[i].Destinations = append(hResult[i].Destinations, dbtxh.DummyNode{UID: v.UID})
		}
	}

	return hResult, nil
}
