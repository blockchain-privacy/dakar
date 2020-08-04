package transaction

import (
	op "dashrpc/db/output"
	"errors"
	"fmt"
)

const DType = "Transaction"

type Transaction struct {
	Uid     string      `json:"uid,omitempty"`
	Outputs []op.Output `json:"tx_outputs,omitempty"`
	Inputs  []op.Output `json:"tx_inputs,omitempty"`
	Hash    string      `json:"txhash,omitempty"`
	DType   []string    `json:"dgraph.type,omitempty"`
}

func (t Transaction) String() string {
	output := fmt.Sprintf("Uid: %s, Hash: %s", t.Uid, t.Hash)

	if t.Outputs != nil {
		output += fmt.Sprintf(", OutputCount: %d", len(t.Outputs))
	}

	if t.Inputs != nil {
		output += fmt.Sprintf(", InputCount: %d", len(t.Inputs))
	}

	return output
}

func (t *Transaction) SetDType() {
	t.DType = []string{DType}
}

// converts a Transaction struct to an updateTransactionData struct
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

// checks if the given transaction has all attributes filled
func (t Transaction) IsComplete() bool {
	return t.Uid != "" && t.Hash != "" && t.DType != nil
}

func (t Transaction) CountInputDenominations() []int {
	return op.CountDenominations(t.Inputs)
}

func (t Transaction) CountOutputDenominations() []int {
	return op.CountDenominations(t.Outputs)
}

// IsCreateDenominations checks if the TX creates denominations
func (t Transaction) IsCreateDenominations() bool {
	denom := t.CountOutputDenominations()
	// todo add fourth denomination?
	return len(t.Inputs) == 1 &&
		(denom[0] > 2 || denom[1] > 2 || denom[2] > 2)
}

func (t Transaction) IsPrivateSend() bool {
	denom := t.CountInputDenominations()
	// todo add fourth denomination?
	return len(t.Outputs) == 1 &&
		(denom[0] > 2 || denom[1] > 2 || denom[2] > 2)
}

// IsOneOrTwoOutputs checks if TX has only 1 or 2 outputs. Used for clustering.
func (t Transaction) IsOneOrTwoOutput() bool {
	return !t.IsMixing() &&
		(len(t.Outputs) == 2 || len(t.Outputs) == 1)
}

// IsMixing checks if TX is mixing
func (t Transaction) IsMixing() bool {
	if len(t.Inputs) != len(t.Outputs) {
		return false
	}
	denomIn := t.CountInputDenominations()
	denomOut := t.CountOutputDenominations()
	sum := 0
	for _, v := range denomIn {
		sum += v
	}
	if sum == 0 {
		return false
	}
	sum = 0
	for _, v := range denomIn {
		sum += v
	}
	if sum == 0 {
		return false
	}
	for i := range denomIn {
		if denomIn[i] != denomOut[i] {
			return false
		}
	}
	return true
}

// This struct is needed to internally convert floats to strings and backwards.
// That is, because the precision of the float type of Dgraph is to low.
type updateTransactionData struct {
	Uid     string                `json:"uid,omitempty"`
	Outputs []op.UpdateOutputData `json:"tx_outputs,omitempty"`
	Inputs  []op.UpdateOutputData `json:"tx_inputs,omitempty"`
	Hash    string                `json:"txhash,omitempty"`
	DType   []string              `json:"dgraph.type,omitempty"`
}

// converts an updateTransactionData struct to a Transaction struct
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

type transactionQuery struct {
	Q []updateTransactionData `json:"q"`
}

func (tq transactionQuery) payload() (tx Transaction, err error) {
	lenQ := len(tq.Q)

	if lenQ == 0 {
		err = errors.New("no transactions found")
		return tx, err
	} else if lenQ > 1 {
		// found more than one transaction, which should not be possible
		err = errors.New("found more than one transaction")
		return tx, err
	}

	return tq.Q[0].toTransaction()
}
