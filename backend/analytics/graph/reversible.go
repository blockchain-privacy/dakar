// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"fmt"
	"maps"
	"reflect"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/iterator"
)

var (
	dg *ReversibleGraph

	_ graph.Graph       = dg
	_ graph.Directed    = dg
	_ graph.NodeRemover = dg
	_ graph.EdgeRemover = dg
)

// ReversibleGraph implements a generalized directed graph which
// direction can be reversed for traversal.
type ReversibleGraph struct {
	reversed bool
	directed bool
	nodes    map[int64]graph.Node
	from     map[int64]map[int64]graph.Edge
	to       map[int64]map[int64]graph.Edge
}

// NewReversibleGraph returns a ReversibleGraph. Initializes internal data structures with the expected number of nodes.
func NewReversibleGraph(numNodesHint int) *ReversibleGraph {
	return &ReversibleGraph{
		reversed: false,
		directed: true,
		nodes:    make(map[int64]graph.Node, numNodesHint),
		from:     make(map[int64]map[int64]graph.Edge, numNodesHint),
		to:       make(map[int64]map[int64]graph.Edge, numNodesHint),
	}
}

// IsReversed returns true if the graph direction is reversed
func (g *ReversibleGraph) IsReversed() bool {
	return g.reversed
}

// SetReverse sets whether the direction of the graph is reversed
func (g *ReversibleGraph) SetReverse(reversed bool) {
	g.reversed = reversed
}

// IsDirected returns true if the graph is directed
func (g *ReversibleGraph) IsDirected() bool {
	return g.directed
}

// SetDirected sets whether the graph is directed
func (g *ReversibleGraph) SetDirected(directed bool) {
	g.directed = directed
}

// AddNode adds n to the graph. It panics if the added node ID matches an existing node ID.
func (g *ReversibleGraph) AddNode(n graph.Node) {
	if _, exists := g.nodes[n.ID()]; exists {
		panic(fmt.Sprintf("reversible: node ID collision: %d", n.ID()))
	}
	g.nodes[n.ID()] = n
}

// UpdateNode updates n. If the node does not already exist, it gets added to the graph.
func (g *ReversibleGraph) UpdateNode(n graph.Node) {
	g.nodes[n.ID()] = n
}

// Edge returns the edge from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
func (g *ReversibleGraph) Edge(uid, vid int64) graph.Edge {
	if !g.directed {
		edge, ok := g.from[uid][vid]
		if !ok {
			edge, ok := g.to[uid][vid]
			if !ok {
				return nil
			}
			return g.NewEdge(edge.To(), edge.From(), edge.(AddressEdge).AddressUIDs)
		}

		return edge
	}

	if g.reversed {
		edge, ok := g.to[uid][vid]
		if !ok {
			return nil
		}

		return g.NewEdge(edge.To(), edge.From(), edge.(AddressEdge).AddressUIDs)
	}

	edge, ok := g.from[uid][vid]
	if !ok {
		return nil
	}
	return edge
}

// Edges returns all the edges in the graph.
func (g *ReversibleGraph) Edges() graph.Edges {
	var edges []graph.Edge
	for _, u := range g.nodes {
		for _, e := range g.from[u.ID()] {
			edges = append(edges, e)
		}
	}
	if len(edges) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedEdges(edges)
}

// From returns all nodes in g that can be reached directly from n.
// The returned graph.Nodes are only valid until the next mutation of
// the receiver.
func (g *ReversibleGraph) From(id int64) graph.Nodes {
	if !g.directed {
		edges := map[int64]graph.Edge{}
		maps.Copy(edges, g.to[id])
		maps.Copy(edges, g.from[id])
		if len(edges) == 0 {
			return graph.Empty
		}

		return iterator.NewNodesByEdge(g.nodes, edges)
	}

	if g.reversed {
		if len(g.to[id]) == 0 {
			return graph.Empty
		}
		return iterator.NewNodesByEdge(g.nodes, g.to[id])
	}

	if len(g.from[id]) == 0 {
		return graph.Empty
	}
	return iterator.NewNodesByEdge(g.nodes, g.from[id])
}

// HasEdgeBetween returns whether an edge exists between nodes x and y without
// considering direction.
func (g *ReversibleGraph) HasEdgeBetween(xid, yid int64) bool {
	if _, ok := g.from[xid][yid]; ok {
		return true
	}
	_, ok := g.from[yid][xid]
	return ok
}

// HasEdgeFromTo returns whether an edge exists in the graph from u to v.
func (g *ReversibleGraph) HasEdgeFromTo(uid, vid int64) bool {
	if g.reversed {
		if _, ok := g.to[uid][vid]; !ok {
			return false
		}
		return true
	}

	if _, ok := g.from[uid][vid]; !ok {
		return false
	}
	return true
}

// NewEdge returns a new Edge from the source to the destination node.
func (g *ReversibleGraph) NewEdge(from, to graph.Node, addresses []int64) graph.Edge {
	return AddressEdge{F: from, T: to, AddressUIDs: addresses}
}

// Node returns the node with the given ID if it exists in the graph,
// and nil otherwise.
func (g *ReversibleGraph) Node(id int64) graph.Node {
	return g.nodes[id]
}

// Nodes returns all the nodes in the graph.
// The returned graph.Nodes are only valid until the next mutation of
// the receiver.
func (g *ReversibleGraph) Nodes() graph.Nodes {
	if len(g.nodes) == 0 {
		return graph.Empty
	}
	return iterator.NewNodes(g.nodes)
}

// RemoveEdge removes the edge with the given end point IDs from the graph, leaving the terminal
// nodes. If the edge does not exist it is a no-op.
func (g *ReversibleGraph) RemoveEdge(fid, tid int64) {
	if _, ok := g.nodes[fid]; !ok {
		return
	}
	if _, ok := g.nodes[tid]; !ok {
		return
	}

	delete(g.from[fid], tid)
	delete(g.to[tid], fid)
}

// RemoveNode removes the node with the given ID from the graph, as well as any edges attached
// to it. If the node is not in the graph it is a no-op.
func (g *ReversibleGraph) RemoveNode(id int64) {
	if _, ok := g.nodes[id]; !ok {
		return
	}
	delete(g.nodes, id)

	for from := range g.from[id] {
		delete(g.to[from], id)
	}
	delete(g.from, id)

	for to := range g.to[id] {
		delete(g.from[to], id)
	}
	delete(g.to, id)
}

// SetEdge adds e, an edge from one node to another. If the nodes do not exist, they are added
// and are set to the nodes of the edge otherwise.
// It will panic if the IDs of the e.From and e.To are equal.
func (g *ReversibleGraph) SetEdge(e graph.Edge) {
	var (
		from = e.From()
		fid  = from.ID()
		to   = e.To()
		tid  = to.ID()
	)

	if fid == tid {
		panic("reversible: adding self edge")
	}

	if _, ok := g.nodes[fid]; !ok {
		g.AddNode(from)
	} else {
		g.nodes[fid] = from
	}
	if _, ok := g.nodes[tid]; !ok {
		g.AddNode(to)
	} else {
		g.nodes[tid] = to
	}

	if fm, ok := g.from[fid]; ok {
		fm[tid] = e
	} else {
		g.from[fid] = map[int64]graph.Edge{tid: e}
	}
	if tm, ok := g.to[tid]; ok {
		tm[fid] = e
	} else {
		g.to[tid] = map[int64]graph.Edge{fid: e}
	}
}

// SetEdgeWithoutOverwrite adds e, an edge from one node to another. If the nodes do not exist, they are added
// and are set to the nodes of the edge otherwise.
// It will panic if the IDs of the e.From and e.To are equal.
func (g *ReversibleGraph) SetEdgeWithoutOverwrite(from graph.Node, to graph.Node, addressUID int64) {
	var (
		fid = from.ID()
		tid = to.ID()
	)

	if fid == tid {
		panic("reversible: adding self edge")
	}

	if _, ok := g.nodes[fid]; !ok {
		g.AddNode(from)
	}

	if _, ok := g.nodes[tid]; !ok {
		g.AddNode(to)
	}

	if fm, ok := g.from[fid]; ok {
		insertEdge(fm, tid, from, to, addressUID)
	} else {
		g.from[fid] = map[int64]graph.Edge{tid: AddressEdge{F: from, T: to, AddressUIDs: []int64{addressUID}}}
	}

	if tm, ok := g.to[tid]; ok {
		if fromEdge, ok := tm[fid]; ok {
			// append address UID
			addrEdge := fromEdge.(AddressEdge)
			addrEdge.AddressUIDs = append(addrEdge.AddressUIDs, addressUID)
			tm[fid] = addrEdge
		} else {
			tm[fid] = AddressEdge{F: from, T: to, AddressUIDs: []int64{addressUID}}
		}
	} else {
		g.to[tid] = map[int64]graph.Edge{fid: AddressEdge{F: from, T: to, AddressUIDs: []int64{addressUID}}}
	}
}

func insertEdge(fm map[int64]graph.Edge, tid int64, from graph.Node, to graph.Node, addressUID int64) {
	if toEdge, ok := fm[tid]; ok {
		// append address UID
		addrEdge := toEdge.(AddressEdge)
		addrEdge.AddressUIDs = append(addrEdge.AddressUIDs, addressUID)
		fm[tid] = addrEdge
	} else {
		fm[tid] = AddressEdge{F: from, T: to, AddressUIDs: []int64{addressUID}}
	}
}

// SetEdgesWithoutOverwrite adds an edge from 'from' node to all nodes contained in 'tos'.
// If the edges nodes do not already exist in the graph, they are created.
// Panics if the IDs of 'from' and one of 'tos' are equal.
func (g *ReversibleGraph) SetEdgesWithoutOverwrite(from graph.Node, tos []graph.Node, addressUIDs []int64) {
	fid := from.ID()

	if _, ok := g.nodes[fid]; !ok {
		g.AddNode(from)
	}

	var fm map[int64]graph.Edge
	fm, fok := g.from[fid]

	var newFromMap map[int64]graph.Edge

	for i, to := range tos {
		tid := to.ID()
		if fid == tid {
			panic("reversible: adding self edge")
		}

		if _, ok := g.nodes[tid]; !ok {
			g.AddNode(to)
		}

		if fok {
			insertEdge(fm, tid, from, to, addressUIDs[i])
		} else {
			if newFromMap == nil {
				newFromMap = make(map[int64]graph.Edge, len(tos))
				newFromMap[tid] = AddressEdge{F: from, T: to, AddressUIDs: []int64{addressUIDs[i]}}
			} else {
				insertEdge(newFromMap, tid, from, to, addressUIDs[i])
			}
		}

		if tm, ok := g.to[tid]; ok {
			if fromEdge, ok := tm[fid]; ok {
				// append address UID
				addrEdge := fromEdge.(AddressEdge)
				addrEdge.AddressUIDs = append(addrEdge.AddressUIDs, addressUIDs[i])
				tm[fid] = addrEdge
			} else {
				tm[fid] = AddressEdge{F: from, T: to, AddressUIDs: []int64{addressUIDs[i]}}
			}
		} else {
			g.to[tid] = map[int64]graph.Edge{fid: AddressEdge{F: from, T: to, AddressUIDs: []int64{addressUIDs[i]}}}
		}
	}

	if !fok {
		g.from[fid] = newFromMap
	}
}

// To returns all nodes in g that can reach directly to n.
// The returned graph.Nodes are only valid until the next mutation of
// the receiver.
func (g *ReversibleGraph) To(id int64) graph.Nodes {
	if g.reversed {
		if len(g.from[id]) == 0 {
			return graph.Empty
		}
		return iterator.NewNodesByEdge(g.nodes, g.from[id])
	}
	if len(g.to[id]) == 0 {
		return graph.Empty
	}
	return iterator.NewNodesByEdge(g.nodes, g.to[id])
}

// IsEqual returns true if g2 is equal to g
func (g *ReversibleGraph) IsEqual(g2 *ReversibleGraph) bool {
	nodesAreEqual := reflect.DeepEqual(g.nodes, g2.nodes)
	fromEqual := reflect.DeepEqual(g.from, g2.from)
	toEqual := reflect.DeepEqual(g.to, g2.to)

	return g.reversed == g2.reversed && nodesAreEqual && fromEqual && toEqual
}
