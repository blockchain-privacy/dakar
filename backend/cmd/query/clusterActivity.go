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
	if err := w.Write([]string{"cluster uid", "number of created transactions", "number of received transactions"}); err != nil {
		warn(err, "msg", "error writing record to file")
		return
	}

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

		addressSlice := addresses[offset:upperBound]

		inputCounts, outputCounts, err := analytics.GetTransactionCountPerAddresses(dgraph, addressSlice)
		if err != nil {
			warn(err)
			return
		}

		for i, address := range addressSlice {
			// column 1: cluster uid
			// column 2: number of created transactions
			// column 3: number of received transactions
			line := []string{address, strconv.Itoa(inputCounts[i]), strconv.Itoa(outputCounts[i])}
			if err := w.Write(line); err != nil {
				warn(err, "msg", "error writing record to file")
				return
			}
		}

		w.Flush()

		offset += step

		if offset >= addressCount {
			break
		}

		if offset%10000 == 0 {
			timePerAddress := time.Since(now) / time.Duration(10000)
			fmt.Printf("received counts for %d addresses. %v/a\n", offset, timePerAddress)
			now = time.Now()
		}
	}

	// handle clusters
	clusters, err := analytics.GetAllFMIClusters(dgraph)
	if err != nil {
		warn(err)
		return
	}

	fmt.Println("number of clusters", len(clusters))
	now = time.Now()
	for i, c := range clusters {
		inputCount, outputCount, err := analytics.GetTransactionCountPerCluster(dgraph, c)
		if err != nil {
			warn(err)
			return
		}

		// column 1: cluster uid
		// column 2: number of created transactions
		// column 3: number of received transactions
		line := []string{c, strconv.Itoa(inputCount), strconv.Itoa(outputCount)}
		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}

		if i%1000 == 0 {
			fmt.Printf("received counts for %d clusters\n", i)
			timePerCluster := time.Since(now) / time.Duration(1000)
			fmt.Printf("received counts for %d clusters. %v/c\n", i, timePerCluster)
			now = time.Now()
		}
	}

	w.Flush()
}
