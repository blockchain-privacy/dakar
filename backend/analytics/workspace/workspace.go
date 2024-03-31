package workspace

import (
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/db/workspace"
	"backend/external"
	"encoding/json"
	"log/slog"
	"slices"
	"strconv"
)

func EncodeAndStoreWorkspaceState(dgraph external.Database, userUID string, workspaceUID string,
	state []workspace.Node, clusterHeight *int64) error {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return cliutil.NewStackError(err)
	}

	return workspace.SetWorkspaceState(dgraph, userUID, workspaceUID, string(stateBytes), clusterHeight)
}

// InsertNodeConnectionsAndHeuristics queries the db for connections between nodes in nodeMap and inserts them.
// Also inserts found heuristics into the node map
func InsertNodeConnectionsAndHeuristics(dgraph external.Database, nodeMap map[string]workspace.Node,
	heuristicMap map[string]workspace.Node, userUID string, workspaceUID string) (int64, error) {
	connections, nodeHeuristics, clusterHeight, err := workspace.GetWorkspaceConnections(dgraph,
		cliutil.GetMapKeys(nodeMap), userUID, workspaceUID)
	if err != nil {
		return 0, err
	}

	for _, node := range connections {
		nodeElement, ok := nodeMap[node.UID]
		if !ok {
			return 0, cliutil.NewStackErrorf("uid %s not found in map", node.UID)
		}
		nodeElement.Children = node.Children
		nodeMap[node.UID] = nodeElement
	}

	// add heuristic nodes to map
	for _, h := range nodeHeuristics {
		// set coordinates
		if oldHeuristic, ok := heuristicMap[h.UID]; ok {
			h.X = oldHeuristic.X
			h.Y = oldHeuristic.Y
		}

		nodeMap[h.UID] = h
	}

	return clusterHeight, nil
}

// IsWorkspaceOutdated returns true if the workspace state is outdated
func IsWorkspaceOutdated(dgraph external.Database,
	w *workspace.DecodedWorkspace) (bool, error) {
	if len(w.Nodes) == 0 {
		return false, nil
	}

	// no timestamp set, therefore it is unkown if the state is outdated -> respond with outdated
	// or timestamp set to zero, therefore an update is necessary
	if w.ClusterHeight == nil || *w.ClusterHeight == 0 {
		return true, nil
	}

	// only one node with an non-zero cluster height -> not oudated
	if len(w.Nodes) == 1 {
		return false, nil
	}

	containsCluster := false
	uids := make([]string, len(w.Nodes))
	for i, n := range w.Nodes {
		if n.Type == workspace.NodeTypeCluster {
			containsCluster = true
		}
		uids[i] = n.UID
	}

	// workspace does not contain a cluster, therefore it can not be outdated
	if !containsCluster {
		return false, nil
	}

	isOutdated, err := workspace.IsWorkspaceStateOutdated(dgraph, *w.ClusterHeight, uids)
	if err != nil {
		return false, err
	}
	return isOutdated, nil
}

// SplitNodesIntoCategories categorizes each node into its own map:
// - general node: transactions and clusters
// - heuristic node: executed heuristics
// - dummy heuristic node: heuristics waiting to be executed
func SplitNodesIntoCategories(nodes []workspace.Node) (map[string]workspace.Node, map[string]workspace.Node, []workspace.Node) {
	nodeMap := map[string]workspace.Node{}
	// save heuristics in separate map, as they are transient. Stored heuristics are used only for coordinates
	heuristicMap := map[string]workspace.Node{}
	var dummyHeuristicMap []workspace.Node

	for _, n := range nodes {
		if n.UID == "" || n.Type == "" {
			continue
		}

		if n.IsLoading() {
			dummyHeuristicMap = append(dummyHeuristicMap, n)
		} else if n.Type == workspace.NodeTypeHeuristic {
			heuristicMap[n.UID] = n
		} else {
			nodeMap[n.UID] = n
		}
	}
	return nodeMap, heuristicMap, dummyHeuristicMap
}

// FilterDummyNodes filters the given nodes based on wether they exist in the worker work log and/or
// if they are finished executing. It returns the filtered list and two flags:
// - removedNode: true if at least one node was removed
// - removedFinishedNode: true if at least one was removed because it was finished executing
func FilterDummyNodes(worker *heuristics.Worker, dummyHeuristics []workspace.Node, userID string) (
	removedNode bool, removedFinishedNode bool, filteredDummies []workspace.Node, errs []error) {
	filteredDummies = slices.DeleteFunc(dummyHeuristics, func(node workspace.Node) bool {
		workID, err := strconv.Atoi(node.UID)
		if err != nil {
			removedNode = true
			errs = append(errs, cliutil.NewStackErrorf("could not convert %s to an integer: %w", node.UID, err))
			// if work ID is invalid delete it as it is useless
			return true
		}

		uid, err := worker.GetFinishedHeuristicUID(workID, userID)
		if err != nil {
			removedNode = true
			errs = append(errs, err)
			return true
		}

		// if the heuristic has finished executing, then remove it from
		// the dummy list
		if uid != "" {
			removedNode = true
			removedFinishedNode = true
			return true
		}

		// remove all dummy heuristics for which no work exists.
		// This can happen if the server was restarted before an heuristic was executed.
		if !worker.DoesWorkExist(workID, userID) {
			removedNode = true
			return true
		}

		return false
	})

	slog.Info("dummy heuristics length", "len(dummyHeuristics)", len(dummyHeuristics), "len(filteredDummies)", len(filteredDummies))

	return
}
