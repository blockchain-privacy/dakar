package db

import (
	"backend/testhelper"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const blockFileName = "testdata/blocks_60000_60020.json"

func TestMain(m *testing.M) {
	RunDgraphTests(m, &dbHandle, ContainerNameDB)
}

func TestInfo(t *testing.T) {
	require.NotPanics(t, func() {
		info("test123")
		info()
		info(-1)
		info(nil)
	})
}

func TestGetBackendContext(t *testing.T) {
	require.NotPanics(t, func() {
		_, cancel := GetBackendContext()
		cancel()
	})
}

func TestGetFrontendContext(t *testing.T) {
	require.NotPanics(t, func() {
		_, cancel := GetFrontendContext()
		cancel()
	})
}

func TestExecTx(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	_, err := execTx(dbHandle, time.Duration(0), &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	})
	require.Error(t, err)

	_, err = execTx(dbHandle, time.Minute, nil)
	require.Error(t, err)

	_, err = execTx(dbHandle, time.Minute, &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	})
	require.NoError(t, err)
}

func TestExecExistingTx(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	_, err := execExistingTx(dbHandle.NewTxn(), time.Duration(0), &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	})
	require.Error(t, err)

	_, err = execExistingTx(dbHandle.NewTxn(), time.Minute, nil)
	require.Error(t, err)

	_, err = execExistingTx(dbHandle.NewTxn(), time.Minute, &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	})
	require.NoError(t, err)
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
	testhelper.SkipIfNotCI(t)
	require.NoError(t, DropAll(dbHandle))
}
