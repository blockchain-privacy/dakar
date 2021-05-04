package analytics

import (
	"backend/cmd/cliutil"
	"backend/db/analytics"

	"errors"
	"fmt"
	"strconv"

	"github.com/dgraph-io/dgo/v2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/traverse"
)

// LoadGraph loads all relevant privacy transactions from the database and builds the returned graph
func LoadGraph(c *dgo.Dgraph) (*simple.DirectedGraph, error) {
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

	return f, nil
}

// toHex returns a hexadecimal string representation of the given integer with the '0x' prefix
func toHex(i int64) string {
	return "0x" + strconv.FormatInt(i, 16)
}

// ReverseLookup returns all leaf nodes of the tree which has uid as its root
func ReverseLookup(g *simple.DirectedGraph, uid string) (map[string]bool, error) {
	w := traverse.BreadthFirst{
		Traverse: func(graph.Edge) bool {
			return true
		},
	}

	nodeUid, err := strconv.ParseInt(uid[2:], 16, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	node := g.Node(nodeUid)
	if node == nil {
		return nil, errors.New("error node not found")
	}

	foundEndpoints := make(map[string]bool)

	w.Walk(g, node, func(n graph.Node, d int) bool {
		from := g.From(n.ID())
		if from.Len() == 0 {
			foundEndpoints[toHex(n.ID())] = true
		}
		return false
	})

	return foundEndpoints, nil
}

// createGraph builds a directed graph based on the provided nodes
func createGraph(nodes []analytics.MixingNode) (*simple.DirectedGraph, error) {
	g := simple.NewDirectedGraph()
	for _, node := range nodes {
		nodeUid, err := strconv.ParseInt(node.Uid[2:], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		for _, input := range node.Inputs {
			inputUid, parseErr := strconv.ParseInt(input.Uid[2:], 16, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), parseErr)
			}
			g.SetEdge(simple.Edge{F: simple.Node(nodeUid), T: simple.Node(inputUid)})
		}
	}

	return g, nil
}
