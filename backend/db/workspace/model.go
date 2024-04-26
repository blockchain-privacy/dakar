package workspace

import (
	"backend/constants"
	"backend/db/analytics/heuristics"
)

const DType = "Workspace"

// possible type of workspace nodes. Addresses are not considered, only their connected FMI cluster
const (
	NodeTypeCluster     = "cluster"
	NodeTypeTransaction = "transaction"
	NodeTypeHeuristic   = "heuristic"
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
		UID     string `json:"uid,omitempty"`
		Cluster []struct {
			UID string `json:"uid,omitempty"`
		} `json:"cluster,omitempty"`
	} `json:"address_cluster,omitempty"`

	Transactions []struct {
		UID     string `json:"uid,omitempty"`
		Outputs []struct {
			InputTransactions []struct {
				UID string `json:"uid,omitempty"`
			} `json:"~tx_inputs,omitempty"`
			Addresses []struct {
				Clusters []struct {
					UID string `json:"uid,omitempty"`
				} `json:"~Cluster.addresses,omitempty"`
			} `json:"~addr_outputs,omitempty"`
		} `json:"tx_outputs,omitempty"`
		Inputs []struct {
			OutputTransactions []struct {
				UID string `json:"uid,omitempty"`
			} `json:"~tx_outputs,omitempty"`
			Addresses []struct {
				Clusters []struct {
					UID string `json:"uid,omitempty"`
				} `json:"~Cluster.addresses,omitempty"`
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
							Clusters []struct {
								UID string `json:"uid,omitempty"`
							} `json:"~Cluster.addresses,omitempty"`
						} `json:"~addr_outputs,omitempty"`
					} `json:"tx_inputs,omitempty"`
					Outputs []struct {
						Addresses []struct {
							Clusters []struct {
								UID string `json:"uid,omitempty"`
							} `json:"~Cluster.addresses,omitempty"`
						} `json:"~addr_outputs,omitempty"`
					} `json:"tx_outputs,omitempty"`
				} `json:"HeuristicResult.destinations,omitempty"`
				Origin struct {
					Inputs []struct {
						Addresses []struct {
							Clusters []struct {
								UID string `json:"uid,omitempty"`
							} `json:"~Cluster.addresses,omitempty"`
						} `json:"~addr_outputs,omitempty"`
					} `json:"tx_inputs,omitempty"`
					Outputs []struct {
						Addresses []struct {
							Clusters []struct {
								UID string `json:"uid,omitempty"`
							} `json:"~Cluster.addresses,omitempty"`
						} `json:"~addr_outputs,omitempty"`
					} `json:"tx_outputs,omitempty"`
				} `json:"HeuristicResult.origin,omitempty"`
			} `json:"HeuristicCluster.results,omitempty"`
		} `json:"Heuristic.clusters,omitempty"`
	} `json:"heuristic_clusters,omitempty"`

	ClusterClusters []struct {
		UID       string `json:"uid,omitempty"`
		Addresses []struct {
			Outputs []struct {
				InputClusters []struct {
					UID string `json:"uid,omitempty"`
				} `json:"~tx_inputs,omitempty"`
				OutputClusters []struct {
					UID string `json:"uid,omitempty"`
				} `json:"~tx_outputs,omitempty"`
			} `json:"addr_outputs,omitempty"`
		} `json:"Cluster.addresses,omitempty"`
	} `json:"cluster_clusters,omitempty"`

	Heuristics []struct {
		UID                 string   `json:"uid,omitempty"`
		Timestamp           string   `json:"ts,omitempty"`
		Type                string   `json:"type,omitempty"`
		Parameter           string   `json:"parameter,omitempty"`
		ExcludeAddresses    bool     `json:"excludeAddresses"`
		ExcludeSpendingGaps bool     `json:"excludeSpendingGaps"`
		ClusterTypes        []string `json:"clusterTypes,omitempty"`
		Transaction         struct {
			UID string `json:"uid,omitempty"`
		} `json:"transaction,omitempty"`
		ParentHeuristic []heuristics.HollowHeuristic `json:"parent,omitempty"`
		ChildHeuristics []heuristics.HollowHeuristic `json:"children,omitempty"`
		ClusterCount    *int                         `json:"clusterCount,omitempty"`
		Clusters        []struct {
			Results []struct {
				Origin struct {
					UID string `json:"uid,omitempty"`
				} `json:"HeuristicResult.origin,omitempty"`
				Destinations []struct {
					UID string `json:"uid,omitempty"`
				} `json:"HeuristicResult.destinations,omitempty"`
			} `json:"HeuristicCluster.results,omitempty"`
		} `json:"Heuristic.clusters,omitempty"`
	} `json:"heuristics,omitempty"`
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

	// heuristic
	HeuristicType       string   `json:"heuristicType,omitempty"`
	Parameter           string   `json:"heuristicParameter,omitempty"`
	ExcludeAddresses    *bool    `json:"heuristicExcludeAddresses,omitempty"`
	ExcludeSpendingGaps *bool    `json:"heuristicExcludeSpendingGaps,omitempty"`
	ClusterTypes        []string `json:"heuristicClusterTypes,omitempty"`
	ClusterCount        *int     `json:"heuristicClusterCount,omitempty"`
	Timestamp           string   `json:"heuristicTs,omitempty"`
	// Loading is true if a new heuristic has been created, which is not yet finished executing
	Loading *bool `json:"loading,omitempty"`
}

func (f Node) IsDestination() bool {
	return f.Type == NodeTypeTransaction && f.PrivacyType != nil && constants.PrivacyType(*f.PrivacyType).IsDestination()
}

func (f Node) IsLoading() bool {
	return f.Loading != nil && *f.Loading
}
