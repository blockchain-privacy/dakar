package graph

import (
	"encoding/csv"
	"gonum.org/v1/gonum/graph"
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
func doDestinationTimestampAnalysis(g *ReversibleGraph) {
	info("export of destination transactions starting")
	dstTransactions := getDestinationTransactions(g)
	info("number of destination transactions in graph", len(dstTransactions))

	writeTxToCSV("destination_transactions", dstTransactions)

	info("done")
}

// doDestinationTimestampAnalysis writes all mixing transactions to a CSV file
func exportMixingTimestamps(g *ReversibleGraph) {
	info("export of mixing transactions starting")

	mixingTransactions := getMixingTransactions(g, false)
	info("number of mixing transactions in graph", len(mixingTransactions))

	writeTxToCSV("mixing_transactions", mixingTransactions)

	info("done")
}

func writeTxToCSV(fileName string, txs []exportTransaction) {
	f, err := os.Create(fileName + ".csv")
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

	for _, t := range txs {
		// first column is destination id
		// second column is timestamp of this transaction
		line := []string{toHex(t.id), strconv.FormatInt(t.ts.Unix(), 10)}
		for _, t := range t.outputTimestamps {
			line = append(line, strconv.FormatInt(t.Unix(), 10))
		}

		if err := w.Write(line); err != nil {
			info("error writing record to file", err)
			return
		}
	}
}

func getDestinationTransactions(g *ReversibleGraph) []exportTransaction {
	var destinations []exportTransaction

	for _, node := range g.nodes {
		if g.To(node.ID()).Len() == 0 {
			txNode, ok := node.(transactionNode)
			if !ok {
				continue
			}

			fromNodes := g.From(node.ID())

			var timestamps []time.Time

			for fromNodes.Next() {
				timestamps = append(timestamps, fromNodes.Node().(transactionNode).ts)
			}

			destinations = append(destinations, exportTransaction{
				outputTimestamps: timestamps,
				id:               node.ID(),
				ts:               txNode.ts,
			})
		}
	}
	return destinations
}

// getMixingTransactions returns all mixing transaction in the graph.
// Depending on the getInputs flag, either the input or
// output timestamps are included
func getMixingTransactions(g *ReversibleGraph, getInputs bool) []exportTransaction {
	var mixingTransactions []exportTransaction
	year2016, err := time.Parse("2006-01-02", "2016-01-01")
	if err != nil {
		info("error while creating date", err)
		return nil
	}
	for _, node := range g.nodes {
		txNode, ok := node.(transactionNode)
		if !ok || !txNode.privacyType.IsMixing() || txNode.ts.Before(year2016) {
			continue
		}

		var reachableNodes graph.Nodes
		if getInputs {
			reachableNodes = g.From(node.ID())
		} else {
			reachableNodes = g.To(node.ID())
		}

		var timestamps []time.Time

		for reachableNodes.Next() {
			timestamps = append(timestamps, reachableNodes.Node().(transactionNode).ts)
		}

		mixingTransactions = append(mixingTransactions, exportTransaction{
			outputTimestamps: timestamps,
			id:               node.ID(),
			ts:               txNode.ts,
		})

	}
	return mixingTransactions
}
