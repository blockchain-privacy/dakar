package user

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateRandomPassword(t *testing.T) {
	pw, err := GenerateRandomPassword()
	require.Nil(t, err)
	require.NotEmpty(t, pw, "password is empty")
	require.EqualValues(t, len(pw), 22, "got random password with wrong size:")
}

func TestGetRandomPasswordAndHash(t *testing.T) {
	pw, pwHash, err := GetRandomPasswordAndHash()

	require.Nil(t, err)
	require.NotEmpty(t, pw, "password is empty")
	require.NotEmpty(t, pwHash, "password hash is empty")
	require.NotEqualValues(t, pw, pwHash)
}

func TestComparePassword(t *testing.T) {
	pw, pwHash, err := GetRandomPasswordAndHash()
	require.Nil(t, err)

	ok, err := ComparePassword(pw, pwHash)
	require.Nil(t, err)
	require.True(t, ok, "did not get a positive result for test. pw:", pw, "hash:", pwHash)

	pw = "some_password"
	pwHash = "some_hash"

	ok, err = ComparePassword(pw, pwHash)

	require.False(t, err == nil || ok, "dummy password and hash should have failed comparison", pw, "with hash", pwHash)
}
