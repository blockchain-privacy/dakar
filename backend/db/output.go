// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

// outputDType is the Dgraph database type for the Output type
const outputDType = "Output"

// Output is the database representation of an output
type Output struct {
	UID         string   `json:"uid,omitempty"`
	OutputIndex *int32   `json:"outputindex,omitempty"`
	InputIndex  *int32   `json:"inputindex,omitempty"`
	Amount      *int64   `json:"amount,omitempty"`
	IsCoinbase  *bool    `json:"iscoinbase,omitempty"`
	DType       []string `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for Dgraph type recognition
func (o *Output) SetDType() {
	o.DType = []string{outputDType}
}
