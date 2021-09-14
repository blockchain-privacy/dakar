package status

import (
	"backend/cmd/cliutil"

	"errors"
	"fmt"
)

// CrawlerStatusDType is the dgraph database type for the CrawlerStatus type
const CrawlerStatusDType = "CrawlerStatus"

// ClassifierStatusDType is the dgraph database type for the ClassifierStatus type
const ClassifierStatusDType = "ClassifierStatus"

// ClusteringHierarchicalMultiInputDType is the dgraph database type for the CHMIStatus type
const ClusteringHierarchicalMultiInputDType = "CHMIStatus"

// ClusteringFlatMultiInputDType is the dgraph database type for the CFMIStatus type
const ClusteringFlatMultiInputDType = "CFMIStatus"

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
	//than LastBlockID, if the crawling was aborted between step 1 and 2.
	LastBlockID   *uint64 `json:"lastblockid,omitempty"`
	LowestBlockID *uint64 `json:"lowestblockid,omitempty"`

	DType []string `json:"dgraph.type,omitempty"`
}

func (c *CrawlerStatus) String() string {
	output := fmt.Sprintf("UID: %s", c.UID)

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
	LastClassifiedBlockID *uint64  `json:"lastclassifiedid,omitempty"`
	DType                 []string `json:"dgraph.type,omitempty"`
}

func (c *ClassifierStatus) String() string {
	output := fmt.Sprintf("UID: %s", c.UID)

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

// ClusteringHierarchicalMultiInputStatus is the database representation of the hierarchical multi-input clustering status
type ClusteringHierarchicalMultiInputStatus struct {
	UID string `json:"uid,omitempty"`

	// IsClustering is true if a multi-input clustering process is currently active
	IsClustering *bool `json:"isclustering,omitempty"`

	// LastClusteredBlockID is the id of the last completely multi-input clustered block
	LastClusteredBlockID *uint64  `json:"lastclusteredid,omitempty"`
	DType                []string `json:"dgraph.type,omitempty"`
}

func (c *ClusteringHierarchicalMultiInputStatus) String() string {
	output := fmt.Sprintf("UID: %s", c.UID)

	if c.IsClustering != nil {
		output += fmt.Sprintf(", IsClustering: %t", *c.IsClustering)
	}

	if c.LastClusteredBlockID != nil {
		output += fmt.Sprintf(", LastClusteredBlockID: %d", *c.LastClusteredBlockID)
	}

	return output
}

// SetDType sets the DType for dgraph type recognition
func (c *ClusteringHierarchicalMultiInputStatus) SetDType() {
	c.DType = []string{ClusteringHierarchicalMultiInputDType}
}

// ClusteringFlatMultiInputStatus is the database representation of the flat multi-input clustering status
type ClusteringFlatMultiInputStatus struct {
	UID string `json:"uid,omitempty"`

	// IsClustering is true if a multi-input clustering process is currently active
	IsClustering *bool `json:"isclustering,omitempty"`

	// LastClusteredBlockID is the id of the last completely multi-input clustered block
	LastClusteredBlockID *uint64  `json:"lastclusteredid,omitempty"`
	DType                []string `json:"dgraph.type,omitempty"`
}

func (c *ClusteringFlatMultiInputStatus) String() string {
	output := fmt.Sprintf("UID: %s", c.UID)

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
	IsCrawling              bool   `json:"iscrawling"`
	IsClassifying           bool   `json:"isclassifying"`
	IsClusteringHMI         bool   `json:"isclusteringhmi"`
	IsClusteringFMI         bool   `json:"isclusteringfmi"`
	LastBlockID             uint64 `json:"lastblockid"`
	LowestBlockID           uint64 `json:"lowestblockid"`
	LastClassifiedBlockID   uint64 `json:"lastclassifiedid"`
	LastClusteredHMIBlockID uint64 `json:"lastclusteredhmiid"`
	LastClusteredFMIBlockID uint64 `json:"lastclusteredfmiid"`
}

func (v FrontendStatus) String() string {
	return fmt.Sprintf("IsCrawling: %t, IsClassifying: %t, IsClusteringHMI: %t, IsClusteringFMI: %t, "+
		"LastBlockID: %d, LastClassifiedBlockID: %d, LastClusteredHMIBlockID: %d, LastClusteredFMIBlockID: %d",
		v.IsCrawling, v.IsClassifying, v.IsClusteringHMI, v.IsClusteringFMI, v.LastBlockID, v.LastClassifiedBlockID,
		v.LastClusteredHMIBlockID, v.LastClusteredFMIBlockID)
}

var (
	// ErrorStatusNotFound is returned if the status has not been set yet
	ErrorStatusNotFound                          = errors.New("no status found")
	errorInvalidNumber                           = errors.New("wrong number of status objects returned")
	errorLastBlockIDNotFound                     = errors.New("last block id not found")
	errorIsCrawlingNotFound                      = errors.New("crawler status not found")
	errorIsClassifyingNotFound                   = errors.New("classifier status not found")
	errorIsClusteringMultiInputNotFound          = errors.New("multi-input clustering status not found")
	errorLastClassifiedBlockIDNotFound           = errors.New("block id of last classified block not found")
	errorLastClusteringMultiInputBlockIDNotFound = errors.New("block id of last clustered multi-input block not found")
	errorTopBlockNotFound                        = errors.New("top block not found")
)

type crawlerStatusQuery struct {
	Q []CrawlerStatus `json:"q"`
}

func (c crawlerStatusQuery) payload() (status CrawlerStatus, err error) {
	lenQ := len(c.Q)

	if lenQ == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorStatusNotFound)
		return
	}

	if lenQ > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorInvalidNumber)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorStatusNotFound)
		return
	}

	if lenQ > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorInvalidNumber)
		return
	}

	status = a.Q[0]
	return
}

type clusteringHMIStatusQuery struct {
	Q []ClusteringHierarchicalMultiInputStatus `json:"q"`
}

func (a clusteringHMIStatusQuery) payload() (status ClusteringHierarchicalMultiInputStatus, err error) {
	lenQ := len(a.Q)

	if lenQ == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorStatusNotFound)
		return
	}

	if lenQ > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorInvalidNumber)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorStatusNotFound)
		return
	}

	if lenQ > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorInvalidNumber)
		return
	}

	status = a.Q[0]
	return
}
