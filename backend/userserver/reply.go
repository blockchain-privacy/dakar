package userserver

import (
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	dbus "backend/db/user"
	dbwork "backend/db/workspace"
	"backend/external"
	"net/http"
)

// getCreateUserReply reads the data from body and constructs a identityReply
func getCreateUserReply(dgraph external.Database) (reply createUserReply, status int) {
	// create dgraph user
	newUserUID, err := dbus.CreateNewUser(dgraph)
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

	if err := attribution.DeleteAllAttributions(dgraph, uid); err != nil {
		reply.Msg = "could not delete users' " + uid + " attributions"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := clustering.DeleteAllClusters(dgraph, uid); err != nil {
		reply.Msg = "could not delete users' " + uid + " clusters"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := dbwork.DeleteAllWorkspaces(dgraph, uid); err != nil {
		reply.Msg = "could not delete users' " + uid + " workspaces"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	if err := dbus.DeleteUser(dgraph, uid); err != nil {
		reply.Msg = "could not delete dgraph user"
		status = http.StatusInternalServerError
		warn(err)
		return
	}

	return
}
