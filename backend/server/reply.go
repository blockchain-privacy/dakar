package server

import (
	"backend/cmd/cliutil"
	dbus "backend/db/user"
	"backend/user"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v2"
	"io"
)

// getLoginReply reads the data from body and constructs a userReply
func getLoginReply(dgraph *dgo.Dgraph, body io.Reader) (reply userReply, userId string) {
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
			serverInfo(cliutil.ShowCallInfo(), getErr)
		}
		return
	}

	passwordValid, pwErr := user.ComparePassword(loginData.Password, u.PasswordHash)
	if pwErr != nil || !passwordValid {
		reply.Msg = invalidUserData
		return
	}

	userId = u.Uid
	reply.Success = true

	return
}

// getCreateUserReply reads the data from body and constructs a userReply
func getCreateUserReply(dgraph *dgo.Dgraph, body io.Reader) (reply userReply) {
	var frontEndUser dbus.FrontendUserCreate

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
			serverInfo(cliutil.ShowCallInfo(), err)
		}
		return
	}

	serverInfo("Generated password(", u.Email, "):", pw)
	reply.Success = true

	return
}
