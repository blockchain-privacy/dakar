// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"backend/external"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidSortOrder(t *testing.T) {
	require.False(t, IsValidSortOrder(-1))
	require.True(t, IsValidSortOrder(SortAscendingByOutputTime))
	require.True(t, IsValidSortOrder(SortDescendingByOutputTime))
	require.True(t, IsValidSortOrder(SortAscendingByInputTime))
	require.True(t, IsValidSortOrder(SortDescendingByInputTime))
	require.True(t, IsValidSortOrder(SortAscendingByAmount))
	require.True(t, IsValidSortOrder(SortDescendingByAmount))
}

func TestIsValidFilter(t *testing.T) {
	require.False(t, IsValidFilter([]int{-1}))
	require.False(t, IsValidFilter([]int{-1, FilterByCoinbase}))
	require.True(t, IsValidFilter([]int{FilterByCoinbase}))
	require.True(t, IsValidFilter([]int{FilterByUnspent}))
	require.True(t, IsValidFilter([]int{FilterByCoinbase, FilterByUnspent}))
}

func TestAddress_String(t *testing.T) {
	addr := Address{UID: "some_uid", Hash: "some_hash"}

	require.NotEmpty(t, addr.String())

	addr.Outputs = []Output{{}}
	require.NotEmpty(t, addr.String())
}

func TestAddress_SetDType(t *testing.T) {
	addr := Address{
		UID:  "some_uid",
		Hash: "some_hash",
	}

	addr.SetDType()
	require.Equal(t, []string{AddressDType}, addr.DType)
}

func TestFrontendOutput_String(t *testing.T) {
	output := FrontendOutput{Amount: 0}
	require.NotEmpty(t, output.String())
}

func TestFrontendAddress_String(t *testing.T) {
	address := FrontendAddress{
		Hash: "some_hash",
	}

	require.NotEmpty(t, address.String())
}

func TestGetFrontendAddress(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	_, err := GetFrontendAddress(t.Context(), dbHandle, "", 1, 1, nil)
	require.Error(t, err)

	const addrHash = "XsE93qsgqTkzumVTaeanYRXqVz3uXjWpkc"

	addr, err := GetFrontendAddress(t.Context(), dbHandle, addrHash, SortAscendingByAmount, 1, nil)
	require.NoError(t, err)
	require.NotNil(t, addr)
	require.NotEmpty(t, addr.Outputs)

	addr, err = GetFrontendAddress(t.Context(), dbHandle, addrHash, SortAscendingByAmount, 1, []int{FilterByUnspent})
	require.NoError(t, err)
	require.NotNil(t, addr)
	require.Empty(t, addr.Outputs)
}

func TestUpsertAddresses(t *testing.T) {
	dbHandle := GetDBConnection(t, "")

	const newAddressHash = "some_address_hash"

	ctx, cancel := GetTaskContext()
	defer cancel()
	_, err := GetFrontendAddress(ctx, dbHandle, newAddressHash, SortAscendingByAmount, 0, nil)
	require.Error(t, err)

	newAddress := Address{Hash: newAddressHash, Outputs: []Output{{
		OutputIndex: getNumPointer[int32](1),
		InputIndex:  getNumPointer[int32](3),
		TxType:      "",
		Amount:      getNumPointer[int64](500),
		DType:       []string{outputDType},
	}}}
	newAddress.SetDType()

	require.NoError(t, UpsertAddresses(ctx, dbHandle, []Address{newAddress}))

	frontendAddress, err := GetFrontendAddress(ctx, dbHandle, newAddressHash, SortAscendingByAmount, 0, nil)
	require.NoError(t, err)
	require.Equal(t, newAddressHash, frontendAddress.Hash)
}

func TestGetAddressesByBlockRange(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)

	ctx, cancel := GetTaskContext()
	defer cancel()
	blockRange, err := GetAddressesByBlockRange(ctx, dbHandle, BlockFileFirstBlock,
		BlockFileLastBlock, true)
	require.NoError(t, err)
	require.NotEmpty(t, blockRange)

	blockRange, err = GetAddressesByBlockRange(ctx, dbHandle, 1, 500, true)
	require.NoError(t, err)
	require.Empty(t, blockRange)
}

func TestGetAddressUIDs(t *testing.T) {
	dbHandle := GetDBConnection(t, UseBlockFile)
	type args struct {
		c             external.Database
		addressHashes []string
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args:    args{c: nil, addressHashes: nil},
			wantErr: true,
		},
		{
			args: args{
				c: dbHandle,
				addressHashes: []string{
					"XgiLmHQ4czfkGvoqLAQJ8SVMNeho1EiFRv",
					"Xe5GhnraNWanA3fY1XrjC1RnKQZfWmWygh",
					"Xrwhr9kHpnk5CmKLitCcm3aeMv5zNYFZcw",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		gotAddresses, err := GetAddressUIDs(t.Context(), tt.args.c, tt.args.addressHashes)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, gotAddresses)
		}
	}
}
