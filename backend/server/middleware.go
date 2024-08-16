package server

import (
	"context"
	mw "github.com/qrest/gomisc/middleware"
	"github.com/qrest/gomisc/serror"
	"net/http"
)

type contextKeyUser int

const middlewareContextUser contextKeyUser = iota

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

func (s *Server) authorization() mw.Adapter {
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
				r.WithContext(context.WithValue(r.Context(), middlewareContextUser, tokenUser{ID: dakarUser})))
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
