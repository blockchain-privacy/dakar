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
	"sync"
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

	txHeuristics, reverseAmountHeuristic, mixingTxType, err := getConstructors(transactionType)
	if err != nil {
		warn(err)
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

	lookbackDurations := []string{"24", "48"}

	workContext, cancel := context.WithCancel(ctx)
	defer cancel()

	const goroutineCount = 100

	// work on jobs
	jobs := make(chan analytics.NodeWithHash, goroutineCount*10)
	results := make(chan []string, goroutineCount*10)
	wg := sync.WaitGroup{}
	for range goroutineCount {
		wg.Add(1)
		go func(jobs <-chan analytics.NodeWithHash, results chan<- []string) {
			defer wg.Done()
			for destination := range jobs {
				select {
				case <-workContext.Done():
					return
				default:
				}
				line, err := executeHeuristics(ctx, dgraph, wrapper, destination, lookbackDurations, txHeuristics, reverseAmountHeuristic)
				if err != nil {
					cancel()
					warn(err)
					return
				}

				results <- line
			}
		}(jobs, results)
	}

	done := make(chan struct{})
	go func() {
		for line := range results {
			select {
			case <-workContext.Done():
				return
			default:
			}
			if err := w.Write(line); err != nil {
				cancel()
				warn(err, "msg", "error writing record to file")
				return
			}
		}
		// signal that all results got processed
		close(done)
	}()

	// create jobs
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

			select {
			case <-workContext.Done():
				return
			case jobs <- destination:
			}
		}

		info("execution duration", "duration/transactions",
			time.Since(now)/time.Duration(len(destinations)), "transaction count", len(destinations))

		if len(destinations) < step {
			break
		}
	}

	close(jobs)
	// wait for workers to finish
	wg.Wait()
	// results channel can now be closed
	close(results)
	// wait until all results got processed
	<-done
}

func getConstructors(transactionType string) ([]heuristics.HeuristicConstructor, heuristics.HeuristicConstructor, string, error) {
	var txHeuristics []heuristics.HeuristicConstructor
	var reverseAmountHeuristic heuristics.HeuristicConstructor
	var mixingTxType string
	switch transactionType {
	case constants.TypeDashDestination:
		mixingTxType = constants.TypeDashMixing
		txHeuristics = []heuristics.HeuristicConstructor{heuristics.NewReverseLookupHeuristic, heuristics.NewOneSourceHeuristic}
		reverseAmountHeuristic = heuristics.NewReverseAmountHeuristic
	case constants.TypeWasabi2Destination:
		mixingTxType = constants.TypeWasabi2Mixing
		txHeuristics = []heuristics.HeuristicConstructor{heuristics.NewWasabi2ReverseLookupByTimeHeuristic, heuristics.NewWasabi2OneSourceByTimeHeuristic}
		reverseAmountHeuristic = heuristics.NewWasabi2ReverseAmountHeuristic
	case constants.TypeWhirlpoolDestination:
		mixingTxType = constants.TypeWhirlpoolMixing
		txHeuristics = []heuristics.HeuristicConstructor{heuristics.NewWhirlpoolReverseLookupByTimeHeuristic, heuristics.NewWhirlpoolOneSourceByTimeHeuristic}
		reverseAmountHeuristic = heuristics.NewWhirlpoolReverseAmountHeuristic
	default:
		return nil, nil, "", serror.FromStrWithContext("invalid transaction type", "type", transactionType)
	}

	return txHeuristics, reverseAmountHeuristic, mixingTxType, nil
}

func executeHeuristics(ctx context.Context, dgraph external.Database, wrapper *graph.Wrapper,
	destination analytics.NodeWithHash, lookbackDurations []string, txHeuristics []heuristics.HeuristicConstructor,
	reverseAmountHeuristic heuristics.HeuristicConstructor) ([]string, error) {
	amountHeuristic := reverseAmountHeuristic()
	if err := amountHeuristic.SetConfig(dbh.Options{TransactionHash: destination.Hash}); err != nil {
		return nil, err
	}

	var sum int64
	for _, t := range destination.Outputs {
		sum += t.Amount
	}

	// column 1: transaction hash
	// column 2: transaction timestamp
	// column 3: input count (only mixing)
	// column 4: input amount
	// column 5: number of clusters (reverse lookup 24h)
	// column 6: number of origins (reverse lookup 24h)
	// column 7: number of clusters (reverse amount for reverse lookup 24)
	// column 8: number of origins (reverse amount for reverse lookup 24)
	// column 9: number of clusters (reverse lookup 48h)
	// column 10: number of origins (reverse lookup 48h)
	// column 11: number of clusters (reverse amount for reverse lookup 48)
	// column 12: number of origins (reverse amount for reverse lookup 48)
	// column 13: number of clusters (one source 24h)
	// column 14: number of origins (one source 24h)
	// column 15: number of clusters (reverse amount for one source 24)
	// column 16: number of origins (reverse amount for one source 24)
	// column 17: number of clusters (one source 48h)
	// column 18: number of origins (one source 48h)
	// column 19: number of clusters (reverse amount for one source 48)
	// column 20: number of origins (reverse amount for one source 48)
	line := []string{destination.Hash, destination.Block[0].TS.Format(time.RFC3339),
		strconv.Itoa(len(destination.Outputs)), strconv.FormatInt(sum, 10)}

	for _, txHeuristic := range txHeuristics {
		thisHeuristic := txHeuristic()
		for _, duration := range lookbackDurations {
			// this heuristic
			if err := thisHeuristic.SetConfig(dbh.Options{Parameter: duration,
				TransactionHash: destination.Hash}); err != nil {
				return nil, err
			}

			thisClusters, err := thisHeuristic.Exec(ctx, dgraph, wrapper, destination.UID, nil)
			if err != nil {
				return nil, err
			}

			var thisOriginCount int
			for _, cluster := range thisClusters {
				thisOriginCount += len(cluster.Results)
			}

			line = append(line, strconv.Itoa(len(thisClusters)), strconv.Itoa(thisOriginCount))

			if len(thisClusters) == 0 {
				line = append(line, "0", "0")
				continue
			}

			// reverse amount heuristic
			amountClusters, err := amountHeuristic.Exec(ctx, dgraph, wrapper, "", thisClusters)
			if err != nil {
				return nil, err
			}

			var amountOriginCount int
			for _, cluster := range amountClusters {
				amountOriginCount += len(cluster.Results)
			}

			line = append(line, strconv.Itoa(len(amountClusters)), strconv.Itoa(amountOriginCount))
		}
	}

	return line, nil
}
