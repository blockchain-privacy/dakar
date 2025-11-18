// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package user

import (
	"backend/db"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateNewUser(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	user, err := CreateNewUser(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)
}

func TestDeleteUser(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// create user
	user, err := CreateNewUser(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	// delete user
	require.NoError(t, DeleteUser(t.Context(), dbHandle, user))

	// try to delete user which does not exist
	require.ErrorIs(t, DeleteUser(t.Context(), dbHandle, "0x9999999"), ErrUserDoesNotExist)
}
