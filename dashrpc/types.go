package dashrpc

import "github.com/btcsuite/btcd/chaincfg/chainhash"

const ChainType_Block = "blk"
const ChainType_TxDetails = "txd"
//const ChainType_TxOutput = "txo"
const ChainType_AddrOutputs = "ado"

const DB_BLOCK_COUNT = "DB_BLOCK_COUNT"


// ChainItem represents a generic blockchain item
type ChainItem struct {
	ItemType string
	Data []byte
}

// Block represents a simple block
type Block struct {
	Hash chainhash.Hash
	PrevBlockHash chainhash.Hash
	NextBlockHash chainhash.Hash
	TxHashes []string
}

// TxTransfer represents simple value transfer from/to address(es)
type TxOutput struct {
	TxHash string
	TxType string
	Amount float64
	// For UTXO this will not be known
	Addresses []string
	// For TXO this represents the index of the output in the transaction
	Index int
	IsCoinbase bool
}

// TxDetails represents transaction data
type TxDetails struct {
	Hash string
	Inputs []TxOutput
	Outputs []TxOutput
	Timestamp int64
}

