package password

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomPassword returns a random string if fixed length of 22
func GenerateRandomPassword() (string, error) {
	// Generate a Salt
	pwByte := make([]byte, 16)
	if _, err := rand.Read(pwByte); err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(pwByte), nil
}
