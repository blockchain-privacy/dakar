package transaction

type Transaction struct {
	Uid     string     `json:"uid,omitempty"`
	Outputs []TxOutput `json:"tx_outputs,omitempty"`
	Inputs  []TxOutput `json:"tx_inputs,omitempty"`
	Hash    string     `json:"txhash,omitempty"`
	DType   []string   `json:"dgraph.type,omitempty"`
}

type TxOutput struct {
	Uid        string   `json:"uid,omitempty"`
	Index      string   `json:"index,omitempty"`
	TxType     string   `json:"txtype,omitempty"`
	Amount     string   `json:"amount,omitempty"`
	IsCoinbase string   `json:"iscoinbase,omitempty"`
	DType      []string `json:"dgraph.type,omitempty"`
}

type transactionQuery struct {
	Q []Transaction `json:"q"`
}
