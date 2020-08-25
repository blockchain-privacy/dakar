package status

import (
	"dashrpc/cmd/cliutil"
	"errors"
	"fmt"
)

const CrawlerStatusDType = "CrawlerStatus"
const AnalyzerStatusDType = "AnalyzerStatus"

type CrawlerStatus struct {
	Uid string `json:"uid,omitempty"`

	// true if a crawling process is currently active
	IsCrawling *bool `json:"iscrawling,omitempty"`

	// Crawling works in three steps:
	// Step 1: Insert block, transaction and output data.
	// Step 2: Connect the outputs of the block with addresses.
	// Step 3: Set LastBlockId to the id of the processed block.
	// LastBlockId is the ID of the last block where both steps have been successful.
	// Thus, it is possible to have a block inside the database with a higher ID
	//than LastBlockId, if the crawling was aborted between step 1 and 2.
	LastBlockId   *uint64 `json:"lastblockid,omitempty"`
	LowestBlockId *uint64 `json:"lowestblockid,omitempty"`

	DType []string `json:"dgraph.type,omitempty"`
}

func (c CrawlerStatus) String() string {
	output := fmt.Sprintf("Uid: %s", c.Uid)

	if c.IsCrawling != nil {
		output += fmt.Sprintf(", IsCrawling: %t", *c.IsCrawling)
	}

	if c.LastBlockId != nil {
		output += fmt.Sprintf(", LastBlockId: %d", *c.LastBlockId)
	}

	return output
}

func (c *CrawlerStatus) SetDType() {
	c.DType = []string{CrawlerStatusDType}
}

type AnalyzerStatus struct {
	Uid string `json:"uid,omitempty"`

	// true if a analyze process is currently active
	IsAnalyzing *bool `json:"isanalyzing,omitempty"`

	// The id of the last completely analysed block
	LastAnalysedBlockId *uint64  `json:"lastanalysedid,omitempty"`
	DType               []string `json:"dgraph.type,omitempty"`
}

func (a AnalyzerStatus) String() string {
	output := fmt.Sprintf("Uid: %s", a.Uid)

	if a.IsAnalyzing != nil {
		output += fmt.Sprintf(", IsAnalyzing: %t", *a.IsAnalyzing)
	}

	if a.LastAnalysedBlockId != nil {
		output += fmt.Sprintf(", LastAnalysedBlockId: %d", *a.LastAnalysedBlockId)
	}

	return output
}

func (a *AnalyzerStatus) SetDType() {
	a.DType = []string{AnalyzerStatusDType}
}

type FrontendStatus struct {
	IsCrawling          bool   `json:"iscrawling"`
	IsAnalyzing         bool   `json:"isanalyzing"`
	LastBlockId         uint64 `json:"lastblockid"`
	LowestBlockId       uint64 `json:"lowestblockid"`
	LastAnalysedBlockId uint64 `json:"lastanalysedid"`
}

func (v FrontendStatus) String() string {
	return fmt.Sprintf("IsCrawling: %t, IsAnalyzing: %t, LastBlockId: %d, LastAnalysedBlockId: %d",
		v.IsCrawling, v.IsAnalyzing, v.LastBlockId, v.LastAnalysedBlockId)
}

var (
	ErrorStatusNotFound      = errors.New("no status found")
	ErrorInvalidNumber       = errors.New("wrong number of status objects returned")
	ErrorLastBlockIdNotFound = errors.New("last block id not found")
	ErrorIsCrawlingNotFound  = errors.New("crawling status not found")
	ErrorIsAnalyzingNotFound = errors.New("analyzing status not found")
	// todo
	//ErrorLastAnalysedBlockIdNotFound = errors.New("block id of last analysed block not found")
	ErrorTopBlockNotFound = errors.New("top block not found")
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

type analyzerStatusQuery struct {
	Q []AnalyzerStatus `json:"q"`
}

func (a analyzerStatusQuery) payload() (status AnalyzerStatus, err error) {
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
