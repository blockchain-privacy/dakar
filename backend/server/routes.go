package server

const (
	routePrefix string = "/api/v1/"

	routeSearch                       string = "search"
	routeTransaction                  string = "tx"
	routeBlock                        string = "blk"
	routeAddress                      string = "address"
	routeMeta                         string = "meta"
	routeHeuristicByWorkID            string = "heuristicByWorkID"
	routeHeuristics                   string = "heuristics"
	routeHeuristicsSummary            string = "heuristicsSummary"
	routeHeuristicsExecution          string = "executeHeuristics"
	routeHeuristicDetails             string = "heuristicDetails"
	routeHeuristicList                string = "heuristicList"
	routeHeuristicDescriptors         string = "heuristicDescriptors"
	routeDeleteHeuristic              string = "deleteHeuristic"
	routeAddressOutputRange           string = "addressOutputRange"
	routeBlockRange                   string = "blkRange"
	routeCreateIdentity               string = "createIdentity"
	routeGetIdentities                string = "getIdentities"
	routeDeleteIdentity               string = "deleteIdentity"
	routeAdminDeleteIdentity          string = "adminDeleteIdentity"
	routeModifyIdentity               string = "modifyIdentity"
	routeShortestTxPath               string = "shortestTransactionPath"
	routeConnectionLookup             string = "connectionLookup"
	routeClusterLookup                string = "clusterLookup"
	routeClusterSummary               string = "clusterSummary"
	routeHMILookup                    string = "hmiLookup"
	routeMixingActivity               string = "mixingActivity"
	routeAddCluster                   string = "addCluster"
	routeDeleteCluster                string = "deleteCluster"
	routeDeleteAllClusters            string = "deleteAllClusters"
	routeClusterOverview              string = "clusterOverview"
	routeAttributionOverview          string = "attributionOverview"
	routeAddPrivateAttribution        string = "addPrivateAttribution"
	routeAddPublicAttribution         string = "addPublicAttribution"
	routeDeletePrivateAttribution     string = "deletePrivateAttribution"
	routeDeletePublicAttribution      string = "deletePublicAttribution"
	routeDeleteAllPrivateAttributions string = "deleteAllPrivateAttributions"
	routeSearchAttributions           string = "searchAttributions"
	routeAddAddressExclusions         string = "addAddressExclusions"
	routeDeleteAddressExclusion       string = "deleteAddressExclusion"
	routeDeleteAllAddressExclusions   string = "deleteAllAddressExclusions"
	routeAddressExclusionOverview     string = "addressExclusionOverview"
	routeAddressExclusionStatus       string = "addressExclusionStatus"
	routeSpendingFingerprint          string = "spendingFingerprint"
	routeAddWorkspaceNode             string = "addWorkspaceNode"
	routeAddWorkspace                 string = "addWorkspace"
	routeWorkspaces                   string = "workspaces"
	routeGetWorkspace                 string = "getWorkspace"
	routeUpdateWorkspace              string = "updateWorkspace"
	routeDeleteWorkspace              string = "deleteWorkspace"
	routeMetrics                      string = "/metrics"
)

const (
	httpGET  = "GET"
	httpPOST = "POST"
)

func buildRoutePattern(httpMethod string, r string, query string) string {
	base := httpMethod + " " + routePrefix + r + "/"

	if query != "" {
		base += "{" + query + "}"
	}
	return base
}

// getRouteTransaction returns a route
func getRouteTransaction() string {
	return buildRoutePattern(httpGET, routeTransaction, "hash")
}

// getRouteBlock returns a route
func getRouteBlock() string {
	return buildRoutePattern(httpGET, routeBlock, "hash")
}

// getRouteAddress returns a route
func getRouteAddress() string {
	return buildRoutePattern(httpGET, routeAddress, "hash")
}

// getRouteMeta returns a route
func getRouteMeta() string {
	return buildRoutePattern(httpGET, routeMeta, "")
}

// getRouteHeuristicByWorkID returns a route
func getRouteHeuristicByWorkID() string {
	return buildRoutePattern(httpGET, routeHeuristicByWorkID, "workID")
}

// getRouteHeuristics returns a route
func getRouteHeuristics() string {
	return buildRoutePattern(httpGET, routeHeuristics, "hash")
}

// getRouteHeuristicsSummary returns a route
func getRouteHeuristicsSummary() string {
	return buildRoutePattern(httpGET, routeHeuristicsSummary, "heuristic_UID")
}

// getRouteHeuristicsExecution returns a route
func getRouteHeuristicsExecution() string {
	return buildRoutePattern(httpPOST, routeHeuristicsExecution, "hash")
}

// getRouteHeuristicDetails returns a route
func getRouteHeuristicDetails() string {
	return buildRoutePattern(httpPOST, routeHeuristicDetails, "")
}

// getRouteHeuristicList returns a route
func getRouteHeuristicList() string {
	return buildRoutePattern(httpGET, routeHeuristicList, "")
}

// getRouteHeuristicDescriptors returns a route
func getRouteHeuristicDescriptors() string {
	return buildRoutePattern(httpGET, routeHeuristicDescriptors, "")
}

// getRouteDeleteHeuristic returns a route
func getRouteDeleteHeuristic() string {
	return buildRoutePattern(httpPOST, routeDeleteHeuristic, "")
}

// getRouteSearch returns a route
func getRouteSearch() string {
	return buildRoutePattern(httpGET, routeSearch, "query")
}

// getRouteAddressOutputRange returns a route
func getRouteAddressOutputRange() string {
	return buildRoutePattern(httpPOST, routeAddressOutputRange, "addressHash")
}

// getRouteBlockRange returns a route
func getRouteBlockRange() string {
	return buildRoutePattern(httpPOST, routeBlockRange, "blockHash")
}

// getRouteCreateIdentity returns a route
func getRouteCreateIdentity() string {
	return buildRoutePattern(httpPOST, routeCreateIdentity, "")
}

// getRouteGetIdentities returns a route
func getRouteGetIdentities() string {
	return buildRoutePattern(httpGET, routeGetIdentities, "")
}

// getRouteDeleteIdentity returns a route
func getRouteDeleteIdentity() string {
	return buildRoutePattern(httpGET, routeDeleteIdentity, "")
}

// getRouteAdminDeleteIdentity returns a route
func getRouteAdminDeleteIdentity() string {
	return buildRoutePattern(httpGET, routeAdminDeleteIdentity, "identityUID")
}

// getRouteModifyIdentity returns a route
func getRouteModifyIdentity() string {
	return buildRoutePattern(httpPOST, routeModifyIdentity, "")
}

// getRouteShortestTransactionPath returns a route
func getRouteShortestTransactionPath() string {
	return buildRoutePattern(httpPOST, routeShortestTxPath, "")
}

// getRouteConnectionLookup returns a route
func getRouteConnectionLookup() string {
	return buildRoutePattern(httpGET, routeConnectionLookup, "txHash")
}

// getRouteClusterLookup returns a route
func getRouteClusterLookup() string {
	return buildRoutePattern(httpGET, routeClusterLookup, "addressHash")
}

// getRouteHMILookup returns a route
func getRouteHMILookup() string {
	return buildRoutePattern(httpGET, routeHMILookup, "hash")
}

// getRouteClusterSummary returns a route
func getRouteClusterSummary() string {
	return buildRoutePattern(httpGET, routeClusterSummary, "addressHash")
}

// getRouteMixingActivity returns a route
func getRouteMixingActivity() string {
	return buildRoutePattern(httpPOST, routeMixingActivity, "")
}

// getRouteAddCluster returns a route
func getRouteAddCluster() string {
	return buildRoutePattern(httpPOST, routeAddCluster, "")
}

// getRouteDeleteCluster returns a route
func getRouteDeleteCluster() string {
	return buildRoutePattern(httpGET, routeDeleteCluster, "cluster_uid")
}

// getRouteDeleteAllClusters returns a route
func getRouteDeleteAllClusters() string {
	return buildRoutePattern(httpGET, routeDeleteAllClusters, "")
}

// getRouteClusterOverview returns a route
func getRouteClusterOverview() string {
	return buildRoutePattern(httpGET, routeClusterOverview, "")
}

// getRouteAddPrivateAttribution returns a route
func getRouteAddPrivateAttribution() string {
	return buildRoutePattern(httpPOST, routeAddPrivateAttribution, "")
}

// getRouteAddPublicAttribution returns a route
func getRouteAddPublicAttribution() string {
	return buildRoutePattern(httpPOST, routeAddPublicAttribution, "")
}

// getRouteAttributionOverview returns a route
func getRouteAttributionOverview() string {
	return buildRoutePattern(httpGET, routeAttributionOverview, "")
}

// getRouteDeletePrivateAttribution returns a route
func getRouteDeletePrivateAttribution() string {
	return buildRoutePattern(httpGET, routeDeletePrivateAttribution, "attribution_uid")
}

// getRouteDeletePublicAttribution returns a route
func getRouteDeletePublicAttribution() string {
	return buildRoutePattern(httpGET, routeDeletePublicAttribution, "attribution_uid")
}

// getRouteDeleteAllPrivateAttributions returns a route
func getRouteDeleteAllPrivateAttributions() string {
	return buildRoutePattern(httpGET, routeDeleteAllPrivateAttributions, "")
}

// getRouteSearchAttributions returns a route
func getRouteSearchAttributions() string {
	return buildRoutePattern(httpPOST, routeSearchAttributions, "")
}

// getRouteAddAddressExclusions returns a route
func getRouteAddAddressExclusions() string {
	return buildRoutePattern(httpPOST, routeAddAddressExclusions, "")
}

// getRouteDeleteAddressExclusion returns a route
func getRouteDeleteAddressExclusion() string {
	return buildRoutePattern(httpGET, routeDeleteAddressExclusion, "addressHash")
}

// getRouteDeleteAllAddressExclusions returns a route
func getRouteDeleteAllAddressExclusions() string {
	return buildRoutePattern(httpGET, routeDeleteAllAddressExclusions, "")
}

// getRouteAddressExclusionOverview returns a route
func getRouteAddressExclusionOverview() string {
	return buildRoutePattern(httpGET, routeAddressExclusionOverview, "")
}

// getRouteAddressExclusionStatus returns a route
func getRouteAddressExclusionStatus() string {
	return buildRoutePattern(httpGET, routeAddressExclusionStatus, "address_hash")
}

// getRouteSpendingFingerprint returns a route
func getRouteSpendingFingerprint() string {
	return buildRoutePattern(httpGET, routeSpendingFingerprint, "hash")
}

// getRouteWorkspaceAddNode returns a route
func getRouteWorkspaceAddNode() string {
	return buildRoutePattern(httpPOST, routeAddWorkspaceNode, "")
}

// getRouteWorkspaces returns a route
func getRouteWorkspaces() string {
	return buildRoutePattern(httpGET, routeWorkspaces, "")
}

// getRouteAddWorkspace returns a route
func getRouteAddWorkspace() string {
	return buildRoutePattern(httpGET, routeAddWorkspace, "name")
}

// getRouteGetWorkspace returns a route
func getRouteGetWorkspace() string {
	return buildRoutePattern(httpGET, routeGetWorkspace, "uid")
}

// getRouteGetWorkspace returns a route
func getRouteUpdateWorkspace() string {
	return buildRoutePattern(httpPOST, routeUpdateWorkspace, "")
}

// getRouteDeleteWorkspace returns a route
func getRouteDeleteWorkspace() string {
	return buildRoutePattern(httpPOST, routeDeleteWorkspace, "")
}

// getRouteMetrics returns a route
func getRouteMetrics() string {
	return httpGET + " " + routeMetrics
}
