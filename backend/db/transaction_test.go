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

	// 2 transaction should return two mappings
	outputs, err = GetTransactionsOutputs(dbHandle, []string{
		"91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb",
		"0cfd028caf97751603255b1467085c3ccc5d476d79810ba9608d63587c7986f8",
	})
	require.Error(t, err)
	require.Equal(t, 2, len(outputs))

	var outputCount int
	for _, mapping := range outputs {
		outputCount += len(mapping.Outputs)
	}

	require.Equal(t, 3, outputCount, "unexpected number of outputs")
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
	require.Equal(t, 4, len(transactions))
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

	SetupDB(t, dbHandle, blockFileName)

	transactions, err := GetTransactionByBlock(dbHandle, 60001)
	require.NoError(t, err)
	require.Equal(t, 4, len(transactions))

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
		require.Equal(t, c.numInputs, inputCount)
		require.Equal(t, c.numOutputs, outputCount)
	}
}

func TestGetFrontendTransaction(t *testing.T) {
	SetupDB(t, dbHandle, blockFileName)
	const blockHash = "00000000000cfe64fca7b5c3a8ad1ee39dd3f380aeb56027bc25e97904d2c99e"
	const txHash1 = "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13"
	transaction, err := GetFrontendTransaction(dbHandle, txHash1)
	require.NoError(t, err)
	require.Equal(t, 1, len(transaction))
	require.Equal(t, txHash1, transaction[0].Hash)
	require.Equal(t, blockHash, transaction[0].BlockHash)
	require.Equal(t, int64(0), transaction[0].Fee)
	require.Equal(t, int64(-1), transaction[0].PrivacyType)

	SetupDB(t, dbHandle, blockFileName)
	const txHash2 = "818dae776566815b8d5307f8597fc8c1db737e933a4605e1841a83f078731638"
	transaction, err = GetFrontendTransaction(dbHandle, txHash2)
	require.NoError(t, err)
	require.Equal(t, 1, len(transaction))
	require.Equal(t, txHash2, transaction[0].Hash)
	require.Equal(t, blockHash, transaction[0].BlockHash)
	require.Equal(t, int64(100000), transaction[0].Fee)
	require.Equal(t, int64(-1), transaction[0].PrivacyType)
}

func TestGetFrontendTransactionsByUID(t *testing.T) {
	SetupDB(t, dbHandle, blockFileName)

	transactions, err := GetTransactionByBlock(dbHandle, 60005)
	require.NoError(t, err)
	require.Equal(t, 7, len(transactions))

	var uids [7]string
	for i, tx := range transactions {
		uids[i] = tx.UID
	}

	frontendTransactions, err := GetFrontendTransactionsByUID(dbHandle, uids[:])
	require.NoError(t, err)
	require.Equal(t, len(transactions), len(frontendTransactions))
}

func TestGetTransactionBlockID(t *testing.T) {
	SetupDB(t, dbHandle, blockFileName)

	const txHash1 = "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13"
	id, err := GetTransactionBlockID(dbHandle, txHash1)
	require.NoError(t, err)
	require.EqualValues(t, 60001, id)

	const txHash2 = "d0bc5aba5a81df73b706d7819956fb298e03baf52a97c736bb588dfd3586e849"
	id, err = GetTransactionBlockID(dbHandle, txHash2)
	require.NoError(t, err)
	require.EqualValues(t, 60007, id)
}

func TestGetTransactionUID(t *testing.T) {
	SetupDB(t, dbHandle, blockFileName)
	const txHash1 = "a9535110536ded94998287e306b9a0c7d9e6b3a7ad88c7e82a60a0515ccc1f13"
	uid, err := GetTransactionUID(dbHandle, txHash1)
	require.NoError(t, err)
	require.NotEmpty(t, uid)
}

func TestGetOutputs(t *testing.T) {
	SetupDB(t, dbHandle, blockFileName)
	transactions, err := GetOutputs(dbHandle, 60000, 60020)
	require.NoError(t, err)
	require.Equal(t, 50, len(transactions))

	transactions, err = GetOutputs(dbHandle, 60007, 60007)
	require.NoError(t, err)
	require.Equal(t, 7, len(transactions))
}
