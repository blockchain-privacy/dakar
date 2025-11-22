// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"context"
	"net/http"

	mw "gitlab.com/blockchain-privacy/gomisc/middleware"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

type ContextKeyUser int

const MiddlewareContextUser ContextKeyUser = iota

// sendUnauthorizedMessage sends an unauthorized message
func sendUnauthorizedMessage(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
	w.WriteHeader(http.StatusUnauthorized)
}

// adapt calls mw.Adapt() and inserts an http.TimeoutHandler into the adapter chain
func (s *Server) adapt(h http.Handler, adapters ...mw.Adapter) http.Handler {
	return mw.Adapt(h, append([]mw.Adapter{s.timeout()}, adapters...)...)
}

func Authorization() mw.Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dakarUser := r.Header.Get("x-dakar-user")
			if dakarUser == "" {
				sendUnauthorizedMessage(w)
				warn(serror.FromStr("dgraph UID not set"))
				return
			}

			// call next handler and add to the request context the identity information
			h.ServeHTTP(w,
				r.WithContext(context.WithValue(r.Context(), MiddlewareContextUser, TokenUser{ID: dakarUser})))
		})
	}
}

func (s *Server) timeout() mw.Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.TimeoutHandler(h, s.handlerTimeout, "request timed out").ServeHTTP(w, r)
		})
	}
}
