package constants

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
	routeCreateUser                   string = "createUser"
	routeGetUsers                     string = "getUsers"
	routeDeleteUser                   string = "deleteUser"
	routeLogin                        string = "login"
	routeLogout                       string = "logout"
	routeModifyUser                   string = "modifyUser"
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
	routeMetrics                      string = "/metrics"
)

func getRoute(r string) string {
	return routePrefix + r + "/"
}

// GetRouteTransaction returns a route
func GetRouteTransaction() string {
	return getRoute(routeTransaction)
}

// GetRouteBlock returns a route
func GetRouteBlock() string {
	return getRoute(routeBlock)
}

// GetRouteAddress returns a route
func GetRouteAddress() string {
	return getRoute(routeAddress)
}

// GetRouteMeta returns a route
func GetRouteMeta() string {
	return getRoute(routeMeta)
}

// GetRouteHeuristics returns a route
func GetRouteHeuristics() string {
	return getRoute(routeHeuristics)
}

// GetRouteHeuristicsSummary returns a route
func GetRouteHeuristicsSummary() string {
	return getRoute(routeHeuristicsSummary)
}

// GetRouteHeuristicsExecution returns a route
func GetRouteHeuristicsExecution() string {
	return getRoute(routeHeuristicsExecution)
}

// GetRouteHeuristicDetails returns a route
func GetRouteHeuristicDetails() string {
	return getRoute(routeHeuristicDetails)
}

// GetRouteHeuristicStatus returns a route
func GetRouteHeuristicStatus() string {
	return getRoute(routeHeuristicStatus)
}

// GetRouteHeuristicList returns a route
func GetRouteHeuristicList() string {
	return getRoute(routeHeuristicList)
}

// GetRouteHeuristicDescriptors returns a route
func GetRouteHeuristicDescriptors() string {
	return getRoute(routeHeuristicDescriptors)
}

// GetRouteDeleteHeuristic returns a route
func GetRouteDeleteHeuristic() string {
	return getRoute(routeDeleteHeuristic)
}

// GetRouteSearch returns a route
func GetRouteSearch() string {
	return getRoute(routeSearch)
}

// GetRouteAddressOutputRange returns a route
func GetRouteAddressOutputRange() string {
	return getRoute(routeAddressOutputRange)
}

// GetRouteBlockRange returns a route
func GetRouteBlockRange() string {
	return getRoute(routeBlockRange)
}

// GetRouteCreateUser returns a route
func GetRouteCreateUser() string {
	return getRoute(routeCreateUser)
}

// GetRouteGetUsers returns a route
func GetRouteGetUsers() string {
	return getRoute(routeGetUsers)
}

// GetRouteDeleteUser returns a route
func GetRouteDeleteUser() string {
	return getRoute(routeDeleteUser)
}

// GetRouteLogin returns a route
func GetRouteLogin() string {
	return getRoute(routeLogin)
}

// GetRouteLogout returns a route
func GetRouteLogout() string {
	return getRoute(routeLogout)
}

// GetRouteModifyUser returns a route
func GetRouteModifyUser() string {
	return getRoute(routeModifyUser)
}

// GetRouteShortestTransactionPath returns a route
func GetRouteShortestTransactionPath() string {
	return getRoute(routeShortestTxPath)
}

// GetRouteConnectionLookup returns a route
func GetRouteConnectionLookup() string {
	return getRoute(routeConnectionLookup)
}

// GetRouteClusterLookup returns a route
func GetRouteClusterLookup() string {
	return getRoute(routeClusterLookup)
}

// GetRouteHMILookup returns a route
func GetRouteHMILookup() string {
	return getRoute(routeHMILookup)
}

// GetRouteClusterSummary returns a route
func GetRouteClusterSummary() string {
	return getRoute(routeClusterSummary)
}

// GetRouteMixingActivity returns a route
func GetRouteMixingActivity() string {
	return getRoute(routeMixingActivity)
}

// GetRouteAddCluster returns a route
func GetRouteAddCluster() string {
	return getRoute(routeAddCluster)
}

// GetRouteDeleteCluster returns a route
func GetRouteDeleteCluster() string {
	return getRoute(routeDeleteCluster)
}

// GetRouteDeleteAllClusters returns a route
func GetRouteDeleteAllClusters() string {
	return getRoute(routeDeleteAllClusters)
}

// GetRouteClusterOverview returns a route
func GetRouteClusterOverview() string {
	return getRoute(routeClusterOverview)
}

// GetRouteAddPrivateAttribution returns a route
func GetRouteAddPrivateAttribution() string {
	return getRoute(routeAddPrivateAttribution)
}

// GetRouteAddPublicAttribution returns a route
func GetRouteAddPublicAttribution() string {
	return getRoute(routeAddPublicAttribution)
}

// GetRouteAttributionOverview returns a route
func GetRouteAttributionOverview() string {
	return getRoute(routeAttributionOverview)
}

// GetRouteDeletePrivateAttribution returns a route
func GetRouteDeletePrivateAttribution() string {
	return getRoute(routeDeletePrivateAttribution)
}

// GetRouteDeletePublicAttribution returns a route
func GetRouteDeletePublicAttribution() string {
	return getRoute(routeDeletePublicAttribution)
}

// GetRouteDeleteAllPrivateAttributions returns a route
func GetRouteDeleteAllPrivateAttributions() string {
	return getRoute(routeDeleteAllPrivateAttributions)
}

// GetRouteSearchAttributions returns a route
func GetRouteSearchAttributions() string {
	return getRoute(routeSearchAttributions)
}

// GetRouteMetrics returns a route
func GetRouteMetrics() string {
	return routeMetrics
}
