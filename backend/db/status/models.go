// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package status

import (
	"gitlab.com/blockchain-privacy/gomisc/serror"

	"errors"
	"fmt"
)

// CrawlerStatusDType is the dgraph database type for the CrawlerStatus type
const CrawlerStatusDType = "CrawlerStatus"

// ClassifierStatusDType is the dgraph database type for the ClassifierStatus type
const ClassifierStatusDType = "ClassifierStatus"

// ClusteringFlatMultiInputDType is the dgraph database type for the CFMIStatus type
const ClusteringFlatMultiInputDType = "CFMIStatus"

// MetaDType is the dgraph database type for the Meta type
const MetaDType = "Meta"

// CrawlerStatus is the database representation of the crawler status
type CrawlerStatus struct {
	UID string `json:"uid,omitempty"`

	// true if a crawling process is currently active
	IsCrawling *bool `json:"iscrawling,omitempty"`

	// Crawling works in three steps:
	// Step 1: Insert block, transaction and output data.
	// Step 2: Connect the outputs of the block with addresses.
	// Step 3: Set LastBlockID to the id of the processed block.
	// LastBlockID is the ID of the last block where both steps have been successful.
	// Thus, it is possible to have a block inside the database with a higher ID
	// than LastBlockID, if the crawling was aborted between step 1 and 2.
	LastBlockID *int64 `json:"lastblockid,omitempty"`

	DType []string `json:"dgraph.type,omitempty"`
}

func (c *CrawlerStatus) String() string {
	output := "UID: " + c.UID

	if c.IsCrawling != nil {
		output += fmt.Sprintf(", IsCrawling: %t", *c.IsCrawling)
	}

	if c.LastBlockID != nil {
		output += fmt.Sprintf(", LastBlockID: %d", *c.LastBlockID)
	}

	return output
}

// SetDType sets the DType for dgraph type recognition
func (c *CrawlerStatus) SetDType() {
	c.DType = []string{CrawlerStatusDType}
}

// ClassifierStatus is the database representation of the classifier status
type ClassifierStatus struct {
	UID string `json:"uid,omitempty"`

	// IsClassifying is true if a classifier process is currently active
	IsClassifying *bool `json:"isclassifying,omitempty"`

	// LastClassifiedBlockID is the id of the last completely classified block
	LastClassifiedBlockID *int64   `json:"lastclassifiedid,omitempty"`
	DType                 []string `json:"dgraph.type,omitempty"`
}

func (c *ClassifierStatus) String() string {
	output := "UID: " + c.UID

	if c.IsClassifying != nil {
		output += fmt.Sprintf(", IsClassifying: %t", *c.IsClassifying)
	}

	if c.LastClassifiedBlockID != nil {
		output += fmt.Sprintf(", LastClassifiedBlockID: %d", *c.LastClassifiedBlockID)
	}

	return output
}

// SetDType sets the DType for dgraph type recognition
func (c *ClassifierStatus) SetDType() {
	c.DType = []string{ClassifierStatusDType}
}

// ClusteringFlatMultiInputStatus is the database representation of the flat multi-input clustering status
type ClusteringFlatMultiInputStatus struct {
	UID string `json:"uid,omitempty"`

	// IsClustering is true if a multi-input clustering process is currently active
	IsClustering *bool `json:"isclustering,omitempty"`

	// LastClusteredBlockID is the id of the last completely multi-input clustered block
	LastClusteredBlockID *int64   `json:"lastclusteredid,omitempty"`
	DType                []string `json:"dgraph.type,omitempty"`
}

func (c *ClusteringFlatMultiInputStatus) String() string {
	output := "UID: " + c.UID

	if c.IsClustering != nil {
		output += fmt.Sprintf(", IsClustering: %t", *c.IsClustering)
	}

	if c.LastClusteredBlockID != nil {
		output += fmt.Sprintf(", LastClusteredBlockID: %d", *c.LastClusteredBlockID)
	}

	return output
}

// SetDType sets the DType for dgraph type recognition
func (c *ClusteringFlatMultiInputStatus) SetDType() {
	c.DType = []string{ClusteringFlatMultiInputDType}
}

// FrontendStatus is the frontend representation of the crawler status
type FrontendStatus struct {
	IsCrawling              *bool  `json:"iscrawling,omitempty"`
	IsClassifying           *bool  `json:"isclassifying,omitempty"`
	IsClusteringFMI         *bool  `json:"isclusteringfmi,omitempty"`
	LastBlockID             *int64 `json:"lastblockid,omitempty"`
	LastClassifiedBlockID   *int64 `json:"lastclassifiedid,omitempty"`
	LastClusteredFMIBlockID *int64 `json:"lastclusteredfmiid,omitempty"`
}

var (
	// ErrStatusNotFound is returned if the status has not been set yet
	ErrStatusNotFound   = errors.New("no status found")
	errInvalidNumber    = errors.New("wrong number of status objects returned")
	errTopBlockNotFound = errors.New("top block not found")
)

type crawlerStatusQuery struct {
	Q []CrawlerStatus `json:"q"`
}

func (c crawlerStatusQuery) payload() (status CrawlerStatus, err error) {
	lenQ := len(c.Q)

	if lenQ == 0 {
		err = serror.New(ErrStatusNotFound)
		return
	}

	if lenQ > 1 {
		err = serror.New(errInvalidNumber)
		return
	}

	status = c.Q[0]
	return
}

type classifierStatusQuery struct {
	Q []ClassifierStatus `json:"q"`
}

func (a classifierStatusQuery) payload() (status ClassifierStatus, err error) {
	lenQ := len(a.Q)

	if lenQ == 0 {
		err = serror.New(ErrStatusNotFound)
		return
	}

	if lenQ > 1 {
		err = serror.New(errInvalidNumber)
		return
	}

	status = a.Q[0]
	return
}

type clusteringFMIStatusQuery struct {
	Q []ClusteringFlatMultiInputStatus `json:"q"`
}

func (a clusteringFMIStatusQuery) payload() (status ClusteringFlatMultiInputStatus, err error) {
	lenQ := len(a.Q)

	if lenQ == 0 {
		err = serror.New(ErrStatusNotFound)
		return
	}

	if lenQ > 1 {
		err = serror.New(errInvalidNumber)
		return
	}

	status = a.Q[0]
	return
}

// Meta holds values regarding metadata of the database
type Meta struct {
	UID string `json:"uid,omitempty"`

	// CreationTime is the time when this database was initialized
	CreationTime string `json:"Meta.creationTime,omitempty"`
	// BlockchainMode of this database
	BlockchainMode string `json:"Meta.blockchainMode,omitempty"`
	// SchemaVersion  of this database
	SchemaVersion *int `json:"Meta.schemaVersion,omitempty"`

	DType []string `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (m *Meta) SetDType() {
	m.DType = []string{MetaDType}
}

type metaQuery struct {
	Q []Meta `json:"q"`
}

func (m metaQuery) payload() (meta Meta, err error) {
	lenQ := len(m.Q)

	if lenQ == 0 {
		err = serror.New(ErrStatusNotFound)
		return
	}

	if lenQ > 1 {
		err = serror.New(errInvalidNumber)
		return
	}

	meta = m.Q[0]
	return
}
