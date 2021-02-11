package server

import (
	"backend/cmd/cliutil"
	"backend/db/analytics/heuristics/transaction"
	dbus "backend/db/user"
	"backend/user"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v2"
	"io"
)

// getLoginReply reads the data from body and constructs a userReply
func getLoginReply(dgraph *dgo.Dgraph, body io.Reader) (reply backendUserReply) {
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
func getCreateUserReply(dgraph *dgo.Dgraph, body io.Reader) (reply userReply) {
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

// getModifyUserReply parses the input and creates a corresponding userReply
func getModifyUserReply(dgraph *dgo.Dgraph, body io.Reader, tUser tokenUser) (reply backendUserReply) {
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
func getDeleteUserReply(dgraph *dgo.Dgraph, delUid string, tUser tokenUser) (reply userReply) {
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
func getShortestTransactionPathReply(dgraph *dgo.Dgraph, body io.Reader) (reply shortestTransactionPathReply) {
	// parse request
	var sPathRequest transaction.ShortestTransactionPathRequest
	if err := json.NewDecoder(body).Decode(&sPathRequest); err != nil {
		reply.Msg = "could not decode request data"
		return
	}

	// do shortest path lookup
	txs, err := transaction.GetShortestTransactionPathAnyDirection(dgraph, sPathRequest.From, sPathRequest.To,
		sPathRequest.IncludePrivacyTransactions)
	if err != nil {
		reply.Msg = "error while searching for paths"
		info(cliutil.ShowCallInfo(), err)
		return
	}

	info(txs)

	reply.Transactions = txs
	reply.Success = true

	return
}
