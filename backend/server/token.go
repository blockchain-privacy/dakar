package server

import (
	"encoding/hex"
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
	// secureCookie controls whether the secure and httpOnly attribute in cookies is set
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
		HttpOnly: secureCookie,
		Secure:   secureCookie,
	}

	http.SetCookie(w, newCookie)
}

func writeNewToken(w http.ResponseWriter, userId string) error {
	newToken, expirationTime, err := issueToken(userId)
	if err != nil {
		return err
	}

	setTokenAsCookie(w, newToken, expirationTime)
	return nil
}

func issueToken(userId string) (token string, expirationTime time.Time, err error) {
	privateKey, _ := getSigningKeys()

	expirationTime = time.Now().Add(tokenExpirationTime)
	jsonToken := paseto.JSONToken{
		Expiration: expirationTime,
	}
	jsonToken.Set(tokenFieldUser, userId)

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
