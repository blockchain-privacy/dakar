// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/dakar/db/analytics"
	"gitlab.com/blockchain-privacy/dakar/external"
)

func doExportClusterActivity(ctx context.Context, dgraph external.Database, fileName string) {
	// handle clusters
	clusters, err := analytics.GetAllFMIClusters(ctx, dgraph)
	if err != nil {
		warn(err)
		return
	}

	fmt.Println("number of clusters", len(clusters))

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

	now := time.Now()
	for i, c := range clusters {
		inputCount, outputCount, err := analytics.GetTransactionCountPerCluster(ctx, dgraph, c)
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
