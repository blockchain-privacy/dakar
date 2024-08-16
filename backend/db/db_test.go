package db

import (
	"backend/external"
	"backend/testhelper"
	"context"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	InitLogger()
	testhelper.RunDgraphTests(m, &dbHandle.DB)
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

func TestExecTx(t *testing.T) {
	testhelper.SkipIfNoDB(t)

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
	testhelper.SkipIfNoDB(t)

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
	testhelper.SkipIfNoDB(t)

	require.Error(t, TxWithRetry(dbHandle, time.Duration(0), &api.Request{
		Query: `{q(func:uid(0x1)){uid}}`, CommitNow: true}))

	require.Error(t, TxWithRetry(dbHandle, time.Minute, nil))

	require.NoError(t, TxWithRetry(dbHandle, time.Minute, &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	}))
}

func TestTxWithRetryAndResponse(t *testing.T) {
	testhelper.SkipIfNoDB(t)

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
	testhelper.SkipIfNoDB(t)

	require.Error(t, ExistingTxWithRetry(dbHandle.NewTxn(), time.Duration(0), &api.Request{
		Query: `{q(func:uid(0x1)){uid}}`, CommitNow: true}))

	require.Error(t, ExistingTxWithRetry(dbHandle.NewTxn(), time.Minute, nil))

	require.NoError(t, ExistingTxWithRetry(dbHandle.NewTxn(), time.Minute, &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	}))
}

func TestExistingTxWithRetryAndResponse(t *testing.T) {
	testhelper.SkipIfNoDB(t)

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
	testhelper.SkipIfNoDB(t)

	_, err := execReadOnlyTx(dbHandle, time.Minute, "", nil)
	require.Error(t, err)

	_, err = execReadOnlyTx(dbHandle, time.Duration(0), "{q(func:uid(0x1)){uid}}", nil)
	require.Error(t, err)

	_, err = execReadOnlyTx(dbHandle, time.Minute, "{q(func:uid(0x1)){uid}}", nil)
	require.NoError(t, err)
}

func TestReadOnlyTxVarWithRetry(t *testing.T) {
	testhelper.SkipIfNoDB(t)

	_, err := ReadOnlyTxVarWithRetry(dbHandle, time.Minute, "", nil)
	require.Error(t, err)

	_, err = ReadOnlyTxVarWithRetry(dbHandle, time.Duration(0), "{q(func:uid(0x1)){uid}}", nil)
	require.Error(t, err)

	_, err = ReadOnlyTxVarWithRetry(dbHandle, time.Minute, "{q(func:uid(0x1)){uid}}", nil)
	require.NoError(t, err)
}

func TestReadOnlyTxWithRetry(t *testing.T) {
	testhelper.SkipIfNoDB(t)

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
	testhelper.SkipIfNoDB(t)
	require.NoError(t, DropAll(dbHandle))
}

func TestCreateClient(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	name, ok := testhelper.GetDBName()
	if !ok {
		t.Fatal("environment variable " + testhelper.EnvDBHostname + " is not set")
	}
	_, c, err := external.CreateClient(name + ":9080")
	require.NoError(t, err)
	require.NoError(t, c.Close())
}

func TestGetTypeByUID(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	SetupDBWithoutData(t, dbHandle)

	// empty db
	_, err := GetTypeByUID(context.Background(), dbHandle, "0x123")
	require.Error(t, err)

	SetupDB(t, dbHandle, testhelper.UseBlockFile)
	ctx := context.Background()
	txUID, err := GetTransactionUID(ctx, dbHandle, "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb")
	require.NoError(t, err)

	typeString, err := GetTypeByUID(ctx, dbHandle, txUID)
	require.NoError(t, err)
	require.Equal(t, "Transaction", typeString)
}
