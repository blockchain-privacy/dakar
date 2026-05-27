// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"gitlab.com/blockchain-privacy/dakar/constants"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestErrNodeNotFound(t *testing.T) {
	require.Error(t, ErrNodeNotFound(1))
	require.Error(t, ErrNodeNotFound(-1))
	require.Error(t, ErrNodeNotFound(0))
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
	for i := range 10 {
		graph.AddNode(TransactionNode{
			TS:   now.Add(time.Hour * time.Duration(i)),
			id:   int64(i),
			Type: constants.TypeDashMixing,
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

// newTestGraph returns a new graph with the structure shown below.
// Each layer of nodes has a time difference of 1 hour.
// Thus, node 11 is the most recent while node 1 has a 3-hour time difference.
func newTestGraph() *ReversibleGraph {
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
	// └──────┤    └──────┴────┴──┬───┘      └──────┘
	//        │           ┌───────┘
	// ┌──────┤        ┌──┴───┐
	// │  4   │        │  13  │
	// └──────┘        └──────┘
	graph := NewReversibleGraph(12)
	now := time.Now()

	threeHoursOld := now.Add(-time.Hour * 3)
	twoHoursOld := now.Add(-time.Hour * 2)
	oneHourOld := now.Add(-time.Hour * 1)

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 11, Type: constants.TypeDashDestination},
		TransactionNode{TS: oneHourOld, id: 8, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: oneHourOld, id: 8, Type: constants.TypeDashMixing},
		TransactionNode{TS: twoHoursOld, id: 5, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: oneHourOld, id: 8, Type: constants.TypeDashMixing},
		TransactionNode{TS: twoHoursOld, id: 6, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: twoHoursOld, id: 6, Type: constants.TypeDashMixing},
		TransactionNode{TS: threeHoursOld, id: 1, Type: constants.TypeDashOrigin}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: twoHoursOld, id: 6, Type: constants.TypeDashMixing},
		TransactionNode{TS: threeHoursOld, id: 2, Type: constants.TypeDashOrigin}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 11, Type: constants.TypeDashDestination},
		TransactionNode{TS: oneHourOld, id: 9, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 12, Type: constants.TypeDashDestination},
		TransactionNode{TS: oneHourOld, id: 10, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: oneHourOld, id: 10, Type: constants.TypeDashMixing},
		TransactionNode{TS: twoHoursOld, id: 7, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: oneHourOld, id: 10, Type: constants.TypeDashMixing},
		TransactionNode{TS: oneHourOld, id: 13, Type: constants.TypeDashOrigin}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: oneHourOld, id: 9, Type: constants.TypeDashMixing},
		TransactionNode{TS: twoHoursOld, id: 7, Type: constants.TypeDashMixing}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: twoHoursOld, id: 7, Type: constants.TypeDashMixing},
		TransactionNode{TS: threeHoursOld, id: 3, Type: constants.TypeDashMixing}, 123456)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: threeHoursOld, id: 3, Type: constants.TypeDashMixing},
		TransactionNode{TS: threeHoursOld, id: 4, Type: constants.TypeDashOrigin}, 0)
	return graph
}

func TestReverseLookupByID(t *testing.T) {
	graph := newTestGraph()
	type args struct {
		nodeID              int64
		maxLookBackTime     time.Duration
		maxDepth            int
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
				maxLookBackTime:     time.Hour * 24,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0x1": true, "0x2": true, "0x4": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              12,
				maxLookBackTime:     time.Hour * 24,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0x4": true, "0xd": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              6,
				maxLookBackTime:     time.Hour * 24,
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
		// Limited look back
		{
			args: args{
				nodeID:              12,
				maxLookBackTime:     time.Hour * 1,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0xd": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              6,
				maxLookBackTime:     time.Minute * 30,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		results, err := ReverseLookupByID(graph, tt.args.nodeID, tt.args.maxLookBackTime, tt.args.maxDepth,
			tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}
}

func TestReverseLookup(t *testing.T) {
	graph := newTestGraph()
	type args struct {
		nodeID              int64
		maxLookBackTime     time.Duration
		maxDepth            int
		excludeSpendingGaps bool
	}
	tests := []struct {
		args    args
		want    map[string]bool
		wantErr bool
	}{
		{
			args: args{
				nodeID:          11,
				maxLookBackTime: time.Hour * 24,
			},
			want:    map[string]bool{"0x1": true, "0x2": true, "0x4": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          12,
				maxLookBackTime: time.Hour * 24,
			},
			want:    map[string]bool{"0x4": true, "0xd": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          6,
				maxLookBackTime: time.Hour * 24,
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
		// Limited look back
		{
			args: args{
				nodeID:          11,
				maxLookBackTime: time.Hour * 2,
			},
			want:    map[string]bool{},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          12,
				maxLookBackTime: time.Hour * 1,
			},
			want:    map[string]bool{"0xd": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          6,
				maxLookBackTime: time.Minute * 30,
			},
			want:    map[string]bool{},
			wantErr: false,
		},
		// depth
		{
			args: args{
				nodeID:   11,
				maxDepth: -100,
			},
			wantErr: true,
		},
		// max depth == 0: just limited by time
		{
			args: args{
				nodeID:          11,
				maxLookBackTime: time.Hour * 24,
				maxDepth:        0,
			},
			want:    map[string]bool{"0x1": true, "0x2": true, "0x4": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          11,
				maxLookBackTime: time.Hour * 24,
				maxDepth:        1,
			},
			want:    map[string]bool{},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          11,
				maxLookBackTime: time.Hour * 24,
				maxDepth:        2,
			},
			want:    map[string]bool{},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          11,
				maxLookBackTime: time.Hour * 24,
				maxDepth:        3,
			},
			want:    map[string]bool{"0x1": true, "0x2": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          11,
				maxLookBackTime: time.Hour * 24,
				maxDepth:        4,
			},
			want:    map[string]bool{"0x1": true, "0x2": true, "0x4": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          12,
				maxLookBackTime: time.Hour * 24,
				maxDepth:        4,
			},
			want:    map[string]bool{"0x4": true, "0xd": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:          3,
				maxLookBackTime: time.Hour * 24,
				maxDepth:        1,
			},
			want:    map[string]bool{"0x4": true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		results, err := ReverseLookup(graph, ToHex(tt.args.nodeID), tt.args.maxLookBackTime, tt.args.maxDepth,
			tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}
}

func TestForwardLookup(t *testing.T) {
	graph := newTestGraph()
	type args struct {
		nodeID              int64
		maxLookForwardTime  time.Duration
		maxDepth            int
		excludeSpendingGaps bool
	}
	tests := []struct {
		args    args
		want    map[string]bool
		wantErr bool
	}{
		{
			args: args{
				nodeID:              1,
				maxLookForwardTime:  time.Hour * 5,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              4,
				maxLookForwardTime:  time.Hour * 5,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true, ToHex(12): true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              1,
				maxDepth:            2,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              3,
				maxDepth:            2,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true, ToHex(12): true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              4,
				maxDepth:            3,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true, ToHex(12): true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              4,
				maxDepth:            2,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              6,
				maxDepth:            1,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		results, err := ForwardLookup(graph, ToHex(tt.args.nodeID), tt.args.maxLookForwardTime, tt.args.maxDepth,
			tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}
}

func Test_getConnectedNodes(t *testing.T) {
	graph := newTestGraph()
	tests := []struct {
		nodeUID int64
		nodes   map[int64]bool
		want    []int64
		wantErr bool
	}{
		{
			nodeUID: 6,
			nodes:   map[int64]bool{6: true, 1: true, 8: true, 11: true, 12: true, 4: true},
			want:    []int64{6, 1, 8, 11},
			wantErr: false,
		},
		{
			nodeUID: 5,
			nodes:   map[int64]bool{1: true, 8: true, 11: true, 12: true, 4: true},
			want:    []int64{5, 8, 11},
			wantErr: false,
		},
		{
			nodeUID: 3,
			nodes:   map[int64]bool{1: true, 8: true, 11: true, 12: true},
			want:    []int64{3},
			wantErr: false,
		},
		{
			nodeUID: 7,
			nodes:   map[int64]bool{7: true, 2: true, 10: true, 6: true, 11: true, 13: true},
			want:    []int64{7, 10, 13},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		results, err := getConnectedNodes(graph, tt.nodeUID, tt.nodes)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			slices.Sort(results)
			slices.Sort(tt.want)
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}
}
