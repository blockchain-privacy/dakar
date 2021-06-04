package server

import (
	heuristic "backend/analytics/heuristics/transaction"
	"backend/cmd/cliutil"
	"backend/constants"
	dbaddr "backend/db/address"
	dbtxh "backend/db/analytics/heuristics/transaction"
	dbstat "backend/db/status"
	dbus "backend/db/user"
	"backend/external"

	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/ristretto"
	"golang.org/x/crypto/ed25519"
)

var (
	errorHeuristicSummary   = "error getting heuristic summary"
	errorHeuristics         = "error getting heuristics"
	errorHeuristicExecution = "error executing heuristics"
	errorHeuristicDetails   = "error getting heuristic details"
	errorInvalidSortOrder   = "error invalid sort order"
	errorInvalidFilter      = "error invalid filter"
	errorInvalidOffset      = "error invalid offset"
)

type searchResponse struct {
	Type    queryResultType `json:"type,omitempty"`
	Payload interface{}     `json:"payload,omitempty"`
}

func setDefaultHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
	w.Header().Set("Content-Type", "application/json")
}

// setCacheHeader sets the client side caching to a third of the server side cache
func setCacheHeader(w http.ResponseWriter, duration time.Duration) {
	if duration == time.Duration(0) {
		duration = time.Hour * 24
	}
	w.Header().Set("Cache-Control", "max-age="+strconv.FormatInt(int64(duration/time.Second/3), 10))
}

// API pattern: "/api/v1/search/<hash>"
func handlerSearch(dgraph *dgo.Dgraph) func(string, []byte) ([]byte, error) {
	return func(query string, body []byte) (response []byte, err error) {
		// set response struct
		resp := searchResponse{
			Type:    typeEmpty,
			Payload: nil,
		}

		if isValid(query) {
			searchOrder := []func(*dgo.Dgraph, string) (SearchResult, bool, error){GetTransaction, GetAddress, GetBlock}

			if isLikelyBlock(query) {
				searchOrder = []func(*dgo.Dgraph, string) (SearchResult, bool, error){GetBlock, GetTransaction, GetAddress}
			} else if isLikelyAddress(query) {
				searchOrder = []func(*dgo.Dgraph, string) (SearchResult, bool, error){GetAddress, GetTransaction, GetBlock}
			}

			// iterate over db access functions
			for _, fn := range searchOrder {
				data, ok, handlerErr := fn(dgraph, query)
				if handlerErr != nil {
					err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), handlerErr)
					return
				}
				// nothing found -> next try
				if !ok {
					continue
				}

				resp.Payload = data.result
				resp.Type = data.resultType
				break
			}
		}

		// encoding
		response, err = json.Marshal(resp)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		return
	}
}

// API pattern: "/api/v1/<type>/<query>"
// API pattern: "/api/v1/blk/<query>"
// API pattern: "/api/v1/address/<query>"
// API pattern: "/api/v1/tx/<query>"
func handlerDetails(dgraph *dgo.Dgraph, fn func(*dgo.Dgraph, string) (
	SearchResult, bool, error)) func(string, []byte) ([]byte, error) {
	return func(query string, body []byte) (response []byte, err error) {
		// set response struct
		resp := searchResponse{
			Type:    "response_empty",
			Payload: nil,
		}

		if isValid(query) {
			data, ok, fnErr := fn(dgraph, query)
			if fnErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), fnErr)
				return
			}
			if ok {
				resp.Payload = data.result
				resp.Type = data.resultType
			}
		}

		// encoding
		response, err = json.Marshal(resp)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		return
	}
}

// API pattern: "/api/v1/<type>/<query>"
// API pattern: "/api/v1/blk/<query>"
// API pattern: "/api/v1/address/<query>"
// API pattern: "/api/v1/tx/<query>"
func handlerDetailsV2(dgraph *dgo.Dgraph, route string, fn func(*dgo.Dgraph, string) (
	SearchResult, bool, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := r.URL.Path[len(route):]

		// set response struct
		reply := searchResponse{
			Type:    "response_empty",
			Payload: nil,
		}

		if isValid(queryString) {
			data, ok, fnErr := fn(dgraph, queryString)
			if fnErr == nil && ok {
				reply.Payload = data.result
				reply.Type = data.resultType
			}
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/addressOutputRange/<address_hash>"
func handlerAddressOutputRange(dgraph *dgo.Dgraph) func(string, []byte) ([]byte, error) {
	return func(query string, body []byte) (response []byte, err error) {
		resp := searchResponse{
			Type:    "response_empty",
			Payload: nil,
		}

		if isValid(query) {
			type request struct {
				Offset int   `json:"offset"`
				Order  int   `json:"order"`
				Filter []int `json:"filter"`
			}

			var addressRequest request
			addressRequest.Offset = -1
			addressRequest.Order = -1

			if decodeErr := json.Unmarshal(body, &addressRequest); decodeErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), decodeErr)
				return
			}

			if !dbaddr.IsValidSortOrder(addressRequest.Order) {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errors.New(errorInvalidSortOrder))
				return
			}

			if !dbaddr.IsValidFilter(addressRequest.Filter) {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errors.New(errorInvalidFilter))
				return
			}

			if addressRequest.Offset < 0 {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errors.New(errorInvalidOffset))
				return
			}

			data, ok, addrErr := GetAddressWithOptions(dgraph, query,
				addressRequest.Order, addressRequest.Offset, addressRequest.Filter)
			if addrErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), addrErr)
				return
			}
			if ok {
				resp.Payload = data.result
				resp.Type = data.resultType
			}
		}

		// encoding
		response, err = json.Marshal(resp)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		return
	}
}

// API pattern: "/api/v1/meta/"
func handlerMeta(dgraph *dgo.Dgraph,
	client external.RPCClient) func(string, []byte) ([]byte, error) {
	return func(query string, body []byte) (response []byte, err error) {
		// async request rpc info
		futureBlockchainInfo := client.GetBlockChainInfoAsync()

		// get data from db
		verboseStatus, err := dbstat.GetFrontendStatus(dgraph)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		// receive async rpc info
		rpcInfo, err := futureBlockchainInfo.Receive()
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		// set response struct
		stat := metaStatus{
			Status: verboseStatus,
			RPCInfo: prunedRPCInfo{
				Blocks:               rpcInfo.Blocks,
				Difficulty:           rpcInfo.Difficulty,
				VerificationProgress: math.Round(rpcInfo.VerificationProgress*10000) / 100,
				Pruned:               rpcInfo.Pruned,
			},
		}

		// encoding
		response, err = json.Marshal(stat)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		return
	}
}

// API pattern: "/api/v1/heuristicsSummary/<hash>"
func handlerHeuristicsSummary(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(constants.GetRouteHeuristicsSummary()):]

		if !isValid(txHashString) {
			http.Error(w, errorHeuristicSummary, http.StatusNotFound)
			return
		}

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			http.Error(w, errorHeuristicSummary, http.StatusNotFound)
			info(cliutil.ShowCallInfo(), err)
			return
		}

		cHeuristic, err := dbtxh.GetFrontendHeuristic(dgraph, txHashString, tUser.Id)
		if err != nil {
			http.Error(w, errorHeuristicSummary, http.StatusNotFound)
			info(cliutil.ShowCallInfo(), err)
			return
		}

		if len(cHeuristic.Heuristics) == 0 {
			http.Error(w, errorHeuristicSummary, http.StatusNotFound)
			return
		}

		// headers for streaming data to client
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", txHashString))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

		csvWriter := csv.NewWriter(w)
		csvWriter.Comma = ';'

		header := []string{"heuristic uid", "parent heuristic uid", "child heuristic uid",
			"heuristic type", "heuristic parameter", "heuristic timestamp",
			"origin uid", "origin transaction hash", "origin timestamp",
			"origin address hash"}

		if err = csvWriter.Write(header); err != nil {
			http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}

		for _, h := range cHeuristic.Heuristics {
			for _, result := range h.Results {
				var row []string
				// per heuristic information
				row = append(row, h.Uid)
				var parentHeuristic string
				if len(h.ParentHeuristic) > 0 {
					// only one parent heuristic is possible
					parentHeuristic = h.ParentHeuristic[0].Uid
				}
				row = append(row, parentHeuristic)

				var childHeuristics string
				for i, c := range h.ChildHeuristics {
					childHeuristics += c.Uid
					if i+1 < len(h.ChildHeuristics) {
						childHeuristics += ","
					}
				}

				row = append(row, childHeuristics)
				row = append(row, h.Type)
				row = append(row, h.Parameter)
				row = append(row, h.Timestamp)

				// per origin information
				row = append(row, result.Uid)
				row = append(row, result.TxHash)
				row = append(row, result.Timestamp)
				row = append(row, result.AddressHash)
				//row = append(row, strconv.Itoa(shortestPaths[result.Uid]))

				if err = csvWriter.Write(row); err != nil {
					http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
					info(cliutil.ShowCallInfo(), err)
				}
			}
			csvWriter.Flush()
		}
	})
}

// API pattern: "/api/v1/heuristics/<hash>"
func handlerHeuristics(dgraph *dgo.Dgraph, worker *heuristic.Worker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(constants.GetRouteHeuristics()):]

		if !isValid(txHashString) {
			http.Error(w, errorHeuristics, http.StatusNotFound)
			return
		}

		var reply heuristicReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getHeuristicReply(dgraph, worker, txHashString, tUser.Id)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristicStatus/<hash>"
func handlerHeuristicStatus(worker *heuristic.Worker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(constants.GetRouteHeuristicStatus()):]

		if !isValid(txHashString) {
			http.Error(w, errorHeuristics, http.StatusNotFound)
			return
		}

		var reply heuristicReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply.Success = true
			reply.Status = worker.GetStatus(txHashString, tUser.Id)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristicDetails/"
func handlerHeuristicsDetails(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			info(cliutil.ShowCallInfo(), err)
			return
		}

		var heuristicRequest struct {
			HeuristicUid string `json:"uid,omitempty"`
		}

		if err = json.NewDecoder(r.Body).Decode(&heuristicRequest); err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			info(cliutil.ShowCallInfo(), err)
			return
		}

		if len(heuristicRequest.HeuristicUid) == 0 {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			return
		}

		frontendHeuristic, err := dbtxh.GetFrontendHeuristicByUid(dgraph, heuristicRequest.HeuristicUid, tUser.Id)
		if err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			info(cliutil.ShowCallInfo(), err)
			return
		}

		// encoding
		if err = json.NewEncoder(w).Encode(frontendHeuristic); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/executeHeuristics/<hash>"
func handlerHeuristicsExecution(dgraph *dgo.Dgraph, worker *heuristic.Worker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(constants.GetRouteHeuristicsExecution()):]

		if !isValid(txHashString) {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			return
		}

		var reply heuristicExecutionReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getHeuristicExecutionReply(dgraph, worker, r.Body, txHashString, tUser.Id)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristicList/"
func handlerHeuristicList(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply heuristicListReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			items, err := dbtxh.GetHeuristicListByUser(dgraph, tUser.Id)
			if err != nil {
				info(cliutil.ShowCallInfo(), err)
			} else {
				reply.Success = true
				reply.Item = items
			}
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deleteHeuristic/"
func handlerDeleteHeuristic(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteHeuristicReply
		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error extracting user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteHeuristicReply(dgraph, r.Body, tUser.Id)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/createUser"
func handlerCreateUser(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getCreateUserReply(dgraph, r.Body)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/getUsers/"
func handlerGetUsers(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		users, err := dbus.GetUsers(dgraph)
		if err != nil {
			handleError(w, err)
			return
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(users); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// API pattern: "/api/v1/logout/"
func handlerLogout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)
		invalidateToken(w)

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(userReply{Success: true}); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// API pattern: "/api/v1/deleteUser/<userUid>"
func handlerDeleteUser(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		userUid := r.URL.Path[len(constants.GetRouteDeleteUser()):]

		var reply userReply

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteUserReply(dgraph, userUid, tUser)
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(reply); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// API pattern: "/api/v1/login/"
func handlerLogin(dgraph *dgo.Dgraph, privateSigningKey ed25519.PrivateKey) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getLoginReply(dgraph, r.Body)

		// set token if login is successful
		if reply.Success {
			token, expirationTime, err := issueToken(reply.User.ToFrontendUserClientState(), privateSigningKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				info(cliutil.ShowCallInfo(), err)
				return
			}
			setTokenAsCookie(w, token, expirationTime)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(reply); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// API pattern: "/api/v1/modifyUser/"
func handlerModifyUser(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply backendUserReply

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getModifyUserReply(dgraph, r.Body, tUser)
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(reply); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// API pattern: "/api/v1/shortestTransactionPath/"
func handlerShortestTransactionPath(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getShortestTransactionPathReply(dgraph, r.Body)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/reverseLookup/<txhash>?forward=true&t=30"
func handlerConnectionLookup(dgraph *dgo.Dgraph, worker *heuristic.Worker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getConnectionLookupReply(dgraph, worker, r.URL.Query(), r.URL.Path)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/clusterLookup0/<addressHash>"
func handlerClusterLookup(dgraph *dgo.Dgraph, worker *heuristic.Worker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getClusterLookupReply(dgraph, worker, r.URL.Path)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// cacheMiddleware caches the response of handler for the specified ttl
func cacheMiddleware(cache *ristretto.Cache, route string, ttl time.Duration,
	handler func(query string, body []byte) ([]byte, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// set headers
		setDefaultHeader(w)
		setCacheHeader(w, ttl)

		// extract body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handleError(w, err)
			return
		}

		query := r.URL.Path[len(route):]
		cacheKey := buildKey(route, query, body)

		// try to get request from cache
		value, found := cache.Get(cacheKey)
		var buf []byte
		if found {
			buf = value.([]byte)
		} else {
			var handlerErr error
			buf, handlerErr = handler(query, body)
			if handlerErr != nil {
				handleError(w, handlerErr)
				return
			}

			cache.SetWithTTL(cacheKey, buf, 1, ttl)
		}

		// write response
		_, err = w.Write(buf)
		if err != nil {
			handleError(w, err)
			return
		}
	}
}

// setupHandlers creates endpoint handlers
func setupHandlers(dgraph *dgo.Dgraph, client external.RPCClient, worker *heuristic.Worker) {
	// get signing keys
	privkey, pubkey, err := GetSigningKeysFromEnv()
	if err != nil {
		panic(fmt.Sprintln("error getting signing keys", err))
	}

	// init cache
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		panic(fmt.Sprintln("error initializing cache", err))
	}

	// API end points

	// Search
	http.HandleFunc(constants.GetRouteSearch(),
		cacheMiddleware(cache, constants.GetRouteSearch(), time.Minute*10, handlerSearch(dgraph)))

	// Common data

	http.Handle(constants.GetRouteTransaction(),
		Adapt(handlerDetailsV2(dgraph, constants.GetRouteTransaction(), GetTransaction),
			cacheMiddlewareAdaptor(cache, constants.GetRouteTransaction(), time.Second*0)))

	//http.HandleFunc(constants.GetRouteTransaction(),
	//	cacheMiddleware(cache, constants.GetRouteTransaction(), time.Second*0, handlerDetails(dgraph, GetTransaction)))

	// setting block cache time to 10 Minutes because blocks at
	// the tip get updated via adding the 'next block' reference
	//http.HandleFunc(constants.GetRouteBlock(),
	//	cacheMiddleware(cache, constants.GetRouteBlock(), time.Minute*10, handlerDetails(dgraph, GetBlock)))

	http.Handle(constants.GetRouteBlock(),
		Adapt(handlerDetailsV2(dgraph, constants.GetRouteBlock(), GetBlock),
			cacheMiddlewareAdaptor(cache, constants.GetRouteBlock(), time.Second*10)))

	//http.HandleFunc(constants.GetRouteAddress(),
	//	cacheMiddleware(cache, constants.GetRouteAddress(), time.Minute*10, handlerDetails(dgraph, GetAddress)))
	http.Handle(constants.GetRouteAddress(),
		Adapt(handlerDetailsV2(dgraph, constants.GetRouteAddress(), GetAddress),
			cacheMiddlewareAdaptor(cache, constants.GetRouteAddress(), time.Second*10)))

	http.HandleFunc(constants.GetRouteAddressOutputRange(),
		cacheMiddleware(cache, constants.GetRouteAddressOutputRange(), time.Minute*10, handlerAddressOutputRange(dgraph)))

	// Meta
	http.HandleFunc(constants.GetRouteMeta(),
		cacheMiddleware(cache, constants.GetRouteMeta(), time.Second*10, handlerMeta(dgraph, client)))

	// Heuristic
	http.Handle(constants.GetRouteHeuristics(),
		Adapt(handlerHeuristics(dgraph, worker),
			authorizationMiddleware(constants.GetRouteHeuristics(), privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicStatus(),
		Adapt(handlerHeuristicStatus(worker),
			authorizationMiddleware(constants.GetRouteHeuristicStatus(), privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicDetails(),
		Adapt(handlerHeuristicsDetails(dgraph),
			authorizationMiddleware(constants.GetRouteHeuristicDetails(), privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicsExecution(),
		Adapt(handlerHeuristicsExecution(dgraph, worker),
			authorizationMiddleware(constants.GetRouteHeuristicsExecution(), privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicsSummary(),
		Adapt(handlerHeuristicsSummary(dgraph),
			authorizationMiddleware(constants.GetRouteHeuristicsSummary(), privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicList(),
		Adapt(handlerHeuristicList(dgraph),
			authorizationMiddleware(constants.GetRouteHeuristicList(), privkey, pubkey)))
	http.Handle(constants.GetRouteDeleteHeuristic(),
		Adapt(handlerDeleteHeuristic(dgraph),
			authorizationMiddleware(constants.GetRouteDeleteHeuristic(), privkey, pubkey)))

	// Analytics
	http.Handle(constants.GetRouteShortestTransactionPath(),
		Adapt(handlerShortestTransactionPath(dgraph),
			authorizationMiddleware(constants.GetRouteShortestTransactionPath(), privkey, pubkey)))

	http.Handle(constants.GetRouteConnectionLookup(),
		Adapt(handlerConnectionLookup(dgraph, worker),
			authorizationMiddleware(constants.GetRouteConnectionLookup(), privkey, pubkey)))

	http.Handle(constants.GetRouteClusterLookup(),
		Adapt(handlerClusterLookup(dgraph, worker),
			authorizationMiddleware(constants.GetRouteClusterLookup(), privkey, pubkey)))

	// User
	http.Handle(constants.GetRouteLogin(), handlerLogin(dgraph, privkey))
	http.Handle(constants.GetRouteLogout(), handlerLogout())
	http.Handle(constants.GetRouteCreateUser(),
		Adapt(handlerCreateUser(dgraph),
			authorizationMiddleware(constants.GetRouteCreateUser(), privkey, pubkey)))
	http.Handle(constants.GetRouteDeleteUser(),
		Adapt(handlerDeleteUser(dgraph),
			authorizationMiddleware(constants.GetRouteDeleteUser(), privkey, pubkey)))
	http.Handle(constants.GetRouteGetUsers(),
		Adapt(handlerGetUsers(dgraph),
			authorizationMiddleware(constants.GetRouteGetUsers(), privkey, pubkey)))
	http.Handle(constants.GetRouteModifyUser(),
		Adapt(handlerModifyUser(dgraph),
			authorizationMiddleware(constants.GetRouteModifyUser(), privkey, pubkey)))
}
