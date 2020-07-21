package block

import (
	tx "dashrpc/db/transaction"
	"fmt"
)

type blockQuery struct {
	Q []Block `json:"q"`
}

type Block struct {
	Uid          string           `json:"uid,omitempty"`
	Hash         string           `json:"blockhash,omitempty"`
	Id           *uint64          `json:"id,omitempty"`
	Timestamp    string           `json:"ts,omitempty"`
	PrevBlock    *Block           `json:"prevblock,omitempty"`
	Transactions []tx.Transaction `json:"transactions,omitempty"`
	DType        []string         `json:"dgraph.type,omitempty"`
}

type UpsertBlockData struct {
	Uid          string           `json:"uid,omitempty"`
	Hash         string           `json:"blockhash,omitempty"`
	Id           *uint64          `json:"id,omitempty"`
	Timestamp    string           `json:"ts,omitempty"`
	PrevBlock    *Block           `json:"prevblock,omitempty"`
	Transactions []tx.Transaction `json:"transactions,omitempty"`
	DType        []string         `json:"dgraph.type,omitempty"`
}

func (b Block) String() string {
	return fmt.Sprintf("Hash: %s, Id: %s, Timestamp: %s", b.Hash, b.Id, b.Timestamp)
}
