package output

import (
	"backend/cmd/cliutil"

	"errors"
	"fmt"
)

// DType is the dgraph database type for the Output type
const DType = "Output"

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

func (o Output) String() string {
	output := fmt.Sprintf("UID: %s, Amount: %d, KeyAsm: %s, SigAsm: %s",
		o.UID, *o.Amount, o.KeyAsm, o.SigAsm)

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
	o.DType = []string{DType}
}

type outputQuery struct {
	GetOutput []struct {
		Outputs []Output `json:"tx_outputs"`
	} `json:"getOutput"`
}

var (
	// ErrorNotFound is returned if an output was not found
	ErrorNotFound      = errors.New("output not found")
	errorMultipleFound = errors.New("found multiple outputs")
)

func (oq outputQuery) payload() (op Output, err error) {
	lenQ := len(oq.GetOutput)
	if lenQ == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorNotFound)
		return
	}

	lenTx := len(oq.GetOutput[0].Outputs)
	if lenTx == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorNotFound)
		return
	}

	if lenQ > 1 || lenTx > 1 {
		// found more than one output, which should not be possible
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorMultipleFound)
		return
	}
	op = oq.GetOutput[0].Outputs[0]
	return
}
