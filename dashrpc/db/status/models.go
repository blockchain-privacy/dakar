package status

import (
	"errors"
	"fmt"
)

const DType = "Status"

type Status struct {
	Uid string `json:"uid,omitempty"`

	// true if a crawling process is currently active
	IsCrawling *bool `json:"iscrawling,omitempty"`

	// Crawling works in two steps:
	// Step 1: Insert block, transaction and output data.
	// Step 2: Connect the outputs of the block with addresses.
	// LastBlockId is the ID of the last block where both steps have been successful.
	// Thus, it is possible to have a Block with a higher ID than LastBlockId,
	// if the crawling was aborted between step 1 and 2.
	LastBlockId   *uint64  `json:"lastblockid,omitempty"`
	LowestBlockId *uint64  `json:"lowestblockid,omitempty"`
	DType         []string `json:"dgraph.type,omitempty"`
}

func (s Status) String() string {
	output := fmt.Sprintf("Uid: %s", s.Uid)

	if s.IsCrawling != nil {
		output += fmt.Sprintf(", IsCrawling: %t", *s.IsCrawling)
	}

	if s.LastBlockId != nil {
		output += fmt.Sprintf(", LastBlockId: %d", *s.LastBlockId)
	}

	return output
}

func (s *Status) SetDType() {
	s.DType = []string{DType}
}

type VerboseStatus struct {
	Uid              string `json:"uid"`
	IsCrawling       bool   `json:"iscrawling"`
	LastBlockId      uint64 `json:"lastblockid"`
	LowestBlockId    uint64 `json:"lowestblockid"`
	TransactionCount uint64 `json:"txcount"`
	BlockCount       uint64 `json:"blkcount"`
	OutputCount      uint64 `json:"outputcount"`
	AddressCount     uint64 `json:"addresscount"`
}

func (v VerboseStatus) String() string {
	return fmt.Sprintf("Uid: %s, IsCrawling: %t, LastBlockId: %d, "+
		"TransactionCount: %d, BlockCount: %d, OutputCount: %d, AddressCount: %d",
		v.Uid, v.IsCrawling, v.LastBlockId, v.TransactionCount, v.BlockCount, v.OutputCount, v.AddressCount)
}

const (
	ErrorStatusNotFound      = "no status found"
	ErrorInvalidNumber       = "wrong number of status objects returned"
	ErrorLastBlockIdNotFound = "last block id not found"
	ErrorIsCrawlingNotFound  = "crawling status not found"
	ErrorTopBlockNotFound    = "top block not found"
)

type statusQuery struct {
	Q []Status `json:"q"`
}

func (s statusQuery) payload() (status Status, err error) {
	lenQ := len(s.Q)

	if lenQ == 0 {
		err = errors.New(ErrorStatusNotFound)
		return
	}

	if lenQ > 1 {
		err = errors.New(ErrorInvalidNumber)
		return
	}

	status = s.Q[0]
	return
}
