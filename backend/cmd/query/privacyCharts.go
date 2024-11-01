package main

import (
	"backend/constants"
	dban "backend/db/analytics"
	"backend/external"
	"context"
	"encoding/csv"
	"os"
	"time"
)

type transactionTypePair struct {
	label           string
	transactionType string
}

// exportTransactionData exports all transaction timestamps in a CSV-file per privacy type (mixing, destination, ...)
func exportTransactionData(ctx context.Context, database external.Database, directory string) {
	info("Creating privacy transaction charts")
	if len(directory) == 0 {
		info("invalid directory: " + directory)
		return
	}

	transactionTypes := []transactionTypePair{
		{label: "mixing", transactionType: constants.TypeDashMixing},
		{label: "origin", transactionType: constants.TypeDashOrigin},
		{label: "destination", transactionType: constants.TypeDashDestination},
		{label: "collateral creation", transactionType: constants.TypeDashCC},
		{label: "collateral payment", transactionType: constants.TypeDashCP},
		{label: "all", transactionType: ""},
	}

	for _, t := range transactionTypes {
		ts, dbErr := dban.GetTransactionTypeData(ctx, database, t.transactionType)
		if dbErr != nil {
			warn(dbErr)
			return
		}
		writeTimestampsToCSV(directory+"/"+t.label, ts)
	}
}

func writeTimestampsToCSV(fileName string, txs []time.Time) {
	f, err := os.Create(fileName + ".csv")
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
		line := []string{t.Format(time.RFC3339)}
		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}
}
