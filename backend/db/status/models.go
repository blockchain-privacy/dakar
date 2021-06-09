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

type FrontendStatus struct {
	IsCrawling            bool   `json:"iscrawling"`
	IsClassifying         bool   `json:"isclassifying"`
	LastBlockID           uint64 `json:"lastblockid"`
	LowestBlockID         uint64 `json:"lowestblockid"`
	LastClassifiedBlockID uint64 `json:"lastclassifiedid"`
}

func (v FrontendStatus) String() string {
	return fmt.Sprintf("IsCrawling: %t, IsClassifying: %t, LastBlockID: %d, "+
		"LastClassifiedBlockID: %d",
		v.IsCrawling, v.IsClassifying, v.LastBlockID, v.LastClassifiedBlockID)
}

var (
	ErrorStatusNotFound                = errors.New("no status found")
	ErrorInvalidNumber                 = errors.New("wrong number of status objects returned")
	ErrorLastBlockIDNotFound           = errors.New("last block id not found")
	ErrorIsCrawlingNotFound            = errors.New("crawler status not found")
	ErrorIsClassifyingNotFound         = errors.New("classifier status not found")
	ErrorLastClassifiedBlockIDNotFound = errors.New("block id of last classified block not found")
	ErrorTopBlockNotFound              = errors.New("top block not found")
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidNumber)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidNumber)
		return
	}

	status = a.Q[0]
	return
}
