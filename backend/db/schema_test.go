// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupSchema(t *testing.T) {
	dbHandle := GetDBConnection(t, "")

	require.NoError(t, WithRetry(func() error { return SetupSchema(dbHandle) }, retrySleepDuration))
}

func TestIsSchemaSet(t *testing.T) {
	dbHandle := GetBareDBConnection(t)
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
