package db

import (
	"backend/external"
	"github.com/stretchr/testify/require"
	"testing"
)

var h external.Database

func TestMain(m *testing.M) {
	RunDgraphTests(m, &h)
}

func TestCreateCommaList(t *testing.T) {
	type testCase struct {
		uids   []string
		result string
	}

	var cases = []testCase{
		{
			uids:   []string{},
			result: "",
		},
		{
			uids:   nil,
			result: "",
		},
		{
			uids:   []string{"123", "456"},
			result: "123,456",
		},
		{
			uids:   []string{"123", ""},
			result: "123,",
		},
	}

	for _, c := range cases {
		require.Equal(t, CreateCommaList(c.uids), c.result)
	}
}

func TestCreateCommaArray(t *testing.T) {
	type testCase struct {
		uids   []string
		result string
	}

	var cases = []testCase{
		{
			uids:   []string{},
			result: "[]",
		},
		{
			uids:   nil,
			result: "[]",
		},
		{
			uids:   []string{"123", "456"},
			result: "[123,456]",
		},
		{
			uids:   []string{"123", ""},
			result: "[123,]",
		},
	}

	for _, c := range cases {
		require.Equal(t, c.result, CreateCommaArray(c.uids))
	}
}

func TestDropAll(t *testing.T) {
	require.NoError(t, DropAll(h))
}
