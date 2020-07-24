package transaction

import (
	op "dashrpc/db/output"
	"errors"
	"fmt"
	"math"
)

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
	t.DType = []string{"Transaction"}
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

// IsCreateDenominations checks if the TX creates denominations
func (t Transaction) IsCreateDenominations() bool {
	denom := countDenominations(t.Outputs)
	// todo add fourth denomination?
	return len(t.Inputs) == 1 &&
		(denom[0] > 2 || denom[1] > 2 || denom[2] > 2)
}

func (t Transaction) IsPrivateSend() bool {
	denom := countDenominations(t.Inputs)
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
	denomIn := countDenominations(t.Inputs)
	denomOut := countDenominations(t.Outputs)
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

// This struct is needed to interally convert floats to strings and backwards.
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

func almostEqual(a, b float64) bool {
	var delta float64
	delta = 0.00001
	return math.Abs(a-b) <= delta
}

func countDenominations(outputs []op.Output) []int {
	denominationsTypes := []float64{1.00001, 0.100001, 0.0100001, 0.00100001}
	denominations := make([]int, len(denominationsTypes))

	for _, o := range outputs {
	inner:
		for i, v := range denominationsTypes {
			if almostEqual(*o.Amount, v) {
				denominations[i]++
				break inner
			}
		}
	}

	return denominations
}
