package graph

import (
	"backend/constants"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
	"time"
)

// newDestinationGraph creates a new graph with the following destination transaction timestamps:
// - 2020-01-01 10:00 - 11:00, 2020-01-02 15:00 - 17:00 - set 1
// - 2020-01-01 08:00 - 09:00, 2020-01-02 13:00 - 16:00 - set 1
// - 2020-02-01 10:00 - 11:00, 2020-02-02 15:00 - 17:00 - set 2
// - 2020-02-01 08:00 - 09:00, 2020-02-02 13:00 - 16:00 - set 2
// - 2020-03-01 10:00 - 11:00, 2020-03-02 15:00 - 17:00 - set 3
func newDestinationGraph(t *testing.T) *ReversibleGraph {
	graph := NewReversibleGraph(5)

	tx1ts1, err := time.Parse(time.RFC3339, "2020-01-01T10:00:00Z")
	require.NoError(t, err)
	tx1ts2, err := time.Parse(time.RFC3339, "2020-01-01T11:00:00Z")
	require.NoError(t, err)
	tx1ts3, err := time.Parse(time.RFC3339, "2020-01-02T15:00:00Z")
	require.NoError(t, err)
	tx1ts4, err := time.Parse(time.RFC3339, "2020-01-02T17:00:00Z")
	require.NoError(t, err)

	tx2ts1, err := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")
	require.NoError(t, err)
	tx2ts2, err := time.Parse(time.RFC3339, "2020-01-01T09:00:00Z")
	require.NoError(t, err)
	tx2ts3, err := time.Parse(time.RFC3339, "2020-01-02T13:00:00Z")
	require.NoError(t, err)
	tx2ts4, err := time.Parse(time.RFC3339, "2020-01-02T16:00:00Z")
	require.NoError(t, err)

	tx3ts1, err := time.Parse(time.RFC3339, "2020-02-01T10:00:00Z")
	require.NoError(t, err)
	tx3ts2, err := time.Parse(time.RFC3339, "2020-02-01T11:00:00Z")
	require.NoError(t, err)
	tx3ts3, err := time.Parse(time.RFC3339, "2020-02-02T15:00:00Z")
	require.NoError(t, err)
	tx3ts4, err := time.Parse(time.RFC3339, "2020-02-02T17:00:00Z")
	require.NoError(t, err)

	tx4ts1, err := time.Parse(time.RFC3339, "2020-02-01T08:00:00Z")
	require.NoError(t, err)
	tx4ts2, err := time.Parse(time.RFC3339, "2020-02-01T09:00:00Z")
	require.NoError(t, err)
	tx4ts3, err := time.Parse(time.RFC3339, "2020-02-02T13:00:00Z")
	require.NoError(t, err)
	tx4ts4, err := time.Parse(time.RFC3339, "2020-02-02T16:00:00Z")
	require.NoError(t, err)

	tx5ts1, err := time.Parse(time.RFC3339, "2020-03-01T10:00:00Z")
	require.NoError(t, err)
	tx5ts2, err := time.Parse(time.RFC3339, "2020-03-01T11:00:00Z")
	require.NoError(t, err)
	tx5ts3, err := time.Parse(time.RFC3339, "2020-03-02T15:00:00Z")
	require.NoError(t, err)
	tx5ts4, err := time.Parse(time.RFC3339, "2020-03-02T17:00:00Z")
	require.NoError(t, err)

	now := time.Now()

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 1, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx1ts1, id: 10, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 1, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx1ts2, id: 11, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 1, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx1ts3, id: 12, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 1, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx1ts4, id: 13, Type: constants.TypeDashMixing}, 0)

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 2, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx2ts1, id: 20, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 2, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx2ts2, id: 21, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 2, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx2ts3, id: 22, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 2, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx2ts4, id: 23, Type: constants.TypeDashMixing}, 0)

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 3, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx3ts1, id: 30, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 3, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx3ts2, id: 31, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 3, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx3ts3, id: 32, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 3, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx3ts4, id: 33, Type: constants.TypeDashMixing}, 0)

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 4, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx4ts1, id: 40, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 4, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx4ts2, id: 41, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 4, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx4ts3, id: 42, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 4, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx4ts4, id: 43, Type: constants.TypeDashMixing}, 0)

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 5, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx5ts1, id: 50, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 5, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx5ts2, id: 51, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 5, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx5ts3, id: 52, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 5, Type: constants.TypeDashDestination},
		TransactionNode{TS: tx5ts4, id: 53, Type: constants.TypeDashMixing}, 0)

	return graph
}

func TestSpendingFingerprint(t *testing.T) {
	g := newDestinationGraph(t)

	tests := []struct {
		uid             string
		wantFingerprint []string
		wantErr         bool
	}{
		{
			uid:             "",
			wantFingerprint: nil,
			wantErr:         true,
		},
		{
			uid:             ToHex(1),
			wantFingerprint: []string{ToHex(2)},
			wantErr:         false,
		},
		{
			uid:             ToHex(2),
			wantFingerprint: []string{ToHex(1)},
			wantErr:         false,
		},
		{
			uid:             ToHex(3),
			wantFingerprint: []string{ToHex(4)},
			wantErr:         false,
		},
		{
			uid:             ToHex(4),
			wantFingerprint: []string{ToHex(3)},
			wantErr:         false,
		},
		{
			uid:             ToHex(5),
			wantFingerprint: []string{},
			wantErr:         false,
		},
		// transaction does not exist in graph
		{
			uid:             ToHex(100),
			wantFingerprint: []string{},
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		fingerprints, err := SpendingFingerprint(g, tt.uid)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			fingerprintUIDs := make([]string, len(fingerprints))
			for j, f := range fingerprints {
				fingerprintUIDs[j] = f.TransactionUID
			}

			sort.Strings(fingerprintUIDs)

			require.Equal(t, tt.wantFingerprint, fingerprintUIDs)
		}
	}
}
