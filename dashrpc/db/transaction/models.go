package transaction

import (
	op "dashrpc/db/output"
	"errors"
	"fmt"
)

const (
	DType              = "Transaction"
	PrivacyOrigin      = "origin"
	PrivacyMixing      = "mixing"
	PrivacyDestination = "destination"
)

var (
	ErrorTransactionNotFound = errors.New("no transaction found")
	ErrorInvalidResult       = errors.New("invalid result")
)

type Transaction struct {
	Uid         string        `json:"uid,omitempty"`
	PrivacyType string        `json:"privacytype,omitempty"`
	Fee         *int64        `json:"fee,omitempty"`
	Outputs     []op.Output   `json:"tx_outputs,omitempty"`
	Inputs      []op.Output   `json:"tx_inputs,omitempty"`
	Hash        string        `json:"txhash,omitempty"`
	Origins     []Transaction `json:"origins,omitempty"`
	DType       []string      `json:"dgraph.type,omitempty"`
}

func (t Transaction) String() string {
	output := fmt.Sprintf("Uid: %s, Hash: %s, Privacy type: %s, Fee: %d",
		t.Uid, t.Hash, t.PrivacyType, *t.Fee)

	if t.Outputs != nil {
		output += fmt.Sprintf(", Output count: %d", len(t.Outputs))
	}

	if t.Inputs != nil {
		output += fmt.Sprintf(", Input count: %d", len(t.Inputs))
	}

	if t.Origins != nil {
		output += fmt.Sprintf(", Origin count: %d", len(t.Origins))
	}
	return output
}

func (t *Transaction) SetDType() {
	t.DType = []string{DType}
}

func (t *Transaction) SetPrivacyOrigin() {
	t.PrivacyType = PrivacyOrigin
}

func (t *Transaction) SetMixing() {
	t.PrivacyType = PrivacyMixing
}

func (t *Transaction) SetPrivacyDestination() {
	t.PrivacyType = PrivacyDestination
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

// IsPrivacyOrigin checks if the TX creates denominations
func (t Transaction) IsPrivacyOrigin(areAllInputAddressesEqual bool) bool {
	return len(t.Inputs) >= 1 && areAllInputAddressesEqual && len(t.Outputs) > 2 && IsPrivacyTransaction(t.CountOutputDenominations())
}

// IsPrivacyDestination checks if the TX is the end receiver of a private send transaction
func (t Transaction) IsPrivacyDestination() bool {
	return len(t.Outputs) == 1 && len(t.Inputs) > 2 && IsPrivacyTransaction(t.CountInputDenominations())
}

// checks if the cumulative amount of inputs and outputs matches
func (t *Transaction) CalculateTransactionFee() (err error) {
	var amountInputs int64
	var amountOutputs int64

	for _, e := range t.Inputs {
		if e.Amount == nil {
			return errors.New("error amount is not set")
		}
		amountInputs += *e.Amount
	}

	for _, e := range t.Outputs {
		if e.Amount == nil {
			return errors.New("error amount is not set")
		}
		amountOutputs += *e.Amount
	}

	fee := amountInputs - amountOutputs
	t.Fee = &fee

	return
}

func IsPrivacyTransaction(denom []int) bool {
	return denom[0] > 2 || denom[1] > 2 || denom[2] > 2 || denom[3] > 2 || denom[4] > 2
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
		err = errors.New("no transactions found")
		return
	} else if lenQ > 1 {
		// found more than one transaction, which should not be possible
		err = errors.New("found more than one transaction")
		return
	}
	tx = tq.Q[0]
	return
}

type FrontendOutput struct {
	Amount      *int64  `json:"amount"`
	InputIndex  *uint32 `json:"inputindex"`
	OutputIndex *uint32 `json:"outputindex"`
	IsCoinbase  bool    `json:"iscoinbase"`
	AddressHash string  `json:"addresshash"`
}

type FrontendTransaction struct {
	Hash           string           `json:"txhash"`
	BlockHash      string           `json:"bhash"`
	Fee            int64            `json:"fee"`
	PrivacyType    string           `json:"privacytype"`
	BlockId        uint64           `json:"bid"`
	BlockTimestamp string           `json:"bts"`
	Outputs        []FrontendOutput `json:"outputs"`
	Inputs         []FrontendOutput `json:"inputs"`
	OriginCount    uint64           `json:"origincount"`
}

func (f FrontendTransaction) String() string {
	return fmt.Sprintf("Hash: %s, BlockHash: %s, BlockId: %d, "+
		"Fee: %d, Privacy type: %s, BlockTimestamp: %s, Output Count: %d, Input Count: %d, Origin Count: %d",
		f.Hash, f.BlockHash, f.BlockId, f.Fee, f.PrivacyType, f.BlockTimestamp,
		len(f.Outputs), len(f.Inputs), f.OriginCount)
}

type FrontendRequest struct {
	Hash        string           `json:"txhash,omitempty"`
	PrivacyType string           `json:"privacytype,omitempty"`
	Fee         string           `json:"fee,omitempty"`
	OriginCount uint64           `json:"origincount,omitempty"`
	Outputs     []FrontendOutput `json:"outputs,omitempty"`
	Inputs      []FrontendOutput `json:"inputs,omitempty"`
	Block       []struct {
		Hash string `json:"blockhash,omitempty"`
		Ts   string `json:"ts,omitempty"`
		Id   uint64 `json:"id,omitempty"`
	} `json:"block,omitempty"`
}

type SourceTransaction struct {
	Hash      string `json:"txhash,omitempty"`
	Timestamp string `json:"ts,omitempty"`
}
