package server

import (
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/ed25519"
)

// tokenExpirationTime is the time a token is valid
const tokenExpirationTime = time.Hour * 24

type Adapter func(http.Handler) http.Handler

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for i := len(adapters) - 1; i >= 0; i-- {
		h = adapters[i](h)
	}
	return h
}

// todo rework, for now dummy keys
func getSigningKeys() (ed25519.PrivateKey, ed25519.PublicKey) {
	a, _ := hex.DecodeString("b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a37741eb9dbbbbc047c03fd70604e0071f0987e16b28b757225c11f00415d0e20b1a2")

	b, _ := hex.DecodeString("1eb9dbbbbc047c03fd70604e0071f0987e16b28b757225c11f00415d0e20b1a2")

	return a, b
}

func authorizationMiddleware() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if false && len(authHeader) != 2 {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Malformed Token"))
			} else {
				jsonToken := paseto.JSONToken{
					Expiration: time.Now().Add(tokenExpirationTime),
				}

				privateKey, publicKey := getSigningKeys()

				// Add custom claim    to the token
				jsonToken.Set("data", "this is a signed message")
				footer := "some footer"

				// Sign data
				paseto2 := paseto.NewV2()
				token, err := paseto2.Sign(privateKey, jsonToken, footer)
				// token = "v2.public.eyJkYXRhIjoidGhpcyBpcyBhIHNpZ25lZCBtZXNzYWdlIiwiZXhwIjoiMjAxOC0wMy0xMlQxOTowODo1NCswMTowMCJ9Ojv0uXlUNXSFhR88KXb568LheLRdeGy2oILR3uyOM_-b7r7i_fX8aljFYUiF-MRr5IRHMBcWPtM0fmn9SOd6Aw.c29tZSBmb290ZXI"

				if err != nil {
					serverInfo(err)
					return
				}

				serverInfo("new token:", token)

				// Verify data
				var newJsonToken paseto.JSONToken
				var newFooter string
				err = paseto2.Verify(token, publicKey, &newJsonToken, &newFooter)
				if err != nil {
					serverInfo(err)
					return
				}

				h.ServeHTTP(w, r)
			}
		})
	}
}
