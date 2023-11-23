package server

import (
	"backend/cmd/cliutil"
	dbus "backend/db/user"
	"bytes"
	"context"
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
	if _, err := w.Write([]byte(msg)); err != nil {
		warn(cliutil.NewStackError(err))
	}
}

// sendUnauthorizedMessage sends an unauthorized message
func sendUnauthorizedMessage(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
	w.WriteHeader(http.StatusUnauthorized)
}

// extractRoles tries to extract roles from the given metadata
func extractRoles(metaDataPublic any) ([]dbus.Role, error) {
	metadata, ok := metaDataPublic.(map[string]any)
	if !ok {
		return nil, cliutil.NewStackErrorStr("identity has no public metadata")
	}

	rolesInterface, ok := metadata["roles"]
	if !ok {
		return nil, cliutil.NewStackErrorStr("identity has no field 'roles'")
	}

	roleInterfaces, ok := rolesInterface.([]any)
	if !ok {
		return nil, cliutil.NewStackErrorStr("roles could not be cast from interface")
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
		return "", cliutil.NewStackErrorStr("identity has no admin metadata")
	}

	dgraphUIDInterface, ok := metadata["dgraph_uid"]
	if !ok {
		return "", cliutil.NewStackErrorStr("identity has no field 'dgraph_uid'")
	}

	dgraphUID, ok := dgraphUIDInterface.(string)
	if !ok {
		return "", cliutil.NewStackErrorStr("dgraph UID could not be cast from interface")
	}

	return dgraphUID, nil
}

func (s *Server) authorization() adapter {
	return func(h http.Handler, route string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, sessionResponse, err := s.auth.FrontendApi.ToSession(r.Context()).
				Cookie(r.Header.Get("Cookie")).Execute() //nolint:bodyclose
			if err != nil {
				sendUnauthorizedMessage(w)
				warn(cliutil.NewStackError(err))
				return
			}

			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(sessionResponse.Body)

			// check if session active and not expired
			if session.Active == nil || session.ExpiresAt == nil ||
				!*session.Active || time.Until(*session.ExpiresAt) <= 0 {
				sendUnauthorizedMessage(w)
				return
			}

			// if only reissueDuration is left of the token lifetime it gets reissued
			const reissueDuration = time.Hour * 24 / 4

			if time.Until(*session.ExpiresAt) <= reissueDuration {
				_, extensionResponse, extensionErr := s.adminAuth.IdentityApi.
					ExtendSession(r.Context(), session.Id).Execute()
				if extensionErr != nil {
					sendUnauthorizedMessage(w)
					warn(cliutil.NewStackError(extensionErr))
					return
				}
				_ = extensionResponse.Body.Close()
			}

			dgraphUID, err := extractDgraphUID(session.Identity.MetadataPublic)
			if err != nil {
				sendUnauthorizedMessage(w)
				warn(err)
				return
			}

			roles, err := extractRoles(session.Identity.MetadataPublic)
			if err != nil {
				sendUnauthorizedMessage(w)
				warn(err)
				return
			}

			// check if route is allowed and get typed role
			routeAllowed := false
			for _, role := range roles {
				routeRole, roleErr := getRoleByName(role.Name)
				if roleErr != nil {
					writeUnauthorized(w, "")
					warn(roleErr)
					return
				}

				if routeRole.IsRouteAllowed(route) {
					routeAllowed = true
					break
				}
			}

			if !routeAllowed {
				writeUnauthorized(w, "route not allowed")
				info("tried to access restricted route: "+route, "session", session)

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
		header     http.Header
		statusCode int
	}
	return func(h http.Handler, route string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// extract body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "an error occurred", http.StatusInternalServerError)
				warn(cliutil.NewStackError(err))
				return
			}
			// reset body, so it can be read by the next handler
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			cacheKey := buildKey(route, r.URL.Path[len(route):]+r.URL.RawQuery, body)

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
				warn(cliutil.NewStackError(err))
			}
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
				warn(cliutil.NewStackErrorf("error received %s request for route %s instead of %s", r.Method, route, method))
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
