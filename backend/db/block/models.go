package block

import (
	tx "backend/db/transaction"

	"errors"
	"fmt"
)

// DType is the dgraph database type for the Block type
const DType = "Block"

var (
	// ErrorBlockNotFound is returned if no block was found
	ErrorBlockNotFound = errors.New("no block found")
	// ErrorInvalidResult is returned if an invalid result was returned from the database
	ErrorInvalidResult = errors.New("invalid result")
)

// Block is the database representation of a block
type Block struct {
	UID          string           `json:"uid,omitempty"`
	Hash         string           `json:"blockhash,omitempty"`
	ID           *uint64          `json:"id,omitempty"`
	Timestamp    string           `json:"ts,omitempty"`
	PrevBlock    *Block           `json:"prevblock,omitempty"`
	Transactions []tx.Transaction `json:"transactions,omitempty"`
	DType        []string         `json:"dgraph.type,omitempty"`
}

func (b Block) String() string {
	output := fmt.Sprintf("UID: %s, Hash: %s, Timestamp: %s", b.UID, b.Hash, b.Timestamp)

	if b.ID != nil {
		output += fmt.Sprintf(", ID: %d", *b.ID)
	}

	if b.PrevBlock != nil {
		output += fmt.Sprintf(", PrevBlockHash: %s", b.PrevBlock.Hash)
	}

	if b.Transactions != nil {
		output += fmt.Sprintf(", TransactionCount: %d", len(b.Transactions))
	}

	return output
}

// SetDType sets the DType for dgraph type recognition
func (b *Block) SetDType() {
	b.DType = []string{DType}
}

// IsComplete checks if the given block has all attributes filled
func (b Block) IsComplete() bool {
	return b.UID != "" && b.Hash != "" && b.ID != nil && b.Timestamp != "" &&
		b.DType != nil && b.Transactions != nil && b.PrevBlock != nil
}

// FrontendBlock holds all block data which is exposed to the frontend
type FrontendBlock struct {
	Hash          string   `json:"blockhash,omitempty"`
	ID            uint64   `json:"id,omitempty"`
	Timestamp     string   `json:"ts,omitempty"`
	PrevBlockHash string   `json:"prevblockhash,omitempty"`
	NextBlockHash string   `json:"nextblockhash,omitempty"`
	Transactions  []string `json:"txhashes,omitempty"`
}

func (v FrontendBlock) String() string {
	output := fmt.Sprintf("ID: %d, Hash: %s, Timestamp: %s, "+
		"PrevBlockHash: %s, NextBlockHash: %s, transaction count: %d",
		v.ID, v.Hash, v.Timestamp, v.PrevBlockHash, v.NextBlockHash, len(v.Transactions))

	return output
}

type blockQuery struct {
	Q []Block `json:"q"`
}

func (bq blockQuery) payload() (blk Block, err error) {
	lenQ := len(bq.Q)

	if lenQ == 0 {
		err = errors.New("no blocks found")
		return
	} else if lenQ > 1 {
		// found more than one block, which should not be possible
		err = errors.New("found more than one block")
		return
	}
	blk = bq.Q[0]
	return
}
