package server

import (
	heuristic "backend/analytics/heuristics/transaction"
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db/address"
	"backend/db/analytics/heuristics/transaction"
	dbtx "backend/db/transaction"
	dbus "backend/db/user"
	"backend/external"
	"backend/user"

	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"
	"time"
)

// getLoginReply reads the data from body and constructs a backendUserReply
func getLoginReply(dgraph *external.GraphDB, body io.Reader) (reply backendUserReply) {
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
func getCreateUserReply(dgraph *external.GraphDB, body io.Reader) (reply userReply) {
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

func getHeuristicReply(dgraph *external.GraphDB, worker *heuristic.Worker,
	txHashString string, userUid string) (reply heuristicReply) {

	heuristics, err := transaction.GetBasicFrontendHeuristic(dgraph, txHashString, userUid)
	if err != nil {
		reply.Msg = "no heuristics found"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Success = true
	reply.Heuristics = heuristics
	reply.Status = worker.GetStatus(txHashString, userUid)

	return
}

func getHeuristicExecutionReply(dgraph *external.GraphDB, worker *heuristic.Worker, body io.Reader,
	txHashString string, userUid string) (reply heuristicExecutionReply) {
	if !worker.IsReady() {
		reply.Success = true
		reply.Status = heuristic.StatusHeuristicWorkerNotReady
		return
	}

	if worker.IsInQueue(txHashString, userUid) {
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

	addedWork := worker.AddWork(txHashString, userUid, work)

	if addedWork {
		reply.Status = heuristic.StatusHeuristicAdded
	} else {
		reply.Status = heuristic.StatusHeuristicDuplicate
	}

	reply.Success = true

	return
}

// getModifyUserReply parses the input and creates a corresponding userReply
func getModifyUserReply(dgraph *external.GraphDB, body io.Reader, tUser tokenUser) (reply backendUserReply) {
	// get clients user state
	var modRequest dbus.ModifyUserRequest
	if err := json.NewDecoder(body).Decode(&modRequest); err != nil {
		reply.Msg = "could not decode user data"
		return
	}

	if len(modRequest.Uid) == 0 ||
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
	if modRequest.Uid != tUser.Id && !isAdmin {
		reply.Msg = "user ids do not match"
		info(cliutil.ShowCallInfo(), "user", tUser.Id, "tried to modify user", modRequest.Uid)
		return
	}

	// check current password if user is not an admin
	if !isAdmin {
		if len(modRequest.CurrentPassword) == 0 {
			reply.Msg = "current password must also be supplied"
			return
		}

		dbUser, err := dbus.GetUser(dgraph, modRequest.Uid)
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
		} else if emailUser.Uid != modRequest.Uid {
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
			info(cliutil.ShowCallInfo(), "user", tUser.Id, "tried to change its roles", modRequest.Roles)
			return
		}
		// check if all roles exists
		for _, r := range modRequest.Roles {
			if _, err := user.GetRoleByName(r.Name); err != nil {
				reply.Msg = "invalid role"
				info(cliutil.ShowCallInfo(), "user", tUser.Id, "provided invalid role", r.Name)
				return
			}
		}
		// delete existing roles if new roles are set
		if err := dbus.RemoveRolesFromUser(dgraph, modRequest.Uid); err != nil {
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
	newUserInfo, err := dbus.GetUser(dgraph, modRequest.Uid)
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

// getDeleteUserReply deletes delUid if is the same uid as tUser.Id or if tUser is an admin
func getDeleteUserReply(dgraph *external.GraphDB, delUid string, tUser tokenUser) (reply userReply) {
	if delUid != tUser.Id {
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
			info(tUser.Id, "tried to delete", delUid)
			return
		}
	}

	if err := dbus.DeleteUser(dgraph, delUid); err != nil {
		reply.Msg = "could not delete user"
		info(cliutil.ShowCallInfo(), err)
	}

	reply.Success = true

	return
}

// getShortestTransactionPathReply searches for the shortest path between two transactions
func getShortestTransactionPathReply(dgraph *external.GraphDB, body io.Reader) (reply shortestTransactionPathReply) {
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

	fromBlockId, err := dbtx.GetTransactionBlockId(dgraph, req.From)
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

	toBlockId, err := dbtx.GetTransactionBlockId(dgraph, req.To)
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

	if fromBlockId == toBlockId {
		// set anyDirection to true, as the direction can not be calculated from the block ids
		// and as the transactions are in the same block the query should be very quick
		anyDirection = true
	}

	oldTx := req.From
	youngTx := req.To

	if !req.AnyDirection {
		// switch transactions if necessary so we are searching in the right direction
		if toBlockId > fromBlockId {
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
func getDeleteHeuristicReply(dgraph *external.GraphDB, body io.Reader, userUid string) (reply deleteHeuristicReply) {
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
		if err := transaction.DeleteAllUserHeuristics(dgraph, userUid); err != nil {
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

	if err := transaction.DeleteAllUserTxHeuristics(dgraph, req.TransactionHash, userUid); err != nil {
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
func getConnectionLookupReply(dgraph *external.GraphDB, worker *heuristic.Worker, urlValues url.Values,
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

	uid, err := dbtx.GetTransactionUid(dgraph, txhash)
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
		endpoints, err = worker.ForwardLookupByTime(uid, time.Hour*24*lookBackTime)
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

	// ------------- Forward lookup start -------------
	//endpointCount := 0
	//var completeDuration time.Duration
	//
	//for k := range endpoints {
	//	fLookupTime := time.Now()
	//	fEndpoints, forwardErr := worker.ForwardLookup(k, uid)
	//	completeDuration += time.Since(fLookupTime)
	//	if forwardErr != nil {
	//		reply.Msg = "Forward lookup not successful"
	//		info(cliutil.ShowCallInfo(), err)
	//		return
	//	}
	//	endpointCount += len(fEndpoints)
	//
	//	if len(fEndpoints) < 20 {
	//		// todo remove
	//		info("forward look up", k, "mem endpoints -- time:", completeDuration, "origins:", len(fEndpoints))
	//	}
	//}
	//info("forward look up -- avg. time:", completeDuration/time.Duration(len(endpoints)),
	//	"avg. endpoints per lookup", endpointCount/len(endpoints))

	// ------------- Forward lookup end -------------

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

	frontendTransactions, err := dbtx.GetFrontendTransactionsByUid(dgraph, transactionUids)
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

// getClusterLookupReply returns the result of a cluster lookup
func getClusterLookupReply(dgraph *external.GraphDB, worker *heuristic.Worker, urlPath string) (reply clusterLookupReply) {
	if !worker.IsReady() {
		reply.Msg = "Worker is not ready to receive cluster lookups. Please try again later."
		reply.Warning = true
		return
	}

	addressHash := urlPath[len(constants.GetRouteClusterLookup()):]

	uid, err := address.GetAddressUid(dgraph, addressHash)
	if err != nil {
		reply.Msg = "Address hash not found"
		if !errors.Is(err, address.ErrorAddressNotFound) {
			info(cliutil.ShowCallInfo(), err)
		}
		return
	}

	addressCluster, err := worker.GetCluster(uid)
	if err != nil {
		reply.Msg = "Lookup not successful"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	addressHashes, err := address.GetAddressesByUid(dgraph, addressCluster)
	if err != nil {
		reply.Msg = "Lookup not successful"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	reply.Addresses = addressHashes
	reply.Success = true

	return
}
