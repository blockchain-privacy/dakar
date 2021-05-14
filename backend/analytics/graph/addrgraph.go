package graph

import (
	"backend/cmd/cliutil"
	"backend/db/analytics"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"log"
	"runtime"
	"runtime/debug"
)

// addAddressEdges adds the edges defined in nodes to g.
func addAddressEdges(g *UndirectedGraph, nodes []analytics.ConnectedNode) error {
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

// loadAddresses loads origin addresses from the database into the graph.
// max is the number of addresses which get maximally loaded. If max is zero all possible addresses are loaded.
func loadAddresses(c *dgo.Dgraph, g *UndirectedGraph, transactionGraph *ReversibleGraph) error {
	txNodes := transactionGraph.Nodes()

	if txNodes.Len() == 0 {
		return errors.New("error nothing to load")
	}

	const step = 50000

	var uidsToLoad []string
	var i int64
	for txNodes.Next() {
		uidsToLoad = append(uidsToLoad, txNodes.Node().(addressGraphNode).String())
		i++
		if len(uidsToLoad) == step || txNodes.Len() == 0 {
			originNodes, err := analytics.GetInputAddresses(c, uidsToLoad)
			if err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}

			err = addAddressEdges(g, originNodes)
			if err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}

			uidsToLoad = []string{}
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
	log.Println("Loading origin nodes")
	if err := loadAddresses(c, g, transactionGraph); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	log.Println("address graph contains", g.Nodes().Len(), "nodes")
	// check
	log.Println("verifying address graph")
	if verificationErr := verifyAddressGraph(g); verificationErr != nil {
		return nil, verificationErr
	}
	debug.SetGCPercent(10)
	runtime.GC()
	log.Println("address graph loaded graph")
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
