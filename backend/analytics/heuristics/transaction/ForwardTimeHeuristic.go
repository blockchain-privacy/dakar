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

// NewForwardTimeHeuristic constructs a ForwardTimeHeuristic. hoursToLookBack in hours.
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
	hoursToLookBack, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}

	h.lookForwardTime = time.Duration(hoursToLookBack) * time.Hour
	h.parameterDescription = strconv.FormatUint(hoursToLookBack, 10)
	return nil
}

func (h ForwardTimeHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h ForwardTimeHeuristic) clone() heuristic {
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
			// get origins from parent heuristic
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
		destinations, err := getOriginTxDestinationsTimeLimited(dgraph, g, []string{o.Origin.UID}, h.lookForwardTime)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		for _, v := range destinations {
			hResult[i].Destinations = append(hResult[i].Destinations, dbtxh.DummyNode{UID: v.UID})
		}
	}

	return hResult, nil
}
