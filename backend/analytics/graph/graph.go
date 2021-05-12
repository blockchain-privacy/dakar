package graph

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db/analytics"
	"log"

	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type transactionNode struct {
	ts          time.Time
	id          int64
	privacyType constants.PrivacyType
}

func (n transactionNode) ID() int64      { return n.id }
func (n transactionNode) String() string { return toHex(n.id) }

// loadOriginTransactions loads origin transactions from the database into the graph.
// max is the number of transactions which get maximally loaded. If max is zero all possible transaction are loaded.
func loadOriginTransactions(c *dgo.Dgraph, g *ReversibleGraph, max int) error {
	const step = 50000
	for i := 0; ; i += step {
		originNodes, err := analytics.GetOriginTransactions(c, step, i)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		err = addSingleNodes(g, originNodes)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(originNodes) < step || (max > 0 && i+step >= max) {
			break
		}
	}

	return nil
}

// loadCCTransactions loads cc transactions from the database into the graph.
//max is the number of transactions which get maximally loaded. If max is zero all possible transaction are loaded.
func loadCCTransactions(c *dgo.Dgraph, g *ReversibleGraph, max int) error {
	const step = 50000
	for i := 0; ; i += step {
		ccNodes, err := analytics.GetCollateralCreationTransactions(c, step, i)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		err = addSingleNodes(g, ccNodes)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(ccNodes) < step || (max > 0 && i+step >= max) {
			break
		}
	}

	return nil
}

// loadMixingTransactions loads mixing transactions from the database into the graph.
//max is the number of transactions which get maximally loaded. If max is zero all possible transaction are loaded.
func loadMixingTransactions(c *dgo.Dgraph, g *ReversibleGraph, max int) error {
	const step = 50000
	for i := 0; ; i += step {
		if i/step > 10 && (i/step)%3 == 0 {
			runtime.GC()
		}
		mixingNodes, err := analytics.GetMixingTransactions(c, step, i)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		err = addEdges(g, mixingNodes)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(mixingNodes) < step || (max > 0 && i+step >= max) {
			break
		}
	}

	return nil
}

// loadDestinationTransactions loads destination transactions from the database into the graph
//max is the number of transactions which get maximally loaded. If max is zero all possible transaction are loaded.
func loadDestinationTransactions(c *dgo.Dgraph, g *ReversibleGraph, max int) error {
	const step = 10000
	for i := 0; ; i += step {
		if i/step > 5 && (i/step)%2 == 0 {
			runtime.GC()
		}
		destinationNodes, err := analytics.GetDestinationTransactions(c, step, i)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		err = addEdges(g, destinationNodes)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if len(destinationNodes) < step || (max > 0 && i+step >= max) {
			break
		}
	}

	return nil
}

func LoadGraphInSteps(c *dgo.Dgraph) (*ReversibleGraph, error) {
	mixingCount, originCount, ccCount, destinationCount, getErr :=
		analytics.GetPrivacyTransactionCount(c)
	if getErr != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), getErr)
	}

	// nothing to do
	if mixingCount == 0 {
		return nil, nil
	}

	log.Println("db stats: mixing count:", mixingCount, "origin count:", originCount,
		"destination count:", destinationCount, "cc count:", ccCount)

	g := NewReversibleGraph(mixingCount + originCount + destinationCount)

	const numTxToLoad = 0

	// load all origin transactions from the database
	log.Println("Loading origin nodes")
	if err := loadOriginTransactions(c, g, numTxToLoad); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// load all cc transactions from the database
	log.Println("Loading cc nodes")
	if err := loadCCTransactions(c, g, numTxToLoad); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// load all mixing transactions from the database
	log.Println("Loading mixing nodes")
	if err := loadMixingTransactions(c, g, numTxToLoad); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	// load all destination transactions from the database
	log.Println("Loading destination nodes")
	if err := loadDestinationTransactions(c, g, numTxToLoad/10); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// only need to prune if a subset of transaction is loaded
	log.Println("pruning nodes")
	if err := pruneNodes(g); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	log.Println("graph contains", g.Nodes().Len(), "nodes")
	// check
	log.Println("verifying graph")
	if verificationErr := verifyGraph(g); verificationErr != nil {
		return nil, verificationErr
	}
	debug.SetGCPercent(10)
	runtime.GC()

	return g, nil
}

// toHex returns a hexadecimal string representation of the given integer with the '0x' prefix
func toHex(i int64) string {
	return "0x" + strconv.FormatInt(i, 16)
}

// toInteger a hex string in the form of "0x123" to an integer
func toInteger(hexString string) (int64, error) {
	return strconv.ParseInt(hexString[2:], 16, 64)
}

// addSingleNodes adds the given nodes to g. Edges will not be set.
func addSingleNodes(g *ReversibleGraph, nodes []analytics.Node) error {
	for _, node := range nodes {
		nodeUid, err := toInteger(node.Uid)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		g.AddNode(transactionNode{id: nodeUid, ts: node.Block[0].Ts, privacyType: node.PrivacyType})
	}

	return nil
}

// addEdges adds the edges defined in nodes to g. Before this function addConnectedNodes should be called with the same set of edges.
func addEdges(g *ReversibleGraph, nodes []analytics.ConnectedNode) error {
	for _, node := range nodes {
		nodeUid, err := toInteger(node.Uid)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		g.UpdateNode(transactionNode{id: nodeUid, ts: node.Block[0].Ts, privacyType: node.PrivacyType})

		for _, input := range node.Inputs {
			inputUid, parseErr := toInteger(input.Uid)
			if parseErr != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), parseErr)
			}

			g.SetEdgeWithoutOverwrite(simple.Edge{F: simple.Node(nodeUid), T: simple.Node(inputUid)})
		}
	}

	return nil
}

// pruneNodes removes all nodes from the graph which are shallow or have no edges
func pruneNodes(g *ReversibleGraph) error {
	var node graph.Node
	var nodeId int64
	var txNode transactionNode
	var ok bool

	nodes := g.Nodes()
	for nodes.Next() {
		node = nodes.Node()
		nodeId = node.ID()

		txNode, ok = node.(transactionNode)

		if !ok {
			g.RemoveNode(nodeId)
		} else if txNode.ts.IsZero() {
			return errors.New("error node timestamp is zero")
		}
	}

	nodes = g.Nodes()
	for nodes.Next() {
		nodeId = nodes.Node().ID()

		if g.To(nodeId).Len() == 0 && g.From(nodeId).Len() == 0 {
			g.RemoveNode(nodeId)
		}
	}

	return nil
}

// verifyGraph checks the integrity of the graph
func verifyGraph(g *ReversibleGraph) error {
	var node graph.Node
	var nodeId int64
	var txNode transactionNode
	var ok bool

	nodes := g.Nodes()
	for nodes.Next() {
		node = nodes.Node()
		nodeId = node.ID()

		if g.To(nodeId).Len() == 0 && g.From(nodeId).Len() == 0 {
			return errors.New("error node exists with no edges")
		}

		txNode, ok = node.(transactionNode)
		if !ok {
			return fmt.Errorf("error node has wrong type: %T", node)
		}

		if txNode.ts.IsZero() {
			return errors.New("error node has invalid timestamp")
		}
	}

	return nil
}
