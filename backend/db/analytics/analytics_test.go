package analytics

import (
	"backend/db"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	db.InitLogger()
	testhelper.RunDgraphTests(m, &dbHandle.DB)
}

func TestGetAllFMIClusters(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)
	clusters, err := GetAllFMIClusters(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, clusters)
}

func TestGetTransactionCountPerCluster(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)
	clusters, err := GetAllFMIClusters(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, clusters)

	for i, c := range clusters {
		inputCount, outputCount, err := GetTransactionCountPerCluster(dbHandle, c)
		require.NoError(t, err)
		require.NotZero(t, inputCount+outputCount)

		if i > 100 {
			break
		}
	}
}
