package constants

const (
	routePrefix string = "/api/v1/"

	routeSearch              string = "search"
	routeTransaction         string = "tx"
	routeBlock               string = "blk"
	routeAddress             string = "address"
	routeMeta                string = "meta"
	routeHeuristics          string = "heuristics"
	routeHeuristicsSummary   string = "heuristicsSummary"
	routeHeuristicsExecution string = "executeHeuristics"
	routeHeuristicDetails    string = "heuristicDetails"
	routeHeuristicStatus     string = "heuristicStatus"
	routeHeuristicList       string = "heuristicList"
	routeDeleteHeuristic     string = "deleteHeuristic"
	routeAddressOutputRange  string = "addressOutputRange"
	routeCreateUser          string = "createUser"
	routeGetUsers            string = "getUsers"
	routeDeleteUser          string = "deleteUser"
	routeLogin               string = "login"
	routeLogout              string = "logout"
	routeModifyUser          string = "modifyUser"
	routeShortestTxPath      string = "shortestTransactionPath"
)

func getRoute(r string) string {
	return routePrefix + r + "/"
}

func GetRouteTransaction() string {
	return getRoute(routeTransaction)
}

func GetRouteBlock() string {
	return getRoute(routeBlock)
}

func GetRouteAddress() string {
	return getRoute(routeAddress)
}

func GetRouteMeta() string {
	return getRoute(routeMeta)
}

func GetRouteHeuristics() string {
	return getRoute(routeHeuristics)
}

func GetRouteHeuristicsSummary() string {
	return getRoute(routeHeuristicsSummary)
}

func GetRouteHeuristicsExecution() string {
	return getRoute(routeHeuristicsExecution)
}

func GetRouteHeuristicDetails() string {
	return getRoute(routeHeuristicDetails)
}

func GetRouteHeuristicStatus() string {
	return getRoute(routeHeuristicStatus)
}

func GetRouteHeuristicList() string {
	return getRoute(routeHeuristicList)
}

func GetRouteDeleteHeuristic() string {
	return getRoute(routeDeleteHeuristic)
}

func GetRouteSearch() string {
	return getRoute(routeSearch)
}

func GetRouteAddressOutputRange() string {
	return getRoute(routeAddressOutputRange)
}

func GetRouteCreateUser() string {
	return getRoute(routeCreateUser)
}

func GetRouteGetUsers() string {
	return getRoute(routeGetUsers)
}

func GetRouteDeleteUser() string {
	return getRoute(routeDeleteUser)
}

func GetRouteLogin() string {
	return getRoute(routeLogin)
}

func GetRouteLogout() string {
	return getRoute(routeLogout)
}

func GetRouteModifyUser() string {
	return getRoute(routeModifyUser)
}

func GetRouteShortestTransactionPath() string {
	return getRoute(routeShortestTxPath)
}
