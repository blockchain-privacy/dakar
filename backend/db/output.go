package db

import (
	"fmt"
	"log"
)

// outputDType is the dgraph database type for the Output type
const outputDType = "Output"

// Output is the database representation of an output
type Output struct {
	UID         string   `json:"uid,omitempty"`
	OutputIndex *uint32  `json:"outputindex,omitempty"`
	InputIndex  *uint32  `json:"inputindex,omitempty"`
	TxType      string   `json:"txtype,omitempty"`
	Amount      *int64   `json:"amount,omitempty"`
	IsCoinbase  *bool    `json:"iscoinbase,omitempty"`
	SigHex      string   `json:"sighex,omitempty"`
	SigAsm      string   `json:"sigasm,omitempty"`
	KeyHex      string   `json:"keyhex,omitempty"`
	KeyAsm      string   `json:"keyasm,omitempty"`
	DType       []string `json:"dgraph.type,omitempty"`
}

func (o *Output) String() string {
	output := fmt.Sprintf("UID: %s, KeyAsm: %s, SigAsm: %s", o.UID, o.KeyAsm, o.SigAsm)

	if o.Amount != nil {
		output += fmt.Sprintf(", Amount: %d", *o.Amount)
	}

	if o.OutputIndex != nil {
		output += fmt.Sprintf(", OutputIndex: %d", *o.OutputIndex)
	}

	if o.InputIndex != nil {
		output += fmt.Sprintf(", InputIndex: %d", *o.InputIndex)
	}

	if o.IsCoinbase != nil {
		output += fmt.Sprintf(", IsCoinbase: %t", *o.IsCoinbase)
	}

	return output
}

// SetDType sets the DType for dgraph type recognition
func (o *Output) SetDType() {
	o.DType = []string{outputDType}
}

// NumDenominations is the number of Dash PrivateSend denominations existing
const NumDenominations = 5

const (
	// MinCollateral is 1/10 of the smallest denomination: round(100001/10).
	MinCollateral = 10000

	// OldMinCollateral is the minimum collateral before the 5th denomination
	// was added in protocol version 70213 it was round(1000010/10): 100000
	// OldMinCollateral = 100000

	// MaxCollateral is the maximum allowed collateral
	MaxCollateral = 40000 // 4*MinCollateral
	// OldMaxCollateral is to old collateral
	OldMaxCollateral = 400000 // 4*OldMinCollateral
)

var denominationsTypes = [NumDenominations]int64{1000010000, 100001000, 10000100, 1000010, 100001}

// CountOutputDenominations returns for each denomination how often it occurred in the given outputs
func CountOutputDenominations(outputs []Output) [NumDenominations]int {
	amounts := make([]int64, len(outputs))

	for i, o := range outputs {
		if o.Amount == nil {
			log.Println("error amount not set")
			return [NumDenominations]int{}
		}
		amounts[i] = *o.Amount
	}

	return CountAmountDenominations(amounts)
}

// CountAmountDenominations returns the number of occurrences of each denomination in the given amounts
func CountAmountDenominations(amounts []int64) (denominations [NumDenominations]int) {
	for _, amt := range amounts {
	inner:
		for i, v := range denominationsTypes {
			if amt == v {
				denominations[i]++
				break inner
			}
		}
	}

	return
}
