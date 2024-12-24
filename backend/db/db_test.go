package db

import (
	"backend/external"
	"backend/testhelper"
	"context"
	"errors"
	"github.com/dgraph-io/dgo/v240"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle = &testhelper.TestDB{}

func TestMain(m *testing.M) {
	testhelper.RunDgraphTests(m, dbHandle)
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
		_, cancel := GetTaskContext()
		cancel()
	})
}

func TestExecTx(t *testing.T) {
	testhelper.SkipIfNoDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	_, err := ExecTx(ctx, dbHandle.NewTxn(), &api.Request{
		Query:     `{q(func:uid(0x1)){uid}}`,
		CommitNow: true,
	})
	require.Error(t, err)

	ctx, cancel = GetShortTaskContext()
	defer cancel()

	_, err = ExecTx(ctx, nil, nil)
	require.Error(t, err)

	_, err = ExecTx(ctx, dbHandle.NewTxn(), &api.Request{
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

func TestWithRetry(t *testing.T) {
	executionCounter := 0
	errorFunction := func() error {
		executionCounter++
		return errors.New("some error")
	}

	require.Error(t, WithRetry(errorFunction, 0))
	require.EqualValues(t, 1, executionCounter)

	executionCounter = 0
	noErrorFunction := func() error {
		executionCounter++
		return nil
	}

	require.NoError(t, WithRetry(noErrorFunction, 0))
	require.EqualValues(t, 1, executionCounter)

	executionCounter = 0
	txAborted := func() error {
		executionCounter++
		return dgo.ErrAborted
	}

	require.Error(t, WithRetry(txAborted, 0))
	require.EqualValues(t, maxRetries, executionCounter)

	executionCounter = 0
	txAbortedThenSuccessful := func() error {
		executionCounter++

		if executionCounter == 3 {
			return nil
		}

		return dgo.ErrAborted
	}

	require.NoError(t, WithRetry(txAbortedThenSuccessful, 0))
	require.EqualValues(t, 3, executionCounter)
}
