package processor

import (
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
	crawler := NewCrawler(context.Background(), nil, nil, nil, 0, Config{})
	unregisterCollectors(crawler)

	// initial state is not set yet
	require.Error(t, crawler.IncrementState())

	// todo use CalculateInitialState and call method again
}

func TestCrawler_Empty(t *testing.T) {
	crawler := NewCrawler(context.Background(), nil, nil, nil, 0, Config{})
	unregisterCollectors(crawler)
	require.True(t, crawler.Empty())

	// todo use CalculateInitialState and call method again
}
