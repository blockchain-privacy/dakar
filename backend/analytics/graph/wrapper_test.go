// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"reflect"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/status"
)

func TestNewWrapper(t *testing.T) {
	w := NewWrapper(t.Context(), nil)
	require.NotNil(t, w)
	require.NotNil(t, w.transactionGraphMutex)
}

func TestWrapper_IsTransactionGraphLoaded(t *testing.T) {
	w := NewWrapper(t.Context(), nil)

	require.False(t, w.IsTransactionGraphLoaded())
	w.transactionGraph = NewReversibleGraph(1)
	require.True(t, w.IsTransactionGraphLoaded())

	// should be thread safe
	numGoroutines := 10
	responses := make(chan bool, numGoroutines)
	for range numGoroutines {
		go func() {
			responses <- w.IsTransactionGraphLoaded()
		}()
	}

	for range numGoroutines {
		r := <-responses
		require.True(t, r)
	}
}

func TestWrapper_ReverseLookup(t *testing.T) {
	w := NewWrapper(t.Context(), nil)

	// transaction graph not loaded -> should produce error
	_, err := w.ReverseLookup("", 0, 0, false)
	require.Error(t, err)

	w.transactionGraph = newTestGraph()

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
	}

	for _, tt := range tests {
		results, err := w.ReverseLookup(ToHex(tt.args.nodeID), tt.args.maxLookBackTime, tt.args.maxDepth,
			tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}

	// test thread safety
	tt := tests[0]
	numGoroutines := 10
	type response struct {
		err error
		res map[string]bool
	}
	responses := make(chan response, numGoroutines)
	for range numGoroutines {
		go func() {
			results, err := w.ReverseLookup(ToHex(tt.args.nodeID), tt.args.maxLookBackTime, tt.args.maxDepth,
				tt.args.excludeSpendingGaps)
			responses <- response{err: err, res: results}
		}()
	}

	for range numGoroutines {
		r := <-responses
		if tt.wantErr {
			require.Error(t, r.err)
		} else {
			require.NoError(t, r.err)
			require.Equal(t, tt.want, r.res)
		}
	}
}

func TestWrapper_ForwardLookup(t *testing.T) {
	w := NewWrapper(t.Context(), nil)

	// transaction graph not loaded -> should produce error
	_, err := w.ForwardLookup("", 0, 0, false)
	require.Error(t, err)

	w.transactionGraph = newTestGraph()

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
	}
	for _, tt := range tests {
		results, err := w.ForwardLookup(ToHex(tt.args.nodeID), tt.args.maxLookForwardTime, tt.args.maxDepth,
			tt.args.excludeSpendingGaps)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, results)
		}
	}

	// test thread safety
	tt := tests[0]
	numGoroutines := 10
	type response struct {
		err error
		res map[string]bool
	}
	responses := make(chan response, numGoroutines)
	for range numGoroutines {
		go func() {
			results, err := w.ForwardLookup(ToHex(tt.args.nodeID), tt.args.maxLookForwardTime, tt.args.maxDepth,
				tt.args.excludeSpendingGaps)
			responses <- response{err: err, res: results}
		}()
	}

	for range numGoroutines {
		r := <-responses
		if tt.wantErr {
			require.Error(t, r.err)
		} else {
			require.NoError(t, r.err)
			require.Equal(t, tt.want, r.res)
		}
	}
}

func TestWrapper_SpendingFingerprint(t *testing.T) {
	w := NewWrapper(t.Context(), nil)

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
	numGoroutines := 10
	type response struct {
		err          error
		fingerprints []FingerPrint
		numSessions  int
	}
	responses := make(chan response, numGoroutines)
	for range numGoroutines {
		go func() {
			fingerprints, sessionCount, err := w.SpendingFingerprint(tt.uid)
			responses <- response{err: err, fingerprints: fingerprints, numSessions: sessionCount}
		}()
	}

	for range numGoroutines {
		r := <-responses
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.wantNumSessions, r.numSessions)
			fingerprintUIDs := make([]string, len(r.fingerprints))
			for j, f := range r.fingerprints {
				fingerprintUIDs[j] = f.TransactionUID
			}

			sort.Strings(fingerprintUIDs)

			require.Equal(t, tt.wantFingerprint, fingerprintUIDs)
		}
	}
}
func TestWrapper_LoadGraphs(t *testing.T) {
	w := NewWrapper(t.Context(), nil)
	w.RegisterMetrics(prometheus.NewRegistry())

	// database is not set
	require.Error(t, w.LoadGraphs(NewDashConfig()))

	dbHandle := db.GetDBConnection(t, "")

	w.db = dbHandle

	// database is empty, therefore classifier status is not set. Should return no error and set isLoading to true.
	require.NoError(t, w.LoadGraphs(NewDashConfig()))
	require.True(t, w.isLoading)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	db.ChangeDBContent(dbHandle, db.UsePrivacyFile)

	// set correct classifier status
	require.NoError(t, status.SetLastClassifiedBlockID(ctx, dbHandle, int64(db.ClassifierFileLastBlock)))

	// set wrapper not loading and set environment variable to
	// only load a small graph (should have no effect, as graph is small anyway)
	w.isLoading = false
	t.Setenv("DEV_GRAPH_LIMIT", "10000")
	require.NoError(t, w.LoadGraphs(NewDashConfig()))
	require.NotNil(t, w.transactionGraph)
}

func TestWrapper_Props(t *testing.T) {
	w := NewWrapper(t.Context(), nil)
	require.NotEmpty(t, w.Props())
}

func TestWrapper_CalculateInitialState(t *testing.T) {
	w := NewWrapper(t.Context(), nil)
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	// error because no graphs were loaded so far
	require.Error(t, w.CalculateInitialState(ctx))

	// simulate a loaded graph -> no error
	w.isLoading = true
	require.NoError(t, w.CalculateInitialState(ctx))
}

func TestWrapper_NextBlock(t *testing.T) {
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	w := NewWrapper(ctx, nil)
	w.RegisterMetrics(prometheus.NewRegistry())

	// db handle not set -> error
	flag, err := w.Next(ctx)
	require.Error(t, err)
	require.False(t, flag)

	dbHandle := db.GetDBConnection(t, db.UsePrivacyFile)
	w.db = dbHandle

	require.NoError(t, status.SetLastClassifiedBlockID(ctx, dbHandle, int64(db.ClassifierFileLastBlock)))
	require.NoError(t, w.LoadGraphs(NewDashConfig()))

	// false because w.state.top is higher than most recent classified block
	flag, err = w.Next(ctx)
	require.NoError(t, err)
	require.False(t, flag)
}

func TestWrapper_PostExecution(t *testing.T) {
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	w := NewWrapper(ctx, nil)
	require.NoError(t, w.PostExecution(ctx))
}

func TestWrapper_IncrementState(t *testing.T) {
	w := NewWrapper(t.Context(), nil)

	require.Zero(t, w.state.ID)
	require.NoError(t, w.IncrementState())
	require.EqualValues(t, 1, w.state.ID)
}

func TestWrapper_Empty(t *testing.T) {
	w := NewWrapper(t.Context(), nil)
	require.False(t, w.Empty())

	w.state.ID = 30
	w.state.Top = 29

	require.True(t, w.Empty())
}

func TestWrapper_Iterate(t *testing.T) {
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	w := NewWrapper(ctx, nil)
	w.RegisterMetrics(prometheus.NewRegistry())

	dbHandle := db.GetDBConnection(t, db.UsePrivacyFile)
	w.db = dbHandle

	require.NoError(t, status.SetLastClassifiedBlockID(ctx, dbHandle, int64(db.ClassifierFileLastBlock)))
	require.NoError(t, w.LoadGraphs(NewDashConfig()))

	// state.ID is set to a block which does not exist,
	// therefore iterate detects no nodes to insert and moves on
	flag, err := w.Iterate(ctx)
	require.NoError(t, err)
	require.True(t, flag)

	// set state.ID to most recent block
	w.state.ID--
	// now there should be something to do
	flag, err = w.Iterate(ctx)
	require.NoError(t, err)
	require.True(t, flag)
}

func TestWrapper_PartitionNodesByDirectConnections(t *testing.T) {
	w := NewWrapper(t.Context(), nil)
	w.transactionGraph = newTestGraph()

	tests := []struct {
		nodes   []string
		want    [][]string
		wantErr bool
	}{
		{
			nodes:   []string{ToHex(7)},
			want:    [][]string{{ToHex(7)}},
			wantErr: false,
		},
		{
			nodes:   []string{ToHex(7), ToHex(6), ToHex(11)},
			want:    [][]string{{ToHex(7)}, {ToHex(6)}, {ToHex(11)}},
			wantErr: false,
		},
		{
			nodes:   []string{ToHex(7), ToHex(2), ToHex(10), ToHex(6), ToHex(11), ToHex(13)},
			want:    [][]string{{ToHex(10), ToHex(7), ToHex(13)}, {ToHex(6), ToHex(2)}, {ToHex(11)}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		results, err := w.PartitionNodesByDirectConnections(tt.nodes)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)

			for _, resultPart := range results {
				slices.Sort(resultPart)
				foundEqual := false
				for _, wantPart := range tt.want {
					slices.Sort(wantPart)
					if reflect.DeepEqual(resultPart, wantPart) {
						foundEqual = true
						break
					}
				}

				require.True(t, foundEqual)
			}
		}
	}
}
