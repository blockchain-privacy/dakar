package main

import (
	"backend/db/analytics"
	"backend/external"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

type clusterActivity struct {
	uid              string
	transactionCount int
}

func doExportClusterActivity(dgraph external.Database, fileName string) {
	addresses, err := analytics.GetAllSingleAddresses(dgraph)
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

	// handle addresses
	activities := make([]clusterActivity, len(addresses))
	for i, a := range addresses {
		txCount, err := analytics.GetTransactionCountPerAddress(dgraph, a)
		if err != nil {
			warn(err)
			return
		}
		activities[i] = clusterActivity{uid: a, transactionCount: txCount}

		if i%100000 == 0 {
			fmt.Printf("received counts for %d addresses", i)
		}
	}

	for _, a := range activities {
		// column 1: cluster or address uid
		// column 2: entity type
		// column 3: number of transactions created by this entity
		line := []string{a.uid, "address", strconv.Itoa(a.transactionCount)}

		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}

	w.Flush()

	// handle clusters
	clusters, err := analytics.GetAllFMIClusters(dgraph)
	if err != nil {
		warn(err)
		return
	}

	activities = make([]clusterActivity, len(clusters))
	for i, c := range clusters {
		txCount, err := analytics.GetTransactionCountPerCluster(dgraph, c)
		if err != nil {
			warn(err)
			return
		}

		activities[i] = clusterActivity{uid: c, transactionCount: txCount}

		if i%100000 == 0 {
			fmt.Printf("received counts for %d clusters", i)
		}
	}

	for _, a := range activities {
		// column 1: cluster or address uid
		// column 2: entity type
		// column 3: number of transactions created by this entity
		line := []string{a.uid, "cluster", strconv.Itoa(a.transactionCount)}

		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}

	w.Flush()
}
