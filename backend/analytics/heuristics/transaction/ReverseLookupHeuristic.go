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

// reverseLookupHeuristic - see exec for description
type reverseLookupHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookBackTime         time.Duration
}

// newReverseLookupHeuristic constructs a reverseLookupHeuristic. hoursToLookBack in hours.
func newReverseLookupHeuristic(hoursToLookBack uint32) *reverseLookupHeuristic {
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	return &reverseLookupHeuristic{
		heuristicType:        "reverse_lookup",
		lookBackTime:         lBackTime,
		parameterDescription: lBackTime.String(),
	}
}

func (h reverseLookupHeuristic) getType() string {
	return h.heuristicType
}

func (h reverseLookupHeuristic) getParameterString() string {
	return h.parameterDescription
}

func (h reverseLookupHeuristic) hasParameter() bool {
	return true
}

func (h *reverseLookupHeuristic) setParameter(p string) error {
	hoursToLookBack, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return err
	}
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	h.lookBackTime = lBackTime
	h.parameterDescription = strconv.FormatUint(hoursToLookBack, 10)
	return nil
}

func (h reverseLookupHeuristic) String() string {
	return fmt.Sprintf("Type: %s, Paramter: %s", h.heuristicType, h.parameterDescription)
}

func (h reverseLookupHeuristic) GetDescriptor() Descriptor {
	return Descriptor{
		Title:    "Reverse Lookup",
		Type:     h.heuristicType,
		Category: heuristicCategoryReverse,
		Description: "Performs a reverse lookup for the given duration and returns " +
			"all found origins. If this heuristic has a parent heuristic, only origins " +
			"which also occur in the parent heuristic will be returned. ",
		Parameter: &struct {
			DefaultValue string `json:"value,omitempty"`
			Description  string `json:"description,omitempty"`
			Type         string `json:"type,omitempty"`
		}{
			DefaultValue: "48",
			Description:  "Look back time in hours",
			Type:         "int",
		},
	}
}

func (h reverseLookupHeuristic) clone() heuristic {
	newHeuristic := h
	return &newHeuristic
}

// reverseLookupHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h reverseLookupHeuristic) exec(dgraph external.Database, g *graph.Wrapper, txHash string,
	parentHeuristicUID string) ([]dbtxh.HeuristicResult, error) {
	// holds all origins from either the parent heuristic or the associated destination transaction
	originLimit := make(map[string]bool)

	parentHeuristicSet := isParentHeuristicSet(parentHeuristicUID)
	if parentHeuristicSet {
		// get origins from parent heuristic
		parentHeuristic, err := dbtxh.GetHeuristic(dgraph, parentHeuristicUID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(parentHeuristic.Results) == 0 {
			return nil, errorNoOriginsAtStart
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
