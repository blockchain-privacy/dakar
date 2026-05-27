// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package userserver

import (
	"errors"
	"net/http"

	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics/clustering"
	dbstat "gitlab.com/blockchain-privacy/dakar/db/status"
	dbus "gitlab.com/blockchain-privacy/dakar/db/user"
	dbwork "gitlab.com/blockchain-privacy/dakar/db/workspace"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/dakar/server"
)

type healthCheckReply struct {
	Healthy bool `json:"healthy"`
}

type createUserReply struct {
	DakarUserUID string `json:"dakarUserUID"`
}

type msgReply struct {
	Msg string `json:"msg"`
}

// getCreateUserReply reads the data from body and constructs a identityReply
func getHealthCheckReply(r *http.Request, dgraph external.Database) (reply healthCheckReply, status int) {
	// do a simple status check to determine if the database connection is healthy
	_, err := dbstat.GetFrontendStatus(r.Context(), dgraph)
	reply.Healthy = err == nil

	return
}

// getCreateUserReply reads the data from body and constructs a identityReply
func getCreateUserReply(r *http.Request, dgraph external.Database) (reply createUserReply, status int) {
	// create dgraph user
	newUserUID, err := dbus.CreateNewUser(r.Context(), dgraph)
	if err != nil {
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	reply.DakarUserUID = newUserUID

	return
}

// getDeleteUserReply deletes the given user
func getDeleteUserReply(r *http.Request, dgraph external.Database) (reply msgReply, status int) {
	uid := r.PathValue("uid")
	if uid == "" {
		status = http.StatusBadRequest
		return
	}

	// not using the request context here, because user deletion process should
	// continue even if the request gets canceled or times out
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	if err := clustering.DeleteAllClusters(ctx, dgraph, uid); err != nil {
		reply.Msg = "could not delete users' " + uid + " clusters"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	// deletes all workspaces, if no workspace UID is set
	if err := dbwork.DeleteWorkspace(ctx, dgraph, uid, ""); err != nil {
		reply.Msg = "could not delete users' " + uid + " workspaces"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := dbus.DeleteUser(ctx, dgraph, uid); err != nil {
		if errors.Is(err, dbus.ErrUserDoesNotExist) {
			status = http.StatusNotFound
			return
		}

		reply.Msg = "could not delete dgraph user"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}

// Create Identity godoc
//
//	@Summary	Create a new user.
//	@Tags		user
//	@Produce	json
//	@Accept		json
//	@Success	200			{object}	server.createUserReply
//	@Failure	400			{object}	server.createUserReply
//	@Failure	500			{object}	server.createUserReply
//	@Router		/users/ [post]
func (s *Server) handlerCreateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getCreateUserReply(r, s.db)

		server.SendReply(w, reply, status)
	})
}

// Delete Arbitrary Identity godoc
//
//	@Summary	Delete an arbitrary user.
//	@Tags		user
//	@Produce	json
//	@Param		uid	path		string	true	"Identity UID"
//	@Success	200	{object}	server.msgReply
//	@Failure	400	{object}	server.msgReply
//	@Failure	404	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/users/{uid} [delete]
func (s *Server) handlerDeleteUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteUserReply(r, s.db)

		server.SendReply(w, reply, status)
	})
}

// Health Check godoc
//
//	@Summary	Check the health of the server. This includes a database connection check.
//	@Tags		user
//	@Produce	json
//	@Success	200	{object}	server.healthCheckReply
//	@Failure	500	{object}	server.healthCheckReply
//	@Router		/health/ [get]
func (s *Server) handlerHealth() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getHealthCheckReply(r, s.db)

		server.SendReply(w, reply, status)
	})
}
