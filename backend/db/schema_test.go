package db

import (
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetupSchema(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	require.NoError(t, SetupSchema(dbHandle))
}

func TestIsSchemaSet(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	// first drop schema
	require.NoError(t, DropAll(dbHandle))

	// should return false because schema is not set
	isSet, err := IsSchemaSet(dbHandle)
	require.NoError(t, err)
	require.False(t, isSet)

	// set up schema
	require.NoError(t, SetupSchema(dbHandle))

	// should return true because schema is now set
	isSet, err = IsSchemaSet(dbHandle)
	require.NoError(t, err)
	require.True(t, isSet)
}
