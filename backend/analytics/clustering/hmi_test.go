// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package clustering

import (
	"backend/db"
	"backend/db/analytics/clustering"
	dbstat "backend/db/status"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewHierarchicalMultiInput(t *testing.T) {
	hm := NewHierarchicalMultiInput(t.Context(), nil)
	require.NotNil(t, hm)
}

func TestHierarchicalMultiInput_CalculateInitialState(t *testing.T) {
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	hm := NewHierarchicalMultiInput(ctx, nil)
	hm.RegisterMetrics(prometheus.NewRegistry())
	// panics because db is not set
	require.Panics(t, func() {
		_ = hm.CalculateInitialState(ctx)
	})

	dbHandle := db.GetDBConnection(t, "")
	hm.db = dbHandle

	// error because classifier status is not set
	require.Error(t, hm.CalculateInitialState(ctx))

	// set classifier status
	require.NoError(t, dbstat.SetClassifying(ctx, dbHandle, true))

	require.NoError(t, hm.CalculateInitialState(ctx))
	require.EqualValues(t, 1, hm.state.ID)
	require.EqualValues(t, 0, hm.state.Top)
}

func TestHierarchicalMultiInput_Iterate(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	hm := NewHierarchicalMultiInput(ctx, dbHandle)
	hm.RegisterMetrics(prometheus.NewRegistry())

	require.NoError(t, dbstat.SetClassifying(ctx, dbHandle, true))
	require.NoError(t, hm.CalculateInitialState(ctx))

	// error because queue is empty
	ok, err := hm.Iterate(ctx)
	require.Error(t, err)
	require.False(t, ok)

	require.NoError(t, dbstat.SetLastClassifiedBlockID(ctx, dbHandle, db.BlockFileLastBlock))
	require.NoError(t, hm.CalculateInitialState(ctx))

	// this block contains transactions with multiple input addresses
	hm.state.ID = 60015

	ok, err = hm.Iterate(ctx)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestHierarchicalMultiInput_NextBlock(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	hm := NewHierarchicalMultiInput(ctx, dbHandle)

	// error because no status is set
	_, err := hm.Next(ctx)
	require.Error(t, err)

	require.NoError(t, dbstat.SetClassifying(ctx, dbHandle, true))

	// error because not classified block is set
	_, err = hm.Next(ctx)
	require.Error(t, err)

	require.NoError(t, dbstat.SetLastClassifiedBlockID(ctx, dbHandle, db.BlockFileLastBlock))
	ok, err := hm.Next(ctx)
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, db.BlockFileLastBlock, hm.state.Top)
}

func TestHierarchicalMultiInput_PostExecution(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	hm := NewHierarchicalMultiInput(ctx, dbHandle)

	require.NoError(t, hm.PostExecution(ctx))
}

func TestHierarchicalMultiInput_IncrementState(t *testing.T) {
	hm := NewHierarchicalMultiInput(t.Context(), nil)

	require.EqualValues(t, 0, hm.state.ID)
	require.NoError(t, hm.IncrementState())
	require.EqualValues(t, 1, hm.state.ID)
}

func TestHierarchicalMultiInput_Empty(t *testing.T) {
	hm := NewHierarchicalMultiInput(t.Context(), nil)

	// initially top and id are 0, so not empty
	require.False(t, hm.Empty())
	hm.state.Top = 5
	hm.state.ID = 6

	require.True(t, hm.Empty())
}

func TestHierarchicalMultiInput_Props(t *testing.T) {
	hm := NewHierarchicalMultiInput(t.Context(), nil)

	require.NotEmpty(t, hm.Props())
}

func Test_setInitialHMIClusteringID(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	require.Error(t, setInitialHMIClusteringID(ctx, dbHandle))
	require.NoError(t, dbstat.SetClusteringHMI(ctx, dbHandle, true))
	require.NoError(t, setInitialHMIClusteringID(ctx, dbHandle))
}

func Test_getClusterRootByCluster(t *testing.T) {
	type args struct {
		clusterMapping map[string]string
		uid            string
	}
	tests := []struct {
		args args
		want string
	}{
		{
			args: args{
				clusterMapping: map[string]string{"a": "b", "b": "c", "1": "2", "c": "d"},
				uid:            "a",
			},
			want: "d",
		},
		{
			args: args{
				clusterMapping: map[string]string{"a": "b", "b": "c", "1": "2", "c": "d"},
				uid:            "b",
			},
			want: "d",
		},
		{
			args: args{
				clusterMapping: map[string]string{"a": "b", "b": "c", "1": "2", "c": "d"},
				uid:            "c",
			},
			want: "d",
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, getClusterRootByCluster(tt.args.clusterMapping, tt.args.uid))
	}
}

func Test_validateClusters(t *testing.T) {
	tests := []struct {
		clusters []clustering.Cluster
		wantErr  bool
	}{
		{
			clusters: []clustering.Cluster{},
			wantErr:  false,
		},
		// error because both children and addresses are empty
		{
			clusters: []clustering.Cluster{{
				Children:  nil,
				Addresses: nil,
			}},
			wantErr: true,
		},
		// error because cluster has multiple parents
		{
			clusters: []clustering.Cluster{
				{Children: []db.UIDNode{{UID: "0x1"}}, Addresses: nil},
				{Children: []db.UIDNode{{UID: "0x1"}}, Addresses: nil},
			},
			wantErr: true,
		},
		// error because address has multiple parents
		{
			clusters: []clustering.Cluster{
				{Addresses: []db.UIDNode{{UID: "0x1"}}, Children: nil},
				{Addresses: []db.UIDNode{{UID: "0x1"}}, Children: nil},
			},
			wantErr: true,
		},
		{
			clusters: []clustering.Cluster{
				{Addresses: []db.UIDNode{{UID: "0x1"}}, Children: []db.UIDNode{{UID: "0x3"}}},
				{Addresses: []db.UIDNode{{UID: "0x2"}}, Children: []db.UIDNode{{UID: "0x4"}}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := validateClusters(tt.clusters)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
