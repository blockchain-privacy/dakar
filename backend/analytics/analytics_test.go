package analytics

import (
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestInitLogger(t *testing.T) {
	require.NotPanics(t, func() {
		InitLogger(log.Writer(), log.Flags())
	})
}
