// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"backend/constants"
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/gomisc/serror"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
)

func ErrNodeNotFound(nodeID int64) error {
	return serror.FromFormat("node %s does not exist in graph", ToHex(nodeID))
}

// ToHex returns a hexadecimal string representation of the given integer with the '0x' prefix.
// Does not handle negative integers.
func ToHex(i int64) string {
	return "0x" + strconv.FormatInt(i, 16)
}

// ToInteger a hex string in the form of "0x123" to an integer. For performance reasons, there is no strong input validation.
func ToInteger(hexString string) (int64, error) {
	if len(hexString) < 3 {
		return 0, serror.FromFormat("string to short: %s", hexString)
	}
	integer, err := strconv.ParseInt(hexString[2:], 16, 64)
	if err != nil {
		return 0, serror.New(err)
	}
	return integer, nil
}

// HasSpendingGap returns true if the spending gap is bigger than 8 hours
func HasSpendingGap(g *ReversibleGraph, edge AddressEdge) bool {
	fromTS := g.Node(edge.F.ID()).(TransactionNode).TS
	toTS := g.Node(edge.T.ID()).(TransactionNode).TS

	if g.reversed {
		return toTS.Sub(fromTS) >= time.Hour*8
	}

	return fromTS.Sub(toTS) >= time.Hour*8
}

// ReverseLookupByID performs a reverse lookup starting at the given nodeId
// Can be limited by time (maxLookBackTime) and number of  hops (maxDepth)
// Set maxLookBackTime and maxDepth to 0 respectively if they should not be considered.
func ReverseLookupByID(g *ReversibleGraph, nodeID int64, maxLookBackTime time.Duration, maxDepthParameter int,
	excludeSpendingGaps bool) (map[string]bool, error) {
	node := g.Node(nodeID)
	if node == nil {
		return nil, ErrNodeNotFound(nodeID)
	}

	if maxDepthParameter < 0 {
		return nil, serror.FromStrWithContext("invalid maximum depth", "depth", maxDepthParameter)
	}

	foundEndpoints := make(map[string]bool)
	nodeTS := node.(TransactionNode).TS
	isReversed := g.IsReversed()
	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			addressEdge := e.(AddressEdge)
			if excludeSpendingGaps && HasSpendingGap(g, addressEdge) {
				return false
			}

			// get node to which the edge leads
			toNode := g.Node(e.To().ID()).(TransactionNode)

			// if a maximum look back time is set check the timestamp
			if maxLookBackTime > 0 {
				// isReversed is true if it is a forward lookup: default case is a reverse
				// lookup so if the graph is reversed a forward lookup is happening
				if isReversed {
					if toNode.TS.Sub(nodeTS) > maxLookBackTime {
						// The forward lookup starts at an origin transaction and looks forward
						// for a certain user-defined duration. Funds from mixing transaction are usually
						// not spent directly after the mixing process is finished. Because of this, the next
						// connected non-mixing transaction must be included to the returned set.
						if !constants.IsMixingTransaction(toNode.Type) {
							foundEndpoints[toNode.String()] = true
						}

						return false
					}
				} else if nodeTS.Sub(toNode.TS) > maxLookBackTime {
					return false
				}
			}

			// if it is not a mixing transaction save it and stop following that edge
			if !constants.IsMixingTransaction(toNode.Type) {
				foundEndpoints[toNode.String()] = true
				return false
			}

			// true: follow this link
			// false: do not follow this link
			return true
		},
	}

	maxDepth := maxDepthParameter
	if maxDepthParameter > 0 {
		if !isReversed {
			// Reduce the maximum depth by 1, because in Traverse() connected nodes are also considered.
			// Therefore, we traverse: maxDepth + 1 - 1 = maxDepth
			// This is limited to reverse lookups. When doing forward lookups we want to traverse
			// one additional hop, which is achieved in Traverse()
			maxDepth--
		}
	}

	w.Walk(g, node, func(n graph.Node, depth int) bool {
		if maxDepthParameter > 0 && depth > maxDepth {
			return true
		}

		if g.From(n.ID()).Len() == 0 && !constants.IsMixingTransaction(n.(TransactionNode).Type) {
			thisNode := n.(TransactionNode)
			foundEndpoints[thisNode.String()] = true
		}

		// true: stop traversing nodes
		// false: do not stop traversing nodes
		return false
	})

	return foundEndpoints, nil
}

// ReverseLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph backward
func ReverseLookup(g *ReversibleGraph, uid string, maxLookBackTime time.Duration, maxDepth int,
	excludeSpendingGaps bool) (map[string]bool, error) {
	nodeUID, err := ToInteger(uid)
	if err != nil {
		return nil, err
	}
	g.SetReverse(false)

	return ReverseLookupByID(g, nodeUID, maxLookBackTime, maxDepth, excludeSpendingGaps)
}

// ForwardLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph forward
func ForwardLookup(g *ReversibleGraph, uid string, maxLookForwardTime time.Duration, maxDepth int,
	excludeSpendingGaps bool) (map[string]bool, error) {
	nodeUID, err := ToInteger(uid)
	if err != nil {
		return nil, err
	}

	g.SetReverse(true)
	origins, err := ReverseLookupByID(g, nodeUID, maxLookForwardTime, maxDepth, excludeSpendingGaps)
	if err != nil {
		return nil, err
	}
	g.SetReverse(false)
	return origins, err
}

// getConnectedNodes returns all nodes (including itself) in nodeSet which
// can be reached via nodeID in both traversal directions.
func getConnectedNodes(g *ReversibleGraph, nodeUID int64, nodeSet map[int64]bool) ([]int64, error) {
	node := g.Node(nodeUID)
	if node == nil {
		return nil, ErrNodeNotFound(nodeUID)
	}

	var visitedNodes []int64

	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			// true: follow this link
			// false: do not follow this link
			return nodeSet[e.To().ID()]
		},
		Visit: func(node graph.Node) {
			visitedNodes = append(visitedNodes, node.ID())
		},
	}

	g.SetDirected(false)
	w.Walk(g, node, nil)
	g.SetDirected(true)

	return visitedNodes, nil
}
