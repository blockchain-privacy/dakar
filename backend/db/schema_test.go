package db

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetupSchema(t *testing.T) {
	require.NoError(t, SetupSchema(h))
}

func TestIsSchemaSet(t *testing.T) {
	// first drop schema
	require.NoError(t, DropAll(h))

	// should return false because schema is not set
	isSet, err := IsSchemaSet(h)
	require.NoError(t, err)
	require.False(t, isSet)

	// set up schema
	require.NoError(t, SetupSchema(h))

	// should return true because schema is now set
	isSet, err = IsSchemaSet(h)
	require.NoError(t, err)
	require.True(t, isSet)
}
