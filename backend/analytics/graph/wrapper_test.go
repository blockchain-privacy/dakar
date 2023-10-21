package graph

import (
	"backend/db"
	"backend/db/status"
	"backend/testhelper"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"os"
	"slices"
	"sort"
	"sync"
	"testing"
	"time"
)

// unregisterCollectors unregisters all collectors of the wrapper.
// This is needed because collectors can not be registered twice with the same default config.
func unregisterCollectors(w *Wrapper) {
	if w == nil {
		return
	}

	prometheus.Unregister(w.blocks)
	prometheus.Unregister(w.transactions)
	prometheus.Unregister(w.blockHeight)
}

func TestNewWrapper(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)
	require.NotNil(t, w)
	require.NotNil(t, w.transactionGraphMutex)
}

func TestWrapper_IsTransactionGraphLoaded(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	require.False(t, w.IsTransactionGraphLoaded())
	w.transactionGraph = NewReversibleGraph(1)
	require.True(t, w.IsTransactionGraphLoaded())

	// should be thread safe
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			require.True(t, w.IsTransactionGraphLoaded())
		}()
	}

	wg.Wait()
}

func TestWrapper_ReverseLookup(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	// transaction graph not loaded -> should produce error
	_, err := w.ReverseLookup("", 0, nil, false)
	require.Error(t, err)

	w.transactionGraph = newTestGraph()

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
				maxLookBackTime:     time.Hour * 24,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{"0x5": true, "0x1": true, "0x2": true, "0x4": true},
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
	}

	for _, tt := range tests {
		results, err := w.ReverseLookup(ToHex(tt.args.nodeID), tt.args.maxLookBackTime, tt.args.addressExclusions, tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}

	// test thread safety
	tt := tests[0]
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := w.ReverseLookup(ToHex(tt.args.nodeID), tt.args.maxLookBackTime, tt.args.addressExclusions, tt.args.excludeSpendingGaps)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, results)
			}
		}()
	}
	wg.Wait()
}

func TestWrapper_ForwardLookup(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	// transaction graph not loaded -> should produce error
	_, err := w.ForwardLookup("", "", nil, false)
	require.Error(t, err)

	w.transactionGraph = newTestGraph()

	type args struct {
		nodeID              int64
		targetID            int64
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
				targetID:            11,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true},
			wantErr: false,
		},
		{
			args: args{
				nodeID:              4,
				targetID:            11,
				addressExclusions:   nil,
				excludeSpendingGaps: false,
			},
			want:    map[string]bool{ToHex(11): true, ToHex(12): true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		results, err := w.ForwardLookup(ToHex(tt.args.nodeID), ToHex(tt.args.targetID), tt.args.addressExclusions, tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}

	// test thread safety
	tt := tests[0]
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := w.ForwardLookup(ToHex(tt.args.nodeID), ToHex(tt.args.targetID), tt.args.addressExclusions, tt.args.excludeSpendingGaps)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, results)
			}
		}()
	}
	wg.Wait()
}

func TestWrapper_ForwardLookupByTime(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	// transaction graph not loaded -> should produce error
	_, err := w.ForwardLookupByTime("", 0, nil, false)
	require.Error(t, err)

	w.transactionGraph = newTestGraph()

	type args struct {
		nodeID              int64
		maxLookForwardTime  time.Duration
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
		results, err := w.ForwardLookupByTime(ToHex(tt.args.nodeID), tt.args.maxLookForwardTime, tt.args.addressExclusions, tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}

	// test thread safety
	tt := tests[0]
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := w.ForwardLookupByTime(ToHex(tt.args.nodeID), tt.args.maxLookForwardTime, tt.args.addressExclusions, tt.args.excludeSpendingGaps)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, results)
			}
		}()
	}
	wg.Wait()
}

func TestWrapper_SpendingFingerprint(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	// transaction graph not loaded -> should produce error
	_, _, err := w.SpendingFingerprint("")
	require.Error(t, err)

	w.transactionGraph = newDestinationGraph(t)

	tests := []struct {
		uid             string
		wantFingerprint []string
		wantNumSessions int
		wantErr         bool
	}{
		{
			uid:             "",
			wantFingerprint: nil,
			wantNumSessions: 0,
			wantErr:         true,
		},
		{
			uid:             ToHex(1),
			wantFingerprint: []string{ToHex(2)},
			wantNumSessions: 2,
			wantErr:         false,
		},
		{
			uid:             ToHex(2),
			wantFingerprint: []string{ToHex(1)},
			wantNumSessions: 2,
			wantErr:         false,
		},
		{
			uid:             ToHex(3),
			wantFingerprint: []string{ToHex(4)},
			wantNumSessions: 2,
			wantErr:         false,
		},
		{
			uid:             ToHex(4),
			wantFingerprint: []string{ToHex(3)},
			wantNumSessions: 2,
			wantErr:         false,
		},
		{
			uid:             ToHex(5),
			wantFingerprint: []string{},
			wantNumSessions: 2,
			wantErr:         false,
		},
		// transaction does not exist in graph
		{
			uid:             ToHex(100),
			wantFingerprint: []string{},
			wantNumSessions: 2,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		fingerprints, i, err := w.SpendingFingerprint(tt.uid)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.wantNumSessions, i)
			fingerprintUIDs := make([]string, len(fingerprints))
			for j, f := range fingerprints {
				fingerprintUIDs[j] = f.TransactionUID
			}

			sort.Strings(fingerprintUIDs)

			require.Equal(t, tt.wantFingerprint, fingerprintUIDs)
		}
	}

	// test thread safety
	tt := tests[0]
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fingerprints, i, err := w.SpendingFingerprint(tt.uid)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantNumSessions, i)
				fingerprintUIDs := make([]string, len(fingerprints))
				for j, f := range fingerprints {
					fingerprintUIDs[j] = f.TransactionUID
				}

				sort.Strings(fingerprintUIDs)

				require.Equal(t, tt.wantFingerprint, fingerprintUIDs)
			}
		}()
	}
	wg.Wait()
}

func TestWrapper_GetInputTransactions(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	// transaction graph not loaded -> should produce error
	_, err := w.GetInputTransactions("")
	require.Error(t, err)

	w.transactionGraph = newTestGraph()

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
		results, err := w.GetInputTransactions(tt.uid)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			// sort results so order stays the same
			slices.Sort(results)
			require.Equal(t, tt.want, results)
		}
	}

	// test thread safety
	tt := tests[0]
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := w.GetInputTransactions(tt.uid)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// sort results so order stays the same
				slices.Sort(results)
				require.Equal(t, tt.want, results)
			}
		}()
	}
	wg.Wait()
}

func TestWrapper_LoadGraphs(t *testing.T) {
	testhelper.SkipIfNoDB(t)

	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	// database is not set
	require.Error(t, w.LoadGraphs())

	db.SetupDBWithoutData(t, dbHandle)

	w.db = dbHandle

	// database is empty, therefore classifier status is not set. Should return no error and set isLoading to true.
	require.NoError(t, w.LoadGraphs())
	require.True(t, w.isLoading)

	db.SetupDB(t, dbHandle, testhelper.UsePrivacyFile)
	// set correct classifier status
	require.NoError(t, status.SetLastClassifiedBlockID(dbHandle, uint64(testhelper.ClassifierFileLastBlock)))

	// set wrapper not loading and set environment variable to
	// only load a small graph (should have no effect, as graph is small anyway)
	w.isLoading = false
	require.NoError(t, os.Setenv("DEV_GRAPH_LIMIT", "10000"))
	require.NoError(t, w.LoadGraphs())
	require.NotNil(t, w.transactionGraph)
}

func TestWrapper_Logger(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)
	// logger is set in TestMain
	require.NotNil(t, w.Logger())
}

func TestWrapper_Context(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)
	require.Equal(t, context.Background(), w.Context())
}

func TestWrapper_Name(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)
	require.NotEmpty(t, w.Name())
}

func TestWrapper_CalculateInitialState(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	// error because no graphs were loaded so far
	require.Error(t, w.CalculateInitialState())

	// simulate a loaded graph -> no error
	w.isLoading = true
	require.NoError(t, w.CalculateInitialState())
}

func TestWrapper_NextBlock(t *testing.T) {
	testhelper.SkipIfNoDB(t)

	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	// db handle not set -> error
	flag, err := w.NextBlock()
	require.Error(t, err)
	require.False(t, flag)

	w.db = dbHandle

	db.SetupDB(t, dbHandle, testhelper.UsePrivacyFile)
	require.NoError(t, status.SetLastClassifiedBlockID(dbHandle, uint64(testhelper.ClassifierFileLastBlock)))
	require.NoError(t, w.LoadGraphs())

	// false because w.state.top is higher than most recent classified block
	flag, err = w.NextBlock()
	require.NoError(t, err)
	require.False(t, flag)
}

func TestWrapper_CurrentBlock(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)
	require.Zero(t, w.CurrentBlock())
}

func TestWrapper_PostExecution(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)
	require.NoError(t, w.PostExecution())
}

func TestWrapper_IncrementState(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	require.Zero(t, w.state.ID)
	require.NoError(t, w.IncrementState())
	require.EqualValues(t, 1, w.state.ID)
}

func TestWrapper_Empty(t *testing.T) {
	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)
	require.False(t, w.Empty())

	w.state.ID = 30
	w.state.Top = 29

	require.True(t, w.Empty())
}

func TestWrapper_Iterate(t *testing.T) {
	testhelper.SkipIfNoDB(t)

	w := NewWrapper(context.Background(), nil)
	unregisterCollectors(w)

	w.db = dbHandle

	db.SetupDB(t, dbHandle, testhelper.UsePrivacyFile)
	require.NoError(t, status.SetLastClassifiedBlockID(dbHandle, uint64(testhelper.ClassifierFileLastBlock)))
	require.NoError(t, w.LoadGraphs())

	// state.ID is set to a block which does not exist,
	// therefore iterate detects no nodes to insert and moves on
	flag, err := w.Iterate()
	require.NoError(t, err)
	require.True(t, flag)

	// set state.ID to most recent block
	w.state.ID--
	// now there should be something to do
	flag, err = w.Iterate()
	require.NoError(t, err)
	require.True(t, flag)
}
