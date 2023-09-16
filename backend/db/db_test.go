package db

import (
	"backend/external"
	"backend/testhelper"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const blockFileName = "testdata/blocks_60000_60020.json"

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	InitLogger()
	testhelper.RunDgraphTests(m, &dbHandle.DB, testhelper.ContainerNameDB)
}

func TestInfo(t *testing.T) {
	require.NotPanics(t, func() {
		info("test123")
		info("")
		info("", -1, nil)
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

func TestTxWithRetry(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	require.Error(t, TxWithRetry(dbHandle, time.Duration(0), &api.Request{
		Query: `{q(func:uid(0x1)){uid}}`, CommitNow: true}))

	require.Error(t, TxWithRetry(dbHandle, time.Minute, nil))

	require.NoError(t, TxWithRetry(dbHandle, time.Minute, &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	}))
}

func TestTxWithRetryAndResponse(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	_, err := TxWithRetryAndResponse(dbHandle, time.Duration(0), &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	})
	require.Error(t, err)

	_, err = TxWithRetryAndResponse(dbHandle, time.Minute, nil)
	require.Error(t, err)

	_, err = TxWithRetryAndResponse(dbHandle, time.Minute, &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	})
	require.NoError(t, err)
}

func TestExistingTxWithRetry(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	require.Error(t, ExistingTxWithRetry(dbHandle.NewTxn(), time.Duration(0), &api.Request{
		Query: `{q(func:uid(0x1)){uid}}`, CommitNow: true}))

	require.Error(t, ExistingTxWithRetry(dbHandle.NewTxn(), time.Minute, nil))

	require.NoError(t, ExistingTxWithRetry(dbHandle.NewTxn(), time.Minute, &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	}))
}

func TestExistingTxWithRetryAndResponse(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	_, err := ExistingTxWithRetryAndResponse(dbHandle.NewTxn(), time.Duration(0), &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	})
	require.Error(t, err)

	_, err = ExistingTxWithRetryAndResponse(dbHandle.NewTxn(), time.Minute, nil)
	require.Error(t, err)

	_, err = ExistingTxWithRetryAndResponse(dbHandle.NewTxn(), time.Minute, &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	})
	require.NoError(t, err)
}

func TestExecReadOnlyTx(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	_, err := execReadOnlyTx(dbHandle, time.Minute, "", nil)
	require.Error(t, err)

	_, err = execReadOnlyTx(dbHandle, time.Duration(0), "{q(func:uid(0x1)){uid}}", nil)
	require.Error(t, err)

	_, err = execReadOnlyTx(dbHandle, time.Minute, "{q(func:uid(0x1)){uid}}", nil)
	require.NoError(t, err)
}

func TestReadOnlyTxVarWithRetry(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	_, err := ReadOnlyTxVarWithRetry(dbHandle, time.Minute, "", nil)
	require.Error(t, err)

	_, err = ReadOnlyTxVarWithRetry(dbHandle, time.Duration(0), "{q(func:uid(0x1)){uid}}", nil)
	require.Error(t, err)

	_, err = ReadOnlyTxVarWithRetry(dbHandle, time.Minute, "{q(func:uid(0x1)){uid}}", nil)
	require.NoError(t, err)
}

func TestReadOnlyTxWithRetry(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	_, err := ReadOnlyTxWithRetry(dbHandle, time.Minute, "")
	require.Error(t, err)

	_, err = ReadOnlyTxWithRetry(dbHandle, time.Duration(0), "{q(func:uid(0x1)){uid}}")
	require.Error(t, err)

	_, err = ReadOnlyTxWithRetry(dbHandle, time.Minute, "{q(func:uid(0x1)){uid}}")
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

func TestCreateClient(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	_, c, err := external.CreateClient(string(testhelper.ContainerNameDB) + ":9080")
	require.NoError(t, err)
	require.NoError(t, c.Close())
}
