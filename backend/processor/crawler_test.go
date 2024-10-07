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

	// reduce fork range limit so there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	crawler := NewCrawler(ctx, dbHandle, client, 0, c)
	crawler.RegisterMetrics(prometheus.NewRegistry())
	// initial state is not set yet
	require.Error(t, crawler.IncrementState())

	require.NoError(t, crawler.CalculateInitialState(ctx))
	block, err := crawler.Next(ctx)
	require.NoError(t, err)
	require.True(t, block)

	require.NoError(t, crawler.IncrementState())
}

func TestCrawler_Empty(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	// reduce fork range limit so there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	crawler := NewCrawler(ctx, dbHandle, client, 0, c)
	crawler.RegisterMetrics(prometheus.NewRegistry())
	require.True(t, crawler.Empty())

	require.NoError(t, crawler.CalculateInitialState(ctx))
	isNewBlock, err := crawler.Next(ctx)
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

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	crawler := NewCrawler(ctx, dbHandle, client, 0, NewBitcoinConfig())
	crawler.RegisterMetrics(prometheus.NewRegistry())
	require.NoError(t, crawler.CalculateInitialState(ctx))
}

func TestCrawler_PostExecution(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	crawler := NewCrawler(ctx, dbHandle, client, 0, NewBitcoinConfig())

	require.NoError(t, crawler.PostExecution(ctx))
}

func TestCrawler_NextBlock(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	// reduce fork range limit so there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	crawler := NewCrawler(ctx, dbHandle, client, 0, c)
	crawler.RegisterMetrics(prometheus.NewRegistry())
	_, err := crawler.Next(ctx)
	require.Error(t, err)

	require.NoError(t, crawler.CalculateInitialState(ctx))
	isNewBlock, err := crawler.Next(ctx)
	require.NoError(t, err)
	require.True(t, isNewBlock)
}

func TestCrawler_Iterate(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	// reduce fork range limit so there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	crawler := NewCrawler(ctx, dbHandle, client, 0, c)
	crawler.RegisterMetrics(prometheus.NewRegistry())
	_, err := crawler.Iterate(ctx)
	require.Error(t, err)

	require.NoError(t, crawler.CalculateInitialState(ctx))
	isNewBlock, err := crawler.Next(ctx)
	require.NoError(t, err)
	require.True(t, isNewBlock)

	// make block height higher so Empty() returns false.
	// This is needed because the testrpc client has fewer blocks than the ForkRangeLimit
	crawler.state.top = 1000000

	success, err := crawler.Iterate(ctx)
	require.NoError(t, err)
	require.True(t, success)
}
