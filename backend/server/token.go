package server

import (
	"backend/cmd/cliutil"
	dbus "backend/db/user"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/o1egl/paseto"
	"net/http"
	"time"

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
)

// todo rework, for now dummy keys
func getSigningKeys() (ed25519.PrivateKey, ed25519.PublicKey) {
	a, _ := hex.DecodeString("b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a37741eb9dbbbbc047c03fd70604e0071f0987e16b28b757225c11f00415d0e20b1a2")
	b, _ := hex.DecodeString("1eb9dbbbbc047c03fd70604e0071f0987e16b28b757225c11f00415d0e20b1a2")
	return a, b
}

// setTokenAsCookie writes the given token with w as a cookie
func setTokenAsCookie(w http.ResponseWriter, token string, expirationTime time.Time) {
	newCookie := &http.Cookie{
		Name:    cookieTokenName,
		Value:   token,
		Expires: expirationTime,
		// todo check if httponly can be set
		HttpOnly: true,
		Secure:   secureCookie,
		// cookie should be able to be sent by request originating form all directories
		Path: "/",
	}

	http.SetCookie(w, newCookie)
}

func writeNewToken(w http.ResponseWriter, user dbus.User) error {
	newToken, expirationTime, err := issueToken(user)
	if err != nil {
		return err
	}

	setTokenAsCookie(w, newToken, expirationTime)
	return nil
}

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

func issueToken(user dbus.User) (token string, expirationTime time.Time, err error) {
	privateKey, _ := getSigningKeys()

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

func verifyToken(token string) (newJsonToken paseto.JSONToken, newFooter string, err error) {
	_, publicKey := getSigningKeys()

	// Verify data
	err = paseto.NewV2().Verify(token, publicKey, &newJsonToken, &newFooter)
	return
}
