package user

import (
	"backend/db"
	"backend/external"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle external.Database

func TestMain(m *testing.M) {
	db.RunDgraphTests(m, &dbHandle, db.ContainerNameUser)
}

func TestGenerateRandomPassword(t *testing.T) {
	pw, err := generateRandomPassword()
	require.Nil(t, err)
	require.NotEmpty(t, pw, "password is empty")
	require.EqualValues(t, len(pw), 22, "got random password with wrong size:")

	pw2, err := generateRandomPassword()
	require.Nil(t, err)
	require.NotEmpty(t, pw, "password is empty")

	require.NotEqual(t, pw, pw2)
}

func TestCreateNewUser(t *testing.T) {
	user, err := CreateNewUser(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)
}

func TestGetUsers(t *testing.T) {
	db.SetupDBWithoutData(t, dbHandle)

	user, err := CreateNewUser(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	users, err := GetUsers(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, users)
}

func TestGetUsersWithCredentials(t *testing.T) {
	db.SetupDBWithoutData(t, dbHandle)

	user, err := CreateNewUser(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	users, err := GetUsersWithCredentials(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, users)
}

func TestDeleteUser(t *testing.T) {
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
