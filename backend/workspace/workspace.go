package workspace

import (
	"backend/cmd/cliutil"
	dbHeuristic "backend/db/analytics/heuristics"
	"backend/db/workspace"
	"backend/external"
	"backend/worker"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
)

// DeleteNode removes a node and all its dependent node from a workspace.
// Returns all node UIDs which have been deleted.
func DeleteNode(dgraph external.Database, workspaceMutex *Mutex, workspaceUID string,
	userUID string, nodeUID string) ([]string, error) {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	if len(w.Nodes) == 0 {
		return nil, cliutil.NewStackErrorf(
			"node deletion request for empty workspace. workspace: %s, node: %s", workspaceUID, nodeUID)
	}

	var deletedNode *workspace.Node
	for _, n := range w.Nodes {
		if n.UID == nodeUID {
			deletedNode = &n // #nosec G601, false positive as of go1.22
			break
		}
	}

	if deletedNode == nil {
		return nil, cliutil.NewStackErrorf(
			"node does not exist in workspace. workspace: %s, node: %s", workspaceUID, nodeUID)
	}

	var deletedNodes []string

	if deletedNode.Type == workspace.NodeTypeHeuristic {
		nodeMap := make(map[string]workspace.Node, len(w.Nodes))
		for _, n := range w.Nodes {
			nodeMap[n.UID] = n
		}

		uids := workspace.FindDescendantHeuristicUIDs(nodeMap, deletedNode.UID)

		// delete the actual heuristics
		if err := dbHeuristic.DeleteUserHeuristics(dgraph, uids, userUID, workspaceUID); err != nil {
			if errors.Is(err, dbHeuristic.ErrNoMutationHappened) {
				return nil, nil
			} else {
				return nil, err
			}
		}

		// remove heuristics from nodes
		w.Nodes = workspace.DeleteNodes(w.Nodes, uids)
		deletedNodes = uids
	} else if deletedNode.IsDestination() {
		nodeMap := make(map[string]workspace.Node, len(w.Nodes))
		for _, n := range w.Nodes {
			nodeMap[n.UID] = n
		}

		// collect all heuristic UIDs
		var children []string
		for _, child := range deletedNode.Children {
			children = append(children, workspace.FindDescendantHeuristicUIDs(nodeMap, child)...)
		}

		if len(children) > 0 {
			// delete the actual heuristics
			if err := dbHeuristic.DeleteUserHeuristics(dgraph, children, userUID, workspaceUID); err != nil {
				if errors.Is(err, dbHeuristic.ErrNoMutationHappened) {
					return nil, nil
				} else {
					return nil, err
				}
			}
		}

		deletedNodes = append(children, deletedNode.UID)
		w.Nodes = workspace.DeleteNodes(w.Nodes, deletedNodes)
	} else {
		w.Nodes = workspace.DeleteNodes(w.Nodes, []string{deletedNode.UID})
		deletedNodes = []string{deletedNode.UID}
	}

	if err := EncodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID,
		w.Nodes, w.ClusterHeight); err != nil {
		return nil, err
	}

	return deletedNodes, nil
}

// AddNode adds a node to a workspace and refreshes the connections between all nodes.
func AddNode(dgraph external.Database, workspaceMutex *Mutex, worker *worker.Worker, workspaceUID string,
	userUID string, newNode *workspace.Node) ([]workspace.Node, error) {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	nodeMap, heuristicMap, dummyHeuristics := SplitNodesIntoCategories(w.Nodes)

	// If the transmitted state is empty, then there are only connections between the new nodes.
	// If newNodes is a destination transaction, it might be connected to heuristics.
	if len(nodeMap) == 0 && !newNode.IsDestination() {
		frontEndNodes := []workspace.Node{*newNode}
		if err := EncodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID,
			frontEndNodes, w.ClusterHeight); err != nil {
			return nil, err
		}

		return frontEndNodes, nil
	}

	if _, ok := nodeMap[newNode.UID]; ok {
		// new node is already in current state, therefore there is nothing to do
		return nil, nil
	}

	nodeMap[newNode.UID] = *newNode
	clusterHeight, nodeMap, err := InsertNodeConnectionsAndHeuristics(dgraph, nodeMap, heuristicMap, userUID, workspaceUID)
	if err != nil {
		return nil, err
	}

	_, dummyHeuristics = FilterDummyNodes(worker, dummyHeuristics, userUID)

	frontEndNodes := append(cliutil.GetMapValues(nodeMap), dummyHeuristics...)

	if err := EncodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID, frontEndNodes, &clusterHeight); err != nil {
		return nil, err
	}

	return frontEndNodes, nil
}
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
	heuristicMap map[string]workspace.Node, userUID string, workspaceUID string) (int64,
	map[string]workspace.Node, error) {
	connections, nodeHeuristics, clusterHeight, err := workspace.GetWorkspaceConnections(dgraph,
		cliutil.GetMapKeys(nodeMap), userUID, workspaceUID)
	if err != nil {
		return 0, nil, err
	}

	newNodeMap := make(map[string]workspace.Node)
	for _, node := range connections {
		nodeElement, ok := nodeMap[node.UID]
		if !ok {
			return 0, nil, cliutil.NewStackErrorf("uid %s not found in map", node.UID)
		}
		nodeElement.Children = node.Children

		// need to add nodes into new map, otherwise duplicate addresses will be left in the map
		newNodeMap[node.UID] = nodeElement
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

	return clusterHeight, newNodeMap, nil
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
func FilterDummyNodes(worker *worker.Worker, dummyHeuristics []workspace.Node,
	userID string) (changedDummies bool, filteredDummies []workspace.Node) {
	filteredDummies = slices.DeleteFunc(dummyHeuristics, func(node workspace.Node) bool {
		workID, err := strconv.Atoi(node.UID)
		if err != nil {
			// if work ID is invalid delete it as it is useless
			return true
		}

		uid, err := worker.GetFinishedHeuristicUID(workID, userID)
		if err != nil {
			// workID can not be retrieved, so remove it
			return true
		}

		// if the heuristic has finished executing, then remove it from
		// the dummy list
		if uid != "" {
			return true
		}

		// remove all dummy heuristics for which no work exists.
		// This can happen if the server was restarted before an heuristic was executed.
		if !worker.DoesWorkExist(workID, userID) {
			return true
		}

		return false
	})
	changedDummies = len(dummyHeuristics) != len(filteredDummies)

	return
}
