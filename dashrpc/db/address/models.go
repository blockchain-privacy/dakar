package address

import (
	op "dashrpc/db/output"
	"fmt"
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

// checks if the given address has all attributes filled
func (a Address) isComplete() bool {
	return a.Uid != "" && a.Hash != "" && a.DType != nil && a.Outputs != nil
}

type addressQuery struct {
	Q []Address `json:"q"`
}
