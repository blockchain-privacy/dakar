package workspace

import (
	"backend/cmd/cliutil"
	"backend/db/workspace"
	"backend/external"
	"encoding/json"
	"time"
)

func EncodeAndStoreWorkspaceState(dgraph external.Database, userUID string,
	workspaceUID string, nodes any, timeStamp time.Time) error {
	stateBytes, err := json.Marshal(nodes)
	if err != nil {
		return cliutil.NewStackError(err)
	}

	err = workspace.SetWorkspaceState(dgraph, userUID, workspaceUID, string(stateBytes), timeStamp)
	if err != nil {
		return cliutil.NewStackError(err)
	}

	return nil
}

// InsertNodeConnections queries the db for connections between nodes in nodeMap and inserts them
func InsertNodeConnections(dgraph external.Database, nodeMap map[string]workspace.FrontendGraphNode) error {
	connections, err := workspace.GetWorkspaceConnections(dgraph, cliutil.GetMapKeys(nodeMap))
	if err != nil {
		return err
	}

	for _, node := range connections {
		nodeElement, ok := nodeMap[node.UID]
		if !ok {
			return cliutil.NewStackErrorf("uid %s not found in map", node.UID)
		}
		nodeElement.Children = node.Children
		nodeMap[node.UID] = nodeElement
	}

	return nil
}
