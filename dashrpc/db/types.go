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
	Uid   string   `json:"uid,omitempty"`
	Hash  string   `json:"txhash,omitempty"`
	Block *Block   `json:"block,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

type blockQuery struct {
	Q []Block `json:"q"`
}
