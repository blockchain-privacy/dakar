package analytics

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitLogger(t *testing.T) {
	require.NotPanics(t, func() {
		InitLogger()
	})
}
