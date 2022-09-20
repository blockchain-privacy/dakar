package db

import (
	"backend/external"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle external.Database

func setupDB(t *testing.T) {
	// reset db
	require.NoError(t, DropAll(dbHandle))

	// set up schema
	require.NoError(t, SetupSchema(dbHandle))
}

func TestGetTransactionsOutputs(t *testing.T) {
	setupDB(t)
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
	setupDB(t)

	// nothing in DB yet, so it should fail
	transactions, err := GetTransactionByBlock(dbHandle, 1)
	require.Error(t, err)
	require.Nil(t, transactions)
}

func TestGetOutputAddressCounts(t *testing.T) {
	setupDB(t)

	// nothing in DB yet, so it should fail
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
