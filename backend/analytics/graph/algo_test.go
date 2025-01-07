package graph

import (
	"backend/constants"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"slices"
	"testing"
	"time"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	testhelper.RunDgraphTests(m, &dbHandle.DB)
}

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
	//        │                   │
	// ┌──────┤                ┌──┴───┐
	// │  4   │                │  13  │
	// └──────┘                └──────┘
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
				maxLookBackTime:     time.Hour * 24,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0x1": true, "0x2": true, "0x4": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              12,
				maxLookBackTime:     time.Hour * 24,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0x4": true, "0xd": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              6,
				maxLookBackTime:     time.Hour * 24,
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
		// Limited look back
		{
			args: args{
				nodeID:              12,
				maxLookBackTime:     time.Hour * 1,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0xd": true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              6,
				maxLookBackTime:     time.Minute * 30,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{},
			wantErr: false,
		},
		// address exclusion
		{
			args: args{
				nodeID:              12,
				maxLookBackTime:     time.Hour * 24,
				addressExclusions:   []string{ToHex(123456)},
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0xd": true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		results, err := ReverseLookupByID(graph, tt.args.nodeID, tt.args.maxLookBackTime, tt.args.maxDepth,
			tt.args.addressExclusions, tt.args.excludeSpendingGaps)
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
		// address exclusion
		{
			args: args{
				nodeID:            12,
				maxLookBackTime:   time.Hour * 24,
				addressExclusions: []string{ToHex(123456)},
			},
			want:    map[string]bool{"0xd": true},
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
			tt.args.addressExclusions, tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}
}

func TestGetInputTransactions(t *testing.T) {
	graph := newTestGraph()
	tests := []struct {
		uid     string
		want    []string
		wantErr bool
	}{
		{
			uid:     ToHex(11),
			want:    []string{ToHex(8), ToHex(9)},
			wantErr: false,
		},
		{
			uid:     ToHex(12),
			want:    []string{ToHex(10)},
			wantErr: false,
		},
		{
			uid:     "a",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		results, err := GetInputTransactions(graph, tt.uid)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			// sort results so order stays the same
			slices.Sort(results)
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
				nodeID:              1,
				maxLookForwardTime:  time.Hour * 5,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              4,
				maxLookForwardTime:  time.Hour * 5,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true, ToHex(12): true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		results, err := ForwardLookup(graph, ToHex(tt.args.nodeID), tt.args.maxLookForwardTime, tt.args.maxDepth,
			tt.args.addressExclusions, tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}
}
