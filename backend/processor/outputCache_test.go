package processor

import (
	"backend/db"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoadUTXOCache(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)
	// nothing in DB so should not return anything
	cache, err := newUTXOCache(dbHandle, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, cache)
	require.Zero(t, cache.getOutputCounts())

	db.SetupDB(t, dbHandle, blockFileName)

	cache, err = newUTXOCache(dbHandle, 60020, 20)
	require.NoError(t, err)
	require.NotNil(t, cache)
	require.Equal(t, 2, cache.getOutputCounts())
}

func TestOutputCacheOffline(t *testing.T) {
	// create cache
	cache := newOutputCache()

	// cache empty -> should all return zero/nil
	require.Zero(t, cache.getOutputCounts())
	require.Nil(t, cache.getOutput("some_invalid_tx_hash", 0))
	require.Nil(t, cache.getAndEvictOutput("some_invalid_tx_hash", 0))

	// try to set zero elements for key
	require.Error(t, cache.setOutputs("txhash1", nil))
	require.Error(t, cache.setOutputs("txhash1", []db.Output{}))

	// try to set empty key
	require.Error(t, cache.setOutputs("", []db.Output{{}, {}}))

	// valid key, but invalid elements
	require.Error(t, cache.setOutputs("txhash1", []db.Output{{}, {}}))

	one := uint32(1)
	const key = "txhash1"
	outputs := []db.Output{{OutputIndex: &one}}
	require.NoError(t, cache.setOutputs(key, outputs))
	// duplicate key is fine
	require.NoError(t, cache.setOutputs(key, outputs))

	require.Equal(t, 1, cache.getOutputCounts())
	require.Equal(t, outputs[0], *cache.getOutput(key, one))
	require.Equal(t, outputs[0], *cache.getAndEvictOutput(key, one))
	require.Zero(t, cache.getOutputCounts())
}
