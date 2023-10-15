package server

import (
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/db"
	dbtxh "backend/db/analytics/heuristics"
	dbstat "backend/db/status"
	"backend/external"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"path"
	"strconv"
	"time"
)

var (
	errorClusterSummary      = "error getting cluster summary"
	errorHeuristicSummary    = "error getting heuristic summary"
	errorHeuristics          = "error getting heuristics"
	errorHeuristicExecution  = "error executing heuristics"
	errorHeuristicDetails    = "error getting heuristic details"
	errorSpendingFingerprint = "error getting spending fingerprintScore details"
	errorInvalidSortOrder    = "error invalid sort order"
	errorInvalidFilter       = "error invalid filter"
	errorInvalidOffset       = "error invalid offset"
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

// Search godoc
//
//	@Summary	Search for blocks, addresses and transactions
//	@Tags		data
//	@Produce	json
//	@Param		hash	path		string	true	"Hash"
//	@Success	200		{object}	server.searchResponse
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/search/{hash} [get]
//
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
					warn(err)
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
			warn(err)
		}
	})
}

// Query for Address, Transaction or Block godoc
//
//	@Summary	Query for address, transaction or block
//	@Tags		data
//	@Produce	json
//	@Param		hash	path		string	true	"Hash"
//	@Success	200		{object}	server.searchResponse
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/blk/{hash} [get]
//	@Router		/address/{hash} [get]
//	@Router		/tx/{hash} [get]
//
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
			warn(err)
		}
	})
}

// Address Output Range godoc
//
//	@Summary	Get outputs of the given address
//	@Tags		data
//	@Produce	json
//	@Param		address_hash	path		string										true	"address hash"
//	@Param		offset			body		server.handlerAddressOutputRange.request	true	"output offset"
//	@Success	200				{object}	server.searchResponse
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/addressOutputRange/{address_hash} [post]
//
// API pattern: "/api/v1/addressOutputRange/<address_hash>"
func (s *Server) handlerAddressOutputRange() http.Handler {
	type request struct {
		Offset int   `json:"offset"`
		Order  int   `json:"order"`
		Filter []int `json:"filter"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := path.Base(r.URL.Path)

		reply := searchResponse{
			Type:    "response_empty",
			Payload: nil,
		}

		if isValid(queryString) {
			var addressRequest request
			addressRequest.Offset = -1
			addressRequest.Order = -1

			body, err := io.ReadAll(r.Body)
			if err != nil {
				handleError(w, err)
				return
			}

			if decodeErr := json.Unmarshal(body, &addressRequest); decodeErr != nil {
				handleError(w, cliutil.NewStackError(decodeErr))
				return
			}

			if !db.IsValidSortOrder(addressRequest.Order) {
				handleError(w, cliutil.NewStackErrorStr(errorInvalidSortOrder))
				return
			}

			if !db.IsValidFilter(addressRequest.Filter) {
				handleError(w, cliutil.NewStackErrorStr(errorInvalidFilter))
				return
			}

			if addressRequest.Offset < 0 {
				handleError(w, cliutil.NewStackErrorStr(errorInvalidOffset))
				return
			}

			data, ok, addrErr := GetAddressWithOptions(s.db, queryString,
				addressRequest.Order, addressRequest.Offset, addressRequest.Filter)
			if addrErr != nil {
				handleError(w, addrErr)
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
			warn(err)
		}
	})
}

// Block Range godoc
//
//	@Summary	Get transactions of the given block
//	@Tags		data
//	@Produce	json
//	@Param		block_hash	path		string								true	"block hash"
//	@Param		offset		body		server.handlerBlockRange.request	true	"transaction offset"
//	@Success	200			{object}	server.searchResponse
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/blkRange/{block_hash} [post]
//
// API pattern: "/api/v1/blkRange/<block_hash>"
func (s *Server) handlerBlockRange() http.Handler {
	type request struct {
		Offset int `json:"offset"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		queryString := path.Base(r.URL.Path)

		reply := searchResponse{
			Type:    "response_empty",
			Payload: nil,
		}

		if isValid(queryString) {
			var blockRequest request
			blockRequest.Offset = -1

			body, err := io.ReadAll(r.Body)
			if err != nil {
				handleError(w, err)
				return
			}

			if decodeErr := json.Unmarshal(body, &blockRequest); decodeErr != nil {
				handleError(w, cliutil.NewStackError(decodeErr))
				return
			}

			if blockRequest.Offset < 0 {
				handleError(w, cliutil.NewStackErrorStr(errorInvalidOffset))
				return
			}

			data, ok, blockErr := GetBlockWithOptions(s.db, queryString, blockRequest.Offset)
			if blockErr != nil {
				handleError(w, blockErr)
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
			warn(err)
		}
	})
}

// Server Status godoc
//
//	@Summary	Get the status of all backend modules
//	@Tags		meta
//	@Produce	json
//	@Success	200	{object}	server.metaStatus
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/meta/ [get]
//
// API pattern: "/api/v1/meta/"
func (s *Server) handlerMeta() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)
		// async request rpc info
		futureBlockchainInfo := s.client.GetBlockChainInfoAsync()

		// get data from db
		verboseStatus, err := dbstat.GetFrontendStatus(s.db)
		if err != nil {
			handleError(w, err)
			return
		}

		// receive async rpc info
		rpcInfo, err := futureBlockchainInfo.Receive()
		if err != nil {
			handleError(w, cliutil.NewStackError(err))
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

// Heuristic Summary godoc
//
//	@Summary	Heuristic Summary
//	@Tags		heuristic
//	@Produce	text/csv
//	@Param		heuristic_UID	path		string	true	"0x123"
//	@Success	200				{string}	string	"comma separated values"
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/heuristicsSummary/{heuristic_UID} [get]
//
// API pattern: "/api/v1/heuristicsSummary/<heuristic_UID>"
func (s *Server) handlerHeuristicsSummary() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		heuristicUID := path.Base(r.URL.Path)

		if heuristicUID == "." || heuristicUID == "" {
			handleError(w, cliutil.NewStackErrorStr("no heuristic UID provided"))
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

// Cluster Summary godoc
//
//	@Summary	Cluster Summary
//	@Tags		cluster
//	@Produce	text/csv
//	@Success	200	{string}	string	"comma separated values"
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/clusterSummary [get]
//
// API pattern: "/api/v1/clusterSummary"
func (s *Server) handlerClusterSummary() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		writeClusterSummary(w, r, s.db)
	})
}

//	@Summary	Add Cluster
//	@Tags		cluster
//	@Produce	json
//	@Param		separator	formData	string	true	"separator of the CSV file"
//	@Param		hasHeader	formData	bool	true	"controls whether the first line should be skiped"
//	@Param		file		formData	file	true	"the CSV file"
//	@Success	200			{object}	server.addClusterReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/addCluster [post]
//
// API pattern: "/api/v1/addCluster"
func (s *Server) handlerAddCluster() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddClusterReply(s.db, r)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Delete Cluster godoc
//
//	@Summary	Delete Cluster
//	@Tags		cluster
//	@Produce	json
//	@Param		cluster_uid	path		string	true	"0x123"
//	@Success	200			{object}	server.deleteClusterReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/deleteCluster/{cluster_uid} [get]
//
// API pattern: "/api/v1/deleteCluster/<cluster_uid>"
func (s *Server) handlerDeleteCluster() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteClusterReply

		clusterUID := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getDeleteClusterReply(s.db, tUser.ID, clusterUID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Delete All Clusters godoc
//
//	@Summary	Delete all clusters of the current user
//	@Tags		cluster
//	@Produce	json
//	@Success	200	{object}	server.deleteClusterReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/deleteAllClusters [get]
//
// API pattern: "/api/v1/deleteAllClusters"
func (s *Server) handlerDeleteAllClusters() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteClusterReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getDeleteAllClustersReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Cluster Overview godoc
//
//	@Summary	Get all custom clusters of the current user
//	@Tags		cluster
//	@Produce	json
//	@Success	200	{object}	server.clusterOverviewReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/clusterOverview [get]
//
// API pattern: "/api/v1/clusterOverview"
func (s *Server) handlerClusterOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply clusterOverviewReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getClusterOverviewReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Attribution Overview godoc
//
//	@Summary	Get all attributions of the current user
//	@Tags		attribution
//	@Produce	json
//	@Success	200	{object}	server.attributionOverviewReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/attributionOverview [get]
//
// API pattern: "/api/v1/attributionOverview"
func (s *Server) handlerAttributionOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply attributionOverviewReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getAttributionOverviewReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Add Private Attribution godoc
//
//	@Summary	Create a new attribution which is only visible for the current user
//	@Tags		attribution
//	@Produce	text/csv
//	@Param		separator	formData	string	true	"separator of the CSV file"
//	@Param		hasHeader	formData	bool	true	"controls whether the first line should be skiped"
//	@Param		file		formData	file	true	"the CSV file"
//	@Success	200			{object}	server.addAttributionReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/addPrivateAttribution [post]
//
// API pattern: "/api/v1/addPrivateAttribution"
func (s *Server) handlerAddPrivateAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddAttributionReply(s.db, r, false)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Add Public Attribution godoc
//
//	@Summary	Create a new attribution which is visible to all users
//	@Tags		attribution
//	@Produce	text/csv
//	@Param		separator	formData	string	true	"separator of the CSV file"
//	@Param		hasHeader	formData	bool	true	"controls whether the first line should be skiped"
//	@Param		file		formData	file	true	"the CSV file"
//	@Success	200			{string}	string	"comma separated values"
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/addPublicAttribution [post]
//
// API pattern: "/api/v1/addPublicAttribution"
func (s *Server) handlerAddPublicAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddAttributionReply(s.db, r, true)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Delete Private Attribution godoc
//
//	@Summary	Delete an attribution belonging to the current user
//	@Tags		attribution
//	@Produce	json
//	@Param		attribution_uid	path		string	true	"0x123"
//	@Success	200				{object}	server.deleteAttributionReply
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/deletePrivateAttribution/{attribution_uid} [get]
//
// API pattern: "/api/v1/deletePrivateAttribution/<attribution_uid>"
func (s *Server) handlerDeletePrivateAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAttributionReply

		attributionUID := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getDeleteAttributionReply(s.db, tUser.ID, attributionUID, false)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Delete Public Attribution godoc
//
//	@Summary	Delete a public attribution
//	@Tags		attribution
//	@Produce	json
//	@Param		attribution_uid	path		string	true	"0x123"
//	@Success	200				{object}	server.deleteAttributionReply
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/deletePublicAttribution/{attribution_uid} [get]
//
// API pattern: "/api/v1/deletePublicAttribution/<attribution_uid>"
func (s *Server) handlerDeletePublicAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAttributionReply

		attributionUID := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getDeleteAttributionReply(s.db, tUser.ID, attributionUID, true)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Delete All Private Attributions godoc
//
//	@Summary	Delete all attributions of the current user
//	@Tags		attribution
//	@Produce	json
//	@Success	200	{object}	server.deleteAttributionReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/deleteAllPrivateAttributions [get]
//
// API pattern: "/api/v1/deleteAllPrivateAttributions"
func (s *Server) handlerDeleteAllPrivateAttributions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAttributionReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getDeleteAllAttributionsReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Search Attributions godoc
//
//	@Summary	Search all public attributions and the attributions belonging to the current user
//	@Tags		attribution
//	@Accept		json
//	@Produce	json
//	@Param		attribution	body		server.getAttributionSearchReply.request	true	"Search query"
//	@Success	200			{object}	server.attributionOverviewReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/searchAttributions [post]
//
// API pattern: "/api/v1/searchAttributions"
func (s *Server) handlerSearchAttributions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply attributionOverviewReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getAttributionSearchReply(s.db, tUser.ID, r.Body)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Add Address Exclusion godoc
//
//	@Summary	Adds new address exclusions for the current user
//	@Tags		address exclusions
//	@Produce	text/csv
//	@Param		file	formData	file	true	"the CSV file"
//	@Success	200		{object}	server.addAddressExclusionsReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/addAddressExclusions [post]
//
// API pattern: "/api/v1/addAddressExclusions"
func (s *Server) handlerAddAddressExclusions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddAddressExclusionsReply(s.db, r)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Delete Address Exclusion godoc
//
//	@Summary	Deletes an address exclusion of the current user
//	@Tags		address exclusions
//	@Produce	json
//	@Param		address_exclusion_uid	path		string	true	"0x123"
//	@Success	200						{object}	server.deleteAddressExclusionReply
//	@Failure	500						{string}	string	"encoding error"
//	@Router		/deleteAddressExclusion/{address_exclusion_uid} [get]
//
// API pattern: "/api/v1/deleteAddressExclusion/<address_exclusion_uid>"
func (s *Server) handlerDeleteAddressExclusion() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAddressExclusionReply

		addressHash := path.Base(r.URL.Path)

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getDeleteAddressExclusionReply(s.db, tUser.ID, addressHash)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Delete All Address Exclusions godoc
//
//	@Summary	Delete all address exclusions of the current user
//	@Tags		address exclusions
//	@Produce	json
//	@Success	200	{object}	server.deleteAddressExclusionReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/deleteAllAddressExclusions [get]
//
// API pattern: "/api/v1/deleteAllAddressExclusions"
func (s *Server) handlerDeleteAllAddressExclusions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteAddressExclusionReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getDeleteAllAddressExclusionsReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Address Exclusion Overview godoc
//
//	@Summary	Address exclusion overview
//	@Tags		address exclusion
//	@Produce	json
//	@Success	200	{object}	server.addressExclusionOverviewReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/addressExclusionOverview [get]
//
// API pattern: "/api/v1/addressExclusionOverview"
func (s *Server) handlerAddressExclusionOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply addressExclusionOverviewReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getAddressExclusionOverviewReply(s.db, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Heuristic godoc
//
//	@Summary	Get all heuristic defined for a transaction
//	@Tags		heuristic
//	@Produce	json
//	@Param		hash	path		string	true	"0x123"
//	@Success	200		{object}	server.heuristicReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/heuristics/{hash} [get]
//
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
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getHeuristicReply(s.db, s.worker, txHashString, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// HMI clusters godoc
//
//	@Summary	Get HMI clusters per transaction
//	@Tags		cluster
//	@Produce	json
//	@Param		hash	path		string	true	"0x123"
//	@Success	200		{object}	server.hmiLookupReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/hmiLookup/{hash} [get]
//
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
			warn(err)
		}
	})
}

// Heuristic Status godoc
//
//	@Summary	Get the status of all heuristics per transaction
//	@Tags		heuristic
//	@Produce	json
//	@Param		hash	path		string	true	"0x123"
//	@Success	200		{object}	server.heuristicReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/heuristicStatus/{hash} [get]
//
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
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply.Success = true
			reply.Status = s.worker.GetStatus(txHashString, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Heuristic Details godoc
//
//	@Summary	Get the details of a heuristic
//	@Tags		heuristic
//	@Produce	json
//	@Accept		json
//	@Param		heuristic	body		server.handlerHeuristicsDetails.request	true	"Heuristic UID"
//	@Success	200			{object}	dbtxh.FrontendHeuristicShort
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/heuristicDetails/ [post]
//
// API pattern: "/api/v1/heuristicDetails/"
func (s *Server) handlerHeuristicsDetails() http.Handler {
	type request struct {
		HeuristicUID string `json:"uid,omitempty"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			warn(err)
			return
		}

		var heuristicRequest request

		if err = json.NewDecoder(r.Body).Decode(&heuristicRequest); err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			warn(err)
			return
		}

		if len(heuristicRequest.HeuristicUID) == 0 {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			return
		}

		frontendHeuristic, err := dbtxh.GetFrontendHeuristicByUID(s.db, heuristicRequest.HeuristicUID, tUser.ID)
		if err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			warn(err)
			return
		}

		// encoding
		if err = json.NewEncoder(w).Encode(frontendHeuristic); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Execute Heuristics godoc
//
//	@Summary		Queues the execution of heuristics for the given transaction
//	@Description	This call queues the given heuristics for the given transaction. Does not wait until the heuristic execution is finished.
//	@Tags			heuristic
//	@Produce		json
//	@Accept			json
//	@Param			hash		path		string										true	"0x123"
//	@Param			heuristic	body		server.getHeuristicExecutionReply.request	true	"Heuristics to queue"
//	@Success		200			{object}	server.heuristicExecutionReply
//	@Failure		500			{string}	string	"encoding error"
//	@Router			/executeHeuristics/{hash} [post]
//
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
			reply.Msg = msgUserNotFound
			warn(err)
		} else {
			reply = getHeuristicExecutionReply(s.db, s.worker, r.Body, txHashString, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// List Heuristics godoc
//
//	@Summary	Lists all heuristics of the current user
//	@Tags		heuristic
//	@Produce	json
//	@Success	200	{object}	server.heuristicListReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/heuristicList/ [get]
//
// API pattern: "/api/v1/heuristicList/"
func (s *Server) handlerHeuristicList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply heuristicListReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error modifying user"
			warn(err)
		} else {
			items, err := dbtxh.GetHeuristicListByUser(s.db, tUser.ID)
			if err != nil {
				warn(err)
			} else {
				reply.Success = true
				reply.Item = items
			}
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Heuristic Descriptors godoc
//
//	@Summary		Gets available heuristic descriptors
//	@Description	Returns available heuristic descriptors, which define the heuristic interface
//	@Tags			heuristic
//	@Produce		json
//	@Success		200	{object}	server.heuristicDescriptorReply
//	@Failure		500	{string}	string	"encoding error"
//	@Router			/heuristicDescriptors/ [get]
//
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
			warn(err)
		}
	})
}

// Delete Heuristic godoc
//
//	@Summary		Deletes either all heuristics or all heuristics of a transaction
//	@Description	Deletes either all heuristics of the current user or all heuristics of a transaction of the current user
//	@Tags			heuristic
//	@Produce		json
//	@Accept			json
//	@Param			heuristic	body		dbtxh.DeleteHeuristicRequest	true	"Heuristic deletion request. Set delete_all to true, only if ALL heuristic should be deleted."
//	@Success		200			{object}	server.deleteHeuristicReply
//	@Failure		500			{string}	string	"encoding error"
//	@Router			/deleteHeuristic/ [post]
//
// API pattern: "/api/v1/deleteHeuristic/"
func (s *Server) handlerDeleteHeuristic() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply deleteHeuristicReply
		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error extracting user"
			warn(err)
		} else {
			reply = getDeleteHeuristicReply(s.db, r.Body, tUser.ID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Create Identity godoc
//
//	@Summary	Create a new identity
//	@Tags		authentication
//	@Produce	json
//	@Accept		json
//	@Param		identity	body		server.getCreateIdentityReply.request	true	"Identity details"
//	@Success	200			{object}	server.identityReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/createIdentity/ [post]
//
// API pattern: "/api/v1/createIdentity/"
// handlerCreateIdentity creates a new identity. This is an admin endpoint.
func (s *Server) handlerCreateIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getCreateIdentityReply(s.db, s.adminAuth, r)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Admin Delete Identity godoc
//
//	@Summary	Delete an arbitrary identity
//	@Tags		authentication
//	@Produce	json
//	@Param		identityUID	path		string	true	"0x123"
//	@Success	200			{object}	server.identityReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/adminDeleteIdentity/{identityUID} [get]
//
// API pattern: "/api/v1/adminDeleteIdentity/<identityUID>"
// handlerAdminDeleteIdentity deletes an arbitrary identity. This is an admin endpoint.
func (s *Server) handlerAdminDeleteIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		identityUID := path.Base(r.URL.Path)

		reply := getDeleteIdentityReply(s.db, s.adminAuth, r, identityUID)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Delete Identity godoc
//
//	@Summary	Delete the identity of the current user
//	@Tags		authentication
//	@Produce	json
//	@Success	200	{object}	server.identityReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/deleteIdentity/ [get]
//
// API pattern: "/api/v1/deleteIdentity/"
// handlerDeleteIdentity deletes the calling users identity.
func (s *Server) handlerDeleteIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply identityReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error deleting user"
			warn(err)
		} else {
			reply = getDeleteIdentityReply(s.db, s.adminAuth, r, tUser.KratosID)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Modify Identity godoc
//
//	@Summary	Modify an arbitrary identity
//	@Tags		authentication
//	@Produce	json
//	@Accept		json
//	@Param		identity	body		server.getModifyIdentityReply.request	true	"Identity modification details"
//	@Success	200			{object}	server.identityReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/modifyIdentity/ [post]
//
// API pattern: "/api/v1/modifyIdentity/"
// handlerModifyIdentity modifies an arbitrary identity. This is an admin endpoint.
func (s *Server) handlerModifyIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getModifyIdentityReply(s.adminAuth, r)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Get Identities godoc
//
//	@Summary	Get all identities
//	@Tags		authentication
//	@Produce	json
//	@Success	200	{object}	server.identitiesReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/getIdentities/ [get]
//
// API pattern: "/api/v1/getIdentities/"
func (s *Server) handlerGetIdentities() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getIdentitiesReply(s.db, s.adminAuth, r)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Shortest Transaction Path godoc
//
//	@Summary	Get the shortest path between two transactions
//	@Tags		tools
//	@Produce	json
//	@Accept		json
//	@Param		transactions	body		dbtxh.ShortestTransactionPathRequest	true	"transactions between which the path should be found"
//	@Success	200				{object}	server.shortestTransactionPathReply
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/shortestTransactionPath/ [post]
//
// API pattern: "/api/v1/shortestTransactionPath/"
func (s *Server) handlerShortestTransactionPath() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getShortestTransactionPathReply(s.db, r.Body)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Connection Lookup Path godoc
//
//	@Summary	Connection lookup
//	@Tags		tools
//	@Produce	json
//	@Param		txhash	path		string	true	"Transaction hash"
//	@Param		forward	query		bool	false	"search direction"
//	@Param		t		query		int		false	"number of days to look back"	maximum(90)
//	@Success	200		{object}	server.connectionLookupReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/reverseLookup/{txhash} [get]
//
// API pattern: "/api/v1/reverseLookup/<txhash>?forward=true&t=30"
func (s *Server) handlerConnectionLookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getConnectionLookupReply(s.db, s.worker, r.URL)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Cluster Lookup Path godoc
//
//	@Summary	Get all clusters of the given address
//	@Tags		cluster
//	@Produce	json
//	@Param		addressHash	path		string	true	"Address hash"
//	@Success	200			{object}	server.clusterLookupReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/clusterLookup/{addressHash} [get]
//
// API pattern: "/api/v1/clusterLookup/<addressHash>"
func (s *Server) handlerClusterLookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		var reply clusterLookupReply

		if tUser, err := extractTokenUser(r.Context()); err != nil {
			reply.Msg = "error modifying user"
			warn(err)
		} else {
			reply = getClusterLookupReply(s.db, r.Body, tUser)
		}

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Mixing Activity Path godoc
//
//	@Summary	Mixing Activity of an address
//	@Tags		tools
//	@Produce	json
//	@Accept		json
//	@Param		heuristic	body		server.getMixingActivity.request	true	"Mixing activity request details"
//	@Success	200			{object}	server.mixingActivityReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/mixingActivity/ [post]
//
// API pattern: "/api/v1/mixingActivity/"
func (s *Server) handlerMixingActivity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getMixingActivity(s.db, r.Body)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Address Exclusion Status godoc
//
//	@Summary	Get the exclusion status of an address
//	@Tags		address exclusions
//	@Produce	json
//	@Param		address_hash	path		string	true	"address hash"
//	@Success	200				{object}	server.addressExclusionStatusReply
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/addressExclusionStatus/{address_hash} [get]
//
// API pattern: "/api/v1/addressExclusionStatus/<address_hash>"
func (s *Server) handlerGetAddressExclusionStatus() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		reply := getAddressExclusionStatusReply(r, s.db, path.Base(r.URL.Path))

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// Spending Fingerprint godoc
//
//	@Summary	Get the spending fingerprint of a transaction
//	@Tags		tools
//	@Produce	json
//	@Param		hash	path		string	true	"transaction hash"
//	@Success	200		{object}	server.spendingFingerprintReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/spendingFingerprint/{hash} [get]
//
// API pattern: "/api/v1/spendingFingerprint/<hash>"
func (s *Server) handlerSpendingFingerprint() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := path.Base(r.URL.Path)

		if !isValid(txHashString) {
			http.Error(w, errorSpendingFingerprint, http.StatusNotFound)
			return
		}

		reply := getSpendingFingerprintReply(s.db, s.worker, txHashString)

		// encoding
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			warn(err)
		}
	})
}

// setupHandlers creates endpoint handlers
func (s *Server) setupHandlers() {
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
	s.handler.Handle(getRouteSpendingFingerprint(),
		adapt(s.handlerSpendingFingerprint(), getRouteSpendingFingerprint(), limitMethod("GET"),
			s.authorization(), s.useCache(time.Minute*10), maxBody()))

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
	s.handler.Handle(getRouteAdminDeleteIdentity(), adapt(s.handlerAdminDeleteIdentity(), getRouteDeleteIdentity(),
		limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteIdentity(), adapt(s.handlerDeleteIdentity(), getRouteDeleteIdentity(),
		limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteModifyIdentity(), adapt(s.handlerModifyIdentity(), getRouteModifyIdentity(),
		limitMethod("POST"), s.authorization(), maxBody()))
}
