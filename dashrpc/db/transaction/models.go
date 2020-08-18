package transaction

import (
	"dashrpc/cmd/cliutil"
	op "dashrpc/db/output"
	"errors"
	"fmt"
)

const (
	DType               = "Transaction"
	PrivacyDenomination = "Denomination"
	PrivacyMixing       = "Mixing"
	PrivacyPrivateSend  = "PrivateSend"
)

var (
	ErrorTransactionNotFound = errors.New("no transaction found")
	ErrorInvalidResult       = errors.New("invalid result")
)

type Transaction struct {
	Uid         string      `json:"uid,omitempty"`
	PrivacyType string      `json:"privacytype,omitempty"`
	Outputs     []op.Output `json:"tx_outputs,omitempty"`
	Inputs      []op.Output `json:"tx_inputs,omitempty"`
	Hash        string      `json:"txhash,omitempty"`
	DType       []string    `json:"dgraph.type,omitempty"`
}

func (t Transaction) String() string {
	output := fmt.Sprintf("Uid: %s, Hash: %s, Privacy type: %s", t.Uid, t.Hash, t.PrivacyType)

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

func (t *Transaction) SetDenomination() {
	t.PrivacyType = PrivacyDenomination
}

func (t *Transaction) SetMixing() {
	t.PrivacyType = PrivacyMixing
}

func (t *Transaction) SetPrivateSend() {
	t.PrivacyType = PrivacyPrivateSend
}

// checks if the given transaction has all attributes filled
func (t Transaction) IsComplete() bool {
	return t.Uid != "" && t.Hash != "" && t.DType != nil
}

func (t Transaction) CountInputDenominations() []int {
	return op.CountOutputDenominations(t.Inputs)
}

func (t Transaction) CountOutputDenominations() []int {
	return op.CountOutputDenominations(t.Outputs)
}

// IsCreateDenominations checks if the TX creates denominations
func (t Transaction) IsCreateDenominations() bool {
	return len(t.Inputs) == 1 && len(t.Outputs) > 2 && IsPrivacyTransaction(t.CountOutputDenominations())
}

// IsPrivateSend checks if the TX is the end receiver of a private send transaction
func (t Transaction) IsPrivateSend() bool {
	return len(t.Outputs) == 1 && len(t.Inputs) > 2 && IsPrivacyTransaction(t.CountInputDenominations())
}

// sets the privacy type of the transaction
func (t *Transaction) SetPrivacyType() {
	if t.IsMixing() {
		t.SetMixing()
	} else if t.IsCreateDenominations() {
		t.SetDenomination()
	} else if t.IsPrivateSend() {
		t.SetPrivateSend()
	}
}

func IsPrivacyTransaction(denom []int) bool {
	// todo add fourth denomination?
	return denom[0] > 2 || denom[1] > 2 || denom[2] > 2
}

// IsOneOrTwoOutputs checks if TX has only 1 or 2 outputs. Used for clustering.
func (t Transaction) IsOneOrTwoOutput() bool {
	return !t.IsMixing() &&
		(len(t.Outputs) == 2 || len(t.Outputs) == 1)
}

// IsMixing checks if TX is mixing
func (t Transaction) IsMixing() bool {
	if len(t.Inputs) < 3 || len(t.Outputs) < 3 || len(t.Inputs) != len(t.Outputs) {
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

type transactionQuery struct {
	Q []Transaction `json:"q"`
}

func (tq transactionQuery) payload() (tx Transaction, err error) {
	lenQ := len(tq.Q)

	if lenQ == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errors.New("no transactions found"))
		return
	} else if lenQ > 1 {
		// found more than one transaction, which should not be possible
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errors.New("found more than one transaction"))
		return
	}
	tx = tq.Q[0]
	return
}

type FrontendOutput struct {
	Amount      string `json:"amount"`
	InputIndex  int    `json:"inputindex"`
	OutputIndex int    `json:"outputindex"`
	IsCoinbase  bool   `json:"iscoinbase"`
	AddressHash string `json:"addresshash"`
}

type FrontendTransaction struct {
	Hash           string           `json:"txhash"`
	BlockHash      string           `json:"bhash"`
	PrivacyType    string           `json:"privacytype"`
	BlockId        uint64           `json:"bid"`
	BlockTimestamp string           `json:"bts"`
	Outputs        []FrontendOutput `json:"outputs"`
	Inputs         []FrontendOutput `json:"inputs"`
}

func (f FrontendTransaction) String() string {
	return fmt.Sprintf("Hash: %s, BlockHash: %s, BlockId: %d, "+
		"Privacy type: %s, BlockTimestamp: %s, Output Count: %d, Input Count: %d",
		f.Hash, f.BlockHash, f.BlockId, f.PrivacyType, f.BlockTimestamp,
		len(f.Outputs), len(f.Inputs))
}
