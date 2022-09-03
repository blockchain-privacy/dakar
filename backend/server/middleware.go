package server

import (
	"backend/cmd/cliutil"
	dbus "backend/db/user"
	"backend/password"
	"errors"

	"bytes"
	"context"
	"crypto/subtle"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
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

// extractRoles tries to extract roles from the given metadata
func extractRoles(metaDataPublic any) ([]dbus.Role, error) {
	metadata, ok := metaDataPublic.(map[string]any)
	if !ok {
		return nil, errors.New("identity has no public metadata")
	}

	rolesInterface, ok := metadata["roles"]
	if !ok {
		return nil, errors.New("identity has no field 'roles'")
	}

	roleInterfaces, ok := rolesInterface.([]any)
	if !ok {
		return nil, errors.New("roles could not be cast from interface")
	}

	var roles []dbus.Role

	for _, r := range roleInterfaces {
		roleString, ok := r.(string)
		if ok {
			roles = append(roles, dbus.Role{
				Name: roleString,
			})
		}
	}

	return roles, nil
}

// extractDgraphUID tries to extract dgraph UID from the given metadata
func extractDgraphUID(metadataPublic any) (string, error) {
	metadata, ok := metadataPublic.(map[string]any)
	if !ok {
		return "", errors.New("identity has no admin metadata")
	}

	dgraphUIDInterface, ok := metadata["dgraph_uid"]
	if !ok {
		return "", errors.New("identity has no field 'dgraph_uid'")
	}

	dgraphUID, ok := dgraphUIDInterface.(string)
	if !ok {
		return "", errors.New("dgraph UID could not be cast from interface")
	}

	return dgraphUID, nil
}

func (s *Server) authorization() adapter {
	return func(h http.Handler, route string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _, err := s.auth.V0alpha2Api.ToSession(r.Context()).Cookie(r.Header.Get("Cookie")).Execute()
			if err != nil {
				sendRedirectMessage(w)
				info(cliutil.ShowCallInfo(), err)
				return
			}

			// check if session active and not expired
			if session.Active == nil || session.ExpiresAt == nil ||
				!*session.Active || time.Until(*session.ExpiresAt) <= 0 {
				sendRedirectMessage(w)
				return
			}

			if time.Until(*session.ExpiresAt) <= reissueDuration {
				if _, _, extensionErr := s.adminAuth.V0alpha2Api.AdminExtendSession(r.Context(), session.Id).Execute(); extensionErr != nil {
					sendRedirectMessage(w)
					info(cliutil.ShowCallInfo(), extensionErr)
					return
				}
			}

			dgraphUID, err := extractDgraphUID(session.Identity.MetadataPublic)
			if err != nil {
				sendRedirectMessage(w)
				info(cliutil.ShowCallInfo(), err)
				return
			}

			roles, err := extractRoles(session.Identity.MetadataPublic)
			if err != nil {
				sendRedirectMessage(w)
				info(cliutil.ShowCallInfo(), err)
				return
			}

			// check if route is allowed and get typed role
			routeAllowed := false
			for _, role := range roles {
				routeRole, roleErr := getRoleByName(role.Name)
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
				info(cliutil.ShowCallInfo(), session, "tried to access", route)
				return
			}

			// call next handler and add to the request context the identity information
			h.ServeHTTP(w,
				r.WithContext(context.WithValue(r.Context(), middlewareContextUser, tokenUser{
					ID:       dgraphUID,
					KratosID: session.Identity.Id,
					Roles:    roles,
				})))
		})
	}
}

func (s *Server) useCache(ttl time.Duration) adapter {
	type cacheElement struct {
		buffer     []byte
		statusCode int
	}
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
			// reset body, so it can be read by the next handler
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			query := r.URL.Path[len(route):]
			cacheKey := buildKey(route, query, body)

			// try to get request from cache
			value, found := s.cache.Get(cacheKey)
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
				resp := recorder.Result()
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					if err != nil {
						info("response body could not be closed")
					}
				}(resp.Body)

				httpStatusCode = resp.StatusCode
				buf = recorder.Body.Bytes()

				// create new cache element
				ce := cacheElement{
					buffer:     buf,
					statusCode: httpStatusCode,
				}

				// only insert in cache if no error occurred
				if httpStatusCode < http.StatusBadRequest {
					s.cache.SetWithTTL(cacheKey, ce, 1, ttl)
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

func (s *Server) basicAuth() adapter {
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
			if subtle.ConstantTimeCompare([]byte(s.basicAuthUser), []byte(requestUser)) != 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// constant time compare and sleep to avoid timing attacks
			if equal, err := password.ComparePassword(requestPassword, s.basicAuthHash); err != nil {
				info(cliutil.ShowCallInfo(), err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			} else if !equal {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

func maxBody() adapter {
	return func(h http.Handler, route string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
			h.ServeHTTP(w, r)
		})
	}
}

func limitMethod(method string) adapter {
	return func(h http.Handler, route string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				info("error received", r.Method, "request for route", route, "instead of", method)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
