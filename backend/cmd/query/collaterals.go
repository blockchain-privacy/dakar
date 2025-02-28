package main

import (
	"backend/constants"
	"backend/db/analytics"
	"backend/external"
	"context"
	"encoding/csv"
	"github.com/qrest/gomisc/serror"
	"os"
	"strconv"
)

func doCollateralAnalysis(ctx context.Context, c external.Database, cpFileName string, ccFileName string) {
	info("collateral analysis starting")
	if cpFileName == "" || ccFileName == "" {
		warn(serror.FromStr("file name is empty"))
		return
	}

	cpCounts, err := analytics.GetOutputCountsPerAddress(ctx, c, constants.TypeDashCP)
	if err != nil {
		warn(err)
		return
	}

	ccCounts, err := analytics.GetOutputCountsPerAddress(ctx, c, constants.TypeDashCC)
	if err != nil {
		warn(err)
		return
	}

	writeOutputCountToCSV(cpFileName, cpCounts)
	writeOutputCountToCSV(ccFileName, ccCounts)
}

func writeOutputCountToCSV(fileName string, counts []analytics.AddressOutputCount) {
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

	for _, count := range counts {
		// first column is address hash
		// second column is output count
		line := []string{count.AddressHash, strconv.Itoa(count.OutputCount)}
		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}
}
