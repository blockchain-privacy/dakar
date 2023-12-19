package server

const (
	routePrefix string = "/api/v1/"

	routeSearch                       string = "search"
	routeTransaction                  string = "tx"
	routeBlock                        string = "blk"
	routeAddress                      string = "address"
	routeMeta                         string = "meta"
	routeHeuristics                   string = "heuristics"
	routeHeuristicsSummary            string = "heuristicsSummary"
	routeHeuristicsExecution          string = "executeHeuristics"
	routeHeuristicDetails             string = "heuristicDetails"
	routeHeuristicStatus              string = "heuristicStatus"
	routeHeuristicList                string = "heuristicList"
	routeHeuristicDescriptors         string = "heuristicDescriptors"
	routeDeleteHeuristic              string = "deleteHeuristic"
	routeAddressOutputRange           string = "addressOutputRange"
	routeBlockRange                   string = "blkRange"
	routeCreateIdentity               string = "createIdentity"
	routeGetIdentities                string = "getIdentities"
	routeDeleteIdentity               string = "deleteIdentity"
	routeAdminDeleteIdentity          string = "adminDeleteIdentity"
	routeModifyUser                   string = "modifyUser"
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

func getRoute(r string) string {
	return routePrefix + r + "/"
}

// getRouteTransaction returns a route
func getRouteTransaction() string {
	return getRoute(routeTransaction)
}

// getRouteBlock returns a route
func getRouteBlock() string {
	return getRoute(routeBlock)
}

// getRouteAddress returns a route
func getRouteAddress() string {
	return getRoute(routeAddress)
}

// getRouteMeta returns a route
func getRouteMeta() string {
	return getRoute(routeMeta)
}

// getRouteHeuristics returns a route
func getRouteHeuristics() string {
	return getRoute(routeHeuristics)
}

// getRouteHeuristicsSummary returns a route
func getRouteHeuristicsSummary() string {
	return getRoute(routeHeuristicsSummary)
}

// getRouteHeuristicsExecution returns a route
func getRouteHeuristicsExecution() string {
	return getRoute(routeHeuristicsExecution)
}

// getRouteHeuristicDetails returns a route
func getRouteHeuristicDetails() string {
	return getRoute(routeHeuristicDetails)
}

// getRouteHeuristicStatus returns a route
func getRouteHeuristicStatus() string {
	return getRoute(routeHeuristicStatus)
}

// getRouteHeuristicList returns a route
func getRouteHeuristicList() string {
	return getRoute(routeHeuristicList)
}

// getRouteHeuristicDescriptors returns a route
func getRouteHeuristicDescriptors() string {
	return getRoute(routeHeuristicDescriptors)
}

// getRouteDeleteHeuristic returns a route
func getRouteDeleteHeuristic() string {
	return getRoute(routeDeleteHeuristic)
}

// getRouteSearch returns a route
func getRouteSearch() string {
	return getRoute(routeSearch)
}

// getRouteAddressOutputRange returns a route
func getRouteAddressOutputRange() string {
	return getRoute(routeAddressOutputRange)
}

// getRouteBlockRange returns a route
func getRouteBlockRange() string {
	return getRoute(routeBlockRange)
}

// getRouteCreateIdentity returns a route
func getRouteCreateIdentity() string {
	return getRoute(routeCreateIdentity)
}

// getRouteGetIdentities returns a route
func getRouteGetIdentities() string {
	return getRoute(routeGetIdentities)
}

// getRouteDeleteIdentity returns a route
func getRouteDeleteIdentity() string {
	return getRoute(routeDeleteIdentity)
}

// getRouteAdminDeleteIdentity returns a route
func getRouteAdminDeleteIdentity() string {
	return getRoute(routeAdminDeleteIdentity)
}

// getRouteModifyUser returns a route
func getRouteModifyUser() string {
	return getRoute(routeModifyUser)
}

// getRouteModifyIdentity returns a route
func getRouteModifyIdentity() string {
	return getRoute(routeModifyIdentity)
}

// getRouteShortestTransactionPath returns a route
func getRouteShortestTransactionPath() string {
	return getRoute(routeShortestTxPath)
}

// getRouteConnectionLookup returns a route
func getRouteConnectionLookup() string {
	return getRoute(routeConnectionLookup)
}

// getRouteClusterLookup returns a route
func getRouteClusterLookup() string {
	return getRoute(routeClusterLookup)
}

// getRouteHMILookup returns a route
func getRouteHMILookup() string {
	return getRoute(routeHMILookup)
}

// getRouteClusterSummary returns a route
func getRouteClusterSummary() string {
	return getRoute(routeClusterSummary)
}

// getRouteMixingActivity returns a route
func getRouteMixingActivity() string {
	return getRoute(routeMixingActivity)
}

// getRouteAddCluster returns a route
func getRouteAddCluster() string {
	return getRoute(routeAddCluster)
}

// getRouteDeleteCluster returns a route
func getRouteDeleteCluster() string {
	return getRoute(routeDeleteCluster)
}

// getRouteDeleteAllClusters returns a route
func getRouteDeleteAllClusters() string {
	return getRoute(routeDeleteAllClusters)
}

// getRouteClusterOverview returns a route
func getRouteClusterOverview() string {
	return getRoute(routeClusterOverview)
}

// getRouteAddPrivateAttribution returns a route
func getRouteAddPrivateAttribution() string {
	return getRoute(routeAddPrivateAttribution)
}

// getRouteAddPublicAttribution returns a route
func getRouteAddPublicAttribution() string {
	return getRoute(routeAddPublicAttribution)
}

// getRouteAttributionOverview returns a route
func getRouteAttributionOverview() string {
	return getRoute(routeAttributionOverview)
}

// getRouteDeletePrivateAttribution returns a route
func getRouteDeletePrivateAttribution() string {
	return getRoute(routeDeletePrivateAttribution)
}

// getRouteDeletePublicAttribution returns a route
func getRouteDeletePublicAttribution() string {
	return getRoute(routeDeletePublicAttribution)
}

// getRouteDeleteAllPrivateAttributions returns a route
func getRouteDeleteAllPrivateAttributions() string {
	return getRoute(routeDeleteAllPrivateAttributions)
}

// getRouteSearchAttributions returns a route
func getRouteSearchAttributions() string {
	return getRoute(routeSearchAttributions)
}

// getRouteAddAddressExclusions returns a route
func getRouteAddAddressExclusions() string {
	return getRoute(routeAddAddressExclusions)
}

// getRouteDeleteAddressExclusion returns a route
func getRouteDeleteAddressExclusion() string {
	return getRoute(routeDeleteAddressExclusion)
}

// getRouteDeleteAllAddressExclusions returns a route
func getRouteDeleteAllAddressExclusions() string {
	return getRoute(routeDeleteAllAddressExclusions)
}

// getRouteAddressExclusionOverview returns a route
func getRouteAddressExclusionOverview() string {
	return getRoute(routeAddressExclusionOverview)
}

// getRouteAddressExclusionStatus returns a route
func getRouteAddressExclusionStatus() string {
	return getRoute(routeAddressExclusionStatus)
}

// getRouteSpendingFingerprint returns a route
func getRouteSpendingFingerprint() string {
	return getRoute(routeSpendingFingerprint)
}

// getRouteWorkspaceAddNode returns a route
func getRouteWorkspaceAddNode() string {
	return getRoute(routeAddWorkspaceNode)
}

// getRouteWorkspaces returns a route
func getRouteWorkspaces() string {
	return getRoute(routeWorkspaces)
}

// getRouteAddWorkspace returns a route
func getRouteAddWorkspace() string {
	return getRoute(routeAddWorkspace)
}

// getRouteGetWorkspace returns a route
func getRouteGetWorkspace() string {
	return getRoute(routeGetWorkspace)
}

// getRouteGetWorkspace returns a route
func getRouteUpdateWorkspace() string {
	return getRoute(routeUpdateWorkspace)
}

// getRouteDeleteWorkspace returns a route
func getRouteDeleteWorkspace() string {
	return getRoute(routeDeleteWorkspace)
}

// getRouteMetrics returns a route
func getRouteMetrics() string {
	return routeMetrics
}
