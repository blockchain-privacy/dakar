package main

import (
	"backend/constants"
	dban "backend/db/analytics"
	"backend/external"

	"encoding/csv"
	"os"
	"time"
)

type privacyTypePair struct {
	label string
	start string
	stop  string
}

// exportTransactionData exports all transaction timestamps in a CSV-file per privacy type (mixing, destination, ...)
func exportTransactionData(database external.Database, directory string) {
	info("Creating privacy transaction charts")
	if len(directory) == 0 {
		info("invalid directory:", directory)
		return
	}

	privacyTypes := []privacyTypePair{
		{label: "mixing", start: "0", stop: constants.StrPrivacyMixingLast},
		// {label: "mixing 0", start: constants.StrPrivacyMixing0, stop: constants.StrPrivacyMixing0},
		// {label: "mixing 1", start: constants.StrPrivacyMixing1, stop: constants.StrPrivacyMixing1},
		// {label: "mixing 2", start: constants.StrPrivacyMixing2, stop: constants.StrPrivacyMixing2},
		// {label: "mixing 3", start: constants.StrPrivacyMixing3, stop: constants.StrPrivacyMixing3},
		// {label: "mixing 4", start: constants.StrPrivacyMixing4, stop: constants.StrPrivacyMixing4},
		{label: "origin", start: constants.StrPrivacyOriginFirst, stop: constants.StrPrivacyOriginLast},
		{label: "destination", start: constants.StrPrivacyDestinationFirst,
			stop: constants.StrPrivacyDestinationLast},
		{label: "collateral creation", start: constants.StrPrivacyCollateralCreationFirst,
			stop: constants.StrPrivacyCollateralCreationLast},
		{label: "collateral payment", start: constants.StrPrivacyCollateralPaymentFirst,
			stop: constants.StrPrivacyCollateralPaymentLast},
		{label: "all", start: "0", stop: "500"},
	}

	for _, privacyType := range privacyTypes {
		ts, dbErr := dban.GetPrivacyTypeData(database, privacyType.start, privacyType.stop)
		if dbErr != nil {
			warn(dbErr)
			return
		}
		writeTimestampsToCSV(directory+"/"+privacyType.label, ts)
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
			info("error writing record to file", err)
			return
		}
	}
}
