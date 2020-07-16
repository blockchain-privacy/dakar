package db

type Block struct {
	Uid       string `json:"uid,omitempty"`
	Hash      string `json:"blockhash,omitempty"`
	Id        uint64 `json:"id,omitempty"`
	Timestamp string `json:"ts,omitempty"`
	PrevBlock *Block `json:"prevblock,omitempty"`
	//NextBlock    *Block         `json:"nextblock,omitempty"`
	Transactions []*Transaction `json:"transactions,omitempty"`
	DType        []string       `json:"dgraph.type,omitempty"`
}

type Transaction struct {
	Uid       string     `json:"uid,omitempty"`
	Outputs   []TxOutput `json:"outputs,omitempty"`
	Inputs    []TxOutput `json:"inputs,omitempty"`
	Hash      string     `json:"txhash,omitempty"`
	Block     *Block     `json:"block,omitempty"`
	Timestamp string     `json:"ts,omitempty"`
	DType     []string   `json:"dgraph.type,omitempty"`
}

type TxOutput struct {
	Uid        string    `json:"uid,omitempty"`
	Index      int       `json:"index"`
	TxType     string    `json:"txtype,omitempty"`
	TxHash     string    `json:"txhash,omitempty"`
	Amount     float64   `json:"amount"`
	IsCoinbase bool      `json:"iscoinbase"`
	DType      []string  `json:"dgraph.type,omitempty"`
	Addresses  []Address `json:"addresses,omitempty"`
}

type Address struct {
	Uid  string `json:"uid,omitempty"`
	Hash string `json:"addresshash,omitempty"`
}

type blockQuery struct {
	Q []Block `json:"q"`
}

type transactionQuery struct {
	Q []Transaction `json:"q"`
}
