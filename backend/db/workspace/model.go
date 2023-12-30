package workspace

import (
	"backend/constants"
	"backend/db/analytics/heuristics"
)

const DType = "Workspace"

type Workspace struct {
	UID              string   `json:"uid,omitempty"`
	Name             string   `json:"Workspace.name,omitempty"`
	ModificationTime string   `json:"Workspace.ts,omitempty"`
	State            string   `json:"Workspace.state,omitempty"`
	DType            []string `json:"dgraph.type,omitempty"`
}

func (w *Workspace) SetDType() {
	w.DType = []string{DType}
}

type FrontendWorkspace struct {
	UID              string `json:"uid,omitempty"`
	Name             string `json:"name,omitempty"`
	ModificationTime string `json:"ts,omitempty"`
	State            string `json:"state,omitempty"`
}

func (w *Workspace) ToFrontendWorkspace() FrontendWorkspace {
	return FrontendWorkspace{
		UID:              w.UID,
		Name:             w.Name,
		ModificationTime: w.ModificationTime,
		State:            w.State,
	}
}

type NodeConnections struct {
	UID      string   `json:"uid,omitempty"`
	Children []string `json:"children,omitempty"`
}

type connectionRequest struct {
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

	TransactionAddresses []struct {
		UID     string `json:"uid,omitempty"`
		Outputs []struct {
			Addresses []struct {
				UID string `json:"uid,omitempty"`
			} `json:"~addr_outputs,omitempty"`
		} `json:"tx_outputs,omitempty"`
		Inputs []struct {
			Addresses []struct {
				UID string `json:"uid,omitempty"`
			} `json:"~addr_outputs,omitempty"`
		} `json:"tx_inputs,omitempty"`
	} `json:"transaction_addresses,omitempty"`

	AddressClusters []struct {
		UID            string `json:"uid,omitempty"`
		AddressOutputs []struct {
			InputTransaction []struct {
				ClusterUID string `json:"uid,omitempty"`
			} `json:"~tx_inputs,omitempty"`
			OutputTransaction []struct {
				ClusterUID string `json:"uid,omitempty"`
			} `json:"~tx_outputs,omitempty"`
		} `json:"addr_outputs,omitempty"`
	} `json:"address_clusters,omitempty"`

	AddressAddresses []struct {
		UID            string `json:"uid,omitempty"`
		AddressOutputs []struct {
			InputTransaction []struct {
				Outputs []struct {
					Addresses []struct {
						UID string `json:"uid,omitempty"`
					} `json:"~addr_outputs,omitempty"`
				} `json:"tx_outputs,omitempty"`
			} `json:"~tx_inputs,omitempty"`
			OutputTransaction []struct {
				Inputs []struct {
					Addresses []struct {
						UID string `json:"uid,omitempty"`
					} `json:"~addr_outputs,omitempty"`
				} `json:"tx_inputs,omitempty"`
			} `json:"~tx_outputs,omitempty"`
		} `json:"addr_outputs,omitempty"`
	} `json:"address_addresses,omitempty"`

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
		UID        string `json:"uid,omitempty"`
		Heuristics []struct {
			UID                 string                       `json:"uid,omitempty"`
			Timestamp           string                       `json:"ts,omitempty"`
			Type                string                       `json:"type,omitempty"`
			Parameter           string                       `json:"parameter,omitempty"`
			ExcludeAddresses    bool                         `json:"excludeAddresses"`
			ExcludeSpendingGaps bool                         `json:"excludeSpendingGaps"`
			ClusterTypes        []string                     `json:"clusterTypes,omitempty"`
			ParentHeuristic     []heuristics.HollowHeuristic `json:"parent,omitempty"`
			ChildHeuristics     []heuristics.HollowHeuristic `json:"children,omitempty"`
			ClusterCount        *int                         `json:"clusterCount,omitempty"`
			Clusters            []struct {
				Results []struct {
					Origin struct {
						UID string `json:"uid,omitempty"`
					} `json:"HeuristicResult.origin,omitempty"`
				} `json:"HeuristicCluster.results,omitempty"`
			} `json:"Heuristic.clusters,omitempty"`
		} `json:"~Heuristic.transaction,omitempty"`
	} `json:"heuristics,omitempty"`
}

type GraphNode struct {
	UID  string `json:"uid,omitempty"`
	Type string `json:"type,omitempty"`

	AddressHash string `json:"addressHash,omitempty"`
	ClusterType string `json:"clusterType,omitempty"`

	TransactionHash string `json:"transactionHash,omitempty"`
	PrivacyType     *int   `json:"privacyType,omitempty"`

	Children []string `json:"children,omitempty"`
}

func (g GraphNode) ToFrontendGraphNode() FrontendGraphNode {
	return FrontendGraphNode{
		UID:             g.UID,
		Type:            g.Type,
		AddressHash:     g.AddressHash,
		ClusterType:     g.ClusterType,
		TransactionHash: g.TransactionHash,
		PrivacyType:     g.PrivacyType,
		Children:        g.Children,
	}
}

func (g GraphNode) IsDestination() bool {
	return g.Type == "transaction" && g.PrivacyType != nil && constants.PrivacyType(*g.PrivacyType).IsDestination()
}

// FrontendGraphNode include the coordinates of the client canvas
type FrontendGraphNode struct {
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
}
