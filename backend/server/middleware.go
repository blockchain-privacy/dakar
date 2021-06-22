package server

import (
	"backend/cmd/cliutil"
	"backend/user"

	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/crypto/ed25519"
)

type contextKeyUser int

const middlewareContextUser contextKeyUser = iota

type adapter func(http.Handler, string) http.Handler

func adapt(h http.Handler, route string, adapters ...adapter) http.Handler {
	for i := len(adapters) - 1; i >= 0; i-- {
		h = adapters[i](h, route)
	}
	return h
}

// writeUnauthorized sets the http.StatusUnauthorized status code and writes an error message
func writeUnauthorized(w http.ResponseWriter, msg string) {
	if len(msg) == 0 {
		msg = "Malformed Token"
	}

	w.WriteHeader(http.StatusUnauthorized)
	if _, writeErr := w.Write([]byte(msg)); writeErr != nil {
		info(writeErr)
	}
}

// sendRedirectMessage sends a redirect message
func sendRedirectMessage(w http.ResponseWriter) {
	setDefaultHeader(w)
	if _, err := w.Write([]byte(`{"invalidToken": true}`)); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
		info(cliutil.ShowCallInfo(), err)
	}
}

func authorizationMiddleware(privkey ed25519.PrivateKey, pubkey ed25519.PublicKey) adapter {
	return func(h http.Handler, route string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieTokenName)
			if err != nil {
				sendRedirectMessage(w)
				return
			}

			token, _, verifyErr := verifyToken(cookie.Value, pubkey)
			if verifyErr != nil {
				sendRedirectMessage(w)
				return
			}

			timeUntilTokenExpires := time.Until(token.Expiration)
			if timeUntilTokenExpires <= 0 {
				sendRedirectMessage(w)
				return
			}

			userFromToken := token.Get(tokenFieldUser)
			if len(userFromToken) == 0 {
				sendRedirectMessage(w)
				info(cliutil.ShowCallInfo(), "user id field not found in token")
				return
			}

			var newUser tokenUser
			if jsonErr := json.Unmarshal([]byte(userFromToken), &newUser); jsonErr != nil {
				writeUnauthorized(w, "")
				info(cliutil.ShowCallInfo(), "user id field not found in token")
				return
			}

			// check if route is allowed
			routeAllowed := false
			for _, uRole := range newUser.Roles {
				routeRole, roleErr := user.GetRoleByName(uRole.Name)
				if roleErr != nil {
					writeUnauthorized(w, "")
					info(cliutil.ShowCallInfo(), roleErr)
					return
				}

				if routeRole.IsRouteAllowed(route) {
					routeAllowed = true
					break
				}
			}

			if !routeAllowed {
				writeUnauthorized(w, "route not allowed")
				info(cliutil.ShowCallInfo(), newUser.ID, "tried to access", route)
				return
			}

			if timeUntilTokenExpires <= reissueDuration {
				if tokenErr := writeNewToken(w, newUser.toUser().ToFrontendUserState(), privkey); tokenErr != nil {
					sendRedirectMessage(w)
					info(cliutil.ShowCallInfo(), tokenErr)
					return
				}
			}
			// call next handler and add to the request context the user information
			h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), middlewareContextUser, newUser)))
		})
	}
}

type cacheElement struct {
	buffer     []byte
	statusCode int
}

func cacheMiddleware(cache *ristretto.Cache, ttl time.Duration) adapter {
	return func(h http.Handler, route string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// set headers
			setDefaultHeader(w)

			// extract body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				handleError(w, err)
				return
			}
			// reset body so it can be read by the next handler
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			query := r.URL.Path[len(route):]
			cacheKey := buildKey(route, query, body)

			// try to get request from cache
			value, found := cache.Get(cacheKey)
			var buf []byte
			var httpStatusCode int
			if found {
				foundCache := value.(cacheElement)

				httpStatusCode = foundCache.statusCode
				buf = foundCache.buffer

			} else {
				// record the writes of the next handler, so the response can be saved in the cache.
				recorder := httptest.NewRecorder()
				// call next handler
				h.ServeHTTP(recorder, r)

				// get recorded values
				httpStatusCode = recorder.Result().StatusCode
				buf = recorder.Body.Bytes()

				// create new cache element
				ce := cacheElement{
					buffer:     buf,
					statusCode: httpStatusCode,
				}

				// only insert in cache if no error occurred
				if httpStatusCode < http.StatusBadRequest {
					cache.SetWithTTL(cacheKey, ce, 1, ttl)
				}
			}

			setCacheHeader(w, ttl)

			w.WriteHeader(httpStatusCode)
			_, err = w.Write(buf)
			if err != nil {
				handleError(w, err)
			}
		})
	}
}

func basicAuthMiddleware(u, pwhash string) adapter {
	return func(h http.Handler, route string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// set headers
			setDefaultHeader(w)
			w.Header().Set("WWW-Authenticate", `Basic realm="Dakar Metrics"`)

			requestUser, requestPassword, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// constant time compare and sleep to avoid timing attacks
			if subtle.ConstantTimeCompare([]byte(u), []byte(requestUser)) != 1 {
				time.Sleep(time.Second * time.Duration(1+rand.Intn(5)))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// constant time compare and sleep to avoid timing attacks
			if equal, err := user.ComparePassword(requestPassword, pwhash); err != nil {
				info(cliutil.ShowCallInfo(), err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			} else if !equal {
				time.Sleep(time.Second * time.Duration(1+rand.Intn(5)))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
