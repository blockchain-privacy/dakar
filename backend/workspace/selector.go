// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"backend/analytics/graph"
	"backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	dbh "backend/db/heuristics"
	"backend/db/workspace"
	"backend/external"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strconv"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

var (
	errInvalidParent  = errors.New("invalid selector parent")
	errInvalidOptions = errors.New("invalid selector option")
	errInvalidType    = errors.New("invalid selector type")
	errTooManyResults = errors.New("selector returned too many results")
)

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
	var errorCode string
	var newNodes []any
	results, totalResultCount, err := workspace.DoSelection(ctx, c, s.opt, s.parentUID)
	if err == nil {
		if len(results) > workspace.MaxResultsPerSelector {
			status = workspace.StatusError
			errorCode = workspace.ErrorCodeResultLimitExceeded
			err = serror.NewWithContext(errTooManyResults, "result count", len(results), "options", s.opt)
		} else {
			newNodes = make([]any, len(results))
			for i, result := range results {
				newNodes[i] = db.UIDNode{UID: result}
			}
		}
	} else {
		// despite the error, we don't return here because we want to store the error state in the db
		status = workspace.StatusError
		err = serror.AddContext(err, "options", s.opt)
	}

	// 2. Store work
	updateErr := updateSelector(ctx, workspaceMutex, c, &workspace.Selector{
		UID:              s.selectorUID,
		Status:           status,
		Results:          newNodes,
		TotalResultCount: &totalResultCount,
		ErrorCode:        errorCode,
	}, s.workspaceUID, s.userUID)
	if updateErr != nil {
		// both running and updating the selector returned an error
		if err != nil {
			return serror.AddContext(err, "updateSelector error", updateErr)
		}
		// only updating the selector returned an error
		return updateErr
	}

	// running may have returned an error
	return err
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
	var errorCode string
	var newNodes []any
	results, totalResultCount, err := workspace.DoGraphSelection(ctx, c, s.opt, s.parentUID)
	if err == nil {
		if len(results) > workspace.MaxResultsPerSelector {
			status = workspace.StatusError
			errorCode = workspace.ErrorCodeResultLimitExceeded
			err = serror.NewWithContext(errTooManyResults, "result count", len(results), "options", s.opt)
		} else {
			newNodes = make([]any, len(results))
			for i, result := range results {
				newNodes[i] = db.UIDNode{UID: result}
			}
		}
	} else {
		// despite the error, we don't return here because we want to store the error state in the db
		status = workspace.StatusError
		err = serror.AddContext(err, "options", s.opt)
	}

	// 2. Store work
	updateErr := updateSelector(ctx, workspaceMutex, c, &workspace.Selector{
		UID:              s.selectorUID,
		Status:           status,
		Results:          newNodes,
		TotalResultCount: &totalResultCount,
		ErrorCode:        errorCode,
	}, s.workspaceUID, s.userUID)
	if updateErr != nil {
		// both running and updating the selector returned an error
		if err != nil {
			return serror.AddContext(err, "updateSelector error", updateErr)
		}
		// only updating the selector returned an error
		return updateErr
	}

	// running the selector may have returned an error
	return err
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
func isValidParent(ctx context.Context, dgraph external.Database, selectorParent string,
	workspaceUID string, userUID string) (bool, error) {
	databaseType, err := db.GetTypeByUID(ctx, dgraph, selectorParent)
	if err != nil {
		return false, err
	}

	if databaseType == "" {
		return false, serror.FromStr("selector parent type not set")
	}

	if databaseType != workspace.SelectorDType {
		return true, nil
	}

	// check if parent belongs to user and that the selector's status is correct
	status, err := workspace.GetSelectorStatus(ctx, dgraph, selectorParent, workspaceUID, userUID)
	if err != nil {
		if errors.Is(err, workspace.ErrInvalidSelector) {
			return false, nil
		}

		return false, err
	}

	return status == workspace.StatusSuccess, nil
}

// checkOptions validates the options and attaches them to a new selector.
// Validates selector type, selector options and selector parent type.
func checkOptions(ctx context.Context, dgraph external.Database, options workspace.Options,
	selectorType string, selectorParent string, workspaceUID string, userUID string) (*workspace.Node, error) {
	if !workspace.IsTypeValid(selectorType) {
		return nil, serror.New(errInvalidType)
	}

	if !options.IsValid(ctx, dgraph, selectorParent) {
		return nil, serror.New(errInvalidOptions)
	}

	newNode := workspace.Node{
		Type:           workspace.NodeTypeSelector,
		SelectorType:   selectorType,
		SelectorStatus: workspace.StatusWaiting,
	}

	// check  if selector type and options match
	switch selectorType {
	case constants.TypeHeuristic:
		opt, ok := any(options).(heuristics.HeuristicOptions)
		if !ok {
			return nil, serror.New(errInvalidOptions)
		}

		newNode.HeuristicOptions = &opt
	case constants.TypeTxProp:
		opt, ok := any(options).(workspace.TxPropOptions)
		if !ok {
			return nil, serror.New(errInvalidOptions)
		}

		newNode.TxPropOptions = &opt
	case constants.TypeTxGraph:
		opt, ok := any(options).(workspace.TxGraphOptions)
		if !ok {
			return nil, serror.New(errInvalidOptions)
		}

		newNode.TxGraphOptions = &opt
	default:
		return nil, serror.New(errInvalidOptions)
	}

	if selectorParent != "" {
		isValid, err := isValidParent(ctx, dgraph, selectorParent, workspaceUID, userUID)
		if err != nil {
			return nil, err
		}
		if !isValid {
			return nil, serror.New(errInvalidParent)
		}
	}
	return &newNode, nil
}

// AddSelector adds and validates a new selector to the workspace. It returns the UID of the updated workspace.
func AddSelector(ctx context.Context, dgraph external.Database, workspaceMutex *Mutex, options workspace.Options,
	selectorType string, selectorParent string, workspaceUID string, userUID string) (string, []workspace.Node, error) {
	newNode, err := checkOptions(ctx, dgraph, options, selectorType, selectorParent, workspaceUID, userUID)
	if err != nil {
		return "", nil, serror.NewWithContext(db.ErrInvalidRequestArgument,
			"options", options, "type", selectorType, "parent", selectorParent)
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
	w.Nodes = append(w.Nodes, *newNode)

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
	selector *workspace.Selector, workspaceUID string, userUID string) error {
	if updateErr := workspace.UpdateSelector(ctx, dgraph, selector, userUID, workspaceUID); updateErr != nil {
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

// getConnectionCount returns the number of new database graph connections required, if the clusters were inserted.
func getConnectionCount(clusters []dbh.HeuristicCluster) int {
	var count int
	for _, c := range clusters {
		// transaction count + 1 cluster
		count += len(c.Results) + 1
	}

	return count
}

// Run processes the heuristic and inserts it into the workspace
func (h HeuristicWork) Run(ctx context.Context, workspaceMutex *Mutex, c external.Database, g *graph.Wrapper) error {
	// 1. Do work
	status := workspace.StatusSuccess
	var errorCode string
	results, err := h.executor.Run(ctx, c, g)
	var newNodes []any
	if err == nil {
		resultCount := getConnectionCount(results)
		if resultCount > workspace.MaxResultsPerSelector {
			status = workspace.StatusError
			errorCode = workspace.ErrorCodeResultLimitExceeded
			err = serror.NewWithContext(errTooManyResults, "result count", resultCount, "selector uid", h.selectorUID)
		} else {
			newNodes = make([]any, len(results))
			for i, result := range results {
				newNodes[i] = result
			}
		}
	} else {
		// despite the error, we don't return here because we want to store the error state in the db
		status = workspace.StatusError
	}

	// 2. Store work
	resultCount := len(newNodes)
	updateErr := updateSelector(ctx, workspaceMutex, c, &workspace.Selector{
		UID:              h.selectorUID,
		Status:           status,
		Results:          newNodes,
		TotalResultCount: &resultCount,
		ErrorCode:        errorCode,
	}, h.workspaceUID, h.userUID)
	if updateErr != nil {
		// both running and updating the selector returned an error
		if err != nil {
			return serror.AddContext(err, "updateSelector error", updateErr)
		}
		// only updating the selector returned an error
		return updateErr
	}

	// running may have returned an error
	return err
}

func NewHeuristicWork(item workspace.WorkItem) (*HeuristicWork, error) {
	if item.SelectorOptions == "" {
		return nil, serror.FromStrWithContext("empty selector options", "item", item)
	}

	var opt heuristics.HeuristicOptions
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

// WriteTransactionsToCsv writes the provided transactions to a csv file
func WriteTransactionsToCsv(writer *csv.Writer, transactions []workspace.TransactionWithTimestamp) error {
	hasClusters := transactions[0].Cluster != nil
	header := []string{"transaction hash", "transaction type", "timestamp"}
	if hasClusters {
		header = append([]string{"cluster ID"}, header...)
	}

	if err := writer.Write(header); err != nil {
		return serror.New(err)
	}

	for _, t := range transactions {
		line := []string{t.Hash, t.Type, t.Timestamp}

		if t.Cluster != nil {
			line = append([]string{strconv.Itoa(*t.Cluster)}, line...)
		}

		if err := writer.Write(line); err != nil {
			return serror.New(err)
		}
	}
	writer.Flush()

	return nil
}
