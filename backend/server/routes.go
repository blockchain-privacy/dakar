// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

const (
	routePrefix string = "/api/v1/"

	routeSearch                   string = "blockchain/search"
	routeTransaction              string = "blockchain/transactions"
	routeBlock                    string = "blockchain/blocks"
	routeAddress                  string = "blockchain/addresses"
	routeAddressOutputRange       string = "blockchain/outputs"
	routeMeta                     string = "meta"
	routeShortestTxPath           string = "shortestTransactionPath"
	routeMixingActivity           string = "mixingActivity"
	routeSpendingFingerprint      string = "spendingFingerprint"
	routeExclusions               string = "exclusions"
	routeClusters                 string = "clusters"
	routeClustersHmi              string = "clusters/hmi"
	routeClustersReport           string = "clusters/report"
	routeAttributions             string = "attributions"
	routeAttributionsPublic       string = "attributions/public"
	routeAttributionsSearch       string = "attributions/search"
	routeWorkspaces               string = "workspaces"
	routeWorkspacesNodes          string = "workspaces/nodes"
	routeWorkspacesNode           string = "workspaces/node"
	routeWorkspacesState          string = "workspaces/state"
	routeAddWorkspaceNote         string = "workspaces/note"
	routeAddWorkspaceSelector     string = "workspaces/selector"
	routeWorkspaceSelectorStatus  string = "workspaces/selector/status"
	routeWorkspaceSelectorResults string = "workspaces/selector/results"
	routeWorkspaceSelectorReport  string = "workspaces/selector/report"
	routeWorkspacesConnection     string = "workspaces/connection"
	routeWorkspaceRename          string = "workspaces/rename"
	routeMetrics                  string = "/metrics"
)

// BuildPattern builds a route pattern which can be used with the stdlib http package
func BuildPattern(httpMethod string, r string, query string) string {
	base := httpMethod + " " + routePrefix + r + "/"

	if query != "" {
		base += "{" + query + "}"
	}

	return base
}
