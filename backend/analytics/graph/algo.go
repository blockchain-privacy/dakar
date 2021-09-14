package graph

import (
	"backend/cmd/cliutil"
	"strconv"

	"fmt"
	"time"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
)

func errNodeNotFound(nodeID int64) error {
	return fmt.Errorf("error node %s does not exist in graph", toHex(nodeID))
}

// toHex returns a hexadecimal string representation of the given integer with the '0x' prefix
func toHex(i int64) string {
	return "0x" + strconv.FormatInt(i, 16)
}

// toInteger a hex string in the form of "0x123" to an integer
func toInteger(hexString string) (int64, error) {
	return strconv.ParseInt(hexString[2:], 16, 64)
}

// ReverseLookupByID performs a reverse lookup starting at the given nodeId
func ReverseLookupByID(g *ReversibleGraph, nodeID int64, maxLookBackTime time.Duration) (map[string]bool, error) {
	node := g.Node(nodeID)
	if node == nil {
		return nil, errNodeNotFound(nodeID)
	}

	foundEndpoints := make(map[string]bool)

	nodeTs := node.(transactionNode).ts

	isReversed := g.IsReversed()

	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			// get node to which the edge leads
			toNode := g.Node(e.To().ID()).(transactionNode)

			// if a maximum look back time is set check the timestamp
			if maxLookBackTime > 0 {
				// isReversed is true if it is a forward lookup: default case is a reverse
				// lookup so if the graph is reversed a forward lookup is happening
				if isReversed {
					if toNode.ts.Sub(nodeTs) > maxLookBackTime {
						// The forward lookup starts at an origin transaction and looks forward
						// for a certain user-defined duration. Funds from mixing transaction are usually
						// not spent directly after the mixing process is finished. Because of this, the next
						// connected non-mixing transaction must be included to the returned set.
						if !toNode.privacyType.IsMixing() {
							foundEndpoints[toNode.String()] = true
						}

						return false
					}
				} else if nodeTs.Sub(toNode.ts) > maxLookBackTime {
					return false
				}
			}

			// if it is not a mixing transaction save it and stop following that edge
			if !toNode.privacyType.IsMixing() {
				foundEndpoints[toNode.String()] = true
				return false
			}

			// true: follow this link
			// false: do not follow this link
			return true
		},
	}

	w.Walk(g, node, func(n graph.Node, d int) bool {
		from := g.From(n.ID())
		if from.Len() == 0 {
			thisNode := n.(transactionNode)
			foundEndpoints[thisNode.String()] = true
		}

		// true: stop traversing nodes
		// false: do not stop traversing nodes
		return false
	})

	return foundEndpoints, nil
}

// ReverseLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph backward
func ReverseLookup(g *ReversibleGraph, uid string,
	maxLookBackTime time.Duration) (map[string]bool, error) {
	nodeUID, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	g.SetReverse(false)

	results, err := ReverseLookupByID(g, nodeUID, maxLookBackTime)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	return results, nil
}

// ForwardLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph forward.
// It does not traverse paths which have a timestamp younger than the node specified by targetUid.
func ForwardLookup(g *ReversibleGraph, uid string, targetUID string) (map[string]bool, error) {
	nodeUID, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	targetNodeUID, err := toInteger(targetUID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	node := g.Node(nodeUID).(transactionNode)
	targetNode := g.Node(targetNodeUID).(transactionNode)

	g.SetReverse(true)
	origins, err := ReverseLookupByID(g, nodeUID, targetNode.ts.Sub(node.ts))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	g.SetReverse(false)
	return origins, err
}

// ForwardLookupByTime returns all leaf nodes of the tree which has uid as its root while traversing the graph forward
// It does not traverse paths which are outside maxLookForwardTime.
func ForwardLookupByTime(g *ReversibleGraph, uid string, maxLookForwardTime time.Duration) (map[string]bool, error) {
	nodeUID, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	g.SetReverse(true)
	origins, err := ReverseLookupByID(g, nodeUID, maxLookForwardTime)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	g.SetReverse(false)
	return origins, err
}

// GetInputTransactions returns the uids of all directly connected input transactions of the tx specified by uid
func GetInputTransactions(g *ReversibleGraph, uid string) ([]string, error) {
	// convert hex string to integer
	nodeUID, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// check if node exists
	if g.Node(nodeUID) == nil {
		return nil, errNodeNotFound(nodeUID)
	}

	var uids []string
	fromNodes := g.From(nodeUID)
	for fromNodes.Next() {
		n := fromNodes.Node()
		uids = append(uids, toHex(n.ID()))
	}

	return uids, nil
}
