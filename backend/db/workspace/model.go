package workspace

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

type NodeConnections struct {
	UID          string   `json:"uid,omitempty"`
	Clusters     []string `json:"clusters,omitempty"`
	Transactions []string `json:"transactions,omitempty"`
	Addresses    []string `json:"addresses,omitempty"`
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
