package clustering

import (
	"strconv"
	"time"
)

type ClusterType string

const (
	TypeHMI    ClusterType = "hmi"
	TypeFMI    ClusterType = "fmi"
	TypeCustom ClusterType = "custom"
	DType                  = "Cluster"
)

type HollowTransaction struct {
	UID string `json:"uid,omitempty"`
}

type HollowUser struct {
	UID string `json:"uid,omitempty"`
}

type HollowAddress struct {
	UID string `json:"uid,omitempty"`
}

type SubCluster struct {
	UID string `json:"uid,omitempty"`
}

type CustomCluster struct {
	UID          string          `json:"uid,omitempty"`
	Type         ClusterType     `json:"Cluster.type,omitempty"`
	Timestamp    string          `json:"Cluster.ts,omitempty"`
	AddressCount *int            `json:"Cluster.addressCount,omitempty"`
	Addresses    []HollowAddress `json:"Cluster.addresses,omitempty"`
	User         HollowUser      `json:"Cluster.user,omitempty"`
	DType        []string        `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (cc *CustomCluster) SetDType() {
	cc.DType = []string{DType}
}

type Cluster struct {
	UID          string            `json:"uid,omitempty"`
	Type         ClusterType       `json:"Cluster.type,omitempty"`
	AddressCount *int              `json:"Cluster.addressCount,omitempty"`
	Transaction  HollowTransaction `json:"Cluster.transaction,omitempty"`
	Children     []SubCluster      `json:"Cluster.children,omitempty"`
	Addresses    []HollowAddress   `json:"Cluster.addresses,omitempty"`
	DType        []string          `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (c *Cluster) SetDType() {
	c.DType = []string{DType}
}

func NewHMICluster(index int, txUID string) Cluster {
	return Cluster{
		UID:         "_:c" + strconv.Itoa(index),
		Type:        TypeHMI,
		Transaction: HollowTransaction{UID: txUID},
		DType:       []string{DType},
	}
}

func NewFMICluster(index int) Cluster {
	return Cluster{
		UID:   "_:c" + strconv.Itoa(index),
		Type:  TypeFMI,
		DType: []string{DType},
	}
}

func NewFMIClusterByUID(uid string) Cluster {
	return Cluster{
		UID:   uid,
		Type:  TypeFMI,
		DType: []string{DType},
	}
}

type ClusterWithParent struct {
	UID          string `json:"uid"`
	AddressCount int    `json:"Cluster.addressCount,omitempty"`
	Parents      []struct {
		UID string `json:"uid"`
	} `json:"parents,omitempty"`
}

type ClusterAddress struct {
	UID      string              `json:"uid"`
	Clusters []ClusterWithParent `json:"clusters,omitempty"`
}

type ClusterTransaction struct {
	UID       string           `json:"uid"`
	Addresses []ClusterAddress `json:"addr,omitempty"`
}

type TransactionWithAddresses struct {
	UID       string          `json:"uid"`
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
	UID             string            `json:"uid,omitempty"`
	Type            ClusterType       `json:"type,omitempty"`
	AddressCount    int               `json:"addressCount,omitempty"`
	TransactionHash string            `json:"txhash,omitempty"`
	BlockID         int               `json:"bid,omitempty"`
	BlockHash       string            `json:"bhash,omitempty"`
	Timestamp       time.Time         `json:"ts,omitempty"`
	Addresses       []FrontendAddress `json:"addresses,omitempty"`
	Attributions    []Attribution     `json:"attributions,omitempty"`
}

type Attribution struct {
	Tag      string `json:"tag,omitempty"`
	IsPublic bool   `json:"isPublic,omitempty"`
}

type ClusterTags struct {
	UID          string        `json:"uid,omitempty"`
	Attributions []Attribution `json:"tags,omitempty"`
}

type FrontendClusterRequest struct {
	UID          string      `json:"uid,omitempty"`
	Type         ClusterType `json:"Cluster.type,omitempty"`
	AddressCount int         `json:"Cluster.addressCount,omitempty"`
	Transaction  []struct {
		TransactionHash string    `json:"txhash,omitempty"`
		BlockID         int       `json:"bid,omitempty"`
		BlockHash       string    `json:"bhash,omitempty"`
		Timestamp       time.Time `json:"ts,omitempty"`
	} `json:"Cluster.transaction,omitempty"`
	Addresses []FrontendAddress `json:"Cluster.addresses,omitempty"`
}

type FrontendHMICluster struct {
	UID             string   `json:"uid,omitempty"`
	AddressCount    int      `json:"addressCount,omitempty"`
	TransactionHash string   `json:"txhash,omitempty"`
	Parent          string   `json:"parent,omitempty"`
	Children        []string `json:"children,omitempty"`
}

type FrontendUserCluster struct {
	UID          string   `json:"uid,omitempty"`
	Timestamp    string   `json:"ts,omitempty"`
	AddressCount int64    `json:"address_count,omitempty"`
	Addresses    []string `json:"addresses,omitempty"`
}
