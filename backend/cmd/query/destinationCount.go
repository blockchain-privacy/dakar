package main

import (
	"backend/db/analytics"
	"backend/external"
	"encoding/csv"
	"os"
	"strconv"
)

func doDestinationCountAnalysis(dgraph external.Database, fileName string) {
	spenders, err := analytics.GetDestinationTransactionSpenders(dgraph)
	if err != nil {
		info(err)
		return
	}

	writeSpendersToCSV(fileName, spenders)
}

func writeSpendersToCSV(fileName string, txs []analytics.SpenderTransaction) {
	f, err := os.Create(fileName)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			info(err)
		}
	}(f)

	if err != nil {
		info(err)
		return
	}

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, t := range txs {
		// column 1: transaction hash of spender transaction
		// column 2: cluster size of spender transaction inputs
		// column 3-x: destination transactions which have been spent by tx in column 1
		line := []string{t.TxHash, strconv.Itoa(t.ClusterSize)}
		line = append(line, t.Destinations...)

		if err := w.Write(line); err != nil {
			info("error writing record to file", err)
			return
		}
	}
}
