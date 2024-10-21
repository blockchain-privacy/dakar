package userserver

import (
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"backend/db/analytics/exclusion"
	dbus "backend/db/user"
	dbwork "backend/db/workspace"
	"backend/external"
	"backend/server"
	"errors"
	"net/http"
)

type createUserReply struct {
	DakarUserUID string `json:"dakarUserUID"`
}

type msgReply struct {
	Msg string `json:"msg"`
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
	// continue even if the request gets cancelled or times out
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	if err := exclusion.DeleteAllAddressExclusions(ctx, dgraph, uid); err != nil {
		reply.Msg = "could not delete users' " + uid + " address exclusions"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := attribution.DeleteAllAttributions(ctx, dgraph, uid); err != nil {
		reply.Msg = "could not delete users' " + uid + " attributions"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := clustering.DeleteAllClusters(ctx, dgraph, uid); err != nil {
		reply.Msg = "could not delete users' " + uid + " clusters"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := dbwork.DeleteAllWorkspaces(ctx, dgraph, uid); err != nil {
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
