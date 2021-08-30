package clustering

import "strconv"

type ClusterType string

const TypeHMI ClusterType = "hmi"
const TypeFMI ClusterType = "fmi"
const DType = "Cluster"

type HollowTransaction struct {
	Uid string `json:"uid"`
}

type HollowAddress struct {
	Uid string `json:"uid"`
}

type SubCluster struct {
	Uid string `json:"uid"`
}

type Cluster struct {
	Uid          string            `json:"uid"`
	Type         ClusterType       `json:"cluster_type"`
	AddressCount *int              `json:"cluster_address_count"`
	Transaction  HollowTransaction `json:"cluster_transaction"`
	Children     []SubCluster      `json:"cluster_children,omitempty"`
	Addresses    []HollowAddress   `json:"cluster_addresses,omitempty"`
	DType        []string          `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (c *Cluster) SetDType() {
	c.DType = []string{DType}
}

func NewHMICluster(index int, txUID string) Cluster {
	return Cluster{
		Uid:         "_:c" + strconv.Itoa(index),
		Type:        TypeHMI,
		Transaction: HollowTransaction{Uid: txUID},
		DType:       []string{DType},
	}
}

func NewFMICluster(index int) Cluster {
	return Cluster{
		Uid:   "_:c" + strconv.Itoa(index),
		Type:  TypeFMI,
		DType: []string{DType},
	}
}

type ClusterWithParent struct {
	Uid          string `json:"uid"`
	AddressCount int    `json:"cluster_address_count,omitempty"`
	Parents      []struct {
		Uid string `json:"uid"`
	} `json:"parents,omitempty"`
}

type ClusterAddress struct {
	Uid      string              `json:"uid"`
	Clusters []ClusterWithParent `json:"clusters,omitempty"`
}

type ClusterTransaction struct {
	Uid       string           `json:"uid"`
	Addresses []ClusterAddress `json:"addr,omitempty"`
}

type TransactionWithAddresses struct {
	Uid       string          `json:"uid"`
	Addresses []HollowAddress `json:"addr,omitempty"`
}
