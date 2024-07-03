package userserver

import (
	"backend/server"
	"net/http"
)

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
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reply, status := getCreateUserReply(s.db)

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
//	@Failure	401	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/users/{uid} [delete]
func (s *Server) handlerDeleteUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteUserReply(r, s.db)

		server.SendReply(w, reply, status)
	})
}
