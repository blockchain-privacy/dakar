package dashrpc

const ChainType_Block = "blk"
const ChainType_TxDetails = "txd"
const ChainType_TxtTransfer = "txt"

const DB_BLOCK_COUNT = "DB_BLOCK_COUNT"


// ChainItem represents a generic blockchain item
type ChainItem struct {
	ChainType string
	Data []byte
}

// Block represents a simple block
type Block struct {
	Hash string
	PrevBlockHash string
	NextBlockHash string
	TxHashes []string
}

// TxTransfer represents simple value transfer from/to address(es)
type TxTransfer struct {
	Details *TxDetails
	Amount float64
	Addresses []string
	Index uint
}

// TxDetails represents transaction data
type TxDetails struct {
	Hash string
	Inputs []TxTransfer
	Outputs []TxTransfer
	IsCoinbase bool
}

