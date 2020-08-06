package address

import (
	op "dashrpc/db/output"
	"errors"
	"fmt"
)

const DType = "Address"

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

type addressQuery struct {
	Q []Address `json:"q"`
}

func (aq addressQuery) payload() (a Address, err error) {
	lenQ := len(aq.Q)

	if lenQ == 0 {
		err = errors.New("no addresses found")
		return a, err
	} else if lenQ > 1 {
		// found more than one transaction, which should not be possible
		err = errors.New("found more than one address")
		return a, err
	}
	a = aq.Q[0]
	return a, err
}

type FrontendAddress struct {
	Hash    string `json:"addresshash"`
	Outputs []struct {
		Amount                string `json:"amount"`
		IsCoinbase            bool   `json:"iscoinbase"`
		InputTransactionHash  string `json:"input_transaction"`
		OutputTransactionHash string `json:"output_transaction"`
	} `json:"addr_outputs"`
}

func (f FrontendAddress) String() string {
	return fmt.Sprintf("Hash: %s, OutputCount: %d", f.Hash, len(f.Outputs))
}
