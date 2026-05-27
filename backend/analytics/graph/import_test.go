// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics"
)

func Test_addSingleNodes(t *testing.T) {
	g := NewReversibleGraph(10)
	tests := []struct {
		nodes   []analytics.Node
		wantErr bool
	}{
		{
			nodes: []analytics.Node{{
				UID: "0x123",
				Block: []struct {
					TS time.Time `json:"ts"`
				}{{TS: time.Now()}},
			}},
			wantErr: false,
		},
		{
			nodes:   []analytics.Node{{UID: "12", Block: nil}},
			wantErr: true,
		},
		{
			nodes: nil,
		},
	}
	for _, tt := range tests {
		if tt.wantErr {
			require.Error(t, addSingleNodes(g, tt.nodes))
		} else {
			require.NoError(t, addSingleNodes(g, tt.nodes))
		}
	}
}

func Test_upsertSingleNodes(t *testing.T) {
	g := NewReversibleGraph(10)
	tests := []struct {
		nodes   []analytics.Node
		wantErr bool
	}{
		{
			nodes: []analytics.Node{{
				UID: "0x123",
				Block: []struct {
					TS time.Time `json:"ts"`
				}{{TS: time.Now()}},
			}},
			wantErr: false,
		},
		{
			nodes: []analytics.Node{{
				UID:   "12",
				Block: nil,
			}},
			wantErr: true,
		},
		{
			nodes: nil,
		},
	}
	for _, tt := range tests {
		if tt.wantErr {
			require.Error(t, upsertSingleNodes(g, tt.nodes))
		} else {
			require.NoError(t, upsertSingleNodes(g, tt.nodes))
		}
	}
}

func Test_addEdges(t *testing.T) {
	g := NewReversibleGraph(10)
	tests := []struct {
		nodes   []analytics.ConnectedNode
		wantErr bool
	}{
		{
			nodes: []analytics.ConnectedNode{{
				UID:    "0x123",
				Inputs: nil,
			}},
			wantErr: false,
		},
		{
			nodes: []analytics.ConnectedNode{{
				UID: "0x1234",
				Inputs: []struct {
					Address          string
					InputTransaction string
				}{{
					Address:          "",
					InputTransaction: "",
				}},
			}},
			wantErr: true,
		},
		{
			nodes: []analytics.ConnectedNode{{
				UID: "0x1235",
				Inputs: []struct {
					Address          string
					InputTransaction string
				}{{
					Address:          "0x1",
					InputTransaction: "",
				}},
			}},
			wantErr: true,
		},
		{
			nodes: []analytics.ConnectedNode{{
				UID: "0x1235",
				Inputs: []struct {
					Address          string
					InputTransaction string
				}{{
					Address:          "",
					InputTransaction: "0x1",
				}},
			}},
			wantErr: true,
		},
		{
			nodes: []analytics.ConnectedNode{{
				UID: "0x1235",
				Inputs: []struct {
					Address          string
					InputTransaction string
				}{{
					Address:          "0x1",
					InputTransaction: "0x2",
				}},
			}},
			wantErr: false,
		},
		{
			nodes: []analytics.ConnectedNode{{
				UID: "12",
			}},
			wantErr: true,
		},
		{
			nodes: nil,
		},
	}
	for _, tt := range tests {
		if tt.wantErr {
			require.Error(t, addEdges(g, tt.nodes))
		} else {
			require.NoError(t, addEdges(g, tt.nodes))
		}
	}
}

func Test_pruneNodes(t *testing.T) {
	g := NewReversibleGraph(1)
	require.NoError(t, pruneNodes(g))

	// unset time should cause error
	g.AddNode(TransactionNode{
		TS:   time.Time{},
		id:   0,
		Type: constants.TypeDashMixing,
	})
	require.Error(t, pruneNodes(g))

	// pruning should have not an effect, as all nodes are connected
	g = newTestGraph()
	numNodes := g.Nodes().Len()
	require.NoError(t, pruneNodes(g))
	require.Equal(t, numNodes, g.Nodes().Len())

	// add unconnected node, therefore pruning should show an effect
	g.AddNode(TransactionNode{
		TS: time.Now(),
		id: 10000,
	})
	numNodes = g.Nodes().Len()
	require.NoError(t, pruneNodes(g))
	require.Equal(t, numNodes-1, g.Nodes().Len())
}

func Test_verifyTransactionGraph(t *testing.T) {
	g := NewReversibleGraph(1)
	require.NoError(t, verifyTransactionGraph(g))

	// unset time should cause error
	g.AddNode(TransactionNode{TS: time.Time{}})
	require.Error(t, verifyTransactionGraph(g))

	// add unconnected node, therefore pruning should show an effect
	g.AddNode(TransactionNode{
		TS: time.Now(),
		id: 10000,
	})
	require.Error(t, verifyTransactionGraph(g))
}

func Test_loadTransactions(t *testing.T) {
	// testdata contains 132 mixing transactions and 557 input transactions
	dbHandle := db.GetDBConnection(t, db.UsePrivacyFile)
	g := NewReversibleGraph(2000)
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	require.NoError(t, loadTransactions(ctx, dbHandle, g, 50000, 0, constants.TypeDashMixing, false))
	require.Equal(t, 689, g.Nodes().Len())

	g = NewReversibleGraph(1)
	require.NoError(t, loadTransactions(ctx, dbHandle, g, 50000, 0, constants.TypeDashDestination, false))
	require.Equal(t, 8, g.Nodes().Len())

	g = NewReversibleGraph(1)
	require.NoError(t, loadTransactions(ctx, dbHandle, g, 50000, 0, constants.TypeDashOrigin, true))
	require.Equal(t, 1, g.Nodes().Len())

	g = NewReversibleGraph(1)
	require.NoError(t, loadTransactions(ctx, dbHandle, g, 50000, 0, constants.TypeDashCC, true))
	require.Equal(t, 0, g.Nodes().Len())
}

func TestLoadDashTransactionGraph(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UsePrivacyFile)
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	graph, err := LoadDashTransactionGraph(ctx, dbHandle, 0)
	require.NoError(t, err)
	// not pruned yet
	require.Equal(t, 696, graph.Nodes().Len())
}

func TestLoadTransactionGraph(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UsePrivacyFile)
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	graph, err := LoadTransactionGraph(ctx, NewDashConfig(), dbHandle, 0)
	require.NoError(t, err)
	// mixing_count=132 origin_count=1 destination_count=1 cc_count=0
	require.Equal(t, 134, graph.Nodes().Len())
}
