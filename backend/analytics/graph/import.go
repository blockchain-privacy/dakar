// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"backend/constants"
	"backend/db/analytics"
	"backend/external"
	"context"
	"errors"
	"fmt"
	"runtime"

	"gitlab.com/blockchain-privacy/gomisc/serror"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

var ErrDBContainsNoClassifiedTransactions = errors.New("db contains no classified transactions")

type Config struct {
	loadFunc func(ctx context.Context, c external.Database, numTxToLoad int) (*ReversibleGraph, error)
}

func NewDashConfig() Config {
	return Config{loadFunc: LoadDashTransactionGraph}
}

func NewBTCConfig() Config {
	return Config{loadFunc: LoadBTCTransactionGraph}
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

// loadDashOriginTransactions loads transactions filtered by type from the database into the graph.
// maxTransactions: the number of transactions which get maximally loaded. If max is zero all possible transaction are loaded.
// step: how many transactions are loaded in a single call to the database
// loadSingleNodes: whether single nodes or edges are supposed to be loaded
func loadTransactions(ctx context.Context, c external.Database, g *ReversibleGraph,
	step int, maxTransactions int, transactionType string, loadSingleNodes bool) error {
	lastNodeUID := "0x0"
	for i := 0; ; i += step {
		var numNodesLoaded int
		if loadSingleNodes {
			singleNodes, err := analytics.GetPrivacyTransactions(ctx, c, step, lastNodeUID, transactionType)
			if err != nil {
				return err
			}

			if len(singleNodes) > 0 {
				err = addSingleNodes(g, singleNodes)
				if err != nil {
					return err
				}
				lastNodeUID = singleNodes[len(singleNodes)-1].UID
			}

			numNodesLoaded = len(singleNodes)
		} else {
			// loading edges allocates a lot of memory, so run gc
			if i/step > 10 && (i/step)%3 == 0 {
				runtime.GC()
			}
			edges, err := analytics.GetConnectedPrivacyTransactions(ctx, c, step, lastNodeUID, transactionType)
			if err != nil {
				return err
			}

			if len(edges) > 0 {
				err = addEdges(g, edges)
				if err != nil {
					return err
				}
				lastNodeUID = edges[len(edges)-1].UID
			}

			numNodesLoaded = len(edges)
		}

		if numNodesLoaded < step || (maxTransactions > 0 && i+step >= maxTransactions) {
			break
		}
	}

	return nil
}

func LoadTransactionGraph(ctx context.Context, config Config, c external.Database, numTxToLoad int) (*ReversibleGraph, error) {
	if config.loadFunc == nil {
		return nil, serror.FromStr("nil loading function")
	}

	g, err := config.loadFunc(ctx, c, numTxToLoad)
	if err != nil {
		return nil, err
	}

	// only need to prune if a subset of transaction is loaded
	if err = pruneNodes(g); err != nil {
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

// LoadDashTransactionGraph loads and constructs the dash CoinJoin transaction graph from the database.
// numTxToLoad == 0: load all transactions
// numTxToLoad > 0: load numTxToLoad transactions of each transaction type
func LoadDashTransactionGraph(ctx context.Context, c external.Database, numTxToLoad int) (*ReversibleGraph, error) {
	mixingCount, originCount, ccCount, cpCount, destinationCount, getErr :=
		analytics.GetDashTransactionTypeCount(ctx, c)
	if getErr != nil {
		return nil, getErr
	}

	// nothing to do
	if mixingCount == 0 {
		return nil, ErrDBContainsNoClassifiedTransactions
	}

	info("db stats", "mixing_count", mixingCount, "origin_count", originCount,
		"destination_count", destinationCount, "cc_count", ccCount, "cp_count", cpCount)

	g := NewReversibleGraph(mixingCount + originCount + destinationCount)

	// load all origin transactions from the database
	info("Loading origin nodes")
	if err := loadTransactions(ctx, c, g, 50000, numTxToLoad,
		constants.TypeDashOrigin, true); err != nil {
		return nil, err
	}

	// load all cc transactions from the database
	info("Loading cc nodes")
	if err := loadTransactions(ctx, c, g, 50000, numTxToLoad,
		constants.TypeDashCC, true); err != nil {
		return nil, err
	}

	// load all mixing transactions from the database
	info("Loading mixing nodes")
	if err := loadTransactions(ctx, c, g, 50000, numTxToLoad,
		constants.TypeDashMixing, false); err != nil {
		return nil, err
	}
	// load all destination transactions from the database
	info("Loading destination nodes")
	if err := loadTransactions(ctx, c, g, 10000, numTxToLoad,
		constants.TypeDashDestination, false); err != nil {
		return nil, err
	}

	return g, nil
}

// LoadBTCTransactionGraph loads and constructs the Bitcoin CoinJoin transaction graph from the database.
// numTxToLoad == 0: load all transactions
// numTxToLoad > 0: load numTxToLoad transactions of each transaction type
func LoadBTCTransactionGraph(ctx context.Context, c external.Database, numTxToLoad int) (*ReversibleGraph, error) {
	wasabi2MixingCount, wasabi2OriginCount, wasabi2Destinationcount, whirlpoolMixingCount,
		whirlpoolOriginCount, whirlpoolDestinationcount, getErr := analytics.GetBTCTransactionTypeCount(ctx, c)
	if getErr != nil {
		return nil, getErr
	}

	// nothing to do
	if wasabi2MixingCount+whirlpoolMixingCount == 0 {
		return nil, ErrDBContainsNoClassifiedTransactions
	}

	info("db stats", "wasabi 2.0 mixing count", wasabi2MixingCount, "wasabi 2.0 origin count", wasabi2OriginCount,
		"wasabi 2.0 destination count", wasabi2Destinationcount, "whirlpool mixing count", whirlpoolMixingCount,
		"whirlpool origin count", whirlpoolOriginCount, "whirlpool destination count", whirlpoolDestinationcount)

	g := NewReversibleGraph(wasabi2MixingCount + wasabi2OriginCount + wasabi2Destinationcount)

	// load all origin transactions from the database
	info("Loading wasabi 2.0 origin nodes")
	if err := loadTransactions(ctx, c, g, 50000, numTxToLoad,
		constants.TypeWasabi2Origin, true); err != nil {
		return nil, err
	}

	info("Loading whirlpool origin nodes")
	// need to get whirlpool origins with connections for peeling chain clustering
	if err := loadTransactions(ctx, c, g, 50000, numTxToLoad,
		constants.TypeWhirlpoolOrigin, false); err != nil {
		return nil, err
	}

	// load all mixing transactions from the database
	info("Loading wasabi 2.0 mixing nodes")
	if err := loadTransactions(ctx, c, g, 50000, numTxToLoad,
		constants.TypeWasabi2Mixing, false); err != nil {
		return nil, err
	}

	// load all mixing transactions from the database
	info("Loading whirlpool mixing nodes")
	if err := loadTransactions(ctx, c, g, 50000, numTxToLoad,
		constants.TypeWhirlpoolMixing, false); err != nil {
		return nil, err
	}
	// load all destination transactions from the database
	info("Loading wasabi 2.0 destination nodes")
	if err := loadTransactions(ctx, c, g, 10000, numTxToLoad,
		constants.TypeWasabi2Destination, false); err != nil {
		return nil, err
	}

	info("Loading whirlpool destination nodes")
	if err := loadTransactions(ctx, c, g, 10000, numTxToLoad,
		constants.TypeWhirlpoolDestination, false); err != nil {
		return nil, err
	}

	return g, nil
}
