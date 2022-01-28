package server

import (
	analytics2 "backend/analytics"
	analyticsClustering "backend/analytics/clustering"
	heuristic "backend/analytics/heuristics/transaction"
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db/analytics"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"backend/db/analytics/heuristics/transaction"
	dbtx "backend/db/transaction"
	dbus "backend/db/user"
	"backend/external"
	"backend/user"
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

// getLoginReply reads the data from body and constructs a backendUserReply
func getLoginReply(dgraph external.Database, body io.Reader) (reply backendUserReply) {
	const invalidUserData = "email and password combination does not match"

	var loginData dbus.FrontendUserLogin

	if err := json.NewDecoder(body).Decode(&loginData); err != nil {
		reply.Msg = "could not decode user data"
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
		reply.Msg = "could not decode user data"
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

func getHeuristicReply(dgraph external.Database, worker *heuristic.Worker,
	txHashString string, userUID string) (reply heuristicReply) {

	heuristics, err := transaction.GetBasicFrontendHeuristic(dgraph, txHashString, userUID)
	if err != nil {
		reply.Msg = "no heuristics found"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true
	reply.Heuristics = heuristics
	reply.Status = worker.GetStatus(txHashString, userUID)

	return
}

func getHeuristicExecutionReply(dgraph external.Database, worker *heuristic.Worker, body io.Reader,
	txHashString string, userUID string) (reply heuristicExecutionReply) {
	if !worker.IsReady() {
		reply.Success = true
		reply.Status = heuristic.StatusHeuristicWorkerNotReady
		return
	}

	if worker.IsInQueue(txHashString, userUID) {
		reply.Success = true
		reply.Status = heuristic.StatusHeuristicDuplicate
		info(cliutil.ShowCallInfo(), "heuristic already in queue")
		return
	}

	type request struct {
		Changed []transaction.FrontendHeuristic `json:"changed,omitempty"`
		Deleted []string                        `json:"deleted,omitempty"`
	}

	var heuristicRequest request

	decoder := json.NewDecoder(body)
	err := decoder.Decode(&heuristicRequest)
	if err != nil {
		reply.Msg = "could not decode request data"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	if len(heuristicRequest.Changed) == 0 && len(heuristicRequest.Deleted) == 0 {
		reply.Msg = "invalid request"
		return
	}

	work, err := heuristic.CreateWork(dgraph, txHashString, heuristicRequest.Changed,
		heuristicRequest.Deleted)
	if err != nil {
		reply.Msg = "invalid request"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	addedWork := worker.AddWork(txHashString, userUID, work)

	if addedWork {
		reply.Status = heuristic.StatusHeuristicAdded
	} else {
		reply.Status = heuristic.StatusHeuristicDuplicate
	}

	reply.Success = true

	return
}

// getModifyUserReply parses the input and creates a corresponding userReply
func getModifyUserReply(dgraph external.Database, body io.Reader, tUser tokenUser) (reply backendUserReply) {
	// get clients user state
	var modRequest dbus.ModifyUserRequest
	if err := json.NewDecoder(body).Decode(&modRequest); err != nil {
		reply.Msg = "could not decode user data"
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

	// check current password if user is not an admin
	if !isAdmin {
		if len(modRequest.CurrentPassword) == 0 {
			reply.Msg = "current password must also be supplied"
			return
		}

		dbUser, err := dbus.GetUser(dgraph, modRequest.UID)
		if err != nil {
			reply.Msg = "error modifying user"
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
			reply.Msg = "error modifying user"
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
			reply.Msg = "error modifying user"
			info(cliutil.ShowCallInfo(), err, modRequest)
			return
		}
	}

	// modify user
	if err := dbus.ModifyUser(dgraph, modRequest.ToUser(newPwHash)); err != nil {
		reply.Msg = "error modifying user"
		info(cliutil.ShowCallInfo(), err, modRequest)
		return
	}

	// get new user information
	newUserInfo, err := dbus.GetUser(dgraph, modRequest.UID)
	if err != nil {
		reply.Msg = "error modifying user"
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
	// parse request
	var req transaction.ShortestTransactionPathRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		reply.Msg = "could not decode request data"
		return
	}

	if req.From == req.To {
		reply.Success = true
		reply.Msg = "Transaction hashes are equal"
		return
	}

	fromBlockID, err := dbtx.GetTransactionBlockID(dgraph, req.From)
	if err != nil {
		if errors.Is(err, dbtx.ErrorTransactionNotFound) {
			reply.Success = true
			reply.Msg = "Transaction " + req.From + " does not exist"
			return
		}

		reply.Msg = "error while searching for paths"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	toBlockID, err := dbtx.GetTransactionBlockID(dgraph, req.To)
	if err != nil {
		if errors.Is(err, dbtx.ErrorTransactionNotFound) {
			reply.Msg = "error transaction" + req.To + " does not exist"
			return
		}

		reply.Msg = "error while searching for paths"
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

	// do shortest transaction path lookup
	txs, err := transaction.GetShortestTransactionPathAnyDirection(dgraph, oldTx, youngTx,
		req.IncludePrivacyTransactions, anyDirection)
	if err != nil {
		reply.Msg = "error while searching for paths"
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
	var req transaction.DeleteHeuristicRequest

	if err := json.NewDecoder(body).Decode(&req); err != nil {
		reply.Msg = "could not decode request data"
		return
	}

	if (req.DeleteAll && len(req.TransactionHash) > 0) ||
		(!req.DeleteAll && len(req.TransactionHash) == 0) {
		reply.Msg = "invalid request"
		return
	}

	if req.DeleteAll {
		if err := transaction.DeleteAllUserHeuristics(dgraph, userUID); err != nil {
			if errors.Is(err, transaction.ErrNoMutationHappened) {
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

	if err := transaction.DeleteAllUserTxHeuristics(dgraph, req.TransactionHash, userUID); err != nil {
		if errors.Is(err, transaction.ErrNoMutationHappened) {
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
func getConnectionLookupReply(dgraph external.Database, worker *heuristic.Worker, urlValues url.Values,
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
		if errors.Is(err, dbtx.ErrorTransactionNotFound) {
			reply.Success = true
			reply.Msg = "Transaction " + txhash + " does not exist"
			return
		}

		reply.Msg = "error while searching for connections"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	info("Reverse Lookup for", txhash, "look back time (days):", int(lookBackTime))

	var endpoints map[string]bool
	rLookupTime := time.Now()
	if isLookupForward {
		endpoints, err = worker.ForwardLookup(uid, time.Hour*24*lookBackTime)
		if err != nil {
			reply.Msg = "Lookup not successful"
			info(cliutil.ShowCallInfo(), err)
			return
		}
	} else {
		endpoints, err = worker.ReverseLookup(uid, time.Hour*24*lookBackTime)
		if err != nil {
			reply.Msg = "Lookup not successful"
			info(cliutil.ShowCallInfo(), err)
			return
		}
	}

	info("time:", time.Since(rLookupTime), "endpoints: origins:", len(endpoints))

	// reply with the first 30 endpoints
	var transactionUids []string
	const numOutputNodes = 30
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
		reply.Msg = "Lookup not successful"
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
func getFrontendCluster(dgraph external.Database, body io.Reader, maxAddresses int) (clusters []clustering.FrontendCluster, msg string, err error) {
	// parse request
	var req clustering.ClusterLookupRequest
	if decodeErr := json.NewDecoder(body).Decode(&req); decodeErr != nil {
		msg = "could not decode request data"
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), decodeErr)
		return
	}

	if req.AddressHash == "" {
		msg = "address hash was not provided"
		return
	}

	if !isValid(req.AddressHash) {
		msg = "address hash was not valid"
		return
	}

	clusterResponse, getErr := clustering.GetClusters(dgraph, req.AddressHash, maxAddresses)
	if getErr != nil {
		msg = "error while searching for clusters"
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), getErr)
		return
	}
	clusters = clusterResponse

	return
}

// getClusterLookupReply returns the result of a cluster lookup
func getClusterLookupReply(dgraph external.Database, body io.Reader) (reply clusterLookupReply) {
	const maxAddresses = 30

	clusters, msg, err := getFrontendCluster(dgraph, body, maxAddresses)
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
func writeHeuristicSummary(w http.ResponseWriter, dgraph external.Database, tUser tokenUser, txHashString string) {
	cHeuristic, err := transaction.GetFrontendHeuristic(dgraph, txHashString, tUser.ID)
	if err != nil {
		handleError(w, err)
		return
	}

	if len(cHeuristic.Heuristics) == 0 {
		http.Error(w, errorHeuristicSummary, http.StatusNotFound)
		return
	}

	// headers for streaming data to client
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", txHashString))
	w.Header().Set("Content-Type", "text/csv")

	// somehow both content-length and transfer-encoding headers are both set, so one must be removed
	//w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	header := []string{"heuristic uid", "parent heuristic uid", "child heuristic uid",
		"heuristic type", "heuristic parameter", "heuristic timestamp",
		"origin uid", "origin transaction hash", "origin timestamp",
		"origin address hash", "destination uid", "destination transaction hash", "destination timestamp"}

	if err = csvWriter.Write(header); err != nil {
		http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
		info(cliutil.ShowCallInfo(), err)
	}

	for _, h := range cHeuristic.Heuristics {
		for _, result := range h.Results {
			var row []string
			// per heuristic information
			row = append(row, h.UID)
			var parentHeuristic string
			if len(h.ParentHeuristic) > 0 {
				// only one parent heuristic is possible
				parentHeuristic = h.ParentHeuristic[0].UID
			}
			row = append(row, parentHeuristic)

			var childHeuristics string
			for i, c := range h.ChildHeuristics {
				childHeuristics += c.UID
				if i+1 < len(h.ChildHeuristics) {
					childHeuristics += ","
				}
			}

			row = append(row, childHeuristics)
			row = append(row, h.Type)
			row = append(row, h.Parameter)
			row = append(row, h.Timestamp)

			// per origin information
			row = append(row, result.Origin.UID)
			row = append(row, result.Origin.TxHash)
			row = append(row, result.Origin.Timestamp)
			row = append(row, result.Origin.AddressHash)

			// add destination data if there exists any
			if len(result.Destinations) > 0 {
				for _, d := range result.Destinations {
					withDestinations := make([]string, len(row))
					// copy because for each destination the row gets reused
					copy(withDestinations, row)

					withDestinations = append(withDestinations, d.UID)
					withDestinations = append(withDestinations, d.TxHash)
					withDestinations = append(withDestinations, d.Timestamp)

					if err = csvWriter.Write(withDestinations); err != nil {
						handleError(w, err)
					}
				}
				csvWriter.Flush()
			} else {
				if err = csvWriter.Write(row); err != nil {
					handleError(w, err)
				}
			}
		}
		csvWriter.Flush()
	}
}

// writeClusterSummary writes heuristic data in CSV format
func writeClusterSummary(w http.ResponseWriter, r *http.Request, dgraph external.Database) {
	clusters, msg, err := getFrontendCluster(dgraph, r.Body, 0)
	if err != nil {
		handleError(w, err)
		info(msg)
		return
	}

	if len(clusters) == 0 {
		http.Error(w, errorHeuristicSummary, http.StatusNotFound)
		return
	}

	// headers for streaming data to client
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=cluster_lookup_%s.csv",
		time.Now().Format("2006-01-02T15:04:05")))
	w.Header().Set("Content-Type", "text/csv")

	// somehow both content-length and transfer-encoding headers are both set, so one must be removed
	//w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	header := []string{"cluster type", "last cluster update (transaction)", "last cluster update (timestamp)", "address hash", "output count", "unspent output count"}

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

	activities, err := analytics.GetMixingActivity(dgraph, req.AddressHash, req.IsClusterLookup)
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
		reply.Msg = "User not found"
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
	} else {
		rSeparator = []rune(separator)[0]
	}

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
			ClusterID:   line[0],
			AddressHash: line[1],
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
		if errors.Is(err, analyticsClustering.ErrTooManyAddresses) {
			reply.Msg = CsvTooManyAddresses
		} else if errors.Is(err, analyticsClustering.ErrShallowCluster) {
			reply.Msg = CsvShallowCluster
		} else if errors.Is(err, analyticsClustering.ErrNonExistentAddress) {
			reply.Msg = err.Error()
		} else {
			reply.Msg = CsvErrorImporting
			info(err)
		}

		return
	}

	reply.Success = true
	return
}

func getAddAttributionReply(dgraph external.Database, r *http.Request) (reply addAttributionReply) {
	tUser, err := extractTokenUser(r.Context())
	if err != nil {
		reply.Msg = "User not found"
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
	} else {
		rSeparator = []rune(separator)[0]
	}

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

	var addresses []analytics2.Attribution
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

		newAttribution := analytics2.Attribution{
			AddressHash: line[0],
			Tag:         line[1],
			Description: line[2],
			Source:      line[3],
			Category:    line[4],
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

	if err := analytics2.ImportAttribution(dgraph, addresses, tUser.ID); err != nil {
		if errors.Is(err, analyticsClustering.ErrTooManyAddresses) {
			reply.Msg = CsvTooManyAddresses
		} else if errors.Is(err, analyticsClustering.ErrShallowCluster) {
			reply.Msg = CsvShallowCluster
		} else if errors.Is(err, analyticsClustering.ErrNonExistentAddress) {
			reply.Msg = err.Error()
		} else {
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

func getDeleteAttributionReply(dgraph external.Database, userUID string, attributionUid string) (reply deleteAttributionReply) {
	if attributionUid == "" {
		reply.Msg = "attribution uid was not set"
		return
	}

	if err := attribution.DeleteAttribution(dgraph, userUID, attributionUid); err != nil {
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

func getAttributionSearchReply(dgraph external.Database, userUID string, body io.ReadCloser) (reply attributionOverviewReply) {
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
