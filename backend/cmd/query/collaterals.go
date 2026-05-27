// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db/analytics"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
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

func doCollateralGapAnalysis(ctx context.Context, c external.Database, filename string) {
	info("collateral gap analysis starting")
	if filename == "" {
		warn(serror.FromStr("file name is empty"))
		return
	}

	transactions, err := analytics.GetCollateralPaymentTimestamps(ctx, c)
	if err != nil {
		warn(err)
		return
	}

	writeTimestampsCountToCSV(filename, transactions)
}

func writeTimestampsCountToCSV(fileName string, transactions []analytics.CollateralPaymentTimestamps) {
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

	for _, transaction := range transactions {
		// first column is transaction hash
		// second column is transaction timestamp
		// third column is transaction input timestamp
		line := []string{transaction.TransactionHash, transaction.Timestamp.UTC().Format(time.RFC3339),
			transaction.InputTimestamp.UTC().Format(time.RFC3339)}
		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}
}
