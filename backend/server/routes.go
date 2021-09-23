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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
func handlerSearch(dgraph external.Database, route string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := r.URL.Path[len(route):]

		// set response struct
		reply := searchResponse{
			Type:    typeEmpty,
			Payload: nil,
		}

		if isValid(queryString) {
			searchOrder := []func(external.Database, string) (SearchResult, bool, error){GetTransaction, GetAddress, GetBlock}

			if isLikelyBlock(queryString) {
				searchOrder = []func(external.Database, string) (SearchResult, bool, error){GetBlock, GetTransaction, GetAddress}
			} else if isLikelyAddress(queryString) {
				searchOrder = []func(external.Database, string) (SearchResult, bool, error){GetAddress, GetTransaction, GetBlock}
			}

			// iterate over db access functions
			for _, fn := range searchOrder {
				data, ok, err := fn(dgraph, queryString)
				if err != nil {
					info(cliutil.ShowCallInfo(), err)
					break
				}
				// nothing found -> next try
				if !ok {
					continue
				}

				reply.Payload = data.result
				reply.Type = data.resultType
				break
			}
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/<type>/<query>"
// API pattern: "/api/v1/blk/<query>"
// API pattern: "/api/v1/address/<query>"
// API pattern: "/api/v1/tx/<query>"
func handlerDetails(dgraph external.Database, route string, fn func(external.Database, string) (
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
			if fnErr != nil {
				handleError(w, fnErr)
				return
			}

			if ok {
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
func handlerAddressOutputRange(dgraph external.Database, route string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := r.URL.Path[len(route):]

		reply := searchResponse{
			Type:    "response_empty",
			Payload: nil,
		}

		if isValid(queryString) {
			type request struct {
				Offset int   `json:"offset"`
				Order  int   `json:"order"`
				Filter []int `json:"filter"`
			}

			var addressRequest request
			addressRequest.Offset = -1
			addressRequest.Order = -1

			body, err := io.ReadAll(r.Body)
			if err != nil {
				handleError(w, err)
				return
			}

			if decodeErr := json.Unmarshal(body, &addressRequest); decodeErr != nil {
				handleError(w, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), decodeErr))
				return
			}

			if !dbaddr.IsValidSortOrder(addressRequest.Order) {
				handleError(w, errors.New(errorInvalidSortOrder))
				return
			}

			if !dbaddr.IsValidFilter(addressRequest.Filter) {
				handleError(w, errors.New(errorInvalidFilter))
				return
			}

			if addressRequest.Offset < 0 {
				handleError(w, errors.New(errorInvalidOffset))
				return
			}

			data, ok, addrErr := GetAddressWithOptions(dgraph, queryString,
				addressRequest.Order, addressRequest.Offset, addressRequest.Filter)
			if addrErr != nil {
				handleError(w, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), addrErr))
				return
			}
			if ok {
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

// API pattern: "/api/v1/blkRange/<address_hash>"
func handlerBlockRange(dgraph external.Database, route string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := r.URL.Path[len(route):]

		reply := searchResponse{
			Type:    "response_empty",
			Payload: nil,
		}

		if isValid(queryString) {
			type request struct {
				Offset int `json:"offset"`
			}

			var blockRequest request
			blockRequest.Offset = -1

			body, err := io.ReadAll(r.Body)
			if err != nil {
				handleError(w, err)
				return
			}

			if decodeErr := json.Unmarshal(body, &blockRequest); decodeErr != nil {
				handleError(w, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), decodeErr))
				return
			}

			if blockRequest.Offset < 0 {
				handleError(w, errors.New(errorInvalidOffset))
				return
			}

			data, ok, blockErr := GetBlockWithOptions(dgraph, queryString, blockRequest.Offset)
			if blockErr != nil {
				handleError(w, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), blockErr))
				return
			}
			if ok {
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

// API pattern: "/api/v1/meta/"
func handlerMeta(dgraph external.Database, client external.RPCClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)
		// async request rpc info
		futureBlockchainInfo := client.GetBlockChainInfoAsync()

		// get data from db
		verboseStatus, err := dbstat.GetFrontendStatus(dgraph)
		if err != nil {
			handleError(w, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
			return
		}

		// receive async rpc info
		rpcInfo, err := futureBlockchainInfo.Receive()
		if err != nil {
			handleError(w, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err))
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
				SizeOnDisk:           rpcInfo.SizeOnDisk,
			},
		}

		// encoding
		if encErr := json.NewEncoder(w).Encode(stat); encErr != nil {
			handleError(w, encErr)
		}
	})
}

// API pattern: "/api/v1/heuristicsSummary/<hash>"
func handlerHeuristicsSummary(dgraph external.Database) http.Handler {
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

		cHeuristic, err := dbtxh.GetFrontendHeuristic(dgraph, txHashString, tUser.ID)
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

		// somehow both content-length and transfer-encoding headers are both set, so one must be removed
		//w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

		csvWriter := csv.NewWriter(w)
		csvWriter.Comma = ';'

		header := []string{"heuristic uid", "parent heuristic uid", "child heuristic uid",
			"heuristic type", "heuristic parameter", "heuristic timestamp",
			"origin uid", "origin transaction hash", "origin timestamp",
			"origin address hash", "destination uid", "destination transaction hash", "destination timestamp"}

		if err = csvWriter.Write(header); err != nil {
			http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}

		for _, h := range cHeuristic.Heuristics {
			for _, result := range h.Results {
				var row []string
				// per heuristic information
				row = append(row, h.UID)
				var parentHeuristic string
				if len(h.ParentHeuristic) > 0 {
					// only one parent heuristic is possible
					parentHeuristic = h.ParentHeuristic[0].UID
				}
				row = append(row, parentHeuristic)

				var childHeuristics string
				for i, c := range h.ChildHeuristics {
					childHeuristics += c.UID
					if i+1 < len(h.ChildHeuristics) {
						childHeuristics += ","
					}
				}

				row = append(row, childHeuristics)
				row = append(row, h.Type)
				row = append(row, h.Parameter)
				row = append(row, h.Timestamp)

				// per origin information
				row = append(row, result.Origin.UID)
				row = append(row, result.Origin.TxHash)
				row = append(row, result.Origin.Timestamp)
				row = append(row, result.Origin.AddressHash)

				// add destination data if there exists any
				if len(result.Destinations) > 0 {
					for _, d := range result.Destinations {
						withDestinations := make([]string, len(row))
						// copy because for each destination the row gets reused
						copy(withDestinations, row)

						withDestinations = append(withDestinations, d.UID)
						withDestinations = append(withDestinations, d.TxHash)
						withDestinations = append(withDestinations, d.Timestamp)

						if err = csvWriter.Write(withDestinations); err != nil {
							http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
							info(cliutil.ShowCallInfo(), err)
						}
					}
					csvWriter.Flush()
				} else {
					if err = csvWriter.Write(row); err != nil {
						http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
						info(cliutil.ShowCallInfo(), err)
					}
				}
			}
			csvWriter.Flush()
		}
	})
}

// API pattern: "/api/v1/heuristics/<hash>"
func handlerHeuristics(dgraph external.Database, worker *heuristic.Worker) http.Handler {
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
			reply = getHeuristicReply(dgraph, worker, txHashString, tUser.ID)
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
			reply.Status = worker.GetStatus(txHashString, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristicDetails/"
func handlerHeuristicsDetails(dgraph external.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			info(cliutil.ShowCallInfo(), err)
			return
		}

		var heuristicRequest struct {
			HeuristicUID string `json:"uid,omitempty"`
		}

		if err = json.NewDecoder(r.Body).Decode(&heuristicRequest); err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			info(cliutil.ShowCallInfo(), err)
			return
		}

		if len(heuristicRequest.HeuristicUID) == 0 {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			return
		}

		frontendHeuristic, err := dbtxh.GetFrontendHeuristicByUID(dgraph, heuristicRequest.HeuristicUID, tUser.ID)
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
func handlerHeuristicsExecution(dgraph external.Database, worker *heuristic.Worker) http.Handler {
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
			reply = getHeuristicExecutionReply(dgraph, worker, r.Body, txHashString, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristicList/"
func handlerHeuristicList(dgraph external.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply heuristicListReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			items, err := dbtxh.GetHeuristicListByUser(dgraph, tUser.ID)
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

// API pattern: "/api/v1/heuristicDescriptors/"
func handlerHeuristicDescriptors() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply heuristicDescriptorReply

		for _, t := range heuristic.ValidHeuristicTypes {
			reply.Descriptors = append(reply.Descriptors, t.GetDescriptor())
		}

		reply.Success = true

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deleteHeuristic/"
func handlerDeleteHeuristic(dgraph external.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteHeuristicReply
		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error extracting user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteHeuristicReply(dgraph, r.Body, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/createUser"
func handlerCreateUser(dgraph external.Database) http.Handler {
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
func handlerGetUsers(dgraph external.Database) http.Handler {
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
func handlerDeleteUser(dgraph external.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		userUID := r.URL.Path[len(constants.GetRouteDeleteUser()):]

		var reply userReply

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteUserReply(dgraph, userUID, tUser)
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(reply); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// API pattern: "/api/v1/login/"
func handlerLogin(dgraph external.Database, privateSigningKey ed25519.PrivateKey) http.Handler {
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
func handlerModifyUser(dgraph external.Database) http.Handler {
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
func handlerShortestTransactionPath(dgraph external.Database) http.Handler {
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
func handlerConnectionLookup(dgraph external.Database, worker *heuristic.Worker) http.Handler {
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

// API pattern: "/api/v1/clusterLookup/<addressHash>"
func handlerClusterLookup(dgraph external.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getClusterLookupReply(dgraph, r.Body)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// setupHandlers creates endpoint handlers
func setupHandlers(dgraph external.Database, client external.RPCClient, worker *heuristic.Worker,
	basicAuthUser string, basicAuthHash string, tokenPublicKey string, tokenPrivateKey string) {

	privkey, err := hex.DecodeString(tokenPrivateKey)
	if err != nil {
		panic(err)
	}

	pubkey, err := hex.DecodeString(tokenPublicKey)
	if err != nil {
		panic(err)
	}

	// init cache
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10 M).
		MaxCost:     1 << 30, // maximum cost of cache (1 GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		panic(fmt.Sprintln("error initializing cache", err))
	}

	// API end points

	// Metrics
	http.Handle(constants.GetRouteMetrics(), adapt(promhttp.Handler(), constants.GetRouteMetrics(),
		basicAuthMiddleware(basicAuthUser, basicAuthHash)))

	// Search
	http.Handle(constants.GetRouteSearch(),
		adapt(handlerSearch(dgraph, constants.GetRouteSearch()), constants.GetRouteSearch(),
			cacheMiddleware(cache, time.Minute*10)))

	// Common data
	http.Handle(constants.GetRouteTransaction(),
		adapt(handlerDetails(dgraph, constants.GetRouteTransaction(), GetTransaction), constants.GetRouteTransaction(),
			cacheMiddleware(cache, time.Second*0)))
	// setting block cache time to 10 Minutes because blocks at
	// the tip get updated via adding the 'next block' reference
	http.Handle(constants.GetRouteBlock(),
		adapt(handlerDetails(dgraph, constants.GetRouteBlock(), GetBlock), constants.GetRouteBlock(),
			cacheMiddleware(cache, time.Second*10)))
	http.Handle(constants.GetRouteAddress(),
		adapt(handlerDetails(dgraph, constants.GetRouteAddress(), GetAddress), constants.GetRouteAddress(),
			cacheMiddleware(cache, time.Second*10)))

	http.Handle(constants.GetRouteAddressOutputRange(),
		adapt(handlerAddressOutputRange(dgraph, constants.GetRouteAddressOutputRange()), constants.GetRouteAddressOutputRange(),
			cacheMiddleware(cache, time.Minute*10)))

	http.Handle(constants.GetRouteBlockRange(),
		adapt(handlerBlockRange(dgraph, constants.GetRouteBlockRange()), constants.GetRouteBlockRange(),
			cacheMiddleware(cache, time.Minute*10)))

	// Meta
	http.Handle(constants.GetRouteMeta(),
		adapt(handlerMeta(dgraph, client), constants.GetRouteMeta(),
			authorizationMiddleware(privkey, pubkey),
			cacheMiddleware(cache, time.Second*10)))

	// Heuristic
	http.Handle(constants.GetRouteHeuristics(),
		adapt(handlerHeuristics(dgraph, worker), constants.GetRouteHeuristics(),
			authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicStatus(),
		adapt(handlerHeuristicStatus(worker), constants.GetRouteHeuristicStatus(),
			authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicDetails(),
		adapt(handlerHeuristicsDetails(dgraph), constants.GetRouteHeuristicDetails(),
			authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicsExecution(),
		adapt(handlerHeuristicsExecution(dgraph, worker), constants.GetRouteHeuristicsExecution(),
			authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicsSummary(),
		adapt(handlerHeuristicsSummary(dgraph), constants.GetRouteHeuristicsSummary(),
			authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicList(),
		adapt(handlerHeuristicList(dgraph), constants.GetRouteHeuristicList(),
			authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicDescriptors(),
		adapt(handlerHeuristicDescriptors(), constants.GetRouteHeuristicDescriptors(),
			authorizationMiddleware(privkey, pubkey), cacheMiddleware(cache, 0)))
	http.Handle(constants.GetRouteDeleteHeuristic(),
		adapt(handlerDeleteHeuristic(dgraph), constants.GetRouteDeleteHeuristic(),
			authorizationMiddleware(privkey, pubkey)))

	// Analytics
	http.Handle(constants.GetRouteShortestTransactionPath(),
		adapt(handlerShortestTransactionPath(dgraph), constants.GetRouteShortestTransactionPath(),
			authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteConnectionLookup(),
		adapt(handlerConnectionLookup(dgraph, worker), constants.GetRouteConnectionLookup(),
			authorizationMiddleware(privkey, pubkey)))

	// Clusters
	http.Handle(constants.GetRouteClusterLookup(),
		adapt(handlerClusterLookup(dgraph), constants.GetRouteClusterLookup(),
			authorizationMiddleware(privkey, pubkey)))

	// User
	http.Handle(constants.GetRouteLogin(), handlerLogin(dgraph, privkey))
	http.Handle(constants.GetRouteLogout(), handlerLogout())
	http.Handle(constants.GetRouteCreateUser(),
		adapt(handlerCreateUser(dgraph), constants.GetRouteCreateUser(), authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteDeleteUser(),
		adapt(handlerDeleteUser(dgraph), constants.GetRouteDeleteUser(), authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteGetUsers(),
		adapt(handlerGetUsers(dgraph), constants.GetRouteGetUsers(), authorizationMiddleware(privkey, pubkey)))
	http.Handle(constants.GetRouteModifyUser(),
		adapt(handlerModifyUser(dgraph), constants.GetRouteModifyUser(), authorizationMiddleware(privkey, pubkey)))
}
