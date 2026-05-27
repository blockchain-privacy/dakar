// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"encoding/csv"
	"os"
	"strconv"

	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/status"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

func doStats(ctx context.Context, dgraph external.Database, fileName string, excludeTransactionType string) {
	if fileName == "" {
		warn(serror.FromStr("file name is empty"))
		return
	}

	crawlerStatus, err := status.GetCrawlerStatus(ctx, dgraph)
	if err != nil {
		warn(err)
		return
	}

	if crawlerStatus.LastBlockID == nil {
		warn(serror.FromStr("last block is nil "))
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

	// load blocks in batches from db
	const steps = 100
	stop := false
	for i := int64(0); !stop; i += steps {
		to := i + steps - 1
		if to >= *crawlerStatus.LastBlockID {
			to = *crawlerStatus.LastBlockID
			stop = true
		}

		counts, err := db.GetTransactionOutputCounts(ctx, dgraph, i, to, excludeTransactionType)
		if err != nil {
			warn(err)
			return
		}

		for _, txCount := range counts {
			// column 1: transaction hash
			// column 2: number of inputs
			// column 3: number of outputs
			line := []string{txCount.Hash, strconv.Itoa(txCount.InputCount), strconv.Itoa(txCount.OutputCount)}
			if err := w.Write(line); err != nil {
				warn(err, "msg", "error writing record to file")
				return
			}
		}

		w.Flush()

		if i%10000 == 0 {
			info("stats processed for block", "count", to)
		}
	}
}
