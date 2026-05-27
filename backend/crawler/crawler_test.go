// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package crawler

import (
	"gitlab.com/blockchain-privacy/dakar/db"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewCrawler(t *testing.T) {
	crawler := NewCrawler(t.Context(), nil, nil, 0, Config{})
	require.NotNil(t, crawler)
}

func TestCrawler_Props(t *testing.T) {
	crawler := NewCrawler(t.Context(), nil, nil, 0, Config{})

	require.NotEmpty(t, crawler.Props())
}

func TestCrawler_IncrementState(t *testing.T) {
	db.SkipIfNoRPC(t)
	dbHandle := db.GetDBConnection(t, "")

	// reduce fork range limit so there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1

	ctx, cancel := db.GetTaskContext()
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
	db.SkipIfNoRPC(t)
	dbHandle := db.GetDBConnection(t, "")

	// reduce fork range limit so there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1

	ctx, cancel := db.GetTaskContext()
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
	db.SkipIfNoRPC(t)
	dbHandle := db.GetDBConnection(t, "")

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	crawler := NewCrawler(ctx, dbHandle, client, 0, NewBitcoinConfig())
	crawler.RegisterMetrics(prometheus.NewRegistry())
	require.NoError(t, crawler.CalculateInitialState(ctx))
}

func TestCrawler_PostExecution(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	crawler := NewCrawler(ctx, dbHandle, client, 0, NewBitcoinConfig())

	require.NoError(t, crawler.PostExecution(ctx))
}

func TestCrawler_NextBlock(t *testing.T) {
	db.SkipIfNoRPC(t)
	dbHandle := db.GetDBConnection(t, "")

	// reduce fork range limit so there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1

	ctx, cancel := db.GetTaskContext()
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
	db.SkipIfNoRPC(t)
	dbHandle := db.GetDBConnection(t, "")

	// reduce fork range limit so there is something to crawl
	c := NewBitcoinConfig()
	c.ForkRangeLimit = 1

	ctx, cancel := db.GetTaskContext()
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

func Test_createTransactionMap(t *testing.T) {
	db.SkipIfNoRPC(t)

	generateToAddress, err := client.GetNewAddress()
	require.NoError(t, err)

	blockHashes, err := client.GenerateToAddress(1, generateToAddress)
	require.NoError(t, err)
	require.NotEmpty(t, blockHashes)

	verboseBlock, err := client.GetBlockVerbose(blockHashes[0])
	require.NoError(t, err)

	txMap, err := createTransactionMap(client, verboseBlock.Tx)
	require.NoError(t, err)
	require.NotEmpty(t, txMap)
}
