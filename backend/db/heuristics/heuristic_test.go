// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateKeyHash(t *testing.T) {
	hash1 := createKeyHash(map[string]bool{"a1111": true, "a2": true, "a33": true})
	require.NotEmpty(t, hash1)

	hash2 := createKeyHash(map[string]bool{"a2": true, "a1111": true, "a33": true})
	require.NotEmpty(t, hash2)
	require.Equal(t, hash1, hash2, "same input elements should create same hash, no matter the order")

	hash3 := createKeyHash(map[string]bool{"a1111": true, "a2": true})
	require.NotEmpty(t, hash3)
	require.NotEqual(t, hash3, hash2, "fewer input elements should create a different hash")

	hash4 := createKeyHash(nil)
	require.Empty(t, hash4)

	hash5 := createKeyHash(map[string]bool{})
	require.Empty(t, hash5)
}

func TestGetClusterUIDFromMergedClusters(t *testing.T) {
	clusterItems := []mergedClusterItem{
		{
			clusterHash: "",
			clusterUIDs: map[string]bool{"a1": true, "a2": true},
		},
		{
			clusterHash: "",
			clusterUIDs: map[string]bool{"a3": true, "a4": true},
		},
		{
			clusterHash: "",
			clusterUIDs: map[string]bool{"a5": true, "a6": true, "a7": true},
		},
	}

	clusters1, m1, err := getClusterUIDFromMergedClusters(clusterItems, "a2")
	require.NoError(t, err)
	require.NotEmpty(t, m1)
	require.NotEmpty(t, clusters1, "should return non-empty cluster map")
	require.NotEmpty(t, clusterItems[0].clusterHash, "cluster hash should have been generated")

	clusters2, m2, err := getClusterUIDFromMergedClusters(clusterItems, "a8")
	require.Error(t, err)
	require.Empty(t, m2)
	require.Empty(t, clusters2, "should return empty cluster map")

	clusters3, m3, err := getClusterUIDFromMergedClusters(nil, "a6")
	require.Error(t, err)
	require.Empty(t, m3)
	require.Empty(t, clusters3, "should return empty cluster map")
}
