// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"encoding/csv"
	"os"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"

	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// doDestinationCountAnalysis investigates if outputs of destination transactions
// which are sent to the same cluster can be linked via destination input
// timestamp fingerprinting. Destination transactions which send funds to large clusters (>1000) are excluded.
func doDestinationCountAnalysis(ctx context.Context, dgraph external.Database, g *graph.ReversibleGraph,
	fileName string, transactionType string) {
	if fileName == "" {
		warn(serror.FromStr("file name is empty"))
		return
	}

	if transactionType == "" {
		warn(serror.FromStr("transaction type is empty"))
		return
	}

	spenders, globalDestinationCount, includedDestinationCount, globalClusterCount, includedClusterCount, err :=
		analytics.GetDestinationTransactionClusterSpenders(ctx, dgraph, transactionType)
	if err != nil {
		warn(err)
		return
	}

	info("destination counts",
		"global destination count", globalDestinationCount, "included destination count", includedDestinationCount,
		"global cluster count", globalClusterCount, "included cluster count", includedClusterCount)

	var foundCountTop30 atomic.Int64
	var foundCountTop20 atomic.Int64
	var foundCountTop10 atomic.Int64

	jobs := make(chan analytics.SpenderTransaction, len(spenders))

	wg := sync.WaitGroup{}

	for range 150 {
		wg.Add(1)
		go func(jobs <-chan analytics.SpenderTransaction) {
			defer wg.Done()
			for spender := range jobs {
				for _, destination := range spender.Destinations {
					fingerprints, _, err := graph.SpendingFingerprint(g, destination.UID, 30)
					if err != nil {
						warn(err)
						return
					}

					foundAny := false

					if len(fingerprints) >= 10 {
						// create map with all fingerprints
						mapFingerprintsTop10 := map[string]bool{}
						for _, fingerprint := range fingerprints[:10] {
							mapFingerprintsTop10[fingerprint.TransactionUID] = true
						}

						// check if one of the fingerprints is one of the other destination transactions
						if slices.ContainsFunc(spender.Destinations, func(transaction db.Transaction) bool {
							return mapFingerprintsTop10[transaction.UID]
						}) {
							// found matching fingerprint for one of the
							// other destination transactions, therefore increase the count
							foundCountTop10.Add(1)
							foundAny = true
						}
					}

					if len(fingerprints) >= 20 {
						// create map with all fingerprints
						mapFingerprintsTop20 := map[string]bool{}
						for _, fingerprint := range fingerprints[:20] {
							mapFingerprintsTop20[fingerprint.TransactionUID] = true
						}

						// check if one of the fingerprints is one of the other destination transactions
						if slices.ContainsFunc(spender.Destinations, func(transaction db.Transaction) bool {
							return mapFingerprintsTop20[transaction.UID]
						}) {
							// found matching fingerprint for one of the
							// other destination transactions, therefore increase the count
							foundCountTop20.Add(1)
							foundAny = true
						}
					}

					// create map with all fingerprints
					mapFingerprintsTop30 := map[string]bool{}
					for _, fingerprint := range fingerprints {
						mapFingerprintsTop30[fingerprint.TransactionUID] = true
					}

					// check if one of the fingerprints is one of the other destination transactions
					if slices.ContainsFunc(spender.Destinations, func(transaction db.Transaction) bool {
						return mapFingerprintsTop30[transaction.UID]
					}) {
						// found matching fingerprint for one of the
						// other destination transactions, therefore increase the count
						foundCountTop30.Add(1)
						foundAny = true
					}

					if foundAny {
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
		"Successful fingerprint count top 30", foundCountTop30.Load(),
		"Percent", float64(foundCountTop30.Load())/float64(len(spenders)),
		"Successful fingerprint count top 20", foundCountTop20.Load(),
		"Percent", float64(foundCountTop20.Load())/float64(len(spenders)),
		"Successful fingerprint count top 10", foundCountTop10.Load(),
		"Percent", float64(foundCountTop10.Load())/float64(len(spenders)),
	)

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
		// column 1: transaction hash of spender transaction or cluster UID
		// column 2: cluster size of spender transaction inputs
		// column 3-x: destination transactions which have been spent by tx in column 1

		key := t.Transaction.Hash
		if t.ClusterUID != "" {
			key = t.ClusterUID
		}

		line := []string{key, strconv.Itoa(t.ClusterSize)}

		for _, d := range t.Destinations {
			line = append(line, d.Hash)
		}

		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}
}
