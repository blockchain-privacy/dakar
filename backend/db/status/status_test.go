package status

import (
	"backend/db"
	"backend/testhelper"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle = &testhelper.TestDB{}

func TestMain(m *testing.M) {
	testhelper.RunDgraphTests(m, dbHandle)
}

func TestGetCrawlerStatus(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	require.NoError(t, db.DropAll(dbHandle))

	ctx, cancel := db.GetTaskContext()
	defer cancel()
	// crawler status not yet set
	_, err := GetCrawlerStatus(ctx, dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set crawling
	require.NoError(t, SetCrawling(ctx, dbHandle, true))

	status, err := GetCrawlerStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsCrawling)

	// set not crawling
	require.NoError(t, SetCrawling(ctx, dbHandle, false))

	status, err = GetCrawlerStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsCrawling)
}

func TestGetClassifierStatus(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	require.NoError(t, db.DropAll(dbHandle))

	ctx, cancel := db.GetTaskContext()
	defer cancel()
	// classifier status not yet set
	_, err := GetClassifierStatus(ctx, dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set classifying
	require.NoError(t, SetClassifying(ctx, dbHandle, true))

	status, err := GetClassifierStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClassifying)

	// set not classifying
	require.NoError(t, SetClassifying(ctx, dbHandle, false))

	status, err = GetClassifierStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClassifying)
}

func TestGetClusteringHMIStatus(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	require.NoError(t, db.DropAll(dbHandle))
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	// clustering status not yet set
	_, err := GetClusteringHMIStatus(ctx, dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set clustering
	require.NoError(t, SetClusteringHMI(ctx, dbHandle, true))

	status, err := GetClusteringHMIStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClustering)

	// set not clustering
	require.NoError(t, SetClusteringHMI(ctx, dbHandle, false))

	status, err = GetClusteringHMIStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClustering)
}

func TestGetClusteringFMIStatus(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	require.NoError(t, db.DropAll(dbHandle))

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// clustering status not yet set
	_, err := GetClusteringFMIStatus(ctx, dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set clustering
	require.NoError(t, SetClusteringFMI(ctx, dbHandle, true))

	status, err := GetClusteringFMIStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClustering)

	// set not clustering
	require.NoError(t, SetClusteringFMI(ctx, dbHandle, false))

	status, err = GetClusteringFMIStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClustering)
}

func TestGetHighestBlockID(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	blockHeight, err := GetHighestBlockID(ctx, dbHandle)
	require.NoError(t, err)
	require.EqualValues(t, testhelper.BlockFileLastBlock, blockHeight)
}

func TestGetFrontendStatus(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	require.NoError(t, db.DropAll(dbHandle))

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// should not fail even if no status is set at all
	status, err := GetFrontendStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.Nil(t, status.IsCrawling)
	require.Nil(t, status.IsClassifying)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set crawling
	require.NoError(t, SetCrawling(ctx, dbHandle, true))
	require.NoError(t, SetLastBlockID(ctx, dbHandle, 50))

	status, err = GetFrontendStatus(context.Background(), dbHandle)
	require.NoError(t, err)
	require.NotNil(t, status.IsCrawling)
	require.True(t, *status.IsCrawling)
	require.EqualValues(t, 50, *status.LastBlockID)
}

func TestGetMeta(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	require.NoError(t, db.DropAll(dbHandle))

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// nothing set yet -> should fail
	_, err := GetMeta(ctx, dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set schema version
	require.NoError(t, InitializeMeta(ctx, dbHandle, "Dash"))

	metaResult, err := GetMeta(ctx, dbHandle)
	require.NoError(t, err)
	require.NotNil(t, metaResult.SchemaVersion)
	require.Equal(t, db.SchemaVersion, *metaResult.SchemaVersion)
	require.Equal(t, "Dash", metaResult.BlockchainMode)
	require.NotEmpty(t, metaResult.CreationTime)
}
