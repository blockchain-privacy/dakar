package server

import (
	"backend/analytics"
	analyticsClustering "backend/analytics/clustering"
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/db"
	dbAnalytics "backend/db/analytics"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"backend/db/analytics/exclusion"
	dbHeuristic "backend/db/analytics/heuristics"
	dbstat "backend/db/status"
	dbus "backend/db/user"
	dbwork "backend/db/workspace"
	"backend/external"
	"backend/worker"
	"backend/workspace"
	"io"
	"strings"
	"time"

	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	ory "github.com/ory/kratos-client-go"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
)

// getSearchReply searches for the given query in the database
func getSearchReply(dgraph external.Database, query string) (searchReply, int) {
	reply := searchReply{
		Type:    typeEmpty,
		Payload: nil,
	}

	if !isValid(query) {
		return reply, http.StatusBadRequest
	}
	searchOrder := []func(external.Database, string) (SearchResult, bool, error){GetTransaction, GetAddress, GetBlock}

	if isLikelyBlock(query) {
		searchOrder = []func(external.Database, string) (SearchResult, bool, error){GetBlock, GetTransaction, GetAddress}
	} else if isLikelyAddress(query) {
		searchOrder = []func(external.Database, string) (SearchResult, bool, error){GetAddress, GetTransaction, GetBlock}
	}

	status := http.StatusOK

	// iterate over db access functions
	for _, fn := range searchOrder {
		data, ok, err := fn(dgraph, query)
		if err != nil {
			status = http.StatusInternalServerError
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

	if reply.Type == typeEmpty {
		status = http.StatusNotFound
	}

	return reply, status
}

func getAddressReply(dgraph external.Database, query string) (reply addressReply, status int) {
	if !isValid(query) {
		return reply, http.StatusBadRequest
	}

	addr, err := db.GetFrontendAddress(dgraph, query, db.SortAscendingByOutputTime, 0, nil)
	if err != nil {
		// only print error if it is not expected
		if errors.Is(err, db.ErrAddressNotFound) {
			status = http.StatusNotFound
			return
		}

		warn(err)
		status = http.StatusInternalServerError
		return
	}

	reply.Address = &addr

	return
}

func getBlockReply(r *http.Request, dgraph external.Database, query string) (reply blockReply, status int) {
	if !isValid(query) {
		return reply, http.StatusBadRequest
	}

	offset := 0
	strOffset := r.URL.Query().Get("offset")

	if strOffset != "" {
		var err error
		offset, err = strconv.Atoi(strOffset)
		if err != nil {
			status = http.StatusBadRequest
			return
		}
	}

	block, err := db.GetFrontendBlock(dgraph, query, offset)
	if err != nil {
		if errors.Is(err, db.ErrBlockNotFound) {
			status = http.StatusNotFound
			return
		}

		warn(err)
		status = http.StatusInternalServerError
		return
	}

	reply.Block = &block

	return
}

func getTransactionReply(dgraph external.Database, query string) (reply transactionReply, status int) {
	if !isValid(query) {
		return reply, http.StatusBadRequest
	}

	transactions, err := db.GetFrontendTransaction(dgraph, query)
	if err != nil {
		// only print error if it is not expected
		if errors.Is(err, db.ErrTransactionNotFound) {
			status = http.StatusNotFound
			return
		}
		warn(err)
		status = http.StatusInternalServerError
		return
	}

	reply.Transactions = transactions

	return
}

// getAddressOutputRangeReply searches for the given address hash in the database with the options stored in the request
func getAddressOutputRangeReply(r *http.Request, dgraph external.Database, addressHash string) (searchReply, int) {
	reply := searchReply{
		Type:    typeEmpty,
		Payload: nil,
	}

	type request struct {
		Offset int   `json:"offset"`
		Order  int   `json:"order"`
		Filter []int `json:"filter"`
	}

	if !isValid(addressHash) {
		return reply, http.StatusBadRequest
	}

	var addressRequest request
	addressRequest.Offset = -1
	addressRequest.Order = -1

	if decodeErr := json.NewDecoder(r.Body).Decode(&addressRequest); decodeErr != nil {
		return reply, http.StatusBadRequest
	}

	if !db.IsValidSortOrder(addressRequest.Order) {
		return reply, http.StatusBadRequest
	}

	if !db.IsValidFilter(addressRequest.Filter) {
		return reply, http.StatusBadRequest
	}

	if addressRequest.Offset < 0 {
		return reply, http.StatusBadRequest
	}

	status := http.StatusOK
	data, ok, addrErr := GetAddressWithOptions(dgraph, addressHash,
		addressRequest.Order, addressRequest.Offset, addressRequest.Filter)
	if addrErr != nil {
		status = http.StatusInternalServerError
	} else if ok {
		reply.Payload = data.result
		reply.Type = data.resultType
	} else {
		status = http.StatusNotFound
	}

	return reply, status
}

func getMetaReply(dgraph external.Database, rpcClient external.RPCClient) (metaReply, int) {
	var reply metaReply

	// get block info
	blocks, err := rpcClient.GetBlockCount()
	if err != nil {
		warn(cliutil.NewStackError(err))
		return reply, http.StatusInternalServerError
	}

	// get data from db
	verboseStatus, err := dbstat.GetFrontendStatus(dgraph)
	if err != nil {
		warn(err)
		return reply, http.StatusInternalServerError
	}

	// set response struct
	return metaReply{
		Status: &verboseStatus,
		Blocks: &blocks,
	}, http.StatusOK
}

func getIdentitiesReply(adminAuth *ory.APIClient, r *http.Request) (reply identitiesReply, status int) {
	// get identity list
	identities, response, err := adminAuth.IdentityAPI.ListIdentities(r.Context()).Execute() //nolint:bodyclose
	if err != nil {
		status = http.StatusInternalServerError
		warn(cliutil.NewStackError(err))
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	sessions, response, err := adminAuth.IdentityAPI.ListSessions(r.Context()).
		Active(true).Expand([]string{"Identity"}).PageSize(100).Execute() //nolint:bodyclose
	if err != nil {
		status = http.StatusInternalServerError
		warn(cliutil.NewStackError(err))
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	var activeSession []ory.Session
	for _, session := range sessions {
		if *session.Active {
			activeSession = append(activeSession, session)
		}
	}

	reply.Identities = identities
	reply.Sessions = activeSession

	return
}

func getHeuristicByWorkIDReply(r *http.Request, dgraph external.Database,
	worker *worker.Worker) (reply heuristicByWorkIDReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		ID           string `json:"id,omitempty"`
		WorkspaceUID string `json:"workspaceUID,omitempty"`
	}

	var workRequest request

	if err := json.NewDecoder(r.Body).Decode(&workRequest); err != nil {
		status = http.StatusBadRequest
		warn(err)
		return
	}

	if workRequest.ID == "" || workRequest.WorkspaceUID == "" {
		status = http.StatusBadRequest
		return
	}

	// convert to integer
	workID, err := strconv.Atoi(workRequest.ID)
	if err != nil {
		status = http.StatusBadRequest
		warn(err)
		return
	}

	uid, err := worker.GetFinishedHeuristicUID(workID, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	// heuristic not finished executing
	if uid == "" {
		return
	}

	h, err := dbHeuristic.GetBasicFrontendHeuristic(dgraph, uid, tUser.ID, workRequest.WorkspaceUID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Heuristic = &h

	return
}

func getHeuristicDetailsReply(r *http.Request, dgraph external.Database) (reply heuristicDetailsReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		HeuristicUID string `json:"heuristicUID,omitempty"`
		WorkspaceUID string `json:"workspaceUID,omitempty"`
	}

	var heuristicRequest request

	if err := json.NewDecoder(r.Body).Decode(&heuristicRequest); err != nil {
		status = http.StatusBadRequest
		warn(cliutil.NewStackError(err))
		return
	}

	if heuristicRequest.HeuristicUID == "" || heuristicRequest.WorkspaceUID == "" {
		status = http.StatusBadRequest
		return
	}

	heuristic, err := dbHeuristic.GetFrontendHeuristicByUID(dgraph, heuristicRequest.HeuristicUID, tUser.ID, heuristicRequest.WorkspaceUID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Heuristic = &heuristic

	return
}

func getHeuristicExecutionReply(r *http.Request, dgraph external.Database, worker *worker.Worker,
	workspaceMutex *workspace.Mutex) (reply heuristicExecutionReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		NewHeuristic dbHeuristic.DatabaseHeuristicRequest `json:"newHeuristic,omitempty"`
	}

	var heuristicRequest request

	if err := json.NewDecoder(r.Body).Decode(&heuristicRequest); err != nil {
		status = http.StatusBadRequest
		warn(cliutil.NewStackError(err))
		return
	}

	work, err := heuristics.ConstructExecutors(heuristicRequest.NewHeuristic, tUser.ID, workspaceMutex)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.WorkID, err = workspace.AddHeuristic(dgraph, worker, workspaceMutex, heuristicRequest.NewHeuristic, tUser.ID, work)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

// getShortestTransactionPathReply searches for the shortest path between two transactions
func getShortestTransactionPathReply(dgraph external.Database, body io.Reader) (reply shortestTransactionPathReply, status int) {
	type request struct {
		// From is the starting point of the shortest path lookup
		From string `json:"from,omitempty"`
		// To is the end point of the shortest path lookup
		To string `json:"to,omitempty"`
		// IncludePrivacyTransactions determines if privacy transactions
		// should be considered when doing the shortest path lookup
		IncludePrivacyTransactions bool `json:"includePrivacyTransactions"`
		// AnyDirection determines the search direction of the shortest transaction path query
		// True: Both inputs and outputs are traversed
		// False: Only inputs are traversed
		AnyDirection bool `json:"anyDirection"`
	}

	// parse request
	var req request
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		return
	}

	if req.From == req.To {
		status = http.StatusBadRequest
		reply.Msg = "Transaction hashes are equal"
		return
	}

	fromBlockID, err := db.GetTransactionBlockID(dgraph, req.From)
	if err != nil {
		if errors.Is(err, db.ErrTransactionNotFound) {
			status = http.StatusNotFound
			reply.Msg = "Transaction " + req.From + " does not exist"
			return
		}

		status = http.StatusInternalServerError
		warn(err)
		return
	}

	toBlockID, err := db.GetTransactionBlockID(dgraph, req.To)
	if err != nil {
		if errors.Is(err, db.ErrTransactionNotFound) {
			status = http.StatusBadRequest
			reply.Msg = "error transaction" + req.To + " does not exist"
			return
		}

		status = http.StatusInternalServerError
		warn(err)
		return
	}

	anyDirection := req.AnyDirection

	if fromBlockID == toBlockID {
		// set anyDirection to true, as the direction can not be calculated from the block ids
		// and as the transactions are in the same block the query should be very quick
		anyDirection = true
	}

	oldTx := req.From
	youngTx := req.To

	if !req.AnyDirection {
		// switch transactions if necessary, so we are searching in the right direction
		if toBlockID > fromBlockID {
			oldTx = req.To
			youngTx = req.From
		}
	}

	// do 'shortest transaction path' lookup
	txs, err := dbAnalytics.GetShortestTransactionPathAnyDirection(dgraph, oldTx, youngTx,
		req.IncludePrivacyTransactions, anyDirection)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if len(txs) == 0 {
		reply.Msg = "No path found"
	} else {
		reply.Transactions = txs
	}

	return
}

// getConnectionLookupReply returns the result of a reverse lookup
func getConnectionLookupReply(dgraph external.Database, worker *worker.Worker,
	transactionHash string, urlHandle *url.URL) (reply connectionLookupReply, status int) {
	if !worker.IsReady() {
		reply.Msg = "Server is not ready to receive connection lookups. Please try again later."
		return
	}

	urlValues := urlHandle.Query()

	isLookupForward, err := strconv.ParseBool(urlValues.Get("forward"))
	if err != nil {
		status = http.StatusBadRequest
		reply.Msg = "invalid value for parameter 'forward'"
		return
	}

	numDays, err := strconv.Atoi(urlValues.Get("t"))
	if err != nil {
		status = http.StatusBadRequest
		reply.Msg = "invalid value for parameter 't'"
		return
	}

	if numDays > 90 {
		status = http.StatusBadRequest
		reply.Msg = "value for parameter 't' must be lower of equal then 90"
		return
	}

	if numDays <= 0 {
		status = http.StatusBadRequest
		reply.Msg = "value for parameter 't' must be greater than 0"
		return
	}

	lookBackTime := time.Duration(numDays)

	uid, err := db.GetTransactionUID(dgraph, transactionHash)
	if err != nil {
		if errors.Is(err, db.ErrTransactionNotFound) {
			status = http.StatusNotFound
			reply.Msg = "Transaction " + transactionHash + " does not exist"
			return
		}

		status = http.StatusInternalServerError
		warn(err)
		return
	}

	var endpoints map[string]bool
	if isLookupForward {
		endpoints, err = worker.ForwardLookup(uid, time.Hour*24*lookBackTime)
		if err != nil {
			status = http.StatusInternalServerError
			warn(err)
			return
		}
	} else {
		endpoints, err = worker.ReverseLookup(uid, time.Hour*24*lookBackTime)
		if err != nil {
			status = http.StatusInternalServerError
			warn(err)
			return
		}
	}

	// reply with the first 30 endpoints
	const numOutputNodes = 30
	transactionUids := make([]string, 0, numOutputNodes)
	i := 0
	for k := range endpoints {
		transactionUids = append(transactionUids, k)
		if i == numOutputNodes {
			break
		}
		i++
	}

	frontendTransactions, err := db.GetFrontendTransactionsByUID(dgraph, transactionUids)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	endpointCount := len(endpoints)
	reply.TransactionCount = &endpointCount
	reply.Transactions = frontendTransactions

	return
}

// getClusterLookupReply returns the result of a cluster lookup
func getClusterLookupReply(r *http.Request, dgraph external.Database) (reply clusterLookupReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	addressHash := r.PathValue("hash")
	if !isValid(addressHash) {
		status = http.StatusBadRequest
		return
	}

	const maxAddresses = 30
	clusters, err := clustering.GetClusters(dgraph, addressHash, maxAddresses, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Clusters = clusters

	return
}

// getHMILookupReply returns all hmi clusters connected to the given address hash
func getHMILookupReply(dgraph external.Database, addressHash string) (reply hmiLookupReply, status int) {
	if !isValid(addressHash) {
		status = http.StatusBadRequest
		return
	}

	addressCluster, clusters, err := clustering.GetHMIClusters(dgraph, addressHash)
	if err != nil {
		warn(err)
		status = http.StatusInternalServerError
		return
	}

	reply.Clusters = clusters
	reply.AddressCluster = addressCluster

	return
}

// writeHeuristicReport writes heuristic data in CSV format
func writeHeuristicReport(w http.ResponseWriter, r *http.Request, dgraph external.Database) {
	setCORSHeaders(w)

	const errReport = "error getting heuristic report"

	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		http.Error(w, errReport, http.StatusBadRequest)
		warn(err)
		return
	}

	type request struct {
		HeuristicUID string `json:"heuristicUID,omitempty"`
		WorkspaceUID string `json:"workspaceUID,omitempty"`
	}

	var heuristicRequest request
	if err := json.NewDecoder(r.Body).Decode(&heuristicRequest); err != nil {
		http.Error(w, errReport, http.StatusBadRequest)
		warn(cliutil.NewStackError(err))
		return
	}

	if heuristicRequest.HeuristicUID == "" || heuristicRequest.WorkspaceUID == "" {
		http.Error(w, errReport, http.StatusBadRequest)
		return
	}

	cHeuristic, err := dbHeuristic.GetFrontendHeuristicByUID(dgraph, heuristicRequest.HeuristicUID,
		tUser.ID, heuristicRequest.WorkspaceUID)
	if err != nil {
		http.Error(w, errReport, http.StatusInternalServerError)
		warn(err)
		return
	}

	if cHeuristic.UID == "" {
		http.Error(w, errReport, http.StatusNotFound)
		return
	}

	// headers for streaming data to client
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", heuristicRequest.HeuristicUID))
	w.Header().Set("Content-Type", "text/csv")

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	header := []string{"cluster ID", "attributions", "origin transaction hash",
		"origin timestamp", "destination count"}

	if err = csvWriter.Write(header); err != nil {
		http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
		warn(cliutil.NewStackError(err))
		return
	}

	var clusterCount int
	for _, c := range cHeuristic.Clusters {
		clusterCount++
		var attributions string

		for i, a := range c.Attributions {
			attributions += a.Tag

			if i+1 < len(c.Attributions) {
				attributions += ","
			}
		}

		for _, transaction := range c.Transactions {
			row := []string{strconv.Itoa(clusterCount), attributions, transaction.Hash,
				transaction.Timestamp, strconv.Itoa(transaction.DestinationCount)}

			if err = csvWriter.Write(row); err != nil {
				// communication with client is not possible, can only log error
				// this is because as soon as we write the CSV header, the HTTP response status is also sent
				warn(cliutil.NewStackError(err))
				return
			}
		}
		csvWriter.Flush()
	}

	csvWriter.Flush()
}

// writeClusterReport writes cluster data in CSV format
func writeClusterReport(w http.ResponseWriter, r *http.Request, dgraph external.Database) {
	setCORSHeaders(w)
	addressHash := r.PathValue("hash")
	if !isValid(addressHash) {
		http.Error(w, "no address hash provided", http.StatusNotFound)
		return
	}

	const errReport = "error getting cluster report"

	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		http.Error(w, errReport, http.StatusNotFound)
		warn(err)
		return
	}

	clusters, err := clustering.GetClusters(dgraph, addressHash, 0, tUser.ID)
	if err != nil {
		http.Error(w, errReport, http.StatusNotFound)
		warn(err)
		return
	}

	if len(clusters) == 0 {
		http.Error(w, errReport, http.StatusNotFound)
		return
	}

	// headers for streaming data to client
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=cluster_lookup_%s.csv",
		time.Now().Format("2006-01-02T15:04:05")))
	w.Header().Set("Content-Type", "text/csv")

	// somehow both content-length and transfer-encoding headers are both set, so one must be removed
	// w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	header := []string{"cluster type", "last cluster update (transaction)", "last cluster update (timestamp)",
		"address hash", "output count", "unspent output count"}

	if err = csvWriter.Write(header); err != nil {
		http.Error(w, "error writing to file", http.StatusInternalServerError)
		warn(cliutil.NewStackError(err))
		return
	}

	for _, c := range clusters {
		for _, a := range c.Addresses {
			var row []string
			// per heuristic information
			row = append(row, string(c.Type))
			row = append(row, c.TransactionHash)
			row = append(row, c.Timestamp.Format(time.RFC3339))
			row = append(row, a.AddressHash)
			row = append(row, strconv.Itoa(a.OutputCount))
			row = append(row, strconv.Itoa(a.OutputCount-a.SpentOutputCount))

			if err = csvWriter.Write(row); err != nil {
				// communication with client is not possible, can only log error
				// this is because as soon as we write the CSV header, the HTTP response status is also sent
				warn(cliutil.NewStackError(err))
				return
			}
		}
		csvWriter.Flush()
	}
}

// getMixingActivity returns the result of a mixing activity lookup
func getMixingActivity(dgraph external.Database, body io.Reader) (reply mixingActivityReply, status int) {
	type request struct {
		// AddressHash is the address hash for which the lookup will be done
		AddressHash string `json:"addressHash,omitempty"`
		// IsClusterLookup determines if all addresses of the cluster will be considered
		IsClusterLookup bool `json:"isClusterLookup,omitempty"`
	}
	var req request
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		warn(cliutil.NewStackError(err))
		return
	}
	const maxAddressCount = 2000
	if req.IsClusterLookup {
		addressCount, err := clustering.GetClusterAddressCount(dgraph, req.AddressHash)
		if err != nil {
			status = http.StatusInternalServerError
			warn(err)
			return
		}

		if addressCount > maxAddressCount {
			reply.Msg = "too_many_addresses"
			status = http.StatusBadRequest
			return
		}
	}

	activities, err := dbAnalytics.GetMixingActivity(dgraph, req.AddressHash, req.IsClusterLookup)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Activities = activities

	return
}

const (
	CsvEmptyHeader       = "empty_header_flag"
	CsvInvalidSeparator  = "unsupported_separator"
	CsvInvalidFieldCount = "file_invalid_field_count"
	CsvNoData            = "file_no_data"
	CsvInvalidData       = "file_invalid_data"
	CsvReadError         = "file_reading_error"
	CsvTooManyAddresses  = "file_too_many_addresses"
	CsvShallowCluster    = "file_shallow_cluster"
	CsvErrorImporting    = "file_error_importing"
)

func getAddClusterReply(dgraph external.Database, r *http.Request) (reply msgReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if err := r.ParseMultipartForm(maxBodySize); err != nil {
		status = http.StatusBadRequest
		return
	}

	separator := r.FormValue("separator")
	if separator == "" {
		reply.Msg = CsvInvalidSeparator
		status = http.StatusBadRequest
		return
	}

	var rSeparator rune
	if separator != ";" && separator != "," {
		reply.Msg = CsvInvalidSeparator
		status = http.StatusBadRequest
		return
	}
	rSeparator = []rune(separator)[0]

	headerFlag := r.FormValue("hasHeader")
	if headerFlag == "" {
		reply.Msg = CsvEmptyHeader
		status = http.StatusBadRequest
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		reply.Msg = CsvReadError
		status = http.StatusBadRequest
		return
	}

	defer func(file multipart.File) {
		if err := file.Close(); err != nil {
			warn(cliutil.NewStackErrorf("error closing CSV-file: %w", err))
		}
	}(file)

	csvReader := csv.NewReader(file)
	csvReader.ReuseRecord = true
	csvReader.Comma = rSeparator
	csvReader.FieldsPerRecord = 2
	var line []string

	var addresses []analyticsClustering.ExternalClusterItem
	var index int
	for ; ; index++ {
		line, err = csvReader.Read()
		if err != nil {
			if errors.Is(err, csv.ErrFieldCount) {
				reply.Msg = CsvInvalidFieldCount
				status = http.StatusBadRequest
				return
			} else if !errors.Is(err, io.EOF) {
				reply.Msg = CsvInvalidData
				status = http.StatusBadRequest
				return
			}
			break
		}

		if index == 0 && headerFlag == "true" {
			continue
		}

		newAddress := analyticsClustering.ExternalClusterItem{
			ClusterID:   strings.TrimSpace(line[0]),
			AddressHash: strings.TrimSpace(line[1]),
		}

		if newAddress.ClusterID == "" || newAddress.AddressHash == "" {
			reply.Msg = CsvInvalidData
			status = http.StatusBadRequest
			return
		}

		addresses = append(addresses, newAddress)
	}

	if len(addresses) == 0 {
		reply.Msg = CsvNoData
		status = http.StatusBadRequest
		return
	}

	if err := analyticsClustering.ImportCluster(dgraph, addresses, tUser.ID); err != nil {
		switch {
		case errors.Is(err, analyticsClustering.ErrTooManyAddresses):
			reply.Msg = CsvTooManyAddresses
			status = http.StatusBadRequest
		case errors.Is(err, analyticsClustering.ErrShallowCluster):
			reply.Msg = CsvShallowCluster
			status = http.StatusBadRequest
		case errors.Is(err, analyticsClustering.ErrNonExistentAddress):
			reply.Msg = CsvInvalidData
			status = http.StatusBadRequest
		default:
			reply.Msg = CsvErrorImporting
			status = http.StatusInternalServerError
			warn(err)
		}

		return
	}

	return
}

func getAddAttributionReply(r *http.Request, dgraph external.Database, isPublic bool) (reply msgReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if err := r.ParseMultipartForm(maxBodySize); err != nil {
		return
	}

	separator := r.FormValue("separator")
	if separator == "" {
		reply.Msg = CsvInvalidSeparator
		status = http.StatusBadRequest
		return
	}

	var rSeparator rune
	if separator != ";" && separator != "," {
		reply.Msg = CsvInvalidSeparator
		status = http.StatusBadRequest
		return
	}
	rSeparator = []rune(separator)[0]

	headerFlag := r.FormValue("hasHeader")
	if headerFlag == "" {
		reply.Msg = CsvEmptyHeader
		status = http.StatusBadRequest
		return
	}

	// Get handler for filename, size and headers
	file, _, err := r.FormFile("file")
	if err != nil {
		reply.Msg = CsvReadError
		status = http.StatusBadRequest
		return
	}

	defer func(file multipart.File) {
		if err := file.Close(); err != nil {
			warn(cliutil.NewStackErrorf("error closing CSV-file: %w", err))
		}
	}(file)

	csvReader := csv.NewReader(file)
	csvReader.ReuseRecord = true
	csvReader.Comma = rSeparator
	csvReader.FieldsPerRecord = 5
	var line []string

	var attributions []analytics.Attribution
	var index int
	for ; ; index++ {
		line, err = csvReader.Read()
		if err != nil {
			if errors.Is(err, csv.ErrFieldCount) {
				reply.Msg = CsvInvalidFieldCount
				status = http.StatusBadRequest
				return
			} else if !errors.Is(err, io.EOF) {
				reply.Msg = CsvInvalidData
				status = http.StatusBadRequest
				return
			}
			break
		}

		if index == 0 && headerFlag == "true" {
			continue
		}

		newAttribution := analytics.Attribution{
			AddressHash: strings.TrimSpace(line[0]),
			Tag:         strings.TrimSpace(line[1]),
			Description: strings.TrimSpace(line[2]),
			Source:      strings.TrimSpace(line[3]),
			Category:    strings.TrimSpace(line[4]),
		}

		if newAttribution.AddressHash == "" || newAttribution.Tag == "" {
			reply.Msg = CsvInvalidData
			status = http.StatusBadRequest
			return
		}

		attributions = append(attributions, newAttribution)
	}

	if len(attributions) == 0 {
		reply.Msg = CsvNoData
		status = http.StatusBadRequest
		return
	}

	if err := analytics.ImportAttribution(dgraph, attributions, tUser.ID, isPublic); err != nil {
		switch {
		case errors.Is(err, analytics.ErrTooManyAddresses):
			reply.Msg = CsvTooManyAddresses
			status = http.StatusBadRequest
		case errors.Is(err, analytics.ErrNonExistentAddress):
			reply.Msg = CsvInvalidData
			status = http.StatusBadRequest
		default:
			reply.Msg = CsvErrorImporting
			status = http.StatusInternalServerError
			warn(err)
		}

		return
	}

	return
}

func getClusterOverviewReply(r *http.Request, dgraph external.Database) (reply clusterOverviewReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	clusters, err := clustering.GetUserClusters(dgraph, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		reply.Msg = "no clusters found"
		warn(err)
		return
	}

	reply.Clusters = clusters

	return
}

func getDeleteClusterReply(r *http.Request, dgraph external.Database) (reply msgReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	clusterUID := r.PathValue("uid")
	if clusterUID == "" {
		status = http.StatusBadRequest
		reply.Msg = "cluster uid was not set"
		return
	}

	if err := clustering.DeleteCluster(dgraph, tUser.ID, clusterUID); err != nil {
		reply.Msg = "could not delete cluster"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getDeleteAllClustersReply(r *http.Request, dgraph external.Database) (reply msgReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if err := clustering.DeleteAllClusters(dgraph, tUser.ID); err != nil {
		reply.Msg = "could not delete clusters"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getAttributionOverviewReply(r *http.Request, dgraph external.Database) (reply attributionOverviewReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	attributions, err := attribution.GetUserAttributions(dgraph, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Attributions = attributions

	return
}

func getDeleteAttributionReply(r *http.Request, dgraph external.Database,
	isPublicDeletion bool) (reply msgReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	attributionUID := r.PathValue("uid")
	if attributionUID == "" {
		reply.Msg = "attribution uid was not set"
		status = http.StatusBadRequest
		return
	}

	if isPublicDeletion {
		err = attribution.DeletePublicAttribution(dgraph, attributionUID)
	} else {
		err = attribution.DeletePrivateAttribution(dgraph, tUser.ID, attributionUID)
	}

	if err != nil {
		reply.Msg = "could not delete attribution"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getDeleteAllAttributionsReply(r *http.Request, dgraph external.Database) (reply msgReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if err := attribution.DeleteAllAttributions(dgraph, tUser.ID); err != nil {
		reply.Msg = "could not delete clusters"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getAttributionSearchReply(r *http.Request, dgraph external.Database, query string) (reply attributionOverviewReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if query == "" {
		status = http.StatusBadRequest
		return
	}

	attributions, err := attribution.SearchAttributions(dgraph, tUser.ID, query)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Attributions = attributions

	return
}

func getAddAddressExclusionsReply(dgraph external.Database, r *http.Request) (reply msgReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if err := r.ParseMultipartForm(maxBodySize); err != nil {
		status = http.StatusBadRequest
		return
	}

	// Get handler for filename, size and headers
	file, _, err := r.FormFile("file")
	if err != nil {
		reply.Msg = CsvReadError
		status = http.StatusBadRequest
		return
	}

	defer func(file multipart.File) {
		if err := file.Close(); err != nil {
			warn(cliutil.NewStackErrorf("closing CSV-file: %w", err))
		}
	}(file)

	csvReader := csv.NewReader(file)
	csvReader.ReuseRecord = true
	csvReader.FieldsPerRecord = 1
	var line []string

	var addresses []string
	var index int
	for ; ; index++ {
		line, err = csvReader.Read()
		if err != nil {
			if errors.Is(err, csv.ErrFieldCount) {
				reply.Msg = CsvInvalidFieldCount
				status = http.StatusBadRequest
				return
			} else if !errors.Is(err, io.EOF) {
				reply.Msg = CsvInvalidData
				status = http.StatusBadRequest
				return
			}
			break
		}

		trimmed := strings.TrimSpace(line[0])

		if trimmed == "" {
			reply.Msg = CsvInvalidData
			status = http.StatusBadRequest
			return
		}

		addresses = append(addresses, trimmed)
	}

	if len(addresses) == 0 {
		reply.Msg = CsvNoData
		status = http.StatusBadRequest
		return
	}

	if err := analytics.ImportAddressExclusions(dgraph, addresses, tUser.ID); err != nil {
		switch {
		case errors.Is(err, analytics.ErrTooManyAddresses):
			reply.Msg = CsvTooManyAddresses
			status = http.StatusBadRequest
		case errors.Is(err, analytics.ErrNonExistentAddress):
			reply.Msg = CsvInvalidData
			status = http.StatusBadRequest
		default:
			reply.Msg = CsvErrorImporting
			status = http.StatusInternalServerError
			warn(err)
		}

		return
	}

	return
}

func getAddressExclusionOverviewReply(r *http.Request, dgraph external.Database) (reply addressExclusionOverviewReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	addresses, count, err := exclusion.GetAddressExclusions(dgraph, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.AddressHashes = addresses
	reply.Count = count

	return
}

func getDeleteAddressExclusionReply(r *http.Request, dgraph external.Database) int {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		warn(err)
		return http.StatusUnauthorized
	}

	addressHash := r.PathValue("hash")
	if !isValid(addressHash) {
		return http.StatusBadRequest
	}

	if err := exclusion.DeleteAddressExclusion(dgraph, tUser.ID, addressHash); err != nil {
		warn(err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getDeleteAllAddressExclusionsReply(r *http.Request, dgraph external.Database) int {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		warn(err)
		return http.StatusUnauthorized
	}

	if err := exclusion.DeleteAllAddressExclusions(dgraph, tUser.ID); err != nil {
		warn(err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getAddressExclusionStatusReply(r *http.Request, dgraph external.Database, addressHash string) (
	reply addressExclusionStatusReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if !isValid(addressHash) {
		status = http.StatusBadRequest
		return
	}

	exclusionStatus, err := exclusion.GetAddressExclusionStatus(dgraph, addressHash, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.IsExclusion = exclusionStatus

	return
}

// getCreateIdentityReply reads the data from body and constructs a identityReply
func getCreateIdentityReply(dgraph external.Database, adminAuth *ory.APIClient,
	r *http.Request) (reply msgReply, status int) {
	type request struct {
		Email string   `json:"email"`
		Roles []string `json:"roles"`
		State string   `json:"state"`
	}

	var frontEndUser request

	if err := json.NewDecoder(r.Body).Decode(&frontEndUser); err != nil {
		status = http.StatusBadRequest
		return
	}

	if len(frontEndUser.Email) == 0 || len(frontEndUser.Roles) == 0 ||
		len(frontEndUser.State) == 0 || !isValidEmail(frontEndUser.Email) {
		status = http.StatusBadRequest
		return
	}

	// check if all roles have valid values
	for _, ur := range frontEndUser.Roles {
		if _, err := getRoleByName(ur); err != nil {
			status = http.StatusBadRequest
			return
		}
	}

	err := dbus.CreateDgraphAndKratosUser(r.Context(), dgraph, adminAuth,
		frontEndUser.Email, nil, frontEndUser.Roles, frontEndUser.State)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
	}

	return
}

// getDeleteIdentityReply deletes the given user
func getDeleteIdentityReply(r *http.Request, dgraph external.Database,
	adminAuth *ory.APIClient, isAdmin bool) (reply msgReply, status int) {
	var kratosID string
	if isAdmin {
		kratosID = r.PathValue("uid")
	} else {
		tUser, err := extractTokenUser(r.Context())
		if err != nil {
			status = http.StatusUnauthorized
			warn(err)
			return
		}

		kratosID = tUser.KratosID
	}

	if kratosID == "" {
		status = http.StatusBadRequest
		return
	}

	// get identity data
	identity, response, err := adminAuth.IdentityAPI.GetIdentity(r.Context(), kratosID).Execute() //nolint:bodyclose
	if err != nil {
		status = http.StatusInternalServerError
		warn(cliutil.NewStackError(err))
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	uid, err := extractDgraphUID(identity.MetadataPublic)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := attribution.DeleteAllAttributions(dgraph, uid); err != nil {
		reply.Msg = "could not delete users " + uid + " attributions"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := clustering.DeleteAllClusters(dgraph, uid); err != nil {
		reply.Msg = "could not delete users " + uid + " clusters"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := dbwork.DeleteAllWorkspaces(dgraph, uid); err != nil {
		reply.Msg = "could not delete users " + uid + " heuristics"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	err = dbus.DeleteUser(dgraph, uid)
	if err != nil {
		reply.Msg = "could not delete dgraph user"
		status = http.StatusInternalServerError
		warn(err)
		return
	}
	response, err = adminAuth.IdentityAPI.DeleteIdentity(r.Context(), kratosID).Execute() //nolint:bodyclose
	if err != nil {
		reply.Msg = "could not delete identity"
		status = http.StatusInternalServerError
		warn(cliutil.NewStackError(err))
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	return
}

// setRoles adds the given roles to the metadata
func setRoles(metaDataPublic any, roles []string) error {
	metadata, ok := metaDataPublic.(map[string]any)
	if !ok {
		return cliutil.NewStackErrorStr("identity has no public metadata")
	}

	metadata["roles"] = roles

	return nil
}

// setEmail sets the given email to the traits
func setEmail(traits any, email string) error {
	metadata, ok := traits.(map[string]any)
	if !ok {
		return cliutil.NewStackErrorStr("identity has no traits")
	}

	metadata["email"] = email

	return nil
}

// getModifyIdentityReply modifies an identity with the given values in the request body
func getModifyIdentityReply(adminAuth *ory.APIClient, r *http.Request) (reply msgReply, status int) {
	type request struct {
		UID   string   `json:"uid,omitempty"`
		Email string   `json:"email,omitempty"`
		State string   `json:"state,omitempty"`
		Roles []string `json:"roles,omitempty"`
	}

	var modRequest request
	if err := json.NewDecoder(r.Body).Decode(&modRequest); err != nil {
		status = http.StatusBadRequest
		return
	}

	if len(modRequest.UID) == 0 || (len(modRequest.Roles) == 0 && len(modRequest.Email) == 0 &&
		len(modRequest.State) == 0) && !isValidEmail(modRequest.Email) {
		status = http.StatusBadRequest
		return
	}

	initialIdentity, getIdentityResponse, err := adminAuth.IdentityAPI.GetIdentity(r.Context(),
		modRequest.UID).Execute() //nolint:bodyclose
	if err != nil {
		status = http.StatusInternalServerError
		warn(cliutil.NewStackError(err), "modification_request", modRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(getIdentityResponse.Body)

	// handle email change
	if len(modRequest.Email) > 0 {
		if !isValidEmail(modRequest.Email) {
			reply.Msg = "invalid email"
			status = http.StatusBadRequest
			return
		}

		// replace email
		if err = setEmail(initialIdentity.Traits, modRequest.Email); err != nil {
			status = http.StatusInternalServerError
			warn(err, "modification_request", modRequest)
			return
		}
	}
	const msgInvalidRole = "invalid role"

	// handle role change
	if len(modRequest.Roles) > 0 {
		// check if all roles exists
		for _, role := range modRequest.Roles {
			if _, err := getRoleByName(role); err != nil {
				reply.Msg = msgInvalidRole
				status = http.StatusBadRequest
				warn(err, "modification_request", modRequest)
				return
			}
		}

		// replace roles
		if err = setRoles(initialIdentity.MetadataPublic, modRequest.Roles); err != nil {
			status = http.StatusInternalServerError
			warn(err, "modification_request", modRequest)
			return
		}
	}

	// check state
	if len(modRequest.State) > 0 {
		if !dbus.IsStateValid(modRequest.State) {
			status = http.StatusBadRequest
			warn(cliutil.NewStackErrorStr("invalid identity state: "+modRequest.State), "modification_request", modRequest)
			return
		}
		initialIdentity.SetState(modRequest.State)
	}

	_, response, err := adminAuth.IdentityAPI.UpdateIdentity(r.Context(), modRequest.UID).UpdateIdentityBody(ory.UpdateIdentityBody{
		MetadataAdmin:  initialIdentity.MetadataAdmin,
		MetadataPublic: initialIdentity.MetadataPublic,
		SchemaId:       initialIdentity.SchemaId,
		State:          *initialIdentity.State,
		Traits:         initialIdentity.Traits.(map[string]any),
	}).Execute() //nolint:bodyclose
	if err != nil {
		status = http.StatusInternalServerError
		warn(cliutil.NewStackError(err), "modification_request", modRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	return
}

func getSpendingFingerprintReply(dgraph external.Database, worker *worker.Worker,
	txhash string) (reply spendingFingerprintReply, status int) {
	if !isValid(txhash) {
		status = http.StatusBadRequest
		return
	}

	if !worker.IsReady() {
		reply.Msg = "Server is not ready to receive lookups. Please try again later."
		return
	}

	uid, err := db.GetTransactionUID(dgraph, txhash)
	if err != nil {
		if errors.Is(err, db.ErrTransactionNotFound) {
			status = http.StatusBadRequest
			reply.Msg = "Transaction " + txhash + " does not exist."
			return
		}

		status = http.StatusInternalServerError
		warn(err)
		return
	}

	similarTransactions, sessionCount, err := worker.SpendingFingerprint(uid)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	// no fingerprints -> nothing to do
	if len(similarTransactions) == 0 {
		return
	}

	uids := make([]string, len(similarTransactions))
	uidToFingerprint := make(map[string]fingerprintScore, len(similarTransactions))
	for i, tx := range similarTransactions {
		uids[i] = tx.TransactionUID
		uidToFingerprint[tx.TransactionUID] = fingerprintScore{Score: tx.Score, SessionCount: tx.SessionCount}
	}

	transactions, err := db.GetTransactionUIDMapping(dgraph, uids)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if len(transactions) != len(uids) {
		status = http.StatusInternalServerError
		warn(cliutil.NewStackErrorf("length of uids and hashes is not equal for %s", txhash))
		return
	}

	for _, tx := range transactions {
		fingerprint, ok := uidToFingerprint[tx.UID]
		if !ok {
			status = http.StatusInternalServerError
			warn(cliutil.NewStackErrorf("could not find uid to tx hash mapping for %s in request for %s", txhash, tx.UID))
			return
		}

		fingerprint.Txhash = tx.Hash
		reply.FingerprintScores = append(reply.FingerprintScores, fingerprint)
	}

	reply.SessionCount = sessionCount

	return
}

func getAddWorkspaceNodeReply(dgraph external.Database, workspaceMutex *workspace.Mutex,
	worker *worker.Worker, r *http.Request) (reply addWorkspaceNodeReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		Query        string `json:"query,omitempty"`
		WorkspaceUID string `json:"workspaceUID,omitempty"`
	}

	var searchRequest request

	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		status = http.StatusBadRequest
		warn(cliutil.NewStackError(err))
		return
	}

	if searchRequest.WorkspaceUID == "" || !isValid(searchRequest.Query) {
		status = http.StatusBadRequest
		return
	}

	newNode, err := dbwork.SearchForNode(dgraph, searchRequest.Query, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err, "query", searchRequest)
		return
	}

	if newNode == nil {
		status = http.StatusBadRequest
		return
	}

	reply.Nodes, err = workspace.AddNode(dgraph, workspaceMutex, worker, searchRequest.WorkspaceUID, tUser.ID, newNode)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getWorkspacesReply(dgraph external.Database, r *http.Request) (reply workspacesReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	workspaces, err := dbwork.GetFrontendWorkspaces(dgraph, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Workspaces = make([]dbwork.FrontendWorkspace, len(workspaces))
	for i, w := range workspaces {
		reply.Workspaces[i] = w.ToFrontendWorkspace()
	}

	return
}

func getGetWorkspaceReply(dgraph external.Database, workspaceMutex *workspace.Mutex, worker *worker.Worker,
	r *http.Request) (reply getWorkspaceReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	workspaceUID := r.PathValue("uid")
	if workspaceUID == "" {
		status = http.StatusBadRequest
		return
	}

	reply.Workspace, err = workspace.GetAndRefreshWorkspace(dgraph, worker, workspaceMutex, workspaceUID, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Descriptors = make([]heuristics.Descriptor, len(heuristics.ValidHeuristicTypes))
	for i, t := range heuristics.ValidHeuristicTypes {
		reply.Descriptors[i] = t.GetDescriptor()
	}

	return
}

func getAddWorkspaceReply(dgraph external.Database, r *http.Request) (status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	workspaceName := r.PathValue("name")
	if workspaceName == "" || len(workspaceName) > 150 {
		status = http.StatusBadRequest
		return
	}

	err = dbwork.AddWorkspace(dgraph, workspaceName, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getUpdateWorkspace(dgraph external.Database, workspaceMutex *workspace.Mutex, r *http.Request) (status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		CurrentState []dbwork.Node `json:"currentState,omitempty"`
		WorkspaceUID string        `json:"workspaceUID,omitempty"`
	}

	var searchRequest request

	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		status = http.StatusBadRequest
		warn(cliutil.NewStackError(err))
		return
	}

	if searchRequest.WorkspaceUID == "" || len(searchRequest.CurrentState) == 0 {
		status = http.StatusBadRequest
		return
	}

	if err = workspace.UpdateNodeCoordinates(dgraph, workspaceMutex,
		searchRequest.WorkspaceUID, tUser.ID, searchRequest.CurrentState); err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getDeleteWorkspaceNodeReply(dgraph external.Database, workspaceMutex *workspace.Mutex,
	r *http.Request) (reply deleteWorkspaceNodeReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		NodeUID      string `json:"nodeUID,omitempty"`
		WorkspaceUID string `json:"workspaceUID,omitempty"`
	}

	var searchRequest request

	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		status = http.StatusBadRequest
		warn(cliutil.NewStackError(err))
		return
	}

	if searchRequest.WorkspaceUID == "" || searchRequest.NodeUID == "" {
		status = http.StatusBadRequest
		return
	}

	reply.DeletedNodeUIDs, err = workspace.DeleteNode(dgraph, workspaceMutex, searchRequest.WorkspaceUID, tUser.ID, searchRequest.NodeUID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getDeleteWorkspaceReply(r *http.Request, dgraph external.Database) (reply msgReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	workspaceUID := r.PathValue("uid")
	if workspaceUID == "" {
		status = http.StatusBadRequest
		return
	}

	if err := dbwork.DeleteWorkspace(dgraph, tUser.ID, workspaceUID); err != nil {
		if errors.Is(err, dbwork.ErrNoMutationHappened) {
			reply.Msg = "No data was deleted. The transaction might not have any workspaces."
			status = http.StatusNotFound
		} else {
			reply.Msg = "could not delete data"
			status = http.StatusInternalServerError
			warn(err)
		}
		return
	}

	return
}

func getDeleteAllWorkspacesReply(r *http.Request, dgraph external.Database) (reply msgReply, status int) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if err := dbwork.DeleteAllWorkspaces(dgraph, tUser.ID); err != nil {
		if errors.Is(err, dbwork.ErrNoMutationHappened) {
			reply.Msg = "No data was deleted. The user might not have any workspaces."
			status = http.StatusNotFound
		} else {
			reply.Msg = "could not delete data"
			status = http.StatusInternalServerError
			warn(err)
		}
		return
	}

	return
}
