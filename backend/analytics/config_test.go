package analytics

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewBitcoinConfig(t *testing.T) {
	require.NotEqual(t, Config{}, NewBitcoinConfig())
}

func TestNewDashConfig(t *testing.T) {
	require.NotEqual(t, Config{}, NewDashConfig())
}

func TestNewDogecoinConfig(t *testing.T) {
	require.NotEqual(t, Config{}, NewDogecoinConfig())
}
