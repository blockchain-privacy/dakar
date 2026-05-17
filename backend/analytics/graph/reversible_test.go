// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/graph"
)

func TestNewReversibleGraph(t *testing.T) {
	require.NotNil(t, NewReversibleGraph(0))
	require.NotNil(t, NewReversibleGraph(5))
	require.NotNil(t, NewReversibleGraph(1000))
}

func TestNewReversibleGraph_Reversible(t *testing.T) {
	g := NewReversibleGraph(1)
	require.False(t, g.IsReversed())
	g.SetReverse(false)
	require.False(t, g.IsReversed())
	g.SetReverse(true)
	require.True(t, g.IsReversed())
}

func TestReversibleGraph_AddNode(t *testing.T) {
	g := NewReversibleGraph(5)
	require.NotPanics(t, func() {
		g.AddNode(TransactionNode{id: 1})
		g.AddNode(TransactionNode{id: 2})
		g.AddNode(TransactionNode{id: 3})
	})

	require.Panics(t, func() {
		g.AddNode(TransactionNode{id: 1})
	})
}

func TestReversibleGraph_UpdateNode(t *testing.T) {
	g := NewReversibleGraph(5)
	require.NotPanics(t, func() {
		g.UpdateNode(TransactionNode{id: 1})
		g.UpdateNode(TransactionNode{id: 2})
		g.UpdateNode(TransactionNode{id: 3})
		g.UpdateNode(TransactionNode{id: 1})
	})
}

func createSimpleGraph() *ReversibleGraph {
	g := NewReversibleGraph(5)
	g.SetEdge(AddressEdge{F: TransactionNode{id: 1}, T: TransactionNode{id: 2}})
	g.SetEdge(AddressEdge{F: TransactionNode{id: 1}, T: TransactionNode{id: 3}})
	g.SetEdge(AddressEdge{F: TransactionNode{id: 3}, T: TransactionNode{id: 4}})

	return g
}

func TestReversibleGraph_Edge(t *testing.T) {
	g := createSimpleGraph()
	type args struct {
		uid int64
		vid int64
	}
	tests := []struct {
		args    args
		wantNil bool
	}{
		{
			args:    args{},
			wantNil: true,
		},
		{
			args:    args{uid: 1, vid: 2},
			wantNil: false,
		},
		{
			args:    args{uid: 1, vid: 3},
			wantNil: false,
		},
		{
			args:    args{uid: 5, vid: 6},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		if tt.wantNil {
			require.Nil(t, g.Edge(tt.args.uid, tt.args.vid))
		} else {
			require.NotNil(t, g.Edge(tt.args.uid, tt.args.vid))
		}
	}
}

func TestReversibleGraph_Edges(t *testing.T) {
	g := NewReversibleGraph(5)
	require.Equal(t, graph.Empty, g.Edges())
	g = createSimpleGraph()
	edges := g.Edges()
	require.Equal(t, 3, edges.Len())
}

func TestReversibleGraph_From(t *testing.T) {
	g := NewReversibleGraph(5)
	require.Equal(t, graph.Empty, g.From(1))
	g.SetReverse(true)
	require.Equal(t, graph.Empty, g.From(1))

	g = createSimpleGraph()

	tests := []struct {
		id           int64
		reversed     bool
		wantNumNodes int
	}{
		{id: 1, reversed: false, wantNumNodes: 2},
		{id: 3, reversed: false, wantNumNodes: 1},
		{id: 4, reversed: false, wantNumNodes: 0},
		{id: 1, reversed: true, wantNumNodes: 0},
		{id: 2, reversed: true, wantNumNodes: 1},
		{id: 3, reversed: true, wantNumNodes: 1},
		{id: 4, reversed: true, wantNumNodes: 1},
	}
	for _, tt := range tests {
		g.SetReverse(tt.reversed)
		require.Equal(t, tt.wantNumNodes, g.From(tt.id).Len())
	}
}

func TestReversibleGraph_HasEdgeBetween(t *testing.T) {
	g := createSimpleGraph()

	type args struct {
		xid int64
		yid int64
	}
	tests := []struct {
		args args
		want bool
	}{
		{args: args{xid: 1, yid: 2}, want: true},
		{args: args{xid: 2, yid: 1}, want: true},
		{args: args{xid: 1, yid: 3}, want: true},
		{args: args{xid: 3, yid: 1}, want: true},
		{args: args{xid: 1, yid: 10}, want: false},
		{args: args{xid: 1, yid: 4}, want: false},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, g.HasEdgeBetween(tt.args.xid, tt.args.yid))
	}
}

func TestReversibleGraph_HasEdgeFromTo(t *testing.T) {
	g := createSimpleGraph()

	type args struct {
		uid      int64
		vid      int64
		reversed bool
	}
	tests := []struct {
		args args
		want bool
	}{
		{args: args{uid: 1, vid: 2, reversed: false}, want: true},
		{args: args{uid: 2, vid: 1, reversed: false}, want: false},
		{args: args{uid: 1, vid: 3, reversed: false}, want: true},
		{args: args{uid: 3, vid: 1, reversed: false}, want: false},
		{args: args{uid: 1, vid: 10, reversed: false}, want: false},
		{args: args{uid: 1, vid: 4, reversed: false}, want: false},

		{args: args{uid: 2, vid: 1, reversed: true}, want: true},
		{args: args{uid: 1, vid: 2, reversed: true}, want: false},
		{args: args{uid: 1, vid: 3, reversed: true}, want: false},
		{args: args{uid: 3, vid: 1, reversed: true}, want: true},
		{args: args{uid: 1, vid: 10, reversed: true}, want: false},
		{args: args{uid: 4, vid: 1, reversed: true}, want: false},
	}
	for _, tt := range tests {
		g.SetReverse(tt.args.reversed)
		require.Equal(t, tt.want, g.HasEdgeFromTo(tt.args.uid, tt.args.vid))
	}
}

func TestReversibleGraph_NewEdge(t *testing.T) {
	require.Equal(t, AddressEdge{F: TransactionNode{id: 1}, T: TransactionNode{id: 2}, AddressUIDs: nil},
		new(ReversibleGraph).NewEdge(TransactionNode{id: 1}, TransactionNode{id: 2}, nil))
}

func TestReversibleGraph_Node(t *testing.T) {
	g := NewReversibleGraph(1)
	require.Nil(t, g.Node(1))
	g.AddNode(TransactionNode{id: 1})
	require.NotNil(t, g.Node(1))
}

func TestReversibleGraph_Nodes(t *testing.T) {
	g := NewReversibleGraph(5)
	require.Equal(t, graph.Empty, g.Nodes())
	g = createSimpleGraph()
	nodes := g.Nodes()
	require.Equal(t, 4, nodes.Len())
}

func TestReversibleGraph_RemoveEdge(t *testing.T) {
	g := createSimpleGraph()

	// test no-ops
	edgeCount := g.Edges().Len()
	g.RemoveEdge(100, 1)
	g.RemoveEdge(1, 100)
	require.Equal(t, edgeCount, g.Edges().Len())

	// remove one edge
	require.True(t, g.HasEdgeBetween(1, 2))
	g.RemoveEdge(1, 2)
	require.False(t, g.HasEdgeBetween(1, 2))
}

func TestReversibleGraph_RemoveNode(t *testing.T) {
	g := createSimpleGraph()

	// test no-ops
	nodeCount := g.Nodes().Len()
	g.RemoveNode(100)
	require.Equal(t, nodeCount, g.Nodes().Len())

	// remove one node
	require.NotNil(t, g.Node(1))
	g.RemoveNode(1)
	require.Nil(t, g.Node(1))
}

func TestReversibleGraph_SetEdge(t *testing.T) {
	require.Panics(t, func() {
		NewReversibleGraph(1).SetEdge(AddressEdge{F: TransactionNode{id: 1}, T: TransactionNode{id: 1}})
	})

	require.NotPanics(t, func() {
		createSimpleGraph()
	})
}

func TestReversibleGraph_SetEdgeWithoutOverwrite(t *testing.T) {
	require.Panics(t, func() {
		NewReversibleGraph(1).SetEdgeWithoutOverwrite(TransactionNode{id: 1}, TransactionNode{id: 1}, 0)
	})

	require.NotPanics(t, func() {
		g := NewReversibleGraph(5)
		g.SetEdgeWithoutOverwrite(TransactionNode{id: 1}, TransactionNode{id: 2}, 0)
		g.SetEdgeWithoutOverwrite(TransactionNode{id: 1}, TransactionNode{id: 3}, 0)
		g.SetEdgeWithoutOverwrite(TransactionNode{id: 3}, TransactionNode{id: 4}, 0)
	})
}

func TestReversibleGraph_To(t *testing.T) {
	g := NewReversibleGraph(5)
	require.Equal(t, graph.Empty, g.To(1))
	g.SetReverse(true)
	require.Equal(t, graph.Empty, g.To(1))

	g = createSimpleGraph()

	tests := []struct {
		id           int64
		reversed     bool
		wantNumNodes int
	}{
		{id: 1, reversed: true, wantNumNodes: 2},
		{id: 3, reversed: true, wantNumNodes: 1},
		{id: 4, reversed: true, wantNumNodes: 0},

		{id: 1, reversed: false, wantNumNodes: 0},
		{id: 2, reversed: false, wantNumNodes: 1},
		{id: 3, reversed: false, wantNumNodes: 1},
		{id: 4, reversed: false, wantNumNodes: 1},
	}
	for _, tt := range tests {
		g.SetReverse(tt.reversed)
		require.Equal(t, tt.wantNumNodes, g.To(tt.id).Len())
	}
}

type testEdgeInput struct {
	f TransactionNode
	t []graph.Node
	a []int64
}

func getSetEdgeData() []testEdgeInput {
	r := rand.New(rand.NewPCG(46434, 7634354)) //nolint:gosec

	const txCount = int64(1000)
	inputs := make([]testEdgeInput, txCount)

	// generate random test data
	for txUID := range txCount {
		// create up to 100 inputs
		numInputs := txCount + max(1, r.Int64N(100))
		froms := make([]graph.Node, numInputs)
		as := make([]int64, numInputs)

		for i := range numInputs {
			froms[i] = TransactionNode{id: txCount + r.Int64N(100)}
			as[i] = txCount + r.Int64N(100)
		}
		inputs[txUID] = testEdgeInput{TransactionNode{id: txUID}, froms, as}
	}

	return inputs
}

func TestResultsOfSetEdgeWithoutOverwrite(t *testing.T) {
	g1 := NewReversibleGraph(1000)
	g2 := NewReversibleGraph(1000)
	inputs := getSetEdgeData()

	for nUID := range inputs {
		for i := range inputs[nUID].t {
			g1.SetEdgeWithoutOverwrite(inputs[nUID].f, inputs[nUID].t[i], inputs[nUID].a[i])
		}
		g2.SetEdgesWithoutOverwrite(inputs[nUID].f, inputs[nUID].t, inputs[nUID].a)
	}

	require.True(t, g1.IsEqual(g2))
}

func BenchmarkReversibleGraph_SetEdgeWithoutOverwrite(b *testing.B) {
	g := NewReversibleGraph(1000)
	inputs := getSetEdgeData()

	for range b.N {
		for i := range inputs {
			for j := range inputs[i].t {
				g.SetEdgeWithoutOverwrite(inputs[i].f, inputs[i].t[j], inputs[i].a[j])
			}
		}
	}
}

func BenchmarkReversibleGraph_SetEdgesWithoutOverwrite(b *testing.B) {
	g := NewReversibleGraph(1000)
	inputs := getSetEdgeData()

	for range b.N {
		for i := range inputs {
			g.SetEdgesWithoutOverwrite(inputs[i].f, inputs[i].t, inputs[i].a)
		}
	}
}
