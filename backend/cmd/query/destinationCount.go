package main

import (
	"backend/analytics/graph"
	"backend/db"
	"backend/db/analytics"
	"backend/external"
	"context"
	"encoding/csv"
	"os"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
)

// doDestinationCountAnalysis investigates if outputs of destination transactions
// which are spent in the same transaction can be linked via destination input
// timestamp fingerprinting. This is achieved by collecting all transactions
// (spending transactions) which are directly connected to multiple (>=2) destination
// transactions. Spending transactions created by large clusters (>1000) are excluded.
func doDestinationCountAnalysis(ctx context.Context, dgraph external.Database, g *graph.ReversibleGraph, fileName string) {
	spenders, globalDestinationCount, spentDestinationTransactionCount, excludedBecauseOfClusterSizeCount, usingDestinationTransactionsCount, err :=
		analytics.GetDestinationTransactionSpenders(ctx, dgraph)
	if err != nil {
		warn(err)
		return
	}

	info("destination counts",
		"global destination count", globalDestinationCount,
		"spent destination transactions", spentDestinationTransactionCount,
		"excluded because of cluster size", excludedBecauseOfClusterSizeCount,
		"using destination transactions count", usingDestinationTransactionsCount)

	var foundCount atomic.Int64

	jobs := make(chan analytics.SpenderTransaction, len(spenders))

	wg := sync.WaitGroup{}

	for range 150 {
		wg.Add(1)
		go func(jobs <-chan analytics.SpenderTransaction) {
			defer wg.Done()
			for spender := range jobs {
				for _, destination := range spender.Destinations {
					fingerprints, _, err := graph.SpendingFingerprint(g, destination.UID)
					if err != nil {
						warn(err)
						return
					}

					// create map with all fingerprint
					mapFingerprints := map[string]bool{}
					for _, fingerprint := range fingerprints {
						mapFingerprints[fingerprint.TransactionUID] = true
					}

					// check if one of the fingerprints is one of the other destination transactions
					if slices.ContainsFunc(spender.Destinations, func(transaction db.Transaction) bool {
						return mapFingerprints[transaction.UID]
					}) {
						// found matching fingerprint for one of the
						// other destination transactions, therefore increase the count
						foundCount.Add(1)
						break
					}
				}
			}
		}(jobs)
	}

	for _, spender := range spenders {
		jobs <- spender
	}

	close(jobs)
	wg.Wait()

	info("fingerprint analysis",
		"Spender count", len(spenders),
		"Successful fingerprint count", foundCount.Load(),
		"Percent", float64(foundCount.Load())/float64(len(spenders)))

	writeSpendersToCSV(fileName, spenders)
}

func writeSpendersToCSV(fileName string, txs []analytics.SpenderTransaction) {
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

	for _, t := range txs {
		// column 1: transaction hash of spender transaction
		// column 2: cluster size of spender transaction inputs
		// column 3-x: destination transactions which have been spent by tx in column 1
		line := []string{t.Transaction.Hash, strconv.Itoa(t.ClusterSize)}

		for _, d := range t.Destinations {
			line = append(line, d.Hash)
		}

		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}
}
