// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"encoding/csv"
	"errors"
	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/analytics/heuristics"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db/analytics"
	"gitlab.com/blockchain-privacy/dakar/external"
	"io"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"

	"gitlab.com/blockchain-privacy/gomisc/serror"
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

	recoveredData, err := tryRecoverFromFile(fileName, 22)
	if err != nil {
		warn(err)
		return
	}

	txHeuristics, reverseAmountHeuristic, mixingTxType, err := getConstructors(transactionType)
	if err != nil {
		warn(err)
		return
	}

	f, err := os.Create(fileName)
	if err != nil {
		warn(serror.New(err))
		return
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			warn(err)
		}
	}(f)

	w := csv.NewWriter(f)
	defer w.Flush()

	if recoveredData != nil {
		info("recovered items", "count", len(recoveredData))
	}

	recoveredDestinations := make(map[string]bool, len(recoveredData))
	for _, recLine := range recoveredData {
		recoveredDestinations[recLine[0]] = true
		if err := w.Write(recLine); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}

	w.Flush()

	const step = 10000
	minDate := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)

	wrapper := graph.NewWrapper(ctx, dgraph)
	wrapper.SetGraph(g)

	lookbackDurations := []string{"24", "48"}

	workContext, cancel := context.WithCancel(ctx)
	defer cancel()

	const goroutineCount = 25

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
	lastNodeUID := "0x0"
	for i := 0; ; i += step {
		now := time.Now()
		destinations, err := analytics.GetPrivacyTransactionsWithHash(ctx, dgraph, step, lastNodeUID, transactionType, mixingTxType)
		if err != nil {
			warn(err)
			return
		}

		if len(destinations) > 0 {
			lastNodeUID = destinations[len(destinations)-1].UID
		}

		for _, destination := range destinations {
			if destination.Block == nil {
				warn(serror.FromStrWithContext("block is nil", "node", destination))
				return
			}

			if recoveredDestinations[destination.Hash] {
				// already in file
				continue
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
	if err := amountHeuristic.SetConfig(heuristics.HeuristicOptions{TransactionHash: destination.Hash}); err != nil {
		return nil, err
	}

	var sum int64
	for _, t := range destination.Inputs {
		sum += t.Amount
	}

	slices.SortFunc(destination.Inputs, func(a, b struct {
		Amount int64     `json:"amount,omitempty"`
		TS     time.Time `json:"ts"`
	}) int {
		return a.TS.Compare(b.TS)
	})

	inputSpread := destination.Inputs[len(destination.Inputs)-1].TS.Sub(destination.Inputs[0].TS)

	// column 1: transaction hash
	// column 2: transaction timestamp
	// column 3: input count (only mixing)
	// column 4: output count
	// column 5: input amount
	// column 6: input timestamp spread (in seconds, only mixing)
	// column 7: number of clusters (reverse lookup 24h)
	// column 8: number of origins (reverse lookup 24h)
	// column 9: number of clusters (reverse amount for reverse lookup 24)
	// column 10: number of origins (reverse amount for reverse lookup 24)
	// column 11: number of clusters (reverse lookup 48h)
	// column 12: number of origins (reverse lookup 48h)
	// column 13: number of clusters (reverse amount for reverse lookup 48)
	// column 14: number of origins (reverse amount for reverse lookup 48)
	// column 15: number of clusters (one source 24h)
	// column 16: number of origins (one source 24h)
	// column 17: number of clusters (reverse amount for one source 24)
	// column 18: number of origins (reverse amount for one source 24)
	// column 19: number of clusters (one source 48h)
	// column 20: number of origins (one source 48h)
	// column 21: number of clusters (reverse amount for one source 48)
	// column 22: number of origins (reverse amount for one source 48)
	line := []string{destination.Hash, destination.Block[0].TS.UTC().Format(time.RFC3339),
		strconv.Itoa(len(destination.Inputs)), strconv.Itoa(destination.OutputCount),
		strconv.FormatInt(sum, 10), strconv.FormatFloat(inputSpread.Seconds(), 'f', 2, 64)}

	for _, txHeuristic := range txHeuristics {
		thisHeuristic := txHeuristic()
		for _, duration := range lookbackDurations {
			// this heuristic
			if err := thisHeuristic.SetConfig(heuristics.HeuristicOptions{Parameter: duration,
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

// tryRecoverFromFile checks if there is a file with results already and reads them
func tryRecoverFromFile(fileName string, columnCount int) ([][]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		// file may not exist, just return no data
		return nil, nil
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			warn(err)
		}
	}(f)

	csvReader := csv.NewReader(f)
	var data [][]string
	for {
		line, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			continue
		}

		if len(line) != columnCount {
			continue
		}

		data = append(data, line)
	}

	if len(data) == 0 {
		return nil, nil
	}

	// remove last line, because it is often only partially flushed to the file
	return data[:len(data)-1], nil
}
