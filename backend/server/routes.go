package server

const (
	routePrefix string = "/api/v1/"

	routeSearch                  string = "blockchain/search"
	routeTransaction             string = "blockchain/transactions"
	routeBlock                   string = "blockchain/blocks"
	routeAddress                 string = "blockchain/addresses"
	routeAddressOutputRange      string = "blockchain/outputs"
	routeMeta                    string = "meta"
	routeHeuristicReport         string = "heuristics/report"
	routeHeuristicDetails        string = "heuristicDetails"
	routeShortestTxPath          string = "shortestTransactionPath"
	routeConnectionLookup        string = "connectionLookup"
	routeMixingActivity          string = "mixingActivity"
	routeSpendingFingerprint     string = "spendingFingerprint"
	routeExclusions              string = "exclusions"
	routeClusters                string = "clusters"
	routeClustersHmi             string = "clusters/hmi"
	routeClustersReport          string = "clusters/report"
	routeAttributions            string = "attributions"
	routeAttributionsPublic      string = "attributions/public"
	routeAttributionsSearch      string = "attributions/search"
	routeWorkspaces              string = "workspaces"
	routeWorkspacesNodes         string = "workspaces/nodes"
	routeWorkspacesNode          string = "workspaces/node"
	routeAddWorkspaceNote        string = "workspaces/note"
	routeAddWorkspaceSelector    string = "workspaces/selector"
	routeWorkspaceSelectorStatus string = "workspaces/selector/status"
	routeWorkspacesConnection    string = "workspaces/connection"
	routeWorkspaceRename         string = "workspaces/rename"
	routeMetrics                 string = "/metrics"
)

// BuildPattern buils a route pattern which can be used with the stdlib http package
func BuildPattern(httpMethod string, r string, query string) string {
	base := httpMethod + " " + routePrefix + r + "/"

	if query != "" {
		base += "{" + query + "}"
	}

	return base
}
