package block

import (
	tx "dashrpc/db/transaction"
	"errors"
	"fmt"
)

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

type blockQuery struct {
	Q []Block `json:"q"`
}

func (bq blockQuery) payload() (blk Block, err error) {
	lenQ := len(bq.Q)

	if lenQ == 0 {
		err = errors.New("no blocks found")
		return blk, err
	} else if lenQ > 1 {
		// found more than one transaction, which should not be possible
		err = errors.New("found more than one block")
		return blk, err
	}
	blk = bq.Q[0]
	return blk, err
}
