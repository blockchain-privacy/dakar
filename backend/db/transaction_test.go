package db

import (
	"backend/constants"
	"backend/testhelper"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransaction_String(t *testing.T) {
	tx := Transaction{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}
	require.NotEmpty(t, tx.String())
}

func TestTransaction_SetDType(t *testing.T) {
	tx := Transaction{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}

	tx.SetDType()

	require.Equal(t, []string{transactionDType}, tx.DType)
}

func getNumPointer[number int64 | uint64 | uint32](n number) *number {
	return &n
}

func TestTransaction_CalculateTransactionFee(t *testing.T) {
	tx := Transaction{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}

	require.NoError(t, tx.CalculateTransactionFee())
	require.Zero(t, *tx.Fee)

	// input amount = 40
	tx.Inputs = []Output{{Amount: getNumPointer[int64](10)}, {Amount: getNumPointer[int64](10)},
		{Amount: getNumPointer[int64](10)}, {Amount: getNumPointer[int64](10)}}
	// output amount = 30
	tx.Outputs = []Output{{Amount: getNumPointer[int64](10)}, {Amount: getNumPointer[int64](10)}, {Amount: getNumPointer[int64](10)}}

	require.NoError(t, tx.CalculateTransactionFee())
	require.Equal(t, int64(10), *tx.Fee)

	// output amount = 40
	tx.Outputs = []Output{{Amount: getNumPointer[int64](10)}, {Amount: getNumPointer[int64](10)},
		{Amount: getNumPointer[int64](10)}, {Amount: getNumPointer[int64](10)}}

	require.NoError(t, tx.CalculateTransactionFee())
	require.Zero(t, *tx.Fee)

	// output amount = 50
	tx.Outputs = []Output{{Amount: getNumPointer[int64](10)}, {Amount: getNumPointer[int64](10)},
		{Amount: getNumPointer[int64](10)}, {Amount: getNumPointer[int64](10)}, {Amount: getNumPointer[int64](10)}}
	require.NoError(t, tx.CalculateTransactionFee())
	require.Equal(t, int64(-10), *tx.Fee)

	tx.Outputs = []Output{{Amount: getNumPointer[int64](10)}, {Amount: nil}}
	require.Error(t, tx.CalculateTransactionFee())

	tx.Outputs = []Output{{Amount: getNumPointer[int64](10)}}
	tx.Inputs = []Output{{Amount: getNumPointer[int64](10)}, {Amount: nil}}
	require.Error(t, tx.CalculateTransactionFee())
}

func TestTransaction_IsMixingTransaction(t *testing.T) {
	tx := Transaction{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}
	require.False(t, tx.IsMixingTransaction())

	tx.Type = constants.TypeMixing
	require.True(t, tx.IsMixingTransaction())

	tx.Type = constants.TypeCC
	require.False(t, tx.IsMixingTransaction())
}

func TestTransaction_IsDestinationTransaction(t *testing.T) {
	tx := Transaction{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}
	require.False(t, tx.IsDestinationTransaction())

	tx.Type = constants.TypeMixing
	require.False(t, tx.IsDestinationTransaction())

	tx.Type = constants.TypeDestination
	require.True(t, tx.IsDestinationTransaction())
}

func TestFrontendTransaction_String(t *testing.T) {
	tx := FrontendTransaction{
		BlockHash: "some_long_hex_hash",
		Hash:      "some_long_hex_hash",
	}
	require.NotEmpty(t, tx.String())
}

func TestGetTransactionsOutputs(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)
	// test null input
	outputs, err := GetTransactionsOutputs(dbHandle, nil)
	require.Error(t, err)
	require.Nil(t, outputs)

	// test invalid input
	outputs, err = GetTransactionsOutputs(dbHandle, []string{"some_invalid_transaction_hash"})
	require.Error(t, err)
	require.Nil(t, outputs)

	// 2 transaction should return two mappings
	outputs, err = GetTransactionsOutputs(dbHandle, []string{
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
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)

	// only blocks beginning from height 60000 are in the DB, so it should fail
	transactions, err := GetTransactionsByBlock(dbHandle, 1, 1)
	require.Error(t, err)
	require.Nil(t, transactions)

	transactions, err = GetTransactionsByBlock(dbHandle, 60001, 60001)
	require.NoError(t, err)
	require.Len(t, transactions, 4)
}

func TestGetOutputAddressCounts(t *testing.T) {
	testhelper.SkipIfNoDB(t)
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

	SetupDB(t, dbHandle, testhelper.UseBlockFile)

	transactions, err := GetTransactionsByBlock(dbHandle, 60001, 60001)
	require.NoError(t, err)
	require.Len(t, transactions, 4)

	type testCase struct {
		txhash     string
		uid        string
		numInputs  uint32
		numOutputs uint32
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
		inputCount, outputCount, err = GetOutputAddressCounts(dbHandle, c.uid)
		require.NoError(t, err)
		require.Equal(t, c.numInputs, inputCount, c.txhash)
		require.Equal(t, c.numOutputs, outputCount, c.txhash)
	}
}

func TestGetFrontendTransaction(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)
	const blockHash = "00000000000cfe64fca7b5c3a8ad1ee39dd3f380aeb56027bc25e97904d2c99e"
	const txHash1 = "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13"
	transaction, err := GetFrontendTransaction(context.Background(), dbHandle, txHash1)
	require.NoError(t, err)
	require.Len(t, transaction, 1)
	require.Equal(t, txHash1, transaction[0].Hash)
	require.Equal(t, blockHash, transaction[0].BlockHash)
	require.Equal(t, int64(0), transaction[0].Fee)
	require.Equal(t, "", transaction[0].Type)

	SetupDB(t, dbHandle, testhelper.UseBlockFile)
	const txHash2 = "818dae776566815b8d5307f8597fc8c1db737e933a4605e1841a83f078731638"
	transaction, err = GetFrontendTransaction(context.Background(), dbHandle, txHash2)
	require.NoError(t, err)
	require.Len(t, transaction, 1)
	require.Equal(t, txHash2, transaction[0].Hash)
	require.Equal(t, blockHash, transaction[0].BlockHash)
	require.Equal(t, int64(100000), transaction[0].Fee)
	require.Equal(t, "", transaction[0].Type)
}

func TestGetFrontendTransactionsByUID(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)

	transactions, err := GetTransactionsByBlock(dbHandle, 60005, 60005)
	require.NoError(t, err)
	require.Len(t, transactions, 7)

	var uids [7]string
	for i, tx := range transactions {
		uids[i] = tx.UID
	}

	frontendTransactions, err := GetFrontendTransactionsByUID(context.Background(), dbHandle, uids[:])
	require.NoError(t, err)
	require.Equal(t, len(transactions), len(frontendTransactions))
}

func TestGetTransactionBlockID(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)

	const txHash1 = "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13"
	id, err := GetTransactionBlockID(context.Background(), dbHandle, txHash1)
	require.NoError(t, err)
	require.EqualValues(t, 60001, id)

	const txHash2 = "d0bc5aba5a81df73b706d7819956fb298e03baf52a97c736bb588dfd3586e849"
	id, err = GetTransactionBlockID(context.Background(), dbHandle, txHash2)
	require.NoError(t, err)
	require.EqualValues(t, 60007, id)
}

func TestUpdateTransactions(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)

	// empty slice should fail
	require.Error(t, UpdateTransactions(dbHandle, nil))
	require.Error(t, UpdateTransactions(dbHandle, []Transaction{}))

	transactions, err := GetTransactionsByBlock(dbHandle, 60001, 60001)
	require.NoError(t, err)

	// no mixing transactions should be in this block
	for _, tx := range transactions {
		require.False(t, tx.IsMixingTransaction())
	}

	// set all transactions to be a mixing transaction
	for i := range transactions {
		transactions[i].Type = constants.TypeMixing
	}

	require.NoError(t, UpdateTransactions(dbHandle, transactions))

	transactions, err = GetTransactionsByBlock(dbHandle, 60001, 60001)
	require.NoError(t, err)

	// all transactions should now have the privacy type set to 'mixing'
	for _, tx := range transactions {
		require.True(t, tx.IsMixingTransaction())
	}
}

func TestGetTransactionUID(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)

	ctx := context.Background()
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
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)
	transactions, err := GetOutputs(dbHandle, testhelper.BlockFileFirstBlock, testhelper.BlockFileLastBlock)
	require.NoError(t, err)
	require.Len(t, transactions, 56)

	transactions, err = GetOutputs(dbHandle, 60007, 60007)
	require.NoError(t, err)
	require.Len(t, transactions, 7)

	// should return an empty transaction slice, because this block range is not included in the database
	transactions, err = GetOutputs(dbHandle, 1, 10)
	require.NoError(t, err)
	require.Empty(t, transactions)
}

func TestGetTransaction(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)

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
		gotTransaction, err := GetTransaction(dbHandle, tt.txHash)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, gotTransaction)
		}
	}
}

func TestGetTransactionUIDMapping(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	SetupDB(t, dbHandle, testhelper.UseBlockFile)

	transactions, err := GetTransactionsByBlock(dbHandle, 60005, 60005)
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
		gotTxs, err := GetTransactionUIDMapping(context.Background(), dbHandle, tt.txUids)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Len(t, gotTxs, tt.wantTxCount)
		}
	}
}
