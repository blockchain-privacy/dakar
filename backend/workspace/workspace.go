package workspace

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/workspace"
	"backend/external"
	"context"
	"encoding/json"
	"errors"
	"github.com/qrest/gomisc/serror"
	"slices"
	"strconv"
	"strings"
	"time"
)

const MaxWorkspaceNameLength = 50

// GetAndRefreshWorkspace returns the specified workspace. If necessary the workspace contents will also be refreshed.
// This becomes necessary if connections become outdated, when new blocks are added to the blockchain.
func GetAndRefreshWorkspace(ctx context.Context, dgraph external.Database,
	workspaceMutex *Mutex, workspaceUID string, userUID string) (*workspace.FrontendWorkspace, error) {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(ctx, dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	// no updated needed because of dummy heuristics, but maybe because clusters are outdated
	isOutdated, err := isWorkspaceOutdated(dgraph, w)
	if err != nil {
		return nil, serror.New(err)
	}

	nodeMap, notes := separateNodes(w.Nodes)

	var clusterHeight int64
	if w.ClusterHeight != nil {
		clusterHeight = *w.ClusterHeight
	}

	// Don't consider destination transactions with heuristic connections, because if
	// it is the only node then the workspace can not be outdated. This is a different
	// behaviour as when inserting node connections when adding a new node, because there
	// the node connections are still unkown.
	if isOutdated && len(nodeMap) > 1 {
		clusterHeight, nodeMap, err = InsertNodeConnectionsAndHeuristics(dgraph, nodeMap, userUID, workspaceUID)
		if err != nil {
			return nil, err
		}

		w.Nodes = append(cliutil.GetMapValues(nodeMap), notes...)

		err = encodeAndStoreWorkspaceState(ctx, dgraph, userUID, workspaceUID, w.Nodes, &clusterHeight)
		if err != nil {
			return nil, err
		}
	}

	return w.ToFrontendWorkspace(), nil
}

// UpdateNodeCoordinates replaces the coordinates of the given workspace with the coordinates from state
func UpdateNodeCoordinates(ctx context.Context, dgraph external.Database, workspaceMutex *Mutex, workspaceUID string,
	userUID string, state []workspace.Node) error {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(ctx, dgraph, workspaceUID, userUID)
	if err != nil {
		return err
	}

	if len(w.Nodes) == 0 {
		return serror.FromStr("received update for empty workspace")
	}

	frontendState := make(map[string]workspace.Node, len(state))
	for _, n := range state {
		frontendState[n.UID] = n
	}

	for i, backendNode := range w.Nodes {
		if frontendNode, ok := frontendState[backendNode.UID]; ok {
			w.Nodes[i].X = frontendNode.X
			w.Nodes[i].Y = frontendNode.Y
		}
	}

	return encodeAndStoreWorkspaceState(ctx, dgraph, userUID, workspaceUID,
		w.Nodes, w.ClusterHeight)
}

func deleteNode(ctx context.Context, dgraph external.Database, node *workspace.Node, workspaceNodes []workspace.Node,
	userUID string, workspaceUID string) ([]string, error) {
	var deletedNodes []string
	if node.Type == workspace.NodeTypeSelector {
		nodeMap := make(map[string]workspace.Node, len(workspaceNodes))
		for _, n := range workspaceNodes {
			nodeMap[n.UID] = n
		}

		uids := workspace.FindDescendantSelectorUIDs(nodeMap, node.UID)

		// delete the actual selectors
		if err := workspace.DeleteUserSelectors(ctx, dgraph, uids, userUID, workspaceUID); err != nil {
			if errors.Is(err, db.ErrNoMutationHappened) {
				return nil, nil
			}

			return nil, err
		}

		deletedNodes = uids
	} else if node.IsDestination() {
		nodeMap := make(map[string]workspace.Node, len(workspaceNodes))
		for _, n := range workspaceNodes {
			nodeMap[n.UID] = n
		}

		// collect all heuristic UIDs
		var children []string
		for _, child := range node.Children {
			children = append(children, workspace.FindDescendantSelectorUIDs(nodeMap, child)...)
		}

		if len(children) > 0 {
			// delete the actual heuristics
			if err := workspace.DeleteUserSelectors(ctx, dgraph, children, userUID, workspaceUID); err != nil {
				if errors.Is(err, db.ErrNoMutationHappened) {
					return nil, nil
				}

				return nil, err
			}
		}

		deletedNodes = append(children, node.UID)
	} else {
		deletedNodes = []string{node.UID}
	}

	return deletedNodes, nil
}

// DeleteNodes removes nodes and all their dependent nodes from a workspace.
// Returns all node UIDs which have been deleted.
func DeleteNodes(ctx context.Context, dgraph external.Database, workspaceMutex *Mutex, workspaceUID string,
	userUID string, nodeUIDs []string) ([]string, error) {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(ctx, dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	if len(w.Nodes) == 0 {
		return nil, serror.FromFormat("node deletion request for empty workspace. workspace: %s", workspaceUID)
	}

	nodesToDelete := make(map[string]*workspace.Node, len(nodeUIDs))
	for _, clientNode := range nodeUIDs {
		found := false
		for _, n := range w.Nodes {
			if n.UID == clientNode {
				nodesToDelete[n.UID] = &n // #nosec G601, false positive as of go1.22
				found = true
				break
			}
		}

		if !found {
			return nil, serror.FromFormat(
				"node does not exist in workspace. workspace: %s, node: %s", workspaceUID, clientNode)
		}
	}

	deleteNodesMap := map[string]bool{}
	for _, n := range nodesToDelete {
		nodes, err := deleteNode(ctx, dgraph, n, w.Nodes, userUID, workspaceUID)
		if err != nil {
			return nil, err
		}

		for _, delNode := range nodes {
			deleteNodesMap[delNode] = true
			// while deleting the current node, other connected nodes might also be deleted.
			// Therefore, remove all these nodes from future iteration as they have been handled already
			delete(nodesToDelete, delNode)
		}
	}

	deletedNodes := cliutil.GetMapKeys(deleteNodesMap)

	// check if any notes need to be deleted
	deletedNodes = append(deletedNodes, findDisconnectedNotes(w.Nodes, deletedNodes)...)

	w.Nodes = workspace.DeleteNodes(w.Nodes, deletedNodes)

	if err := encodeAndStoreWorkspaceState(ctx, dgraph, userUID, workspaceUID,
		w.Nodes, w.ClusterHeight); err != nil {
		return nil, err
	}

	return deletedNodes, nil
}

// findDisconnectedNotes finds all notes which would be unconnected if the nodes in deletedNodes where deleted.
func findDisconnectedNotes(nodes []workspace.Node, deletedNodes []string) []string {
	var orphanNotes []string
	for _, n := range nodes {
		// note can only have one child
		if n.Type == workspace.NodeTypeNote && len(n.Children) > 0 && slices.Contains(deletedNodes, n.Children[0]) {
			orphanNotes = append(orphanNotes, n.UID)
		}
	}

	return orphanNotes
}

// AddNodes adds a node to a workspace and refreshes the connections between all nodes.
func AddNodes(ctx context.Context, dgraph external.Database, workspaceMutex *Mutex, workspaceUID string,
	userUID string, newNodes []*workspace.Node) ([]workspace.Node, error) {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(ctx, dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	nodeMap, notes := separateNodes(w.Nodes)

	// If the transmitted state is empty, then there are only connections between the new nodes.
	// If newNodes is a destination transaction, it might be connected to heuristics.
	if len(nodeMap) == 0 && len(newNodes) == 1 && !newNodes[0].IsDestination() {
		frontEndNodes := []workspace.Node{*newNodes[0]}
		if err := encodeAndStoreWorkspaceState(ctx, dgraph, userUID, workspaceUID,
			frontEndNodes, w.ClusterHeight); err != nil {
			return nil, err
		}

		return frontEndNodes, nil
	}

	// check if all new nodes already exist in the workspace
	receivedNewNode := false
	for _, newNode := range newNodes {
		if _, ok := nodeMap[newNode.UID]; !ok {
			receivedNewNode = true
			break
		}
	}

	if !receivedNewNode {
		// all new nodes are already in current state, therefore there is nothing to do
		return nil, nil
	}

	for _, newNode := range newNodes {
		nodeMap[newNode.UID] = *newNode
	}

	clusterHeight, nodeMap, err := InsertNodeConnectionsAndHeuristics(dgraph, nodeMap, userUID, workspaceUID)
	if err != nil {
		return nil, err
	}

	frontEndNodes := slices.Concat(cliutil.GetMapValues(nodeMap), notes)

	if err := encodeAndStoreWorkspaceState(ctx, dgraph, userUID, workspaceUID, frontEndNodes, &clusterHeight); err != nil {
		return nil, err
	}

	return frontEndNodes, nil
}

const noteUIDPrefix = "note_"

// Generates a note uid using unix time. The note uid only has to be unique between the notes of a workspace.
// Therefore it is fine to not use cryptographically secure functions.
func generateNoteUID() string {
	return noteUIDPrefix + strconv.FormatInt(time.Now().Unix(), 10)
}

// AddNote adds a note to a workspace
func AddNote(ctx context.Context, dgraph external.Database, workspaceMutex *Mutex, workspaceUID string,
	userUID string, note workspace.Node) ([]workspace.Node, error) {
	if len(note.Children) == 0 {
		return nil, serror.FromStr("note has no children")
	}

	if note.UID != "" && !strings.HasPrefix(note.UID, noteUIDPrefix) {
		return nil, serror.FromFormat("invalid note uid: %s", note.UID)
	}

	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(ctx, dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	nodeMap, notes := separateNodes(w.Nodes)

	if len(nodeMap) == 0 {
		return nil, serror.FromStr("trying to add note to an empty workspace")
	}

	if _, ok := nodeMap[note.Children[0]]; !ok {
		// parent does not exist
		return nil, serror.FromFormat("trying to add note with non-existing child %s", note.Children[0])
	}

	// if it is a new node generate a uid
	if note.UID == "" {
		note.UID = generateNoteUID()
	} else {
		// remove existing note
		notes = slices.DeleteFunc(notes, func(node workspace.Node) bool {
			return node.UID == note.UID
		})
	}

	note.Type = workspace.NodeTypeNote

	notes = append(notes, note)
	frontEndNodes := append(cliutil.GetMapValues(nodeMap), notes...)

	if err := encodeAndStoreWorkspaceState(ctx, dgraph, userUID, workspaceUID, frontEndNodes, nil); err != nil {
		return nil, err
	}

	return frontEndNodes, nil
}

// encodeAndStoreWorkspaceState transforms the workspace state into JSON and stores it in the database
func encodeAndStoreWorkspaceState(ctx context.Context, dgraph external.Database, userUID string, workspaceUID string,
	state []workspace.Node, clusterHeight *int64) error {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return serror.New(err)
	}

	return workspace.SetWorkspaceState(ctx, dgraph, userUID, workspaceUID, string(stateBytes), clusterHeight)
}

// InsertNodeConnectionsAndHeuristics queries the db for connections between nodes in nodeMap and inserts them.
// Also inserts found heuristics into the node map
func InsertNodeConnectionsAndHeuristics(dgraph external.Database, nodeMap map[string]workspace.Node,
	userUID string, workspaceUID string) (int64, map[string]workspace.Node, error) {
	connections, nodeSelectors, clusterHeight, err := workspace.GetWorkspaceConnections(dgraph,
		cliutil.GetMapKeys(nodeMap), userUID, workspaceUID)
	if err != nil {
		return 0, nil, err
	}

	newNodeMap := make(map[string]workspace.Node)
	for _, node := range connections {
		nodeElement, ok := nodeMap[node.UID]
		if !ok {
			return 0, nil, serror.FromFormat("uid %s not found in map", node.UID)
		}
		nodeElement.Children = node.Children

		// need to add nodes into new map, otherwise duplicate addresses will be left in the map
		newNodeMap[node.UID] = nodeElement
	}

	// add selector nodes to map
	for _, s := range nodeSelectors {
		// set coordinates
		if oldSelector, ok := nodeMap[s.UID]; ok {
			s.X = oldSelector.X
			s.Y = oldSelector.Y
		}

		newNodeMap[s.UID] = s
	}

	return clusterHeight, newNodeMap, nil
}

// isWorkspaceOutdated returns true if the workspace state is outdated
func isWorkspaceOutdated(dgraph external.Database,
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

	var clusterUIDs []string //nolint:prealloc
	for _, n := range w.Nodes {
		if n.Type != workspace.NodeTypeCluster {
			continue
		}
		clusterUIDs = append(clusterUIDs, n.UID)
	}

	// workspace does not contain a cluster, therefore it can not be outdated
	if len(clusterUIDs) == 0 {
		return false, nil
	}

	isOutdated, err := workspace.IsWorkspaceStateOutdated(dgraph, *w.ClusterHeight, clusterUIDs)
	if err != nil {
		return false, err
	}
	return isOutdated, nil
}

// separateNodes separates nodes:
// - general node: transactions, heuristics, selectors and clusters
// - notes node: notes
func separateNodes(nodes []workspace.Node) (map[string]workspace.Node, []workspace.Node) {
	nodeMap := map[string]workspace.Node{}
	var notes []workspace.Node
	for _, n := range nodes {
		if n.UID == "" || n.Type == "" {
			continue
		}

		if n.Type == workspace.NodeTypeNote {
			notes = append(notes, n)
		} else {
			nodeMap[n.UID] = n
		}
	}
	return nodeMap, notes
}
