package cliutil

import (
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestShowCallInfo(t *testing.T) {
	info := ShowCallInfo()
	require.NotEmpty(t, info, "call info is empty")
}

func TestBuildEndpoint(t *testing.T) {
	cases := []struct {
		host string
		port uint
	}{
		{"localhost", uint(8081)},
		{"asfd", uint(1)},
		{"localhost", uint(0)},
		{"", uint(50)},
		{"     ", uint(10)},
	}

	for _, c := range cases {
		endpoint, err := BuildEndpoint(c.host, c.port)
		if len(strings.TrimSpace(c.host)) == 0 || c.port == 0 {
			require.NotNil(t, err)
			require.Empty(t, endpoint, "endpoint should be empty")
		} else {
			require.Nil(t, err)
			require.NotEmpty(t, endpoint, "created endpoint is empty")
		}
	}
}

func TestGetLogfile(t *testing.T) {
	// this should not work
	logfile, err := GetLogfile("")
	require.NotNil(t, err)
	require.Nil(t, logfile)

	// this should work

	// getting the name for a temporary file
	file, err := os.CreateTemp("", "go_test_logfile")
	require.Nil(t, err)
	fName := file.Name()
	err = file.Close()
	require.Nil(t, err)
	logfile, err = GetLogfile(fName)
	require.Nil(t, err)
	require.NotNil(t, logfile)
}
