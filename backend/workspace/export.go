package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/analytics/heuristics"
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
	Primitives []ExportNode `json:"primitives,omitempty"`
	Notes      []ExportNode `json:"notes,omitempty"`
	Selectors  []ExportNode `json:"selectors,omitempty"`
}

func NewExport(name string, blockchainMode string) Export {
	return Export{Meta: ExportMeta{
		Version:        exportVersion1,
		ExportDate:     time.Now().UTC(),
		Name:           name,
		BlockchainMode: blockchainMode,
	}}
}

type ExportNode struct {
	UID      string   `json:"uid,omitempty"`
	Children []string `json:"children,omitempty"`
	Type     string   `json:"type,omitempty"`

	// tx hash or address hash
	Identifier string `json:"identifier,omitempty"`

	SelectorType     string                       `json:"selectorType,omitempty"`
	TxPropOptions    *workspace.TxPropOptions     `json:"txPropOptions,omitempty"`
	TxGraphOptions   *workspace.TxGraphOptions    `json:"txGraphOptions,omitempty"`
	HeuristicOptions *heuristics.HeuristicOptions `json:"heuristicOptions,omitempty"`

	// Note
	Text string `json:"text,omitempty"`
}

func (e ExportNode) IsSelector() bool {
	return e.Type == workspace.NodeTypeSelector
}

// ExportWorkspace returns the name of the workspace and all if its nodes, stripped of unnecessary details
func ExportWorkspace(ctx context.Context, dgraph external.Database,
	workspaceMutex *Mutex, blockchainMode string, workspaceUID string, userUID string) (*Export, error) {
	w, err := GetAndRefreshWorkspace(ctx, dgraph, workspaceMutex, workspaceUID, userUID)
	if err != nil {
		return nil, err
	}

	// nonSelectorNodes holds all nodes which are not selectors
	nonSelectorNodes := make([]ExportNode, 0, len(w.Nodes))
	var notes []ExportNode
	var selectors []ExportNode

	for _, node := range w.Nodes {
		if node.IsTransaction() {
			nonSelectorNodes = append(nonSelectorNodes, ExportNode{UID: node.UID, Identifier: node.TransactionHash, Children: node.Children, Type: node.Type})
		} else if node.IsCluster() {
			nonSelectorNodes = append(nonSelectorNodes, ExportNode{UID: node.UID, Identifier: node.AddressHash, Children: node.Children, Type: node.Type})
		} else if node.IsNote() {
			notes = append(notes, ExportNode{UID: node.UID, Text: node.Text, Children: node.Children, Type: node.Type})
		} else if node.IsSelector() {
			selectors = append(selectors, ExportNode{UID: node.UID, Children: node.Children,
				Type:             node.Type,
				SelectorType:     node.SelectorType,
				HeuristicOptions: node.HeuristicOptions,
				TxGraphOptions:   node.TxGraphOptions,
				TxPropOptions:    node.TxPropOptions})
		}
	}

	export := NewExport(w.Name, blockchainMode)
	export.Selectors = selectors
	export.Notes = notes
	export.Primitives = nonSelectorNodes

	return &export, nil
}

// importPrimitiveNodes imports the given primitive nodes and returns their mapping from external UID to internal UID
func importPrimitiveNodes(ctx context.Context, m *Mutex, c external.Database,
	workspaceUID string, userUID string, nodes []ExportNode) (map[string]string, error) {
	newNodes := map[string]workspace.Node{}
	for _, node := range nodes {
		newNode, err := workspace.SearchForNode(ctx, c, node.Identifier)
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
				return n.Identifier == node.AddressHash || n.Identifier == node.TransactionHash
			})

			if index == -1 {
				return nil, serror.FromStrWithContext("could not find mapping", "node", n)
			}

			externalToInternalMapping[n.UID] = workspaceNodes[index].UID
		}
	}

	return externalToInternalMapping, nil
}

func getExportMap(export Export) map[string]ExportNode {
	allNodes := make(map[string]ExportNode, len(export.Primitives)+len(export.Selectors)+len(export.Notes))

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

func importSelectors(ctx context.Context, m *Mutex, c external.Database, workspaceUID string, userUID string,
	allNodes map[string]ExportNode, imported map[string]bool, externalToInternalMapping map[string]string) (int, error) {
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

			var opt workspace.Options
			if childNode.TxPropOptions != nil {
				opt = *childNode.TxPropOptions
			} else if childNode.TxGraphOptions != nil {
				opt = *childNode.TxGraphOptions
			} else if childNode.HeuristicOptions != nil {
				opt = *childNode.HeuristicOptions
			} else {
				return 0, serror.FromStr("invalid selector type")
			}

			selectorUID, _, err := AddSelector(ctx, c, m, opt, childNode.SelectorType,
				internalParentUID, workspaceUID, userUID)
			if err != nil {
				return 0, err
			}
			// update mapping
			externalToInternalMapping[childNode.UID] = selectorUID
			imported[childNode.UID] = true
			newSelectors = append(newSelectors, selectorUID)
		}
	}

	if err := waitForSelectors(ctx, c, newSelectors, workspaceUID, userUID); err != nil {
		return 0, err
	}

	return len(newSelectors), nil
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
		if err = workspace.DeleteImportAndSetStatus(ctx, c, i.WorkspaceUID, importStatus); err != nil {
			warn(err)
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

	for {
		var numImported int
		numImported, err = importSelectors(ctx, m, c, i.WorkspaceUID, i.UserUID, allNodes, imported, externalToInternalMapping)
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
		for i, child := range note.Children {
			internalChild, ok := externalToInternalMapping[child]
			if !ok {
				err = serror.FromStrWithContext("unable to find internal uid", "external UID", child, "map", externalToInternalMapping)
				return
			}
			children[i] = internalChild
		}

		if _, err = AddNote(ctx, c, m, i.WorkspaceUID, i.UserUID, workspace.Node{
			UID: note.UID, Text: note.Text, Children: children,
		}); err != nil {
			return
		}
	}

	return
}
