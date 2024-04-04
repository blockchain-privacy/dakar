package server

const (
	routePrefix string = "/api/v1/"

	routeSearch               string = "blockchain/search"
	routeTransaction          string = "blockchain/transactions"
	routeBlock                string = "blockchain/blocks"
	routeAddress              string = "blockchain/addresses"
	routeAddressOutputRange   string = "blockchain/outputs"
	routeMeta                 string = "meta"
	routeHeuristicByWorkID    string = "heuristicByWorkID"
	routeHeuristicReport      string = "heuristics/report"
	routeHeuristicsExecution  string = "executeHeuristics"
	routeHeuristicDetails     string = "heuristicDetails"
	routeHeuristicDescriptors string = "heuristicDescriptors"
	routeShortestTxPath       string = "shortestTransactionPath"
	routeConnectionLookup     string = "connectionLookup"
	routeMixingActivity       string = "mixingActivity"
	routeSpendingFingerprint  string = "spendingFingerprint"
	routeIdentities           string = "identities"
	routeExclusions           string = "exclusions"
	routeClusters             string = "clusters"
	routeHMILookup            string = "clusters/hmi"
	clusterReport             string = "clusters/report"
	routeAttributions         string = "attributions"
	routeAttributionsPublic   string = "attributions/public"
	routeAttributionsSearch   string = "attributions/search"
	routeWorkspaces           string = "workspaces"
	routeAddWorkspaceNode     string = "workspaces/node"
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

func getRouteTransaction() string {
	return buildRoutePattern(httpGET, routeTransaction, "hash")
}

func getRouteBlock() string {
	return buildRoutePattern(httpGET, routeBlock, "hash")
}

func getRouteAddress() string {
	return buildRoutePattern(httpGET, routeAddress, "hash")
}

func getRouteMeta() string {
	return buildRoutePattern(httpGET, routeMeta, "")
}

func getRouteHeuristicByWorkID() string {
	return buildRoutePattern(httpPOST, routeHeuristicByWorkID, "")
}

func getRouteHeuristicReport() string {
	return buildRoutePattern(httpPOST, routeHeuristicReport, "")
}

func getRouteHeuristicsExecution() string {
	return buildRoutePattern(httpPOST, routeHeuristicsExecution, "")
}

func getRouteHeuristicDetails() string {
	return buildRoutePattern(httpPOST, routeHeuristicDetails, "")
}

func getRouteHeuristicDescriptors() string {
	return buildRoutePattern(httpGET, routeHeuristicDescriptors, "")
}

func getRouteSearch() string {
	return buildRoutePattern(httpGET, routeSearch, "query")
}

func getRouteAddressOutputRange() string {
	return buildRoutePattern(httpPOST, routeAddressOutputRange, "hash")
}

func getRouteCreateIdentity() string {
	return buildRoutePattern(httpPOST, routeIdentities, "")
}

func getRouteGetIdentities() string {
	return buildRoutePattern(httpGET, routeIdentities, "")
}

func getRouteDeleteIdentity() string {
	return buildRoutePattern(httpDELETE, routeIdentities, "")
}

func getRouteAdminDeleteIdentity() string {
	return buildRoutePattern(httpDELETE, routeIdentities, "uid")
}

func getRouteModifyIdentity() string {
	return buildRoutePattern(httpPUT, routeIdentities, "")
}

func getRouteShortestTransactionPath() string {
	return buildRoutePattern(httpPOST, routeShortestTxPath, "")
}

func getRouteConnectionLookup() string {
	return buildRoutePattern(httpGET, routeConnectionLookup, "hash")
}

func getRouteClusterLookup() string {
	return buildRoutePattern(httpGET, routeClusters, "hash")
}

func getRouteHMILookup() string {
	return buildRoutePattern(httpGET, routeHMILookup, "hash")
}

func getRouteClusterReport() string {
	return buildRoutePattern(httpGET, clusterReport, "hash")
}

func getRouteMixingActivity() string {
	return buildRoutePattern(httpPOST, routeMixingActivity, "")
}

func getRouteAddCluster() string {
	return buildRoutePattern(httpPOST, routeClusters, "")
}

func getRouteDeleteCluster() string {
	return buildRoutePattern(httpDELETE, routeClusters, "uid")
}

func getRouteDeleteAllClusters() string {
	return buildRoutePattern(httpDELETE, routeClusters, "")
}

func getRouteClusterOverview() string {
	return buildRoutePattern(httpGET, routeClusters, "")
}

func getRouteAddPrivateAttribution() string {
	return buildRoutePattern(httpPOST, routeAttributions, "")
}

func getRouteAddPublicAttribution() string {
	return buildRoutePattern(httpPOST, routeAttributionsPublic, "")
}

func getRouteAttributionList() string {
	return buildRoutePattern(httpGET, routeAttributions, "")
}

func getRouteDeletePrivateAttribution() string {
	return buildRoutePattern(httpDELETE, routeAttributions, "uid")
}

func getRouteDeletePublicAttribution() string {
	return buildRoutePattern(httpDELETE, routeAttributionsPublic, "uid")
}

func getRouteDeleteAllPrivateAttributions() string {
	return buildRoutePattern(httpDELETE, routeAttributions, "")
}

func getRouteSearchAttributions() string {
	return buildRoutePattern(httpGET, routeAttributionsSearch, "query")
}

func getRouteAddAddressExclusions() string {
	return buildRoutePattern(httpPOST, routeExclusions, "")
}

func getRouteDeleteAddressExclusion() string {
	return buildRoutePattern(httpDELETE, routeExclusions, "hash")
}

func getRouteDeleteAllAddressExclusions() string {
	return buildRoutePattern(httpDELETE, routeExclusions, "")
}

func getRouteAddressExclusionList() string {
	return buildRoutePattern(httpGET, routeExclusions, "")
}

func getRouteAddressExclusionStatus() string {
	return buildRoutePattern(httpGET, routeExclusions, "hash")
}

func getRouteSpendingFingerprint() string {
	return buildRoutePattern(httpGET, routeSpendingFingerprint, "hash")
}

func getRouteWorkspaceAddNode() string {
	return buildRoutePattern(httpPOST, routeAddWorkspaceNode, "")
}

func getRouteWorkspaceDeleteNode() string {
	return buildRoutePattern(httpDELETE, routeAddWorkspaceNode, "")
}

func getRouteWorkspaces() string {
	return buildRoutePattern(httpGET, routeWorkspaces, "")
}

func getRouteAddWorkspace() string {
	return buildRoutePattern(httpPOST, routeWorkspaces, "name")
}

func getRouteGetWorkspace() string {
	return buildRoutePattern(httpGET, routeWorkspaces, "uid")
}

func getRouteUpdateWorkspace() string {
	return buildRoutePattern(httpPUT, routeWorkspaces, "")
}

func getRouteDeleteWorkspace() string {
	return buildRoutePattern(httpDELETE, routeWorkspaces, "uid")
}

func getRouteDeleteAllWorkspaces() string {
	return buildRoutePattern(httpDELETE, routeWorkspaces, "")
}

func getRouteMetrics() string {
	return httpGET + " " + routeMetrics
}
