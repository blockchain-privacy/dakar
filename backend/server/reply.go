package server

import (
	"backend/analytics"
	analyticsClustering "backend/analytics/clustering"
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/constants"
	dbAnalytics "backend/db/analytics"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"backend/db/analytics/exclusion"
	dbHeuristic "backend/db/analytics/heuristics"
	dbtx "backend/db/transaction"
	dbus "backend/db/user"
	"backend/external"
	"backend/user"
	"strings"

	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const msgCouldNotDecodeRequest = "could not decode request data"
const msgCouldNotDecodeUser = "could not decode user data"
const msgInvalidRequest = "invalid request"
const msgUserNotFound = "User not found"

// getLoginReply reads the data from body and constructs a backendUserReply
func getLoginReply(dgraph external.Database, body io.Reader) (reply backendUserReply) {
	const invalidUserData = "email and password combination does not match"

	var loginData dbus.FrontendUserLogin

	if err := json.NewDecoder(body).Decode(&loginData); err != nil {
		reply.Msg = msgCouldNotDecodeUser
		return
	}

	if !loginData.IsValid() {
		reply.Msg = "user not valid"
		return
	}

	u, getErr := dbus.GetUserByEmail(dgraph, loginData.Email)
	if getErr != nil {
		reply.Msg = invalidUserData

		// only log error if not expected
		if !errors.Is(dbus.ErrorUsersNotFound, getErr) {
			info(cliutil.ShowCallInfo(), getErr)
		}
		return
	}

	passwordValid, pwErr := user.ComparePassword(loginData.Password, u.PasswordHash)
	if pwErr != nil || !passwordValid {
		reply.Msg = invalidUserData
		return
	}

	u.PasswordHash = ""

	loginReplyUser := u.ToFrontendUserBackendState()

	reply.Success = true

	reply.User = &loginReplyUser

	return
}

// getCreateUserReply reads the data from body and constructs a userReply
func getCreateUserReply(dgraph external.Database, body io.Reader) (reply userReply) {
	var frontEndUser dbus.FrontendUserRoles

	if err := json.NewDecoder(body).Decode(&frontEndUser); err != nil {
		reply.Msg = msgCouldNotDecodeUser
		return
	}

	if !frontEndUser.IsValid() {
		reply.Msg = "user not valid"
		return
	}

	pw, pwHash, pwErr := user.GetRandomPasswordAndHash()
	if pwErr != nil {
		reply.Msg = "could not create password"
		return
	}

	u := frontEndUser.ToUser()
	u.PasswordHash = pwHash
	if err := dbus.CreateUser(dgraph, u); err != nil {
		reply.Success = false
		// check if special error
		if errors.Is(err, dbus.ErrorEmailExists) {
			reply.Msg = "duplicate e-mail"
		} else {
			reply.Msg = "could not create user"
			info(cliutil.ShowCallInfo(), err)
		}
		return
	}

	info("Generated password(", u.Email, "):", pw)
	reply.Success = true

	return
}

func getHeuristicReply(dgraph external.Database, worker *heuristics.Worker,
	txHashString string, userUID string) (reply heuristicReply) {

	results, err := dbHeuristic.GetBasicFrontendHeuristic(dgraph, txHashString, userUID)
	if err != nil {
		reply.Msg = "no heuristics found"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true
	reply.Heuristics = results
	reply.Status = worker.GetStatus(txHashString, userUID)

	return
}

func getHeuristicExecutionReply(dgraph external.Database, worker *heuristics.Worker, body io.Reader,
	txHashString string, userUID string) (reply heuristicExecutionReply) {
	if !worker.IsReady() {
		reply.Success = true
		reply.Status = heuristics.StatusHeuristicWorkerNotReady
		return
	}

	if worker.IsInQueue(txHashString, userUID) {
		reply.Success = true
		reply.Status = heuristics.StatusHeuristicDuplicate
		info(cliutil.ShowCallInfo(), "heuristic already in queue")
		return
	}

	var heuristicRequest struct {
		Changed []dbHeuristic.FrontendHeuristicRequest `json:"changed,omitempty"`
		Deleted []string                               `json:"deleted,omitempty"`
	}

	if err := json.NewDecoder(body).Decode(&heuristicRequest); err != nil {
		reply.Msg = msgCouldNotDecodeRequest
		info(cliutil.ShowCallInfo(), err)
		return
	}

	if len(heuristicRequest.Changed) == 0 && len(heuristicRequest.Deleted) == 0 {
		reply.Msg = msgInvalidRequest
		return
	}

	work, err := heuristics.CreateWork(dgraph, txHashString, heuristicRequest.Changed,
		heuristicRequest.Deleted, userUID)
	if err != nil {
		reply.Msg = msgInvalidRequest
		info(cliutil.ShowCallInfo(), err)
		return
	}

	if worker.AddWork(txHashString, userUID, work) {
		reply.Status = heuristics.StatusHeuristicAdded
	} else {
		reply.Status = heuristics.StatusHeuristicDuplicate
	}

	reply.Success = true

	return
}

// getModifyUserReply parses the input and creates a corresponding userReply
func getModifyUserReply(dgraph external.Database, body io.Reader, tUser tokenUser) (reply backendUserReply) {
	// get clients user state
	var modRequest dbus.ModifyUserRequest
	if err := json.NewDecoder(body).Decode(&modRequest); err != nil {
		reply.Msg = msgCouldNotDecodeUser
		return
	}

	if len(modRequest.UID) == 0 ||
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
	if modRequest.UID != tUser.ID && !isAdmin {
		reply.Msg = "user ids do not match"
		info(cliutil.ShowCallInfo(), "user", tUser.ID, "tried to modify user", modRequest.UID)
		return
	}

	const msgErrModifyingUser = "error modifying user"

	// check current password if user is not an admin
	if !isAdmin {
		if len(modRequest.CurrentPassword) == 0 {
			reply.Msg = "current password must also be supplied"
			return
		}

		dbUser, err := dbus.GetUser(dgraph, modRequest.UID)
		if err != nil {
			reply.Msg = msgErrModifyingUser
			info(cliutil.ShowCallInfo(), err, modRequest)
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
				info(cliutil.ShowCallInfo(), err, modRequest)
				return
			}
		} else if emailUser.UID != modRequest.UID {
			reply.Msg = "duplicate email"
			info(cliutil.ShowCallInfo(), err, modRequest)
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
			reply.Msg = msgErrModifyingUser
			return
		}
	}

	// handle role change
	if len(modRequest.Roles) > 0 {
		if !isAdmin {
			reply.Msg = "user can not change its roles"
			info(cliutil.ShowCallInfo(), "user", tUser.ID, "tried to change its roles", modRequest.Roles)
			return
		}
		// check if all roles exists
		for _, r := range modRequest.Roles {
			if _, err := user.GetRoleByName(r.Name); err != nil {
				reply.Msg = "invalid role"
				info(cliutil.ShowCallInfo(), "user", tUser.ID, "provided invalid role", r.Name)
				return
			}
		}
		// delete existing roles if new roles are set
		if err := dbus.RemoveRolesFromUser(dgraph, modRequest.UID); err != nil {
			reply.Msg = msgErrModifyingUser
			info(cliutil.ShowCallInfo(), err, modRequest)
			return
		}
	}

	// modify user
	if err := dbus.ModifyUser(dgraph, modRequest.ToUser(newPwHash)); err != nil {
		reply.Msg = msgErrModifyingUser
		info(cliutil.ShowCallInfo(), err, modRequest)
		return
	}

	// get new user information
	newUserInfo, err := dbus.GetUser(dgraph, modRequest.UID)
	if err != nil {
		reply.Msg = msgErrModifyingUser
		info(cliutil.ShowCallInfo(), err, modRequest)
		return
	}

	// set new user info
	newUserState := newUserInfo.ToFrontendUserBackendState()
	reply.User = &newUserState
	reply.Success = true

	return
}

// getDeleteUserReply deletes delUid if is the same uid as tUser.ID or if tUser is an admin
func getDeleteUserReply(dgraph external.Database, delUID string, tUser tokenUser) (reply userReply) {
	if delUID != tUser.ID {
		// is user an admin
		isAdmin := false
		for _, r := range tUser.Roles {
			if r.Name == user.AdminRoleName {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			reply.Msg = "user can only delete his own account"
			info(tUser.ID, "tried to delete", delUID)
			return
		}
	}

	if err := dbus.DeleteUser(dgraph, delUID); err != nil {
		reply.Msg = "could not delete user"
		info(cliutil.ShowCallInfo(), err)
	}

	reply.Success = true

	return
}

// getShortestTransactionPathReply searches for the shortest path between two transactions
func getShortestTransactionPathReply(dgraph external.Database, body io.Reader) (reply shortestTransactionPathReply) {
	const msgErrorPathSearch = "error while searching for paths"

	// parse request
	var req dbHeuristic.ShortestTransactionPathRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		reply.Msg = msgCouldNotDecodeRequest
		return
	}

	if req.From == req.To {
		reply.Success = true
		reply.Msg = "Transaction hashes are equal"
		return
	}

	fromBlockID, err := dbtx.GetTransactionBlockID(dgraph, req.From)
	if err != nil {
		if errors.Is(err, dbtx.ErrTransactionNotFound) {
			reply.Success = true
			reply.Msg = "Transaction " + req.From + " does not exist"
			return
		}

		reply.Msg = msgErrorPathSearch
		info(cliutil.ShowCallInfo(), err)
		return
	}

	toBlockID, err := dbtx.GetTransactionBlockID(dgraph, req.To)
	if err != nil {
		if errors.Is(err, dbtx.ErrTransactionNotFound) {
			reply.Msg = "error transaction" + req.To + " does not exist"
			return
		}

		reply.Msg = msgErrorPathSearch
		info(cliutil.ShowCallInfo(), err)
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
	txs, err := dbHeuristic.GetShortestTransactionPathAnyDirection(dgraph, oldTx, youngTx,
		req.IncludePrivacyTransactions, anyDirection)
	if err != nil {
		reply.Msg = msgErrorPathSearch
		info(cliutil.ShowCallInfo(), err)
		return
	}

	if len(txs) == 0 {
		reply.Msg = "No path found"
	} else {
		reply.Transactions = txs
	}

	reply.Success = true

	return
}

// getDeleteHeuristicReply reads the data from body and constructs a deleteHeuristicReply
func getDeleteHeuristicReply(dgraph external.Database, body io.Reader, userUID string) (reply deleteHeuristicReply) {
	var req dbHeuristic.DeleteHeuristicRequest

	if err := json.NewDecoder(body).Decode(&req); err != nil {
		reply.Msg = msgCouldNotDecodeRequest
		return
	}

	if (req.DeleteAll && len(req.TransactionHash) > 0) ||
		(!req.DeleteAll && len(req.TransactionHash) == 0) {
		reply.Msg = msgInvalidRequest
		return
	}

	if req.DeleteAll {
		if err := dbHeuristic.DeleteAllUserHeuristics(dgraph, userUID); err != nil {
			if errors.Is(err, dbHeuristic.ErrNoMutationHappened) {
				reply.Msg = "No data was deleted. The user may not have any heuristics."
			} else {
				reply.Msg = "could not delete data"
				info(cliutil.ShowCallInfo(), err)
			}
			return
		}

		reply.Success = true
		return
	}

	if err := dbHeuristic.DeleteAllUserTxHeuristics(dgraph, req.TransactionHash, userUID); err != nil {
		if errors.Is(err, dbHeuristic.ErrNoMutationHappened) {
			reply.Msg = "No data was deleted. The transaction may not have any heuristics."
		} else {
			reply.Msg = "could not delete data"
			info(cliutil.ShowCallInfo(), err)
		}
		return
	}

	reply.Success = true

	return
}

// getConnectionLookupReply returns the result of a reverse lookup
func getConnectionLookupReply(dgraph external.Database, worker *heuristics.Worker, urlValues url.Values,
	urlPath string) (reply connectionLookupReply) {

	if !worker.IsReady() {
		reply.Msg = "Server is not ready to receive connection lookups. Please try again later."
		reply.Warning = true
		return
	}

	// get time parameter
	fLockBackTime := urlValues.Get("t")
	var lookBackTime time.Duration
	if len(fLockBackTime) > 0 {
		n, err := strconv.Atoi(fLockBackTime)
		if err != nil {
			reply.Msg = "error parsing input"
			info(cliutil.ShowCallInfo(), err)
			return
		}

		lookBackTime = time.Duration(n)
	}

	// get direction parameter
	direction := urlValues.Get("forward")
	var isLookupForward bool
	// direction is either "0" or "1", thus check for a string with length equal to 1.
	if len(direction) == 1 {
		n, err := strconv.Atoi(direction)
		if err != nil {
			reply.Msg = "error parsing input"
			info(cliutil.ShowCallInfo(), err)
			return
		}

		isLookupForward = n == 1
	}

	txhash := urlPath[len(constants.GetRouteConnectionLookup()):]

	uid, err := dbtx.GetTransactionUID(dgraph, txhash)
	if err != nil {
		if errors.Is(err, dbtx.ErrTransactionNotFound) {
			reply.Success = true
			reply.Msg = "Transaction " + txhash + " does not exist"
			return
		}

		reply.Msg = "error while searching for connections"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	info("Reverse Lookup for", txhash, "look back time (days):", int(lookBackTime))

	const msgLookupNotSuccessful = "Lookup not successful"

	var endpoints map[string]bool
	rLookupTime := time.Now()
	if isLookupForward {
		endpoints, err = worker.ForwardLookup(uid, time.Hour*24*lookBackTime)
		if err != nil {
			reply.Msg = msgLookupNotSuccessful
			info(cliutil.ShowCallInfo(), err)
			return
		}
	} else {
		endpoints, err = worker.ReverseLookup(uid, time.Hour*24*lookBackTime)
		if err != nil {
			reply.Msg = msgLookupNotSuccessful
			info(cliutil.ShowCallInfo(), err)
			return
		}
	}

	info("time:", time.Since(rLookupTime), "endpoints: origins:", len(endpoints))

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

	frontendTransactions, err := dbtx.GetFrontendTransactionsByUID(dgraph, transactionUids)
	if err != nil {
		reply.Msg = msgLookupNotSuccessful
		info(cliutil.ShowCallInfo(), err)
		return
	}

	endpointCount := len(endpoints)
	reply.TransactionCount = &endpointCount
	reply.Transactions = frontendTransactions
	reply.Success = true

	return
}

// getFrontendCluster returns the requested (by body) clusters. In case an error occurred msg and err is filled.
func getFrontendCluster(dgraph external.Database, body io.Reader, maxAddresses int,
	userID string) (clusters []clustering.FrontendCluster, msg string, err error) {
	// parse request
	var req clustering.ClusterLookupRequest
	if decodeErr := json.NewDecoder(body).Decode(&req); decodeErr != nil {
		msg = msgCouldNotDecodeRequest
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), decodeErr)
		return
	}

	if req.AddressHash == "" {
		msg = "address hash is empty"
		return
	}

	if !isValid(req.AddressHash) {
		msg = "address hash was not valid"
		return
	}

	clusterResponse, getErr := clustering.GetClusters(dgraph, req.AddressHash, maxAddresses, userID)
	if getErr != nil {
		msg = "error while searching for clusters"
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), getErr)
		return
	}
	clusters = clusterResponse

	return
}

// getClusterLookupReply returns the result of a cluster lookup
func getClusterLookupReply(dgraph external.Database, body io.Reader, user tokenUser) (reply clusterLookupReply) {
	const maxAddresses = 30

	clusters, msg, err := getFrontendCluster(dgraph, body, maxAddresses, user.ID)
	reply.Msg = msg
	if err != nil {
		info(err)
		return
	}

	reply.Clusters = clusters
	reply.Success = true

	return
}

// getHMILookupReply returns all hmi clusters connected to the given address hash
func getHMILookupReply(dgraph external.Database, addressHash string) (reply hmiLookupReply) {
	addressCluster, clusters, err := clustering.GetHMIClusters(dgraph, addressHash)
	if err != nil {
		info(cliutil.ShowCallInfo(), err)
		return reply
	}
	reply.Success = true
	reply.Clusters = clusters
	reply.AddressCluster = addressCluster

	return reply
}

// writeHeuristicSummary writes heuristic data in CSV format
func writeHeuristicSummary(w http.ResponseWriter, dgraph external.Database, tUser tokenUser, heuristicUID string) {
	cHeuristic, err := dbHeuristic.GetFrontendHeuristicByUID(dgraph, heuristicUID, tUser.ID)
	if err != nil {
		handleError(w, err)
		info(cliutil.ShowCallInfo(), err)
		return
	}

	if cHeuristic.UID == "" {
		http.Error(w, errorHeuristicSummary, http.StatusNotFound)
		return
	}

	// headers for streaming data to client
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", heuristicUID))
	w.Header().Set("Content-Type", "text/csv")

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	header := []string{"cluster ID", "attributions", "origin transaction hash",
		"origin timestamp", "destination count"}

	if err = csvWriter.Write(header); err != nil {
		http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
		info(cliutil.ShowCallInfo(), err)
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
				handleError(w, err)
				info(cliutil.ShowCallInfo(), err)
				return
			}
		}
		csvWriter.Flush()
	}

	csvWriter.Flush()
}

// writeClusterSummary writes heuristic data in CSV format
func writeClusterSummary(w http.ResponseWriter, r *http.Request, dgraph external.Database) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		http.Error(w, errorClusterSummary, http.StatusNotFound)
		return
	}

	clusters, msg, err := getFrontendCluster(dgraph, r.Body, 0, tUser.ID)
	if err != nil {
		handleError(w, err)
		info(msg)
		return
	}

	if len(clusters) == 0 {
		http.Error(w, errorClusterSummary, http.StatusNotFound)
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
		http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
		info(cliutil.ShowCallInfo(), err)
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
				handleError(w, err)
			}
		}
		csvWriter.Flush()
	}
}

// getMixingActivity returns the result of a mixing activity lookup
func getMixingActivity(dgraph external.Database, body io.Reader) (reply mixingActivityReply) {
	var req struct {
		// AddressHash is the address hash for which the lookup will be done
		AddressHash string `json:"addressHash,omitempty"`
		// IsClusterLookup determines if all addresses of the cluster will be considered
		IsClusterLookup bool `json:"isClusterLookup,omitempty"`
	}
	if decodeErr := json.NewDecoder(body).Decode(&req); decodeErr != nil {
		info(cliutil.ShowCallInfo(), decodeErr)
		return
	}
	const maxAddressCount = 2000
	if req.IsClusterLookup {
		addressCount, err := clustering.GetClusterAddressCount(dgraph, req.AddressHash)
		if err != nil {
			info(cliutil.ShowCallInfo(), err)
			return
		}

		if addressCount > maxAddressCount {
			reply.Msg = "too_many_addresses"
			reply.Success = true
			return
		}
	}

	activities, err := dbAnalytics.GetMixingActivity(dgraph, req.AddressHash, req.IsClusterLookup)
	if err != nil {
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Activities = activities
	reply.Success = true

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

func getAddClusterReply(dgraph external.Database, r *http.Request) (reply addClusterReply) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		reply.Msg = msgUserNotFound
		info(cliutil.ShowCallInfo(), err)
		return
	}

	if err := r.ParseMultipartForm(maxBodySize); err != nil {
		return
	}

	separator := r.FormValue("separator")
	if separator == "" {
		reply.Msg = CsvInvalidSeparator
		return
	}

	var rSeparator rune
	if separator != ";" && separator != "," {
		reply.Msg = CsvInvalidSeparator
		return
	}
	rSeparator = []rune(separator)[0]

	headerFlag := r.FormValue("hasHeader")
	if headerFlag == "" {
		reply.Msg = CsvEmptyHeader
		return
	}

	// Get handler for filename, size and headers
	file, _, err := r.FormFile("file")
	if err != nil {
		reply.Msg = CsvReadError
		return
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			info("Error closing CSV-file")
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
				return
			} else if !errors.Is(err, io.EOF) {
				reply.Msg = CsvInvalidData
				return
			}
			break
		}

		if index == 0 && headerFlag == "1" {
			continue
		}

		newAddress := analyticsClustering.ExternalClusterItem{
			ClusterID:   strings.TrimSpace(line[0]),
			AddressHash: strings.TrimSpace(line[1]),
		}

		if newAddress.ClusterID == "" || newAddress.AddressHash == "" {
			reply.Msg = CsvInvalidData
			return
		}

		addresses = append(addresses, newAddress)
	}

	if len(addresses) == 0 {
		reply.Msg = CsvNoData
		return
	}

	if err := analyticsClustering.ImportCluster(dgraph, addresses, tUser.ID); err != nil {
		switch {
		case errors.Is(err, analyticsClustering.ErrTooManyAddresses):
			reply.Msg = CsvTooManyAddresses
		case errors.Is(err, analyticsClustering.ErrShallowCluster):
			reply.Msg = CsvShallowCluster
		case errors.Is(err, analyticsClustering.ErrNonExistentAddress):
			reply.Msg = err.Error()
		default:
			reply.Msg = CsvErrorImporting
			info(err)
		}

		return
	}

	reply.Success = true
	return
}

func getAddAttributionReply(dgraph external.Database, r *http.Request, isPublic bool) (reply addAttributionReply) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		reply.Msg = msgUserNotFound
		info(cliutil.ShowCallInfo(), err)
		return
	}

	if err := r.ParseMultipartForm(maxBodySize); err != nil {
		return
	}

	separator := r.FormValue("separator")
	if separator == "" {
		reply.Msg = CsvInvalidSeparator
		return
	}

	var rSeparator rune
	if separator != ";" && separator != "," {
		reply.Msg = CsvInvalidSeparator
		return
	}
	rSeparator = []rune(separator)[0]

	headerFlag := r.FormValue("hasHeader")
	if headerFlag == "" {
		reply.Msg = CsvEmptyHeader
		return
	}

	// Get handler for filename, size and headers
	file, _, err := r.FormFile("file")
	if err != nil {
		reply.Msg = CsvReadError
		return
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			info("Error closing CSV-file")
		}
	}(file)

	csvReader := csv.NewReader(file)
	csvReader.ReuseRecord = true
	csvReader.Comma = rSeparator
	csvReader.FieldsPerRecord = 5
	var line []string

	var addresses []analytics.Attribution
	var index int
	for ; ; index++ {
		line, err = csvReader.Read()
		if err != nil {
			if errors.Is(err, csv.ErrFieldCount) {
				reply.Msg = CsvInvalidFieldCount
				return
			} else if !errors.Is(err, io.EOF) {
				reply.Msg = CsvInvalidData
				return
			}
			break
		}

		if index == 0 && headerFlag == "1" {
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
			return
		}

		addresses = append(addresses, newAttribution)
	}

	if len(addresses) == 0 {
		reply.Msg = CsvNoData
		return
	}

	if err := analytics.ImportAttribution(dgraph, addresses, tUser.ID, isPublic); err != nil {
		switch {
		case errors.Is(err, analytics.ErrTooManyAddresses):
			reply.Msg = CsvTooManyAddresses
		case errors.Is(err, analytics.ErrNonExistentAddress):
			reply.Msg = err.Error()
		default:
			reply.Msg = CsvErrorImporting
			info(err)
		}

		return
	}

	reply.Success = true
	return
}

func getClusterOverviewReply(dgraph external.Database, userUID string) (reply clusterOverviewReply) {
	clusters, err := clustering.GetUserClusters(dgraph, userUID)
	if err != nil {
		reply.Msg = "no clusters found"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true
	reply.Clusters = clusters

	return
}

func getDeleteClusterReply(dgraph external.Database, userUID string, clusterUID string) (reply deleteClusterReply) {
	if clusterUID == "" {
		reply.Msg = "cluster uid was not set"
		return
	}

	if err := clustering.DeleteCluster(dgraph, userUID, clusterUID); err != nil {
		reply.Msg = "could not delete cluster"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true

	return
}

func getDeleteAllClustersReply(dgraph external.Database, userUID string) (reply deleteClusterReply) {
	if err := clustering.DeleteAllClusters(dgraph, userUID); err != nil {
		reply.Msg = "could not delete clusters"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true

	return
}

func getAttributionOverviewReply(dgraph external.Database, userUID string) (reply attributionOverviewReply) {
	attributions, err := attribution.GetUserAttributions(dgraph, userUID)
	if err != nil {
		reply.Msg = "no attributions found"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true
	reply.Attributions = attributions

	return
}

func getDeleteAttributionReply(dgraph external.Database, userUID string,
	attributionUID string, isPublicDeletion bool) (reply deleteAttributionReply) {
	if attributionUID == "" {
		reply.Msg = "attribution uid was not set"
		return
	}

	var err error
	if isPublicDeletion {
		err = attribution.DeletePublicAttribution(dgraph, attributionUID)
	} else {
		err = attribution.DeletePrivateAttribution(dgraph, userUID, attributionUID)
	}

	if err != nil {
		reply.Msg = "could not delete attribution"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true

	return
}

func getDeleteAllAttributionsReply(dgraph external.Database, userUID string) (reply deleteAttributionReply) {
	if err := attribution.DeleteAllAttributions(dgraph, userUID); err != nil {
		reply.Msg = "could not delete clusters"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true

	return
}

func getAttributionSearchReply(dgraph external.Database, userUID string,
	body io.ReadCloser) (reply attributionOverviewReply) {
	var searchRequest struct {
		Query string `json:"q,omitempty"`
	}

	if err := json.NewDecoder(body).Decode(&searchRequest); err != nil {
		reply.Success = false
		reply.Msg = "error decoding request"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	if searchRequest.Query == "" {
		reply.Success = false
		reply.Msg = "empty query string"
		return
	}

	attributions, err := attribution.SearchAttributions(dgraph, userUID, searchRequest.Query)
	if err != nil {
		reply.Msg = "no attributions found"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true
	reply.Attributions = attributions

	return
}

func getAddAddressExclusionsReply(dgraph external.Database, r *http.Request) (reply addAddressExclusionsReply) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		reply.Msg = msgUserNotFound
		info(cliutil.ShowCallInfo(), err)
		return
	}

	if err := r.ParseMultipartForm(maxBodySize); err != nil {
		return
	}

	// Get handler for filename, size and headers
	file, _, err := r.FormFile("file")
	if err != nil {
		reply.Msg = CsvReadError
		return
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			info("Error closing CSV-file")
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
				return
			} else if !errors.Is(err, io.EOF) {
				reply.Msg = CsvInvalidData
				return
			}
			break
		}

		trimmed := strings.TrimSpace(line[0])

		if trimmed == "" {
			reply.Msg = CsvInvalidData
			return
		}

		addresses = append(addresses, trimmed)
	}

	if len(addresses) == 0 {
		reply.Msg = CsvNoData
		return
	}

	if err := analytics.ImportAddressExclusions(dgraph, addresses, tUser.ID); err != nil {
		switch {
		case errors.Is(err, analytics.ErrTooManyAddresses):
			reply.Msg = CsvTooManyAddresses
		case errors.Is(err, analytics.ErrNonExistentAddress):
			reply.Msg = err.Error()
		default:
			reply.Msg = CsvErrorImporting
			info(err)
		}

		return
	}

	reply.Success = true
	return
}

func getAddressExclusionOverviewReply(dgraph external.Database, userUID string) (reply addressExclusionOverviewReply) {
	addresses, count, err := exclusion.GetAddressExclusions(dgraph, userUID)
	if err != nil {
		reply.Msg = "no addresses found"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true
	reply.AddressHashes = addresses
	reply.Count = count

	return
}

func getDeleteAddressExclusionReply(dgraph external.Database, userUID string,
	addressHash string) (reply deleteAddressExclusionReply) {
	if addressHash == "" {
		reply.Msg = "address hash was not set"
		return
	}

	if err := exclusion.DeleteAddressExclusion(dgraph, userUID, addressHash); err != nil {
		reply.Msg = "could not delete address exclusion"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true

	return
}

func getDeleteAllAddressExclusionsReply(dgraph external.Database, userUID string) (reply deleteAddressExclusionReply) {
	if err := exclusion.DeleteAllAddressExclusions(dgraph, userUID); err != nil {
		reply.Msg = "could not delete address exclusions"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true

	return
}

func getAddressExclusionStatusReply(r *http.Request, dgraph external.Database, addressHash string) (
	reply addressExclusionStatusReply) {
	if !isValid(addressHash) {
		reply.Msg = "address hash is not valid"
		return
	}

	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		reply.Msg = msgUserNotFound
		info(cliutil.ShowCallInfo(), err)
		return
	}

	status, err := exclusion.GetAddressExclusionStatus(dgraph, addressHash, tUser.ID)
	if err != nil {
		reply.Msg = "error getting exclusion status"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true
	reply.IsExclusion = status

	return
}
