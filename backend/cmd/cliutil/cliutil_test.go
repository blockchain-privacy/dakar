package cliutil

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

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
			require.Error(t, err)
			require.Empty(t, endpoint, "endpoint should be empty")
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, endpoint, "created endpoint is empty")
		}
	}
}

func TestGetOneKey(t *testing.T) {
	require.Equal(t, "a", GetOneKey(map[string]int{"a": 1}))
	require.Equal(t, "", GetOneKey(map[string]int{}))
	var m map[string]int
	require.Equal(t, "", GetOneKey(m))
	// can not check for return value as key is indeterminate with map len > 1
	GetOneKey(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4})
}

func TestGetOneItem(t *testing.T) {
	k1, v1 := GetOneItem(map[string]int{"a": 1})
	require.Equal(t, "a", k1)
	require.Equal(t, 1, v1)

	k2, v2 := GetOneItem(map[string]int{})
	require.Equal(t, "", k2)
	require.Equal(t, 0, v2)

	var m map[string]int
	k3, v3 := GetOneItem(m)
	require.Equal(t, "", k3)
	require.Equal(t, 0, v3)
	// can not check for return value as key is indeterminate with map len > 1
	GetOneItem(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4})
}

func TestGetMapKeys(t *testing.T) {
	type testCase struct {
		args        map[string]int
		wantNumKeys int
	}
	tests := []testCase{
		{
			args:        map[string]int{},
			wantNumKeys: 0,
		},
		{
			args:        map[string]int{"a": 1},
			wantNumKeys: 1,
		},
		{
			args:        map[string]int{"asdf": 5, "aaa": 3},
			wantNumKeys: 2,
		},
		{
			args:        map[string]int(nil),
			wantNumKeys: 0,
		},
	}
	for _, tt := range tests {
		require.Len(t, GetMapKeys(tt.args), tt.wantNumKeys)
	}
}
