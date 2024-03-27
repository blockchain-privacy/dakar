package workspace

import (
	"backend/cmd/cliutil"
	"backend/db/workspace"
	"backend/external"
	"encoding/json"
	"time"
)

func EncodeAndStoreWorkspaceState(dgraph external.Database, userUID string,
	workspaceUID string, state []workspace.FrontendGraphNode, timeStamp time.Time, clusterHeight *int64) error {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return cliutil.NewStackError(err)
	}

	return workspace.SetWorkspaceState(dgraph, userUID, workspaceUID, string(stateBytes), timeStamp, clusterHeight)
}

// InsertNodeConnectionsAndHeuristics queries the db for connections between nodes in nodeMap and inserts them
func InsertNodeConnectionsAndHeuristics(dgraph external.Database, nodeMap map[string]workspace.FrontendGraphNode,
	heuristicMap map[string]workspace.FrontendGraphNode, userUID string) (int64, error) {
	connections, heuristics, clusterHeight, err := workspace.GetWorkspaceConnections(dgraph, cliutil.GetMapKeys(nodeMap), userUID)
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
	if len(w.Nodes) > 0 {
		// no timestamp set, therefore we can not know if it is outdated -> respond with outdated
		if w.ClusterHeight == nil {
			return true, nil
		}

		// not outdated as there are no nodes or only one node
		if len(w.Nodes) < 2 {
			return false, nil
		}

		containsCluster := false
		uids := make([]string, len(w.Nodes))
		for i, n := range w.Nodes {
			if n.Type == "cluster" {
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

	return false, nil
}
