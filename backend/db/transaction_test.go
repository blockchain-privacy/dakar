// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/constants"
)

func TestTransaction_SetDType(t *testing.T) {
	tx := Transaction{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}

	tx.SetDType()

	require.Equal(t, []string{TransactionDType}, tx.DType)
}

func TestTransaction_CalculateTransactionFee(t *testing.T) {
	tx := Transaction{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}

	require.NoError(t, tx.CalculateTransactionFee())
	require.Zero(t, *tx.Fee)

	// input amount = 40
	tx.Inputs = []Output{{Amount: new(int64(10))}, {Amount: new(int64(10))},
		{Amount: new(int64(10))}, {Amount: new(int64(10))}}
	// output amount = 30
	tx.Outputs = []Output{{Amount: new(int64(10))}, {Amount: new(int64(10))}, {Amount: new(int64(10))}}

	require.NoError(t, tx.CalculateTransactionFee())
	require.Equal(t, int64(10), *tx.Fee)

	// output amount = 40
	tx.Outputs = []Output{{Amount: new(int64(10))}, {Amount: new(int64(10))},
		{Amount: new(int64(10))}, {Amount: new(int64(10))}}

	require.NoError(t, tx.CalculateTransactionFee())
	require.Zero(t, *tx.Fee)

	// output amount = 50
	tx.Outputs = []Output{{Amount: new(int64(10))}, {Amount: new(int64(10))},
		{Amount: new(int64(10))}, {Amount: new(int64(10))}, {Amount: new(int64(10))}}
	require.NoError(t, tx.CalculateTransactionFee())
	require.Equal(t, int64(-10), *tx.Fee)

	tx.Outputs = []Output{{Amount: new(int64(10))}, {Amount: nil}}
	require.Error(t, tx.CalculateTransactionFee())

	tx.Outputs = []Output{{Amount: new(int64(10))}}
	tx.Inputs = []Output{{Amount: new(int64(10))}, {Amount: nil}}
	require.Error(t, tx.CalculateTransactionFee())
}

func TestFrontendTransaction_String(t *testing.T) {
	tx := FrontendTransaction{
		BlockHash: "some_long_hex_hash",
		Hash:      "some_long_hex_hash",
	}
	require.NotEmpty(t, tx.String())
}

func TestGetTransactionsOutputs(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	ctx, cancel := GetTaskContext()
	defer cancel()

	// test null input
	outputs, err := GetTransactionsOutputs(ctx, dbHandle, nil)
	require.Error(t, err)
	require.Nil(t, outputs)

	// test invalid input
	outputs, err = GetTransactionsOutputs(ctx, dbHandle, []string{"some_invalid_transaction_hash"})
	require.Error(t, err)
	require.Nil(t, outputs)

	// 2 transaction should return two mappings
	outputs, err = GetTransactionsOutputs(ctx, dbHandle, []string{
		"91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb",
		"0cfd028caf97751603255b1467085c3ccc5d476d79810ba9608d63587c7986f8",
	})
	require.NoError(t, err)
	require.Len(t, outputs, 2)

	var outputCount int
	for _, mapping := range outputs {
		outputCount += len(mapping.Outputs)
	}

	require.Equal(t, 3, outputCount, "unexpected number of outputs")
}

func TestGetTransactionByBlock(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	ctx, cancel := GetTaskContext()
	defer cancel()

	// only blocks beginning from height 60000 are in the DB, so it should fail
	transactions, err := GetTransactionsByBlock(ctx, dbHandle, 1, 1, nil)
	require.Error(t, err)
	require.Nil(t, transactions)

	transactions, err = GetTransactionsByBlock(ctx, dbHandle, 60001, 60001, nil)
	require.NoError(t, err)
	require.Len(t, transactions, 4)
}

func TestGetInputOutputAddressCounts(t *testing.T) {
	dbHandle := GetDBConnection(t, "")
	ctx, cancel := GetTaskContext()
	defer cancel()

	// invalid input
	inputCount, outputCount, err := GetInputOutputAddressCounts(ctx, dbHandle, "")
	require.Error(t, err)
	require.Zero(t, inputCount)
	require.Zero(t, outputCount)

	// invalid input should return no error but two zero counts
	inputCount, outputCount, err = GetInputOutputAddressCounts(ctx, dbHandle, "0x123FFFF")
	require.NoError(t, err)
	require.Zero(t, inputCount)
	require.Zero(t, outputCount)

	ChangeDBContent(dbHandle, UseBlockFile)

	transactions, err := GetTransactionsByBlock(ctx, dbHandle, 60001, 60001, nil)
	require.NoError(t, err)
	require.Len(t, transactions, 4)

	type testCase struct {
		txhash     string
		uid        string
		numInputs  int
		numOutputs int
	}

	cases := []testCase{
		{
			txhash:     "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13",
			numInputs:  3,
			numOutputs: 9,
		},
		{
			txhash:     "38c54271cb439357f1c02d4db06c6a2715d77b68e920f6ccb9f13d66fe233384",
			numInputs:  0,
			numOutputs: 1,
		},
		{
			txhash:     "818dae776566815b8d5307f8597fc8c1db737e933a4605e1841a83f078731638",
			numInputs:  1,
			numOutputs: 2,
		},
		{
			txhash:     "1f22cdab5f3543a49f2b0ab8a0ea7858fd4459e94f3074c58c39044bd63a8aff",
			numInputs:  1,
			numOutputs: 1,
		},
	}
	var found int
	for _, tx := range transactions {
		for i, c := range cases {
			if c.txhash == tx.Hash {
				cases[i].uid = tx.UID
				found++
				break
			}
		}
	}

	if len(cases) != found {
		t.Error(len(cases), "cases, but", found, "found")
	}

	for _, c := range cases {
		inputCount, outputCount, err = GetInputOutputAddressCounts(ctx, dbHandle, c.uid)
		require.NoError(t, err)
		require.Equal(t, c.numInputs, inputCount, c.txhash)
		require.Equal(t, c.numOutputs, outputCount, c.txhash)
	}
}

func TestGetOutputAddressCounts(t *testing.T) {
	dbHandle := GetDBConnection(t, "")
	ctx, cancel := GetTaskContext()
	defer cancel()

	// invalid input
	outputCount, err := GetOutputAddressCounts(ctx, dbHandle, "")
	require.Error(t, err)
	require.Zero(t, outputCount)

	// invalid input should return no error but two zero counts
	outputCount, err = GetOutputAddressCounts(ctx, dbHandle, "0x123FFFF")
	require.NoError(t, err)
	require.Zero(t, outputCount)

	ChangeDBContent(dbHandle, UseBlockFile)

	transactions, err := GetTransactionsByBlock(ctx, dbHandle, 60001, 60001, nil)
	require.NoError(t, err)
	require.Len(t, transactions, 4)

	type testCase struct {
		txhash     string
		uid        string
		numOutputs int
	}

	cases := []testCase{
		{
			txhash:     "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13",
			numOutputs: 9,
		},
		{
			txhash:     "38c54271cb439357f1c02d4db06c6a2715d77b68e920f6ccb9f13d66fe233384",
			numOutputs: 1,
		},
		{
			txhash:     "818dae776566815b8d5307f8597fc8c1db737e933a4605e1841a83f078731638",
			numOutputs: 2,
		},
		{
			txhash:     "1f22cdab5f3543a49f2b0ab8a0ea7858fd4459e94f3074c58c39044bd63a8aff",
			numOutputs: 1,
		},
	}
	var found int
	for _, tx := range transactions {
		for i, c := range cases {
			if c.txhash == tx.Hash {
				cases[i].uid = tx.UID
				found++
				break
			}
		}
	}

	if len(cases) != found {
		t.Error(len(cases), "cases, but", found, "found")
	}

	for _, c := range cases {
		outputCount, err = GetOutputAddressCounts(ctx, dbHandle, c.uid)
		require.NoError(t, err)
		require.Equal(t, c.numOutputs, outputCount, c.txhash)
	}
}

func TestGetFrontendTransaction(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)
	const blockHash = "00000000000cfe64fca7b5c3a8ad1ee39dd3f380aeb56027bc25e97904d2c99e"
	const txHash1 = "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13"
	transaction, err := GetFrontendTransaction(t.Context(), dbHandle, txHash1)
	require.NoError(t, err)
	require.Len(t, transaction, 1)
	require.Equal(t, txHash1, transaction[0].Hash)
	require.Equal(t, blockHash, transaction[0].BlockHash)
	require.Equal(t, int64(0), transaction[0].Fee)
	require.Empty(t, transaction[0].Type)

	const txHash2 = "818dae776566815b8d5307f8597fc8c1db737e933a4605e1841a83f078731638"
	transaction, err = GetFrontendTransaction(t.Context(), dbHandle, txHash2)
	require.NoError(t, err)
	require.Len(t, transaction, 1)
	require.Equal(t, txHash2, transaction[0].Hash)
	require.Equal(t, blockHash, transaction[0].BlockHash)
	require.Equal(t, int64(100000), transaction[0].Fee)
	require.Empty(t, transaction[0].Type)
}

func TestGetTransactionBlockID(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	const txHash1 = "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13"
	id, err := GetTransactionBlockID(t.Context(), dbHandle, txHash1)
	require.NoError(t, err)
	require.EqualValues(t, 60001, id)

	const txHash2 = "d0bc5aba5a81df73b706d7819956fb298e03baf52a97c736bb588dfd3586e849"
	id, err = GetTransactionBlockID(t.Context(), dbHandle, txHash2)
	require.NoError(t, err)
	require.EqualValues(t, 60007, id)
}

func TestUpdateTransactions(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	ctx, cancel := GetTaskContext()
	defer cancel()

	// empty slice should fail
	require.Error(t, UpdateTransactions(ctx, dbHandle, nil))
	require.Error(t, UpdateTransactions(ctx, dbHandle, []Transaction{}))

	transactions, err := GetTransactionsByBlock(ctx, dbHandle, 60001, 60001, nil)
	require.NoError(t, err)

	// no mixing transactions should be in this block
	for _, tx := range transactions {
		require.NotEqual(t, constants.TypeDashMixing, tx.Type)
	}

	// set all transactions to be a mixing transaction
	for i := range transactions {
		transactions[i].Type = constants.TypeDashMixing
	}

	require.NoError(t, UpdateTransactions(ctx, dbHandle, transactions))

	transactions, err = GetTransactionsByBlock(ctx, dbHandle, 60001, 60001, nil)
	require.NoError(t, err)

	// all transactions should now have the transaction type set to 'mixing'
	for _, tx := range transactions {
		require.Equal(t, constants.TypeDashMixing, tx.Type)
	}
}

func TestGetTransactionUID(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	ctx := t.Context()
	_, err := GetTransactionUID(ctx, dbHandle, "")
	require.Error(t, err)

	_, err = GetTransactionUID(ctx, dbHandle, "transaction_hash_which_does_not_exist")
	require.Error(t, err)

	const txHash1 = "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13"
	uid, err := GetTransactionUID(ctx, dbHandle, txHash1)
	require.NoError(t, err)
	require.NotEmpty(t, uid)
}

func TestGetOutputs(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	ctx, cancel := GetTaskContext()
	defer cancel()

	transactions, err := GetOutputs(ctx, dbHandle, BlockFileFirstBlock, BlockFileLastBlock)
	require.NoError(t, err)
	require.Len(t, transactions, 56)

	transactions, err = GetOutputs(ctx, dbHandle, 60007, 60007)
	require.NoError(t, err)
	require.Len(t, transactions, 7)

	// should return an empty transaction slice, because this block range is not included in the database
	transactions, err = GetOutputs(ctx, dbHandle, 1, 10)
	require.NoError(t, err)
	require.Empty(t, transactions)
}

func TestGetTransaction(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	ctx, cancel := GetTaskContext()
	defer cancel()

	tests := []struct {
		txHash  string
		wantErr bool
	}{
		{
			txHash:  "",
			wantErr: true,
		},
		{
			txHash:  "0cfd028caf97751603255b1467085c3ccc5d476d79810ba9608d63587c7986f8",
			wantErr: false,
		},
		{
			txHash:  "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		gotTransaction, err := GetTransaction(ctx, dbHandle, tt.txHash)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, gotTransaction)
		}
	}
}

func TestGetTransactionUIDMapping(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	ctx, cancel := GetTaskContext()
	defer cancel()

	transactions, err := GetTransactionsByBlock(ctx, dbHandle, 60005, 60005, nil)
	require.NoError(t, err)
	require.Len(t, transactions, 7)

	var uids [7]string
	for i, tx := range transactions {
		uids[i] = tx.UID
	}

	tests := []struct {
		txUids      []string
		wantTxCount int
		wantErr     bool
	}{
		{
			txUids:      nil,
			wantTxCount: 0,
			wantErr:     true,
		},
		{
			txUids:      uids[:],
			wantTxCount: 7,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		gotTxs, err := GetTransactionUIDMapping(t.Context(), dbHandle, tt.txUids)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Len(t, gotTxs, tt.wantTxCount)
		}
	}
}
