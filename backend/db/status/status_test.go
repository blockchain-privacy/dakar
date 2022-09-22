package status

import (
	"backend/db"
	"backend/external"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle external.Database

func TestMain(m *testing.M) {
	db.RunDgraphTests(m, &dbHandle, "dgraph_status")
}

func TestGetCrawlerStatus(t *testing.T) {
	// crawler status not yet set
	_, err := GetCrawlerStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set crawling
	require.NoError(t, SetCrawling(dbHandle, true))

	status, err := GetCrawlerStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsCrawling)

	// set not crawling
	require.NoError(t, SetCrawling(dbHandle, false))

	status, err = GetCrawlerStatus(dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsCrawling)
}

func TestGetClassifierStatus(t *testing.T) {
	// classifier status not yet set
	_, err := GetClassifierStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set classifying
	require.NoError(t, SetClassifying(dbHandle, true))

	status, err := GetClassifierStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClassifying)

	// set not classifying
	require.NoError(t, SetClassifying(dbHandle, false))

	status, err = GetClassifierStatus(dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClassifying)
}

func TestGetClusteringHMIStatus(t *testing.T) {
	// clustering status not yet set
	_, err := GetClusteringHMIStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set clustering
	require.NoError(t, SetClusteringHMI(dbHandle, true))

	status, err := GetClusteringHMIStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClustering)

	// set not clustering
	require.NoError(t, SetClusteringHMI(dbHandle, false))

	status, err = GetClusteringHMIStatus(dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClustering)
}

func TestGetClusteringFMIStatus(t *testing.T) {
	// clustering status not yet set
	_, err := GetClusteringFMIStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set clustering
	require.NoError(t, SetClusteringFMI(dbHandle, true))

	status, err := GetClusteringFMIStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClustering)

	// set not clustering
	require.NoError(t, SetClusteringFMI(dbHandle, false))

	status, err = GetClusteringFMIStatus(dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClustering)
}

func TestGetHighestBlockID(t *testing.T) {
	// nothing set yet -> should fail
	_, err := GetHighestBlockID(dbHandle)
	require.Error(t, err)
}

func TestGetFrontendStatus(t *testing.T) {
	require.NoError(t, db.DropAll(dbHandle))

	// nothing set yet -> should fail
	_, err := GetFrontendStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set crawling
	require.NoError(t, SetCrawling(dbHandle, true))
	require.NoError(t, SetLastBlockID(dbHandle, 50))

	status, err := GetFrontendStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, status.IsCrawling)
	require.Equal(t, status.LastBlockID, uint64(50))
}

func TestGetMeta(t *testing.T) {
	// nothing set yet -> should fail
	_, err := GetMeta(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set schema version
	require.NoError(t, InitializeMeta(dbHandle, "Dash"))

	metaResult, err := GetMeta(dbHandle)
	require.NoError(t, err)
	require.NotNil(t, metaResult.SchemaVersion)
	require.Equal(t, *metaResult.SchemaVersion, db.SchemaVersion)
	require.Equal(t, metaResult.BlockchainMode, "Dash")
	require.NotEmpty(t, metaResult.CreationTime)
}
