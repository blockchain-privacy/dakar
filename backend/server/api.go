// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"net/http"
	"time"

	mw "gitlab.com/blockchain-privacy/gomisc/middleware"
)

// Search godoc
//
//	@Summary		Search for blocks, addresses and transactions
//	@Description	Search for blocks, addresses and transactions. Supports searching by hash or block ID.
//	@Tags			data
//	@Produce		json
//	@Param			query	path		string	true	"Query string (block height, transaction hash, address hash or block hash)"
//	@Success		200		{object}	server.searchReply
//	@Failure		400		{object}	server.searchReply
//	@Failure		404		{object}	server.searchReply
//	@Failure		500		{object}	server.searchReply
//	@Router			/blockchain/search/{query} [get]
func (s *Server) handlerSearch() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getSearchReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Query for an Address godoc
//
//	@Summary	Query for an address
//	@Tags		data
//	@Produce	json
//	@Param		hash	path		string	true	"Hash"
//	@Success	200		{object}	server.addressReply
//	@Failure	400		{object}	server.addressReply
//	@Failure	404		{object}	server.addressReply
//	@Failure	500		{object}	server.addressReply
//	@Router		/blockchain/addresses/{hash} [get]
func (s *Server) handlerAddress() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddressReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Query for a Block godoc
//
//	@Summary	Query for a block
//	@Tags		data
//	@Produce	json
//	@Param		hash	path		string	true	"Hash"
//	@Param		offset	query		int		false	"transaction offset"
//	@Success	200		{object}	server.blockReply
//	@Failure	400		{object}	server.blockReply
//	@Failure	404		{object}	server.blockReply
//	@Failure	500		{object}	server.blockReply
//	@Router		/blockchain/blocks/{hash} [get]
func (s *Server) handlerBlock() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getBlockReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Query for a Transaction godoc
//
//	@Summary	Query for a transaction
//	@Tags		data
//	@Produce	json
//	@Param		hash	path		string	true	"Hash"
//	@Success	200		{object}	server.transactionReply
//	@Failure	400		{object}	server.transactionReply
//	@Failure	404		{object}	server.transactionReply
//	@Failure	500		{object}	server.transactionReply
//	@Router		/blockchain/transactions/{hash} [get]
func (s *Server) handlerTransaction() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getTransactionReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Address Output Range godoc
//
//	@Summary	Get outputs of the given address
//	@Tags		data
//	@Produce	json
//	@Param		hash	path		string										true	"Address hash"
//	@Param		options	body		server.getAddressOutputRangeReply.request	true	"query options"
//	@Success	200		{object}	server.addressReply
//	@Failure	400		{object}	server.addressReply
//	@Failure	404		{object}	server.addressReply
//	@Failure	500		{object}	server.addressReply
//	@Router		/blockchain/outputs/{hash} [post]
func (s *Server) handlerAddressOutputRange() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddressOutputRangeReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Server Status godoc
//
//	@Summary	Get the status of all backend modules
//	@Tags		meta
//	@Produce	json
//	@Success	200	{object}	server.metaReply
//	@Failure	500	{object}	server.metaReply
//	@Router		/meta/ [get]
func (s *Server) handlerMeta() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getMetaReply(s.db, s.client, r)

		SendReply(w, reply, status)
	})
}

// Selector Report godoc
//
//	@Summary	Get a CSV file containing results of the specified selector
//	@Tags		workspace
//	@Produce	text/csv
//	@Param		selector	body		server.writeSelectorReport.request	true	"selector request"
//	@Success	200			{file}		file								"comma separated values"
//	@Failure	400			{string}	string								"bad request"
//	@Failure	404			{string}	string								"resource not found"
//	@Failure	500			{string}	string								"encoding error"
//	@Router		/workspaces/selector/report/ [post]
func (s *Server) handlerSelectorReport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeSelectorReport(s.db, w, r)
	})
}

// Cluster Report godoc
//
//	@Summary	Get a CSV file containing all clusters for the given address
//	@Tags		cluster
//	@Produce	text/csv
//	@Accept		json
//	@Param		hash	path		string	true	"Address hash"
//	@Success	200		{file}		file	"comma separated values"
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/clusters/report/{hash} [get]
func (s *Server) handlerClusterReport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeClusterReport(s.db, w, r)
	})
}

// @Summary	Add a custom cluster
// @Tags		cluster
// @Produce	json
// @Param		separator	formData	string	true	"separator of the CSV file; only comma and semicolon are allowed."
// @Param		hasHeader	formData	bool	true	"controls whether the first line should be skipped"
// @Param		file		formData	file	true	"the CSV file"
// @Success	200			{object}	server.msgReply
// @Failure	400			{object}	server.msgReply
// @Failure	401			{object}	server.msgReply
// @Failure	500			{object}	server.msgReply
// @Router		/clusters/ [post]
func (s *Server) handlerAddCluster() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddClusterReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Delete a Custom Cluster godoc
//
//	@Summary	Delete a custom cluster
//	@Tags		cluster
//	@Produce	json
//	@Param		uid	path		string	true	"Cluster UID"
//	@Success	200	{object}	server.msgReply
//	@Failure	400	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/clusters/{uid} [delete]
func (s *Server) handlerDeleteCluster() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteClusterReply(r, s.db)

		SendReply(w, reply, status)
	})
}

// Delete All Custom Clusters godoc
//
//	@Summary	Delete all custom clusters of the current user
//	@Tags		cluster
//	@Produce	json
//	@Success	200	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/clusters/ [delete]
func (s *Server) handlerDeleteAllClusters() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteAllClustersReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Cluster Overview godoc
//
//	@Summary	Get all custom clusters of the current user
//	@Tags		cluster
//	@Produce	json
//	@Success	200	{object}	server.clusterOverviewReply
//	@Failure	401	{object}	server.clusterOverviewReply
//	@Failure	500	{object}	server.clusterOverviewReply
//	@Router		/clusters/ [get]
func (s *Server) handlerClusterOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getClusterOverviewReply(r, s.db)

		SendReply(w, reply, status)
	})
}

// Attribution Overview godoc
//
//	@Summary	Get all attributions of the current user
//	@Tags		attribution
//	@Produce	json
//	@Success	200	{object}	server.attributionOverviewReply
//	@Failure	401	{object}	server.attributionOverviewReply
//	@Failure	500	{object}	server.attributionOverviewReply
//	@Router		/attributions/ [get]
func (s *Server) handlerAttributionList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAttributionOverviewReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Add Private Attribution godoc
//
//	@Summary	Create a new attribution which is only visible for the current user
//	@Tags		attribution
//	@Produce	text/csv
//	@Param		separator	formData	string	true	"separator of the CSV file"
//	@Param		hasHeader	formData	bool	true	"controls whether the first line should be skipped"
//	@Param		file		formData	file	true	"the CSV file"
//	@Success	200			{object}	server.msgReply
//	@Failure	400			{object}	server.msgReply
//	@Failure	401			{object}	server.msgReply
//	@Failure	500			{object}	server.msgReply
//	@Router		/attributions/ [post]
func (s *Server) handlerAddPrivateAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddAttributionReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Delete Private Attribution godoc
//
//	@Summary	Delete an attribution belonging to the current user
//	@Tags		attribution
//	@Produce	json
//	@Param		uid	path		string	true	"Attribution UID"
//	@Success	200	{object}	server.msgReply
//	@Failure	400	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/attributions/{uid} [delete]
func (s *Server) handlerDeletePrivateAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteAttributionReply(r, s.db, false)

		SendReply(w, reply, status)
	})
}

// Delete Public Attribution godoc
//
//	@Summary	Delete a public attribution
//	@Tags		attribution
//	@Produce	json
//	@Param		uid	path		string	true	"Attribution UID"
//	@Success	200	{object}	server.msgReply
//	@Failure	400	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/attributions/public/{uid} [delete]
func (s *Server) handlerDeletePublicAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteAttributionReply(r, s.db, true)

		SendReply(w, reply, status)
	})
}

// Delete All Private Attributions godoc
//
//	@Summary	Delete all attributions of the current user
//	@Tags		attribution
//	@Produce	json
//	@Success	200	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/attributions/ [delete]
func (s *Server) handlerDeleteAllPrivateAttributions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteAllAttributionsReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Search Attributions godoc
//
//	@Summary	Search all public attributions and the attributions belonging to the current user
//	@Tags		attribution
//	@Produce	json
//	@Param		query	path		string	true	"Attribution query"
//	@Success	200		{object}	server.attributionOverviewReply
//	@Failure	400		{object}	server.attributionOverviewReply
//	@Failure	401		{object}	server.attributionOverviewReply
//	@Failure	500		{object}	server.attributionOverviewReply
//	@Router		/attributions/search/{query} [get]
func (s *Server) handlerSearchAttributions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAttributionSearchReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Selector Status godoc
//
//	@Summary		Checks if the given selector is finished executing
//	@Description	Checks if the given selector is finished executing. Returns the updated workspace in that case.
//	@Tags			workspace
//	@Produce		json
//	@Param			selector	body		server.getSelectorStatus.request	true	"selector identifier"
//	@Success		200			{object}	server.selectorStatusReply
//	@Failure		400			{object}	server.selectorStatusReply
//	@Failure		401			{object}	server.selectorStatusReply
//	@Failure		500			{object}	server.selectorStatusReply
//	@Router			/workspaces/selector/status/ [post]
func (s *Server) handlerSelectorByUID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getSelectorStatus(s.workspaceMutex, s.db, r)

		SendReply(w, reply, status)
	})
}

// HMI clusters godoc
//
//	@Summary	Get HMI clusters per address
//	@Tags		cluster
//	@Produce	json
//	@Param		hash	path		string	true	"Transaction hash"
//	@Success	200		{object}	server.hmiLookupReply
//	@Failure	400		{object}	server.hmiLookupReply
//	@Failure	500		{object}	server.hmiLookupReply
//	@Router		/clusters/hmi/{hash} [get]
func (s *Server) handlerHMILookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getHMILookupReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Selector Results godoc
//
//	@Summary		Get the results of a selector.
//	@Description	Get the results of a selector.
//	@Description	Depending on the selector type, one of the result properties is set.
//	@Tags			workspace
//	@Produce		json
//	@Accept			json
//	@Param			selector	body		server.getSelectorResultsReply.request	true	"Selector request"
//	@Success		200			{object}	server.selectorResultsReply
//	@Failure		400			{object}	server.selectorResultsReply
//	@Failure		401			{object}	server.selectorResultsReply
//	@Failure		500			{object}	server.selectorResultsReply
//	@Router			/workspaces/selector/results/ [post]
func (s *Server) handlerSelectorDetails() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getSelectorResultsReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Add Selector godoc
//
//	@Summary		Adds a new selector to a workspace
//	@Description	Adds a new selector to a workspace. The selector will be executed eventually
//	@Description	and the results stored in the database. Only one options parameter may be set.
//	@Description	Returns the updated workspace.
//	@Tags			workspace
//	@Produce		json
//	@Accept			json
//	@Param			selector	body		server.getAddWorkspaceSelectorReply.request	true	"Selector properties"
//	@Success		200			{object}	server.addWorkspaceSelectorReply
//	@Failure		400			{object}	server.addWorkspaceSelectorReply
//	@Failure		401			{object}	server.addWorkspaceSelectorReply
//	@Failure		500			{object}	server.addWorkspaceSelectorReply
//	@Router			/workspaces/selector/ [post]
func (s *Server) handlerAddWorkspaceSelector() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddWorkspaceSelectorReply(s.db, r, s.workspaceMutex, s.worker)

		SendReply(w, reply, status)
	})
}

// Shortest Transaction Path godoc
//
//	@Summary	Get the shortest path between two transactions
//	@Tags		tools
//	@Produce	json
//	@Accept		json
//	@Param		transactions	body		server.getShortestTransactionPathReply.request	true	"transactions between which the path should be found"
//	@Success	200				{object}	server.shortestTransactionPathReply
//	@Failure	400				{object}	server.shortestTransactionPathReply
//	@Failure	500				{object}	server.shortestTransactionPathReply
//	@Router		/shortestTransactionPath/ [post]
func (s *Server) handlerShortestTransactionPath() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getShortestTransactionPathReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Cluster Lookup Path godoc
//
//	@Summary	Get all clusters of the given address
//	@Tags		cluster
//	@Produce	json
//	@Param		hash	path		string	true	"Address hash"
//	@Success	200		{object}	server.clusterLookupReply
//	@Failure	400		{object}	server.clusterLookupReply
//	@Failure	401		{object}	server.clusterLookupReply
//	@Failure	500		{object}	server.clusterLookupReply
//	@Router		/clusters/{hash} [get]
func (s *Server) handlerClusterLookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getClusterLookupReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Mixing Activity Path godoc
//
//	@Summary	Mixing Activity of an address
//	@Tags		tools
//	@Produce	json
//	@Accept		json
//	@Param		activity	body		server.getMixingActivity.request	true	"Mixing activity request details"
//	@Success	200			{object}	server.mixingActivityReply
//	@Failure	400			{object}	server.mixingActivityReply
//	@Failure	500			{object}	server.mixingActivityReply
//	@Router		/mixingActivity/ [post]
func (s *Server) handlerMixingActivity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getMixingActivity(s.db, r)

		SendReply(w, reply, status)
	})
}

// Spending Fingerprint godoc
//
//	@Summary	Get the spending fingerprint of a transaction
//	@Tags		tools
//	@Produce	json
//	@Param		hash	path		string	true	"Transaction hash"
//	@Success	200		{object}	server.spendingFingerprintReply
//	@Failure	400		{object}	server.spendingFingerprintReply
//	@Failure	500		{object}	server.spendingFingerprintReply
//	@Router		/spendingFingerprint/{hash} [get]
func (s *Server) handlerSpendingFingerprint() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getSpendingFingerprintReply(s.db, s.graphWrapper, r)

		SendReply(w, reply, status)
	})
}

// Add Nodes godoc
//
//	@Summary	Add nodes to a workspace. This will also refresh connections between existing nodes.
//	@Tags		workspace
//	@Accept		json
//	@Produce	json
//	@Param		query	body		server.getAddWorkspaceNodesReply.request	true	"Search query"
//	@Success	200		{object}	server.addWorkspaceNodesReply
//	@Failure	400		{object}	server.addWorkspaceNodesReply
//	@Failure	404		{object}	server.addWorkspaceNodesReply
//	@Failure	500		{object}	server.addWorkspaceNodesReply
//	@Router		/workspaces/nodes/ [post]
func (s *Server) handlerAddWorkspaceNodes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddWorkspaceNodesReply(s.db, s.workspaceMutex, r)

		SendReply(w, reply, status)
	})
}

// Add Note godoc
//
//	@Summary		Add a note or update a note
//	@Description	Add a new note (empty uid) to a workspace or update an existing one. 100-character limit for the note text.
//	@Tags			workspace
//	@Accept			json
//	@Produce		json
//	@Param			note	body		server.getAddWorkspaceNoteReply.request	true	"New note"
//	@Success		200		{object}	server.addWorkspaceNoteReply
//	@Failure		400		{object}	server.addWorkspaceNoteReply
//	@Failure		500		{object}	server.addWorkspaceNoteReply
//	@Router			/workspaces/note/ [post]
func (s *Server) handlerAddWorkspaceNote() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddWorkspaceNoteReply(s.db, s.workspaceMutex, r)

		SendReply(w, reply, status)
	})
}

// List all Workspaces godoc
//
//	@Summary	List all workspaces of the current user
//	@Tags		workspace
//	@Produce	json
//	@Success	200	{object}	server.workspacesReply
//	@Failure	400	{object}	server.workspacesReply
//	@Failure	500	{object}	server.workspacesReply
//	@Router		/workspaces/ [get]
func (s *Server) handlerWorkspaces() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getWorkspacesReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Add Workspace godoc
//
//	@Summary	Creates a new workspace
//	@Tags		workspace
//	@Produce	json
//	@Param		name	path		string	true	"Workspace name"
//	@Success	200		{string}	string
//	@Failure	400		{string}	string
//	@Failure	500		{string}	string
//	@Router		/workspaces/{name} [post]
func (s *Server) handlerAddWorkspace() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := getAddWorkspaceReply(s.db, r)

		SendReply(w, "", status)
	})
}

// Rename Workspace godoc
//
//	@Summary	Renames a workspace
//	@Tags		workspace
//	@Produce	json
//	@Param		workspace	body		server.getRenameWorkspaceReply.request	true	"Workspace"
//	@Success	200			{string}	string
//	@Failure	400			{string}	string
//	@Failure	500			{string}	string
//	@Router		/workspaces/rename/ [post]
func (s *Server) handlerRenameWorkspace() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := getRenameWorkspaceReply(s.db, r)

		SendReply(w, "", status)
	})
}

// Get Workspace godoc
//
//	@Summary	Returns the specified workspace
//	@Tags		workspace
//	@Produce	json
//	@Param		uid	path		string	true	"Workspace UID"
//	@Success	200	{object}	server.getWorkspaceReply
//	@Failure	400	{object}	server.getWorkspaceReply
//	@Failure	500	{object}	server.getWorkspaceReply
//	@Router		/workspaces/{uid} [get]
func (s *Server) handlerGetWorkspace() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getGetWorkspaceReply(s.db, s.workspaceMutex, r)

		SendReply(w, reply, status)
	})
}

// Get Workspace State godoc
//
//	@Summary	Returns the state of the specified workspace. Calling this endpoint performs no workspace update checks.
//	@Tags		workspace
//	@Produce	json
//	@Param		uid	path		string	true	"Workspace UID"
//	@Success	200	{object}	server.getWorkspaceStateReply
//	@Failure	400	{object}	server.getWorkspaceStateReply
//	@Failure	500	{object}	server.getWorkspaceStateReply
//	@Router		/workspaces/state/{uid} [get]
func (s *Server) handlerGetWorkspaceState() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getGetWorkspaceStateReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Update Workspace godoc
//
//	@Summary	Update a workspace
//	@Tags		workspace
//	@Accept		json
//	@Produce	json
//	@Param		state	body		server.getUpdateWorkspace.request	true	"Workspace state"
//	@Success	200		{string}	string
//	@Failure	400		{string}	string
//	@Failure	500		{string}	string
//	@Router		/workspaces/ [put]
func (s *Server) handlerUpdateWorkspace() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := getUpdateWorkspace(s.db, s.workspaceMutex, r)

		SendReply(w, "", status)
	})
}

// Delete a Workspace Node godoc
//
//	@Summary	Deletes a workspace node
//	@Tags		workspace
//	@Produce	json
//	@Param		state	body		server.getDeleteWorkspaceNodeReply.request	true	"Node which shall be deleted"
//	@Success	200		{object}	server.deleteWorkspaceNodeReply
//	@Failure	400		{object}	server.deleteWorkspaceNodeReply
//	@Failure	500		{object}	server.deleteWorkspaceNodeReply
//	@Router		/workspaces/node/ [delete]
func (s *Server) handlerDeleteWorkspaceNode() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteWorkspaceNodeReply(s.db, s.workspaceMutex, r)

		SendReply(w, reply, status)
	})
}

// Get the connection between two nodes godoc
//
//	@Summary	Get the connection between two nodes
//	@Tags		workspace
//	@Produce	json
//	@Param		state	body		server.getWorkspaceConnectionReply.request	true	"Node which shall be deleted"
//	@Success	200		{object}	server.workspaceConnectionReply
//	@Failure	400		{object}	server.workspaceConnectionReply
//	@Failure	500		{object}	server.workspaceConnectionReply
//	@Router		/workspaces/connection/ [post]
func (s *Server) handlerWorkspaceConnection() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getWorkspaceConnectionReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// Delete a Workspace godoc
//
//	@Summary	Deletes a workspace of the current user
//	@Tags		workspace
//	@Produce	json
//	@Param		uid	path		string	true	"Workspace UID"
//	@Success	200	{object}	server.msgReply
//	@Failure	400	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	404	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/workspaces/{uid} [delete]
func (s *Server) handlerDeleteWorkspace() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteWorkspaceReply(s.db, r)

		SendReply(w, reply, status)
	})
}

// setupHandlers creates endpoint handlers
func (s *Server) setupHandlers() {
	// Search
	s.handler.Handle(BuildPattern(http.MethodGet, routeSearch, "query"),
		s.adapt(s.handlerSearch(), s.cacheFactory(time.Minute*10), mw.MaxBody5MiB()))

	// Common data
	s.handler.Handle(BuildPattern(http.MethodGet, routeTransaction, "hash"),
		// transaction data never changes so set cache time to be infinite
		s.adapt(s.handlerTransaction(), mw.MaxBody5MiB(), s.cacheFactory(time.Duration(0))))
	s.handler.Handle(BuildPattern(http.MethodGet, routeBlock, "hash"),
		s.adapt(s.handlerBlock(), mw.MaxBody5MiB(), s.cacheFactory(time.Minute*10)))
	s.handler.Handle(BuildPattern(http.MethodGet, routeAddress, "hash"),
		s.adapt(s.handlerAddress(), mw.MaxBody5MiB(), s.cacheFactory(time.Minute*10)))
	s.handler.Handle(BuildPattern(http.MethodPost, routeAddressOutputRange, "hash"),
		s.adapt(s.handlerAddressOutputRange(), mw.MaxBody5MiB(), s.cacheFactory(time.Minute*10)))

	// Meta
	s.handler.Handle(BuildPattern(http.MethodGet, routeMeta, ""),
		s.adapt(s.handlerMeta(), Authorization(), mw.MaxBody5MiB(), s.cacheFactory(time.Second*10)))

	// Analytics
	s.handler.Handle(BuildPattern(http.MethodPost, routeShortestTxPath, ""),
		s.adapt(s.handlerShortestTransactionPath(), Authorization(), mw.MaxBody5MiB(), s.cacheFactory(time.Minute*10)))
	s.handler.Handle(BuildPattern(http.MethodPost, routeMixingActivity, ""),
		s.adapt(s.handlerMixingActivity(), Authorization(), mw.MaxBody5MiB(), s.cacheFactory(time.Minute*10)))
	s.handler.Handle(BuildPattern(http.MethodGet, routeSpendingFingerprint, "hash"),
		s.adapt(s.handlerSpendingFingerprint(), Authorization(), mw.MaxBody5MiB(), s.cacheFactory(time.Minute*10)))

	// Clusters
	s.handler.Handle(BuildPattern(http.MethodGet, routeClusters, "hash"),
		s.adapt(s.handlerClusterLookup(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodGet, routeClustersHmi, "hash"),
		s.adapt(s.handlerHMILookup(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodGet, routeClustersReport, "hash"),
		s.adapt(s.handlerClusterReport(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodPost, routeClusters, ""),
		s.adapt(s.handlerAddCluster(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodDelete, routeClusters, "uid"),
		s.adapt(s.handlerDeleteCluster(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodDelete, routeClusters, ""),
		s.adapt(s.handlerDeleteAllClusters(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodGet, routeClusters, ""),
		s.adapt(s.handlerClusterOverview(), Authorization(), mw.MaxBody5MiB()))

	// Attributions
	s.handler.Handle(BuildPattern(http.MethodPost, routeAttributions, ""),
		s.adapt(s.handlerAddPrivateAttribution(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodGet, routeAttributions, ""),
		s.adapt(s.handlerAttributionList(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodDelete, routeAttributions, "uid"),
		s.adapt(s.handlerDeletePrivateAttribution(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodDelete, routeAttributionsPublic, "uid"),
		s.adapt(s.handlerDeletePublicAttribution(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodDelete, routeAttributions, ""),
		s.adapt(s.handlerDeleteAllPrivateAttributions(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodGet, routeAttributionsSearch, "query"),
		s.adapt(s.handlerSearchAttributions(), Authorization(), mw.MaxBody5MiB()))

	// Workspace
	s.handler.Handle(BuildPattern(http.MethodPost, routeWorkspaceRename, ""),
		s.adapt(s.handlerRenameWorkspace(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodPost, routeWorkspacesNodes, ""),
		s.adapt(s.handlerAddWorkspaceNodes(), Authorization(), mw.MaxBody(50)))
	s.handler.Handle(BuildPattern(http.MethodPost, routeAddWorkspaceNote, ""),
		s.adapt(s.handlerAddWorkspaceNote(), Authorization(), mw.MaxBody(50)))
	s.handler.Handle(BuildPattern(http.MethodPost, routeAddWorkspaceSelector, ""),
		s.adapt(s.handlerAddWorkspaceSelector(), Authorization(), mw.MaxBody(50)))
	s.handler.Handle(BuildPattern(http.MethodPost, routeWorkspaceSelectorStatus, ""),
		s.adapt(s.handlerSelectorByUID(), Authorization(), mw.MaxBody(50)))
	s.handler.Handle(BuildPattern(http.MethodDelete, routeWorkspacesNode, ""),
		s.adapt(s.handlerDeleteWorkspaceNode(), Authorization(), mw.MaxBody(50)))
	s.handler.Handle(BuildPattern(http.MethodGet, routeWorkspaces, ""),
		s.adapt(s.handlerWorkspaces(), Authorization()))
	s.handler.Handle(BuildPattern(http.MethodPost, routeWorkspaces, "name"),
		s.adapt(s.handlerAddWorkspace(), Authorization()))
	s.handler.Handle(BuildPattern(http.MethodGet, routeWorkspaces, "uid"),
		s.adapt(s.handlerGetWorkspace(), Authorization()))
	s.handler.Handle(BuildPattern(http.MethodGet, routeWorkspacesState, "uid"),
		s.adapt(s.handlerGetWorkspaceState(), Authorization()))
	s.handler.Handle(BuildPattern(http.MethodPut, routeWorkspaces, ""),
		s.adapt(s.handlerUpdateWorkspace(), Authorization(), mw.MaxBody(50)))
	s.handler.Handle(BuildPattern(http.MethodDelete, routeWorkspaces, "uid"),
		s.adapt(s.handlerDeleteWorkspace(), Authorization(), mw.MaxBody5MiB()))
	s.handler.Handle(BuildPattern(http.MethodPost, routeWorkspacesConnection, ""),
		s.adapt(s.handlerWorkspaceConnection(), Authorization(), mw.MaxBody5MiB(), s.cacheFactory(0)))
	s.handler.Handle(BuildPattern(http.MethodPost, routeWorkspaceSelectorResults, ""),
		s.adapt(s.handlerSelectorDetails(), Authorization(), mw.MaxBody5MiB(), s.cacheFactory(0)))
	s.handler.Handle(BuildPattern(http.MethodPost, routeWorkspaceSelectorReport, ""),
		s.adapt(s.handlerSelectorReport(), Authorization(), mw.MaxBody5MiB(), s.cacheFactory(0)))
}
