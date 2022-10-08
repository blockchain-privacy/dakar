package processor

import (
	"backend/db"
	"backend/testhelper"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"testing"
)

// unregisterCollectors unregisters all collectors of the classifier.
// This is needed because collectors can not be registered twice with the same default config.
func unregisterCollectors(c *Crawler) {
	if c == nil {
		return
	}

	prometheus.Unregister(c.blocks)
	prometheus.Unregister(c.transactions)
	prometheus.Unregister(c.blockHeight)
}

func TestNewCrawler(t *testing.T) {
	crawler := NewCrawler(context.Background(), nil, nil, nil, 0, Config{})
	unregisterCollectors(crawler)
	require.NotNil(t, crawler)
}

func TestCrawler_Name(t *testing.T) {
	crawler := NewCrawler(context.Background(), nil, nil, nil, 0, Config{})
	unregisterCollectors(crawler)
	require.NotEmpty(t, crawler.Name())
}

func TestCrawler_Logger(t *testing.T) {
	crawler := NewCrawler(context.Background(), nil, nil, nil, 0, Config{})
	unregisterCollectors(crawler)
	require.NotNil(t, crawler.Logger())
}

func TestCrawler_Context(t *testing.T) {
	crawler := NewCrawler(context.Background(), nil, nil, nil, 0, Config{})
	unregisterCollectors(crawler)
	require.NotNil(t, crawler.Context())
}

func TestCrawler_DB(t *testing.T) {
	crawler := NewCrawler(context.Background(), nil, nil, nil, 0, Config{})
	unregisterCollectors(crawler)
	require.Nil(t, crawler.DB())
}

func TestCrawler_IncrementState(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)

	crawler := NewCrawler(context.Background(), dbHandle, client, batchClient, 0, NewBitcoinConfig())
	unregisterCollectors(crawler)

	// initial state is not set yet
	require.Error(t, crawler.IncrementState())

	require.NoError(t, crawler.CalculateInitialState())
	block, err := crawler.NextBlock()
	require.NoError(t, err)
	require.True(t, block)

	require.NoError(t, crawler.IncrementState())
}

func TestCrawler_Empty(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)

	crawler := NewCrawler(context.Background(), dbHandle, client, batchClient, 0, NewBitcoinConfig())
	unregisterCollectors(crawler)
	require.True(t, crawler.Empty())

	require.NoError(t, crawler.CalculateInitialState())
	isNewBlock, err := crawler.NextBlock()
	require.NoError(t, err)
	require.True(t, isNewBlock)

	// make block height higher so Empty() returns false
	// This is needed because the testrpc client has fewer blocks than the ForkRangeLimit
	crawler.state.top = 1000000

	require.False(t, crawler.Empty())
}

func TestCrawler_CalculateInitialState(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)

	crawler := NewCrawler(context.Background(), dbHandle, client, batchClient, 0, NewBitcoinConfig())
	unregisterCollectors(crawler)
	require.NoError(t, crawler.CalculateInitialState())
}

func TestCrawler_PostExecution(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)

	crawler := NewCrawler(context.Background(), dbHandle, client, batchClient, 0, NewBitcoinConfig())
	unregisterCollectors(crawler)

	require.NoError(t, crawler.PostExecution())
}

func TestCrawler_CurrentBlock(t *testing.T) {
	crawler := NewCrawler(context.Background(), nil, nil, nil, 0, Config{})
	unregisterCollectors(crawler)
	require.Zero(t, crawler.CurrentBlock())
}

func TestCrawler_NextBlock(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)

	crawler := NewCrawler(context.Background(), dbHandle, client, batchClient, 0, NewBitcoinConfig())
	unregisterCollectors(crawler)

	_, err := crawler.NextBlock()
	require.Error(t, err)

	require.NoError(t, crawler.CalculateInitialState())
	isNewBlock, err := crawler.NextBlock()
	require.NoError(t, err)
	require.True(t, isNewBlock)
}

func TestCrawler_Iterate(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)

	crawler := NewCrawler(context.Background(), dbHandle, client, batchClient, 0, NewBitcoinConfig())
	unregisterCollectors(crawler)

	_, err := crawler.Iterate()
	require.Error(t, err)

	require.NoError(t, crawler.CalculateInitialState())
	isNewBlock, err := crawler.NextBlock()
	require.NoError(t, err)
	require.True(t, isNewBlock)

	// make block height higher so Empty() returns false.
	// This is needed because the testrpc client has fewer blocks than the ForkRangeLimit
	crawler.state.top = 1000000

	success, err := crawler.Iterate()
	require.NoError(t, err)
	require.True(t, success)
}
