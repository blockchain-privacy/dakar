package db

import (
	"backend/external"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle external.Database

func TestGetTransactionsOutputs(t *testing.T) {
	SetupDB(t, dbHandle, blockFileName)
	// test null input
	outputs, err := GetTransactionsOutputs(dbHandle, nil)
	require.Error(t, err)
	require.Nil(t, outputs)

	// test invalid input
	outputs, err = GetTransactionsOutputs(dbHandle, []string{"some_invalid_transaction_hash"})
	require.Error(t, err)
	require.Nil(t, outputs)
}

func TestGetTransactionByBlock(t *testing.T) {
	require.NoError(t, DropAll(dbHandle))

	// nothing in DB yet, so it should fail
	transactions, err := GetTransactionByBlock(dbHandle, 1)
	require.Error(t, err)
	require.Nil(t, transactions)

	SetupDB(t, dbHandle, blockFileName)

	// only blocks beginning from height 60000 are in the DB, so it should fail
	transactions, err = GetTransactionByBlock(dbHandle, 1)
	require.Error(t, err)
	require.Nil(t, transactions)

	transactions, err = GetTransactionByBlock(dbHandle, 60001)
	require.NoError(t, err)
	require.NotEmpty(t, transactions)
}

func TestGetOutputAddressCounts(t *testing.T) {
	// invalid input
	inputCount, outputCount, err := GetOutputAddressCounts(dbHandle, "")
	require.Error(t, err)
	require.Zero(t, inputCount)
	require.Zero(t, outputCount)

	// invalid input should return no error but two zero counts
	inputCount, outputCount, err = GetOutputAddressCounts(dbHandle, "0x123FFFF")
	require.NoError(t, err)
	require.Zero(t, inputCount)
	require.Zero(t, outputCount)
}
