package address

import (
	op "backend/db/output"
	"errors"

	"fmt"
)

const DType = "Address"

const (
	// SortAscendingByOutputTime sort outputs ascending by the output transaction timestamp
	SortAscendingByOutputTime int = iota
	// SortDescendingByOutputTime sort outputs descending by the output transaction timestamp
	SortDescendingByOutputTime
	// SortAscendingByInputTime sort outputs ascending by the input transaction timestamp
	SortAscendingByInputTime
	// SortDescendingByInputTime sort outputs descending by the input transaction timestamp
	SortDescendingByInputTime
	// SortAscendingByAmount sort outputs ascending by the output amount
	SortAscendingByAmount
	// SortDescendingByAmount sort outputs ascending by the output amount
	SortDescendingByAmount
)

const (
	// FilterByCoinbase filters outputs if they are a coinbase output
	FilterByCoinbase int = iota
	// FilterByUnspent filters outputs if they are unspent
	FilterByUnspent
)

var (
	ErrorAddressNotFound = errors.New("no address found")
	ErrorInvalidResult   = errors.New("invalid result")
)

// IsValidSortOrder returns true if sortOrder has a valid sort order value
func IsValidSortOrder(sortOrder int) bool {
	return sortOrder == SortAscendingByInputTime || sortOrder == SortDescendingByInputTime ||
		sortOrder == SortAscendingByOutputTime || sortOrder == SortDescendingByOutputTime ||
		sortOrder == SortAscendingByAmount || sortOrder == SortDescendingByAmount
}

// IsValidFilter returns true if filters had a valid value
func IsValidFilter(filters []int) bool {
	for _, f := range filters {
		if f != FilterByUnspent && f != FilterByCoinbase {
			return false
		}
	}

	return true
}

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
	IsCoinbase            bool   `json:"is_coinbase"`
	InputIndex            int    `json:"input_index"`
	InputTransactionHash  string `json:"input_transaction"`
	InputTimestamp        string `json:"input_ts"`
	OutputIndex           int    `json:"output_index"`
	OutputTransactionHash string `json:"output_transaction"`
	OutputTimestamp       string `json:"output_ts"`
}

func (o FrontendOutput) String() string {
	return fmt.Sprintf("Amount: %d", o.Amount)
}

type FrontendAddress struct {
	Hash          string           `json:"addresshash"`
	QueryMaxCount int64            `json:"query_max_count"`
	CoinbaseCount int64            `json:"coinbase_count"`
	OutputCount   int64            `json:"output_count"`
	InputCount    int64            `json:"input_count"`
	InputSum      int64            `json:"input_sum"`
	OutputSum     int64            `json:"output_sum"`
	Outputs       []FrontendOutput `json:"addr_outputs"`
}

func (f FrontendAddress) String() string {
	return fmt.Sprintf("Hash: %s, OutputCount: %d", f.Hash, len(f.Outputs))
}
