package workspace

import (
	"backend/analytics/graph"
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/db"
	dbHeuristic "backend/db/analytics/heuristics"
	"backend/db/workspace"
	"backend/external"
	"context"
	"encoding/json"
	"github.com/qrest/gomisc/serror"
)

type SelectorWork struct {
	opt          workspace.Options
	selectorUID  string
	workspaceUID string
	userUID      string
	parentUID    string
}

type Options interface {
	workspace.Options | dbHeuristic.Options
	// IsValid returns true if the Options are valid
	IsValid() bool
}

func NewSelectorWork(item workspace.WorkItem) (*SelectorWork, error) {
	if item.SelectorOptions == "" {
		return nil, serror.FromStrWithContext("empty selector options", "item", item)
	}

	var opt workspace.Options
	if err := json.Unmarshal([]byte(item.SelectorOptions), &opt); err != nil {
		return nil, err
	}

	return &SelectorWork{
		opt:          opt,
		workspaceUID: item.WorkspaceUID,
		userUID:      item.UserUID,
		selectorUID:  item.SelectorUID,
		parentUID:    item.ParentUID,
	}, nil
}

// Run processes the selector and updates it into the workspace
func (s SelectorWork) Run(workspaceMutex *Mutex, c external.Database, _ *graph.Wrapper) error {
	ctx, cancel := db.GetBackendContext()
	defer cancel()

	// 1. Do work
	status := workspace.StatusSuccess
	var newNodes []any
	results, err := workspace.DoSelection(ctx, c, s.opt)
	if err == nil {
		newNodes = make([]any, len(results))
		for i, result := range results {
			newNodes[i] = db.UIDNode{UID: result}
		}
	} else {
		// despite the error, we don't return here because we want to store the error state in the db
		status = workspace.StatusError
		warn(err, "options", s.opt)
	}

	// 2. Store work
	if updateErr := workspace.UpdateSelector(ctx, c, &workspace.Selector{
		UID:     s.selectorUID,
		Status:  status,
		Results: newNodes,
	}, s.userUID, s.workspaceUID); updateErr != nil {
		return updateErr
	}

	// 3. Update workspace
	lock := workspaceMutex.Lock(s.workspaceUID)
	defer lock.Unlock()

	w, err := workspace.GetFrontendWorkspace(ctx, c, s.workspaceUID, s.userUID)
	if err != nil {
		return err
	}

	nodeMap, notes := separateNodes(w.Nodes)

	clusterHeight, nodeMap, err := InsertNodeConnectionsAndHeuristics(c, nodeMap, s.userUID, s.workspaceUID)
	if err != nil {
		return err
	}

	// try to set node status
	if n, ok := nodeMap[s.selectorUID]; ok {
		n.SelectorStatus = status
		nodeMap[s.selectorUID] = n
	}

	frontEndNodes := append(cliutil.GetMapValues(nodeMap), notes...)

	return encodeAndStoreWorkspaceState(ctx, c, s.userUID, s.workspaceUID, frontEndNodes, &clusterHeight)
}

func getSelectorParent(selectorParent string, nodes []workspace.Node) (int, *db.UIDNode, error) {
	parentIndex := -1
	var parentNode *db.UIDNode
	// find the index of the selector's parent
	if selectorParent != "" {
		for i, n := range nodes {
			if n.UID == selectorParent {
				parentIndex = i
				break
			}
		}

		// no parent found
		if parentIndex == -1 {
			return parentIndex, nil, serror.FromStr("parent of selector is not contained in workspace")
		}

		parentNode = &db.UIDNode{UID: selectorParent}
	}

	return parentIndex, parentNode, nil
}

// AddSelector adds a new selector to the workspace. It returns UID the updated workspace.
func AddSelector[O Options](ctx context.Context, dgraph external.Database, workspaceMutex *Mutex, options O,
	selectorType string, selectorParent string, workspaceUID string, userUID string) (string, []workspace.Node, error) {
	if !workspace.IsTypeValid(selectorType) {
		return "", nil, serror.FromStrWithContext("invalid type", "type", selectorType)
	}

	if !options.IsValid() {
		return "", nil, serror.FromStrWithContext("invalid options", "options", options, "type", selectorType)
	}

	newNode := workspace.Node{
		Type:           workspace.NodeTypeSelector,
		SelectorType:   selectorType,
		SelectorStatus: workspace.StatusWaiting,
	}

	// check  if selector type and options match
	switch selectorType {
	case workspace.TypeHeuristic:
		opt, ok := any(options).(dbHeuristic.Options)
		if !ok {
			return "", nil, serror.FromStrWithContext("options type mismatch", "options", options, "type", selectorType)
		}
		newNode.HeuristicOptions = &opt
	case workspace.TypeTransactionProperties:
		opt, ok := any(options).(workspace.Options)
		if !ok {
			return "", nil, serror.FromStrWithContext("options type mismatch", "options", options, "type", selectorType)
		}
		newNode.SelectorOptions = &opt
	default:
		return "", nil, serror.FromStrWithContext("invalid selector type", "options", options, "type", selectorType)
	}

	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(ctx, dgraph, workspaceUID, userUID)
	if err != nil {
		return "", nil, err
	}

	parentIndex, parentNode, err := getSelectorParent(selectorParent, w.Nodes)
	if err != nil {
		return "", nil, serror.AddContext(err, "options", options)
	}

	optionStr, err := json.Marshal(options)
	if err != nil {
		return "", nil, serror.NewWithContext(err, "options", options)
	}

	newNode.UID, err = workspace.InsertSelector(ctx, dgraph, &workspace.Selector{
		Type:    selectorType,
		Status:  workspace.StatusWaiting,
		Parent:  parentNode,
		Options: string(optionStr),
	}, userUID, workspaceUID)
	if err != nil {
		return "", nil, err
	}

	if parentIndex >= 0 {
		// add new selector uid to children of parent
		w.Nodes[parentIndex].Children = append(w.Nodes[parentIndex].Children, newNode.UID)
	}

	// add node
	w.Nodes = append(w.Nodes, newNode)

	if err = encodeAndStoreWorkspaceState(ctx, dgraph, userUID, workspaceUID, w.Nodes, w.ClusterHeight); err != nil {
		return "", nil, err
	}

	return newNode.UID, w.Nodes, nil
}

type HeuristicWork struct {
	executor     heuristics.Executor
	workspaceUID string
	selectorUID  string
	userUID      string
}

// Run processes the heuristic and inserts it into the workspace
func (h HeuristicWork) Run(workspaceMutex *Mutex, dgraph external.Database, g *graph.Wrapper) error {
	// 1. Do work
	status := workspace.StatusSuccess
	results, err := h.executor.Run(dgraph, g)
	var newNodes []any
	if err == nil {
		newNodes = make([]any, len(results))
		for i, result := range results {
			newNodes[i] = result
		}
	} else {
		// despite the error, we don't return here because we want to store the error state in the db
		status = workspace.StatusError
		warn(err)
	}

	// 2. Store work
	ctx, cancel := db.GetBackendContext()
	defer cancel()

	if updateErr := workspace.UpdateSelector(ctx, dgraph, &workspace.Selector{
		UID:     h.selectorUID,
		Status:  status,
		Results: newNodes,
	}, h.userUID, h.workspaceUID); updateErr != nil {
		return updateErr
	}

	// 3. Update workspace
	lock := workspaceMutex.Lock(h.workspaceUID)
	defer lock.Unlock()

	// update workspace
	w, err := workspace.GetFrontendWorkspace(ctx, dgraph, h.workspaceUID, h.userUID)
	if err != nil {
		return err
	}

	nodeMap, notes := separateNodes(w.Nodes)

	clusterHeight, nodeMap, err := InsertNodeConnectionsAndHeuristics(dgraph, nodeMap, h.userUID, h.workspaceUID)
	if err != nil {
		return err
	}

	// try to set node status
	if newNode, ok := nodeMap[h.selectorUID]; ok {
		newNode.SelectorStatus = status
		nodeMap[h.selectorUID] = newNode
	}

	frontEndNodes := append(cliutil.GetMapValues(nodeMap), notes...)

	return encodeAndStoreWorkspaceState(ctx, dgraph, h.userUID, h.workspaceUID, frontEndNodes, &clusterHeight)
}

func NewHeuristicWork(item workspace.WorkItem) (*HeuristicWork, error) {
	if item.SelectorOptions == "" {
		return nil, serror.FromStrWithContext("empty selector options", "item", item)
	}

	var opt dbHeuristic.Options
	if err := json.Unmarshal([]byte(item.SelectorOptions), &opt); err != nil {
		return nil, err
	}

	executor, err := heuristics.ConstructExecutors(opt, item.UserUID, item.ParentUID)
	if err != nil {
		return nil, err
	}

	return &HeuristicWork{
		executor:     executor,
		workspaceUID: item.WorkspaceUID,
		userUID:      item.UserUID,
		selectorUID:  item.SelectorUID,
	}, nil
}
