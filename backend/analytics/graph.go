package analytics

import (
	"backend/cmd/cliutil"
	"backend/db/analytics"

	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/traverse"
)

// LoadGraph loads all relevant privacy transactions from the database and builds the returned graph
func LoadGraph(c *dgo.Dgraph) (*ReversibleGraph, error) {
	// load all mixing transaction from the database
	nodes, err := analytics.GetMixingTransactions(c)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// crates the graph
	f, err := createGraph(nodes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	//nodeSet := f.Nodes()
	//
	//const rootNodeId = int64(99999999999)
	//
	//var tips []int64
	//for nodeSet.Next() {
	//	nid := nodeSet.Node().ID()
	//	if f.To(nid).Len() == 0 {
	//		tips = append(tips, nid)
	//	}
	//}
	//
	//for _, t := range tips {
	//	f.SetEdge(simple.Edge{F: simple.Node(rootNodeId), T: simple.Node(t)})
	//}
	//
	//info("found", len(tips), "tips")
	//
	//dom := flow.Dominators(f.Node(rootNodeId), f)
	//_ = dom
	//
	//now := time.Now()
	//allOrigins, err := ReverseLookupById(f, rootNodeId)
	//if err != nil {
	//	return nil, err
	//}
	//info("time for complete reverse lookup", time.Since(now))
	//
	//info("allOrigins", len(allOrigins))

	// Render Graph
	//err = RenderGraph(f)
	//if err != nil {
	//	return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	//}

	// Export to Gephi
	//if exportErr := ExportToGephi("/home/dark/Downloads/graph.csv", f); exportErr != nil {
	//	return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), exportErr)
	//}

	return f, nil
}

// toHex returns a hexadecimal string representation of the given integer with the '0x' prefix
func toHex(i int64) string {
	return "0x" + strconv.FormatInt(i, 16)
}

func ReverseLookupById(g *ReversibleGraph, nodeId int64, maxGap time.Duration,
	maxLookBackTime time.Duration) (map[string]bool, error) {
	node := g.Node(nodeId)
	if node == nil {
		return nil, errors.New("error node not found")
	}

	foundEndpoints := make(map[string]bool)

	nodeTs := node.(transactionNode).ts

	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			// ALL nodes are of type transactionNode, therefore unsafe cast is okay
			if maxGap > 0 {
				toNode := e.To().(transactionNode)
				if toNode.ts.IsZero() {
					return true
				}
				fromNode := e.From().(transactionNode)

				if fromNode.ts.Sub(toNode.ts) > maxGap {
					return false
				}
			}
			// todo check if direction impacts time comparisons
			if maxLookBackTime > 0 {
				toNode := e.To().(transactionNode)
				if toNode.ts.IsZero() {
					return true
				}

				if nodeTs.Sub(toNode.ts) > maxLookBackTime {
					return false
				}
			}

			return true
		},
	}

	w.Walk(g, node, func(n graph.Node, d int) bool {
		from := g.From(n.ID())
		if from.Len() == 0 {
			foundEndpoints[toHex(n.ID())] = true
		}
		return false
	})

	return foundEndpoints, nil
}

// ReverseLookup returns all leaf nodes of the tree which has uid as its root
func ReverseLookup(g *ReversibleGraph, uid string, maxGap time.Duration,
	maxLookBackTime time.Duration) (map[string]bool, error) {
	nodeUid, err := strconv.ParseInt(uid[2:], 16, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	g.SetReverse(false)
	return ReverseLookupById(g, nodeUid, maxGap, maxLookBackTime)
}

// ForwardLookup returns all leaf nodes of the tree which has uid as its root
func ForwardLookup(g *ReversibleGraph, uid string, maxGap time.Duration,
	maxLookBackTime time.Duration) (map[string]bool, error) {
	nodeUid, err := strconv.ParseInt(uid[2:], 16, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	g.SetReverse(true)
	origins, err := ReverseLookupById(g, nodeUid, maxGap, maxLookBackTime)
	if err != nil {
		return nil, err
	}
	g.SetReverse(false)
	return origins, err
}

type transactionNode struct {
	ts time.Time
	id int64
}

func (n transactionNode) ID() int64      { return n.id }
func (n transactionNode) String() string { return toHex(n.id) }

// createGraph builds a directed reversible graph based on the provided nodes
func createGraph(nodes []analytics.MixingNode) (*ReversibleGraph, error) {
	g := NewReversibleGraph()

	// creating the nodes and edges in two steps so already existing nodes are not overwritten in step 2

	// 1. create all nodes
	for _, node := range nodes {
		nodeUid, err := strconv.ParseInt(node.Uid[2:], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		g.AddNode(transactionNode{ts: node.Block[0].Ts, id: nodeUid})
	}

	// 2. set edges between nodes
	for _, node := range nodes {
		// todo get uid from map
		nodeUid, err := strconv.ParseInt(node.Uid[2:], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		newNode := g.Node(nodeUid)
		for _, input := range node.Inputs {
			inputUid, parseErr := strconv.ParseInt(input.Uid[2:], 16, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), parseErr)
			}

			toNode := g.Node(inputUid)
			if toNode == nil {
				toNode = transactionNode{ts: time.Time{}, id: inputUid}
			}

			g.SetEdge(simple.Edge{F: newNode, T: toNode})
		}
	}

	return g, nil
}
