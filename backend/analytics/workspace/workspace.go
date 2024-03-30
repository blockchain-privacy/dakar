package workspace

import (
	"backend/cmd/cliutil"
	"backend/db/workspace"
	"backend/external"
	"encoding/json"
)

func EncodeAndStoreWorkspaceState(dgraph external.Database, userUID string, workspaceUID string,
	state []workspace.Node, clusterHeight *int64) error {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return cliutil.NewStackError(err)
	}

	return workspace.SetWorkspaceState(dgraph, userUID, workspaceUID, string(stateBytes), clusterHeight)
}

// InsertNodeConnectionsAndHeuristics queries the db for connections between nodes in nodeMap and inserts them
func InsertNodeConnectionsAndHeuristics(dgraph external.Database, nodeMap map[string]workspace.Node,
	heuristicMap map[string]workspace.Node, userUID string, workspaceUID string) (int64, error) {
	connections, heuristics, clusterHeight, err := workspace.GetWorkspaceConnections(dgraph,
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
	for _, h := range heuristics {
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
