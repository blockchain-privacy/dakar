package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/cmd/cliutil"
	"gitlab.com/blockchain-privacy/dakar/db/workspace"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

const exportVersion1 = "dakar_workspace_export_v1"

type ExportMeta struct {
	Version        string    `json:"version,omitempty"`
	ExportDate     time.Time `json:"exportDate,omitempty"`
	Name           string    `json:"name,omitempty"`
	BlockchainMode string    `json:"blockchainMode,omitempty"`
}
type Export struct {
	Meta ExportMeta `json:"meta,omitempty"`
	// Transactions and Clusters
	Primitives []workspace.Node `json:"primitives,omitempty"`
	Notes      []workspace.Node `json:"notes,omitempty"`
	Selectors  []workspace.Node `json:"selectors,omitempty"`
}

func NewExport(name string, blockchainMode string) Export {
	return Export{Meta: ExportMeta{
		Version:        exportVersion1,
		ExportDate:     time.Now().UTC(),
		Name:           name,
		BlockchainMode: blockchainMode,
	}}
}

// ExportWorkspace returns all if its nodes, stripped of unnecessary details
func ExportWorkspace(ctx context.Context, dgraph external.Database,
	workspaceMutex *Mutex, blockchainMode string, workspaceUID string, userUID string) (*Export, error) {
	w, err := GetAndRefreshWorkspace(ctx, dgraph, workspaceMutex, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	// nonSelectorNodes holds all nodes which are not selectors
	nonSelectorNodes := make([]workspace.Node, 0, len(w.Nodes))
	var notes []workspace.Node
	var selectors []workspace.Node

	for _, node := range w.Nodes {
		if node.IsTransaction() || node.IsCluster() {
			// reset not needed data
			node.TransactionType = ""
			node.ClusterType = ""
			nonSelectorNodes = append(nonSelectorNodes, node)
		} else if node.IsNote() {
			notes = append(notes, node)
		} else if node.IsSelector() {
			// reset not needed data
			node.SelectorCreated = ""
			node.SelectorModified = ""
			node.SelectorErrorCode = ""
			node.SelectorResultCount = nil
			node.SelectorTotalResultCount = nil
			selectors = append(selectors, node)
		}
	}

	export := NewExport(w.Name, blockchainMode)
	export.Selectors = selectors
	export.Notes = notes
	export.Primitives = nonSelectorNodes

	return &export, nil
}

// ExportBasic returns a two column string array:
// [0]: transactions
// [1]: clusters (addresses)
func ExportBasic(ctx context.Context, dgraph external.Database,
	workspaceMutex *Mutex, workspaceUID string, userUID string) ([][]string, error) {
	w, err := GetAndRefreshWorkspace(ctx, dgraph, workspaceMutex, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	nodes := [][]string{{}, {}}
	for _, node := range w.Nodes {
		if node.IsTransaction() {
			nodes[0] = append(nodes[0], node.TransactionHash)
		} else if node.IsCluster() {
			nodes[1] = append(nodes[1], node.AddressHash)
		}
	}

	return nodes, nil
}

// importPrimitiveNodes imports the given primitive nodes and returns their mapping from external UID to internal UID
func importPrimitiveNodes(ctx context.Context, m *Mutex, c external.Database,
	workspaceUID string, userUID string, nodes []workspace.Node) (map[string]string, error) {
	newNodes := map[string]workspace.Node{}
	for _, node := range nodes {
		identifier := node.AddressHash
		if node.IsTransaction() {
			identifier = node.TransactionHash
		}

		// ignore node if no transaction hash or address hash is present
		if len(identifier) == 0 {
			continue
		}

		newNode, err := workspace.SearchForNode(ctx, c, identifier)
		if err != nil {
			if errors.Is(err, workspace.ErrNodeNotFound) {
				// ignore nodes which are not found
				continue
			}

			warn(err, "import node", node)
			return nil, err
		}

		newNodes[newNode.UID] = *newNode
	}

	externalToInternalMapping := make(map[string]string, len(newNodes))

	if len(newNodes) > 0 {
		workspaceNodes, _, err := AddNodes(ctx, c, m, workspaceUID, userUID, cliutil.GetMapValues(newNodes))
		if err != nil {
			return nil, err
		}

		for _, n := range nodes {
			index := slices.IndexFunc(workspaceNodes, func(node workspace.Node) bool {
				if n.AddressHash != "" && n.AddressHash == node.AddressHash {
					return true
				}
				if n.TransactionHash != "" && n.TransactionHash == node.TransactionHash {
					return true
				}

				return false
			})

			if index == -1 {
				return nil, serror.FromStrWithContext("could not find mapping", "node", n)
			}

			externalToInternalMapping[n.UID] = workspaceNodes[index].UID
		}
	}

	return externalToInternalMapping, nil
}

func getExportMap(export Export) map[string]workspace.Node {
	allNodes := make(map[string]workspace.Node, len(export.Primitives)+len(export.Selectors)+len(export.Notes))

	for _, n := range export.Primitives {
		allNodes[n.UID] = n
	}

	for _, n := range export.Selectors {
		allNodes[n.UID] = n
	}

	for _, n := range export.Notes {
		allNodes[n.UID] = n
	}

	return allNodes
}

// waitForSelectors periodically checks the status of the provided selectors.
// It returns when all selectors have a status other than 'waiting'.
func waitForSelectors(ctx context.Context, c external.Database, selectorUIDs []string,
	workspaceUID string, userUID string) error {
	var err error
	for {
		selectorUIDs = slices.DeleteFunc(selectorUIDs, func(selectorUID string) bool {
			var status string
			status, err = workspace.GetSelectorStatus(ctx, c, selectorUID, workspaceUID, userUID)
			if err != nil {
				return false
			}

			if status != workspace.StatusWaiting {
				return true
			}

			return false
		})

		if err != nil {
			return err
		}

		if len(selectorUIDs) == 0 {
			break
		}

		time.Sleep(time.Second)
	}

	return nil
}

// getRootSelectors returns the UIDs of all selectors that have no parent
func getRootSelectors(allNodes map[string]workspace.Node) []workspace.Node {
	allSelectors := map[string]workspace.Node{}
	for k, n := range allNodes {
		if n.IsSelector() {
			allSelectors[k] = n
		}
	}

	// remove all selectors that are a child of a node
	for _, n := range allNodes {
		for _, child := range n.Children {
			delete(allSelectors, child)
		}
	}
	return cliutil.GetMapValues(allSelectors)
}

// importConnectedSelectors imports all selectors connected to the parent nodes
func importConnectedSelectors(ctx context.Context, c external.Database, m *Mutex, workspaceUID string, userUID string,
	allNodes map[string]workspace.Node, imported map[string]bool, externalToInternalMapping map[string]string) (int, error) {
	var newSelectors []string
	for n := range imported {
		node, ok := allNodes[n]
		if !ok {
			return 0, serror.FromStrWithContext("imported node not in node map", "node", n, "map", allNodes)
		}

		internalParentUID, ok := externalToInternalMapping[node.UID]
		if !ok {
			return 0, serror.FromStrWithContext("unable to find internal parent uid", "external UID", node.UID, "map", externalToInternalMapping)
		}

		var addedSelectors []string

		for _, child := range node.Children {
			if imported[child] {
				// child is already imported
				continue
			}

			childNode, ok := allNodes[child]
			if !ok {
				return 0, serror.FromStrWithContext("child node not in map", "node", child, "map", allNodes)
			}

			if !childNode.IsSelector() {
				continue
			}

			selectorUID, err := importSelector(ctx, c, m, workspaceUID, userUID, &childNode, internalParentUID)
			if err != nil {
				return 0, err
			}

			// update mapping
			externalToInternalMapping[childNode.UID] = selectorUID
			imported[childNode.UID] = true
			newSelectors = append(newSelectors, selectorUID)
			addedSelectors = append(addedSelectors, selectorUID)
		}

		if err := waitForSelectors(ctx, c, addedSelectors, workspaceUID, userUID); err != nil {
			return 0, err
		}
	}

	return len(newSelectors), nil
}

func importSelector(ctx context.Context, c external.Database, m *Mutex, workspaceUID string,
	userUID string, selector *workspace.Node, parentUID string) (string, error) {
	var opt workspace.Options
	if selector.TxPropOptions != nil {
		opt = *selector.TxPropOptions
	} else if selector.TxGraphOptions != nil {
		opt = *selector.TxGraphOptions
	} else if selector.HeuristicOptions != nil {
		opt = *selector.HeuristicOptions
	} else {
		return "", serror.FromStr("invalid selector type")
	}

	selectorUID, _, err := AddSelector(ctx, c, m, opt, selector.SelectorType,
		parentUID, workspaceUID, userUID)
	if err != nil {
		return "", err
	}

	return selectorUID, nil
}

// ImportWorkspace adds a new workspace. Its state gets constructed from the provided export.
func ImportWorkspace(ctx context.Context, c external.Database,
	blockchainMode string, export Export, userUID string) error {
	if export.Meta.Version != exportVersion1 {
		return serror.FromStrWithContext("invalid export version", "version", export.Meta.Version)
	}

	if export.Meta.Name == "" {
		return serror.FromStr("export name is empty")
	}

	if export.Meta.BlockchainMode != blockchainMode {
		return serror.FromStrWithContext("invalid blockchain mode",
			"server blockchain mode", blockchainMode, "import blockchain mode", export.Meta.BlockchainMode)
	}

	marshalledExport, err := json.Marshal(export)
	if err != nil {
		return serror.New(err)
	}

	if _, err = workspace.AddWorkspace(ctx, c, export.Meta.Name, userUID, string(marshalledExport)); err != nil {
		return err
	}

	return nil
}

type ImportWork struct {
	UserUID      string
	ImportFile   string
	WorkspaceUID string
}

func NewImportWork(item workspace.ImportWorkItem) *ImportWork {
	return &ImportWork{
		UserUID:      item.UserUID,
		ImportFile:   item.ImportFile,
		WorkspaceUID: item.WorkspaceUID,
	}
}

func (i *ImportWork) Run(ctx context.Context, m *Mutex, c external.Database, _ *graph.Wrapper) (err error) {
	// in this function use a pointer to the named return to detect an error state
	// and run the deferred cleanup function depending on it.
	errorPointer := &err
	defer func() {
		importStatus := workspace.ImportStatusError
		if *errorPointer == nil {
			importStatus = workspace.ImportStatusSuccess
		}
		if cleanUpErr := workspace.DeleteImportAndSetStatus(ctx, c, i.WorkspaceUID, importStatus); cleanUpErr != nil {
			warn(cleanUpErr)
		}
	}()

	var e Export
	if err = json.Unmarshal([]byte(i.ImportFile), &e); err != nil {
		err = serror.New(err)
		return
	}

	var externalToInternalMapping map[string]string
	externalToInternalMapping, err = importPrimitiveNodes(ctx, m, c, i.WorkspaceUID, i.UserUID, e.Primitives)
	if err != nil {
		return
	}

	imported := make(map[string]bool, len(e.Primitives))
	for _, n := range e.Primitives {
		imported[n.UID] = true
	}

	allNodes := getExportMap(e)
	rootSelectors := getRootSelectors(allNodes)
	var addedSelectors []string
	// import selectors that have no parent node
	for _, rootSelector := range rootSelectors {
		var selectorUID string
		selectorUID, err = importSelector(ctx, c, m, i.WorkspaceUID, i.UserUID, &rootSelector, "")
		if err != nil {
			return
		}

		// update mapping
		externalToInternalMapping[rootSelector.UID] = selectorUID
		imported[rootSelector.UID] = true
		addedSelectors = append(addedSelectors, selectorUID)
	}

	if err = waitForSelectors(ctx, c, addedSelectors, i.WorkspaceUID, i.UserUID); err != nil {
		return
	}

	// import selectors connected to parent nodes
	for {
		var numImported int
		numImported, err = importConnectedSelectors(ctx, c, m, i.WorkspaceUID, i.UserUID, allNodes, imported, externalToInternalMapping)
		if err != nil {
			return err
		}
		if numImported == 0 {
			// done importing selectors
			break
		}
	}

	for _, note := range e.Notes {
		children := make([]string, len(note.Children))
		for index, child := range note.Children {
			internalChild, ok := externalToInternalMapping[child]
			if !ok {
				err = serror.FromStrWithContext("unable to find internal uid", "external UID", child, "map", externalToInternalMapping)
				return
			}
			children[index] = internalChild
		}

		note.Children = children

		if _, err = AddNote(ctx, c, m, i.WorkspaceUID, i.UserUID, note); err != nil {
			return
		}
	}

	nodesWithInternalUIDs := make([]workspace.Node, len(allNodes))
	index := 0
	for _, n := range allNodes {
		if n.IsNote() {
			nodesWithInternalUIDs[index] = n
			index++
			// note UIDs do not have internal UID
			continue
		}

		internalUID, ok := externalToInternalMapping[n.UID]
		if !ok {
			err = serror.FromStrWithContext("unable to find internal uid", "external UID", n.UID, "map", externalToInternalMapping)
			return
		}

		n.UID = internalUID
		nodesWithInternalUIDs[index] = n

		index++
	}

	if err = UpdateNodeCoordinates(ctx, c, m, i.WorkspaceUID, i.UserUID, nodesWithInternalUIDs); err != nil {
		return err
	}

	return
}
