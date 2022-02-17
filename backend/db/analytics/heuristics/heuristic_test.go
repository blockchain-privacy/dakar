package heuristics

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateKeyHash(t *testing.T) {
	hash1 := createKeyHash(map[string]bool{"a1111": true, "a2": true, "a33": true})
	require.NotEmpty(t, hash1)

	hash2 := createKeyHash(map[string]bool{"a2": true, "a1111": true, "a33": true})
	require.NotEmpty(t, hash2)
	require.EqualValues(t, hash1, hash2, "same input elements should create same hash, no matter the order")

	hash3 := createKeyHash(map[string]bool{"a1111": true, "a2": true})
	require.NotEmpty(t, hash3)
	require.NotEqual(t, hash3, hash2, "fewer input elements should create a different hash")

	hash4 := createKeyHash(nil)
	require.Empty(t, hash4)

	hash5 := createKeyHash(map[string]bool{})
	require.Empty(t, hash5)
}
