package clustering

import (
	"backend/db/analytics/clustering"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAddClustersToMergeList(t *testing.T) {
	clusterMergeMap := make(map[string]*newCluster)
	addressMergeMap := make(map[string]*newCluster)
	clusterStore := make(map[string]clustering.Cluster)
	newClusters := make(map[string]bool)
	newAddresses := make(map[string]bool)

	// case: no data given
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx1", newClusters, newAddresses)
	require.Len(t, clusterMergeMap, 0)
	require.Len(t, addressMergeMap, 0)
	require.Equal(t, 0, countPointer(clusterMergeMap))
	require.Equal(t, 0, countPointer(addressMergeMap))

	newAddresses["A1"] = true
	newAddresses["A2"] = true
	newAddresses["A3"] = true
	newAddresses["A4"] = true
	newAddresses["A5"] = true

	// case: 5 new addresses
	addClustersToMergeList(clusterMergeMap, addressMergeMap, clusterStore, "tx2", newClusters, newAddresses)
	require.Len(t, clusterMergeMap, 0)
	require.Len(t, addressMergeMap, 5)
	require.Equal(t, 0, countPointer(clusterMergeMap))
	require.Equal(t, 1, countPointer(addressMergeMap))

	one := 1

	clusterStore["C1"] = clustering.Cluster{
		Uid:          "C1",
		Type:         "fmi",
		AddressCount: &one,
	}

	clusterStore["C2"] = clustering.Cluster{
		Uid:          "C2",
		Type:         "fmi",
		AddressCount: &one,
	}

	clusterStore["C3"] = clustering.Cluster{
		Uid:          "C3",
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
}

func countPointer(data map[string]*newCluster) int {
	found := make(map[*newCluster]bool)
	for _, v := range data {
		found[v] = true
	}
	return len(found)
}
