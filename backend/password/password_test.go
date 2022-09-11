package password

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
