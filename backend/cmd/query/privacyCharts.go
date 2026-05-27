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
	dban "gitlab.com/blockchain-privacy/dakar/db/analytics"
	"gitlab.com/blockchain-privacy/dakar/db/status"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// exportTransactionData exports all transaction timestamps in a CSV-file per transaction type.
func exportTransactionData(ctx context.Context, database external.Database, directory string) {
	info("Creating transaction type charts")
	if len(directory) == 0 {
		info("invalid directory: " + directory)
		return
	}

	meta, err := status.GetMeta(ctx, database)
	if err != nil {
		warn(err)
		return
	}

	var transactionTypes []string
	switch meta.BlockchainMode {
	case constants.BlockchainModeDash:
		transactionTypes = []string{constants.TypeDashMixing, constants.TypeDashOrigin,
			constants.TypeDashDestination, constants.TypeDashCC, constants.TypeDashCP}
	case constants.BlockchainModeBTC:
		transactionTypes = []string{constants.TypeWasabi2Origin, constants.TypeWasabi2Mixing,
			constants.TypeWasabi2Destination, constants.TypeWhirlpoolOrigin,
			constants.TypeWhirlpoolMixing, constants.TypeWhirlpoolDestination,
		}
	default:
		warn(serror.FromStrWithContext("invalid blockchain mode", "mode", meta.BlockchainMode))
		return
	}

	for _, t := range transactionTypes {
		ts, counts, err := dban.GetTransactionTypeData(ctx, database, t)
		if err != nil {
			warn(err)
			return
		}

		if len(ts) == 0 {
			info("no transactions found, continuing with next transaction type", "transaction type", t)
			continue
		}

		if len(ts) != len(counts) {
			warn(serror.FromStr("number of returned timestamps is different than number of outputs"))
			return
		}

		writeTimestampsToCSV(directory+"/"+t+".csv", ts, counts)
	}
}

func writeTimestampsToCSV(fileName string, txs []time.Time, counts []int) {
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

	for i, t := range txs {
		line := []string{t.Format(time.RFC3339), strconv.Itoa(counts[i])}
		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}
}
