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
	"backend/db/workspace"
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

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
}

// setCacheHeader sets the client side caching to a third of the server side cache
func setCacheHeader(w http.ResponseWriter, duration time.Duration) {
	if duration == time.Duration(0) {
		duration = time.Hour * 24
	}
	w.Header().Set("Cache-Control", "max-age="+strconv.FormatInt(int64(duration/time.Second/3), 10))
}

// sendReply encodes the given reply into JSON and sends it
func sendReply(w http.ResponseWriter, reply any, statusCode int) {
	setCORSHeaders(w)

	// use marshalling instead of encoding (streaming), as it gives better error handling
	// and because encoding buffers all data before writing: https://github.com/golang/go/issues/7872
	// todo check if https://github.com/golang/go/discussions/63397 has been accepted, merged and released and then rework json handling.
	replyBuffer, err := json.Marshal(reply)
	if err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
		warn(cliutil.NewStackError(err))
		return
	}

	if reply == "" {
		w.Header().Set("Content-Type", "text/plain")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}

	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	w.WriteHeader(statusCode)

	if _, err := w.Write(replyBuffer); err != nil {
		// not possible to send response to client, so just log error
		warn(cliutil.NewStackError(err))
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

type searchReply struct {
	Type    queryResultType `json:"type,omitempty"`
	Payload interface{}     `json:"payload,omitempty"`
}

type metaReply struct {
	Status *dbstat.FrontendStatus `json:"status,omitempty"`
	Blocks *int64                 `json:"blocks,omitempty"`
}

type heuristicsReply struct {
	Heuristics []dbh.FrontendHeuristic `json:"heuristics,omitempty"`
}
type heuristicByWorkIDReply struct {
	Heuristic *dbh.FrontendHeuristic `json:"heuristic,omitempty"`
}

type heuristicExecutionReply struct {
	WorkID string `json:"workID"`
}

type heuristicDetailsReply struct {
	Heuristic *dbh.FrontendHeuristicShort `json:"heuristic,omitempty"`
}

type shortestTransactionPathReply struct {
	Msg          string                   `json:"msg,omitempty"`
	Transactions []db.FrontendTransaction `json:"transactions"`
}

type heuristicListReply struct {
	Item []dbh.HeuristicListItem `json:"items"`
}

type heuristicDescriptorReply struct {
	Descriptors []heuristics.Descriptor `json:"descriptors"`
}

type fingerprintScore struct {
	Score        float64 `json:"score"`
	Txhash       string  `json:"txhash"`
	SessionCount int     `json:"session_count"`
}

type spendingFingerprintReply struct {
	Msg               string             `json:"msg,omitempty"`
	SessionCount      int                `json:"session_count"`
	FingerprintScores []fingerprintScore `json:"fingerprint_scores"`
}

type addWorkspaceNodeReply struct {
	Nodes []workspace.Node `json:"nodes,omitempty"`
}

type deleteWorkspaceNodeReply struct {
	DeletedNodeUIDs []string `json:"deletedNodeUIDs,omitempty"`
}

type workspacesReply struct {
	Workspaces []workspace.FrontendWorkspace `json:"workspaces"`
}

type getWorkspaceReply struct {
	Workspace *workspace.FrontendWorkspace `json:"workspace,omitempty"`
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

// buildKey build a key from the given arguments
func buildKey(requestURI string, body []byte) string {
	if len(body) > 0 {
		requestURI += string(body)
	}

	return requestURI
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
	ID       string   `json:"uid,omitempty"`
	KratosID string   `json:"kratos_id,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

// extractTokenUser extracts a tokenUser from the context.
func extractTokenUser(ctx context.Context) (t tokenUser, err error) {
	userInfo := ctx.Value(middlewareContextUser)
	if userInfo == nil {
		err = cliutil.NewStackErrorStr("could not extract token user from context")
		return
	}

	tUser, ok := userInfo.(tokenUser)
	if !ok || len(tUser.ID) == 0 {
		err = cliutil.NewStackErrorStr("invalid user extracted from context")
		return
	}

	t = tUser

	return
}

type connectionLookupReply struct {
	Msg              string                   `json:"msg,omitempty"`
	Transactions     []db.FrontendTransaction `json:"transactions"`
	TransactionCount *int                     `json:"count,omitempty"`
}

type clusterLookupReply struct {
	Clusters []clustering.FrontendCluster `json:"clusters"`
}

type hmiLookupReply struct {
	Clusters       []clustering.FrontendHMICluster `json:"clusters,omitempty"`
	AddressCluster string                          `json:"address_cluster,omitempty"`
}

type mixingActivityReply struct {
	Msg        string                     `json:"msg,omitempty"`
	Activities []analytics.MixingActivity `json:"activities,omitempty"`
}

type clusterOverviewReply struct {
	Msg      string                           `json:"msg,omitempty"`
	Clusters []clustering.FrontendUserCluster `json:"clusters"`
}

type attributionOverviewReply struct {
	Attributions []attribution.FrontendAttribution `json:"attributions"`
}

type msgReply struct {
	Msg string `json:"msg"`
}

type addressExclusionOverviewReply struct {
	AddressHashes []string `json:"addresses"`
	Count         int64    `json:"addressCount,omitempty"`
}

type addressExclusionStatusReply struct {
	IsExclusion bool `json:"isExclusion"`
}

type identitiesReply struct {
	Identities []client.Identity `json:"identities"`
	Sessions   []client.Session  `json:"sessions"`
}

type blockReply struct {
	Block *db.FrontendBlock `json:"block"`
}

type transactionReply struct {
	Transactions []db.FrontendTransaction `json:"transactions"`
}

type addressReply struct {
	Address *db.FrontendAddress `json:"address"`
}

// isValidEmail is a regex filter which checks if the input conforms to an email string
var isValidEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]" +
	"{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$").MatchString
