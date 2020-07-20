package address

import tx "dashrpc/db/transaction"

type Address struct {
	Uid     string        `json:"uid,omitempty"`
	Hash    string        `json:"addresshash,omitempty"`
	Outputs []tx.TxOutput `json:"addr_outputs,omitempty"`
	DType   []string      `json:"dgraph.type,omitempty"`
}

type addressQuery struct {
	Q []Address `json:"q"`
}
