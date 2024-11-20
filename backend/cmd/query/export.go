package main

import (
	mgraph "backend/analytics/graph"
	"backend/constants"
	"backend/db/analytics"
	"backend/external"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
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
func doDestinationTimestampAnalysis(g *mgraph.ReversibleGraph) {
	info("export of destination transactions starting")
	dstTransactions := getDestinationTransactions(g)
	info(fmt.Sprintf("number of destination transactions in graph %d", len(dstTransactions)))

	writeTxToCSV("destination_transactions", dstTransactions)

	info("done")
}

// doDestinationTimestampAnalysis writes all mixing transactions to a CSV file
func exportMixingTimestamps(g *mgraph.ReversibleGraph, getInputs bool) {
	info("export of mixing transactions starting")

	mixingTransactions := getMixingTransactions(g, getInputs)
	info(fmt.Sprintf("number of mixing transactions in graph %d", len(mixingTransactions)))

	writeTxToCSV("mixing_transactions", mixingTransactions)

	info("done")
}

func writeTxToCSV(fileName string, txs []exportTransaction) {
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
		// first column is destination id
		// second column is timestamp of this transaction
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

func getDestinationTransactions(g *mgraph.ReversibleGraph) []exportTransaction {
	var destinations []exportTransaction

	nodes := g.Nodes()
	for nodes.Next() {
		node := nodes.Node()
		if g.To(node.ID()).Len() == 0 {
			txNode, ok := node.(mgraph.TransactionNode)
			if !ok {
				continue
			}

			fromNodes := g.From(node.ID())

			var timestamps []time.Time

			for fromNodes.Next() {
				timestamps = append(timestamps, fromNodes.Node().(mgraph.TransactionNode).TS)
			}

			destinations = append(destinations, exportTransaction{
				outputTimestamps: timestamps,
				id:               node.ID(),
				ts:               txNode.TS,
			})
		}
	}
	return destinations
}

// getMixingTransactions returns all mixing transaction in the graph.
// Depending on the getInputs flag, either the input or
// output timestamps are included
func getMixingTransactions(g *mgraph.ReversibleGraph, getInputs bool) []exportTransaction {
	year2016, err := time.Parse("2006-01-02", "2016-01-01")
	if err != nil {
		warn(err, "msg", "error while creating date")
		return nil
	}

	nodes := g.Nodes()
	mixingTransactions := make([]exportTransaction, 0, nodes.Len())
	for nodes.Next() {
		txNode, ok := nodes.Node().(mgraph.TransactionNode)
		if !ok || txNode.Type != constants.TypeDashMixing || txNode.TS.Before(year2016) {
			continue
		}

		var reachableNodes graph.Nodes
		if getInputs {
			reachableNodes = g.To(txNode.ID())
		} else {
			reachableNodes = g.From(txNode.ID())
		}

		var timestamps []time.Time

		for reachableNodes.Next() {
			timestamps = append(timestamps, reachableNodes.Node().(mgraph.TransactionNode).TS)
		}

		mixingTransactions = append(mixingTransactions, exportTransaction{
			outputTimestamps: timestamps,
			id:               txNode.ID(),
			ts:               txNode.TS,
		})
	}
	return mixingTransactions
}

// exportReverseLookup writes the mixing transactions including their timestamps to a CSV-file
func exportReverseLookup(g *mgraph.ReversibleGraph, nodeIDStr string,
	maxLookBackTime int, addressExclusions []string, getInputs bool, checkSpendingGaps bool) {
	nodeID, err := mgraph.ToInteger(nodeIDStr)
	if err != nil {
		warn(err)
		return
	}

	node := g.Node(nodeID)
	if node == nil {
		warn(mgraph.ErrNodeNotFound(nodeID))
	}

	nodeTS := node.(mgraph.TransactionNode).TS

	isReversed := g.IsReversed()

	exclusionsMap := make(map[int64]bool, len(addressExclusions))

	for _, e := range addressExclusions {
		integer, err := mgraph.ToInteger(e)
		if err != nil {
			warn(err)
			return
		}

		exclusionsMap[integer] = true
	}

	spendingGapCounter := 0

	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			if !mgraph.CheckAddressExclusions(exclusionsMap, e.(mgraph.AddressEdge)) {
				return false
			}

			if checkSpendingGaps && !mgraph.HasSpendingGap(g, e.(mgraph.AddressEdge)) {
				spendingGapCounter++
				return false
			}

			// get node to which the edge leads
			toNode := g.Node(e.To().ID()).(mgraph.TransactionNode)

			// if a maximum look back time is set check the timestamp
			if maxLookBackTime > 0 {
				// isReversed is true if it is a forward lookup: default case is a reverse
				// lookup so if the graph is reversed a forward lookup is happening
				if isReversed {
					if toNode.TS.Sub(nodeTS) > time.Duration(maxLookBackTime) {
						return false
					}
				} else if nodeTS.Sub(toNode.TS) > time.Duration(maxLookBackTime) {
					return false
				}
			}

			// if it is not a mixing transaction save it and stop following that edge
			if toNode.Type != constants.TypeDashMixing {
				return false
			}

			// true: follow this link
			// false: do not follow this link
			return true
		},
	}

	var exportTransactions []exportTransaction

	w.Walk(g, node, func(n graph.Node, _ int) bool {
		var reachableNodes graph.Nodes
		if getInputs {
			reachableNodes = g.To(n.ID())
		} else {
			reachableNodes = g.From(n.ID())
		}

		// collect mixing nodes start
		txNode, ok := n.(mgraph.TransactionNode)
		if !ok || txNode.Type != constants.TypeDashMixing {
			return false
		}

		var timestamps []time.Time
		for reachableNodes.Next() {
			timestamps = append(timestamps, reachableNodes.Node().(mgraph.TransactionNode).TS)
		}

		exportTransactions = append(exportTransactions, exportTransaction{
			outputTimestamps: timestamps,
			id:               n.ID(),
			ts:               txNode.TS,
		})
		// collect mixing nodes end

		// true: stop traversing nodes
		// false: do not stop traversing nodes
		return false
	})
	info(fmt.Sprintf("number of edges not traversed due to spending gap: %d", spendingGapCounter))

	writeTxToCSV(mgraph.ToHex(nodeID)+"_mixing_transactions", exportTransactions)
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

	err = json.NewEncoder(file).Encode(toEncode)
	if err != nil {
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

	err = json.NewEncoder(file).Encode(toEncode)
	if err != nil {
		warn(err, "msg", "error encoding data")
		return
	}
}
