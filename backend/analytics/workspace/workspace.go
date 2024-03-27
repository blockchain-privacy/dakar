package workspace

import (
	"backend/cmd/cliutil"
	"backend/db/workspace"
	"backend/external"
	"encoding/json"
	"time"
)

func EncodeAndStoreWorkspaceState(dgraph external.Database, userUID string,
	workspaceUID string, state workspace.State, timeStamp time.Time) error {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return cliutil.NewStackError(err)
	}

	return workspace.SetWorkspaceState(dgraph, userUID, workspaceUID, string(stateBytes), timeStamp)
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
