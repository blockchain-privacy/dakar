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

func (b Block) String() string {
	output := fmt.Sprintf("Uid: %s, Hash: %s, Timestamp: %s", b.Uid, b.Hash, b.Timestamp)

	if b.Id != nil {
		output += fmt.Sprintf(", Id: %d", *b.Id)
	}

	if b.PrevBlock != nil {
		output += fmt.Sprintf(", PrevBlockHash: %s", b.PrevBlock.Hash)
	}

	if b.Transactions != nil {
		output += fmt.Sprintf(", TransactionCount: %d", len(b.Transactions))
	}

	return output
}

// checks if the given block has all attributes filled
func (b Block) isComplete() bool {
	return b.Uid != "" && b.Hash != "" && b.Id != nil && b.Timestamp != "" ||
		b.DType != nil && b.Transactions != nil && b.PrevBlock != nil
}
