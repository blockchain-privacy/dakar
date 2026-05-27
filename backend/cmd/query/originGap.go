// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"

	mgraph "gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
)

func doOriginGapAnalysis(g *mgraph.ReversibleGraph, minGap time.Duration, filename string) {
	info("origin gap analysis starting")
	getSpendingGapTransactions(g, minGap, filename)
}

// getSpendingGapTransactions returns all mixing transactions including their spending gap information
func getSpendingGapTransactions(g *mgraph.ReversibleGraph, minGap time.Duration, filename string) {
	year2016, err := time.Parse("2006-01-02", "2016-01-01")
	if err != nil {
		warn(err)
		return
	}

	f, err := os.Create(filename)
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

	nodes := g.Nodes()
	for nodes.Next() {
		txNode, ok := nodes.Node().(mgraph.TransactionNode)
		if !ok || txNode.Type != constants.TypeDashMixing || txNode.TS.Before(year2016) {
			continue
		}

		to := g.To(txNode.ID())
		hasSpendingGap := false
		for to.Next() {
			node := to.Node()
			if node.(mgraph.TransactionNode).TS.Sub(txNode.TS) >= minGap {
				hasSpendingGap = true
				break
			}
		}

		distances, err := getOriginDistance(g, txNode.ID())
		if err != nil {
			warn(err)
			return
		}

		line := []string{mgraph.ToHex(txNode.ID()), strconv.FormatBool(hasSpendingGap)}

		for _, d := range distances {
			line = append(line, strconv.Itoa(d))
		}

		if err := w.Write(line); err != nil {
			warn(err, "msg", "error writing record to file")
			return
		}
	}
}

// getOriginDistance returns the distance between the given node and the first origin found in breadth-first search
func getOriginDistance(g *mgraph.ReversibleGraph, nodeID int64) (distances []int, err error) {
	node := g.Node(nodeID)
	if node == nil {
		err = mgraph.ErrNodeNotFound(nodeID)
		return
	}

	w := traverse.BreadthFirst{
		Traverse: func(_ graph.Edge) bool {
			// true: follow this link
			// false: do not follow this link
			return true
		},
	}

	maxDepth := 6
	numDistances := 40
	distances = make([]int, numDistances)
	i := 0
	w.Walk(g, node, func(n graph.Node, depth int) bool {
		if depth > maxDepth {
			// maximal depth reached
			return true
		}

		from := g.From(n.ID())
		if from.Len() == 0 {
			distances[i] = depth

			if i+1 == numDistances {
				// found enough origin distances
				return true
			}
			i++
		}

		// true: stop traversing nodes
		// false: do not stop traversing nodes
		return false
	})

	return
}
