package transaction

import (
	"backend/constants"
	op "backend/db/output"
	"errors"
	"fmt"
)

// DType is the dgraph database type for the Transaction type
const DType = "Transaction"

var (
	ErrorTransactionNotFound = errors.New("no transaction found")
	ErrorInvalidResult       = errors.New("invalid result")
)

type Transaction struct {
	UID         string                 `json:"uid,omitempty"`
	PrivacyType *constants.PrivacyType `json:"privacytype,omitempty"`
	Fee         *int64                 `json:"fee,omitempty"`
	Outputs     []op.Output            `json:"tx_outputs,omitempty"`
	Inputs      []op.Output            `json:"tx_inputs,omitempty"`
	Hash        string                 `json:"txhash,omitempty"`
	Origins     []Transaction          `json:"origins,omitempty"`
	DType       []string               `json:"dgraph.type,omitempty"`
}

func (t Transaction) String() string {
	output := fmt.Sprintf("UID: %s, Hash: %s, Privacy type: %d, Fee: %d",
		t.UID, t.Hash, t.PrivacyType, *t.Fee)

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

// SetDType sets the DType for dgraph type recognition
func (t *Transaction) SetDType() {
	t.DType = []string{DType}
}

// CalculateTransactionFee checks if the cumulative amount of inputs and outputs matches
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

func (t *Transaction) IsMixingTransaction() bool {
	return t.PrivacyType != nil && t.PrivacyType.IsMixing()
}

func (t *Transaction) IsDestinationTransaction() bool {
	return t.PrivacyType != nil && t.PrivacyType.IsDestination()
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
	SigAsm      string  `json:"sigasm,omitempty"`
	KeyAsm      string  `json:"keyasm,omitempty"`
}

type FrontendTransaction struct {
	Hash           string           `json:"txhash,omitempty"`
	BlockHash      string           `json:"bhash,omitempty"`
	Fee            int64            `json:"fee"`
	PrivacyType    int64            `json:"privacytype,omitempty"`
	BlockID        uint64           `json:"bid"`
	BlockTimestamp string           `json:"bts,omitempty"`
	Outputs        []FrontendOutput `json:"outputs,omitempty"`
	Inputs         []FrontendOutput `json:"inputs,omitempty"`
	OriginCount    uint64           `json:"origincount"`
}

func (f FrontendTransaction) String() string {
	return fmt.Sprintf("Hash: %s, BlockHash: %s, BlockID: %d, "+
		"Fee: %d, Privacy type: %d, BlockTimestamp: %s, Output Count: %d, Input Count: %d, Origin Count: %d",
		f.Hash, f.BlockHash, f.BlockID, f.Fee, f.PrivacyType, f.BlockTimestamp,
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
		ID   uint64 `json:"id,omitempty"`
	} `json:"block,omitempty"`
}
