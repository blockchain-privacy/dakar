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

// ForwardTimeHeuristic - see exec for description
type ForwardTimeHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookForwardTime      time.Duration
}

// NewForwardTimeHeuristic constructs a ForwardTimeHeuristic. hoursToLookForward in hours.
func NewForwardTimeHeuristic(hoursToLookForward uint32) *ForwardTimeHeuristic {
	lForwardTime := time.Duration(hoursToLookForward) * time.Hour
	return &ForwardTimeHeuristic{
		heuristicType:        "forward_time",
		lookForwardTime:      lForwardTime,
		parameterDescription: lForwardTime.String(),
	}
}

func (h ForwardTimeHeuristic) getType() string {
	return h.heuristicType
}

func (h ForwardTimeHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h ForwardTimeHeuristic) hasParameter() bool {
	return true
}

func (h *ForwardTimeHeuristic) setParameter(p string) error {
	hoursToLookForward, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}

	h.lookForwardTime = time.Duration(hoursToLookForward) * time.Hour
	h.parameterDescription = strconv.FormatUint(hoursToLookForward, 10)
	return nil
}

func (h ForwardTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h ForwardTimeHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title: "Forward Lookup",
		Type:  h.heuristicType,
		Description: "Performs a forward lookup for each origin " +
			"transaction of the parent Heuristic. If this Heuristic " +
			"is placed at the root level a reverse lookup with 48h " +
			"look back time will be performed.",
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

func (h ForwardTimeHeuristic) clone() Heuristic {
	newHeuristic := h
	return &newHeuristic
}

// ForwardTimeHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h ForwardTimeHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string) ([]dbtxh.HeuristicResult, error) {

	var hResult []dbtxh.HeuristicResult
	{ // separate enclosure so the results slice can be garbage collected
		var results []dbtxh.HeuristicTransaction

		if isParentHeuristicSet(parentHeuristicUID) {
			// get origins from parent Heuristic
			var err error
			results, err = dbtxh.GetHeuristicResults(dgraph, parentHeuristicUID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		} else {
			var err error
			results, err = getDestinationTxOriginsTimeLimited(dgraph, g, txHash, time.Hour*48)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		if len(results) == 0 {
			return nil, ErrorNoOriginsAtStart
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
