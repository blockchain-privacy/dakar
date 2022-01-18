package server

import (
	heuristic "backend/analytics/heuristics/transaction"
	"backend/cmd/cliutil"
	dbaddr "backend/db/address"
	"backend/db/analytics"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	dbh "backend/db/analytics/heuristics/transaction"
	dbblk "backend/db/block"
	dbstat "backend/db/status"
	dbtx "backend/db/transaction"
	dbus "backend/db/user"
	"backend/external"

	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
)

// isValidInput is a regex filter which checks if the input only consists of numbers and letters
var isValidInput = regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString

// isValid checks if user input is valid.
// Should be used to check address, transaction and block hashes, as well as block ids.
func isValid(input string) bool {
	inputLen := len(input)
	// 64 -> length of transaction hash and block hash
	if inputLen == 0 || inputLen > 64 {
		return false
	}

	// 33,34 -> address length; if smaller than it must be a block id
	if inputLen < 33 {
		// attempt to convert input to an integer; if it succeeds the input is valid.
		_, err := strconv.Atoi(input)
		return err == nil
	}

	return isValidInput(input)
}

// isLikelyBlock returns true if the given query string is likely a block hash
func isLikelyBlock(query string) bool {
	return query[0:1] == "0"
}

// isLikelyAddress returns true if the given query string is likely an address hash
func isLikelyAddress(query string) bool {
	return query[0:1] == "X" || query[0:1] == "7"
}

type prunedRPCInfo struct {
	Difficulty           float64 `json:"difficulty"`
	VerificationProgress float64 `json:"verificationprogress,omitempty"`
	SizeOnDisk           int64   `json:"size_on_disk,omitempty"`
	Pruned               bool    `json:"pruned"`
	Blocks               int32   `json:"blocks"`
}

type metaStatus struct {
	Status  dbstat.FrontendStatus `json:"status"`
	RPCInfo prunedRPCInfo         `json:"rpcinfo"`
}

type heuristicReply struct {
	Success    bool                           `json:"success"`
	Msg        string                         `json:"msg,omitempty"`
	Heuristics []dbh.FrontendHeuristic        `json:"heuristics,omitempty"`
	Status     heuristic.HeuristicQueueStatus `json:"status"`
}

type heuristicExecutionReply struct {
	Success bool                           `json:"success"`
	Msg     string                         `json:"msg,omitempty"`
	Status  heuristic.HeuristicQueueStatus `json:"status"`
}

type userReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type backendUserReply struct {
	Success bool                           `json:"success"`
	Msg     string                         `json:"msg,omitempty"`
	User    *dbus.FrontendUserBackendState `json:"user,omitempty"`
}

type shortestTransactionPathReply struct {
	Success      bool                       `json:"success"`
	Msg          string                     `json:"msg,omitempty"`
	Transactions []dbtx.FrontendTransaction `json:"transactions"`
}

type heuristicListReply struct {
	Success bool                    `json:"success"`
	Msg     string                  `json:"msg,omitempty"`
	Item    []dbh.HeuristicListItem `json:"items"`
}

type heuristicDescriptorReply struct {
	Success     bool                   `json:"success"`
	Msg         string                 `json:"msg,omitempty"`
	Descriptors []heuristic.Descriptor `json:"descriptors"`
}

type deleteHeuristicReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type queryResultType string

const typeBlock queryResultType = "block"
const typeAddr queryResultType = "addr"
const typeTx queryResultType = "tx"
const typeEmpty queryResultType = "response_empty"

// SearchResult holds result data of a block, address or transaction search request
type SearchResult struct {
	resultType queryResultType
	result     interface{}
}

// handleError conditionally logs and writes the error to the http response
func handleError(w http.ResponseWriter, err error) {
	if err == nil || w == nil {
		return
	}

	http.Error(w, "an error occurred", http.StatusInternalServerError)
	info(cliutil.ShowCallInfo(), err)
}

// buildKey build a key from the given arguments
func buildKey(route string, query string, body []byte) (key string) {
	key = route + query
	if len(body) > 0 {
		key += string(body[:])
	}

	return
}

// GetBlock searches for the hash specified in query. If a block is found the returned bool is true
func GetBlock(dgraph external.Database, query string) (SearchResult, bool, error) {
	return GetBlockWithOptions(dgraph, query, 0)
}

// GetBlockWithOptions searches for the hash specified in query. If an address is found the returned bool is true.
// It supports an offset. A maximum of 5 transactions is returned.
func GetBlockWithOptions(dgraph external.Database, query string, offset int) (SearchResult, bool, error) {
	block, err := dbblk.GetFrontendBlock(dgraph, query, offset)
	if err != nil {
		// only print error if it is not expected
		if !errors.Is(err, dbblk.ErrorBlockNotFound) {
			return SearchResult{}, false, err
		}
		return SearchResult{}, false, nil
	}

	return SearchResult{resultType: typeBlock, result: block}, true, nil
}

// GetTransaction searches for the hash specified in query. If a transaction is found the returned bool is true
func GetTransaction(dgraph external.Database, query string) (SearchResult, bool, error) {
	tx, err := dbtx.GetFrontendTransaction(dgraph, query)
	if err != nil {
		// only print error if it is not expected
		if !errors.Is(err, dbtx.ErrorTransactionNotFound) {
			return SearchResult{}, false, err
		}
		return SearchResult{}, false, nil
	}

	return SearchResult{resultType: typeTx, result: tx}, true, nil
}

// GetAddress searches for the hash specified in query. If an address is found the returned bool is true.
// A maximum of 20 elements is returned.
func GetAddress(dgraph external.Database, query string) (SearchResult, bool, error) {
	return GetAddressWithOptions(dgraph, query, dbaddr.SortAscendingByOutputTime, 0, nil)
}

// GetAddressWithOptions searches for the hash specified in query. If an address is found the returned bool is true.
// It supports sorting and setting an offset. For sorting use the constants defined in the db address module.
// A maximum of 20 elements is returned.
func GetAddressWithOptions(dgraph external.Database, query string, sortOrder int, offset int, filters []int) (SearchResult, bool, error) {
	addr, err := dbaddr.GetFrontendAddress(dgraph, query, sortOrder, offset, filters)
	if err != nil {
		// only print error if it is not expected
		if !errors.Is(err, dbaddr.ErrorAddressNotFound) {
			return SearchResult{}, false, err
		}
		return SearchResult{}, false, nil
	}

	return SearchResult{resultType: typeAddr, result: addr}, true, nil
}

// extractTokenUser extracts the tokenUser from the context
func extractTokenUser(ctx context.Context) (t tokenUser, err error) {
	userInfo := ctx.Value(middlewareContextUser)
	if userInfo == nil {
		err = errors.New("could not extract token user from context")
		return
	}
	tUser := userInfo.(tokenUser)
	if len(tUser.ID) == 0 {
		err = errors.New("invalid user id extracted from context")
		return
	}

	t = tUser

	return
}

type connectionLookupReply struct {
	Success          bool                       `json:"success"`
	Warning          bool                       `json:"warning,omitempty"`
	Msg              string                     `json:"msg,omitempty"`
	Transactions     []dbtx.FrontendTransaction `json:"transactions"`
	TransactionCount *int                       `json:"count,omitempty"`
}

type clusterLookupReply struct {
	Success  bool                         `json:"success"`
	Msg      string                       `json:"msg,omitempty"`
	Clusters []clustering.FrontendCluster `json:"clusters"`
}

type hmiLookupReply struct {
	Success        bool                            `json:"success"`
	Clusters       []clustering.FrontendHMICluster `json:"clusters,omitempty"`
	AddressCluster string                          `json:"address_cluster,omitempty"`
}

type mixingActivityReply struct {
	Success    bool                       `json:"success"`
	Activities []analytics.MixingActivity `json:"activities,omitempty"`
}

type addClusterReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type clusterOverviewReply struct {
	Success  bool                             `json:"success"`
	Msg      string                           `json:"msg,omitempty"`
	Clusters []clustering.FrontendUserCluster `json:"clusters"`
}

type deleteClusterReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type addAttributionReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type attributionOverviewReply struct {
	Success      bool                              `json:"success"`
	Msg          string                            `json:"msg,omitempty"`
	Attributions []attribution.FrontendAttribution `json:"attributions"`
}
