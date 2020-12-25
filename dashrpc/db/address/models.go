package address

import (
	op "dashrpc/db/output"
	"errors"
	"fmt"
)

const DType = "Address"

const (
	// SortAscendingByInputTime sort addresses ascending by the input transaction timestamp
	SortAscendingByInputTime int = iota
	// SortDescendingByInputTime sort addresses descending by the input transaction timestamp
	SortDescendingByInputTime
	// SortAscendingByOutputTime sort addresses ascending by the output transaction timestamp
	SortAscendingByOutputTime
	// SortDescendingByOutputTime sort addresses descending by the output transaction timestamp
	SortDescendingByOutputTime
)

var (
	ErrorAddressNotFound = errors.New("no address found")
	ErrorInvalidResult   = errors.New("invalid result")
)

type Address struct {
	Uid     string      `json:"uid,omitempty"`
	Hash    string      `json:"addresshash,omitempty"`
	Outputs []op.Output `json:"addr_outputs,omitempty"`
	DType   []string    `json:"dgraph.type,omitempty"`
}

func (a Address) String() string {
	output := fmt.Sprintf("Uid: %s, Hash: %s", a.Uid, a.Hash)

	if a.Outputs != nil {
		output += fmt.Sprintf(", OutputCount: %d", len(a.Outputs))
	}

	return output
}

func (a *Address) SetDType() {
	a.DType = []string{DType}
}

// checks if the given address has all attributes filled
func (a Address) isComplete() bool {
	return a.Uid != "" && a.Hash != "" && a.DType != nil && a.Outputs != nil
}

type FrontendOutput struct {
	Amount                uint64 `json:"amount"`
	IsCoinbase            bool   `json:"iscoinbase"`
	InputTransactionHash  string `json:"input_transaction"`
	InputTimestamp        string `json:"input_ts"`
	OutputTransactionHash string `json:"output_transaction"`
	OutputTimestamp       string `json:"output_ts"`
}

func (o FrontendOutput) String() string {
	return fmt.Sprintf("Amount: %d", o.Amount)
}

type FrontendAddress struct {
	Hash       string           `json:"addresshash"`
	NumOutputs int64            `json:"num_outputs"`
	Outputs    []FrontendOutput `json:"addr_outputs"`
}

func (f FrontendAddress) String() string {
	return fmt.Sprintf("Hash: %s, OutputCount: %d", f.Hash, len(f.Outputs))
}
