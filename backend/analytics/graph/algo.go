package graph

import (
	"backend/cmd/cliutil"
	"strconv"

	"errors"
	"fmt"
	"time"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
)

var (
	ErrNodeNotFound = errors.New("error node does not exist in graph")
)

// toHex returns a hexadecimal string representation of the given integer with the '0x' prefix
func toHex(i int64) string {
	return "0x" + strconv.FormatInt(i, 16)
}

// toInteger a hex string in the form of "0x123" to an integer
func toInteger(hexString string) (int64, error) {
	return strconv.ParseInt(hexString[2:], 16, 64)
}

func ReverseLookupById(g *ReversibleGraph, nodeId int64, maxLookBackTime time.Duration) (map[string]bool, error) {
	node := g.Node(nodeId)
	if node == nil {
		return nil, ErrNodeNotFound
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
				foundEndpoints[toNode.String()] = true
				return false
			}

			return true
		},
	}

	w.Walk(g, node, func(n graph.Node, d int) bool {
		from := g.From(n.ID())
		if from.Len() == 0 {
			thisNode := n.(transactionNode)
			foundEndpoints[thisNode.String()] = true
		}
		return false
	})

	return foundEndpoints, nil
}

// ReverseLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph backward
func ReverseLookup(g *ReversibleGraph, uid string,
	maxLookBackTime time.Duration) (map[string]bool, error) {
	nodeUid, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	g.SetReverse(false)
	return ReverseLookupById(g, nodeUid, maxLookBackTime)
}

// ForwardLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph forward.
// It does not traverse paths which have a timestamp younger than the node specified by targetUid.
func ForwardLookup(g *ReversibleGraph, uid string, targetUid string) (map[string]bool, error) {
	nodeUid, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	targetNodeUid, err := toInteger(targetUid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	node := g.Node(nodeUid).(transactionNode)
	targetNode := g.Node(targetNodeUid).(transactionNode)

	g.SetReverse(true)
	origins, err := ReverseLookupById(g, nodeUid, targetNode.ts.Sub(node.ts))
	if err != nil {
		return nil, err
	}
	g.SetReverse(false)
	return origins, err
}

// ForwardLookupByTime returns all leaf nodes of the tree which has uid as its root while traversing the graph forward
// It does not traverse paths which are outside maxLookForwardTime.
func ForwardLookupByTime(g *ReversibleGraph, uid string, maxLookForwardTime time.Duration) (map[string]bool, error) {
	nodeUid, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	g.SetReverse(true)
	origins, err := ReverseLookupById(g, nodeUid, maxLookForwardTime)
	if err != nil {
		return nil, err
	}
	g.SetReverse(false)
	return origins, err
}

// GetInputTransactions returns the uids of all directly connected input transactions of the tx specified by uid
func GetInputTransactions(g *ReversibleGraph, uid string) ([]string, error) {
	// convert hex string to integer
	nodeUid, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// check if node exists
	if g.Node(nodeUid) == nil {
		return nil, fmt.Errorf("error node %s (base 10: %d) not found", uid, nodeUid)
	}

	var uids []string
	fromNodes := g.From(nodeUid)
	for fromNodes.Next() {
		n := fromNodes.Node()
		uids = append(uids, toHex(n.ID()))
	}

	return uids, nil
}

type ClusterId uint

// GetClusters returns a mapping between address uids and ClusterId's
func GetClusters(g *UndirectedGraph, addressUids map[string]bool) (map[string]ClusterId, error) {
	var w traverse.BreadthFirst
	clusterMap := make(map[string]ClusterId)

	i := ClusterId(0)

	for uid := range addressUids {
		if _, ok := clusterMap[uid]; ok {
			// uid already processed
			continue
		}

		nodeUid, err := toInteger(uid)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		node := g.Node(nodeUid)
		if node == nil {
			return nil, ErrNodeNotFound
		}

		var addressesInCluster []string

		w.Walk(g, node, func(n graph.Node, d int) bool {
			addrNode := g.Node(n.ID()).(addressGraphNode)
			// todo optimize map lookup from string to int64
			if addrNode.isAddress && addressUids[toHex(n.ID())] {
				addressesInCluster = append(addressesInCluster, addrNode.String())
			}

			return false
		})

		// set all address to same cluster, the initial address is included in the cluster
		for _, a := range addressesInCluster {
			clusterMap[a] = i
		}

		i++
	}

	return clusterMap, nil
}
