// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	dban "backend/db/analytics"
	"backend/db/status"
	"backend/external"
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

func doUniqueAddressAnalysis(ctx context.Context, database external.Database, option int, fileName string) {
	info("Starting unique address analysis")

	// get the highest clustered block ID
	fmiStatus, err := status.GetClusteringFMIStatus(ctx, database)
	if err != nil {
		warn(err)
		return
	}

	// get start date
	startDateStr, err := dban.BlockHeightToTimestamp(ctx, database, 1)
	if err != nil {
		return
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		warn(err)
		return
	}

	fromDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(),
		0, 0, 0, 0, startDate.Location())

	// get end date
	endDateStr, err := dban.BlockHeightToTimestamp(ctx, database, *fmiStatus.LastClusteredBlockID)
	if err != nil {
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		warn(err)
		return
	}

	toDate := time.Date(endDate.Year(), endDate.Month(), endDate.Day(),
		0, 0, 0, 0, endDate.Location())

	f, err := os.Create(fileName + "_option" + strconv.Itoa(option) + ".csv")
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

	// write header
	header := []string{"Timestamp", "Addresses", "Clusters", "Addresses Without Cluster"}
	if err := w.Write(header); err != nil {
		warn(err, "msg", "error writing header to file")
		return
	}

	// write data
	for i := fromDate.UTC(); toDate.UTC().After(i); i = i.Add(time.Hour * 24) {
		addressCount, clusterCount, addressesWithClusterCount, err :=
			dban.GetUniqueAddressCountsPerBlock(ctx, database, i, option)
		if err != nil {
			warn(err)
			return
		}

		line := []string{i.Format(time.RFC3339), strconv.Itoa(addressCount), strconv.Itoa(clusterCount),
			strconv.Itoa(addressesWithClusterCount)}

		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}

		info("processed", i)
		w.Flush()
	}
}
