package server

import (
	"backend/cmd/cliutil"
	dbus "backend/db/user"

	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/ed25519"
)

const (
	// tokenExpirationTime is the time a token is valid
	tokenExpirationTime = time.Hour * 24
	// cookieTokenName is the name of the cookie where the token is saved
	cookieTokenName = "token"
	// secureCookie controls whether the secure attribute in cookies is set
	secureCookie = false
	// if only reissueDuration is left of the token lifetime it gets reissued
	reissueDuration = tokenExpirationTime / 8
	// tokenFieldUser is the name of the user field in the token
	tokenFieldUser = "user_id"
	// SigningPubkeyEnvironmentField is the name of the os environment field for the public signing key
	SigningPubkeyEnvironmentField = "TOKEN_PUB_KEY"
	// SigningPrivkeyEnvironmentField is the name of the os environment field for the private signing key
	SigningPrivkeyEnvironmentField = "TOKEN_PRIV_KEY"
)

type tokenUser struct {
	Id    string      `json:"uid,omitempty"`
	Roles []dbus.Role `json:"roles,omitempty"`
}

// toUser creates a new dbus.User and fill it with data from t
func (t tokenUser) toUser() dbus.User {
	return dbus.User{
		Uid:   t.Id,
		Roles: t.Roles,
	}
}

// GetSigningKeysFromEnv returns a public key pair, an error is returned if
// SigningPubkeyEnvironmentField or SigningPrivkeyEnvironmentField are not set
func GetSigningKeysFromEnv() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	pubKeyEncoded := os.Getenv(SigningPubkeyEnvironmentField)

	if len(pubKeyEncoded) == 0 {
		return nil, nil, errors.New("public key environment variable not set")
	}

	privkeyEncoded := os.Getenv(SigningPrivkeyEnvironmentField)
	if len(privkeyEncoded) == 0 {
		return nil, nil, errors.New("private key environment variable not set")
	}

	privkey, err := hex.DecodeString(privkeyEncoded)
	if err != nil {
		return nil, nil, err
	}

	pubkey, err := hex.DecodeString(pubKeyEncoded)
	if err != nil {
		return nil, nil, err
	}

	return privkey, pubkey, nil
}

// setTokenAsCookie writes the given token with w as a cookie
func setTokenAsCookie(w http.ResponseWriter, token string, expirationTime time.Time) {
	newCookie := &http.Cookie{
		Name:     cookieTokenName,
		Value:    token,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   secureCookie,
		// cookie should be able to be sent by request originating form all directories
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, newCookie)
}

// writeNewToken writes the data from user into as a cookie to w
func writeNewToken(w http.ResponseWriter, user dbus.FrontendUserState, privkey ed25519.PrivateKey) error {
	newToken, expirationTime, err := issueToken(user, privkey)
	if err != nil {
		return err
	}

	setTokenAsCookie(w, newToken, expirationTime)
	return nil
}

// invalidateToken invalidates the token
func invalidateToken(w http.ResponseWriter) {
	setTokenAsCookie(w, "", time.Now().Add(-100*time.Hour))
}

// issueToken creates a token from user
func issueToken(user dbus.FrontendUserState, privateKey ed25519.PrivateKey) (token string, expirationTime time.Time, err error) {
	newTokenUser := tokenUser{
		Id:    user.Uid,
		Roles: user.Roles,
	}

	jsonUser, jsonErr := json.Marshal(&newTokenUser)
	if jsonErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), jsonErr)
		return
	}
	expirationTime = time.Now().Add(tokenExpirationTime)
	jsonToken := paseto.JSONToken{
		Expiration: expirationTime,
	}
	jsonToken.Set(tokenFieldUser, string(jsonUser))

	// Sign data
	token, err = paseto.NewV2().Sign(privateKey, jsonToken, nil)
	return
}

// verifyToken checks if token is valid
func verifyToken(token string, publicKey ed25519.PublicKey) (newJsonToken paseto.JSONToken, newFooter string, err error) {
	// Verify data
	err = paseto.NewV2().Verify(token, publicKey, &newJsonToken, &newFooter)
	return
}
