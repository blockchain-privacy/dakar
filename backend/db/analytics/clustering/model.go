package clustering

import (
	"strconv"
	"time"
)

type ClusterType string

const TypeHMI ClusterType = "hmi"
const TypeFMI ClusterType = "fmi"
const DType = "Cluster"

type HollowTransaction struct {
	Uid string `json:"uid,omitempty"`
}

type HollowUser struct {
	Uid string `json:"uid,omitempty"`
}

type HollowAddress struct {
	Uid string `json:"uid,omitempty"`
}

type SubCluster struct {
	Uid string `json:"uid,omitempty"`
}

type CustomCluster struct {
	Uid          string          `json:"uid,omitempty"`
	Type         ClusterType     `json:"cluster_type,omitempty"`
	Timestamp    string          `json:"cluster_ts,omitempty"`
	AddressCount *int            `json:"cluster_address_count,omitempty"`
	Addresses    []HollowAddress `json:"cluster_addresses,omitempty"`
	User         HollowUser      `json:"cluster_user,omitempty"`
	DType        []string        `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (cc *CustomCluster) SetDType() {
	cc.DType = []string{DType}
}

type Cluster struct {
	Uid          string            `json:"uid,omitempty"`
	Type         ClusterType       `json:"cluster_type,omitempty"`
	AddressCount *int              `json:"cluster_address_count,omitempty"`
	Transaction  HollowTransaction `json:"cluster_transaction,omitempty"`
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

func NewFMIClusterByUID(UID string) Cluster {
	return Cluster{
		Uid:   UID,
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

// ClusterLookupRequest holds all configuration data for a cluster lookup request
type ClusterLookupRequest struct {
	// AddressHash is either the address hash for which to find clusters
	AddressHash string `json:"addressHash,omitempty"`
}

type FrontendAddress struct {
	AddressHash      string `json:"addresshash,omitempty"`
	OutputCount      int    `json:"output_count,omitempty"`
	SpentOutputCount int    `json:"spent_output_count,omitempty"`
}

type FrontendCluster struct {
	Uid             string            `json:"uid,omitempty"`
	Type            ClusterType       `json:"cluster_type,omitempty"`
	AddressCount    int               `json:"cluster_address_count,omitempty"`
	TransactionHash string            `json:"txhash,omitempty"`
	BlockID         int               `json:"bid,omitempty"`
	BlockHash       string            `json:"bhash,omitempty"`
	Timestamp       time.Time         `json:"ts,omitempty"`
	Addresses       []FrontendAddress `json:"cluster_addresses,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
}

type ClusterTags struct {
	Uid  string `json:"uid,omitempty"`
	Tags []struct {
		Tag string `json:"tag,omitempty"`
	}
}

type FrontendClusterRequest struct {
	Uid          string      `json:"uid,omitempty"`
	Type         ClusterType `json:"cluster_type,omitempty"`
	AddressCount int         `json:"cluster_address_count,omitempty"`
	Transaction  []struct {
		TransactionHash string    `json:"txhash,omitempty"`
		BlockID         int       `json:"bid,omitempty"`
		BlockHash       string    `json:"bhash,omitempty"`
		Timestamp       time.Time `json:"ts,omitempty"`
	} `json:"cluster_transaction,omitempty"`
	Addresses []FrontendAddress `json:"cluster_addresses,omitempty"`
}

type FrontendHMICluster struct {
	Uid             string   `json:"uid,omitempty"`
	AddressCount    int      `json:"cluster_address_count,omitempty"`
	TransactionHash string   `json:"txhash,omitempty"`
	Parent          string   `json:"cluster_parent,omitempty"`
	Children        []string `json:"cluster_children,omitempty"`
}

type FrontendUserCluster struct {
	Uid          string   `json:"uid,omitempty"`
	Timestamp    string   `json:"ts,omitempty"`
	AddressCount int64    `json:"address_count,omitempty"`
	Addresses    []string `json:"addresses,omitempty"`
}
