package user

import (
	"backend/db"
	"backend/testhelper"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	db.InitLogger()
	testhelper.RunDgraphTests(m, &dbHandle.DB)
}

func TestCreateNewUser(t *testing.T) {
	testhelper.SkipIfNoDB(t)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	user, err := CreateNewUser(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)
}

func TestDeleteUser(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// create user
	user, err := CreateNewUser(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	// delete user
	require.NoError(t, DeleteUser(context.Background(), dbHandle, user))

	// try to delete user which does not exist
	require.Error(t, DeleteUser(context.Background(), dbHandle, "0x9999999"))
}
