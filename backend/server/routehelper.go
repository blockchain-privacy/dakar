package server

import (
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	dbh "backend/db/analytics/heuristics"
	dbstat "backend/db/status"
	dbus "backend/db/user"
	"backend/external"
	"encoding/json"
	"time"

	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	client "github.com/ory/kratos-client-go"
)

// isValidInput is a regex filter which checks if the input only consists of numbers and letters
var isValidInput = regexp.MustCompile(`^[a-zA-Z\d]*$`).MatchString

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

// writeReply encodes the given reply into JSON
func writeReply(w http.ResponseWriter, reply any) {
	if err := json.NewEncoder(w).Encode(reply); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
		warn(err)
	}
}

// isLikelyBlock returns true if the given query string is likely a block hash
func isLikelyBlock(query string) bool {
	return query[0:1] == "0"
}

// isLikelyAddress returns true if the given query string is likely an address hash
func isLikelyAddress(query string) bool {
	return query[0:1] == "X" || query[0:1] == "7"
}

type searchResponse struct {
	Type    queryResultType `json:"type,omitempty"`
	Payload interface{}     `json:"payload,omitempty"`
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
	Success    bool                               `json:"success"`
	Msg        string                             `json:"msg,omitempty"`
	Heuristics []dbh.TransformedFrontendHeuristic `json:"heuristics,omitempty"`
	Status     heuristics.HeuristicQueueStatus    `json:"status"`
}

type heuristicExecutionReply struct {
	Success bool                            `json:"success"`
	Msg     string                          `json:"msg,omitempty"`
	Status  heuristics.HeuristicQueueStatus `json:"status"`
}

type identityReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type shortestTransactionPathReply struct {
	Success      bool                     `json:"success"`
	Msg          string                   `json:"msg,omitempty"`
	Transactions []db.FrontendTransaction `json:"transactions"`
}

type heuristicListReply struct {
	Success bool                    `json:"success"`
	Msg     string                  `json:"msg,omitempty"`
	Item    []dbh.HeuristicListItem `json:"items"`
}

type heuristicDescriptorReply struct {
	Success     bool                    `json:"success"`
	Msg         string                  `json:"msg,omitempty"`
	Descriptors []heuristics.Descriptor `json:"descriptors"`
}

type deleteHeuristicReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type fingerprintScore struct {
	Score        float64 `json:"score"`
	Txhash       string  `json:"txhash"`
	SessionCount int     `json:"session_count,omitempty"`
}

type spendingFingerprintReply struct {
	Success           bool               `json:"success"`
	Warning           bool               `json:"warning,omitempty"`
	Msg               string             `json:"msg,omitempty"`
	SessionCount      int                `json:"session_count,omitempty"`
	FingerprintScores []fingerprintScore `json:"fingerprint_scores,omitempty"`
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
	warn(err)
}

// buildKey build a key from the given arguments
func buildKey(route string, query string, body []byte) (key string) {
	key = route + query
	if len(body) > 0 {
		key += string(body)
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
	block, err := db.GetFrontendBlock(dgraph, query, offset)
	if err != nil {
		// only print error if it is not expected
		if !errors.Is(err, db.ErrBlockNotFound) {
			return SearchResult{}, false, err
		}
		return SearchResult{}, false, nil
	}

	return SearchResult{resultType: typeBlock, result: block}, true, nil
}

// GetTransaction searches for the hash specified in query. If a transaction is found the returned bool is true
func GetTransaction(dgraph external.Database, query string) (SearchResult, bool, error) {
	tx, err := db.GetFrontendTransaction(dgraph, query)
	if err != nil {
		// only print error if it is not expected
		if !errors.Is(err, db.ErrTransactionNotFound) {
			return SearchResult{}, false, err
		}
		return SearchResult{}, false, nil
	}

	return SearchResult{resultType: typeTx, result: tx}, true, nil
}

// GetAddress searches for the hash specified in query. If an address is found the returned bool is true.
// A maximum of 20 elements is returned.
func GetAddress(dgraph external.Database, query string) (SearchResult, bool, error) {
	return GetAddressWithOptions(dgraph, query, db.SortAscendingByOutputTime, 0, nil)
}

// GetAddressWithOptions searches for the hash specified in query. If an address is found the returned bool is true.
// It supports sorting and setting an offset. For sorting use the constants defined in the db address module.
// A maximum of 20 elements is returned.
func GetAddressWithOptions(dgraph external.Database, query string, sortOrder int,
	offset int, filters []int) (SearchResult, bool, error) {
	addr, err := db.GetFrontendAddress(dgraph, query, sortOrder, offset, filters)
	if err != nil {
		// only print error if it is not expected
		if !errors.Is(err, db.ErrAddressNotFound) {
			return SearchResult{}, false, err
		}
		return SearchResult{}, false, nil
	}

	return SearchResult{resultType: typeAddr, result: addr}, true, nil
}

type tokenUser struct {
	ID       string      `json:"uid,omitempty"`
	KratosID string      `json:"kratos_id,omitempty"`
	Roles    []dbus.Role `json:"roles,omitempty"`
}

// extractTokenUser extracts the tokenUser from the context.
// Returns an error if no user data could be extracted
func extractTokenUser(ctx context.Context) (t tokenUser, err error) {
	userInfo := ctx.Value(middlewareContextUser)
	if userInfo == nil {
		err = cliutil.NewStackErrorStr("could not extract token user from context")
		return
	}
	tUser := userInfo.(tokenUser)
	if len(tUser.ID) == 0 {
		err = cliutil.NewStackErrorStr("invalid user id extracted from context")
		return
	}

	t = tUser

	return
}

type connectionLookupReply struct {
	Success          bool                     `json:"success"`
	Warning          bool                     `json:"warning,omitempty"`
	Msg              string                   `json:"msg,omitempty"`
	Transactions     []db.FrontendTransaction `json:"transactions"`
	TransactionCount *int                     `json:"count,omitempty"`
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
	Msg        string                     `json:"msg,omitempty"`
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

type deleteAttributionReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type addAddressExclusionsReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type addressExclusionOverviewReply struct {
	Success       bool     `json:"success"`
	Msg           string   `json:"msg,omitempty"`
	AddressHashes []string `json:"addresses"`
	Count         int64    `json:"addressCount,omitempty"`
}

type deleteAddressExclusionReply struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg,omitempty"`
}

type addressExclusionStatusReply struct {
	Success     bool   `json:"success"`
	IsExclusion bool   `json:"isExclusion"`
	Msg         string `json:"msg,omitempty"`
}

type identitiesReply struct {
	Success    bool                            `json:"success"`
	Users      []dbus.FrontendUserBackendState `json:"users"`
	Identities []client.Identity               `json:"identities"`
	Sessions   []client.Session                `json:"sessions"`
}

// isValidEmail is a regex filter which checks if the input conforms to an email string
var isValidEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]" +
	"{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$").MatchString
