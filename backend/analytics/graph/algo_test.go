package graph

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestErrNodeNotFound(t *testing.T) {
	require.NotNil(t, ErrNodeNotFound(1))
	require.NotNil(t, ErrNodeNotFound(-1))
	require.NotNil(t, ErrNodeNotFound(0))
}

func TestToHex(t *testing.T) {
	tests := []struct {
		arg  int64
		want string
	}{
		{arg: 5, want: "0x5"},
		{arg: 0, want: "0x0"},
		{arg: 11, want: "0xb"},
		{arg: 123456, want: "0x1e240"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, ToHex(tt.arg))
	}
}

func TestToInteger(t *testing.T) {
	tests := []struct {
		hexString string
		want      int64
		wantErr   bool
	}{
		{hexString: "0x1e240", want: 123456, wantErr: false},
		{hexString: "0x1", want: 1, wantErr: false},
		{hexString: "0x0", want: 0, wantErr: false},
		{hexString: "0xb", want: 11, wantErr: false},
		{hexString: "0x", want: -1, wantErr: true},
		{hexString: "", want: -1, wantErr: true},
		// This is a valid input for performance reasons, but returns an unexpected result
		{hexString: "x12", want: 2, wantErr: false},
	}
	for _, tt := range tests {
		integer, err := ToInteger(tt.hexString)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, integer)
		}
	}
}

func TestHasSpendingGap(t *testing.T) {
	graph := NewReversibleGraph(10)
	now := time.Now()
	for i := 0; i < 10; i++ {
		graph.AddNode(TransactionNode{
			TS:          now.Add(time.Hour * time.Duration(i)),
			id:          int64(i),
			PrivacyType: 0,
		})
	}

	tests := []struct {
		edge     AddressEdge
		want     bool
		reversed bool
	}{
		{edge: AddressEdge{F: graph.Node(0), T: graph.Node(0)}, want: false},
		{edge: AddressEdge{F: graph.Node(1), T: graph.Node(0)}, want: false},
		{edge: AddressEdge{F: graph.Node(2), T: graph.Node(0)}, want: false},
		{edge: AddressEdge{F: graph.Node(3), T: graph.Node(0)}, want: false},
		{edge: AddressEdge{F: graph.Node(4), T: graph.Node(0)}, want: false},
		{edge: AddressEdge{F: graph.Node(5), T: graph.Node(0)}, want: false},
		{edge: AddressEdge{F: graph.Node(6), T: graph.Node(0)}, want: false},
		{edge: AddressEdge{F: graph.Node(7), T: graph.Node(0)}, want: false},
		{edge: AddressEdge{F: graph.Node(8), T: graph.Node(0)}, want: true},
		{edge: AddressEdge{F: graph.Node(9), T: graph.Node(0)}, want: true},
		{edge: AddressEdge{F: graph.Node(9), T: graph.Node(1)}, want: true},
		{edge: AddressEdge{F: graph.Node(0), T: graph.Node(8)}, want: true, reversed: true},
		{edge: AddressEdge{F: graph.Node(0), T: graph.Node(7)}, want: false, reversed: true},
	}
	for _, tt := range tests {
		graph.SetReverse(tt.reversed)
		require.Equal(t, tt.want, HasSpendingGap(graph, tt.edge))
	}
}

func TestCheckAddressExclusions(t *testing.T) {
	type args struct {
		exclusions map[int64]bool
		edge       AddressEdge
	}
	tests := []struct {
		args args
		want bool
	}{
		{
			args: args{exclusions: nil},
			want: true,
		},
		{
			args: args{
				exclusions: map[int64]bool{1: true},
				edge:       AddressEdge{},
			},
			want: false,
		},
		{
			args: args{
				exclusions: map[int64]bool{1: true, 2: true},
				edge:       AddressEdge{AddressUIDs: []int64{3}},
			},
			want: true,
		},
		{
			args: args{
				exclusions: map[int64]bool{1: true, 2: true},
				edge:       AddressEdge{AddressUIDs: []int64{1, 2}},
			},
			want: false,
		},
		{
			args: args{
				exclusions: map[int64]bool{1: true, 2: true},
				edge:       AddressEdge{AddressUIDs: []int64{2}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, CheckAddressExclusions(tt.args.exclusions, tt.args.edge))
	}
}

func TestReverseLookupByID(t *testing.T) {
	// *** Test Graph structure  ***
	//              ┌──────┐
	//              │  5   ├─────┐
	//              └──────┘     │
	// ┌──────┐               ┌──┴───┐
	// │  1   ├────┐       ┌──┤  8   ├───────┐
	// └──────┘    └┬──────┤  └──────┘       ├──────┐
	//              │  6   │                 │  11  │
	// ┌──────┬─────┴──────┘   ┌──────┬──────┴──────┘
	// │  2   │                │  9   │
	// └──────┘          ┌─────┴──────┘
	//                   │
	// ┌──────┬────┬─────┴┐    ┌──────┬──────┬──────┐
	// │  3   │    │  7   │    │  10  │      │  12  │
	// └──────┤    └──────┴────┴──────┘      └──────┘
	//        │
	// ┌──────┤
	// │  4   │
	// └──────┘
	graph := NewReversibleGraph(12)
	now := time.Now()

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 11, PrivacyType: 0},
		TransactionNode{TS: now, id: 8, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 8, PrivacyType: 0},
		TransactionNode{TS: now, id: 5, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 8, PrivacyType: 0},
		TransactionNode{TS: now, id: 6, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 6, PrivacyType: 0},
		TransactionNode{TS: now, id: 1, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 6, PrivacyType: 0},
		TransactionNode{TS: now, id: 2, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 11, PrivacyType: 0},
		TransactionNode{TS: now, id: 9, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 12, PrivacyType: 0},
		TransactionNode{TS: now, id: 10, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 10, PrivacyType: 0},
		TransactionNode{TS: now, id: 7, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 9, PrivacyType: 0},
		TransactionNode{TS: now, id: 7, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 7, PrivacyType: 0},
		TransactionNode{TS: now, id: 3, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 3, PrivacyType: 0},
		TransactionNode{TS: now, id: 4, PrivacyType: 0}, 0)
	type args struct {
		nodeID              int64
		maxLookBackTime     time.Duration
		addressExclusions   []string
		excludeSpendingGaps bool
	}
	tests := []struct {
		args    args
		want    map[string]bool
		wantErr bool
	}{
		{
			args: args{
				nodeID:              11,
				maxLookBackTime:     time.Hour * 24 * 60,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0x5": true, "0x1": true, "0x2": true, "0x4": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              12,
				maxLookBackTime:     time.Hour * 24 * 60,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0x4": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              6,
				maxLookBackTime:     time.Hour * 24 * 60,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0x1": true, "0x2": true},
			wantErr: false,
		},
		{
			args:    args{nodeID: 500},
			wantErr: true,
		},
		{
			args:    args{nodeID: -1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		results, err := ReverseLookupByID(graph, tt.args.nodeID, tt.args.maxLookBackTime, nil, false)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}
}
