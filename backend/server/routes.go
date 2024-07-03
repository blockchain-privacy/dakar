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
	routeShortestTxPath       string = "shortestTransactionPath"
	routeConnectionLookup     string = "connectionLookup"
	routeMixingActivity       string = "mixingActivity"
	routeSpendingFingerprint  string = "spendingFingerprint"
	routeUsers                string = "users"
	routeExclusions           string = "exclusions"
	routeClusters             string = "clusters"
	routeClustersHmi          string = "clusters/hmi"
	routeClustersReport       string = "clusters/report"
	routeAttributions         string = "attributions"
	routeAttributionsPublic   string = "attributions/public"
	routeAttributionsSearch   string = "attributions/search"
	routeWorkspaces           string = "workspaces"
	routeWorkspacesNodes      string = "workspaces/nodes"
	routeWorkspacesNode       string = "workspaces/node"
	routeAddWorkspaceNote     string = "workspaces/note"
	routeWorkspacesConnection string = "workspaces/connection"
	routeWorkspaceRename      string = "workspaces/rename"
	routeMetrics              string = "/metrics"
)

// buildPattern buils a route pattern which can be used with the stdlib http package
func buildPattern(httpMethod string, r string, query string) string {
	base := httpMethod + " " + routePrefix + r + "/"

	if query != "" {
		base += "{" + query + "}"
	}

	return base
}
