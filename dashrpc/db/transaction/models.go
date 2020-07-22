package transaction

import op "dashrpc/db/output"

type Transaction struct {
	Uid     string      `json:"uid,omitempty"`
	Outputs []op.Output `json:"tx_outputs,omitempty"`
	Inputs  []op.Output `json:"tx_inputs,omitempty"`
	Hash    string      `json:"txhash,omitempty"`
	DType   []string    `json:"dgraph.type,omitempty"`
}

func (t updateTransactionData) toTransaction() (tx Transaction, err error) {
	tx.Uid = t.Uid
	tx.Hash = t.Hash
	tx.DType = t.DType

	for _, e := range t.Inputs {
		o, err := e.ToOutput()
		if err != nil {
			return tx, err
		}

		tx.Inputs = append(tx.Inputs, o)
	}

	for _, e := range t.Outputs {
		o, err := e.ToOutput()
		if err != nil {
			return tx, err
		}

		tx.Outputs = append(tx.Outputs, o)
	}

	return tx, err
}

// This struct is needed to interally convert floats to strings and backwards.
// That is, because the precision of the float type of Dgraph is to low.
type updateTransactionData struct {
	Uid     string                `json:"uid,omitempty"`
	Outputs []op.UpdateOutputData `json:"tx_outputs,omitempty"`
	Inputs  []op.UpdateOutputData `json:"tx_inputs,omitempty"`
	Hash    string                `json:"txhash,omitempty"`
	DType   []string              `json:"dgraph.type,omitempty"`
}

func (t Transaction) toUpdate() (tx updateTransactionData) {
	tx.Uid = t.Uid
	tx.Hash = t.Hash
	tx.DType = t.DType

	for _, e := range t.Inputs {
		tx.Inputs = append(tx.Inputs, e.ToUpdate())
	}

	for _, e := range t.Outputs {
		tx.Outputs = append(tx.Outputs, e.ToUpdate())
	}

	return tx
}

type transactionQuery struct {
	Q []updateTransactionData `json:"q"`
}
