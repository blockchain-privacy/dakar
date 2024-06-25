package workspace

import (
	"backend/analytics/graph"
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	dbHeuristic "backend/db/analytics/heuristics"
	"backend/db/workspace"
	"backend/external"
	"backend/worker"
	"encoding/json"
	"errors"
	"github.com/qrest/gomisc/serror"
	"slices"
	"strconv"
	"strings"
	"time"
)

const MaxWorkspaceNameLength = 50

type HeuristicWork struct {
	executor       heuristics.Executor
	workspaceMutex *Mutex
	workspaceUID   string
	userUID        string
}

// Run processes the heuristic and inserts it into the workspace
func (h HeuristicWork) Run(dgraph external.Database, g *graph.Wrapper, workID int) (string, error) {
	newHeuristic, err := h.executor.Run(dgraph, g)
	if err != nil {
		return "", err
	}

	lock := h.workspaceMutex.Lock(h.workspaceUID)
	defer lock.Unlock()

	newHeuristicUID, err := dbHeuristic.InsertHeuristic(dgraph, newHeuristic, h.userUID, h.workspaceUID)
	if err != nil {
		return "", err
	}

	// update workspace

	w, err := workspace.GetFrontendWorkspace(dgraph, h.workspaceUID, h.userUID)
	if err != nil {
		return "", err
	}

	nodeMap, dummyHeuristics, notes := separateNodes(w.Nodes)

	clusterHeight, nodeMap, err := InsertNodeConnectionsAndHeuristics(dgraph, nodeMap, h.userUID, h.workspaceUID)
	if err != nil {
		return "", err
	}

	workIDString := strconv.Itoa(workID)

	dummyHeuristics = slices.DeleteFunc(dummyHeuristics, func(node workspace.Node) bool {
		if node.UID == workIDString {
			// try to set node position
			if newNode, ok := nodeMap[newHeuristicUID]; ok {
				newNode.X = node.X
				newNode.Y = node.Y
				nodeMap[newHeuristicUID] = newNode
			}

			return true
		}

		return false
	})

	frontEndNodes := slices.Concat(cliutil.GetMapValues(nodeMap), dummyHeuristics, notes)

	if err := encodeAndStoreWorkspaceState(dgraph, h.userUID, h.workspaceUID, frontEndNodes, &clusterHeight); err != nil {
		return "", err
	}

	return newHeuristicUID, nil
}

// CreateWork creates a new work package, which can be run at a later time
func CreateWork(newHeuristic dbHeuristic.DatabaseHeuristicRequest, workspaceUID string,
	userUID string, workspaceMutex *Mutex) (worker.Work, error) {
	if workspaceUID == "" {
		return nil, serror.NewStackErrorStr("workspace UID not set for heuristic request")
	}

	executor, err := heuristics.ConstructExecutors(newHeuristic, userUID)
	if err != nil {
		return nil, err
	}

	return HeuristicWork{
		executor:       executor,
		workspaceMutex: workspaceMutex,
		workspaceUID:   workspaceUID,
		userUID:        userUID,
	}, err
}

// AddHeuristic adds a new heuristic to the workspace. It returns a work ID,
// which can be used to check the execution status of the heuristic.
func AddHeuristic(dgraph external.Database, worker *worker.Worker, workspaceMutex *Mutex,
	heuristicRequest dbHeuristic.DatabaseHeuristicRequest, workspaceUID string, userUID string) (string, error) {
	work, err := CreateWork(heuristicRequest, workspaceUID, userUID, workspaceMutex)
	if err != nil {
		return "", err
	}

	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(dgraph, workspaceUID, userUID)
	if err != nil {
		return "", err
	}

	// sanity check
	if len(w.Nodes) == 0 {
		return "", serror.NewStackErrorStr("received update for empty workspace")
	}

	// find the index of the hew heuristic's parent
	parentIndex := -1
	if heuristicRequest.ParentHeuristicUID == "" {
		for i, n := range w.Nodes {
			if n.TransactionHash == heuristicRequest.TransactionHash {
				parentIndex = i
				break
			}
		}
	} else {
		for i, n := range w.Nodes {
			if n.UID == heuristicRequest.ParentHeuristicUID {
				parentIndex = i
				break
			}
		}
	}

	// no parent found
	if parentIndex == -1 {
		return "", serror.NewStackErrorStr("could not determine parent for new heuristic")
	}

	clusterTypes := make([]string, len(heuristicRequest.ClusterTypes))
	for i, c := range heuristicRequest.ClusterTypes {
		clusterTypes[i] = string(c)
	}

	yes := true
	workID := strconv.Itoa(worker.AddWork(userUID, work))

	// add new heuristic uid to children of parent
	w.Nodes[parentIndex].Children = append(w.Nodes[parentIndex].Children, workID)
	// add node
	w.Nodes = append(w.Nodes, workspace.Node{
		UID:                 workID,
		Type:                workspace.NodeTypeHeuristic,
		HeuristicType:       heuristicRequest.Type,
		Parameter:           heuristicRequest.Parameter,
		ExcludeAddresses:    &heuristicRequest.ExcludeAddresses,
		ExcludeSpendingGaps: &heuristicRequest.ExcludeSpendingGaps,
		ClusterTypes:        clusterTypes,
		Loading:             &yes,
	})

	if err = encodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID, w.Nodes, w.ClusterHeight); err != nil {
		return "", err
	}

	return workID, nil
}

// GetAndRefreshWorkspace returns the specified workspace. If necessary the workspace contents will also be refreshed.
// This becomes necessary if connections become outdated, when new blocks are added to the blockchain.
func GetAndRefreshWorkspace(dgraph external.Database, worker *worker.Worker, workspaceMutex *Mutex, workspaceUID string,
	userUID string) (*workspace.FrontendWorkspace, error) {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	// no updated needed because of dummy heuristics, but maybe because clusters are outdated
	isOutdated, err := isWorkspaceOutdated(dgraph, w)
	if err != nil {
		return nil, serror.NewStackError(err)
	}

	nodeMap, dummyHeuristics, notes := separateNodes(w.Nodes)

	var clusterHeight int64
	if w.ClusterHeight != nil {
		clusterHeight = *w.ClusterHeight
	}
	var updatedConnections bool
	// Don't consider destination transactions with heuristic connections, because if
	// it is the only node then the workspace can not be outdated. This is a different
	// behaviour as when inserting node connections when adding a new node, because there
	// the node connections are still unkown.
	if isOutdated && len(nodeMap) > 1 {
		clusterHeight, nodeMap, err = InsertNodeConnectionsAndHeuristics(dgraph, nodeMap, userUID, workspaceUID)
		if err != nil {
			return nil, err
		}

		updatedConnections = true
	}

	needToStore, dummyHeuristics := filterDummyNodes(worker, dummyHeuristics, userUID)

	if updatedConnections || needToStore {
		w.Nodes = slices.Concat(cliutil.GetMapValues(nodeMap), dummyHeuristics, notes)

		err = encodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID, w.Nodes, &clusterHeight)
		if err != nil {
			return nil, err
		}
	}

	return w.ToFrontendWorkspace(), nil
}

// UpdateNodeCoordinates replaces the coordinates of the given workspace with the coordinates from state
func UpdateNodeCoordinates(dgraph external.Database, workspaceMutex *Mutex, workspaceUID string,
	userUID string, state []workspace.Node) error {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(dgraph, workspaceUID, userUID)
	if err != nil {
		return err
	}

	if len(w.Nodes) == 0 {
		return serror.NewStackErrorStr("received update for empty workspace")
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

	return encodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID,
		w.Nodes, w.ClusterHeight)
}

func deleteNode(dgraph external.Database, node *workspace.Node, workspaceNodes []workspace.Node,
	userUID string, workspaceUID string) ([]string, error) {
	var deletedNodes []string
	if node.Type == workspace.NodeTypeHeuristic {
		nodeMap := make(map[string]workspace.Node, len(workspaceNodes))
		for _, n := range workspaceNodes {
			nodeMap[n.UID] = n
		}

		uids := workspace.FindDescendantHeuristicUIDs(nodeMap, node.UID)

		// delete the actual heuristics
		if err := dbHeuristic.DeleteUserHeuristics(dgraph, uids, userUID, workspaceUID); err != nil {
			if errors.Is(err, dbHeuristic.ErrNoMutationHappened) {
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
			children = append(children, workspace.FindDescendantHeuristicUIDs(nodeMap, child)...)
		}

		if len(children) > 0 {
			// delete the actual heuristics
			if err := dbHeuristic.DeleteUserHeuristics(dgraph, children, userUID, workspaceUID); err != nil {
				if errors.Is(err, dbHeuristic.ErrNoMutationHappened) {
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
func DeleteNodes(dgraph external.Database, workspaceMutex *Mutex, workspaceUID string,
	userUID string, nodeUIDs []string) ([]string, error) {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	if len(w.Nodes) == 0 {
		return nil, serror.NewStackErrorf(
			"node deletion request for empty workspace. workspace: %s", workspaceUID)
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
			return nil, serror.NewStackErrorf(
				"node does not exist in workspace. workspace: %s, node: %s", workspaceUID, clientNode)
		}
	}

	deleteNodesMap := map[string]bool{}
	for _, n := range nodesToDelete {
		nodes, err := deleteNode(dgraph, n, w.Nodes, userUID, workspaceUID)
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

	if err := encodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID,
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
func AddNodes(dgraph external.Database, workspaceMutex *Mutex, worker *worker.Worker, workspaceUID string,
	userUID string, newNodes []*workspace.Node) ([]workspace.Node, error) {
	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	nodeMap, dummyHeuristics, notes := separateNodes(w.Nodes)

	// If the transmitted state is empty, then there are only connections between the new nodes.
	// If newNodes is a destination transaction, it might be connected to heuristics.
	if len(nodeMap) == 0 && len(newNodes) == 1 && !newNodes[0].IsDestination() {
		frontEndNodes := []workspace.Node{*newNodes[0]}
		if err := encodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID,
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

	_, dummyHeuristics = filterDummyNodes(worker, dummyHeuristics, userUID)

	frontEndNodes := slices.Concat(cliutil.GetMapValues(nodeMap), dummyHeuristics, notes)

	if err := encodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID, frontEndNodes, &clusterHeight); err != nil {
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
func AddNote(dgraph external.Database, workspaceMutex *Mutex, workspaceUID string,
	userUID string, note workspace.Node) ([]workspace.Node, error) {
	if len(note.Children) == 0 {
		return nil, serror.NewStackErrorStr("note has no children")
	}

	if note.UID != "" && !strings.HasPrefix(note.UID, noteUIDPrefix) {
		return nil, serror.NewStackErrorf("invalid note uid: %s", note.UID)
	}

	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(dgraph, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	nodeMap, dummyHeuristics, notes := separateNodes(w.Nodes)

	if len(nodeMap) == 0 {
		return nil, serror.NewStackErrorStr("trying to add note to an empty workspace")
	}

	if _, ok := nodeMap[note.Children[0]]; !ok {
		// parent does not exist
		return nil, serror.NewStackErrorf("trying to add note with non-existing child %s", note.Children[0])
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
	frontEndNodes := slices.Concat(cliutil.GetMapValues(nodeMap), dummyHeuristics, notes)

	if err := encodeAndStoreWorkspaceState(dgraph, userUID, workspaceUID, frontEndNodes, nil); err != nil {
		return nil, err
	}

	return frontEndNodes, nil
}

// encodeAndStoreWorkspaceState transforms the workspace state into JSON and stores it in the database
func encodeAndStoreWorkspaceState(dgraph external.Database, userUID string, workspaceUID string,
	state []workspace.Node, clusterHeight *int64) error {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return serror.NewStackError(err)
	}

	return workspace.SetWorkspaceState(dgraph, userUID, workspaceUID, string(stateBytes), clusterHeight)
}

// InsertNodeConnectionsAndHeuristics queries the db for connections between nodes in nodeMap and inserts them.
// Also inserts found heuristics into the node map
func InsertNodeConnectionsAndHeuristics(dgraph external.Database, nodeMap map[string]workspace.Node,
	userUID string, workspaceUID string) (int64, map[string]workspace.Node, error) {
	connections, nodeHeuristics, clusterHeight, err := workspace.GetWorkspaceConnections(dgraph,
		cliutil.GetMapKeys(nodeMap), userUID, workspaceUID)
	if err != nil {
		return 0, nil, err
	}

	newNodeMap := make(map[string]workspace.Node)
	for _, node := range connections {
		nodeElement, ok := nodeMap[node.UID]
		if !ok {
			return 0, nil, serror.NewStackErrorf("uid %s not found in map", node.UID)
		}
		nodeElement.Children = node.Children

		// need to add nodes into new map, otherwise duplicate addresses will be left in the map
		newNodeMap[node.UID] = nodeElement
	}

	// add heuristic nodes to map
	for _, h := range nodeHeuristics {
		// set coordinates
		if oldHeuristic, ok := nodeMap[h.UID]; ok {
			h.X = oldHeuristic.X
			h.Y = oldHeuristic.Y
		}

		newNodeMap[h.UID] = h
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
// - general node: transactions, heuristics and clusters
// - dummy heuristic node: heuristics waiting to be executed
// - notes node: notes
func separateNodes(nodes []workspace.Node) (map[string]workspace.Node, []workspace.Node, []workspace.Node) {
	nodeMap := map[string]workspace.Node{}
	var dummyHeuristics []workspace.Node
	var notes []workspace.Node
	for _, n := range nodes {
		if n.UID == "" || n.Type == "" {
			continue
		}

		if n.Type == workspace.NodeTypeNote {
			notes = append(notes, n)
		} else if n.IsLoading() {
			dummyHeuristics = append(dummyHeuristics, n)
		} else {
			nodeMap[n.UID] = n
		}
	}
	return nodeMap, dummyHeuristics, notes
}

// filterDummyNodes filters the given nodes based on wether they exist in the worker work log and/or
// if they are finished executing. It returns the filtered list and two flags:
// - removedNode: true if at least one node was removed
// - removedFinishedNode: true if at least one was removed because it was finished executing
func filterDummyNodes(worker *worker.Worker, dummyHeuristics []workspace.Node,
	userID string) (changedDummies bool, filteredDummies []workspace.Node) {
	filteredDummies = slices.DeleteFunc(dummyHeuristics, func(node workspace.Node) bool {
		workID, err := strconv.Atoi(node.UID)
		if err != nil {
			// if work ID is invalid delete it as it is useless
			return true
		}

		uid, err := worker.GetFinishedDatabaseUID(workID, userID)
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
