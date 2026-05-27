// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package clustering

import (
	"backend/db"
	"strconv"
	"time"
)

type ClusterType string

const (
	TypeFMI    ClusterType = "fmi"
	TypeCustom ClusterType = "custom"
	DType                  = "Cluster"
)

type CustomCluster struct {
	UID          string       `json:"uid,omitempty"`
	Type         ClusterType  `json:"Cluster.type,omitempty"`
	Timestamp    string       `json:"Cluster.ts,omitempty"`
	AddressCount *int         `json:"Cluster.addressCount,omitempty"`
	Addresses    []db.UIDNode `json:"Cluster.addresses,omitempty"`
	User         db.UIDNode   `json:"Cluster.user,omitempty"`
	DType        []string     `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (cc *CustomCluster) SetDType() {
	cc.DType = []string{DType}
}

type Cluster struct {
	UID          string       `json:"uid,omitempty"`
	Type         ClusterType  `json:"Cluster.type,omitempty"`
	AddressCount *int         `json:"Cluster.addressCount,omitempty"`
	Transaction  db.UIDNode   `json:"Cluster.transaction,omitempty"`
	Children     []db.UIDNode `json:"Cluster.children,omitempty"`
	Addresses    []db.UIDNode `json:"Cluster.addresses,omitempty"`
	DType        []string     `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (c *Cluster) SetDType() {
	c.DType = []string{DType}
}

func NewFMICluster(index int) *Cluster {
	return &Cluster{
		UID:   "_:c" + strconv.Itoa(index),
		Type:  TypeFMI,
		DType: []string{DType},
	}
}

func NewFMIClusterByUID(uid string) *Cluster {
	return &Cluster{
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

type BasicCluster struct {
	UID          string `json:"uid"`
	AddressCount int    `json:"Cluster.addressCount,omitempty"`
}

type AddressWithClusters struct {
	UID      string              `json:"uid"`
	Clusters []ClusterWithParent `json:"clusters,omitempty"`
}

type AddressWithCluster struct {
	UID     string        `json:"uid"`
	Cluster *BasicCluster `json:"clusters,omitempty"`
}

type TransactionWithAddressClusters struct {
	UID       string                `json:"uid"`
	Addresses []AddressWithClusters `json:"addr,omitempty"`
}

type TransactionWithInputOutputAddressCluster struct {
	UID             string               `json:"uid"`
	Type            string               `json:"Transaction.type,omitempty"`
	InputAddresses  []AddressWithCluster `json:"input_addr,omitempty"`
	OutputAddresses []AddressWithCluster `json:"output_addr,omitempty"`
}

type TransactionWithAddresses struct {
	UID       string       `json:"uid"`
	Addresses []db.UIDNode `json:"addr,omitempty"`
}

type TransactionWithInputOutputAddresses struct {
	UID             string       `json:"uid"`
	Type            string       `json:"Transaction.type,omitempty"`
	InputAddresses  []db.UIDNode `json:"input_addr,omitempty"`
	OutputAddresses []db.UIDNode `json:"output_addr,omitempty"`
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

type FrontendUserCluster struct {
	UID          string   `json:"uid,omitempty"`
	Timestamp    string   `json:"ts,omitempty"`
	AddressCount int64    `json:"address_count,omitempty"`
	Addresses    []string `json:"addresses,omitempty"`
}
