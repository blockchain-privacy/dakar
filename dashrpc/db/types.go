package db

type Block struct {
	Uid          string         `json:"uid,omitempty"`
	Hash         string         `json:"blockhash,omitempty"`
	Id           string         `json:"id,omitempty"`
	Timestamp    string         `json:"ts,omitempty"`
	PrevBlock    *Block         `json:"prevblock,omitempty"`
	Transactions []*Transaction `json:"transactions,omitempty"`
	DType        []string       `json:"dgraph.type,omitempty"`
}

type Transaction struct {
	Uid       string     `json:"uid,omitempty"`
	Outputs   []TxOutput `json:"outputs,omitempty"`
	Inputs    []TxOutput `json:"inputs,omitempty"`
	Hash      string     `json:"txhash,omitempty"`
	Timestamp string     `json:"ts,omitempty"`
	DType     []string   `json:"dgraph.type,omitempty"`
}

type TxOutput struct {
	Uid        string    `json:"uid,omitempty"`
	Index      string    `json:"index,omitempty"`
	TxType     string    `json:"txtype,omitempty"`
	Amount     string    `json:"amount,omitempty"`
	IsCoinbase string    `json:"iscoinbase,omitempty"`
	Addresses  []Address `json:"addresses,omitempty"`
	DType      []string  `json:"dgraph.type,omitempty"`
}

type Address struct {
	Uid     string     `json:"uid,omitempty"`
	Hash    string     `json:"addresshash,omitempty"`
	Outputs []TxOutput `json:"outputs,omitempty"`
	DType   []string   `json:"dgraph.type,omitempty"`
}

type blockQuery struct {
	Q []Block `json:"q"`
}

type transactionQuery struct {
	Q []Transaction `json:"q"`
}

type addressQuery struct {
	Q []Address `json:"q"`
}
