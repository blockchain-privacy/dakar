package db

import (
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
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
	testhelper.SkipIfNotCI(t)
	SetupDB(t, dbHandle, blockFileName)

	_, err := GetFrontendAddress(dbHandle, "", 1, 1, nil)
	require.Error(t, err)

	const addrHash = "XsE93qsgqTkzumVTaeanYRXqVz3uXjWpkc"

	addr, err := GetFrontendAddress(dbHandle, addrHash, SortAscendingByAmount, 1, nil)
	require.NoError(t, err)
	require.NotEmpty(t, addr.Outputs)

	addr, err = GetFrontendAddress(dbHandle, addrHash, SortAscendingByAmount, 1, []int{FilterByUnspent})
	require.NoError(t, err)
	require.Empty(t, addr.Outputs)
}

func TestUpsertAddresses(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	SetupDBWithoutData(t, dbHandle)

	const newAddressHash = "some_address_hash"

	_, err := GetFrontendAddress(dbHandle, newAddressHash, SortAscendingByAmount, 0, nil)
	require.Error(t, err)

	newAddress := Address{Hash: newAddressHash, Outputs: []Output{{
		OutputIndex: getNumPointer[uint32](1),
		InputIndex:  getNumPointer[uint32](3),
		TxType:      "",
		Amount:      getNumPointer[int64](500),
		DType:       []string{outputDType},
	}}}
	newAddress.SetDType()

	require.NoError(t, UpsertAddresses(dbHandle, []Address{newAddress}))

	frontendAddress, err := GetFrontendAddress(dbHandle, newAddressHash, SortAscendingByAmount, 0, nil)
	require.NoError(t, err)
	require.Equal(t, newAddressHash, frontendAddress.Hash)
}
