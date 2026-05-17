// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	mgraph "backend/analytics/graph"
	"backend/db"
	"backend/db/analytics"
	"backend/external"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type exportTransaction struct {
	outputTimestamps []time.Time
	id               int64
	ts               time.Time
}

// doDestinationTimestampAnalysis writes all destination transactions to a CSV file
func doDestinationTimestampAnalysis(g *mgraph.ReversibleGraph, transactionType string, filename string) {
	info(fmt.Sprintf("export of '%s' transactions starting", transactionType))
	graphTransactions := getGraphTransactions(g, transactionType)
	info(fmt.Sprintf("number of '%s' transactions in graph %d", transactionType, len(graphTransactions)))

	if len(graphTransactions) == 0 {
		return
	}

	writeTxToCSV(filename, graphTransactions)
}

func writeTxToCSV(fileName string, txs []exportTransaction) {
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
		// column 1: destination uid
		// column 2: destination timestamp
		// column 3-n: input timestamps
		line := []string{mgraph.ToHex(t.id), strconv.FormatInt(t.ts.Unix(), 10)}
		for _, t := range t.outputTimestamps {
			line = append(line, strconv.FormatInt(t.Unix(), 10))
		}

		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}
}

func getGraphTransactions(g *mgraph.ReversibleGraph, transactionType string) []exportTransaction {
	year2018, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		warn(err, "msg", "error while creating date")
		return nil
	}

	var destinations []exportTransaction

	nodes := g.Nodes()
	for nodes.Next() {
		node := nodes.Node()

		txNode, ok := node.(mgraph.TransactionNode)
		if !ok || txNode.Type != transactionType || txNode.TS.Before(year2018) {
			continue
		}

		fromNodes := g.From(node.ID())
		timestamps := make([]time.Time, fromNodes.Len())
		i := 0
		for fromNodes.Next() {
			timestamps[i] = fromNodes.Node().(mgraph.TransactionNode).TS
			i++
		}

		destinations = append(destinations, exportTransaction{
			outputTimestamps: timestamps,
			id:               node.ID(),
			ts:               txNode.TS,
		})
	}

	return destinations
}

func doExportBlocks(ctx context.Context, dgraph external.Database, fileName string, startBlock int, endBlock int) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		warn(err, "msg", "error creating file")
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	blockRange, err := getBlockRange(ctx, dgraph, startBlock, endBlock)
	if err != nil {
		warn(err, "msg", "error getting blocks")
		return
	}

	if len(blockRange) == 0 {
		info("no blocks to write")
		return
	}

	addressRange, err := getAddressRange(ctx, dgraph, startBlock, endBlock)
	if err != nil {
		warn(err, "msg", "error getting addresses")
		return
	}

	if len(addressRange) == 0 {
		info("no addresses to write")
		return
	}

	clusterRange, err := getClusterRange(ctx, dgraph, startBlock, endBlock)
	if err != nil {
		warn(err, "msg", "error getting clusters")
		return
	}

	// merge addresses and blocks
	toEncode := make([]any, len(blockRange)+len(addressRange)+len(clusterRange))

	for i, b := range blockRange {
		toEncode[i] = b
	}

	for i, a := range addressRange {
		toEncode[i+len(blockRange)] = a
	}

	for i, c := range clusterRange {
		toEncode[i+len(blockRange)+len(addressRange)] = c
	}

	if err = json.NewEncoder(file).Encode(toEncode); err != nil {
		warn(err, "msg", "error encoding data")
		return
	}
}

func doExportPrivacyGraph(ctx context.Context, dgraph external.Database, fileName string, startTransaction string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		warn(err, "msg", "error creating file")
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	blocks, addresses, transactions, err := analytics.GetForwardLookupTransactions(ctx, dgraph, startTransaction)
	if err != nil {
		return
	}

	if len(addresses) == 0 {
		info("no addresses to write")
		return
	}

	if len(blocks) == 0 {
		info("no blocks to write")
		return
	}

	if len(transactions) == 0 {
		info("no transactions to write")
		return
	}

	// merge addresses and blocks
	toEncode := make([]any, 0, len(blocks)+len(addresses)+len(transactions))
	for _, b := range blocks {
		toEncode = append(toEncode, b)
	}

	for _, a := range addresses {
		toEncode = append(toEncode, a)
	}

	for _, t := range transactions {
		toEncode = append(toEncode, t)
	}

	if err = json.NewEncoder(file).Encode(toEncode); err != nil {
		warn(err, "msg", "error encoding data")
		return
	}
}

func doExportTransactions(ctx context.Context, dgraph external.Database,
	fileName string, startBlock int64, endBlock int64, transactionTypes []string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		warn(err, "msg", "error creating file")
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	transactions, err := db.GetTransactionsByBlock(ctx, dgraph, startBlock, endBlock, transactionTypes)
	if err != nil {
		warn(err)
		return
	}

	if err = json.NewEncoder(file).Encode(transactions); err != nil {
		warn(err, "msg", "error encoding data")
		return
	}
}
