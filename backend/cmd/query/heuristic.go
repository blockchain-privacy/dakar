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

func doHeuristicAnalysis(ctx context.Context, dgraph external.Database, g *graph.ReversibleGraph,
	fileName string, transactionType string) {
	info("heuristic analysis starting")
	if fileName == "" {
		warn(serror.FromStr("file name is empty"))
		return
	}

	if transactionType == "" {
		warn(serror.FromStr("transaction type is empty"))
		return
	}

	var txHeuristics []heuristics.Heuristic
	var mixingTxType string
	switch transactionType {
	case constants.TypeDashDestination:
		mixingTxType = constants.TypeDashMixing
		txHeuristics = []heuristics.Heuristic{heuristics.NewReverseLookupHeuristic(), heuristics.NewOneSourceHeuristic()}
	case constants.TypeWasabi2Destination:
		mixingTxType = constants.TypeWasabi2Mixing
		txHeuristics = []heuristics.Heuristic{heuristics.NewWasabi2ReverseLookupByTimeHeuristic(), heuristics.NewWasabi2OneSourceByTimeHeuristic()}
	case constants.TypeWhirlpoolDestination:
		mixingTxType = constants.TypeWhirlpoolMixing
		txHeuristics = []heuristics.Heuristic{heuristics.NewWhirlpoolReverseLookupByTimeHeuristic(), heuristics.NewWhirlpoolOneSourceByTimeHeuristic()}
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

	lookbackDurations := []int{12, 24, 48}
	for i := 0; ; i += step {
		now := time.Now()
		destinations, err := analytics.GetPrivacyTransactionsWithHash(ctx, dgraph, step, i, transactionType, mixingTxType)
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

			var sum int64
			for _, t := range destination.Outputs {
				sum += t.Amount
			}

			// column 1: transaction hash
			// column 2: transaction timestamp
			// column 3: input count (only mixing)
			// column 4: input amount
			// column 5: number of clusters (reverse lookup 12h)
			// column 6: number of origins (reverse lookup 12h)
			// column 7: number of clusters (reverse lookup 24h)
			// column 8: number of origins (reverse lookup 24h)
			// column 9: number of clusters (reverse lookup 48h)
			// column 10: number of origins (reverse lookup 48h)
			// column 11: number of clusters (one source 12h)
			// column 12: number of origins (one source 12h)
			// column 13: number of clusters (one source 24h)
			// column 14: number of origins (one source 24h)
			// column 15: number of clusters (one source 48h)
			// column 16: number of origins (one source 48h)
			line := []string{destination.Hash, ts.Format(time.RFC3339),
				strconv.Itoa(len(destination.Outputs)), strconv.FormatInt(sum, 10)}
			for _, txHeuristic := range txHeuristics {
				for _, duration := range lookbackDurations {
					if err = txHeuristic.SetConfig(dbh.Options{Parameter: strconv.Itoa(duration),
						TransactionHash: destination.Hash}); err != nil {
						warn(err)
						return
					}

					clusters, err := txHeuristic.Exec(ctx, dgraph, wrapper, destination.UID)
					if err != nil {
						warn(err)
						return
					}

					var originCount int
					for _, cluster := range clusters {
						originCount += len(cluster.Results)
					}

					line = append(line, strconv.Itoa(len(clusters)), strconv.Itoa(originCount))
				}
			}

			if err := w.Write(line); err != nil {
				warn(err, "msg", "error writing record to file")
				return
			}
		}

		info("execution duration", "duration/transactions",
			time.Since(now)/time.Duration(len(destinations)), "transaction count", len(destinations))

		if len(destinations) < step {
			break
		}
	}
}
