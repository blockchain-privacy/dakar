package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupSchema(t *testing.T) {
	SkipIfNoDB(t)
	dbHandle := GetDBConnection("")
	require.NoError(t, SetupSchema(dbHandle))
}

func TestIsSchemaSet(t *testing.T) {
	SkipIfNoDB(t)
	dbHandle := GetBareDBConnection()
	ctx, cancel := GetTaskContext()
	defer cancel()

	// first drop schema
	require.NoError(t, dbHandle.DropData(ctx))

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
