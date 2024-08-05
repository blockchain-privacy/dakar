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
	workspace.Options | dbHeuristic.Config
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
	var err error
	status := workspace.StatusSuccess
	var newNodes []db.UIDNode
	results, err := workspace.DoSelection(ctx, c, s.opt)
	if err == nil {
		newNodes = make([]db.UIDNode, len(results))
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
		Type:    workspace.TypeTransactionProperties,
		Status:  status,
		Results: newNodes,
	}, s.userUID, s.workspaceUID); updateErr != nil {
		return updateErr
	}

	lock := workspaceMutex.Lock(s.workspaceUID)
	defer lock.Unlock()

	// 3. Update workspace
	w, err := workspace.GetFrontendWorkspace(ctx, c, s.workspaceUID, s.userUID)
	if err != nil {
		return err
	}

	nodeMap, notes := separateNodes(w.Nodes)

	clusterHeight, nodeMap, err := InsertNodeConnectionsAndHeuristics(c, nodeMap, s.userUID, s.workspaceUID)
	if err != nil {
		return err
	}

	// try to set node position
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

// AddSelector adds a new selector to the workspace. It returns UID of the newly created selector,
// which can be used to check its execution status.
func AddSelector[O Options](ctx context.Context, dgraph external.Database, workspaceMutex *Mutex, options O,
	selectorType string, selectorParent string, workspaceUID string, userUID string) (string, error) {
	if !workspace.IsTypeValid(selectorType) {
		return "", serror.FromStrWithContext("invalid type", "type", selectorType)
	}

	workspaceLock := workspaceMutex.Lock(workspaceUID)
	defer workspaceLock.Unlock()

	w, err := workspace.GetFrontendWorkspace(ctx, dgraph, workspaceUID, userUID)
	if err != nil {
		return "", err
	}

	var parentIndex int
	var parentNode *db.UIDNode
	if selectorType == workspace.TypeHeuristic {
		opt, ok := any(options).(dbHeuristic.Config)
		if !ok {
			return "", serror.FromStrWithContext("options type mismatch", "options", options, "type", selectorType)
		}

		parentIndex, parentNode, err = getHeuristicParent(selectorParent, opt.TransactionHash, w.Nodes)
		if err != nil {
			return "", serror.AddContext(err, "options", options)
		}
	} else if selectorType == workspace.TypeTransactionProperties {
		if _, ok := any(options).(workspace.Options); !ok {
			return "", serror.FromStrWithContext("options type mismatch", "options", options, "type", selectorType)
		}

		parentIndex, parentNode, err = getSelectorParent(selectorParent, w.Nodes)
		if err != nil {
			return "", serror.AddContext(err, "options", options)
		}
	} else {
		return "", serror.FromStrWithContext("invalid selector type", "options", options, "type", selectorType)
	}

	optionStr, err := json.Marshal(options)
	if err != nil {
		return "", serror.NewWithContext(err, "options", options)
	}

	selectorUID, err := workspace.InsertSelector(ctx, dgraph, &workspace.Selector{
		Type:    selectorType,
		Status:  workspace.StatusWaiting,
		Parent:  parentNode,
		Options: string(optionStr),
	}, userUID, workspaceUID)
	if err != nil {
		return "", err
	}

	if parentIndex >= 0 {
		// add new selector uid to children of parent
		w.Nodes[parentIndex].Children = append(w.Nodes[parentIndex].Children, selectorUID)
	}

	// add node
	w.Nodes = append(w.Nodes, workspace.Node{
		UID:            selectorUID,
		Type:           workspace.NodeTypeSelector,
		SelectorType:   selectorType,
		SelectorStatus: workspace.StatusWaiting,
	})

	if err = encodeAndStoreWorkspaceState(ctx, dgraph, userUID, workspaceUID, w.Nodes, w.ClusterHeight); err != nil {
		return "", err
	}

	return selectorUID, nil
}

type HeuristicWork struct {
	executor     heuristics.Executor
	workspaceUID string
	selectorUID  string
	userUID      string
}

// Run processes the heuristic and inserts it into the workspace
func (h HeuristicWork) Run(workspaceMutex *Mutex, dgraph external.Database, g *graph.Wrapper) error {
	newHeuristic, err := h.executor.Run(dgraph, g)
	if err != nil {
		return err
	}

	lock := workspaceMutex.Lock(h.workspaceUID)
	defer lock.Unlock()

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	newHeuristicUID, err := dbHeuristic.InsertHeuristic(dgraph, newHeuristic, h.userUID, h.workspaceUID)
	if err != nil {
		return err
	}

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

	// try to set node position
	if newNode, ok := nodeMap[newHeuristicUID]; ok {
		// todo set actual status
		newNode.SelectorStatus = workspace.StatusSuccess
		nodeMap[newHeuristicUID] = newNode
	}

	frontEndNodes := append(cliutil.GetMapValues(nodeMap), notes...)

	return encodeAndStoreWorkspaceState(ctx, dgraph, h.userUID, h.workspaceUID, frontEndNodes, &clusterHeight)
}

func NewHeuristicWork(item workspace.WorkItem) (*HeuristicWork, error) {
	if item.SelectorOptions == "" {
		return nil, serror.FromStrWithContext("empty selector options", "item", item)
	}

	var opt dbHeuristic.Config
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

func getHeuristicParent(parentheuristicUID string, txHash string, nodes []workspace.Node) (int, *db.UIDNode, error) {
	// find the index of the hew heuristic's parent
	parentIndex := -1
	if parentheuristicUID == "" {
		for i, n := range nodes {
			if n.TransactionHash == txHash {
				parentIndex = i
				break
			}
		}
	} else {
		for i, n := range nodes {
			if n.UID == parentheuristicUID {
				parentIndex = i
				break
			}
		}
	}

	// no parent found
	if parentIndex == -1 {
		return parentIndex, nil, serror.FromStr("could not determine parent for new heuristic")
	}

	return parentIndex, &db.UIDNode{UID: nodes[parentIndex].UID}, nil
}
