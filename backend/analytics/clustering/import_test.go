// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package clustering

import (
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics/clustering"
)

func Test_buildClusterSet(t *testing.T) {
	tests := []struct {
		clusters []ExternalClusterItem
		want     map[string]map[string]bool
	}{
		{
			clusters: []ExternalClusterItem{},
			want:     map[string]map[string]bool{},
		},
		{
			clusters: []ExternalClusterItem{
				{ClusterID: "1", AddressHash: "a"},
				{ClusterID: "1", AddressHash: "b"},
			},
			want: map[string]map[string]bool{"1": {"a": true, "b": true}},
		},
		{
			clusters: []ExternalClusterItem{
				{ClusterID: "1", AddressHash: "a"},
				{ClusterID: "2", AddressHash: "b"},
				{ClusterID: "1", AddressHash: "c"},
				{ClusterID: "2", AddressHash: "d"},
				{ClusterID: "2", AddressHash: "e"},
			},
			want: map[string]map[string]bool{
				"1": {"a": true, "c": true},
				"2": {"b": true, "d": true, "e": true},
			},
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, buildClusterSet(tt.clusters))
	}
}

func Test_buildDatabaseClusters(t *testing.T) {
	type args struct {
		clusters  []ExternalClusterItem
		hashToUID map[string]string
	}
	tests := []struct {
		args args
		want []clustering.CustomCluster
	}{
		{
			args: args{clusters: nil, hashToUID: nil},
			want: []clustering.CustomCluster{},
		},
		{
			args: args{
				clusters: []ExternalClusterItem{
					{ClusterID: "1", AddressHash: "a"},
					{ClusterID: "2", AddressHash: "b"},
					{ClusterID: "1", AddressHash: "c"},
					{ClusterID: "2", AddressHash: "d"},
					{ClusterID: "2", AddressHash: "e"},
				},
				hashToUID: map[string]string{"a": "0x1", "b": "0x2", "c": "0x3", "d": "0x4", "e": "0x5"},
			},
			want: []clustering.CustomCluster{
				{Addresses: []db.UIDNode{{UID: "0x1"}, {UID: "0x3"}}},
				{Addresses: []db.UIDNode{{UID: "0x2"}, {UID: "0x4"}, {UID: "0x5"}}},
			},
		},
	}
	for _, tt := range tests {
		clusters := buildDatabaseClusters(tt.args.clusters, "0x1234", tt.args.hashToUID)

		sort.Slice(clusters, func(i, j int) bool {
			return *clusters[i].AddressCount < *clusters[j].AddressCount
		})

		for i, c := range clusters {
			// sort slice so order is the same as in the expected result
			sort.Slice(c.Addresses, func(i, j int) bool {
				return c.Addresses[i].UID < c.Addresses[j].UID
			})
			require.Equal(t, tt.want[i].Addresses, c.Addresses)
		}
	}
}

func Test_validateAddresses(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	// generate 1001 address hashes
	oneThousandAndOneAddresses := make([]ExternalClusterItem, 1001)
	for i := range 1001 {
		oneThousandAndOneAddresses[i] = ExternalClusterItem{ClusterID: "a", AddressHash: strconv.Itoa(i)}
	}

	tests := []struct {
		clusters    []ExternalClusterItem
		wantMapLen  int
		wantErr     bool
		wantErrType error
	}{
		{clusters: nil, wantErr: true},
		{
			clusters: []ExternalClusterItem{
				{ClusterID: "1", AddressHash: "XccPicDbg7HhRfFwPU3Z2CFVpHSYAhuxtu"},
				{ClusterID: "2", AddressHash: "XggCE7j3QozdXavADMLVEmREPvKXKQexqP"},
				{ClusterID: "1", AddressHash: "Xu4ebhFCAp8vq8oJkyhfB7uQFXqWkTuqur"},
				{ClusterID: "2", AddressHash: "XhqJaPve64tjUjootwQDjghyJ7RkBxVBYy"},
				{ClusterID: "2", AddressHash: "XdJPEBAEckf5JVkcN24Eh84S39Az4hwErN"},
			},
			wantMapLen: 5,
			wantErr:    false,
		},
		{
			clusters: []ExternalClusterItem{
				{ClusterID: "1", AddressHash: "XccPicDbg7HhRfFwPU3Z2CFVpHSYAhuxtu"},
				{ClusterID: "2", AddressHash: "XggCE7j3QozdXavADMLVEmREPvKXKQexqP"},
			},
			wantErr:     true,
			wantErrType: ErrShallowCluster,
		},
		{
			clusters:    oneThousandAndOneAddresses,
			wantErr:     true,
			wantErrType: ErrTooManyAddresses,
		},
	}
	for _, tt := range tests {
		addresses, err := validateAddresses(t.Context(), dbHandle, tt.clusters)
		if tt.wantErr {
			require.Error(t, err)
			if tt.wantErrType != nil {
				require.ErrorIs(t, err, tt.wantErrType)
			}
		} else {
			require.NoError(t, err)
			require.Len(t, addresses, tt.wantMapLen)
		}
	}
}

func TestImportCluster(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	type args struct {
		clusters []ExternalClusterItem
		userID   string
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args:    args{userID: ""},
			wantErr: true,
		},
		{
			args: args{
				clusters: []ExternalClusterItem{
					{ClusterID: "1", AddressHash: "XccPicDbg7HhRfFwPU3Z2CFVpHSYAhuxtu"},
					{ClusterID: "2", AddressHash: "XggCE7j3QozdXavADMLVEmREPvKXKQexqP"},
					{ClusterID: "1", AddressHash: "Xu4ebhFCAp8vq8oJkyhfB7uQFXqWkTuqur"},
					{ClusterID: "2", AddressHash: "XhqJaPve64tjUjootwQDjghyJ7RkBxVBYy"},
					{ClusterID: "2", AddressHash: "XdJPEBAEckf5JVkcN24Eh84S39Az4hwErN"},
				},
				userID: "0x1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := ImportCluster(t.Context(), dbHandle, tt.args.clusters, tt.args.userID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
