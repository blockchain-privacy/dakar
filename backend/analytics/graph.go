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
	// todo actually load all transactions
	// set to zero to load ALL nodes
	const numNodesToLoad = 100000

	info("Loading mixing nodes")
	// load all mixing transactions from the database
	mixingNodes, err := analytics.GetMixingTransactions(c, numNodesToLoad)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	info("Loading destination nodes")
	// load all destination transactions from the database
	destinationNodes, err := analytics.GetDestinationTransactions(c, numNodesToLoad)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	info("Loading origin nodes")
	// load all origin transactions from the database
	originNodes, err := analytics.GetOriginTransactions(c, numNodesToLoad)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	info("Loading collateral creation nodes")
	// load all cc transactions from the database
	ccNodes, err := analytics.GetCollateralCreationTransactions(c, numNodesToLoad)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// crates the graph
	f, err := createGraph(mixingNodes, destinationNodes, originNodes, ccNodes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if numNodesToLoad > 0 {
		// only need to prune if a subset of transaction is loaded
		info("pruning nodes")
		pruneNodes(f)
	}

	// check
	info("verifying graph")
	if verificationErr := verifyGraph(f); verificationErr != nil {
		return nil, verificationErr
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
	isGraphReversed := g.IsReversed()

	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			// ALL nodes are of type transactionNode, therefore unsafe cast is okay
			if maxGap > 0 {
				toNode := e.To().(transactionNode)
				fromNode := e.From().(transactionNode)

				if fromNode.ts.Sub(toNode.ts) > maxGap {
					return false
				}
			}

			if maxLookBackTime > 0 {
				toNode := e.To().(transactionNode)
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

// addConnectedNodes adds the given nodes to g. Edges will not be set.
func addConnectedNodes(g *ReversibleGraph, nodes []analytics.ConnectedNode) error {
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

		newNode := g.Node(nodeUid)
		for _, input := range node.Inputs {
			inputUid, parseErr := toInteger(input.Uid)
			if parseErr != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), parseErr)
			}

			toNode := g.Node(inputUid)
			if toNode == nil {
				toNode = transactionNode{ts: time.Time{}, id: inputUid}
			}

			g.SetEdge(simple.Edge{F: newNode, T: toNode})
		}
	}

	return nil
}

// createGraph builds a directed reversible graph based on the provided nodes
func createGraph(mixingNodes []analytics.ConnectedNode, destinationNodes []analytics.ConnectedNode,
	originNodes []analytics.Node, ccNodes []analytics.Node) (*ReversibleGraph, error) {
	g := NewReversibleGraph()

	// Nodes and edges are created in separate steps so already existing nodes are not overwritten in step 2

	// 1. add nodes to graph
	// 1.1. create origin nodes
	err := addSingleNodes(g, originNodes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// 1.2. create cc nodes
	err = addSingleNodes(g, ccNodes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// 1.3. create mixing nodes
	err = addConnectedNodes(g, mixingNodes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// 1.4. create destination nodes
	err = addConnectedNodes(g, destinationNodes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// 2. set edges between nodes
	// 2.1 set edges defined by mixing nodes
	err = addEdges(g, mixingNodes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// 2.2 set edges defined by destination nodes
	err = addEdges(g, destinationNodes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return g, nil
}

func pruneNodes(g *ReversibleGraph) {
	var toRemove []int64

	nodes := g.Nodes()
	for nodes.Next() {
		node := nodes.Node()
		nodeId := node.ID()

		txNode := node.(transactionNode)

		if txNode.ts.IsZero() {
			toRemove = append(toRemove, nodeId)
		}
	}
	// remove nodes with zero ts
	for _, n := range toRemove {
		g.RemoveNode(n)
	}

	toRemove = []int64{}

	nodes = g.Nodes()
	for nodes.Next() {
		nodeId := nodes.Node().ID()

		if g.To(nodeId).Len() == 0 && g.From(nodeId).Len() == 0 {
			toRemove = append(toRemove, nodeId)
			continue
		}
	}
	// remove nodes with no connections
	for _, n := range toRemove {
		g.RemoveNode(n)
	}
}

// verifyGraph checks the integrity of the graph
func verifyGraph(g *ReversibleGraph) error {
	nodes := g.Nodes()

	for nodes.Next() {
		node := nodes.Node()
		nodeId := node.ID()

		if g.To(nodeId).Len() == 0 && g.From(nodeId).Len() == 0 {
			return errors.New("error node exists with no edges")
		}

		txNode := node.(transactionNode)

		if txNode.ts.IsZero() {
			return errors.New("error node has invalid timestamp")
		}
	}

	return nil
}
