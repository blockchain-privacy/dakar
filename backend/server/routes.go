package server

const (
	routePrefix string = "/api/v1/"

	routeSearch               string = "search"
	routeTransaction          string = "tx"
	routeBlock                string = "blk"
	routeAddress              string = "address"
	routeMeta                 string = "meta"
	routeHeuristicByWorkID    string = "heuristicByWorkID"
	routeHeuristics           string = "heuristics"
	routeHeuristicsSummary    string = "heuristicsSummary"
	routeHeuristicsExecution  string = "executeHeuristics"
	routeHeuristicDetails     string = "heuristicDetails"
	routeHeuristicList        string = "heuristicList"
	routeHeuristicDescriptors string = "heuristicDescriptors"
	routeDeleteHeuristic      string = "deleteHeuristic"
	routeAddressOutputRange   string = "addressOutputRange"
	routeBlockRange           string = "blkRange"
	routeShortestTxPath       string = "shortestTransactionPath"
	routeConnectionLookup     string = "connectionLookup"
	routeClusterSummary       string = "clusterSummary"
	routeHMILookup            string = "hmiLookup"
	routeMixingActivity       string = "mixingActivity"
	routeSpendingFingerprint  string = "spendingFingerprint"
	routeIdentities           string = "identities"
	routeExclusions           string = "exclusions"
	routeClusters             string = "clusters"
	routeAttributions         string = "attributions"
	routeAttributionsPublic   string = "attributions/public"
	routeAttributionsSearch   string = "attributions/search"
	routeAddWorkspaceNode     string = "addWorkspaceNode"
	routeWorkspaces           string = "workspaces"
	routeMetrics              string = "/metrics"
)

const (
	httpGET    = "GET"
	httpPOST   = "POST"
	httpDELETE = "DELETE"
	httpPUT    = "PUT"
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
	return buildRoutePattern(httpGET, routeHeuristicsSummary, "uid")
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
	return buildRoutePattern(httpPOST, routeAddressOutputRange, "hash")
}

// getRouteBlockRange returns a route
func getRouteBlockRange() string {
	return buildRoutePattern(httpPOST, routeBlockRange, "hash")
}

// getRouteCreateIdentity returns a route
func getRouteCreateIdentity() string {
	return buildRoutePattern(httpPOST, routeIdentities, "")
}

// getRouteGetIdentities returns a route
func getRouteGetIdentities() string {
	return buildRoutePattern(httpGET, routeIdentities, "")
}

// getRouteDeleteIdentity returns a route
func getRouteDeleteIdentity() string {
	return buildRoutePattern(httpDELETE, routeIdentities, "")
}

// getRouteAdminDeleteIdentity returns a route
func getRouteAdminDeleteIdentity() string {
	return buildRoutePattern(httpDELETE, routeIdentities, "uid")
}

// getRouteModifyIdentity returns a route
func getRouteModifyIdentity() string {
	return buildRoutePattern(httpPUT, routeIdentities, "")
}

// getRouteShortestTransactionPath returns a route
func getRouteShortestTransactionPath() string {
	return buildRoutePattern(httpPOST, routeShortestTxPath, "")
}

// getRouteConnectionLookup returns a route
func getRouteConnectionLookup() string {
	return buildRoutePattern(httpGET, routeConnectionLookup, "hash")
}

// getRouteClusterLookup returns a route
func getRouteClusterLookup() string {
	return buildRoutePattern(httpGET, routeClusters, "hash")
}

// getRouteHMILookup returns a route
func getRouteHMILookup() string {
	return buildRoutePattern(httpGET, routeHMILookup, "hash")
}

// getRouteClusterSummary returns a route
func getRouteClusterSummary() string {
	return buildRoutePattern(httpGET, routeClusterSummary, "hash")
}

// getRouteMixingActivity returns a route
func getRouteMixingActivity() string {
	return buildRoutePattern(httpPOST, routeMixingActivity, "")
}

// getRouteAddCluster returns a route
func getRouteAddCluster() string {
	return buildRoutePattern(httpPOST, routeClusters, "")
}

// getRouteDeleteCluster returns a route
func getRouteDeleteCluster() string {
	return buildRoutePattern(httpDELETE, routeClusters, "uid")
}

// getRouteDeleteAllClusters returns a route
func getRouteDeleteAllClusters() string {
	return buildRoutePattern(httpDELETE, routeClusters, "")
}

// getRouteClusterOverview returns a route
func getRouteClusterOverview() string {
	return buildRoutePattern(httpGET, routeClusters, "")
}

// getRouteAddPrivateAttribution returns a route
func getRouteAddPrivateAttribution() string {
	return buildRoutePattern(httpPOST, routeAttributions, "")
}

// getRouteAddPublicAttribution returns a route
func getRouteAddPublicAttribution() string {
	return buildRoutePattern(httpPOST, routeAttributionsPublic, "")
}

// getRouteAttributionList returns a route
func getRouteAttributionList() string {
	return buildRoutePattern(httpGET, routeAttributions, "")
}

// getRouteDeletePrivateAttribution returns a route
func getRouteDeletePrivateAttribution() string {
	return buildRoutePattern(httpDELETE, routeAttributions, "uid")
}

// getRouteDeletePublicAttribution returns a route
func getRouteDeletePublicAttribution() string {
	return buildRoutePattern(httpDELETE, routeAttributionsPublic, "uid")
}

// getRouteDeleteAllPrivateAttributions returns a route
func getRouteDeleteAllPrivateAttributions() string {
	return buildRoutePattern(httpDELETE, routeAttributions, "")
}

// getRouteSearchAttributions returns a route
func getRouteSearchAttributions() string {
	return buildRoutePattern(httpGET, routeAttributionsSearch, "query")
}

// getRouteAddAddressExclusions returns a route
func getRouteAddAddressExclusions() string {
	return buildRoutePattern(httpPOST, routeExclusions, "")
}

// getRouteDeleteAddressExclusion returns a route
func getRouteDeleteAddressExclusion() string {
	return buildRoutePattern(httpDELETE, routeExclusions, "hash")
}

// getRouteDeleteAllAddressExclusions returns a route
func getRouteDeleteAllAddressExclusions() string {
	return buildRoutePattern(httpDELETE, routeExclusions, "")
}

// getRouteAddressExclusionList returns a route
func getRouteAddressExclusionList() string {
	return buildRoutePattern(httpGET, routeExclusions, "")
}

// getRouteAddressExclusionStatus returns a route
func getRouteAddressExclusionStatus() string {
	return buildRoutePattern(httpGET, routeExclusions, "hash")
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
	return buildRoutePattern(httpPOST, routeWorkspaces, "name")
}

// getRouteGetWorkspace returns a route
func getRouteGetWorkspace() string {
	return buildRoutePattern(httpGET, routeWorkspaces, "uid")
}

// getRouteGetWorkspace returns a route
func getRouteUpdateWorkspace() string {
	return buildRoutePattern(httpPUT, routeWorkspaces, "")
}

// getRouteDeleteWorkspace returns a route
func getRouteDeleteWorkspace() string {
	return buildRoutePattern(httpDELETE, routeWorkspaces, "uid")
}

// getRouteDeleteAllWorkspaces returns a route
func getRouteDeleteAllWorkspaces() string {
	return buildRoutePattern(httpDELETE, routeWorkspaces, "")
}

// getRouteMetrics returns a route
func getRouteMetrics() string {
	return httpGET + " " + routeMetrics
}
