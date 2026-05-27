// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"errors"
	"testing"

	"github.com/dgraph-io/dgo/v250"
	"github.com/dgraph-io/dgo/v250/protos/api"
	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/external"
)

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
	dbHandle := GetDBConnection(t, "")

	ctx, cancel := context.WithTimeout(t.Context(), 0)
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

func TestCreateCommaListQuotationMarks(t *testing.T) {
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
			result: "\"123\",\"456\"",
		},
		{
			uids:   []string{"123", ""},
			result: "\"123\",\"\"",
		},
	}

	for _, c := range cases {
		require.Equal(t, CreateCommaListQuotationMarks(c.uids), c.result)
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

func TestCreateClient(t *testing.T) {
	if !DoDBTests() {
		t.SkipNow()
	}

	name, ok := GetDBName()
	if !ok {
		t.Fatal("environment variable " + EnvDBHostname + " is not set")
	}

	d, err := external.CreateClient(name + ":9080")
	require.NoError(t, err)
	defer d.Close()
}

func TestCreateClientWithNamespace(t *testing.T) {
	if !DoDBTests() {
		t.SkipNow()
	}

	name, ok := GetDBName()
	if !ok {
		t.Fatal("environment variable " + EnvDBHostname + " is not set")
	}

	d, err := external.CreateClientWithNamespace(t.Context(), name+":9080",
		GetDBUser(), GetDBPassword(), 0)
	require.NoError(t, err)
	defer d.Close()
}

func TestGetTypeByUID(t *testing.T) {
	dbHandle := GetDBConnection(t, "")

	// empty db
	_, err := GetTypeByUID(t.Context(), dbHandle, "0x123")
	require.Error(t, err)

	ChangeDBContent(dbHandle, UseBlockFile)

	ctx := t.Context()
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
	require.Equal(t, 1, executionCounter)

	executionCounter = 0
	noErrorFunction := func() error {
		executionCounter++
		return nil
	}

	require.NoError(t, WithRetry(noErrorFunction, 0))
	require.Equal(t, 1, executionCounter)

	executionCounter = 0
	txAborted := func() error {
		executionCounter++
		return dgo.ErrAborted
	}

	require.Error(t, WithRetry(txAborted, 0))
	require.Equal(t, maxRetries, executionCounter)

	executionCounter = 0
	txAbortedThenSuccessful := func() error {
		executionCounter++

		if executionCounter == 3 {
			return nil
		}

		return dgo.ErrAborted
	}

	require.NoError(t, WithRetry(txAbortedThenSuccessful, 0))
	require.Equal(t, 3, executionCounter)
}
