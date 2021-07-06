package graph

import (
	"backend/cmd/cliutil"
	"backend/db/analytics"
	"backend/external"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// addAddressEdges adds the edges defined in nodes to g.
func addAddressEdges(g *UndirectedGraph, nodes []analytics.AddressNode) error {
	for _, node := range nodes {
		nodeUID, err := toInteger(node.UID)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		g.UpdateNode(addressGraphNode{id: nodeUID})

		// inputs are here addresses
		for _, input := range node.Inputs {
			inputUID, parseErr := toInteger(input.UID)
			if parseErr != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), parseErr)
			}

			g.UpdateNode(addressGraphNode{id: inputUID, isAddress: true})

			g.SetEdgeWithoutOverwrite(simple.Edge{F: simple.Node(nodeUID), T: simple.Node(inputUID)})
		}
	}

	return nil
}

// loadAddresses loads origin addresses from the database into the graph
func loadAddresses(c external.Database, g *UndirectedGraph, transactionGraph *ReversibleGraph) error {
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

		txNode := txNodes.Node().(transactionNode)
		if isEndpoint(transactionGraph, txNodes.Node().ID()) {
			uidsToLoad = append(uidsToLoad, txNode.String())
		}
	}

	return nil
}

func isEndpoint(transactionGraph *ReversibleGraph, nodeID int64) bool {
	if transactionGraph.From(nodeID).Len() == 0 {
		// no connections -> must be an endpoint
		return true
	} else {
		txNode := transactionGraph.Node(nodeID).(transactionNode)
		if !txNode.privacyType.IsMixing() {
			// check if the non mixing node has a mixing node as a parent
			parentNodes := transactionGraph.To(txNode.ID())
			for parentNodes.Next() {
				if parentNodes.Node().(transactionNode).privacyType.IsMixing() {
					return true
				}
			}
		}
	}

	return false
}

// LoadAddressGraph returns a graph containing all origins of transactionGraph and their input addresses
func LoadAddressGraph(c external.Database, transactionGraph *ReversibleGraph) (*UndirectedGraph, error) {
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
	var nodeID int64
	var ok bool

	nodes := g.Nodes()
	for nodes.Next() {
		node = nodes.Node()
		nodeID = node.ID()

		if g.From(nodeID).Len() == 0 {
			return fmt.Errorf("error node %s has no edges", toHex(nodeID))
		}

		_, ok = node.(addressGraphNode)
		if !ok {
			return fmt.Errorf("error node %s has wrong type: %T", toHex(node.ID()), node)
		}
	}

	return nil
}
