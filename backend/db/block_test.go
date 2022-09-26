package db

import (
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBlock_String(t *testing.T) {
	blk := Block{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}
	require.NotEmpty(t, blk.String())
}

func TestBlock_SetDType(t *testing.T) {
	blk := Block{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}

	blk.SetDType()

	require.Equal(t, []string{blockDType}, blk.DType)
}

func TestBlock_IsComplete(t *testing.T) {
	blk := Block{
		UID:  "some_uid",
		Hash: "some_long_hex_hash",
	}

	require.False(t, blk.IsComplete())

	blk.ID = getUInt64Pointer(100)
	blk.Timestamp = time.Now().Format(time.RFC3339)
	blk.SetDType()
	blk.Transactions = []Transaction{{}}
	blk.PrevBlock = &Block{}
	require.True(t, blk.IsComplete())
}

func TestFrontendBlock_String(t *testing.T) {
	blk := FrontendBlock{
		Hash: "some_long_hex_hash",
	}
	require.NotEmpty(t, blk.String())
}

func TestBlockQueryPayload(t *testing.T) {
	q := blockQuery{Q: []Block{{}}}

	_, err := q.payload()
	require.NoError(t, err)

	q = blockQuery{}
	_, err = q.payload()
	require.Error(t, err)

	q = blockQuery{Q: []Block{{}, {}}}
	_, err = q.payload()
	require.Error(t, err)
}

func TestGetBlock(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	SetupDB(t, dbHandle, blockFileName)

	const blockHash = "0000000000422dc68b3a1ab79869010747851e53d7732a10477759271da13caa"
	const blockHash2 = "000000000020ef46c4026cb77aee959224b9bb0a23b24bc46c429d8c9cab498b"

	block, err := GetBlock(dbHandle, blockHash)
	require.NoError(t, err)
	require.Len(t, block.Transactions, 2)

	block, err = GetBlock(dbHandle, blockHash2)
	require.NoError(t, err)
	require.Len(t, block.Transactions, 1)

	_, err = GetBlock(dbHandle, "")
	require.Error(t, err)
}
