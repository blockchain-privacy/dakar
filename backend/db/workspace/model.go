// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"gitlab.com/blockchain-privacy/dakar/analytics/heuristics"
	"gitlab.com/blockchain-privacy/dakar/db"
)

const DType = "Workspace"

// possible type of workspace nodes. Addresses are not considered, only their connected FMI cluster
const (
	NodeTypeCluster     = "cluster"
	NodeTypeTransaction = "transaction"
	NodeTypeSelector    = "selector"
	NodeTypeNote        = "note"
)

type Workspace struct {
	UID              string   `json:"uid,omitempty"`
	Name             string   `json:"Workspace.name,omitempty"`
	ModificationTime string   `json:"Workspace.ts,omitempty"`
	ClusterHeight    *int64   `json:"Workspace.clusterHeight,omitempty"`
	State            string   `json:"Workspace.state,omitempty"`
	ImportFile       string   `json:"Workspace.importFile,omitempty"`
	ImportStatus     string   `json:"Workspace.importStatus,omitempty"`
	ImportTime       string   `json:"Workspace.importTs,omitempty"`
	DType            []string `json:"dgraph.type,omitempty"`
}

func (w *Workspace) SetDType() {
	w.DType = []string{DType}
}

type DecodedWorkspace struct {
	UID              string
	Name             string
	ModificationTime string
	ClusterHeight    *int64
	Nodes            []Node
}

func (w *DecodedWorkspace) ToFrontendWorkspace() *FrontendWorkspace {
	return &FrontendWorkspace{
		UID:              w.UID,
		Name:             w.Name,
		ModificationTime: w.ModificationTime,
		Nodes:            w.Nodes,
	}
}

type FrontendWorkspace struct {
	UID              string `json:"uid,omitempty"`
	Name             string `json:"name,omitempty"`
	ModificationTime string `json:"ts,omitempty"`
	ImportStatus     string `json:"importStatus,omitempty"`
	ImportTime       string `json:"importTs,omitempty"`
	Nodes            []Node `json:"nodes,omitempty"`
}

func (w *Workspace) ToFrontendWorkspace() FrontendWorkspace {
	return FrontendWorkspace{
		UID:              w.UID,
		Name:             w.Name,
		ModificationTime: w.ModificationTime,
		ImportStatus:     w.ImportStatus,
		ImportTime:       w.ImportTime,
	}
}

type NodeConnections struct {
	UID      string   `json:"uid,omitempty"`
	Children []string `json:"children,omitempty"`
}

type NodeConnectionsMap struct {
	UID      string
	children map[string]bool
}

type connectionRequest struct {
	ClusterHeight []struct {
		LastClusteredID *int64 `json:"lastclusteredid,omitempty"`
	} `json:"cluster_height,omitempty"`

	AddressClusters []struct {
		UID     string       `json:"uid,omitempty"`
		Cluster []db.UIDNode `json:"cluster,omitempty"`
	} `json:"address_cluster,omitempty"`

	Transactions []struct {
		UID     string `json:"uid,omitempty"`
		Outputs []struct {
			InputTransactions []db.UIDNode `json:"~tx_inputs,omitempty"`
			Addresses         []struct {
				Clusters []db.UIDNode `json:"~Cluster.addresses,omitempty"`
			} `json:"~addr_outputs,omitempty"`
		} `json:"tx_outputs,omitempty"`
		Inputs []struct {
			OutputTransactions []db.UIDNode `json:"~tx_outputs,omitempty"`
			Addresses          []struct {
				Clusters []db.UIDNode `json:"~Cluster.addresses,omitempty"`
			} `json:"~addr_outputs,omitempty"`
		} `json:"tx_inputs,omitempty"`
	} `json:"transactions,omitempty"`

	HeuristicClusters []struct {
		UID      string `json:"uid,omitempty"`
		Clusters []struct {
			Results []struct {
				Inputs []struct {
					Addresses []struct {
						Clusters []db.UIDNode `json:"~Cluster.addresses,omitempty"`
					} `json:"~addr_outputs,omitempty"`
				} `json:"tx_inputs,omitempty"`
			} `json:"HeuristicCluster.results,omitempty"`
		} `json:"results,omitempty"`
	} `json:"heuristic_clusters,omitempty"`

	SelectorClusters []struct {
		UID     string `json:"uid,omitempty"`
		Results []struct {
			Inputs []struct {
				Addresses []struct {
					Clusters []db.UIDNode `json:"~Cluster.addresses,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_inputs,omitempty"`
		} `json:"results,omitempty"`
	} `json:"selector_clusters,omitempty"`

	ClusterClusters []struct {
		UID       string `json:"uid,omitempty"`
		Addresses []struct {
			Outputs []struct {
				InputClusters  []db.UIDNode `json:"~tx_inputs,omitempty"`
				OutputClusters []db.UIDNode `json:"~tx_outputs,omitempty"`
			} `json:"addr_outputs,omitempty"`
		} `json:"Cluster.addresses,omitempty"`
	} `json:"cluster_clusters,omitempty"`

	Heuristics []struct {
		UID         string       `json:"uid,omitempty"`
		Created     string       `json:"created,omitempty"`
		Modified    string       `json:"modified,omitempty"`
		Type        string       `json:"type,omitempty"`
		Status      string       `json:"status,omitempty"`
		ErrorCode   string       `json:"errorCode,omitempty"`
		Options     string       `json:"options,omitempty"`
		Parent      *db.UIDNode  `json:"parent,omitempty"`
		Children    []db.UIDNode `json:"children,omitempty"`
		ResultCount *int         `json:"resultCount,omitempty"`
		Clusters    []struct {
			Results []db.UIDNode `json:"HeuristicCluster.results,omitempty"`
		} `json:"results,omitempty"`
	} `json:"heuristics,omitempty"`

	Selectors []struct {
		UID              string       `json:"uid,omitempty"`
		Created          string       `json:"created,omitempty"`
		Modified         string       `json:"modified,omitempty"`
		Type             string       `json:"type,omitempty"`
		Status           string       `json:"status,omitempty"`
		ErrorCode        string       `json:"errorCode,omitempty"`
		Options          string       `json:"options,omitempty"`
		Parent           *db.UIDNode  `json:"parent,omitempty"`
		Children         []db.UIDNode `json:"children,omitempty"`
		ResultCount      *int         `json:"resultCount,omitempty"`
		TotalResultCount *int         `json:"totalResultCount,omitempty"`
		Results          []db.UIDNode `json:"results,omitempty"`
	} `json:"selectors,omitempty"`
}

// Node is the data model of a workspace node
type Node struct {
	// UID can also hold the work ID of a pending heuristic
	UID      string   `json:"uid,omitempty"`
	Type     string   `json:"type,omitempty"`
	Children []string `json:"children,omitempty"`
	X        *float32 `json:"x,omitempty"`
	Y        *float32 `json:"y,omitempty"`

	// address
	AddressHash string `json:"addressHash,omitempty"`
	ClusterType string `json:"clusterType,omitempty"`

	// transaction
	TransactionHash string `json:"transactionHash,omitempty"`
	TransactionType string `json:"txtype,omitempty"`

	// selector
	SelectorCreated          string                       `json:"selectorCreated,omitempty"`
	SelectorModified         string                       `json:"selectorModified,omitempty"`
	SelectorType             string                       `json:"selectorType,omitempty"`
	SelectorStatus           string                       `json:"selectorStatus,omitempty"`
	SelectorErrorCode        string                       `json:"selectorErrorCode,omitempty"`
	SelectorResultCount      *int                         `json:"selectorResultCount,omitempty"`
	SelectorTotalResultCount *int                         `json:"selectorTotalResultCount,omitempty"`
	TxPropOptions            *TxPropOptions               `json:"txPropOptions,omitempty"`
	TxGraphOptions           *TxGraphOptions              `json:"txGraphOptions,omitempty"`
	HeuristicOptions         *heuristics.HeuristicOptions `json:"heuristicOptions,omitempty"`

	// note
	Text string `json:"text,omitempty"`
}

func (f Node) IsTransaction() bool {
	return f.Type == NodeTypeTransaction
}

func (f Node) IsCluster() bool {
	return f.Type == NodeTypeCluster
}

func (f Node) IsSelector() bool {
	return f.Type == NodeTypeSelector
}
func (f Node) IsNote() bool {
	return f.Type == NodeTypeNote
}

type dummyUser struct {
	UID        string      `json:"uid,omitempty"`
	Workspaces []Workspace `json:"User.workspaces,omitempty"`
}
