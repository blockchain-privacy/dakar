package server

import (
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/constants"
	dbaddr "backend/db/address"
	dbtxh "backend/db/analytics/heuristics"
	dbstat "backend/db/status"
	dbus "backend/db/user"
	"backend/external"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	errorClusterSummary     = "error getting cluster summary"
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
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
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
func (s *Server) handlerSearch() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := path.Base(r.URL.Path)

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
				data, ok, err := fn(s.db, queryString)
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
func (s *Server) handlerDetails(fn func(external.Database, string) (SearchResult, bool, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := path.Base(r.URL.Path)

		// set response struct
		reply := searchResponse{
			Type:    "response_empty",
			Payload: nil,
		}

		if isValid(queryString) {
			data, ok, fnErr := fn(s.db, queryString)
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
func (s *Server) handlerAddressOutputRange() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := path.Base(r.URL.Path)

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

			data, ok, addrErr := GetAddressWithOptions(s.db, queryString,
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
func (s *Server) handlerBlockRange() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := path.Base(r.URL.Path)

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

			data, ok, blockErr := GetBlockWithOptions(s.db, queryString, blockRequest.Offset)
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
func (s *Server) handlerMeta() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)
		// async request rpc info
		futureBlockchainInfo := s.client.GetBlockChainInfoAsync()

		// get data from db
		verboseStatus, err := dbstat.GetFrontendStatus(s.db)
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

// API pattern: "/api/v1/heuristicsSummary/<heuristic_UID>"
func (s *Server) handlerHeuristicsSummary() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		heuristicUID := path.Base(r.URL.Path)

		if heuristicUID == "." || heuristicUID == "" {
			handleError(w, errors.New("no heuristic UID provided"))
			return
		}

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			handleError(w, err)
			return
		}

		writeHeuristicSummary(w, s.db, tUser, heuristicUID)
	})
}

// API pattern: "/api/v1/clusterSummary"
func (s *Server) handlerClusterSummary() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		writeClusterSummary(w, r, s.db)
	})
}

// API pattern: "/api/v1/addCluster"
func (s *Server) handlerAddCluster() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddClusterReply(s.db, r)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deleteCluster/<cluster_uid>"
func (s *Server) handlerDeleteCluster() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteClusterReply

		clusterUid := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteClusterReply(s.db, tUser.ID, clusterUid)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deleteAllClusters"
func (s *Server) handlerDeleteAllClusters() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteClusterReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteAllClustersReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/clusterOverview"
func (s *Server) handlerClusterOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply clusterOverviewReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getClusterOverviewReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/attributionOverview"
func (s *Server) handlerAttributionOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply attributionOverviewReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getAttributionOverviewReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/addPrivateAttribution"
func (s *Server) handlerAddPrivateAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddAttributionReply(s.db, r, false)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/addPublicAttribution"
func (s *Server) handlerAddPublicAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddAttributionReply(s.db, r, true)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deletePrivateAttribution/<cluster_uid>"
func (s *Server) handlerDeletePrivateAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAttributionReply

		attributionUid := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteAttributionReply(s.db, tUser.ID, attributionUid, false)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deletePublicAttribution/<cluster_uid>"
func (s *Server) handlerDeletePublicAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAttributionReply

		attributionUid := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteAttributionReply(s.db, tUser.ID, attributionUid, true)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deleteAllPrivateAttributions"
func (s *Server) handlerDeleteAllPrivateAttributions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAttributionReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteAllAttributionsReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/searchAttributions"
func (s *Server) handlerSearchAttributions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply attributionOverviewReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getAttributionSearchReply(s.db, tUser.ID, r.Body)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/addAddressExclusions"
func (s *Server) handlerAddAddressExclusions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddAddressExclusionsReply(s.db, r)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deleteAddressExclusion/<address_exclusion_uid>"
func (s *Server) handlerDeleteAddressExclusion() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAddressExclusionReply

		addressHash := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteAddressExclusionReply(s.db, tUser.ID, addressHash)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deleteAllAddressExclusions"
func (s *Server) handlerDeleteAllAddressExclusions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAddressExclusionReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteAllAddressExclusionsReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/addressExclusionOverview"
func (s *Server) handlerAddressExclusionOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply addressExclusionOverviewReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getAddressExclusionOverviewReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristics/<hash>"
func (s *Server) handlerHeuristics() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := path.Base(r.URL.Path)

		if !isValid(txHashString) {
			http.Error(w, errorHeuristics, http.StatusNotFound)
			return
		}

		var reply heuristicReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getHeuristicReply(s.db, s.worker, txHashString, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/hmiLookup/<hash>"
func (s *Server) handlerHMILookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		addressHash := path.Base(r.URL.Path)

		if !isValid(addressHash) {
			http.Error(w, errorHeuristics, http.StatusNotFound)
			return
		}

		reply := getHMILookupReply(s.db, addressHash)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristicStatus/<hash>"
func (s *Server) handlerHeuristicStatus() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := path.Base(r.URL.Path)

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
			reply.Status = s.worker.GetStatus(txHashString, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristicDetails/"
func (s *Server) handlerHeuristicsDetails() http.Handler {
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

		frontendHeuristic, err := dbtxh.GetFrontendHeuristicByUID(s.db, heuristicRequest.HeuristicUID, tUser.ID)
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
func (s *Server) handlerHeuristicsExecution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := path.Base(r.URL.Path)

		if !isValid(txHashString) {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			return
		}

		var reply heuristicExecutionReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getHeuristicExecutionReply(s.db, s.worker, r.Body, txHashString, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/heuristicList/"
func (s *Server) handlerHeuristicList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply heuristicListReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			items, err := dbtxh.GetHeuristicListByUser(s.db, tUser.ID)
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
func (s *Server) handlerHeuristicDescriptors() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply heuristicDescriptorReply

		for _, t := range heuristics.ValidHeuristicTypes {
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
func (s *Server) handlerDeleteHeuristic() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteHeuristicReply
		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error extracting user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteHeuristicReply(s.db, r.Body, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/createUser"
func (s *Server) handlerCreateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getCreateUserReply(s.db, r.Body)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/getUsers/"
func (s *Server) handlerGetUsers() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		users, err := dbus.GetUsers(s.db)
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
func (s *Server) handlerLogout() http.Handler {
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
func (s *Server) handlerDeleteUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		userUID := path.Base(r.URL.Path)

		var reply userReply

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteUserReply(s.db, userUID, tUser)
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(reply); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// API pattern: "/api/v1/login/"
func (s *Server) handlerLogin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getLoginReply(s.db, r.Body)

		// set token if login is successful
		if reply.Success {
			token, expirationTime, err := issueToken(reply.User.ToFrontendUserClientState(), s.tokenPrivateKey)
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
func (s *Server) handlerModifyUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply backendUserReply

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getModifyUserReply(s.db, r.Body, tUser)
		}

		// encoding
		if encodingErr := json.NewEncoder(w).Encode(reply); encodingErr != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), encodingErr)
		}
	})
}

// API pattern: "/api/v1/shortestTransactionPath/"
func (s *Server) handlerShortestTransactionPath() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getShortestTransactionPathReply(s.db, r.Body)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/reverseLookup/<txhash>?forward=true&t=30"
func (s *Server) handlerConnectionLookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getConnectionLookupReply(s.db, s.worker, r.URL.Query(), r.URL.Path)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/clusterLookup/<addressHash>"
func (s *Server) handlerClusterLookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply clusterLookupReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getClusterLookupReply(s.db, r.Body, tUser)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/mixingActivity/"
func (s *Server) handlerMixingActivity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getMixingActivity(s.db, r.Body)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/addressExclusionStatus/<address_hash>"
func (s *Server) handlerGetAddressExclusionStatus() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddressExclusionStatusReply(r, s.db, path.Base(r.URL.Path))

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// setupHandlers creates endpoint handlers
func (s *Server) setupHandlers() {
	// Metrics
	s.handler.Handle(constants.GetRouteMetrics(), adapt(promhttp.Handler(), constants.GetRouteMetrics(),
		limitMethod("GET"), s.basicAuth(), maxBody()))

	// Search
	s.handler.Handle(constants.GetRouteSearch(), adapt(s.handlerSearch(), constants.GetRouteSearch(),
		limitMethod("GET"), s.useCache(time.Minute*10), maxBody()))

	// Common data
	s.handler.Handle(constants.GetRouteTransaction(),
		adapt(s.handlerDetails(GetTransaction), constants.GetRouteTransaction(),
			limitMethod("GET"), s.useCache(time.Second*0), maxBody()))
	// setting block cache time to 10 Minutes because blocks at
	// the tip get updated via adding the 'next block' reference
	s.handler.Handle(constants.GetRouteBlock(), adapt(s.handlerDetails(GetBlock), constants.GetRouteBlock(),
		limitMethod("GET"), s.useCache(time.Second*10), maxBody()))
	s.handler.Handle(constants.GetRouteAddress(), adapt(s.handlerDetails(GetAddress), constants.GetRouteAddress(),
		limitMethod("GET"), s.useCache(time.Second*10), maxBody()))

	s.handler.Handle(constants.GetRouteAddressOutputRange(),
		adapt(s.handlerAddressOutputRange(), constants.GetRouteAddressOutputRange(),
			limitMethod("POST"), s.useCache(time.Minute*10), maxBody()))

	s.handler.Handle(constants.GetRouteBlockRange(), adapt(s.handlerBlockRange(), constants.GetRouteBlockRange(),
		limitMethod("POST"), s.useCache(time.Minute*10), maxBody()))

	// Meta
	s.handler.Handle(constants.GetRouteMeta(), adapt(s.handlerMeta(), constants.GetRouteMeta(),
		limitMethod("GET"), s.authorization(), s.useCache(time.Second*10), maxBody()))

	// heuristic
	s.handler.Handle(constants.GetRouteHeuristics(),
		adapt(s.handlerHeuristics(), constants.GetRouteHeuristics(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteHeuristicStatus(),
		adapt(s.handlerHeuristicStatus(), constants.GetRouteHeuristicStatus(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteHeuristicDetails(),
		adapt(s.handlerHeuristicsDetails(), constants.GetRouteHeuristicDetails(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteHeuristicsExecution(),
		adapt(s.handlerHeuristicsExecution(), constants.GetRouteHeuristicsExecution(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteHeuristicsSummary(),
		adapt(s.handlerHeuristicsSummary(), constants.GetRouteHeuristicsSummary(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteHeuristicList(),
		adapt(s.handlerHeuristicList(), constants.GetRouteHeuristicList(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteHeuristicDescriptors(),
		adapt(s.handlerHeuristicDescriptors(), constants.GetRouteHeuristicDescriptors(),
			limitMethod("GET"), s.authorization(), s.useCache(0), maxBody()))
	s.handler.Handle(constants.GetRouteDeleteHeuristic(),
		adapt(s.handlerDeleteHeuristic(), constants.GetRouteDeleteHeuristic(),
			limitMethod("POST"), s.authorization(), maxBody()))

	// Analytics
	s.handler.Handle(constants.GetRouteShortestTransactionPath(),
		adapt(s.handlerShortestTransactionPath(), constants.GetRouteShortestTransactionPath(),
			limitMethod("POST"), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(constants.GetRouteConnectionLookup(),
		adapt(s.handlerConnectionLookup(), constants.GetRouteConnectionLookup(),
			limitMethod("GET"), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(constants.GetRouteMixingActivity(),
		adapt(s.handlerMixingActivity(), constants.GetRouteMixingActivity(),
			limitMethod("POST"), s.authorization(), s.useCache(time.Minute*10), maxBody()))

	// Clusters
	s.handler.Handle(constants.GetRouteClusterLookup(),
		adapt(s.handlerClusterLookup(), constants.GetRouteClusterLookup(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteHMILookup(),
		adapt(s.handlerHMILookup(), constants.GetRouteHMILookup(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteClusterSummary(),
		adapt(s.handlerClusterSummary(), constants.GetRouteClusterSummary(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteAddCluster(),
		adapt(s.handlerAddCluster(), constants.GetRouteAddCluster(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteDeleteCluster(),
		adapt(s.handlerDeleteCluster(), constants.GetRouteDeleteCluster(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteDeleteAllClusters(),
		adapt(s.handlerDeleteAllClusters(), constants.GetRouteDeleteAllClusters(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteClusterOverview(),
		adapt(s.handlerClusterOverview(), constants.GetRouteClusterOverview(),
			limitMethod("GET"), s.authorization(), maxBody()))

	// Attributions
	s.handler.Handle(constants.GetRouteAddPrivateAttribution(),
		adapt(s.handlerAddPrivateAttribution(), constants.GetRouteAddPrivateAttribution(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteAddPublicAttribution(),
		adapt(s.handlerAddPublicAttribution(), constants.GetRouteAddPublicAttribution(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteAttributionOverview(),
		adapt(s.handlerAttributionOverview(), constants.GetRouteAttributionOverview(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteDeletePrivateAttribution(),
		adapt(s.handlerDeletePrivateAttribution(), constants.GetRouteDeletePrivateAttribution(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteDeletePublicAttribution(),
		adapt(s.handlerDeletePublicAttribution(), constants.GetRouteDeletePublicAttribution(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteDeleteAllPrivateAttributions(),
		adapt(s.handlerDeleteAllPrivateAttributions(), constants.GetRouteDeleteAllPrivateAttributions(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteSearchAttributions(),
		adapt(s.handlerSearchAttributions(), constants.GetRouteSearchAttributions(),
			limitMethod("POST"), s.authorization(), maxBody()))

	// Address Exclusions
	s.handler.Handle(constants.GetRouteAddAddressExclusions(),
		adapt(s.handlerAddAddressExclusions(), constants.GetRouteAddAddressExclusions(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteDeleteAddressExclusion(),
		adapt(s.handlerDeleteAddressExclusion(), constants.GetRouteDeleteAddressExclusion(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteDeleteAllAddressExclusions(),
		adapt(s.handlerDeleteAllAddressExclusions(), constants.GetRouteDeleteAllAddressExclusions(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteAddressExclusionOverview(),
		adapt(s.handlerAddressExclusionOverview(), constants.GetRouteAddressExclusionOverview(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteAddressExclusionStatus(),
		adapt(s.handlerGetAddressExclusionStatus(), constants.GetRouteAddressExclusionStatus(),
			limitMethod("GET"), s.authorization(), maxBody()))

	// User
	s.handler.Handle(constants.GetRouteLogin(), adapt(s.handlerLogin(),
		constants.GetRouteLogin(), limitMethod("POST"), maxBody()))
	s.handler.Handle(constants.GetRouteLogout(), adapt(s.handlerLogout(),
		constants.GetRouteLogout(), limitMethod("GET"), maxBody()))

	s.handler.Handle(constants.GetRouteCreateUser(), adapt(s.handlerCreateUser(), constants.GetRouteCreateUser(),
		limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteDeleteUser(), adapt(s.handlerDeleteUser(), constants.GetRouteDeleteUser(),
		limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteGetUsers(), adapt(s.handlerGetUsers(), constants.GetRouteGetUsers(),
		limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(constants.GetRouteModifyUser(), adapt(s.handlerModifyUser(), constants.GetRouteModifyUser(),
		limitMethod("POST"), s.authorization(), maxBody()))
}
