package server

import (
	"backend/cmd/cliutil"
	"backend/user"
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
func writeUnauthorized(w http.ResponseWriter, msg string) {
	if len(msg) == 0 {
		msg = "Malformed Token"
	}

	w.WriteHeader(http.StatusUnauthorized)
	if _, writeErr := w.Write([]byte(msg)); writeErr != nil {
		serverInfo(writeErr)
	}
}

// sendRedirectMessage sends a redirect message
func sendRedirectMessage(w http.ResponseWriter) {
	setDefaultHeader(w)
	if _, err := w.Write([]byte(`{"invalidToken": true}`)); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
		serverInfo(cliutil.ShowCallInfo(), err)
	}
}

func authorizationMiddleware(route string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieTokenName)
			if err != nil {
				sendRedirectMessage(w)
				return
			}

			token, _, verifyErr := verifyToken(cookie.Value)
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
				serverInfo(cliutil.ShowCallInfo(), "user id field not found in token")
				return
			}

			var newUser tokenUser
			if jsonErr := json.Unmarshal([]byte(userFromToken), &newUser); jsonErr != nil {
				writeUnauthorized(w, "")
				serverInfo(cliutil.ShowCallInfo(), "user id field not found in token")
				return
			}

			// check if route is allowed
			routeAllowed := false
			for _, uRole := range newUser.Roles {
				routeRole, roleErr := user.GetRoleByName(uRole.Name)
				if roleErr != nil {
					writeUnauthorized(w, "")
					serverInfo(cliutil.ShowCallInfo(), roleErr)
					return
				}

				if routeRole.IsRouteAllowed(route) {
					routeAllowed = true
					break
				}
			}

			if !routeAllowed {
				writeUnauthorized(w, "route not allowed")
				serverInfo(cliutil.ShowCallInfo(), newUser.Id, "tried to access", route)
				return
			}

			if timeUntilTokenExpires <= reissueDuration {
				if tokenErr := writeNewToken(w, newUser.toUser().ToFrontendUserState()); tokenErr != nil {
					sendRedirectMessage(w)
					serverInfo(cliutil.ShowCallInfo(), tokenErr)
					return
				}
			}

			h.ServeHTTP(w, r)
		})
	}
}
