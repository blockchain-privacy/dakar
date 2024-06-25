package server

import (
	"bytes"
	"context"
	mw "github.com/qrest/gomisc/middleware"
	"github.com/qrest/gomisc/serror"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

type contextKeyUser int

const middlewareContextUser contextKeyUser = iota

// sendUnauthorizedMessage sends an unauthorized message
func sendUnauthorizedMessage(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
	w.WriteHeader(http.StatusUnauthorized)
}

func (s *Server) authorization() mw.Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			kratosID := r.Header.Get("x-user")
			if kratosID == "" {
				sendUnauthorizedMessage(w)
				warn(serror.FromStr("kratos ID not set"))
				return
			}

			dakarUser := r.Header.Get("x-dakar-user")
			if dakarUser == "" {
				sendUnauthorizedMessage(w)
				warn(serror.FromStr("dgraph UID not set"))
				return
			}

			// call next handler and add to the request context the identity information
			h.ServeHTTP(w,
				r.WithContext(context.WithValue(r.Context(), middlewareContextUser, tokenUser{
					ID:       dakarUser,
					KratosID: kratosID,
				})))
		})
	}
}

func (s *Server) useCache(ttl time.Duration) mw.Adapter {
	type cacheElement struct {
		buffer     []byte
		header     http.Header
		statusCode int
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// extract body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "an error occurred", http.StatusInternalServerError)
				warn(serror.New(err))
				return
			}
			// reset body, so it can be read by the next handler
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			cacheKey := buildKey(r.RequestURI, body)

			// try to get request from cache
			value, found := s.cache.Get(cacheKey)
			var response cacheElement
			if found {
				response = value.(cacheElement)
			} else {
				// record the writes of the next handler, so the response can be saved in the cache.
				recorder := httptest.NewRecorder()
				// call next handler
				h.ServeHTTP(recorder, r)

				// get recorded values
				resp := recorder.Result() //nolint:bodyclose
				defer func(Body io.ReadCloser) {
					_ = Body.Close()
				}(resp.Body)

				response.statusCode = resp.StatusCode
				response.buffer = recorder.Body.Bytes()
				response.header = resp.Header.Clone()

				// only insert in cache if no error occurred
				if response.statusCode < http.StatusBadRequest {
					s.cache.SetWithTTL(cacheKey, response, 1, ttl)
				}
			}

			// write original header
			for k := range response.header {
				for _, v := range response.header.Values(k) {
					w.Header().Add(k, v)
				}
			}
			// write cache header
			setCacheHeader(w, ttl)
			// write status code
			w.WriteHeader(response.statusCode)

			if _, err = w.Write(response.buffer); err != nil {
				warn(serror.New(err))
			}
		})
	}
}
