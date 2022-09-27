package db

import (
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
