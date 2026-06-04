// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	analyticsClustering "gitlab.com/blockchain-privacy/dakar/analytics/clustering"
	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/analytics/heuristics"
	"gitlab.com/blockchain-privacy/dakar/cmd/cliutil"
	"gitlab.com/blockchain-privacy/dakar/db"
	dbAnalytics "gitlab.com/blockchain-privacy/dakar/db/analytics"
	"gitlab.com/blockchain-privacy/dakar/db/analytics/clustering"
	dbstat "gitlab.com/blockchain-privacy/dakar/db/status"
	dbwork "gitlab.com/blockchain-privacy/dakar/db/workspace"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/dakar/workspace"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// getSearchReply searches for the given query in the database
func getSearchReply(dgraph external.Database, r *http.Request) (reply searchReply, status int) {
	query := r.PathValue("query")

	if !isValid(query) {
		return reply, http.StatusBadRequest
	}

	queryType, err := db.Search(r.Context(), dgraph, query)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if queryType == "" {
		status = http.StatusNotFound
		reply.Type = "response_empty"
	} else {
		reply.Type = queryType
	}

	return reply, status
}

func getAddressReply(dgraph external.Database, r *http.Request) (reply addressReply, status int) {
	query := r.PathValue("hash")
	if !isValid(query) {
		return reply, http.StatusBadRequest
	}

	addr, err := db.GetFrontendAddress(r.Context(), dgraph, query, db.SortAscendingByOutputTime, 0, nil)
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

	reply.Address = addr

	return
}

func getBlockReply(dgraph external.Database, r *http.Request) (reply blockReply, status int) {
	query := r.PathValue("hash")
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

	block, err := db.GetFrontendBlock(r.Context(), dgraph, query, offset)
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

func getTransactionReply(dgraph external.Database, r *http.Request) (reply transactionReply, status int) {
	query := r.PathValue("hash")
	if !isValid(query) {
		return reply, http.StatusBadRequest
	}

	transactions, err := db.GetFrontendTransaction(r.Context(), dgraph, query)
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
func getAddressOutputRangeReply(dgraph external.Database, r *http.Request) (reply addressReply, status int) {
	addressHash := r.PathValue("hash")

	type request struct {
		Offset int   `json:"offset"`
		Order  int   `json:"order"`
		Filter []int `json:"filter"`
	}

	if !isValid(addressHash) {
		status = http.StatusBadRequest
		return
	}

	var addressRequest request
	addressRequest.Offset = -1
	addressRequest.Order = -1

	if decodeErr := json.NewDecoder(r.Body).Decode(&addressRequest); decodeErr != nil {
		status = http.StatusBadRequest
		return
	}

	if !db.IsValidSortOrder(addressRequest.Order) {
		status = http.StatusBadRequest
		return
	}

	if !db.IsValidFilter(addressRequest.Filter) {
		status = http.StatusBadRequest
		return
	}

	if addressRequest.Offset < 0 {
		status = http.StatusBadRequest
		return
	}

	data, err := db.GetFrontendAddress(r.Context(), dgraph, addressHash,
		addressRequest.Order, addressRequest.Offset, addressRequest.Filter)
	if err != nil {
		if errors.Is(err, db.ErrAddressNotFound) {
			status = http.StatusNotFound
			return
		}

		status = http.StatusInternalServerError
		return
	}

	status = http.StatusOK
	reply.Address = data

	return
}

func getMetaReply(dgraph external.Database, rpcClient external.RPCClient, r *http.Request) (reply metaReply, status int) {
	if rpcClient != nil {
		blocks, err := rpcClient.GetBlockCount()
		if err != nil {
			status = http.StatusInternalServerError
			warn(err)
			return
		}
		reply.Blocks = &blocks
	}

	verboseStatus, err := dbstat.GetFrontendStatus(r.Context(), dgraph)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Status = &verboseStatus

	return
}

func getSelectorStatus(workspaceMutex *workspace.Mutex, dgraph external.Database, r *http.Request) (reply selectorStatusReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		SelectorUID  string `json:"selectorUID"`
		WorkspaceUID string `json:"workspaceUID"`
	}

	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		warn(err)
		return
	}

	if req.SelectorUID == "" || req.WorkspaceUID == "" {
		status = http.StatusBadRequest
		return
	}

	selectorStatus, err := dbwork.GetSelectorStatus(r.Context(), dgraph,
		req.SelectorUID, req.WorkspaceUID, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if selectorStatus == dbwork.StatusWaiting {
		return
	}

	lock := workspaceMutex.Lock(req.WorkspaceUID)
	defer lock.Unlock()

	w, err := dbwork.GetFrontendWorkspace(r.Context(), dgraph, req.WorkspaceUID, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Nodes = w.Nodes

	return
}

func getSelectorResultsReply(dgraph external.Database, r *http.Request) (reply selectorResultsReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		SelectorUID  string `json:"selectorUID"`
		WorkspaceUID string `json:"workspaceUID"`
	}

	var heuristicRequest request

	if err := json.NewDecoder(r.Body).Decode(&heuristicRequest); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}

	if heuristicRequest.SelectorUID == "" || heuristicRequest.WorkspaceUID == "" {
		status = http.StatusBadRequest
		return
	}

	reply.Results, err = dbwork.GetSelectorResultsByUID(r.Context(), dgraph,
		heuristicRequest.SelectorUID, tUser.ID, heuristicRequest.WorkspaceUID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getAddWorkspaceSelectorReply(dgraph external.Database, r *http.Request,
	workspaceMutex *workspace.Mutex, worker *workspace.Worker) (reply addWorkspaceSelectorReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		Type             string                       `json:"type"`
		Parent           string                       `json:"parent"`
		HeuristicOptions *heuristics.HeuristicOptions `json:"heuristicOptions,omitempty"`
		TxPropOptions    *dbwork.TxPropOptions        `json:"txPropOptions,omitempty"`
		TxGraphOptions   *dbwork.TxGraphOptions       `json:"txGraphOptions,omitempty"`
		WorkspaceUID     string                       `json:"workspaceUID"`
	}

	var selectorRequest request
	if err := json.NewDecoder(r.Body).Decode(&selectorRequest); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}

	var options dbwork.Options
	if selectorRequest.TxPropOptions != nil {
		options = *selectorRequest.TxPropOptions
	} else if selectorRequest.TxGraphOptions != nil {
		options = *selectorRequest.TxGraphOptions
	} else if selectorRequest.HeuristicOptions != nil {
		options = *selectorRequest.HeuristicOptions
	} else {
		status = http.StatusBadRequest
		return
	}

	if _, reply.Nodes, err = workspace.AddSelector(r.Context(), dgraph, workspaceMutex, options,
		selectorRequest.Type, selectorRequest.Parent, selectorRequest.WorkspaceUID, tUser.ID); err != nil {
		if errors.Is(err, db.ErrInvalidRequestArgument) {
			status = http.StatusBadRequest
		} else {
			status = http.StatusInternalServerError
		}

		warn(err)
		return
	}

	worker.TriggerSearch()

	return
}

// getShortestTransactionPathReply searches for the shortest path between two transactions
func getShortestTransactionPathReply(dgraph external.Database, r *http.Request) (reply shortestTransactionPathReply, status int) {
	type request struct {
		// From is the starting point of the shortest path lookup
		From string `json:"from"`
		// To is the end point of the shortest path lookup
		To string `json:"to"`
		// IncludePrivacyTransactions determines if classified transactions
		// should be considered when doing the shortest path lookup
		IncludePrivacyTransactions bool `json:"includePrivacyTransactions"`
		// AnyDirection determines the search direction of the shortest transaction path query
		// True: Both inputs and outputs are traversed
		// False: Only inputs are traversed
		AnyDirection bool `json:"anyDirection"`
	}

	// parse request
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		return
	}

	if req.From == "" || req.To == "" {
		status = http.StatusBadRequest
		reply.Msg = "Hash is empty"
		return
	}

	if req.From == req.To {
		status = http.StatusBadRequest
		reply.Msg = "Transaction hashes are equal"
		return
	}

	fromBlockID, err := db.GetTransactionBlockID(r.Context(), dgraph, req.From)
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

	toBlockID, err := db.GetTransactionBlockID(r.Context(), dgraph, req.To)
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

	// limit shortest task query duration
	ctx, cancelFunc := context.WithTimeout(r.Context(), time.Second*30)
	defer cancelFunc()

	txs, err := dbAnalytics.GetShortestTransactionPathAnyDirection(ctx, dgraph,
		oldTx, youngTx, req.IncludePrivacyTransactions, anyDirection)
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

// getClusterLookupReply returns the result of a cluster lookup
func getClusterLookupReply(dgraph external.Database, r *http.Request) (reply clusterLookupReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
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
	clusters, err := clustering.GetClusters(r.Context(), dgraph, addressHash, maxAddresses, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Clusters = clusters

	return
}

// writeSelectorReport writes selector data in CSV format
func writeSelectorReport(dgraph external.Database, w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	const errReport = "error getting selector report"

	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		http.Error(w, errReport, http.StatusBadRequest)
		warn(err)
		return
	}

	type request struct {
		SelectorUID  string `json:"selectorUID"`
		WorkspaceUID string `json:"workspaceUID"`
	}

	var selectorRequest request
	if err := json.NewDecoder(r.Body).Decode(&selectorRequest); err != nil {
		http.Error(w, errReport, http.StatusBadRequest)
		warn(serror.New(err))
		return
	}

	if selectorRequest.SelectorUID == "" || selectorRequest.WorkspaceUID == "" {
		http.Error(w, errReport, http.StatusBadRequest)
		return
	}

	selectorResults, err := dbwork.GetSelectorResultsByUID(r.Context(), dgraph,
		selectorRequest.SelectorUID, tUser.ID, selectorRequest.WorkspaceUID)
	if err != nil {
		http.Error(w, errReport, http.StatusInternalServerError)
		warn(err)
		return
	}

	// headers for streaming data to client
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", selectorRequest.SelectorUID))
	w.Header().Set("Content-Type", "text/csv")

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	if len(selectorResults) > 0 {
		err = workspace.WriteTransactionsToCsv(csvWriter, selectorResults)
	} else {
		http.Error(w, "invalid selector results", http.StatusInternalServerError)
		warn(serror.FromStr("invalid selector results"))
		return
	}

	if err != nil {
		http.Error(w, "error writing to csv stream", http.StatusInternalServerError)
		warn(err)
		return
	}
}

// writeClusterReport writes cluster data in CSV format
func writeClusterReport(dgraph external.Database, w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	addressHash := r.PathValue("hash")
	if !isValid(addressHash) {
		http.Error(w, "no address hash provided", http.StatusNotFound)
		return
	}

	const errReport = "error getting cluster report"

	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		http.Error(w, errReport, http.StatusNotFound)
		warn(err)
		return
	}

	clusters, err := clustering.GetClusters(r.Context(), dgraph, addressHash, 0, tUser.ID)
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
		warn(serror.New(err))
		return
	}

	for _, c := range clusters {
		for _, a := range c.Addresses {
			if err = csvWriter.Write([]string{string(c.Type), c.TransactionHash, c.Timestamp.Format(time.RFC3339),
				a.AddressHash, strconv.Itoa(a.OutputCount), strconv.Itoa(a.OutputCount - a.SpentOutputCount)}); err != nil {
				// communication with client is not possible, can only log error
				// this is because as soon as we write the CSV header, the HTTP response status is also sent
				warn(serror.New(err))
				return
			}
		}
		csvWriter.Flush()
	}
}

// getMixingActivity returns the result of a mixing activity lookup
func getMixingActivity(dgraph external.Database, r *http.Request) (reply mixingActivityReply, status int) {
	type request struct {
		// AddressHash is the address hash for which the lookup will be done
		AddressHash string `json:"addressHash"`
		// IsClusterLookup determines if all addresses of the cluster will be considered
		IsClusterLookup bool `json:"isClusterLookup"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}
	const maxAddressCount = 2000
	if req.IsClusterLookup {
		addressCount, err := clustering.GetClusterAddressCount(r.Context(), dgraph, req.AddressHash)
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

	activities, err := dbAnalytics.GetMixingActivity(r.Context(), dgraph, req.AddressHash, req.IsClusterLookup)
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
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if err := r.ParseMultipartForm(maxBodySize); err != nil { //nolint:gosec
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
			warn(serror.FromFormat("error closing file: %w", err))
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

	if err := analyticsClustering.ImportCluster(r.Context(), dgraph, addresses, tUser.ID); err != nil {
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

func getClusterOverviewReply(r *http.Request, dgraph external.Database) (reply clusterOverviewReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	clusters, err := clustering.GetUserClusters(r.Context(), dgraph, tUser.ID)
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
	tUser, err := ExtractTokenUser(r.Context())
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

	if err := clustering.DeleteCluster(r.Context(), dgraph, tUser.ID, clusterUID); err != nil {
		reply.Msg = "could not delete cluster"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getDeleteAllClustersReply(dgraph external.Database, r *http.Request) (reply msgReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if err := clustering.DeleteAllClusters(r.Context(), dgraph, tUser.ID); err != nil {
		reply.Msg = "could not delete clusters"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getSpendingFingerprintReply(dgraph external.Database, graphWrapper *graph.Wrapper,
	r *http.Request) (reply spendingFingerprintReply, status int) {
	txhash := r.PathValue("hash")
	if !isValid(txhash) {
		status = http.StatusBadRequest
		return
	}

	if !graphWrapper.IsTransactionGraphLoaded() {
		reply.Msg = "Server is not ready to receive lookups. Please try again later."
		return
	}

	uid, err := db.GetTransactionUID(r.Context(), dgraph, txhash)
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

	similarTransactions, sessionCount, err := graphWrapper.SpendingFingerprint(uid)
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

	transactions, err := db.GetTransactionUIDMapping(r.Context(), dgraph, uids)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if len(transactions) != len(uids) {
		status = http.StatusInternalServerError
		warn(serror.FromFormat("length of uids and hashes is not equal for %s", txhash))
		return
	}

	for _, tx := range transactions {
		fingerprint, ok := uidToFingerprint[tx.UID]
		if !ok {
			status = http.StatusInternalServerError
			warn(serror.FromFormat("could not find uid to tx hash mapping for %s in request for %s", txhash, tx.UID))
			return
		}

		fingerprint.Txhash = tx.Hash
		reply.FingerprintScores = append(reply.FingerprintScores, fingerprint)
	}

	reply.SessionCount = sessionCount

	return
}

func getImportWorkspaceReply(dgraph external.Database, blockchainMode string, r *http.Request) (reply msgReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	if err := r.ParseMultipartForm(maxBodySize); err != nil { //nolint:gosec
		status = http.StatusBadRequest
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		status = http.StatusBadRequest
		return
	}

	defer func(file multipart.File) {
		if err := file.Close(); err != nil {
			warn(serror.FromFormat("error closing file: %w", err))
		}
	}(file)

	var export workspace.Export
	if err = json.NewDecoder(file).Decode(&export); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}

	if err = workspace.ImportWorkspace(r.Context(), dgraph, blockchainMode, export, tUser.ID); err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func writeExportWorkspace(dgraph external.Database, workspaceMutex *workspace.Mutex, blockchainMode string,
	w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	const errReport = "error getting workspace export"

	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		http.Error(w, errReport, http.StatusUnauthorized)
		warn(err)
		return
	}

	type request struct {
		WorkspaceUID string `json:"workspaceUID"`
	}

	var searchRequest request

	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		http.Error(w, errReport, http.StatusBadRequest)
		warn(serror.New(err))
		return
	}

	if searchRequest.WorkspaceUID == "" {
		http.Error(w, errReport, http.StatusBadRequest)
		return
	}

	export, err := workspace.ExportWorkspace(r.Context(), dgraph, workspaceMutex, blockchainMode, searchRequest.WorkspaceUID, tUser.ID)
	if err != nil {
		http.Error(w, errReport, http.StatusInternalServerError)
		warn(err)
		return
	}

	// headers for streaming data to client
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=workspace_export_%s_%s.json",
		export.Meta.ExportDate.Format("2006-01-02"), export.Meta.Name))
	w.Header().Set("Content-Type", "application/json")

	// use marshalling instead of encoding (streaming), as it gives better error handling
	// and because encoding buffers all data before writing: https://github.com/golang/go/issues/7872
	// todo check if https://github.com/golang/go/discussions/63397 has been released and then rework json handling.
	replyBuffer, err := json.Marshal(export)
	if err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
		warn(serror.New(err))
		return
	}
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(replyBuffer); err != nil {
		// not possible to send response to client, so just log error
		warn(serror.New(err))
	}
}

// writeEntityExport exports basic workspace node data
func writeEntityExport(dgraph external.Database, workspaceMutex *workspace.Mutex, w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	const errBasicExport = "error getting workspace report"

	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		http.Error(w, errBasicExport, http.StatusBadRequest)
		warn(err)
		return
	}

	type request struct {
		WorkspaceUID string `json:"workspaceUID"`
	}

	var exportRequest request
	if err := json.NewDecoder(r.Body).Decode(&exportRequest); err != nil {
		http.Error(w, errBasicExport, http.StatusBadRequest)
		warn(serror.New(err))
		return
	}

	if exportRequest.WorkspaceUID == "" {
		http.Error(w, errBasicExport, http.StatusBadRequest)
		return
	}

	nodes, err := workspace.ExportBasic(r.Context(), dgraph, workspaceMutex,
		exportRequest.WorkspaceUID, tUser.ID)
	if err != nil {
		http.Error(w, errBasicExport, http.StatusInternalServerError)
		warn(err)
		return
	}

	// headers for streaming data to client
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", exportRequest.WorkspaceUID))
	w.Header().Set("Content-Type", "text/csv")

	csvWriter := csv.NewWriter(w)

	header := []string{"transactions", "addresses"}
	if err := csvWriter.Write(header); err != nil {
		http.Error(w, errBasicExport, http.StatusInternalServerError)
		warn(serror.New(err))
		return
	}

	maximum := max(len(nodes[0]), len(nodes[1]))
	for i := range maximum {
		line := make([]string, 2)

		if len(nodes[0]) > i {
			line[0] = nodes[0][i]
		}

		if len(nodes[1]) > i {
			line[1] = nodes[1][i]
		}

		if err := csvWriter.Write(line); err != nil {
			http.Error(w, errBasicExport, http.StatusInternalServerError)
			warn(serror.New(err))
			return
		}
	}
	csvWriter.Flush()

	return
}

func getAddWorkspaceNodesReply(dgraph external.Database, workspaceMutex *workspace.Mutex,
	r *http.Request) (reply addWorkspaceNodesReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		Queries      []string `json:"queries"`
		WorkspaceUID string   `json:"workspaceUID"`
	}

	var searchRequest request

	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}

	if searchRequest.WorkspaceUID == "" || len(searchRequest.Queries) == 0 {
		status = http.StatusBadRequest
		return
	}

	for _, query := range searchRequest.Queries {
		if !isValid(query) {
			status = http.StatusBadRequest
			return
		}
	}

	newNodes := map[string]dbwork.Node{}
	for _, query := range searchRequest.Queries {
		newNode, err := dbwork.SearchForNode(r.Context(), dgraph, query)
		if err != nil {
			if errors.Is(err, dbwork.ErrNodeNotFound) {
				// ignore nodes which are not found
				continue
			}

			status = http.StatusInternalServerError
			warn(err, "query", searchRequest)
			return
		}

		newNodes[newNode.UID] = *newNode
	}

	if len(newNodes) == 0 {
		status = http.StatusNotFound
		return
	}

	reply.Nodes, reply.DuplicateNodeUID, err = workspace.AddNodes(r.Context(), dgraph, workspaceMutex, searchRequest.WorkspaceUID,
		tUser.ID, cliutil.GetMapValues(newNodes))
	if err != nil {
		if errors.Is(err, workspace.ErrTooManyOutputs) {
			reply.ClusterTooLarge = true
			return
		}

		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getAddWorkspaceNoteReply(dgraph external.Database, workspaceMutex *workspace.Mutex,
	r *http.Request) (reply addWorkspaceNoteReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		UID string `json:"uid"`
		// the note text
		Text string `json:"text"`
		// the node to which the note is connected to
		ChildUID     string `json:"childUID"`
		WorkspaceUID string `json:"workspaceUID"`
	}

	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}

	if req.WorkspaceUID == "" || req.Text == "" || req.ChildUID == "" || len(req.Text) > 100 {
		status = http.StatusBadRequest
		return
	}

	reply.Nodes, err = workspace.AddNote(r.Context(), dgraph, workspaceMutex, req.WorkspaceUID,
		tUser.ID, dbwork.Node{
			UID:      req.UID,
			Text:     req.Text,
			Children: []string{req.ChildUID},
		})
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getWorkspacesReply(dgraph external.Database, r *http.Request) (reply workspacesReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	workspaces, err := dbwork.GetFrontendWorkspaces(r.Context(), dgraph, tUser.ID)
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

func getGetWorkspaceReply(dgraph external.Database, workspaceMutex *workspace.Mutex,
	r *http.Request) (reply getWorkspaceReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
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

	reply.Workspace, err = workspace.GetAndRefreshWorkspace(r.Context(), dgraph, workspaceMutex, workspaceUID, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.Descriptors = make([]heuristics.Descriptor, 0, len(heuristics.ConstructorMap))
	for _, v := range heuristics.ConstructorMap {
		reply.Descriptors = append(reply.Descriptors, v().GetDescriptor())
	}

	return
}

func getGetWorkspaceStateReply(dgraph external.Database, r *http.Request) (reply getWorkspaceStateReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
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

	reply.State, err = dbwork.GetWorkspaceState(r.Context(), dgraph, workspaceUID, tUser.ID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getAddWorkspaceReply(dgraph external.Database, r *http.Request) (status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	workspaceName := r.PathValue("name")
	if workspaceName == "" || len(workspaceName) > workspace.MaxWorkspaceNameLength {
		status = http.StatusBadRequest
		return
	}

	_, err = dbwork.AddWorkspace(r.Context(), dgraph, workspaceName, tUser.ID, "")
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getRenameWorkspaceReply(dgraph external.Database, r *http.Request) (status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		Name         string `json:"name"`
		WorkspaceUID string `json:"workspaceUID"`
	}

	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}

	if req.Name == "" || len(req.Name) > workspace.MaxWorkspaceNameLength || req.WorkspaceUID == "" {
		status = http.StatusBadRequest
		return
	}

	err = dbwork.RenameWorkspace(r.Context(), dgraph, req.Name, tUser.ID, req.WorkspaceUID)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getUpdateWorkspace(dgraph external.Database, workspaceMutex *workspace.Mutex, r *http.Request) (status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		CurrentState []dbwork.Node `json:"currentState"`
		WorkspaceUID string        `json:"workspaceUID"`
	}

	var searchRequest request

	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}

	if searchRequest.WorkspaceUID == "" || len(searchRequest.CurrentState) == 0 {
		status = http.StatusBadRequest
		return
	}

	if err = workspace.UpdateNodeCoordinates(r.Context(), dgraph, workspaceMutex,
		searchRequest.WorkspaceUID, tUser.ID, searchRequest.CurrentState); err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getDeleteWorkspaceNodeReply(dgraph external.Database, workspaceMutex *workspace.Mutex,
	r *http.Request) (reply deleteWorkspaceNodeReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		NodeUIDs     []string `json:"nodeUIDs"`
		WorkspaceUID string   `json:"workspaceUID"`
	}

	var delRequest request

	if err := json.NewDecoder(r.Body).Decode(&delRequest); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}

	if delRequest.WorkspaceUID == "" || len(delRequest.NodeUIDs) == 0 {
		status = http.StatusBadRequest
		return
	}

	reply.DeletedNodeUIDs, err = workspace.DeleteNodes(r.Context(), dgraph, workspaceMutex,
		delRequest.WorkspaceUID, tUser.ID, delRequest.NodeUIDs)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

func getDeleteWorkspaceReply(dgraph external.Database, r *http.Request) (reply msgReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
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

	if err := dbwork.DeleteWorkspace(r.Context(), dgraph, tUser.ID, workspaceUID); err != nil {
		if errors.Is(err, db.ErrNoMutationHappened) {
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

func getWorkspaceConnectionReply(dgraph external.Database, r *http.Request) (reply workspaceConnectionReply, status int) {
	tUser, err := ExtractTokenUser(r.Context())
	if err != nil {
		status = http.StatusUnauthorized
		warn(err)
		return
	}

	type request struct {
		FirstNode struct {
			UID  string `json:"uid"`
			Type string `json:"type"`
		} `json:"firstNode"`
		SecondNode struct {
			UID  string `json:"uid"`
			Type string `json:"type"`
		} `json:"secondNode"`
		WorkspaceUID string `json:"workspaceUID"`
	}

	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		warn(serror.New(err))
		return
	}

	// validate input
	if req.FirstNode.UID == "" || req.SecondNode.UID == "" ||
		req.WorkspaceUID == "" || req.FirstNode.UID == req.SecondNode.UID {
		status = http.StatusBadRequest
		return
	}

	if req.FirstNode.Type == dbwork.NodeTypeCluster && req.SecondNode.Type == dbwork.NodeTypeCluster {
		reply.AmountTransactions, err = dbwork.GetConnectionClusterToCluster(r.Context(), dgraph,
			req.FirstNode.UID, req.SecondNode.UID)
		if err != nil {
			status = http.StatusInternalServerError
			warn(err)
			return
		}
	} else if req.FirstNode.Type == dbwork.NodeTypeCluster && req.SecondNode.Type == dbwork.NodeTypeSelector ||
		req.FirstNode.Type == dbwork.NodeTypeSelector && req.SecondNode.Type == dbwork.NodeTypeCluster {
		clusterUID := req.FirstNode.UID
		heuristicUID := req.SecondNode.UID

		if req.SecondNode.Type == dbwork.NodeTypeCluster {
			clusterUID = req.SecondNode.UID
			heuristicUID = req.FirstNode.UID
		}

		reply.AmountTransactions, err = dbwork.GetConnectionClusterToSelector(r.Context(), dgraph, clusterUID,
			heuristicUID, tUser.ID, req.WorkspaceUID)
		if err != nil {
			status = http.StatusInternalServerError
			warn(err)
			return
		}
	} else if req.FirstNode.Type == dbwork.NodeTypeCluster && req.SecondNode.Type == dbwork.NodeTypeTransaction ||
		req.FirstNode.Type == dbwork.NodeTypeTransaction && req.SecondNode.Type == dbwork.NodeTypeCluster {
		clusterUID := req.FirstNode.UID
		transactionUID := req.SecondNode.UID

		if req.FirstNode.Type == dbwork.NodeTypeTransaction {
			clusterUID = req.SecondNode.UID
			transactionUID = req.FirstNode.UID
		}

		reply.FrontendTransactions, err = dbwork.GetConnectionClusterToTransaction(r.Context(), dgraph, clusterUID, transactionUID)
		if err != nil {
			status = http.StatusInternalServerError
			warn(err)
			return
		}
	} else {
		// wrong combination of types
		status = http.StatusBadRequest
		return
	}

	return
}
