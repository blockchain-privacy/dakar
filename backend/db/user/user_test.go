package user

import (
	"backend/db"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	testhelper.RunDgraphTests(m, &dbHandle.DB)
}

func TestGenerateRandomPassword(t *testing.T) {
	const numPasswords = 10000
	pwMap := make(map[string]bool, numPasswords)
	for i := 0; i < numPasswords; i++ {
		pw, err := generateRandomPassword()
		require.NoError(t, err)
		require.NotEmpty(t, pw, "password is empty")
		require.EqualValues(t, len(pw), 22, "got random password with wrong size:")
		pwMap[pw] = true
	}

	// all generated password should be unique
	require.Len(t, pwMap, numPasswords)
}

func TestCreateNewUser(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	user, err := CreateNewUser(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)
}

func TestGetUsers(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)

	user, err := CreateNewUser(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	users, err := GetUsers(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, users)
}

func TestGetUsersWithCredentials(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)

	user, err := CreateNewUser(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	users, err := GetUsersWithCredentials(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, users)
}

func TestDeleteUser(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)
	// create user
	user, err := CreateNewUser(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	// delete user
	require.NoError(t, DeleteUser(dbHandle, user))

	// try to delete user which does not exist
	require.Error(t, DeleteUser(dbHandle, "some_random_uid_which_does_not_exist"))
}
