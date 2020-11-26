package transaction

import (
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"time"
)

type TimeConstraintHeuristic struct {
	heuristicType        string
	parameterDescription string
	lookBackTime         time.Duration
}

// TimeConstraintHeuristic constructor
// hoursToLookBack in hours
func NewTimeConstraintHeuristic(hoursToLookBack uint32) TimeConstraintHeuristic {
	lBackTime := time.Duration(hoursToLookBack) * time.Hour
	return TimeConstraintHeuristic{
		heuristicType:        "timeconstraint",
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

func (h TimeConstraintHeuristic) setParameter(p string) error {
	return nil
}

func (h TimeConstraintHeuristic) String() string {
	return h.heuristicType
}

// TimeConstraintHeuristic applies the following heuristics:
// - filter all origins, which are not created in the time span defined by lookBackTime
func (h TimeConstraintHeuristic) exec(dgraph *dgo.Dgraph, txHash string, parentHeuristicUid string) ([]string, error) {
	var origins []string
	parentHeuristicSet := isParentHeuristicSet(parentHeuristicUid)
	if parentHeuristicSet {
		// get origins from parent heuristic
		parentHeuristic, err := dbtxh.GetHeuristic(dgraph, parentHeuristicUid)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(parentHeuristic.Origins) == 0 {
			return nil, ErrorNoOriginsAtStart
		}

		for _, o := range parentHeuristic.Origins {
			origins = append(origins, o.Uid)
		}
	}

	// gather input information
	inputTransactions, err := dbtxh.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	allTimeLimitedOrigins := make(map[string]bool)
	// holds all origins from either the parent heuristic or the associated destination transaction
	originLimit := make(map[string]bool)

	for _, o := range origins {
		originLimit[o] = true
	}

	for _, it := range inputTransactions {
		timeLimitedOrigins, err := getTimeLimitedOrigins(dgraph, it, h.lookBackTime)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// save all origins only once
		for _, t := range timeLimitedOrigins {
			// only save the uid also exists in the maximal origin set
			if parentHeuristicSet && !originLimit[t.Uid] {
				continue
			}
			allTimeLimitedOrigins[t.Uid] = true
		}
	}

	// convert map to string slice
	var filteredOrigins []string
	for k := range allTimeLimitedOrigins {
		filteredOrigins = append(filteredOrigins, k)
	}

	return filteredOrigins, nil
}
