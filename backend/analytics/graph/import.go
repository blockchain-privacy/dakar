package graph

import (
	"backend/db/analytics"
	"backend/external"
	"errors"
	"fmt"
	"github.com/qrest/gomisc/serror"

	"runtime"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

var ErrDBContainsNoPrivacyTransactions = errors.New("db contains no privacy transactions")

// loadOriginTransactions loads origin transactions from the database into the graph.
// max is the number of transactions which get maximally loaded. If max is zero all possible transaction are loaded.
func loadOriginTransactions(c external.Database, g *ReversibleGraph, max int) error {
	const step = 50000
	for i := 0; ; i += step {
		originNodes, err := analytics.GetOriginTransactions(c, step, i)
		if err != nil {
			return err
		}

		err = addSingleNodes(g, originNodes)
		if err != nil {
			return err
		}

		if len(originNodes) < step || (max > 0 && i+step >= max) {
			break
		}
	}

	return nil
}

// loadCCTransactions loads cc transactions from the database into the graph.
// max is the number of transactions which get maximally loaded. If max is zero all possible transaction are loaded.
func loadCCTransactions(c external.Database, g *ReversibleGraph, max int) error {
	const step = 50000
	for i := 0; ; i += step {
		ccNodes, err := analytics.GetCollateralCreationTransactions(c, step, i)
		if err != nil {
			return err
		}

		err = addSingleNodes(g, ccNodes)
		if err != nil {
			return err
		}

		if len(ccNodes) < step || (max > 0 && i+step >= max) {
			break
		}
	}

	return nil
}

// loadMixingTransactions loads mixing transactions from the database into the graph.
// max is the number of transactions which get maximally loaded. If max is zero all possible transaction are loaded.
func loadMixingTransactions(c external.Database, g *ReversibleGraph, max int) error {
	const step = 50000
	for i := 0; ; i += step {
		if i/step > 10 && (i/step)%3 == 0 {
			runtime.GC()
		}
		mixingNodes, err := analytics.GetMixingTransactions(c, step, i)
		if err != nil {
			return err
		}

		err = addEdges(g, mixingNodes)
		if err != nil {
			return err
		}

		if len(mixingNodes) < step || (max > 0 && i+step >= max) {
			break
		}
	}

	return nil
}

// loadDestinationTransactions loads destination transactions from the database into the graph
// max is the number of transactions which get maximally loaded. If max is zero all possible transaction are loaded.
func loadDestinationTransactions(c external.Database, g *ReversibleGraph, max int) error {
	const step = 10000
	for i := 0; ; i += step {
		if i/step > 5 && (i/step)%2 == 0 {
			runtime.GC()
		}
		destinationNodes, err := analytics.GetDestinationTransactions(c, step, i)
		if err != nil {
			return err
		}

		err = addEdges(g, destinationNodes)
		if err != nil {
			return err
		}

		if len(destinationNodes) < step || (max > 0 && i+step >= max) {
			break
		}
	}

	return nil
}

// LoadTransactionGraph loads and constructs a transaction graph from the database.
// numTxToLoad == 0: load all transactions
// numTxToLoad > 0: load numTxToLoad transactions of each privacy type
func LoadTransactionGraph(c external.Database, numTxToLoad int) (*ReversibleGraph, error) {
	mixingCount, originCount, ccCount, cpCount, destinationCount, getErr :=
		analytics.GetTransactionTypeCount(c)
	if getErr != nil {
		return nil, getErr
	}

	// nothing to do
	if mixingCount == 0 {
		return nil, ErrDBContainsNoPrivacyTransactions
	}

	info("db stats", "mixing_count", mixingCount, "origin_count", originCount,
		"destination_count", destinationCount, "cc_count", ccCount, "cp_count", cpCount)

	g := NewReversibleGraph(mixingCount + originCount + destinationCount)

	// load all origin transactions from the database
	info("Loading origin nodes")
	if err := loadOriginTransactions(c, g, numTxToLoad); err != nil {
		return nil, err
	}

	// load all cc transactions from the database
	info("Loading cc nodes")
	if err := loadCCTransactions(c, g, numTxToLoad); err != nil {
		return nil, err
	}

	// load all mixing transactions from the database
	info("Loading mixing nodes")
	if err := loadMixingTransactions(c, g, numTxToLoad); err != nil {
		return nil, err
	}
	// load all destination transactions from the database
	info("Loading destination nodes")
	if err := loadDestinationTransactions(c, g, numTxToLoad/10); err != nil {
		return nil, err
	}

	// only need to prune if a subset of transaction is loaded
	if err := pruneNodes(g); err != nil {
		return nil, err
	}
	info(fmt.Sprintf("transaction graph contains %d nodes", g.Nodes().Len()))
	// check
	info("verifying transaction graph")
	if verificationErr := verifyTransactionGraph(g); verificationErr != nil {
		return nil, verificationErr
	}
	runtime.GC()
	info("transaction graph loaded")

	return g, nil
}

// addSingleNodes adds the given nodes to g. Edges will not be set.
func addSingleNodes(g *ReversibleGraph, nodes []analytics.Node) error {
	for _, node := range nodes {
		nodeUID, err := ToInteger(node.UID)
		if err != nil {
			return err
		}

		g.AddNode(TransactionNode{id: nodeUID, TS: node.Block[0].TS, Type: node.TransactionType})
	}

	return nil
}

// upsertSingleNodes adds the given nodes to g or updates existing ones. Edges will not be set.
func upsertSingleNodes(g *ReversibleGraph, nodes []analytics.Node) error {
	for _, node := range nodes {
		nodeUID, err := ToInteger(node.UID)
		if err != nil {
			return err
		}

		g.UpdateNode(TransactionNode{id: nodeUID, TS: node.Block[0].TS, Type: node.TransactionType})
	}

	return nil
}

// addEdges adds the edges defined in nodes to g.
func addEdges(g *ReversibleGraph, nodes []analytics.ConnectedNode) error {
	for _, node := range nodes {
		nodeUID, err := ToInteger(node.UID)
		if err != nil {
			return err
		}

		g.UpdateNode(TransactionNode{id: nodeUID, TS: node.TS, Type: node.Type})

		addressUIDs := make([]int64, len(node.Inputs))
		inputUIDs := make([]graph.Node, len(node.Inputs))

		for i, input := range node.Inputs {
			inputUID, err := ToInteger(input.InputTransaction)
			if err != nil {
				return err
			}
			inputUIDs[i] = simple.Node(inputUID)

			addressUIDs[i], err = ToInteger(input.Address)
			if err != nil {
				return err
			}
		}

		g.SetEdgesWithoutOverwrite(simple.Node(nodeUID), inputUIDs, addressUIDs)
	}

	return nil
}

// pruneNodes removes all nodes from the graph which are shallow or have no edges
func pruneNodes(g *ReversibleGraph) error {
	var node graph.Node
	var nodeID int64
	var txNode TransactionNode
	var ok bool

	nodes := g.Nodes()
	for nodes.Next() {
		node = nodes.Node()
		nodeID = node.ID()

		txNode, ok = node.(TransactionNode)

		if !ok {
			g.RemoveNode(nodeID)
		} else if txNode.TS.IsZero() {
			return serror.FromStr("node timestamp is zero")
		}
	}

	nodes = g.Nodes()
	for nodes.Next() {
		nodeID = nodes.Node().ID()

		if g.To(nodeID).Len() == 0 && g.From(nodeID).Len() == 0 {
			g.RemoveNode(nodeID)
		}
	}

	return nil
}

// verifyTransactionGraph checks the integrity of the graph
func verifyTransactionGraph(g *ReversibleGraph) error {
	var node graph.Node
	var nodeID int64
	var txNode TransactionNode
	var ok bool

	nodes := g.Nodes()
	for nodes.Next() {
		node = nodes.Node()
		nodeID = node.ID()

		if g.To(nodeID).Len() == 0 && g.From(nodeID).Len() == 0 {
			return serror.FromStr("node exists with no edges")
		}

		txNode, ok = node.(TransactionNode)
		if !ok {
			return serror.FromFormat("node has wrong type: %T", node)
		}

		if txNode.TS.IsZero() {
			return serror.FromStr("node has invalid timestamp")
		}
	}

	return nil
}
