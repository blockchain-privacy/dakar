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
// lookBackTime in hours
func NewTimeConstraintHeuristic(hoursToLookBack time.Duration) TimeConstraintHeuristic {
	lBackTime := hoursToLookBack * time.Hour
	return TimeConstraintHeuristic{
		heuristicType:        "timeconstraint",
		lookBackTime:         lBackTime,
		parameterDescription: lBackTime.String(),
	}
}

func (b TimeConstraintHeuristic) getType() string {
	return b.heuristicType
}

func (b TimeConstraintHeuristic) getParameter() string {
	return b.parameterDescription
}

// time limitation
func (b TimeConstraintHeuristic) exec(dgraph *dgo.Dgraph, txHash string, origins []string) ([]string, error) {
	// gather input information
	inputTransactions, err := dbtxh.GetInputTransactions(dgraph, txHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	allTimeLimitedOrigins := make(map[string]bool)

	for _, it := range inputTransactions {
		// calculate look back time
		t, err := time.Parse(time.RFC3339, it.Timestamp)
		if err != nil {
			return nil, err
		}
		// get time limited origins
		t = t.Add(-1 * b.lookBackTime)
		timeLimitedOrigins, err := dbtxh.GetOriginsByDate(dgraph, it.Uid, t.Format(time.RFC3339))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// save all origins only once
		for _, t := range timeLimitedOrigins {
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
