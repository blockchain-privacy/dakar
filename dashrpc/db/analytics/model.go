package analytics

import (
	"fmt"
)

type Origin struct {
	Uid            string
	Hash           string
	BlockHash      string
	BlockId        uint64
	BlockTimestamp string
}

func (o Origin) String() string {
	return fmt.Sprintf("Uid: %s, Hash: %s, block hash: %s, block id: %d, timestamp: %s",
		o.Uid, o.Hash, o.BlockHash, o.BlockId, o.BlockTimestamp)
}

type input struct {
	Uid         string       `json:"uid,omitempty"`
	Transaction *transaction `json:"~tx_outputs,omitempty"`
}

type transaction struct {
	Uid   string `json:"uid,omitempty"`
	Input *input `json:"tx_inputs,omitempty"`
}

type PathElement struct {
	uid           string
	isTransaction bool
}
