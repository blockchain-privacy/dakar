package graph

import (
	"backend/cmd/cliutil"
	"backend/db/analytics"

	"errors"
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/dgraph-io/dgo/v210"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// addAddressEdges adds the edges defined in nodes to g.
func addAddressEdges(g *UndirectedGraph, nodes []analytics.AddressNode) error {
	for _, node := range nodes {
		nodeUid, err := toInteger(node.Uid)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		g.UpdateNode(addressGraphNode{id: nodeUid})

		// inputs are here addresses
		for _, input := range node.Inputs {
			inputUid, parseErr := toInteger(input.Uid)
			if parseErr != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), parseErr)
			}

			g.UpdateNode(addressGraphNode{id: inputUid, isAddress: true})

			g.SetEdgeWithoutOverwrite(simple.Edge{F: simple.Node(nodeUid), T: simple.Node(inputUid)})
		}
	}

	return nil
}

// loadAddresses loads origin addresses from the database into the graph
func loadAddresses(c *dgo.Dgraph, g *UndirectedGraph, transactionGraph *ReversibleGraph) error {
	transactionGraph.SetReverse(false)
	txNodes := transactionGraph.Nodes()

	if txNodes.Len() == 0 {
		return errors.New("error nothing to load")
	}

	const step = 20000

	var uidsToLoad []string
	for txNodes.Next() {
		// either step was reached or this was the last node to load
		if len(uidsToLoad) == step || txNodes.Len() == 0 {
			runtime.GC()
			originNodes, err := analytics.GetInputAddresses(c, uidsToLoad)
			if err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}

			if err = addAddressEdges(g, originNodes); err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}

			uidsToLoad = []string{}
		}

		n := transactionGraph.From(txNodes.Node().ID())
		if n.Len() == 0 {
			uidsToLoad = append(uidsToLoad, txNodes.Node().(transactionNode).String())
		}
	}

	return nil
}

// LoadAddressGraph returns a graph containing all origins of transactionGraph and their input addresses
func LoadAddressGraph(c *dgo.Dgraph, transactionGraph *ReversibleGraph) (*UndirectedGraph, error) {
	mixingCount, originCount, ccCount, _, getErr := analytics.GetPrivacyTransactionCount(c)
	if getErr != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), getErr)
	}

	// nothing to do
	if mixingCount == 0 {
		return nil, nil
	}

	g := NewUndirectedGraph(originCount + ccCount)

	// load all origin transactions from the database
	info("Loading origin nodes")
	if err := loadAddresses(c, g, transactionGraph); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	info("address graph contains", g.Nodes().Len(), "nodes")
	// check
	info("verifying address graph")
	if verificationErr := verifyAddressGraph(g); verificationErr != nil {
		return nil, verificationErr
	}
	debug.SetGCPercent(30)
	runtime.GC()
	info("address graph loaded")
	return g, nil
}

// verifyAddressGraph checks the integrity of the graph
func verifyAddressGraph(g *UndirectedGraph) error {
	var node graph.Node
	var nodeId int64
	var ok bool

	nodes := g.Nodes()
	for nodes.Next() {
		node = nodes.Node()
		nodeId = node.ID()

		if g.From(nodeId).Len() == 0 {
			return errors.New("error node exists with no edges")
		}

		_, ok = node.(addressGraphNode)
		if !ok {
			return fmt.Errorf("error node has wrong type: %T", node)
		}
	}

	return nil
}
