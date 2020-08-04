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
	Uid string `json:"uid,omitempty"`

	// true if a crawling process is currently active
	IsCrawling *bool `json:"iscrawling,omitempty"`

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

type VerboseStatus struct {
	Uid              string
	IsCrawling       bool
	LastBlockId      uint64
	HighestBlockId   uint64
	TransactionCount uint64
	BlockCount       uint64
	OutputCount      uint64
	AddressCount     uint64
}

func (v VerboseStatus) String() string {
	return fmt.Sprintf("Uid: %s, IsCrawling: %t, LastBlockId: %d, HighestBlockId: %d, "+
		"TransactionCount: %d, BlockCount: %d, OutputCount: %d, AddressCount: %d",
		v.Uid, v.IsCrawling, v.LastBlockId, v.HighestBlockId, v.TransactionCount, v.BlockCount, v.OutputCount, v.AddressCount)
}

const (
	ErrorStatusNotFound       = "no status found"
	ErrorInvalidNumber        = "wrong number of status objects returned"
	ErrorLastBlockIdNotFound  = "last block id not found"
	ErrorIsCrawlingNotFound   = "crawling status not found"
	ErrorHighestBlockNotFound = "highest block not found"
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
