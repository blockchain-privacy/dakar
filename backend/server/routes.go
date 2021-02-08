package server

import (
	heuristic "backend/analytics/heuristics/transaction"
	"backend/cmd/cliutil"
	"backend/constants"
	dbaddr "backend/db/address"
	dban "backend/db/analytics"
	dbtxh "backend/db/analytics/heuristics/transaction"
	dbstat "backend/db/status"
	dbtx "backend/db/transaction"
	dbus "backend/db/user"
	"backend/user"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/ed25519"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/btcsuite/btcd/rpcclient"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/ristretto"
)

const (
	maxOrigins = 1000
)

var (
	errorPath               = "error getting paths"
	errorHeuristics         = "error getting heuristics"
	errorHeuristicExecution = "error executing heuristics"
	errorHeuristicDetails   = "error getting heuristic details"
	errorInvalidSortOrder   = "error invalid sort order"
	errorInvalidFilter      = "error invalid filter"
	errorInvalidOffset      = "error invalid offset"
)

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

type searchResponse struct {
	Type    queryResultType `json:"type,omitempty"`
	Payload interface{}     `json:"payload,omitempty"`
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

var (
	errHttpDefault = errors.New("an error occurred")
)

// API pattern: "/api/v1/meta/"
func handlerMeta(dgraph *dgo.Dgraph,
	client *rpcclient.Client) func(string, []byte) ([]byte, error) {
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

var lock sync.Mutex

// API pattern: "/api/v1/paths/"
func handlerPaths(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(constants.GetRouteOrigins()):]

		if !isValid(txHashString) {
			http.Error(w, errorPath, http.StatusNotFound)
			return
		}

		originCount, err := dban.GetOriginCount(dgraph, txHashString)
		if err != nil {
			http.Error(w, errorPath, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		// returned data is getting to big
		if originCount > maxOrigins {
			http.Error(w, "getting paths is only supported up to "+strconv.Itoa(maxOrigins)+" origins", http.StatusNotFound)
			return
		}

		lock.Lock()
		paths, transactions, err := dban.GetPaths(dgraph, txHashString)
		if err != nil {
			http.Error(w, errorPath, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}
		lock.Unlock()

		if len(paths) == 0 {
			http.Error(w, errorPath, http.StatusNotFound)
			return
		}

		// headers for streaming data to client
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", txHashString))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

		csvWriter := csv.NewWriter(w)
		csvWriter.Comma = ';'

		header := []string{"path id", "path step", "tx hash", "type", "block hash", "block height", "timestamp"}
		if err = csvWriter.Write(header); err != nil {
			http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}

		for i, p := range paths {
			for j, e := range p {
				tx := transactions[e.Hash]
				var row []string
				row = append(row, strconv.Itoa(i+1))
				row = append(row, strconv.Itoa(j+1))
				row = append(row, e.Hash)

				if e.IsOrigin {
					row = append(row, dbtx.PrivacyOrigin)
				} else {
					row = append(row, dbtx.PrivacyMixing)
				}

				row = append(row, tx.BlockHash)
				row = append(row, strconv.FormatUint(tx.BlockId, 10))
				row = append(row, tx.BlockTimestamp)
				if err = csvWriter.Write(row); err != nil {
					http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
					serverInfo(cliutil.ShowCallInfo(), err)
				}
			}
			csvWriter.Flush()
		}
	}
}

// API pattern: "/api/v1/heuristicsSummary/<hash>"
func handlerHeuristicsSummary(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(constants.GetRouteHeuristicsSummary()):]

		if !isValid(txHashString) {
			http.Error(w, errorPath, http.StatusNotFound)
			return
		}

		cHeuristic, err := dbtxh.GetFrontendHeuristic(dgraph, txHashString)
		if err != nil {
			log.Println(err)
			return
		}

		if len(cHeuristic.Heuristics) == 0 {
			http.Error(w, errorPath, http.StatusNotFound)
			return
		}

		// calculate shortest path
		//shortestPaths := make(map[string]int)
		//lock.Lock()
		//for _, h := range cHeuristic.Heuristics {
		//
		//	for _, r := range h.Results {
		//		if _, ok := shortestPaths[r.Uid]; !ok {
		//			pathLen, pathErr := dbtxh.GetShortestPathLength(dgraph, cHeuristic.Uid, r.Uid)
		//			if pathErr != nil {
		//				log.Println(pathErr)
		//				return
		//			}
		//			shortestPaths[r.Uid] = pathLen
		//		}
		//	}
		//}
		//lock.Unlock()

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
		//header := []string{"heuristic uid", "parent heuristic uid", "child heuristic uid",
		//	"heuristic type", "heuristic parameter", "heuristic timestamp",
		//	"origin uid", "origin transaction hash", "origin timestamp",
		//	"origin address hash", "origin shortest path"}
		if err = csvWriter.Write(header); err != nil {
			http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
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
					serverInfo(cliutil.ShowCallInfo(), err)
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

		heuristics, err := dbtxh.GetBasicFrontendHeuristic(dgraph, txHashString)
		if err != nil {
			http.Error(w, errorHeuristics, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		resp := heuristicReply{
			Heuristics: heuristics,
			Status:     worker.GetStatus(txHashString),
		}

		// encoding
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
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

		resp := heuristicReply{
			Status: worker.GetStatus(txHashString),
		}

		// encoding
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristicDetails/<hash>"
func handlerHeuristicsDetails(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(constants.GetRouteHeuristicDetails()):]

		if !isValid(txHashString) {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			return
		}

		type request struct {
			HeuristicUid string `json:"uid,omitempty"`
		}

		var heuristicRequest request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&heuristicRequest)
		if err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		if len(heuristicRequest.HeuristicUid) == 0 {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			return
		}

		frontendHeuristic, err := dbtxh.GetFrontendHeuristicByUid(dgraph, heuristicRequest.HeuristicUid, txHashString)
		if err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		// encoding
		err = json.NewEncoder(w).Encode(frontendHeuristic)
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
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

		resp := heuristicReply{}

		if worker.IsInQueue(txHashString) {
			resp.Status = heuristic.StatusHeuristicDuplicate
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				http.Error(w, "encoding error", http.StatusInternalServerError)
				serverInfo(cliutil.ShowCallInfo(), err)
			}

			serverInfo(cliutil.ShowCallInfo(), "heuristic already in queue")
			return
		}

		type request struct {
			Changed []dbtxh.FrontendHeuristic `json:"changed,omitempty"`
			Deleted []string                  `json:"deleted,omitempty"`
		}

		var heuristicRequest request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&heuristicRequest)
		if err != nil {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		if len(heuristicRequest.Changed) == 0 && len(heuristicRequest.Deleted) == 0 {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			return
		}

		work, err := heuristic.CreateWork(dgraph, txHashString, heuristicRequest.Changed,
			heuristicRequest.Deleted)
		if err != nil {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		addedWork := worker.AddWork(txHashString, work)

		if addedWork {
			resp.Status = heuristic.StatusHeuristicAdded
		} else {
			resp.Status = heuristic.StatusHeuristicDuplicate
		}

		// encoding
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
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
			serverInfo(cliutil.ShowCallInfo(), err)
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
			serverInfo(cliutil.ShowCallInfo(), encodingErr)
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
			serverInfo(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// API pattern: "/api/v1/deleteUser/<userUid>"
func handlerDeleteUser(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		userUid := r.URL.Path[len(constants.GetRouteDeleteUser()):]

		reply := userReply{
			Success: true,
		}

		if err := dbus.DeleteUser(dgraph, userUid); err != nil {
			reply.Success = false
			reply.Msg = "could not delete user"
			serverInfo(err)
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(reply); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), encodingErr)
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
				serverInfo(cliutil.ShowCallInfo(), err)
				return
			}
			setTokenAsCookie(w, token, expirationTime)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(reply); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// getModifyUserReply parses the input and creates a corresponding userReply
func getModifyUserReply(dgraph *dgo.Dgraph, body io.Reader, tUser tokenUser) (reply backendUserReply) {
	// get clients user state
	var modRequest dbus.ModifyUserRequest
	if err := json.NewDecoder(body).Decode(&modRequest); err != nil {
		reply.Msg = "could not decode user data"
		return
	}

	if len(modRequest.Uid) == 0 ||
		(len(modRequest.Roles) == 0 && len(modRequest.Email) == 0 && len(modRequest.NewPassword) == 0) {
		reply.Msg = "nothing to change"
		return
	}

	// check if passwords are equal
	if len(modRequest.CurrentPassword) > 0 && len(modRequest.NewPassword) > 0 &&
		modRequest.NewPassword == modRequest.CurrentPassword {
		reply.Msg = "passwords are equal"
		return
	}

	// is user an admin
	isAdmin := false
	for _, r := range tUser.Roles {
		if r.Name == user.AdminRoleName {
			isAdmin = true
			break
		}
	}

	// if user ids does not match, check if this is a request from an admin user
	if modRequest.Uid != tUser.Id && !isAdmin {
		reply.Msg = "user ids do not match"
		serverInfo(cliutil.ShowCallInfo(), "user", tUser.Id, "tried to modify user", modRequest.Uid)
		return
	}

	// check current password if user is not an admin
	if !isAdmin {
		if len(modRequest.CurrentPassword) == 0 {
			reply.Msg = "current password must also be supplied"
			return
		}

		dbUser, err := dbus.GetUser(dgraph, modRequest.Uid)
		if err != nil {
			reply.Msg = "error modifying user"
			serverInfo(cliutil.ShowCallInfo(), err, modRequest)
			return
		}

		if ok, err := user.ComparePassword(modRequest.CurrentPassword, dbUser.PasswordHash); !ok || err != nil {
			reply.Msg = "wrong current password"
			return
		}
	}

	// check email
	if len(modRequest.Email) > 0 {
		if !dbus.IsValidEmail(modRequest.Email) {
			reply.Msg = "invalid email"
			return
		}

		emailUser, err := dbus.GetUserByEmail(dgraph, modRequest.Email)
		if err != nil {
			if !errors.Is(dbus.ErrorUsersNotFound, err) {
				reply.Msg = "invalid email"
				serverInfo(cliutil.ShowCallInfo(), err, modRequest)
				return
			}
		} else if emailUser.Uid != modRequest.Uid {
			reply.Msg = "duplicate email"
			serverInfo(cliutil.ShowCallInfo(), err, modRequest)
			return
		}
	}

	var newPwHash string
	// check if password matches
	if len(modRequest.NewPassword) > 0 {
		if len(modRequest.NewPassword) < 10 {
			reply.Msg = "new password must be at least 10 characters long"
			return
		}

		var generatePwErr error
		if newPwHash, generatePwErr = user.GeneratePasswordHash(user.DefaultPasswordConfig,
			modRequest.NewPassword); generatePwErr != nil {
			reply.Msg = "error modifying user"
			return
		}
	}

	// handle role change
	if len(modRequest.Roles) > 0 {
		if !isAdmin {
			reply.Msg = "user can not change its roles"
			serverInfo(cliutil.ShowCallInfo(), "user", tUser.Id, "tried to change its roles", modRequest.Roles)
			return
		}
		// check if all roles exists
		for _, r := range modRequest.Roles {
			if _, err := user.GetRoleByName(r.Name); err != nil {
				reply.Msg = "invalid role"
				serverInfo(cliutil.ShowCallInfo(), "user", tUser.Id, "provided invalid role", r.Name)
				return
			}
		}
		// delete existing roles if new roles are set
		if err := dbus.RemoveRolesFromUser(dgraph, modRequest.Uid); err != nil {
			reply.Msg = "error modifying user"
			serverInfo(cliutil.ShowCallInfo(), err, modRequest)
			return
		}
	}

	// modify user
	if err := dbus.ModifyUser(dgraph, modRequest.ToUser(newPwHash)); err != nil {
		reply.Msg = "error modifying user"
		serverInfo(cliutil.ShowCallInfo(), err, modRequest)
		return
	}

	// get new user information
	newUserInfo, err := dbus.GetUser(dgraph, modRequest.Uid)
	if err != nil {
		reply.Msg = "error modifying user"
		serverInfo(cliutil.ShowCallInfo(), err, modRequest)
		return
	}

	// set new user info
	newUserState := newUserInfo.ToFrontendUserBackendState()
	reply.User = &newUserState
	reply.Success = true

	return
}

// API pattern: "/api/v1/modifyUser/"
func handlerModifyUser(dgraph *dgo.Dgraph) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply backendUserReply

		if userInfo := r.Context().Value(middlewareContextUser); userInfo == nil {
			reply.Msg = "error modifying user"
		} else {
			if tUser := userInfo.(tokenUser); len(tUser.Id) == 0 {
				reply.Msg = "error modifying user"
			} else {
				reply = getModifyUserReply(dgraph, r.Body, tUser)
			}
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(reply); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), encodingErr)
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
		body, err := ioutil.ReadAll(r.Body)
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
func setupHandlers(ctx context.Context, dgraph *dgo.Dgraph, client *rpcclient.Client) {
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

	// init worker
	worker := heuristic.NewWorker()
	worker.StartWorking(ctx, dgraph)

	// API end points

	// Search
	http.HandleFunc(constants.GetRouteSearch(),
		cacheMiddleware(cache, constants.GetRouteSearch(), time.Minute*10, handlerSearch(dgraph)))

	// Common data
	// setting block cache time to 10 Minutes because blocks at
	// the tip get updated via adding the 'next block' reference
	http.HandleFunc(constants.GetRouteBlock(),
		cacheMiddleware(cache, constants.GetRouteBlock(), time.Minute*10, handlerDetails(dgraph, GetBlock)))
	http.HandleFunc(constants.GetRouteTransaction(),
		cacheMiddleware(cache, constants.GetRouteTransaction(), time.Second*0, handlerDetails(dgraph, GetTransaction)))
	http.HandleFunc(constants.GetRouteAddress(),
		cacheMiddleware(cache, constants.GetRouteAddress(), time.Minute*10, handlerDetails(dgraph, GetAddress)))
	http.HandleFunc(constants.GetRouteAddressOutputRange(),
		cacheMiddleware(cache, constants.GetRouteAddressOutputRange(), time.Minute*10, handlerAddressOutputRange(dgraph)))

	// Meta
	http.HandleFunc(constants.GetRouteMeta(),
		cacheMiddleware(cache, constants.GetRouteMeta(), time.Second*10, handlerMeta(dgraph, client)))

	// Origins
	http.HandleFunc(constants.GetRouteOrigins(), handlerPaths(dgraph))

	// Heuristic
	http.Handle(constants.GetRouteHeuristics(),
		Adapt(handlerHeuristics(dgraph, &worker),
			authorizationMiddleware(constants.GetRouteHeuristics(), privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicStatus(),
		Adapt(handlerHeuristicStatus(&worker),
			authorizationMiddleware(constants.GetRouteHeuristicStatus(), privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicDetails(),
		Adapt(handlerHeuristicsDetails(dgraph),
			authorizationMiddleware(constants.GetRouteHeuristicDetails(), privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicsExecution(),
		Adapt(handlerHeuristicsExecution(dgraph, &worker),
			authorizationMiddleware(constants.GetRouteHeuristicsExecution(), privkey, pubkey)))
	http.Handle(constants.GetRouteHeuristicsSummary(),
		Adapt(handlerHeuristicsSummary(dgraph),
			authorizationMiddleware(constants.GetRouteHeuristicsSummary(), privkey, pubkey)))

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
