package graph

import (
	"backend/cmd/cliutil"

	"errors"
	"fmt"
	"time"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
)

func ReverseLookupById(g *ReversibleGraph, nodeId int64,
	maxLookBackTime time.Duration) (map[string]bool, map[string]bool, map[string]bool, error) {
	node := g.Node(nodeId)
	if node == nil {
		return nil, nil, nil, errors.New("error node not found")
	}

	foundOrigins := make(map[string]bool)
	foundCC := make(map[string]bool)
	foundOther := make(map[string]bool)

	nodeTs := node.(transactionNode).ts

	isReversed := g.IsReversed()

	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			// get node to which the edge leads
			toNode := g.Node(e.To().ID()).(transactionNode)

			// if a maximum look back time is set check the timestamp
			if maxLookBackTime > 0 {
				if isReversed {
					if toNode.ts.Sub(nodeTs) > maxLookBackTime {
						return false
					}
				}

				if nodeTs.Sub(toNode.ts) > maxLookBackTime {
					return false
				}
			}

			// if it is not a mixing transaction save it and stop following that edge
			if !toNode.privacyType.IsMixing() {
				if toNode.privacyType.IsOrigin() {
					foundOrigins[toNode.String()] = true
				} else if toNode.privacyType.IsCC() {
					foundCC[toNode.String()] = true
				} else {
					foundOther[toNode.String()] = true
				}
				return false
			}

			return true
		},
	}

	w.Walk(g, node, func(n graph.Node, d int) bool {
		from := g.From(n.ID())
		if from.Len() == 0 {
			thisNode := n.(transactionNode)
			if thisNode.privacyType.IsOrigin() {
				foundOrigins[thisNode.String()] = true
			} else if thisNode.privacyType.IsCC() {
				foundCC[thisNode.String()] = true
			} else {
				foundOther[thisNode.String()] = true
			}
		}
		return false
	})

	return foundOrigins, foundCC, foundOther, nil
}

// ReverseLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph backward
func ReverseLookup(g *ReversibleGraph, uid string,
	maxLookBackTime time.Duration) (map[string]bool, map[string]bool, map[string]bool, error) {
	nodeUid, err := toInteger(uid)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	g.SetReverse(false)
	return ReverseLookupById(g, nodeUid, maxLookBackTime)
}

// ForwardLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph forward
func ForwardLookup(g *ReversibleGraph, uid string, targetUid string) (map[string]bool,
	map[string]bool, map[string]bool, error) {
	nodeUid, err := toInteger(uid)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	targetNodeUid, err := toInteger(targetUid)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	node := g.Node(nodeUid).(transactionNode)
	targetNode := g.Node(targetNodeUid).(transactionNode)

	g.SetReverse(true)
	origins, cc, destination, err := ReverseLookupById(g, nodeUid, targetNode.ts.Sub(node.ts))
	if err != nil {
		return nil, nil, nil, err
	}
	g.SetReverse(false)
	return origins, cc, destination, err
}

// GetInputTransactions returns all transactions which
func GetInputTransactions() {

}
