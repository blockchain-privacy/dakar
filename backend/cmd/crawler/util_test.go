package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	require.NotPanics(t, func() {
		printVersion("Dash")
	})
}
