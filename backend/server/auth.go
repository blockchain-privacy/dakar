package server

import (
	"errors"
	"os"
)

const (
	// BasicAuthUserEnvironmentField is the name of the os environment field for the basic auth user
	BasicAuthUserEnvironmentField = "BASIC_AUTH_USER"
	// BasicAuthPasswordHashEnvironmentField is the name of the os environment field for the basic auth password hash
	BasicAuthPasswordHashEnvironmentField = "BASIC_AUTH_PW_HASH"
)

// GetBasicAuthCredentialsFromEnv returns the user and password hash for basic auth
func GetBasicAuthCredentialsFromEnv() (string, string, error) {
	user := os.Getenv(BasicAuthUserEnvironmentField)
	if len(user) == 0 {
		return "", "", errors.New("basic auth user environment variable not set")
	}

	passwordHash := os.Getenv(BasicAuthPasswordHashEnvironmentField)
	if len(passwordHash) == 0 {
		return "", "", errors.New("basic auth password hash environment variable not set")
	}

	return user, passwordHash, nil
}
