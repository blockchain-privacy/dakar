package server

import (
	"backend/external"
	"net/http"
	"path"
	"time"
)

// Search godoc
//
//	@Summary		Search for blocks, addresses and transactions
//	@Description	Search for blocks, addresses and transactions. Supports searching by hash or block ID.
//	@Tags			data
//	@Produce		json
//	@Param			query	path		string	true	"Hash"
//	@Success		200		{object}	server.searchReply
//	@Failure		500		{string}	string	"encoding error"
//	@Router			/search/{query} [get]
func (s *Server) handlerSearch() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getSearchReply(s.db, path.Base(r.URL.Path))

		sendReply(w, reply)
	})
}

// Query for Address, Transaction or Block godoc
//
//	@Summary	Query for address, transaction or block
//	@Tags		data
//	@Produce	json
//	@Param		hash	path		string	true	"Hash"
//	@Success	200		{object}	server.searchReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/blk/{hash} [get]
//	@Router		/address/{hash} [get]
//	@Router		/tx/{hash} [get]
func (s *Server) handlerDetails(fn func(external.Database, string) (SearchResult, bool, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDataDetailsReply(s.db, fn, path.Base(r.URL.Path))

		sendReply(w, reply)
	})
}

// Address Output Range godoc
//
//	@Summary	Get outputs of the given address
//	@Tags		data
//	@Produce	json
//	@Param		addressHash	path		string										true	"address hash"
//	@Param		options		body		server.getAddressOutputRangeReply.request	true	"query options"
//	@Success	200			{object}	server.searchReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/addressOutputRange/{addressHash} [post]
func (s *Server) handlerAddressOutputRange() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getAddressOutputRangeReply(r, s.db, path.Base(r.URL.Path))

		sendReply(w, reply)
	})
}

// Block Range godoc
//
//	@Summary	Get transactions of the given block
//	@Tags		data
//	@Produce	json
//	@Param		blockHash	path		string								true	"block hash"
//	@Param		offset		body		server.getBlockRangeReply.request	true	"transaction offset"
//	@Success	200			{object}	server.searchReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/blkRange/{blockHash} [post]
func (s *Server) handlerBlockRange() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getBlockRangeReply(r, s.db, path.Base(r.URL.Path))

		sendReply(w, reply)
	})
}

// Server Status godoc
//
//	@Summary	Get the status of all backend modules
//	@Tags		meta
//	@Produce	json
//	@Success	200	{object}	server.metaReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/meta/ [get]
func (s *Server) handlerMeta() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getMetaReply(s.db, s.client)

		sendReply(w, reply)
	})
}

// Heuristic Summary godoc
//
//	@Summary	Heuristic Summary
//	@Tags		heuristic
//	@Produce	text/csv
//	@Param		heuristic_UID	path		string	true	"0x123"
//	@Success	200				{file}		file	"comma separated values"
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/heuristicsSummary/{heuristic_UID} [get]
func (s *Server) handlerHeuristicsSummary() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeHeuristicSummary(w, r, s.db)
	})
}

// Cluster Summary godoc
//
//	@Summary	Get all clusters of an address
//	@Tags		cluster
//	@Produce	text/csv
//	@Accept		json
//	@Param		addressHash	path		string	true	"Address hash"
//	@Success	200			{file}		file	"comma separated values"
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/clusterSummary/{addressHash} [get]
func (s *Server) handlerClusterSummary() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeClusterSummary(w, r, s.db)
	})
}

// @Summary	Add Cluster
// @Tags		cluster
// @Produce	json
// @Param		separator	formData	string	true	"separator of the CSV file"
// @Param		hasHeader	formData	bool	true	"controls whether the first line should be skiped"
// @Param		file		formData	file	true	"the CSV file"
// @Success	200			{object}	server.addClusterReply
// @Failure	500			{string}	string	"encoding error"
// @Router		/addCluster/ [post]
func (s *Server) handlerAddCluster() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getAddClusterReply(s.db, r)

		sendReply(w, reply)
	})
}

// Delete Cluster godoc
//
//	@Summary	Delete Cluster
//	@Tags		cluster
//	@Produce	json
//	@Param		cluster_uid	path		string	true	"0x123"
//	@Success	200			{object}	server.deleteClusterReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/deleteCluster/{cluster_uid} [get]
func (s *Server) handlerDeleteCluster() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteClusterReply(r, s.db)

		sendReply(w, reply)
	})
}

// Delete All Clusters godoc
//
//	@Summary	Delete all clusters of the current user
//	@Tags		cluster
//	@Produce	json
//	@Success	200	{object}	server.deleteClusterReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/deleteAllClusters/ [get]
func (s *Server) handlerDeleteAllClusters() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteAllClustersReply(r, s.db)

		sendReply(w, reply)
	})
}

// Cluster Overview godoc
//
//	@Summary	Get all custom clusters of the current user
//	@Tags		cluster
//	@Produce	json
//	@Success	200	{object}	server.clusterOverviewReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/clusterOverview/ [get]
func (s *Server) handlerClusterOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getClusterOverviewReply(r, s.db)

		sendReply(w, reply)
	})
}

// Attribution Overview godoc
//
//	@Summary	Get all attributions of the current user
//	@Tags		attribution
//	@Produce	json
//	@Success	200	{object}	server.attributionOverviewReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/attributionOverview/ [get]
func (s *Server) handlerAttributionOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getAttributionOverviewReply(r, s.db)

		sendReply(w, reply)
	})
}

// Add Private Attribution godoc
//
//	@Summary	Create a new attribution which is only visible for the current user
//	@Tags		attribution
//	@Produce	text/csv
//	@Param		separator	formData	string	true	"separator of the CSV file"
//	@Param		hasHeader	formData	bool	true	"controls whether the first line should be skiped"
//	@Param		file		formData	file	true	"the CSV file"
//	@Success	200			{object}	server.addAttributionReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/addPrivateAttribution/ [post]
func (s *Server) handlerAddPrivateAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getAddAttributionReply(r, s.db, false)

		sendReply(w, reply)
	})
}

// Add Public Attribution godoc
//
//	@Summary	Create a new attribution which is visible to all users
//	@Tags		attribution
//	@Produce	text/csv
//	@Param		separator	formData	string	true	"separator of the CSV file"
//	@Param		hasHeader	formData	bool	true	"controls whether the first line should be skiped"
//	@Param		file		formData	file	true	"the CSV file"
//	@Success	200			{string}	string	"comma separated values"
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/addPublicAttribution/ [post]
func (s *Server) handlerAddPublicAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getAddAttributionReply(r, s.db, true)

		sendReply(w, reply)
	})
}

// Delete Private Attribution godoc
//
//	@Summary	Delete an attribution belonging to the current user
//	@Tags		attribution
//	@Produce	json
//	@Param		attribution_uid	path		string	true	"0x123"
//	@Success	200				{object}	server.deleteAttributionReply
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/deletePrivateAttribution/{attribution_uid} [get]
func (s *Server) handlerDeletePrivateAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteAttributionReply(r, s.db, false)

		sendReply(w, reply)
	})
}

// Delete Public Attribution godoc
//
//	@Summary	Delete a public attribution
//	@Tags		attribution
//	@Produce	json
//	@Param		attribution_uid	path		string	true	"0x123"
//	@Success	200				{object}	server.deleteAttributionReply
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/deletePublicAttribution/{attribution_uid} [get]
func (s *Server) handlerDeletePublicAttribution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteAttributionReply(r, s.db, true)

		sendReply(w, reply)
	})
}

// Delete All Private Attributions godoc
//
//	@Summary	Delete all attributions of the current user
//	@Tags		attribution
//	@Produce	json
//	@Success	200	{object}	server.deleteAttributionReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/deleteAllPrivateAttributions/ [get]
func (s *Server) handlerDeleteAllPrivateAttributions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteAllAttributionsReply(r, s.db)

		sendReply(w, reply)
	})
}

// Search Attributions godoc
//
//	@Summary	Search all public attributions and the attributions belonging to the current user
//	@Tags		attribution
//	@Accept		json
//	@Produce	json
//	@Param		attribution	body		server.getAttributionSearchReply.request	true	"Search query"
//	@Success	200			{object}	server.attributionOverviewReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/searchAttributions/ [post]
func (s *Server) handlerSearchAttributions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getAttributionSearchReply(r, s.db)

		sendReply(w, reply)
	})
}

// Add Address Exclusion godoc
//
//	@Summary	Adds new address exclusions for the current user
//	@Tags		address exclusions
//	@Produce	text/csv
//	@Param		file	formData	file	true	"the CSV file"
//	@Success	200		{object}	server.addAddressExclusionsReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/addAddressExclusions/ [post]
func (s *Server) handlerAddAddressExclusions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getAddAddressExclusionsReply(s.db, r)

		sendReply(w, reply)
	})
}

// Delete Address Exclusion godoc
//
//	@Summary	Deletes an address exclusion of the current user
//	@Tags		address exclusions
//	@Produce	json
//	@Param		addressHash	path		string	true	"0x123"
//	@Success	200			{object}	server.deleteAddressExclusionReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/deleteAddressExclusion/{addressHash} [get]
func (s *Server) handlerDeleteAddressExclusion() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteAddressExclusionReply(r, s.db)

		sendReply(w, reply)
	})
}

// Delete All Address Exclusions godoc
//
//	@Summary	Delete all address exclusions of the current user
//	@Tags		address exclusions
//	@Produce	json
//	@Success	200	{object}	server.deleteAddressExclusionReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/deleteAllAddressExclusions/ [get]
func (s *Server) handlerDeleteAllAddressExclusions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteAllAddressExclusionsReply(r, s.db)

		sendReply(w, reply)
	})
}

// Address Exclusion Overview godoc
//
//	@Summary	Address exclusion overview
//	@Tags		address exclusions
//	@Produce	json
//	@Success	200	{object}	server.addressExclusionOverviewReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/addressExclusionOverview/ [get]
func (s *Server) handlerAddressExclusionOverview() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getAddressExclusionOverviewReply(r, s.db)

		sendReply(w, reply)
	})
}

// Heuristic godoc
//
//	@Summary	Get all heuristic defined for a transaction
//	@Tags		heuristic
//	@Produce	json
//	@Param		hash	path		string	true	"0x123"
//	@Success	200		{object}	server.heuristicReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/heuristics/{hash} [get]
func (s *Server) handlerHeuristics() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getHeuristicReply(r, s.db, s.worker)

		sendReply(w, reply)
	})
}

// HMI clusters godoc
//
//	@Summary	Get HMI clusters per transaction
//	@Tags		cluster
//	@Produce	json
//	@Param		hash	path		string	true	"0x123"
//	@Success	200		{object}	server.hmiLookupReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/hmiLookup/{hash} [get]
func (s *Server) handlerHMILookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getHMILookupReply(s.db, path.Base(r.URL.Path))

		sendReply(w, reply)
	})
}

// Heuristic Status godoc
//
//	@Summary	Get the status of all heuristics per transaction
//	@Tags		heuristic
//	@Produce	json
//	@Param		hash	path		string	true	"0x123"
//	@Success	200		{object}	server.heuristicReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/heuristicStatus/{hash} [get]
func (s *Server) handlerHeuristicStatus() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getHeuristicStatusReply(r, s.worker)

		sendReply(w, reply)
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
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/heuristicDetails/ [post]
func (s *Server) handlerHeuristicsDetails() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getHeuristicDetailsReply(r, s.db)

		sendReply(w, reply)
	})
}

// Execute Heuristics godoc
//
//	@Summary		Queues the execution of heuristics for the given transaction
//	@Description	This call queues the given heuristics for the given transaction. Does not wait until the heuristic execution is finished.
//	@Tags			heuristic
//	@Produce		json
//	@Accept			json
//	@Param			hash		path		string										true	"0x123"
//	@Param			heuristic	body		server.getHeuristicExecutionReply.request	true	"Heuristics to queue"
//	@Success		200			{object}	server.heuristicExecutionReply
//	@Failure		500			{string}	string	"encoding error"
//	@Router			/executeHeuristics/{hash} [post]
func (s *Server) handlerHeuristicsExecution() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getHeuristicExecutionReply(r, s.db, s.worker)

		sendReply(w, reply)
	})
}

// List Heuristics godoc
//
//	@Summary	Lists all heuristics of the current user
//	@Tags		heuristic
//	@Produce	json
//	@Success	200	{object}	server.heuristicListReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/heuristicList/ [get]
func (s *Server) handlerHeuristicList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getHeuristicListReply(r, s.db)

		sendReply(w, reply)
	})
}

// Heuristic Descriptors godoc
//
//	@Summary		Gets available heuristic descriptors
//	@Description	Returns available heuristic descriptors, which define the heuristic interface
//	@Tags			heuristic
//	@Produce		json
//	@Success		200	{object}	server.heuristicDescriptorReply
//	@Failure		500	{string}	string	"encoding error"
//	@Router			/heuristicDescriptors/ [get]
func (s *Server) handlerHeuristicDescriptors() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getHeuristicDescriptorReply()

		sendReply(w, reply)
	})
}

// Delete Heuristic godoc
//
//	@Summary		Deletes either all heuristics or all heuristics of a transaction
//	@Description	Deletes either all heuristics of the current user or all heuristics of a transaction of the current user
//	@Tags			heuristic
//	@Produce		json
//	@Accept			json
//	@Param			heuristic	body		server.getDeleteHeuristicReply.request	true	"Heuristic deletion request. Set delete_all to true, only if ALL heuristic should be deleted."
//	@Success		200			{object}	server.deleteHeuristicReply
//	@Failure		500			{string}	string	"encoding error"
//	@Router			/deleteHeuristic/ [post]
func (s *Server) handlerDeleteHeuristic() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteHeuristicReply(r, s.db)

		sendReply(w, reply)
	})
}

// Create Identity godoc
//
//	@Summary	Create a new identity
//	@Tags		authentication
//	@Produce	json
//	@Accept		json
//	@Param		identity	body		server.getCreateIdentityReply.request	true	"Identity details"
//	@Success	200			{object}	server.identityReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/createIdentity/ [post]
//
// handlerCreateIdentity creates a new identity. This is an admin endpoint.
func (s *Server) handlerCreateIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getCreateIdentityReply(s.db, s.adminAuth, r)

		sendReply(w, reply)
	})
}

// Admin Delete Identity godoc
//
//	@Summary	Delete an arbitrary identity
//	@Tags		authentication
//	@Produce	json
//	@Param		identityUID	path		string	true	"0x123"
//	@Success	200			{object}	server.identityReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/adminDeleteIdentity/{identityUID} [get]
//
// handlerAdminDeleteIdentity deletes an arbitrary identity. This is an admin endpoint.
func (s *Server) handlerAdminDeleteIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteIdentityReply(r, s.db, s.adminAuth, true)

		sendReply(w, reply)
	})
}

// Delete Identity godoc
//
//	@Summary	Delete the identity of the current user
//	@Tags		authentication
//	@Produce	json
//	@Success	200	{object}	server.identityReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/deleteIdentity/ [get]
//
// handlerDeleteIdentity deletes the calling users identity.
func (s *Server) handlerDeleteIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getDeleteIdentityReply(r, s.db, s.adminAuth, false)

		sendReply(w, reply)
	})
}

// Modify Identity godoc
//
//	@Summary	Modify an arbitrary identity
//	@Tags		authentication
//	@Produce	json
//	@Accept		json
//	@Param		identity	body		server.getModifyIdentityReply.request	true	"Identity modification details"
//	@Success	200			{object}	server.identityReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/modifyIdentity/ [post]
//
// handlerModifyIdentity modifies an arbitrary identity. This is an admin endpoint.
func (s *Server) handlerModifyIdentity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getModifyIdentityReply(s.adminAuth, r)

		sendReply(w, reply)
	})
}

// Get Identities godoc
//
//	@Summary	Get all identities
//	@Tags		authentication
//	@Produce	json
//	@Success	200	{object}	server.identitiesReply
//	@Failure	500	{string}	string	"encoding error"
//	@Router		/getIdentities/ [get]
func (s *Server) handlerGetIdentities() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getIdentitiesReply(s.db, s.adminAuth, r)

		sendReply(w, reply)
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
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/shortestTransactionPath/ [post]
func (s *Server) handlerShortestTransactionPath() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getShortestTransactionPathReply(s.db, r.Body)

		sendReply(w, reply)
	})
}

// Connection Lookup Path godoc
//
//	@Summary	Connection lookup
//	@Tags		tools
//	@Produce	json
//	@Param		txhash	path		string	true	"Transaction hash"
//	@Param		forward	query		bool	false	"search direction"
//	@Param		t		query		int		false	"number of days to look back"	maximum(90)
//	@Success	200		{object}	server.connectionLookupReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/connectionLookup/{txhash} [get]
func (s *Server) handlerConnectionLookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getConnectionLookupReply(s.db, s.worker, r.URL)

		sendReply(w, reply)
	})
}

// Cluster Lookup Path godoc
//
//	@Summary	Get all clusters of the given address
//	@Tags		cluster
//	@Produce	json
//	@Param		addressHash	path		string	true	"Address hash"
//	@Success	200			{object}	server.clusterLookupReply
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/clusterLookup/{addressHash} [get]
func (s *Server) handlerClusterLookup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getClusterLookupReply(r, s.db)

		sendReply(w, reply)
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
//	@Failure	500			{string}	string	"encoding error"
//	@Router		/mixingActivity/ [post]
func (s *Server) handlerMixingActivity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getMixingActivity(s.db, r.Body)

		sendReply(w, reply)
	})
}

// Address Exclusion Status godoc
//
//	@Summary	Get the exclusion status of an address
//	@Tags		address exclusions
//	@Produce	json
//	@Param		address_hash	path		string	true	"address hash"
//	@Success	200				{object}	server.addressExclusionStatusReply
//	@Failure	500				{string}	string	"encoding error"
//	@Router		/addressExclusionStatus/{address_hash} [get]
func (s *Server) handlerGetAddressExclusionStatus() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getAddressExclusionStatusReply(r, s.db, path.Base(r.URL.Path))

		sendReply(w, reply)
	})
}

// Spending Fingerprint godoc
//
//	@Summary	Get the spending fingerprint of a transaction
//	@Tags		tools
//	@Produce	json
//	@Param		hash	path		string	true	"transaction hash"
//	@Success	200		{object}	server.spendingFingerprintReply
//	@Failure	500		{string}	string	"encoding error"
//	@Router		/spendingFingerprint/{hash} [get]
func (s *Server) handlerSpendingFingerprint() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reply := getSpendingFingerprintReply(s.db, s.worker, path.Base(r.URL.Path))

		sendReply(w, reply)
	})
}

// todo: when https://github.com/golang/go/issues/61410 is merged and released,
// refactor API design to use consistent endpoint naming:
// - GET clusters - returns all clusters
// - GET clusters/{addresshash} - returns all clusters belonging to a specific address
// - DELETE clusters - deletes all clusters
// - DELETE clusters/{addresshash} - deletes all clusters belonging to a specific address
// - POST clusters - creates a new clusters
// setupHandlers creates endpoint handlers
func (s *Server) setupHandlers() {
	// Search
	s.handler.Handle(getRouteSearch(), adapt(s.handlerSearch(), getRouteSearch(),
		limitMethod("GET"), s.useCache(time.Minute*10), maxBody()))

	// Common data
	s.handler.Handle(getRouteTransaction(),
		adapt(s.handlerDetails(GetTransaction), getRouteTransaction(),
			limitMethod("GET"), s.useCache(time.Second*0), maxBody()))
	// setting block cache time to 10 Minutes because blocks at
	// the tip get updated via adding the 'next block' reference
	s.handler.Handle(getRouteBlock(), adapt(s.handlerDetails(GetBlock), getRouteBlock(),
		limitMethod("GET"), s.useCache(time.Second*10), maxBody()))
	s.handler.Handle(getRouteAddress(), adapt(s.handlerDetails(GetAddress), getRouteAddress(),
		limitMethod("GET"), s.useCache(time.Second*10), maxBody()))

	s.handler.Handle(getRouteAddressOutputRange(),
		adapt(s.handlerAddressOutputRange(), getRouteAddressOutputRange(),
			limitMethod("POST"), s.useCache(time.Minute*10), maxBody()))

	s.handler.Handle(getRouteBlockRange(), adapt(s.handlerBlockRange(), getRouteBlockRange(),
		limitMethod("POST"), s.useCache(time.Minute*10), maxBody()))

	// Meta
	s.handler.Handle(getRouteMeta(), adapt(s.handlerMeta(), getRouteMeta(),
		limitMethod("GET"), s.authorization(), s.useCache(time.Second*10), maxBody()))

	// heuristic
	s.handler.Handle(getRouteHeuristics(),
		adapt(s.handlerHeuristics(), getRouteHeuristics(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicStatus(),
		adapt(s.handlerHeuristicStatus(), getRouteHeuristicStatus(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicDetails(),
		adapt(s.handlerHeuristicsDetails(), getRouteHeuristicDetails(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicsExecution(),
		adapt(s.handlerHeuristicsExecution(), getRouteHeuristicsExecution(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicsSummary(),
		adapt(s.handlerHeuristicsSummary(), getRouteHeuristicsSummary(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicList(),
		adapt(s.handlerHeuristicList(), getRouteHeuristicList(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHeuristicDescriptors(),
		adapt(s.handlerHeuristicDescriptors(), getRouteHeuristicDescriptors(),
			limitMethod("GET"), s.authorization(), s.useCache(0), maxBody()))
	s.handler.Handle(getRouteDeleteHeuristic(),
		adapt(s.handlerDeleteHeuristic(), getRouteDeleteHeuristic(),
			limitMethod("POST"), s.authorization(), maxBody()))

	// Analytics
	s.handler.Handle(getRouteShortestTransactionPath(),
		adapt(s.handlerShortestTransactionPath(), getRouteShortestTransactionPath(),
			limitMethod("POST"), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(getRouteConnectionLookup(),
		adapt(s.handlerConnectionLookup(), getRouteConnectionLookup(),
			limitMethod("GET"), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(getRouteMixingActivity(),
		adapt(s.handlerMixingActivity(), getRouteMixingActivity(),
			limitMethod("POST"), s.authorization(), s.useCache(time.Minute*10), maxBody()))
	s.handler.Handle(getRouteSpendingFingerprint(),
		adapt(s.handlerSpendingFingerprint(), getRouteSpendingFingerprint(), limitMethod("GET"),
			s.authorization(), s.useCache(time.Minute*10), maxBody()))

	// Clusters
	s.handler.Handle(getRouteClusterLookup(),
		adapt(s.handlerClusterLookup(), getRouteClusterLookup(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteHMILookup(),
		adapt(s.handlerHMILookup(), getRouteHMILookup(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteClusterSummary(),
		adapt(s.handlerClusterSummary(), getRouteClusterSummary(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddCluster(),
		adapt(s.handlerAddCluster(), getRouteAddCluster(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteCluster(),
		adapt(s.handlerDeleteCluster(), getRouteDeleteCluster(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllClusters(),
		adapt(s.handlerDeleteAllClusters(), getRouteDeleteAllClusters(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteClusterOverview(),
		adapt(s.handlerClusterOverview(), getRouteClusterOverview(),
			limitMethod("GET"), s.authorization(), maxBody()))

	// Attributions
	s.handler.Handle(getRouteAddPrivateAttribution(),
		adapt(s.handlerAddPrivateAttribution(), getRouteAddPrivateAttribution(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddPublicAttribution(),
		adapt(s.handlerAddPublicAttribution(), getRouteAddPublicAttribution(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAttributionOverview(),
		adapt(s.handlerAttributionOverview(), getRouteAttributionOverview(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeletePrivateAttribution(),
		adapt(s.handlerDeletePrivateAttribution(), getRouteDeletePrivateAttribution(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeletePublicAttribution(),
		adapt(s.handlerDeletePublicAttribution(), getRouteDeletePublicAttribution(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllPrivateAttributions(),
		adapt(s.handlerDeleteAllPrivateAttributions(), getRouteDeleteAllPrivateAttributions(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteSearchAttributions(),
		adapt(s.handlerSearchAttributions(), getRouteSearchAttributions(),
			limitMethod("POST"), s.authorization(), maxBody()))

	// Address Exclusions
	s.handler.Handle(getRouteAddAddressExclusions(),
		adapt(s.handlerAddAddressExclusions(), getRouteAddAddressExclusions(),
			limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAddressExclusion(),
		adapt(s.handlerDeleteAddressExclusion(), getRouteDeleteAddressExclusion(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteAllAddressExclusions(),
		adapt(s.handlerDeleteAllAddressExclusions(), getRouteDeleteAllAddressExclusions(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddressExclusionOverview(),
		adapt(s.handlerAddressExclusionOverview(), getRouteAddressExclusionOverview(),
			limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAddressExclusionStatus(),
		adapt(s.handlerGetAddressExclusionStatus(), getRouteAddressExclusionStatus(),
			limitMethod("GET"), s.authorization(), maxBody()))

	// User
	s.handler.Handle(getRouteGetIdentities(), adapt(s.handlerGetIdentities(), getRouteGetIdentities(),
		limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteCreateIdentity(), adapt(s.handlerCreateIdentity(), getRouteCreateIdentity(),
		limitMethod("POST"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteAdminDeleteIdentity(), adapt(s.handlerAdminDeleteIdentity(), getRouteDeleteIdentity(),
		limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteDeleteIdentity(), adapt(s.handlerDeleteIdentity(), getRouteDeleteIdentity(),
		limitMethod("GET"), s.authorization(), maxBody()))
	s.handler.Handle(getRouteModifyIdentity(), adapt(s.handlerModifyIdentity(), getRouteModifyIdentity(),
		limitMethod("POST"), s.authorization(), maxBody()))
}
