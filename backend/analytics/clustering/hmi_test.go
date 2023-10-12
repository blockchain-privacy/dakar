package clustering

import (
	"backend/db"
	"backend/db/analytics/clustering"
	dbstat "backend/db/status"
	"backend/testhelper"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"testing"
)

// unregisterCollectors unregisters all collectors of the hierarchical multi input clusterer.
// This is needed because collectors can not be registered twice with the same default config.
func unregisterCollectorsHM(hm *HierarchicalMultiInput) {
	if hm == nil {
		return
	}

	prometheus.Unregister(hm.blocks)
	prometheus.Unregister(hm.transactions)
	prometheus.Unregister(hm.mergedClusters)
	prometheus.Unregister(hm.newAddresses)
	prometheus.Unregister(hm.blockHeight)
}

func TestNewHierarchicalMultiInput(t *testing.T) {
	hm := NewHierarchicalMultiInput(context.Background(), nil)
	unregisterCollectorsHM(hm)
	require.NotNil(t, hm)
}

func TestHierarchicalMultiInput_CalculateInitialState(t *testing.T) {
	testhelper.SkipIfNoDB(t)

	hm := NewHierarchicalMultiInput(context.Background(), nil)
	unregisterCollectorsHM(hm)

	// panics because db is not set
	require.Panics(t, func() {
		_ = hm.CalculateInitialState()
	})

	hm.db = dbHandle
	db.SetupDBWithoutData(t, dbHandle)

	// error because classifier status is not set
	require.Error(t, hm.CalculateInitialState())

	// set classifier status
	require.NoError(t, dbstat.SetClassifying(dbHandle, true))

	require.NoError(t, hm.CalculateInitialState())
	require.EqualValues(t, 1, hm.state.ID)
	require.EqualValues(t, 0, hm.state.Top)
}

func TestHierarchicalMultiInput_Iterate(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)
	hm := NewHierarchicalMultiInput(context.Background(), dbHandle)
	unregisterCollectorsHM(hm)

	require.NoError(t, dbstat.SetClassifying(dbHandle, true))
	require.NoError(t, hm.CalculateInitialState())

	// error because queue is empty
	ok, err := hm.Iterate()
	require.Error(t, err)
	require.False(t, ok)

	require.NoError(t, dbstat.SetLastClassifiedBlockID(dbHandle, testhelper.BlockFileLastBlock))
	require.NoError(t, hm.CalculateInitialState())

	ok, err = hm.Iterate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestHierarchicalMultiInput_NextBlock(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	hm := NewHierarchicalMultiInput(context.Background(), dbHandle)
	unregisterCollectorsHM(hm)

	// error because no status is set
	_, err := hm.NextBlock()
	require.Error(t, err)

	require.NoError(t, dbstat.SetClassifying(dbHandle, true))

	// error because not classified block is set
	_, err = hm.NextBlock()
	require.Error(t, err)

	require.NoError(t, dbstat.SetLastClassifiedBlockID(dbHandle, testhelper.BlockFileLastBlock))
	ok, err := hm.NextBlock()
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, testhelper.BlockFileLastBlock, hm.state.Top)
}

func TestHierarchicalMultiInput_PostExecution(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)

	hm := NewHierarchicalMultiInput(context.Background(), dbHandle)
	unregisterCollectorsHM(hm)

	require.NoError(t, hm.PostExecution())
}

func TestHierarchicalMultiInput_IncrementState(t *testing.T) {
	hm := NewHierarchicalMultiInput(context.Background(), dbHandle)
	unregisterCollectorsHM(hm)

	require.EqualValues(t, 0, hm.state.ID)
	require.NoError(t, hm.IncrementState())
	require.EqualValues(t, 1, hm.state.ID)
}

func TestHierarchicalMultiInput_Empty(t *testing.T) {
	hm := NewHierarchicalMultiInput(context.Background(), nil)
	unregisterCollectorsHM(hm)

	// initially top and id are 0, so not empty
	require.False(t, hm.Empty())
	hm.state.Top = 5
	hm.state.ID = 6

	require.True(t, hm.Empty())
}

func TestHierarchicalMultiInput_CurrentBlock(t *testing.T) {
	hm := NewHierarchicalMultiInput(context.Background(), nil)
	unregisterCollectorsHM(hm)

	require.Zero(t, hm.CurrentBlock())
}

func TestHierarchicalMultiInput_Logger(t *testing.T) {
	hm := NewHierarchicalMultiInput(context.Background(), nil)
	unregisterCollectorsHM(hm)

	require.NotNil(t, hm.Logger())
}

func TestHierarchicalMultiInput_Context(t *testing.T) {
	hm := NewHierarchicalMultiInput(context.Background(), nil)
	unregisterCollectorsHM(hm)

	require.NotNil(t, hm.Context())
}

func TestHierarchicalMultiInput_Name(t *testing.T) {
	hm := NewHierarchicalMultiInput(context.Background(), nil)
	unregisterCollectorsHM(hm)

	require.NotEmpty(t, hm.Name())
}

func Test_setInitialHMIClusteringID(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)

	require.Error(t, setInitialHMIClusteringID(dbHandle))
	require.NoError(t, dbstat.SetClusteringHMI(dbHandle, true))
	require.NoError(t, setInitialHMIClusteringID(dbHandle))
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
				{Children: []clustering.SubCluster{{UID: "0x1"}}, Addresses: nil},
				{Children: []clustering.SubCluster{{UID: "0x1"}}, Addresses: nil},
			},
			wantErr: true,
		},
		// error because address has multiple parents
		{
			clusters: []clustering.Cluster{
				{Addresses: []clustering.HollowAddress{{UID: "0x1"}}, Children: nil},
				{Addresses: []clustering.HollowAddress{{UID: "0x1"}}, Children: nil},
			},
			wantErr: true,
		},
		{
			clusters: []clustering.Cluster{
				{Addresses: []clustering.HollowAddress{{UID: "0x1"}}, Children: []clustering.SubCluster{{UID: "0x3"}}},
				{Addresses: []clustering.HollowAddress{{UID: "0x2"}}, Children: []clustering.SubCluster{{UID: "0x4"}}},
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
