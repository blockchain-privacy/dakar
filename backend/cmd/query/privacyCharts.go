package main

import (
	"backend/constants"
	dban "backend/db/analytics"
	"backend/db/status"
	"backend/external"
	"context"
	"encoding/csv"
	"github.com/qrest/gomisc/serror"
	"os"
	"time"
)

type transactionTypePair struct {
	label           string
	transactionType string
}

// exportTransactionData exports all transaction timestamps in a CSV-file per transaction type.
func exportTransactionData(ctx context.Context, database external.Database, directory string) {
	info("Creating privacy transaction charts")
	if len(directory) == 0 {
		info("invalid directory: " + directory)
		return
	}

	meta, err := status.GetMeta(ctx, database)
	if err != nil {
		warn(err)
		return
	}

	var transactionTypes []transactionTypePair
	switch meta.BlockchainMode {
	case constants.BlockchainModeDash:
		transactionTypes = []transactionTypePair{
			{transactionType: constants.TypeDashMixing},
			{transactionType: constants.TypeDashOrigin},
			{transactionType: constants.TypeDashDestination},
			{transactionType: constants.TypeDashCC},
			{transactionType: constants.TypeDashCP},
			{label: "all", transactionType: ""},
		}
	case constants.BlockchainModeBTC:
		transactionTypes = []transactionTypePair{
			{transactionType: constants.TypeWasabi2Origin},
			{transactionType: constants.TypeWasabi2Mixing},
			{transactionType: constants.TypeWasabi2Destination},
			{transactionType: constants.TypeWhirlpoolOrigin},
			{transactionType: constants.TypeWhirlpoolMixing},
			{transactionType: constants.TypeWhirlpoolDestination},
			{label: "all", transactionType: ""},
		}
	default:
		warn(serror.FromStrWithContext("invalid blockchain mode", "mode", meta.BlockchainMode))
		return
	}

	for _, t := range transactionTypes {
		ts, err := dban.GetTransactionTypeData(ctx, database, t.transactionType)
		if err != nil {
			warn(err)
			return
		}

		if len(ts) == 0 {
			info("no transactions found, continuing with next transaction type", "transaction type", t.transactionType)
			continue
		}

		label := t.label
		if label == "" {
			label = t.transactionType
		}

		writeTimestampsToCSV(directory+"/"+label+".csv", ts)
	}
}

func writeTimestampsToCSV(fileName string, txs []time.Time) {
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
		line := []string{t.Format(time.RFC3339)}
		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}
}
