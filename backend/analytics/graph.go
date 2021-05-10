package analytics

import (
	"backend/cmd/cliutil"
	"backend/db/analytics"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/traverse"
)

// loadOriginTransactions loads origin transactions from the database into the graph
func loadOriginTransactions(c *dgo.Dgraph, g *ReversibleGraph) error {
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

		if len(originNodes) < step {
			break
		}
	}

	return nil
}

// loadCCTransactions loads cc transactions from the database into the graph
func loadCCTransactions(c *dgo.Dgraph, g *ReversibleGraph) error {
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

		if len(ccNodes) < step {
			break
		}
	}

	return nil
}

// loadMixingTransactions loads mixing transactions from the database into the graph
func loadMixingTransactions(c *dgo.Dgraph, g *ReversibleGraph) error {
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

		if len(mixingNodes) < step {
			break
		}
	}

	return nil
}

// loadDestinationTransactions loads destination transactions from the database into the graph
func loadDestinationTransactions(c *dgo.Dgraph, g *ReversibleGraph) error {
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

		if len(destinationNodes) < step {
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

	info("db stats: mixing count:", mixingCount, "origin count:", originCount,
		"destination count:", destinationCount, "cc count:", ccCount)

	g := NewReversibleGraph(mixingCount + originCount + destinationCount)

	// load all origin transactions from the database
	info("Loading origin nodes")
	if err := loadOriginTransactions(c, g); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// load all cc transactions from the database
	info("Loading cc nodes")
	if err := loadCCTransactions(c, g); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// load all mixing transactions from the database
	info("Loading mixing nodes")
	if err := loadMixingTransactions(c, g); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	// load all destination transactions from the database
	info("Loading destination nodes")
	if err := loadDestinationTransactions(c, g); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// only need to prune if a subset of transaction is loaded
	info("pruning nodes")
	if err := pruneNodes(g); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	info("graph contains", g.Nodes().Len(), "nodes")
	// check
	info("verifying graph")
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

func ReverseLookupById(g *ReversibleGraph, nodeId int64, maxGap time.Duration,
	maxLookBackTime time.Duration) (map[string]bool, error) {
	node := g.Node(nodeId)
	if node == nil {
		return nil, errors.New("error node not found")
	}

	foundEndpoints := make(map[string]bool)
	nodeTs := node.(transactionNode).ts
	isGraphReversed := g.IsReversed()

	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			// ALL nodes are of type transactionNode, therefore unsafe cast is okay
			if maxGap > 0 {

				toNode := g.Node(e.To().ID()).(transactionNode)
				fromNode := g.Node(e.From().ID()).(transactionNode)

				if fromNode.ts.Sub(toNode.ts) > maxGap {
					return false
				}
			}

			if maxLookBackTime > 0 {
				toNode := g.Node(e.To().ID()).(transactionNode)
				if isGraphReversed {
					return true
					//return !(toNode.ts.Sub(nodeTs) >= maxLookBackTime)
				}

				return !(nodeTs.Sub(toNode.ts) >= maxLookBackTime)
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
	nodeUid, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	g.SetReverse(false)
	return ReverseLookupById(g, nodeUid, maxGap, maxLookBackTime)
}

// toInteger a hex string in the form of "0x123" to an integer
func toInteger(hexString string) (int64, error) {
	return strconv.ParseInt(hexString[2:], 16, 64)
}

// ForwardLookup returns all leaf nodes of the tree which has uid as its root
func ForwardLookup(g *ReversibleGraph, uid string, maxGap time.Duration,
	maxLookBackTime time.Time) (map[string]bool, error) {
	nodeUid, err := toInteger(uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// get duration from time
	maxLookBackTime.Sub(g.Node(nodeUid).(transactionNode).ts)
	g.SetReverse(true)
	// todo somehow get correct time from original node for which the reverse lookup was done
	origins, err := ReverseLookupById(g, nodeUid, maxGap, maxLookBackTime.Sub(g.Node(nodeUid).(transactionNode).ts))
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

// addSingleNodes adds the given nodes to g. Edges will not be set.
func addSingleNodes(g *ReversibleGraph, nodes []analytics.Node) error {
	for _, node := range nodes {
		nodeUid, err := toInteger(node.Uid)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		g.AddNode(transactionNode{ts: node.Block[0].Ts, id: nodeUid})
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

		g.UpdateNode(transactionNode{ts: node.Block[0].Ts, id: nodeUid})

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
