package main

import (
	"backend/db/analytics"
	"backend/external"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

func doExportClusterActivity(dgraph external.Database, fileName string) {
	addresses, err := analytics.GetAllSingleAddresses(dgraph)
	if err != nil {
		warn(err)
		return
	}
	fmt.Println("number of addresses", len(addresses))

	results := map[int]int{}

	// handle addresses
	const step = 200
	var offset = 0
	addressCount := len(addresses)
	now := time.Now()
	for {
		upperBound := offset + step
		if upperBound > addressCount {
			upperBound = addressCount
		}

		counts, err := analytics.GetTransactionCountPerAddresses(dgraph, addresses[offset:upperBound])
		if err != nil {
			warn(err)
			return
		}

		for _, c := range counts {
			results[c]++
		}

		offset += step

		if offset >= addressCount {
			break
		}

		if offset%1000 == 0 {
			timePerAddress := time.Since(now) / time.Duration(offset)
			fmt.Printf("received counts for %d addresses. %v/a\n", offset, timePerAddress)
		}
	}

	// handle clusters
	clusters, err := analytics.GetAllFMIClusters(dgraph)
	if err != nil {
		warn(err)
		return
	}

	fmt.Println("number of clusters", len(clusters))

	for i, c := range clusters {
		txCount, err := analytics.GetTransactionCountPerCluster(dgraph, c)
		if err != nil {
			warn(err)
			return
		}

		results[txCount]++

		if i%1000 == 0 {
			fmt.Printf("received counts for %d clusters\n", i)
		}
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

	// header
	if err := w.Write([]string{"number of transactions", "cluster count"}); err != nil {
		warn(err, "msg", "error writing record to file")
		return
	}

	for k, v := range results {
		// column 1: number of transactions
		// column 2: cluster count
		line := []string{strconv.Itoa(k), strconv.Itoa(v)}
		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}

	w.Flush()
}
