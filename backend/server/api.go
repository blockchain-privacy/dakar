package server

import (
	"net/http"
	"time"
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
		reply, status := getSearchReply(s.db, r.PathValue("query"))

		sendReply(w, reply, status)
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
		reply, status := getAddressReply(s.db, r.PathValue("hash"))

		sendReply(w, reply, status)
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
		reply, status := getBlockReply(r, s.db, r.PathValue("hash"))

		sendReply(w, reply, status)
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
		reply, status := getTransactionReply(s.db, r.PathValue("hash"))

		sendReply(w, reply, status)
	})
}

// Address Output Range godoc
//
//	@Summary	Get outputs of the given address
//	@Tags		data
//	@Produce	json
//	@Param		hash	path		string										true	"Address hash"
//	@Param		options	body		server.getAddressOutputRangeReply.request	true	"query options"
//	@Success	200		{object}	server.searchReply
//	@Failure	400		{object}	server.searchReply
//	@Failure	404		{object}	server.searchReply
//	@Failure	500		{object}	server.searchReply
//	@Router		/blockchain/outputs/{hash} [post]
func (s *Server) handlerAddressOutputRange() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddressOutputRangeReply(r, s.db, r.PathValue("hash"))

		sendReply(w, reply, status)
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
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reply, status := getMetaReply(s.db, s.client)

		sendReply(w, reply, status)
	})
}

// Heuristic Report godoc
//
//	@Summary	Get a CSV file containing results of the specified heuristic
//	@Tags		heuristic
//	@Produce	text/csv
//	@Param		work	body		server.writeHeuristicReport.request	true	"work item"
//	@Success	200		{file}		file								"comma separated values"
//	@Failure	400		{string}	string								"bad request"
//	@Failure	404		{string}	string								"resource not found"
//	@Failure	500		{string}	string								"encoding error"
//	@Router		/heuristics/report/ [post]
func (s *Server) handlerHeuristicsReport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeHeuristicReport(w, r, s.db)
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
		writeClusterReport(w, r, s.db)
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

		sendReply(w, reply, status)
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

		sendReply(w, reply, status)
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
		reply, status := getDeleteAllClustersReply(r, s.db)

		sendReply(w, reply, status)
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

		sendReply(w, reply, status)
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
		reply, status := getAttributionOverviewReply(r, s.db)

		sendReply(w, reply, status)
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
		reply, status := getAddAttributionReply(r, s.db, false)

		sendReply(w, reply, status)
	})
}

// Add Public Attribution godoc
//
//	@Summary	Create a new attribution which is visible to all users
//	@Tags		attribution
//	@Produce	text/csv
//	@Param		separator	formData	string	true	"separator of the CSV file"
//	@Param		hasHeader	formData	bool	true	"controls whether the first line should be skipped"
//	@Param		file		formData	file	true	"the CSV file"
//	@Success	200			{object}	server.msgReply
//	@Failure	400			{object}	server.msgReply
//	@Failure	401			{object}	server.msgReply
//	@Failure	500			{object}	server.msgReply
//	@Router		/attributions/public/ [post]
func (s *Server) handlerAddPublicAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddAttributionReply(r, s.db, true)

		sendReply(w, reply, status)
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

		sendReply(w, reply, status)
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

		sendReply(w, reply, status)
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
		reply, status := getDeleteAllAttributionsReply(r, s.db)

		sendReply(w, reply, status)
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
		reply, status := getAttributionSearchReply(r, s.db, r.PathValue("query"))

		sendReply(w, reply, status)
	})
}

// Add Address Exclusion godoc
//
//	@Summary	Adds new address exclusions for the current user
//	@Tags		address exclusions
//	@Produce	text/csv
//	@Param		file	formData	file	true	"the CSV file"
//	@Success	200		{object}	server.msgReply
//	@Failure	400		{object}	server.msgReply
//	@Failure	401		{object}	server.msgReply
//	@Failure	500		{object}	server.msgReply
//	@Router		/exclusions/ [post]
func (s *Server) handlerAddAddressExclusions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddAddressExclusionsReply(s.db, r)

		sendReply(w, reply, status)
	})
}

// Delete an Address Exclusion godoc
//
//	@Summary	Deletes an address exclusion of the current user
//	@Tags		address exclusions
//	@Produce	text/plain
//	@Param		hash	path		string	true	"Address hash"
//	@Success	200		{string}	string
//	@Failure	400		{string}	string
//	@Failure	401		{string}	string
//	@Failure	500		{string}	string
//	@Router		/exclusions/{hash} [delete]
func (s *Server) handlerDeleteAddressExclusion() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := getDeleteAddressExclusionReply(r, s.db)

		sendReply(w, "", status)
	})
}

// Delete All Address Exclusions godoc
//
//	@Summary	Delete all address exclusions of the current user
//	@Tags		address exclusions
//	@Produce	text/plain
//	@Success	200	{string}	string
//	@Failure	401	{string}	string
//	@Failure	500	{string}	string
//	@Router		/exclusions/ [delete]
func (s *Server) handlerDeleteAllAddressExclusions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := getDeleteAllAddressExclusionsReply(r, s.db)

		sendReply(w, "", status)
	})
}

// List Address Exclusions godoc
//
//	@Summary	Lists all address exclusions of the current user
//	@Tags		address exclusions
//	@Produce	json
//	@Success	200	{object}	server.addressExclusionOverviewReply
//	@Failure	401	{object}	server.addressExclusionOverviewReply
//	@Failure	500	{object}	server.addressExclusionOverviewReply
//	@Router		/exclusions/ [get]
func (s *Server) handlerAddressExclusionList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddressExclusionOverviewReply(r, s.db)

		sendReply(w, reply, status)
	})
}

// Heuristic by Work ID godoc
//
//	@Summary		Get a specific heuristic by work ID
//	@Description	Get a specific heuristic by work ID
//	@Tags			heuristic
//	@Produce		json
//	@Param			work	body		server.getHeuristicByWorkIDReply.request	true	"work item"
//	@Success		200		{object}	server.heuristicByWorkIDReply
//	@Failure		400		{object}	server.heuristicByWorkIDReply
//	@Failure		401		{object}	server.heuristicByWorkIDReply
//	@Failure		500		{object}	server.heuristicByWorkIDReply
//	@Router			/heuristicByWorkID/ [post]
func (s *Server) handlerHeuristicByWorkID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getHeuristicByWorkIDReply(r, s.db, s.worker)

		sendReply(w, reply, status)
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
		reply, status := getHMILookupReply(s.db, r.PathValue("hash"))

		sendReply(w, reply, status)
	})
}

// Heuristic Details godoc
//
//	@Summary	Get the details of a heuristic
//	@Tags		heuristic
//	@Produce	json
//	@Accept		json
//	@Param		heuristic	body		server.getHeuristicDetailsReply.request	true	"Heuristic UID"
//	@Success	200			{object}	server.heuristicDetailsReply
//	@Failure	400			{object}	server.heuristicDetailsReply
//	@Failure	401			{object}	server.heuristicDetailsReply
//	@Failure	500			{object}	server.heuristicDetailsReply
//	@Router		/heuristicDetails/ [post]
func (s *Server) handlerHeuristicsDetails() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getHeuristicDetailsReply(r, s.db)

		sendReply(w, reply, status)
	})
}

// Execute Heuristics godoc
//
//	@Summary		Queues the execution of heuristics for the given transaction
//	@Description	This call queues the given heuristics for the given transaction. Does not wait until the heuristic execution is finished.
//	@Tags			heuristic
//	@Produce		json
//	@Accept			json
//	@Param			heuristic	body		server.getHeuristicExecutionReply.request	true	"Heuristics to queue"
//	@Success		200			{object}	server.heuristicExecutionReply
//	@Failure		400			{object}	server.heuristicExecutionReply
//	@Failure		401			{object}	server.heuristicExecutionReply
//	@Failure		500			{object}	server.heuristicExecutionReply
//	@Router			/executeHeuristics/ [post]
func (s *Server) handlerHeuristicsExecution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getHeuristicExecutionReply(r, s.db, s.worker, s.workspaceMutex)

		sendReply(w, reply, status)
	})
}

// Create Identity godoc
//
//	@Summary	Create a new identity. This is an admin endpoint.
//	@Tags		identity
//	@Produce	json
//	@Accept		json
//	@Param		identity	body		server.getCreateIdentityReply.request	true	"Identity details"
//	@Success	200			{object}	server.msgReply
//	@Failure	400			{object}	server.msgReply
//	@Failure	500			{object}	server.msgReply
//	@Router		/identities/ [post]
func (s *Server) handlerCreateIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getCreateIdentityReply(s.db, s.adminAuth, r)

		sendReply(w, reply, status)
	})
}

// Admin Delete Identity godoc
//
//	@Summary	Delete an identity. This is an admin endpoint.
//	@Tags		identity
//	@Produce	json
//	@Param		uid	path		string	true	"Identity UID"
//	@Success	200	{object}	server.msgReply
//	@Failure	400	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/identities/{uid} [delete]
func (s *Server) handlerAdminDeleteIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteIdentityReply(r, s.db, s.adminAuth, true)

		sendReply(w, reply, status)
	})
}

// Delete Identity godoc
//
//	@Summary	Delete the identity of the calling user
//	@Tags		identity
//	@Produce	json
//	@Success	200	{object}	server.msgReply
//	@Failure	400	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/identities/ [delete]
func (s *Server) handlerDeleteIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteIdentityReply(r, s.db, s.adminAuth, false)

		sendReply(w, reply, status)
	})
}

// Modify Identity godoc
//
//	@Summary	Modify an arbitrary identity. This is an admin endpoint.
//	@Tags		identity
//	@Produce	json
//	@Accept		json
//	@Param		identity	body		server.getModifyIdentityReply.request	true	"Identity modification details"
//	@Success	200			{object}	server.msgReply
//	@Failure	400			{object}	server.msgReply
//	@Failure	500			{object}	server.msgReply
//	@Router		/identities/ [put]
func (s *Server) handlerModifyIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getModifyIdentityReply(s.adminAuth, r)

		sendReply(w, reply, status)
	})
}

// Get Identities godoc
//
//	@Summary	Get all identities. This is an admin endpoint.
//	@Tags		identity
//	@Produce	json
//	@Success	200	{object}	server.identitiesReply
//	@Failure	500	{object}	server.identitiesReply
//	@Router		/identities/ [get]
func (s *Server) handlerGetIdentities() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getIdentitiesReply(s.adminAuth, r)

		sendReply(w, reply, status)
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
		reply, status := getShortestTransactionPathReply(s.db, r.Body)

		sendReply(w, reply, status)
	})
}

// Connection Lookup Path godoc
//
//	@Summary	Connection lookup
//	@Tags		tools
//	@Produce	json
//	@Param		hash	path		string	true	"Transaction hash"
//	@Param		forward	query		bool	true	"search direction"
//	@Param		t		query		int		true	"time range in number of days"	maximum(90)	minimum(1)
//	@Success	200		{object}	server.connectionLookupReply
//	@Failure	400		{object}	server.connectionLookupReply
//	@Failure	404		{object}	server.connectionLookupReply
//	@Failure	500		{object}	server.connectionLookupReply
//	@Router		/connectionLookup/{hash} [get]
func (s *Server) handlerConnectionLookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getConnectionLookupReply(s.db, s.graphWrapper, r.PathValue("hash"), r.URL)

		sendReply(w, reply, status)
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
		reply, status := getClusterLookupReply(r, s.db)

		sendReply(w, reply, status)
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
		reply, status := getMixingActivity(s.db, r.Body)

		sendReply(w, reply, status)
	})
}

// Address Exclusion Status godoc
//
//	@Summary	Get the exclusion status of an address
//	@Tags		address exclusions
//	@Produce	json
//	@Param		hash	path		string	true	"Address hash"
//	@Success	200		{object}	server.addressExclusionStatusReply
//	@Failure	400		{object}	server.addressExclusionStatusReply
//	@Failure	401		{object}	server.addressExclusionStatusReply
//	@Failure	500		{object}	server.addressExclusionStatusReply
//	@Router		/exclusions/{hash} [get]
func (s *Server) handlerGetAddressExclusionStatus() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddressExclusionStatusReply(r, s.db, r.PathValue("hash"))

		sendReply(w, reply, status)
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
		reply, status := getSpendingFingerprintReply(s.db, s.graphWrapper, r.PathValue("hash"))

		sendReply(w, reply, status)
	})
}

// Add Node godoc
//
//	@Summary	Add a node to a workspace. This will also add connections between existing nodes.
//	@Tags		workspace
//	@Accept		json
//	@Produce	json
//	@Param		query	body		server.getAddWorkspaceNodeReply.request	true	"Search query"
//	@Success	200		{object}	server.addWorkspaceNodeReply
//	@Failure	400		{object}	server.addWorkspaceNodeReply
//	@Failure	500		{object}	server.addWorkspaceNodeReply
//	@Router		/workspaces/node/ [post]
func (s *Server) handlerAddWorkspaceNode() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddWorkspaceNodeReply(s.db, s.workspaceMutex, s.worker, r)

		sendReply(w, reply, status)
	})
}

// Add Note godoc
//
//	@Summary	Add a note to a workspace. 100 character limit.
//	@Tags		workspace
//	@Accept		json
//	@Produce	json
//	@Param		note	body		server.getAddWorkspaceNoteReply.request	true	"New note"
//	@Success	200		{object}	server.addWorkspaceNoteReply
//	@Failure	400		{object}	server.addWorkspaceNoteReply
//	@Failure	500		{object}	server.addWorkspaceNoteReply
//	@Router		/workspaces/note/ [post]
func (s *Server) handlerAddWorkspaceNote() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getAddWorkspaceNoteReply(s.db, s.workspaceMutex, r)

		sendReply(w, reply, status)
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

		sendReply(w, reply, status)
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

		sendReply(w, "", status)
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

		sendReply(w, "", status)
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
		reply, status := getGetWorkspaceReply(s.db, s.workspaceMutex, s.worker, r)

		sendReply(w, reply, status)
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

		sendReply(w, "", status)
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

		sendReply(w, reply, status)
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

		sendReply(w, reply, status)
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
		reply, status := getDeleteWorkspaceReply(r, s.db)

		sendReply(w, reply, status)
	})
}

// Delete All Workspace godoc
//
//	@Summary	Deletes all workspaces of the current user
//	@Tags		workspace
//	@Produce	json
//	@Success	200	{object}	server.msgReply
//	@Failure	400	{object}	server.msgReply
//	@Failure	401	{object}	server.msgReply
//	@Failure	404	{object}	server.msgReply
//	@Failure	500	{object}	server.msgReply
//	@Router		/workspaces/ [delete]
func (s *Server) handlerDeleteAllWorkspaces() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply, status := getDeleteAllWorkspacesReply(r, s.db)

		sendReply(w, reply, status)
	})
}

// setupHandlers creates endpoint handlers
func (s *Server) setupHandlers() {
	// Search
	s.handler.Handle(getRouteSearch(),
		adapt(s.handlerSearch(), getRouteSearch(), s.useCache(time.Minute*10), maxBody()))

	// Common data
	s.handler.Handle(getRouteTransaction(),
		adapt(s.handlerTransaction(), getRouteTransaction(), s.useCache(time.Second*0), maxBody()))
	// setting block cache time to 10 Minutes because blocks at
	// the tip get updated via adding the 'next block' reference
	s.handler.Handle(getRouteBlock(),
		adapt(s.handlerBlock(), getRouteBlock(), s.useCache(time.Second*10), maxBody()))
	s.handler.Handle(getRouteAddress(),
		adapt(s.handlerAddress(), getRouteAddress(), s.useCache(time.Second*10), maxBody()))
	s.handler.Handle(getRouteAddressOutputRange(),
		adapt(s.handlerAddressOutputRange(), getRouteAddressOutputRange(), s.useCache(time.Minute*10), maxBody()))

	// Meta
	s.handler.Handle(getRouteMeta(),
		adapt(s.handlerMeta(), getRouteMeta(), s.authorization(), s.useCache(time.Second*10), maxBody()))

	// heuristic
	s.handler.Handle(getRouteHeuristicByWorkID(),
		adapt(s.handlerHeuristicByWorkID(), getRouteHeuristicByWorkID(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicDetails(),
		adapt(s.handlerHeuristicsDetails(), getRouteHeuristicDetails(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicsExecution(),
		adapt(s.handlerHeuristicsExecution(), getRouteHeuristicsExecution(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicReport(),
		adapt(s.handlerHeuristicsReport(), getRouteHeuristicReport(), s.authorization(), maxBody()))

	// Analytics
	s.handler.Handle(getRouteShortestTransactionPath(),
		adapt(s.handlerShortestTransactionPath(), getRouteShortestTransactionPath(), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(getRouteConnectionLookup(),
		adapt(s.handlerConnectionLookup(), getRouteConnectionLookup(), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(getRouteMixingActivity(),
		adapt(s.handlerMixingActivity(), getRouteMixingActivity(), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(getRouteSpendingFingerprint(),
		adapt(s.handlerSpendingFingerprint(), getRouteSpendingFingerprint(), s.authorization(), s.useCache(time.Minute*10), maxBody()))

	// Clusters
	s.handler.Handle(getRouteClusterLookup(),
		adapt(s.handlerClusterLookup(), getRouteClusterLookup(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHMILookup(),
		adapt(s.handlerHMILookup(), getRouteHMILookup(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteClusterReport(),
		adapt(s.handlerClusterReport(), getRouteClusterReport(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddCluster(),
		adapt(s.handlerAddCluster(), getRouteAddCluster(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteCluster(),
		adapt(s.handlerDeleteCluster(), getRouteDeleteCluster(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllClusters(),
		adapt(s.handlerDeleteAllClusters(), getRouteDeleteAllClusters(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteClusterOverview(),
		adapt(s.handlerClusterOverview(), getRouteClusterOverview(), s.authorization(), maxBody()))

	// Attributions
	s.handler.Handle(getRouteAddPrivateAttribution(),
		adapt(s.handlerAddPrivateAttribution(), getRouteAddPrivateAttribution(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddPublicAttribution(),
		adapt(s.handlerAddPublicAttribution(), getRouteAddPublicAttribution(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAttributionList(),
		adapt(s.handlerAttributionList(), getRouteAttributionList(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeletePrivateAttribution(),
		adapt(s.handlerDeletePrivateAttribution(), getRouteDeletePrivateAttribution(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeletePublicAttribution(),
		adapt(s.handlerDeletePublicAttribution(), getRouteDeletePublicAttribution(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllPrivateAttributions(),
		adapt(s.handlerDeleteAllPrivateAttributions(), getRouteDeleteAllPrivateAttributions(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteSearchAttributions(),
		adapt(s.handlerSearchAttributions(), getRouteSearchAttributions(), s.authorization(), maxBody()))

	// Address Exclusions
	s.handler.Handle(getRouteAddAddressExclusions(),
		adapt(s.handlerAddAddressExclusions(), getRouteAddAddressExclusions(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAddressExclusion(),
		adapt(s.handlerDeleteAddressExclusion(), getRouteDeleteAddressExclusion(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllAddressExclusions(),
		adapt(s.handlerDeleteAllAddressExclusions(), getRouteDeleteAllAddressExclusions(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddressExclusionList(),
		adapt(s.handlerAddressExclusionList(), getRouteAddressExclusionList(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddressExclusionStatus(),
		adapt(s.handlerGetAddressExclusionStatus(), getRouteAddressExclusionStatus(), s.authorization(), maxBody()))

	// User
	s.handler.Handle(getRouteGetIdentities(),
		adapt(s.handlerGetIdentities(), getRouteGetIdentities(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteCreateIdentity(),
		adapt(s.handlerCreateIdentity(), getRouteCreateIdentity(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAdminDeleteIdentity(),
		adapt(s.handlerAdminDeleteIdentity(), getRouteAdminDeleteIdentity(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteIdentity(),
		adapt(s.handlerDeleteIdentity(), getRouteDeleteIdentity(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteModifyIdentity(),
		adapt(s.handlerModifyIdentity(), getRouteModifyIdentity(), s.authorization(), maxBody()))

	// Workspace
	s.handler.Handle(getRouteRenameWorkspace(),
		adapt(s.handlerRenameWorkspace(), getRouteRenameWorkspace(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteWorkspaceAddNode(),
		adapt(s.handlerAddWorkspaceNode(), getRouteWorkspaceAddNode(), s.authorization(), maxBodyConfig(50)))
	s.handler.Handle(getRouteWorkspaceAddNote(),
		adapt(s.handlerAddWorkspaceNote(), getRouteWorkspaceAddNote(), s.authorization(), maxBodyConfig(50)))
	s.handler.Handle(getRouteWorkspaceDeleteNode(),
		adapt(s.handlerDeleteWorkspaceNode(), getRouteWorkspaceDeleteNode(), s.authorization(), maxBodyConfig(50)))
	s.handler.Handle(getRouteWorkspaces(), adapt(s.handlerWorkspaces(), getRouteWorkspaces(), s.authorization()))
	s.handler.Handle(getRouteAddWorkspace(), adapt(s.handlerAddWorkspace(), getRouteAddWorkspace(), s.authorization()))
	s.handler.Handle(getRouteGetWorkspace(), adapt(s.handlerGetWorkspace(), getRouteGetWorkspace(), s.authorization()))
	s.handler.Handle(getRouteUpdateWorkspace(),
		adapt(s.handlerUpdateWorkspace(), getRouteUpdateWorkspace(), s.authorization(), maxBodyConfig(50)))
	s.handler.Handle(getRouteDeleteWorkspace(),
		adapt(s.handlerDeleteWorkspace(), getRouteDeleteWorkspace(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllWorkspaces(),
		adapt(s.handlerDeleteAllWorkspaces(), getRouteDeleteAllWorkspaces(), s.authorization(), maxBody()))
	s.handler.Handle(getRouteWorkspacesConnection(),
		adapt(s.handlerWorkspaceConnection(), getRouteWorkspacesConnection(), s.authorization(), s.useCache(time.Minute*10), maxBody()))
}
