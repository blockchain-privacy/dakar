package workspace

import (
	"backend/constants"
	"backend/db"
	dbHeuristic "backend/db/analytics/heuristics"
)

const DType = "Workspace"

// possible type of workspace nodes. Addresses are not considered, only their connected FMI cluster
const (
	NodeTypeCluster     = "cluster"
	NodeTypeTransaction = "transaction"
	NodeTypeHeuristic   = "heuristic"
	NodeTypeSelector    = "selector"
	NodeTypeNote        = "note"
)

type Workspace struct {
	UID              string   `json:"uid,omitempty"`
	Name             string   `json:"Workspace.name,omitempty"`
	ModificationTime string   `json:"Workspace.ts,omitempty"`
	ClusterHeight    *int64   `json:"Workspace.clusterHeight,omitempty"`
	State            string   `json:"Workspace.state,omitempty"`
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
	Nodes            []Node `json:"nodes,omitempty"`
}

func (w *Workspace) ToFrontendWorkspace() FrontendWorkspace {
	return FrontendWorkspace{
		UID:              w.UID,
		Name:             w.Name,
		ModificationTime: w.ModificationTime,
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
				Destinations []struct {
					Inputs []struct {
						Addresses []struct {
							Clusters []db.UIDNode `json:"~Cluster.addresses,omitempty"`
						} `json:"~addr_outputs,omitempty"`
					} `json:"tx_inputs,omitempty"`
				} `json:"HeuristicResult.destinations,omitempty"`
				Origin struct {
					Inputs []struct {
						Addresses []struct {
							Clusters []db.UIDNode `json:"~Cluster.addresses,omitempty"`
						} `json:"~addr_outputs,omitempty"`
					} `json:"tx_inputs,omitempty"`
				} `json:"HeuristicResult.origin,omitempty"`
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
		Options     string       `json:"options,omitempty"`
		Parent      *db.UIDNode  `json:"parent,omitempty"`
		Children    []db.UIDNode `json:"children,omitempty"`
		ResultCount *int         `json:"resultCount,omitempty"`
		Clusters    []struct {
			Results []struct {
				Origin       db.UIDNode   `json:"HeuristicResult.origin,omitempty"`
				Destinations []db.UIDNode `json:"HeuristicResult.destinations,omitempty"`
			} `json:"HeuristicCluster.results,omitempty"`
		} `json:"results,omitempty"`
	} `json:"heuristics,omitempty"`

	Selectors []struct {
		UID         string       `json:"uid,omitempty"`
		Created     string       `json:"created,omitempty"`
		Modified    string       `json:"modified,omitempty"`
		Type        string       `json:"type,omitempty"`
		Status      string       `json:"status,omitempty"`
		Options     string       `json:"options,omitempty"`
		Parent      *db.UIDNode  `json:"parent,omitempty"`
		Children    []db.UIDNode `json:"children,omitempty"`
		ResultCount *int         `json:"resultCount,omitempty"`
		Results     []db.UIDNode `json:"results,omitempty"`
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
	PrivacyType     *int   `json:"privacyType,omitempty"`

	// selector
	SelectorCreated     string               `json:"selectorCreated,omitempty"`
	SelectorModified    string               `json:"selectorModified,omitempty"`
	SelectorType        string               `json:"selectorType,omitempty"`
	SelectorStatus      string               `json:"selectorStatus,omitempty"`
	SelectorResultCount *int                 `json:"selectorResultCount,omitempty"`
	SelectorOptions     *Options             `json:"selectorOptions,omitempty"`
	HeuristicOptions    *dbHeuristic.Options `json:"heuristicOptions,omitempty"`

	// note
	Text string `json:"text,omitempty"`

	// Loading is true if a new heuristic has been created, which is not yet finished executing
	Loading *bool `json:"loading,omitempty"`
}

func (f Node) IsDestination() bool {
	return f.Type == NodeTypeTransaction && f.PrivacyType != nil && constants.PrivacyType(*f.PrivacyType).IsDestination()
}

func (f Node) IsLoading() bool {
	return f.Loading != nil && *f.Loading
}

type dummyUser struct {
	UID        string      `json:"uid,omitempty"`
	Workspaces []Workspace `json:"User.workspaces,omitempty"`
}
