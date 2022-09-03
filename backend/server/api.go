package server

import (
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	dbaddr "backend/db/address"
	dbtxh "backend/db/analytics/heuristics"
	dbstat "backend/db/status"
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

		clusterUID := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteClusterReply(s.db, tUser.ID, clusterUID)
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

		attributionUID := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteAttributionReply(s.db, tUser.ID, attributionUID, false)
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

		attributionUID := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "User not found"
			info(cliutil.ShowCallInfo(), err)
		} else {
			reply = getDeleteAttributionReply(s.db, tUser.ID, attributionUID, true)
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

// API pattern: "/api/v1/createIdentity/"
// handlerCreateIdentity creates a new identity. This is an admin endpoint.
func (s *Server) handlerCreateIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getCreateIdentityReply(s.db, s.adminAuth, r)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/deleteIdentity/<identityUID>"
// handlerDeleteIdentity deletes an arbitrary identity. This is an admin endpoint.
func (s *Server) handlerDeleteIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		identityUID := path.Base(r.URL.Path)

		reply := getDeleteIdentityReply(s.db, s.adminAuth, r, identityUID)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/modifyIdentity/"
// handlerModifyIdentity modifies an arbitrary identity. This is an admin endpoint.
func (s *Server) handlerModifyIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getModifyIdentityReply(s.adminAuth, r)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			info(cliutil.ShowCallInfo(), err)
		}
	})
}

// API pattern: "/api/v1/getIdentities/"
func (s *Server) handlerGetIdentities() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getIdentitiesReply(s.db, s.adminAuth, r)

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

		reply := getConnectionLookupReply(s.db, s.worker, r.URL)

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
	s.handler.Handle(getRouteMetrics(), adapt(promhttp.Handler(), getRouteMetrics(),
		limitMethod("GET"), s.basicAuth(), maxBody()))

	// Search
	s.handler.Handle(getRouteSearch(), adapt(s.handlerSearch(), getRouteSearch(),
		limitMethod("GET"), s.useCache(time.Minute*10), maxBody()))

	// Common data
	s.handler.Handle(getRouteTransaction(),
		adapt(s.handlerDetails(GetTransaction), getRouteTransaction(),
			limitMethod("GET"), s.useCache(time.Second*0), maxBody()))
	// setting block cache time to 10 Minutes because blocks at
	// the tip get updated via adding the 'next block' reference
	s.handler.Handle(getRouteBlock(), adapt(s.handlerDetails(GetBlock), getRouteBlock(),
		limitMethod("GET"), s.useCache(time.Second*10), maxBody()))
	s.handler.Handle(getRouteAddress(), adapt(s.handlerDetails(GetAddress), getRouteAddress(),
		limitMethod("GET"), s.useCache(time.Second*10), maxBody()))

	s.handler.Handle(getRouteAddressOutputRange(),
		adapt(s.handlerAddressOutputRange(), getRouteAddressOutputRange(),
			limitMethod("POST"), s.useCache(time.Minute*10), maxBody()))

	s.handler.Handle(getRouteBlockRange(), adapt(s.handlerBlockRange(), getRouteBlockRange(),
		limitMethod("POST"), s.useCache(time.Minute*10), maxBody()))

	// Meta
	s.handler.Handle(getRouteMeta(), adapt(s.handlerMeta(), getRouteMeta(),
		limitMethod("GET"), s.authorization(), s.useCache(time.Second*10), maxBody()))

	// heuristic
	s.handler.Handle(getRouteHeuristics(),
		adapt(s.handlerHeuristics(), getRouteHeuristics(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicStatus(),
		adapt(s.handlerHeuristicStatus(), getRouteHeuristicStatus(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicDetails(),
		adapt(s.handlerHeuristicsDetails(), getRouteHeuristicDetails(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicsExecution(),
		adapt(s.handlerHeuristicsExecution(), getRouteHeuristicsExecution(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicsSummary(),
		adapt(s.handlerHeuristicsSummary(), getRouteHeuristicsSummary(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicList(),
		adapt(s.handlerHeuristicList(), getRouteHeuristicList(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicDescriptors(),
		adapt(s.handlerHeuristicDescriptors(), getRouteHeuristicDescriptors(),
			limitMethod("GET"), s.authorization(), s.useCache(0), maxBody()))
	s.handler.Handle(getRouteDeleteHeuristic(),
		adapt(s.handlerDeleteHeuristic(), getRouteDeleteHeuristic(),
			limitMethod("POST"), s.authorization(), maxBody()))

	// Analytics
	s.handler.Handle(getRouteShortestTransactionPath(),
		adapt(s.handlerShortestTransactionPath(), getRouteShortestTransactionPath(),
			limitMethod("POST"), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(getRouteConnectionLookup(),
		adapt(s.handlerConnectionLookup(), getRouteConnectionLookup(),
			limitMethod("GET"), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(getRouteMixingActivity(),
		adapt(s.handlerMixingActivity(), getRouteMixingActivity(),
			limitMethod("POST"), s.authorization(), s.useCache(time.Minute*10), maxBody()))

	// Clusters
	s.handler.Handle(getRouteClusterLookup(),
		adapt(s.handlerClusterLookup(), getRouteClusterLookup(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHMILookup(),
		adapt(s.handlerHMILookup(), getRouteHMILookup(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteClusterSummary(),
		adapt(s.handlerClusterSummary(), getRouteClusterSummary(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddCluster(),
		adapt(s.handlerAddCluster(), getRouteAddCluster(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteCluster(),
		adapt(s.handlerDeleteCluster(), getRouteDeleteCluster(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllClusters(),
		adapt(s.handlerDeleteAllClusters(), getRouteDeleteAllClusters(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteClusterOverview(),
		adapt(s.handlerClusterOverview(), getRouteClusterOverview(),
			limitMethod("GET"), s.authorization(), maxBody()))

	// Attributions
	s.handler.Handle(getRouteAddPrivateAttribution(),
		adapt(s.handlerAddPrivateAttribution(), getRouteAddPrivateAttribution(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddPublicAttribution(),
		adapt(s.handlerAddPublicAttribution(), getRouteAddPublicAttribution(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAttributionOverview(),
		adapt(s.handlerAttributionOverview(), getRouteAttributionOverview(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeletePrivateAttribution(),
		adapt(s.handlerDeletePrivateAttribution(), getRouteDeletePrivateAttribution(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeletePublicAttribution(),
		adapt(s.handlerDeletePublicAttribution(), getRouteDeletePublicAttribution(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllPrivateAttributions(),
		adapt(s.handlerDeleteAllPrivateAttributions(), getRouteDeleteAllPrivateAttributions(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteSearchAttributions(),
		adapt(s.handlerSearchAttributions(), getRouteSearchAttributions(),
			limitMethod("POST"), s.authorization(), maxBody()))

	// Address Exclusions
	s.handler.Handle(getRouteAddAddressExclusions(),
		adapt(s.handlerAddAddressExclusions(), getRouteAddAddressExclusions(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAddressExclusion(),
		adapt(s.handlerDeleteAddressExclusion(), getRouteDeleteAddressExclusion(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllAddressExclusions(),
		adapt(s.handlerDeleteAllAddressExclusions(), getRouteDeleteAllAddressExclusions(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddressExclusionOverview(),
		adapt(s.handlerAddressExclusionOverview(), getRouteAddressExclusionOverview(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddressExclusionStatus(),
		adapt(s.handlerGetAddressExclusionStatus(), getRouteAddressExclusionStatus(),
			limitMethod("GET"), s.authorization(), maxBody()))

	// User
	s.handler.Handle(getRouteGetIdentities(), adapt(s.handlerGetIdentities(), getRouteGetIdentities(),
		limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteCreateIdentity(), adapt(s.handlerCreateIdentity(), getRouteCreateIdentity(),
		limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteIdentity(), adapt(s.handlerDeleteIdentity(), getRouteDeleteIdentity(),
		limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteModifyIdentity(), adapt(s.handlerModifyIdentity(), getRouteModifyIdentity(),
		limitMethod("POST"), s.authorization(), maxBody()))
}
