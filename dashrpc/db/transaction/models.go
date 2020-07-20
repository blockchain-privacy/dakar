package transaction

import op "dashrpc/db/output"

type Transaction struct {
	Uid     string      `json:"uid,omitempty"`
	Outputs []op.Output `json:"tx_outputs,omitempty"`
	Inputs  []op.Output `json:"tx_inputs,omitempty"`
	Hash    string      `json:"txhash,omitempty"`
	DType   []string    `json:"dgraph.type,omitempty"`
}

type transactionQuery struct {
	Q []Transaction `json:"q"`
}
