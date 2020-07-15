package db

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"time"
)

type Block struct {
	Hash          chainhash.Hash `json:"blockhash,omitempty"`
	Id            uint64         `json:"id,omitempty"`
	Timestamp     time.Time      `json:"ts,omitempty"`
	PrevBlockHash chainhash.Hash `json:"prevblockhash,omitempty"`
	NextBlockHash chainhash.Hash `json:"nextblockhash,omitempty"`
	TxHashes      []string       `json:"txhashes,omitempty"`
	DType         []string       `json:"dgraph.type,omitempty"`
}

type DatabaseBlock struct {
	Uid          string                 `json:"uid,omitempty"`
	Hash         string                 `json:"blockhash,omitempty"`
	Id           uint64                 `json:"id,omitempty"`
	Timestamp    string                 `json:"ts,omitempty"`
	PrevBlock    *DatabaseBlock         `json:"prevblock,omitempty"`
	Transactions []*DataBaseTransaction `json:"txhashes,omitempty"`
	DType        []string               `json:"dgraph.type,omitempty"`
}

type DataBaseTransaction struct {
	Uid  string `json:"uid,omitempty"`
	Hash string `json:"txhash,omitempty"`
}

// If omitempty is not set, then edges with empty values (0 for int/float, "" for string, false
// for bool) would be created for values not specified explicitly.

type Transaction struct {
	Uid   string   `json:"uid,omitempty"`
	Hash  string   `json:"hash,omitempty"`
	Block Block    `json:"block,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

type blockQuery struct {
	Q []DatabaseBlock `json:"q"`
}
