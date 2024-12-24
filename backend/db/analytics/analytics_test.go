package analytics

import (
	"backend/db"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle = &testhelper.TestDB{}

func TestMain(m *testing.M) {
	testhelper.RunDgraphTests(m, dbHandle)
}

func TestGetAllFMIClusters(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	clusters, err := GetAllFMIClusters(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, clusters)
}

func TestGetTransactionCountPerCluster(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	clusters, err := GetAllFMIClusters(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, clusters)

	for i, c := range clusters {
		inputCount, outputCount, err := GetTransactionCountPerCluster(ctx, dbHandle, c)
		require.NoError(t, err)
		require.NotZero(t, inputCount+outputCount)

		if i > 100 {
			break
		}
	}
}
