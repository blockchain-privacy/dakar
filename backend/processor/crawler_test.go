package processor

import (
	"backend/db"
	"backend/testhelper"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCrawler(t *testing.T) {
	crawler := NewCrawler(context.Background(), nil, nil, 0, Config{})
	require.NotNil(t, crawler)
}

func TestCrawler_Props(t *testing.T) {
	crawler := NewCrawler(context.Background(), nil, nil, 0, Config{})

	require.NotEmpty(t, crawler.Props())
}

func TestCrawler_IncrementState(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	// reduce fork range limit there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1

	crawler := NewCrawler(context.Background(), dbHandle, client, 0, c)
	crawler.RegisterMetrics(prometheus.NewRegistry())
	// initial state is not set yet
	require.Error(t, crawler.IncrementState())

	require.NoError(t, crawler.CalculateInitialState())
	block, err := crawler.Next()
	require.NoError(t, err)
	require.True(t, block)

	require.NoError(t, crawler.IncrementState())
}

func TestCrawler_Empty(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	// reduce fork range limit there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1
	crawler := NewCrawler(context.Background(), dbHandle, client, 0, c)
	crawler.RegisterMetrics(prometheus.NewRegistry())
	require.True(t, crawler.Empty())

	require.NoError(t, crawler.CalculateInitialState())
	isNewBlock, err := crawler.Next()
	require.NoError(t, err)
	require.True(t, isNewBlock)

	// make block height higher so Empty() returns false
	// This is needed because the testrpc client has fewer blocks than the ForkRangeLimit
	crawler.state.top = 1000000

	require.False(t, crawler.Empty())
}

func TestCrawler_CalculateInitialState(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	crawler := NewCrawler(context.Background(), dbHandle, client, 0, NewBitcoinConfig())
	crawler.RegisterMetrics(prometheus.NewRegistry())
	require.NoError(t, crawler.CalculateInitialState())
}

func TestCrawler_PostExecution(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)

	crawler := NewCrawler(context.Background(), dbHandle, client, 0, NewBitcoinConfig())

	require.NoError(t, crawler.PostExecution())
}

func TestCrawler_NextBlock(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	// reduce fork range limit there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1
	crawler := NewCrawler(context.Background(), dbHandle, client, 0, c)
	crawler.RegisterMetrics(prometheus.NewRegistry())
	_, err := crawler.Next()
	require.Error(t, err)

	require.NoError(t, crawler.CalculateInitialState())
	isNewBlock, err := crawler.Next()
	require.NoError(t, err)
	require.True(t, isNewBlock)
}

func TestCrawler_Iterate(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	// reduce fork range limit there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1
	crawler := NewCrawler(context.Background(), dbHandle, client, 0, c)
	crawler.RegisterMetrics(prometheus.NewRegistry())
	_, err := crawler.Iterate()
	require.Error(t, err)

	require.NoError(t, crawler.CalculateInitialState())
	isNewBlock, err := crawler.Next()
	require.NoError(t, err)
	require.True(t, isNewBlock)

	// make block height higher so Empty() returns false.
	// This is needed because the testrpc client has fewer blocks than the ForkRangeLimit
	crawler.state.top = 1000000

	success, err := crawler.Iterate()
	require.NoError(t, err)
	require.True(t, success)
}
