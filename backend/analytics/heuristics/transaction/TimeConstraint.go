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

// TimeConstraintHeuristic - see exec for description
type TimeConstraintHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookBackTime         time.Duration
}

// NewTimeConstraintHeuristic constructs a TimeConstraintHeuristic. hoursToLookBack in hours.
func NewTimeConstraintHeuristic(hoursToLookBack uint32) *TimeConstraintHeuristic {
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	return &TimeConstraintHeuristic{
		heuristicType:        "time_constraint",
		lookBackTime:         lBackTime,
		parameterDescription: lBackTime.String(),
	}
}

func (h TimeConstraintHeuristic) getType() string {
	return h.heuristicType
}

func (h TimeConstraintHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h TimeConstraintHeuristic) hasParameter() bool {
	return true
}

func (h *TimeConstraintHeuristic) setParameter(p string) error {
	hoursToLookBack, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	h.lookBackTime = lBackTime
	h.parameterDescription = strconv.FormatUint(hoursToLookBack, 10)
	return nil
}

func (h TimeConstraintHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h TimeConstraintHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:       "Time Constraint",
		Type:        h.heuristicType,
		Description: "Filters by time.",
		Parameter: struct {
			DefaultValue string `json:"default_value,omitempty"`
			Description  string `json:"description,omitempty"`
			Type         string `json:"type,omitempty"`
		}{
			DefaultValue: "48",
			Description:  "Look back time in hours",
			Type:         "int",
		},
	}
}

func (h TimeConstraintHeuristic) clone() Heuristic {
	newHeuristic := h
	return &newHeuristic
}

// TimeConstraintHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h TimeConstraintHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string) ([]dbtxh.HeuristicResult, error) {
	// holds all origins from either the parent Heuristic or the associated destination transaction
	originLimit := make(map[string]bool)

	parentHeuristicSet := isParentHeuristicSet(parentHeuristicUID)
	if parentHeuristicSet {
		// get origins from parent Heuristic
		parentHeuristic, err := dbtxh.GetHeuristic(dgraph, parentHeuristicUID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(parentHeuristic.Results) == 0 {
			return nil, ErrorNoOriginsAtStart
		}

		for _, r := range parentHeuristic.Results {
			originLimit[r.Origin.UID] = true
		}
	}

	// gather input information
	inputTransactions, err := dbtxh.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	allTimeLimitedOrigins := make(map[string]bool)

	for _, it := range inputTransactions {
		timeLimitedOrigins, err := getTimeLimitedOrigins(dgraph, g, it, h.lookBackTime)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// save all origins only once
		for _, t := range timeLimitedOrigins {
			// only save if the uid also exists in the parent origin set
			if parentHeuristicSet && !originLimit[t.UID] {
				continue
			}
			allTimeLimitedOrigins[t.UID] = true
		}
	}

	var ret []dbtxh.HeuristicResult
	for k := range allTimeLimitedOrigins {
		ret = append(ret, dbtxh.HeuristicResult{
			Origin: dbtxh.DummyNode{UID: k},
		})
	}

	return ret, nil
}
