// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"fmt"
)

// outputDType is the dgraph database type for the Output type
const outputDType = "Output"

// Output is the database representation of an output
type Output struct {
	UID         string   `json:"uid,omitempty"`
	OutputIndex *int32   `json:"outputindex,omitempty"`
	InputIndex  *int32   `json:"inputindex,omitempty"`
	TxType      string   `json:"txtype,omitempty"`
	Amount      *int64   `json:"amount,omitempty"`
	IsCoinbase  *bool    `json:"iscoinbase,omitempty"`
	SigAsm      string   `json:"sigasm,omitempty"`
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
