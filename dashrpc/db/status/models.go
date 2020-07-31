package status

import (
	"errors"
	"fmt"
)

// Old Meta struct // todo
//type Meta struct {
//	LastBlockId      uint64 `json:"lastblockid"`
//	StopBlockId      uint64 `json:"stopblockid"`
//	LastBlockHash    string `json:"lastblockhash"`
//	Status           string `json:"status"`
//	GlobalBlockCount uint64 `json:"globalblockcount"`
//	GlobalTxCount    uint64 `json:"globaltxcount"`
//}

const DType = "Status"

type Status struct {
	Uid        string `json:"uid,omitempty"`
	IsCrawling *bool  `json:"iscrawling,omitempty"`

	// Crawling works in two steps:
	// Step 1: Insert block, transaction and output data.
	// Step 2: Connect the outputs of the block with addresses.
	// LastBlockId is the ID of the last block where both steps have been successful.
	// Thus, it is possible to have a Block with a higher ID than LastBlockId,
	// if the crawling was aborted between step 1 and 2.
	LastBlockId *uint64  `json:"lastblockid,omitempty"`
	DType       []string `json:"dgraph.type,omitempty"`
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

type statusQuery struct {
	Q []Status `json:"q"`
}

const (
	ErrorStatusNotFound = "no status found"
	ErrorInvalidNumber  = "wrong number of status objects returned"
)

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
