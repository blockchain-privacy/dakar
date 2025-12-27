// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

	blk.ID = getNumPointer[int64](100)
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
	dbHandle := GetDBConnection(t, UseBlockFile)

	const blockHash = "0000000000422dc68b3a1ab79869010747851e53d7732a10477759271da13caa"
	const blockHash2 = "000000000020ef46c4026cb77aee959224b9bb0a23b24bc46c429d8c9cab498b"

	ctx, cancel := GetTaskContext()
	defer cancel()

	block, err := GetBlock(ctx, dbHandle, blockHash)
	require.NoError(t, err)
	require.Len(t, block.Transactions, 2)

	block, err = GetBlock(ctx, dbHandle, blockHash2)
	require.NoError(t, err)
	require.Len(t, block.Transactions, 1)

	_, err = GetBlock(ctx, dbHandle, "")
	require.Error(t, err)
}

func TestGetFullBlock(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)
	ctx, cancel := GetTaskContext()
	defer cancel()
	block, err := GetFullBlock(ctx, dbHandle, 60003, true)
	require.NoError(t, err)
	require.Len(t, block.Transactions, 2)

	block, err = GetFullBlock(ctx, dbHandle, 60004, true)
	require.NoError(t, err)
	require.Len(t, block.Transactions, 1)

	// not in block file
	_, err = GetFullBlock(ctx, dbHandle, 3, true)
	require.Error(t, err)
}

func TestGetFrontendBlock(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	const blockHash = "0000000000422dc68b3a1ab79869010747851e53d7732a10477759271da13caa"
	const blockHash2 = "000000000020ef46c4026cb77aee959224b9bb0a23b24bc46c429d8c9cab498b"

	block, err := GetFrontendBlock(t.Context(), dbHandle, blockHash, 0)
	require.NoError(t, err)
	require.Len(t, block.Transactions, 2)

	block, err = GetFrontendBlock(t.Context(), dbHandle, blockHash2, 0)
	require.NoError(t, err)
	require.Len(t, block.Transactions, 1)

	_, err = GetFrontendBlock(t.Context(), dbHandle, "", 0)
	require.Error(t, err)
}

func TestUpsertBlock(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)
	ctx, cancel := GetTaskContext()
	defer cancel()

	transactions, err := GetTransactionsByBlock(ctx, dbHandle, BlockFileFirstBlock, BlockFileFirstBlock, nil)
	require.NoError(t, err)

	tests := []struct {
		block   Block
		wantErr bool
	}{
		{
			block: Block{
				Transactions: transactions,
				Hash:         "some_hash",
				ID:           getNumPointer[int64](5),
				Timestamp:    time.Now().Format(time.RFC3339),
				PrevBlock: &Block{
					Hash: "some_other_hash",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := UpsertBlock(ctx, dbHandle, tt.block)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
