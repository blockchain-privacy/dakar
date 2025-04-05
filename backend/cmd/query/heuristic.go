package main

import (
	"backend/analytics/graph"
	"backend/analytics/heuristics"
	"backend/constants"
	"backend/db/analytics"
	dbh "backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"encoding/csv"
	"github.com/qrest/gomisc/serror"
	"os"
	"strconv"
	"time"
)

func doHeuristicAnalysis(ctx context.Context, dgraph external.Database,
	g *graph.ReversibleGraph, fileName string, transactionType string, heuristicType string, lookBackHours int) {
	info("heuristic analysis starting")
	if fileName == "" {
		warn(serror.FromStr("file name is empty"))
		return
	}

	if transactionType == "" {
		warn(serror.FromStr("transaction type is empty"))
		return
	}

	if lookBackHours == 0 {
		warn(serror.FromStr("lookback time is zero"))
		return
	}

	var mixingTxType string
	switch transactionType {
	case constants.TypeDashDestination:
		mixingTxType = constants.TypeDashMixing
	case constants.TypeWasabi2Destination:
		mixingTxType = constants.TypeWasabi2Mixing
	case constants.TypeWhirlpoolDestination:
		mixingTxType = constants.TypeWhirlpoolMixing
	default:
		warn(serror.FromStrWithContext("invalid transaction type", "type", transactionType))
		return
	}

	f, err := os.Create(fileName)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			warn(err)
		}
	}(f)

	if err != nil {
		warn(err)
		return
	}

	w := csv.NewWriter(f)
	defer w.Flush()

	const step = 10000
	minDate := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)

	wrapper := graph.NewWrapper(ctx, dgraph)
	wrapper.SetGraph(g)

	heuristicConstructor, ok := heuristics.ConstructorMap[heuristicType]
	if !ok {
		warn(serror.FromStrWithContext("invalid heuristic type", "type", heuristicType))
		return
	}

	h := heuristicConstructor()

	for i := 0; ; i += step {
		destinations, err := analytics.GetPrivacyTransactionsWithHash(ctx, dgraph, step, i, transactionType)
		if err != nil {
			warn(err)
			return
		}

		for _, destination := range destinations {
			if destination.Block == nil {
				warn(serror.FromStrWithContext("block is nil", "node", destination))
				return
			}

			ts := destination.Block[0].TS
			if ts.Sub(minDate) < 0 {
				// don't consider transactions before the minimum date
				continue
			}

			tx, err := dbh.GetInputAmounts(ctx, dgraph, destination.Hash, mixingTxType)
			if err != nil {
				warn(err)
				return
			}

			var sum int64
			for _, t := range tx.Outputs {
				sum += t.Amount
			}

			if err = h.SetConfig(dbh.Options{Parameter: strconv.Itoa(lookBackHours),
				TransactionHash: destination.Hash}); err != nil {
				warn(err)
				return
			}

			clusters, err := h.Exec(ctx, dgraph, wrapper, "")
			if err != nil {
				return
			}

			var originCount int
			for _, cluster := range clusters {
				originCount += len(cluster.Results)
			}

			// column 1: transaction UID
			// column 2: transaction timestamp
			// column 3: input amount
			// column 4: number of clusters
			// column 5: number of origins
			line := []string{destination.UID, ts.Format(time.RFC3339), strconv.FormatInt(sum, 10), strconv.Itoa(len(clusters)), strconv.Itoa(originCount)}
			if err := w.Write(line); err != nil {
				warn(err, "msg", "error writing record to file")
				return
			}
		}

		if len(destinations) < step {
			break
		}
	}
}
