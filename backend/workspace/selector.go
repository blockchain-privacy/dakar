package workspace

import (
	"backend/analytics/graph"
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
	}, nil
}

// Run processes the selector and updates it into the workspace
func (s SelectorWork) Run(workspaceMutex *Mutex, c external.Database, _ *graph.Wrapper) error {
	ctx, cancel := db.GetBackendContext()
	defer cancel()

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
		status = workspace.StatusError
		// todo add error message into selector
	}

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

	// update workspace
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
		// todo set tx hash
		parentIndex, parentNode, err = getHeuristicParent(selectorParent, "", w.Nodes)
		if err != nil {
			return "", serror.AddContext(err, "options", options)
		}
	} else {
		parentIndex, parentNode, err = getSelectorParent(selectorParent, w.Nodes)
		if err != nil {
			return "", serror.AddContext(err, "options", options)
		}
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
