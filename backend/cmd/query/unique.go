package main

import (
	dban "backend/db/analytics"
	"backend/db/status"
	"backend/external"
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

func doUniqueAddressAnalysis(database external.Database, option int, fileName string) {

	// get the highest clustered block ID
	fmiStatus, err := status.GetClusteringFMIStatus(database)
	if err != nil {
		info(err)
		return
	}

	// get start date
	startDateStr, err := dban.BlockHeightToTimestamp(database, 1)
	if err != nil {
		return
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		info(err)
		return
	}

	fromDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(),
		0, 0, 0, 0, startDate.Location())

	// get end date
	endDateStr, err := dban.BlockHeightToTimestamp(database, *fmiStatus.LastClusteredBlockID)
	if err != nil {
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		info(err)
		return
	}

	toDate := time.Date(endDate.Year(), endDate.Month(), endDate.Day(),
		0, 0, 0, 0, endDate.Location())

	f, err := os.Create(fileName + "_option" + strconv.Itoa(option) + ".csv")
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

	// write header
	header := []string{"Timestamp", "Addresses", "Clusters", "Addresses Without Cluster"}
	if err := w.Write(header); err != nil {
		info("error writing header to file", err)
		return
	}

	// write data
	for i := fromDate.UTC(); toDate.UTC().After(i); i = i.Add(time.Hour * 24) {
		addressCount, clusterCount, addressesWithClusterCount, err :=
			dban.GetUniqueAddressCountsPerBlock(database, i, option)
		if err != nil {
			info(err)
			return
		}

		line := []string{i.Format(time.RFC3339), strconv.FormatUint(addressCount, 10),
			strconv.FormatUint(clusterCount, 10),
			strconv.FormatUint(addressesWithClusterCount, 10)}

		if err := w.Write(line); err != nil {
			info("error writing record to file", err)
			return
		}

		info("processed", i)
		w.Flush()
	}
}
