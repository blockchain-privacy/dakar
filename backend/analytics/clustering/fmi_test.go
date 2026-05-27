// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package clustering

import (
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics/clustering"
	dbstat "gitlab.com/blockchain-privacy/dakar/db/status"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func Test_addClustersToMergeList(t *testing.T) {
	clusterMergeMap := make(map[string]*newCluster)
	addressMergeMap := make(map[string]*newCluster)
	clusterStore := make(map[string]clustering.Cluster)
	newClusters := make(map[string]bool)
	newAddresses := make(map[string]bool)

	// case: no data given
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx1", newClusters, newAddresses)
	require.Empty(t, clusterMergeMap)
	require.Empty(t, addressMergeMap)
	require.Equal(t, 0, countPointer(clusterMergeMap))
	require.Equal(t, 0, countPointer(addressMergeMap))

	newAddresses["A1"] = true
	newAddresses["A2"] = true
	newAddresses["A3"] = true
	newAddresses["A4"] = true
	newAddresses["A5"] = true

	// case: 5 new addresses
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx2", newClusters, newAddresses)
	require.Empty(t, clusterMergeMap)
	require.Len(t, addressMergeMap, 5)
	require.Equal(t, 0, countPointer(clusterMergeMap))
	require.Equal(t, 1, countPointer(addressMergeMap))

	one := 1

	clusterStore["C1"] = clustering.Cluster{
		UID:          "C1",
		Type:         "fmi",
		AddressCount: &one,
	}

	clusterStore["C2"] = clustering.Cluster{
		UID:          "C2",
		Type:         "fmi",
		AddressCount: &one,
	}

	clusterStore["C3"] = clustering.Cluster{
		UID:          "C3",
		Type:         "fmi",
		AddressCount: &one,
	}

	clusterStore["C4"] = clustering.Cluster{
		UID:          "C4",
		Type:         "fmi",
		AddressCount: &one,
	}

	newAddresses2 := make(map[string]bool)

	newClusters["C1"] = true
	newClusters["C2"] = true
	newClusters["C3"] = true
	newAddresses2["A6"] = true
	// case: 3 new clusters, 1 new address
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx3", newClusters, newAddresses2)
	require.Len(t, clusterMergeMap, 3)
	require.Len(t, addressMergeMap, 6)
	require.Equal(t, 1, countPointer(clusterMergeMap))
	require.Equal(t, 2, countPointer(addressMergeMap))

	newAddresses3 := make(map[string]bool)

	newAddresses3["A1"] = true
	newAddresses3["A6"] = true
	newAddresses3["A7"] = true
	// case: link the two clusters
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx4", newClusters, newAddresses3)
	require.Len(t, clusterMergeMap, 3)
	require.Len(t, addressMergeMap, 7)
	require.Equal(t, 1, countPointer(clusterMergeMap))
	require.Equal(t, 1, countPointer(addressMergeMap))

	newAddresses4 := make(map[string]bool)
	newAddresses4["A8"] = true

	// case: single address cluster
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx5", nil, newAddresses4)
	require.Len(t, clusterMergeMap, 3)
	require.Len(t, addressMergeMap, 8)
	require.Equal(t, 1, countPointer(clusterMergeMap))
	require.Equal(t, 2, countPointer(addressMergeMap))
	require.Equal(t, "tx5", addressMergeMap["A8"].changeTransaction)

	// case: update single address with new transaction
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx6", nil, newAddresses4)
	require.Len(t, clusterMergeMap, 3)
	require.Len(t, addressMergeMap, 8)
	require.Equal(t, 1, countPointer(clusterMergeMap))
	require.Equal(t, 2, countPointer(addressMergeMap))
	require.Equal(t, "tx6", addressMergeMap["A8"].changeTransaction)

	// case: update single address cluster with new transaction
	singeCluster := map[string]bool{"C4": true}
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx7", singeCluster, nil)
	require.Len(t, clusterMergeMap, 4)
	require.Len(t, addressMergeMap, 8)
	require.Equal(t, 2, countPointer(clusterMergeMap))
	require.Equal(t, 2, countPointer(addressMergeMap))
	require.Equal(t, "tx7", clusterMergeMap["C4"].changeTransaction)

	// case: update single address cluster with new transaction again
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx8", singeCluster, nil)
	require.Len(t, clusterMergeMap, 4)
	require.Len(t, addressMergeMap, 8)
	require.Equal(t, 2, countPointer(clusterMergeMap))
	require.Equal(t, 2, countPointer(addressMergeMap))
	require.Equal(t, "tx8", clusterMergeMap["C4"].changeTransaction)
}

func countPointer(data map[string]*newCluster) int {
	found := make(map[*newCluster]bool)
	for _, v := range data {
		found[v] = true
	}
	return len(found)
}

func Test_buildDBOperation(t *testing.T) {
	five := 5
	ten := 10

	someCluster := newCluster{
		changeTransaction: "0x1",
		mergeList:         nil,
		addresses:         map[string]bool{"0x10": true, "0x20": true},
	}

	otherCluster := newCluster{}

	type args struct {
		processedClusters map[*newCluster]bool
		items             map[string]*newCluster
		clusterIndex      int
	}
	tests := []struct {
		args              args
		wantNumOperations int
		wantErr           bool
	}{
		{
			args:              args{processedClusters: nil, items: map[string]*newCluster{}, clusterIndex: 0},
			wantNumOperations: 0,
			wantErr:           false,
		},
		// mergeList and addresses empty
		{
			args: args{processedClusters: map[*newCluster]bool{}, items: map[string]*newCluster{"": {
				changeTransaction: "",
				mergeList:         nil,
				addresses:         nil,
			}}, clusterIndex: 0},
			wantNumOperations: 0,
			wantErr:           true,
		},
		{
			args: args{
				processedClusters: map[*newCluster]bool{},
				items: map[string]*newCluster{
					"a": {
						changeTransaction: "",
						mergeList:         nil,
						addresses:         nil,
					}},
				clusterIndex: 0},
			wantNumOperations: 0,
			wantErr:           true,
		},
		{
			args: args{
				processedClusters: map[*newCluster]bool{},
				items: map[string]*newCluster{
					"a": &someCluster},
				clusterIndex: 0},
			wantNumOperations: 1,
			wantErr:           false,
		},
		{
			args: args{
				processedClusters: map[*newCluster]bool{&someCluster: true},
				items: map[string]*newCluster{
					"a": &someCluster,
					"b": {
						changeTransaction: "0x2",
						mergeList:         []clustering.Cluster{{UID: "0x100", AddressCount: &five}, {UID: "0x200", AddressCount: &ten}},
						addresses:         map[string]bool{"0x30": true, "0x40": true},
					}},
				clusterIndex: 0},
			wantNumOperations: 1,
			wantErr:           false,
		},
		{
			args: args{
				processedClusters: map[*newCluster]bool{&otherCluster: true},
				items: map[string]*newCluster{
					"a": &someCluster,
					"b": {
						changeTransaction: "0x2",
						mergeList:         []clustering.Cluster{{UID: "0x100", AddressCount: &five}, {UID: "0x200", AddressCount: &ten}},
						addresses:         map[string]bool{"0x30": true, "0x40": true},
					}},
				clusterIndex: 0},
			wantNumOperations: 2,
			wantErr:           false,
		},
		{
			args: args{
				processedClusters: map[*newCluster]bool{&otherCluster: true},
				items: map[string]*newCluster{
					"a": &someCluster,
					"b": {
						changeTransaction: "0x2",
						mergeList:         []clustering.Cluster{{UID: "0x100", AddressCount: nil}},
					}},
				clusterIndex: 0},
			wantNumOperations: 0,
			wantErr:           true,
		},
		{
			args: args{
				processedClusters: map[*newCluster]bool{&otherCluster: true},
				items: map[string]*newCluster{
					"a": &someCluster,
					"b": {
						changeTransaction: "0x2",
						mergeList:         []clustering.Cluster{{UID: "0x100", AddressCount: new(50000)}, {UID: "0x200", AddressCount: &ten}},
						addresses:         map[string]bool{"0x30": true, "0x40": true},
					}},
				clusterIndex: 0},
			wantNumOperations: 2,
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		operation, err := buildDBOperation(tt.args.processedClusters, tt.args.items, tt.args.clusterIndex)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Len(t, operation, tt.wantNumOperations)
		}
	}
}

func Test_buildDBOperationLimitClusterSize(t *testing.T) {
	operation, err := buildDBOperation(map[*newCluster]bool{}, map[string]*newCluster{
		"a": {
			changeTransaction: "0x2",
			mergeList: []clustering.Cluster{
				{UID: "0x100", AddressCount: new(30000)},
				{UID: "0x200", AddressCount: new(8000)},
				{UID: "0x300", AddressCount: new(5000)},
				{UID: "0x400", AddressCount: new(4000)},
				{UID: "0x300", AddressCount: new(9000)},
				{UID: "0x500", AddressCount: new(1000)},
				{UID: "0x600", AddressCount: new(1000)},
				{UID: "0x700", AddressCount: new(6000)},
			},

			addresses: map[string]bool{"0x30": true, "0x40": true},
		}}, 0)
	require.NoError(t, err)
	// 34000 (cluster) + 2 (addresses)
	require.Equal(t, 34002, *operation[0].NewCluster.AddressCount)
	require.Equal(t, "0x300", operation[0].NewCluster.UID)
	require.Len(t, operation[0].OldClusters, 6)

	operation, err = buildDBOperation(map[*newCluster]bool{}, map[string]*newCluster{
		"a": {
			changeTransaction: "0x2",
			mergeList: []clustering.Cluster{
				{UID: "0x100", AddressCount: new(8000)},
				{UID: "0x200", AddressCount: new(5000)},
				{UID: "0x300", AddressCount: new(4000)},
				{UID: "0x400", AddressCount: new(9000)},
				{UID: "0x500", AddressCount: new(1000)},
				{UID: "0x600", AddressCount: new(1000)},
				{UID: "0x700", AddressCount: new(6000)},
			},

			addresses: map[string]bool{"0x30": true, "0x40": true},
		}}, 0)
	require.NoError(t, err)
	// 34000 (cluster) + 2 (addresses)
	require.Equal(t, 34002, *operation[0].NewCluster.AddressCount)
	require.Equal(t, "0x400", operation[0].NewCluster.UID)
	require.Len(t, operation[0].OldClusters, 6)

	operation, err = buildDBOperation(map[*newCluster]bool{}, map[string]*newCluster{
		"a": {
			changeTransaction: "0x2",
			mergeList: []clustering.Cluster{
				{UID: "0x300", AddressCount: new(5000)},
				{UID: "0x400", AddressCount: new(4000)},
				{UID: "0x500", AddressCount: new(9000)},
				{UID: "0x600", AddressCount: new(1000)},
				{UID: "0x700", AddressCount: new(1000)},
				{UID: "0x800", AddressCount: new(6000)},
			},

			addresses: map[string]bool{"0x30": true, "0x40": true},
		}}, 0)
	require.NoError(t, err)
	// 26000 (cluster) + 2 (addresses)
	require.Equal(t, 26002, *operation[0].NewCluster.AddressCount)
	require.Equal(t, "0x500", operation[0].NewCluster.UID)
	require.Len(t, operation[0].OldClusters, 5)

	operation, err = buildDBOperation(map[*newCluster]bool{}, map[string]*newCluster{
		"a": {
			changeTransaction: "0x2",
			mergeList: []clustering.Cluster{
				{UID: "0x300", AddressCount: new(60000)},
			},

			addresses: map[string]bool{"0x30": true, "0x40": true},
		}}, 0)
	require.NoError(t, err)
	// 0 (cluster) + 2 (addresses)
	require.Equal(t, 2, *operation[0].NewCluster.AddressCount)
	require.Equal(t, "_:c1", operation[0].NewCluster.UID)
	require.Empty(t, operation[0].OldClusters)

	operation, err = buildDBOperation(map[*newCluster]bool{}, map[string]*newCluster{
		"a": {
			changeTransaction: "0x2",
			addresses:         map[string]bool{"0x30": true, "0x40": true},
		}}, 0)
	require.NoError(t, err)
	// 0 (cluster) + 2 (addresses)
	require.Equal(t, 2, *operation[0].NewCluster.AddressCount)
	require.Equal(t, "_:c1", operation[0].NewCluster.UID)
	require.Empty(t, operation[0].OldClusters)
}

func Test_calculateMetrics(t *testing.T) {
	operations := []clustering.DBOperation{
		{
			NewCluster: clustering.Cluster{
				Addresses: []db.UIDNode{{UID: "0x1"}, {UID: "0x2"}, {UID: "0x3"}},
			},
			OldClusters: []string{"1", "2", "3"},
		},
		{
			NewCluster: clustering.Cluster{
				Addresses: []db.UIDNode{{UID: "0x1"}, {UID: "0x2"}},
			},
			OldClusters: []string{"1", "2"},
		},
		{
			NewCluster: clustering.Cluster{
				Addresses: []db.UIDNode{{UID: "0x1"}},
			},
			OldClusters: []string{"1"},
		},
	}

	clusterCount, addressCount := calculateMetrics(operations)
	require.Equal(t, 6, clusterCount)
	require.Equal(t, 6, addressCount)
}

func TestNewFlatMultiInput(t *testing.T) {
	fm := NewFlatMultiInput(t.Context(), nil, NewDashConfig())

	require.NotNil(t, fm)
}

func TestFlatMultiInput_CalculateInitialState(t *testing.T) {
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	fm := NewFlatMultiInput(ctx, nil, NewDashConfig())
	fm.RegisterMetrics(prometheus.NewRegistry())
	// panics because db is not set
	require.Panics(t, func() {
		_ = fm.CalculateInitialState(ctx)
	})

	dbHandle := db.GetDBConnection(t, "")
	fm.db = dbHandle

	// error because classifier status is not set
	require.Error(t, fm.CalculateInitialState(ctx))

	// set classifier status
	require.NoError(t, dbstat.SetClassifying(ctx, dbHandle, true))

	require.NoError(t, fm.CalculateInitialState(ctx))
	require.EqualValues(t, 1, fm.state.ID)
	require.EqualValues(t, 0, fm.state.Top)
}

func TestFlatMultiInput_Iterate(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	fm := NewFlatMultiInput(ctx, dbHandle, NewDashConfig())
	fm.RegisterMetrics(prometheus.NewRegistry())
	require.NoError(t, dbstat.SetClassifying(ctx, dbHandle, true))
	require.NoError(t, fm.CalculateInitialState(ctx))

	// error because queue is empty
	ok, err := fm.Iterate(ctx)
	require.Error(t, err)
	require.False(t, ok)

	require.NoError(t, dbstat.SetLastClassifiedBlockID(ctx, dbHandle, db.BlockFileLastBlock))
	require.NoError(t, fm.CalculateInitialState(ctx))

	// this block contains transactions with multiple input addresses
	fm.state.ID = 60015

	ok, err = fm.Iterate(ctx)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestFlatMultiInput_NextBlock(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	fm := NewFlatMultiInput(ctx, dbHandle, NewDashConfig())

	// error because no status is set
	_, err := fm.Next(ctx)
	require.Error(t, err)

	require.NoError(t, dbstat.SetClassifying(ctx, dbHandle, true))

	// error because not classified block is set
	_, err = fm.Next(ctx)
	require.Error(t, err)

	require.NoError(t, dbstat.SetLastClassifiedBlockID(ctx, dbHandle, db.BlockFileLastBlock))
	ok, err := fm.Next(ctx)
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, db.BlockFileLastBlock, fm.state.Top)
}

func TestFlatMultiInput_PostExecution(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	fm := NewFlatMultiInput(ctx, dbHandle, NewDashConfig())

	require.NoError(t, fm.PostExecution(ctx))
}

func TestFlatMultiInput_IncrementState(t *testing.T) {
	fm := NewFlatMultiInput(t.Context(), nil, NewDashConfig())

	require.EqualValues(t, 0, fm.state.ID)
	// need to simulate a block being processed
	fm.blocksProcessed = 10
	require.NoError(t, fm.IncrementState())
	require.EqualValues(t, 10, fm.state.ID)
}

func TestFlatMultiInput_Empty(t *testing.T) {
	fm := NewFlatMultiInput(t.Context(), nil, NewDashConfig())

	// initially top and id are 0, so not empty
	require.False(t, fm.Empty())
	fm.state.Top = 5
	fm.state.ID = 6

	require.True(t, fm.Empty())
}

func TestFlatMultiInput_Props(t *testing.T) {
	fm := NewFlatMultiInput(t.Context(), nil, NewDashConfig())

	require.NotEmpty(t, fm.Props())
}

func Test_setInitialFMIClusteringID(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.Error(t, setInitialFMIClusteringID(ctx, dbHandle))
	require.NoError(t, dbstat.SetClusteringFMI(ctx, dbHandle, true))
	require.NoError(t, setInitialFMIClusteringID(ctx, dbHandle))
}

func Test_isGenericCoinJoin(t *testing.T) {
	tests := []struct {
		t    clustering.TransactionWithInputOutputAddressCluster
		c    Config
		want bool
	}{
		{
			c: Config{excludeInputCountThreshold: 5, excludeOutputCountThreshold: 5},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 5),
				OutputAddresses: make([]clustering.AddressWithCluster, 5),
			},
			want: true,
		},
		{
			c: Config{excludeInputCountThreshold: 5, excludeOutputCountThreshold: 5},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 4),
				OutputAddresses: make([]clustering.AddressWithCluster, 5),
			},
			want: false,
		},
		{
			c: Config{excludeInputCountThreshold: 5, excludeOutputCountThreshold: 5},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 5),
				OutputAddresses: make([]clustering.AddressWithCluster, 4),
			},
			want: false,
		},
		{
			c: Config{excludeInputCountThreshold: 5, excludeOutputCountThreshold: 5},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 100),
				OutputAddresses: make([]clustering.AddressWithCluster, 100),
			},
			want: true,
		},
		{
			c: Config{excludeInputCountThreshold: 0, excludeOutputCountThreshold: 5},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 1),
				OutputAddresses: make([]clustering.AddressWithCluster, 6),
			},
			want: true,
		},
		{
			c: Config{excludeInputCountThreshold: 5, excludeOutputCountThreshold: 0},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 0),
				OutputAddresses: make([]clustering.AddressWithCluster, 100),
			},
			want: false,
		},
		{
			c: Config{excludeInputCountThreshold: 50, excludeOutputCountThreshold: 0},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 50),
				OutputAddresses: make([]clustering.AddressWithCluster, 44),
			},
			want: true,
		},
		{
			c: Config{excludeInputCountThreshold: 0, excludeOutputCountThreshold: 0},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 50),
				OutputAddresses: make([]clustering.AddressWithCluster, 44),
			},
			want: false,
		},
		{
			c: Config{excludeInputCountThreshold: 0, excludeOutputCountThreshold: 0},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 1),
				OutputAddresses: make([]clustering.AddressWithCluster, 2),
			},
			want: false,
		},
		{
			c: Config{excludeInputCountThreshold: 0, excludeOutputCountThreshold: 0},
			t: clustering.TransactionWithInputOutputAddressCluster{
				InputAddresses:  make([]clustering.AddressWithCluster, 0),
				OutputAddresses: make([]clustering.AddressWithCluster, 0),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, isGenericCoinJoin(tt.t, tt.c))
	}
}
