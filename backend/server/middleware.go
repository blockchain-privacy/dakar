package server

import (
	"backend/cmd/cliutil"
	"encoding/json"
	"net/http"
	"time"
)

type Adapter func(http.Handler) http.Handler

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for i := len(adapters) - 1; i >= 0; i-- {
		h = adapters[i](h)
	}
	return h
}

// writeUnauthorized sets the http.StatusUnauthorized status code and writes an error message
func writeUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	if _, writeErr := w.Write([]byte("Malformed Token")); writeErr != nil {
		serverInfo(writeErr)
	}
}

func authorizationMiddleware() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverInfo(r.Cookies())

			if cookie, err := r.Cookie(cookieTokenName); err != nil {
				writeUnauthorized(w)
				return
			} else {
				token, _, verifyErr := verifyToken(cookie.Value)
				if verifyErr != nil {
					writeUnauthorized(w)
					return
				}

				timeUntilTokenExpires := time.Until(token.Expiration)
				// todo instead of returning an error message, return a status that the user should
				// be redirected to the login page
				if timeUntilTokenExpires <= 0 {
					writeUnauthorized(w)
					return
				}

				if timeUntilTokenExpires <= reissueDuration {
					userFromToken := token.Get(tokenFieldUser)
					if len(userFromToken) == 0 {
						serverInfo(cliutil.ShowCallInfo(), "user id field not found in token")
						return
					}

					var newUser tokenUser
					if jsonErr := json.Unmarshal([]byte(userFromToken), &newUser); jsonErr != nil {
						serverInfo(cliutil.ShowCallInfo(), "user id field not found in token")
						return
					}

					if tokenErr := writeNewToken(w, newUser.toUser()); tokenErr != nil {
						serverInfo(cliutil.ShowCallInfo(), tokenErr)
						return
					}
				}

				h.ServeHTTP(w, r)
			}
		})
	}
}
