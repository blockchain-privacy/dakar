package graph

import (
	"backend/db"
	"backend/db/analytics"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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
		TS:          time.Time{},
		id:          0,
		PrivacyType: 0,
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

func Test_loadOriginTransactions(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	g := NewReversibleGraph(1)
	// testdata contains 1 origin transaction
	db.SetupDB(t, dbHandle, testhelper.UsePrivacyFile)
	require.NoError(t, loadOriginTransactions(dbHandle, g, 0))
	require.Equal(t, 1, g.Nodes().Len())
}

func Test_loadDestinationTransactions(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	g := NewReversibleGraph(1)
	// testdata contains 1 destination transaction with 7 input transactions
	db.SetupDB(t, dbHandle, testhelper.UsePrivacyFile)
	require.NoError(t, loadDestinationTransactions(dbHandle, g, 0))
	require.Equal(t, 8, g.Nodes().Len())
}

func Test_loadMixingTransactions(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	g := NewReversibleGraph(132)
	// testdata contains 132 mixing transactions and 557 input transactions
	db.SetupDB(t, dbHandle, testhelper.UsePrivacyFile)
	require.NoError(t, loadMixingTransactions(dbHandle, g, 0))
	require.Equal(t, 689, g.Nodes().Len())
}

func Test_loadCCTransactions(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	g := NewReversibleGraph(132)
	// testdata contains 0 cc transactions
	db.SetupDB(t, dbHandle, testhelper.UsePrivacyFile)
	require.NoError(t, loadCCTransactions(dbHandle, g, 0))
	require.Equal(t, 0, g.Nodes().Len())
}

func TestLoadTransactionGraph(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, testhelper.UsePrivacyFile)

	graph, err := LoadTransactionGraph(dbHandle, 0)
	require.NoError(t, err)
	// mixing_count=132 origin_count=1 destination_count=1 cc_count=0
	require.Equal(t, 134, graph.Nodes().Len())
}
