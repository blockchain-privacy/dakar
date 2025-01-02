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
	"encoding/csv"
	"encoding/json"
	"errors"
	"github.com/qrest/gomisc/serror"
	"strconv"
)

type Options interface {
	workspace.TxPropOptions | workspace.TxGraphOptions | dbHeuristic.Options
	// IsValid returns true if the Options are valid. hasParent should be set to true if the associated selector has a parent.
	IsValid(hasParent bool) bool
}

type TxPropWork struct {
	opt          workspace.TxPropOptions
	selectorUID  string
	workspaceUID string
	userUID      string
	parentUID    string
}

func NewTxPropWork(item workspace.WorkItem) (*TxPropWork, error) {
	if item.SelectorOptions == "" {
		return nil, serror.FromStrWithContext("empty selector options", "item", item)
	}

	var opt workspace.TxPropOptions
	if err := json.Unmarshal([]byte(item.SelectorOptions), &opt); err != nil {
		return nil, serror.New(err)
	}

	return &TxPropWork{
		opt:          opt,
		workspaceUID: item.WorkspaceUID,
		userUID:      item.UserUID,
		selectorUID:  item.SelectorUID,
		parentUID:    item.ParentUID,
	}, nil
}

// Run processes the selector and updates it into the workspace
func (s TxPropWork) Run(ctx context.Context, workspaceMutex *Mutex, c external.Database, _ *graph.Wrapper) error {
	// 1. Do work
	status := workspace.StatusSuccess
	var newNodes []any
	results, totalResultCount, err := workspace.DoSelection(ctx, c, s.opt, s.parentUID)
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
	return updateSelector(ctx, workspaceMutex, c, s.selectorUID, s.workspaceUID, s.userUID, status, newNodes, totalResultCount)
}

type TxGraphWork struct {
	opt          workspace.TxGraphOptions
	selectorUID  string
	workspaceUID string
	userUID      string
	parentUID    string
}

func NewTxGraphWork(item workspace.WorkItem) (*TxGraphWork, error) {
	if item.SelectorOptions == "" || item.ParentUID == "" {
		return nil, serror.FromStrWithContext("invalid selector", "item", item)
	}

	var opt workspace.TxGraphOptions
	if err := json.Unmarshal([]byte(item.SelectorOptions), &opt); err != nil {
		return nil, serror.New(err)
	}

	return &TxGraphWork{
		opt:          opt,
		workspaceUID: item.WorkspaceUID,
		userUID:      item.UserUID,
		selectorUID:  item.SelectorUID,
		parentUID:    item.ParentUID,
	}, nil
}

// Run processes the selector and updates it into the workspace
func (s TxGraphWork) Run(ctx context.Context, workspaceMutex *Mutex, c external.Database, _ *graph.Wrapper) error {
	// 1. Do work
	status := workspace.StatusSuccess
	var newNodes []any
	results, totalResultCount, err := workspace.DoGraphSelection(ctx, c, s.opt, s.parentUID)
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
	return updateSelector(ctx, workspaceMutex, c, s.selectorUID, s.workspaceUID, s.userUID, status, newNodes, totalResultCount)
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

// isValidParent checks if the parent is a selector, if it belongs to the user and if it was successfully executed
func isValidParent(ctx context.Context, dgraph external.Database, selectorParent string, workspaceUID string, userUID string) (bool, error) {
	parentType, err := db.GetTypeByUID(ctx, dgraph, selectorParent)
	if err != nil {
		return false, err
	}

	if parentType != workspace.SelectorDType {
		return true, nil
	}

	// check if parent belongs to user and if status is correct
	status, err := workspace.GetSelectorStatus(ctx, dgraph, selectorParent, workspaceUID, userUID)
	if err != nil {
		if errors.Is(err, workspace.ErrInvalidSelector) {
			return false, nil
		}

		return false, err
	}

	return status == workspace.StatusSuccess, nil
}

// AddSelector adds a new selector to the workspace. It returns UID the updated workspace.
func AddSelector[O Options](ctx context.Context, dgraph external.Database, workspaceMutex *Mutex, options O,
	selectorType string, selectorParent string, workspaceUID string, userUID string) (string, []workspace.Node, error) {
	if !workspace.IsTypeValid(selectorType) {
		return "", nil, serror.NewWithContext(db.ErrInvalidRequestArgument, "type", selectorType)
	}

	if !options.IsValid(selectorParent != "") {
		return "", nil, serror.NewWithContext(db.ErrInvalidRequestArgument, "options", options, "type", selectorType)
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
			return "", nil, serror.NewWithContext(db.ErrInvalidRequestArgument, "options", options, "type", selectorType)
		}
		newNode.HeuristicOptions = &opt
	case workspace.TypeTxProp:
		opt, ok := any(options).(workspace.TxPropOptions)
		if !ok {
			return "", nil, serror.NewWithContext(db.ErrInvalidRequestArgument, "options", options, "type", selectorType)
		}
		newNode.TxPropOptions = &opt
	case workspace.TypeTxGraph:
		opt, ok := any(options).(workspace.TxGraphOptions)
		if !ok {
			return "", nil, serror.NewWithContext(db.ErrInvalidRequestArgument, "options", options, "type", selectorType)
		}
		newNode.TxGraphOptions = &opt
	default:
		return "", nil, serror.NewWithContext(db.ErrInvalidRequestArgument, "options", options, "type", selectorType)
	}

	if selectorParent != "" {
		isValid, err := isValidParent(ctx, dgraph, selectorParent, workspaceUID, userUID)
		if err != nil {
			return "", nil, err
		}
		if !isValid {
			return "", nil, serror.NewWithContext(db.ErrInvalidRequestArgument, "parent", selectorParent)
		}
	}

	optionStr, err := json.Marshal(options)
	if err != nil {
		return "", nil, serror.NewWithContext(err, "options", options)
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

// updateSelector updates the selector both in the workspace state and in the db.
func updateSelector(ctx context.Context, workspaceMutex *Mutex, dgraph external.Database,
	selectorUID string, workspaceUID string, userUID string, status string, newNodes []any, totalResults int) error {
	if updateErr := workspace.UpdateSelector(ctx, dgraph, &workspace.Selector{
		UID:              selectorUID,
		Status:           status,
		Results:          newNodes,
		TotalResultCount: &totalResults,
	}, userUID, workspaceUID); updateErr != nil {
		return updateErr
	}

	// 3. Update workspace
	lock := workspaceMutex.Lock(workspaceUID)
	defer lock.Unlock()

	w, err := workspace.GetFrontendWorkspace(ctx, dgraph, workspaceUID, userUID)
	if err != nil {
		return err
	}

	nodeMap, notes := separateNodes(w.Nodes)

	clusterHeight, nodeMap, err := InsertNodeConnectionsAndHeuristics(ctx, dgraph, nodeMap, userUID, workspaceUID)
	if err != nil {
		return err
	}

	frontEndNodes := append(cliutil.GetMapValues(nodeMap), notes...)

	return encodeAndStoreWorkspaceState(ctx, dgraph, userUID, workspaceUID, frontEndNodes, &clusterHeight)
}

// Run processes the heuristic and inserts it into the workspace
func (h HeuristicWork) Run(ctx context.Context, workspaceMutex *Mutex, c external.Database, g *graph.Wrapper) error {
	// 1. Do work
	status := workspace.StatusSuccess
	results, err := h.executor.Run(ctx, c, g)
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
	return updateSelector(ctx, workspaceMutex, c, h.selectorUID, h.workspaceUID, h.userUID, status, newNodes, len(newNodes))
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

// WriteClustersToCsv writes the provided cluster to a csv file
func WriteClustersToCsv(writer *csv.Writer, clusters []workspace.HeuristicCluster) error {
	header := []string{"cluster ID", "attributions", "transaction hash", "timestamp"}

	if err := writer.Write(header); err != nil {
		return serror.New(err)
	}

	var clusterCount int
	for _, c := range clusters {
		clusterCount++
		var attributions string

		for i, a := range c.Attributions {
			attributions += a.Tag

			if i+1 < len(c.Attributions) {
				attributions += ","
			}
		}

		for _, transaction := range c.Transactions {
			row := []string{strconv.Itoa(clusterCount), attributions, transaction.Hash, transaction.Timestamp}

			if err := writer.Write(row); err != nil {
				return serror.New(err)
			}
		}
		writer.Flush()
	}

	writer.Flush()
	return nil
}

// WriteTransactionsToCsv writes the provided transactions to a csv file
func WriteTransactionsToCsv(writer *csv.Writer, transactions []workspace.TransactionWithTimestamp) error {
	header := []string{"transaction hash", "timestamp"}

	if err := writer.Write(header); err != nil {
		return serror.New(err)
	}

	for _, transaction := range transactions {
		if err := writer.Write([]string{transaction.Hash, transaction.Timestamp}); err != nil {
			return serror.New(err)
		}
	}
	writer.Flush()

	return nil
}
