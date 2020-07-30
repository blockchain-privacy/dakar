package output

import (
	"errors"
	"fmt"
	"strconv"
)

const DType = "TxOutput"

type Output struct {
	Uid         string   `json:"uid,omitempty"`
	OutputIndex *uint64  `json:"outputindex,omitempty"`
	InputIndex  *uint64  `json:"inputindex,omitempty"`
	TxType      string   `json:"txtype,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	IsCoinbase  *bool    `json:"iscoinbase,omitempty"`
	DType       []string `json:"dgraph.type,omitempty"`
}

func (o Output) String() string {
	output := fmt.Sprintf("Uid: %s", o.Uid)

	if o.OutputIndex != nil {
		output += fmt.Sprintf(", OutputIndex: %d", *o.OutputIndex)
	}

	if o.InputIndex != nil {
		output += fmt.Sprintf(", InputIndex: %d", *o.InputIndex)
	}

	if o.Amount != nil {
		output += fmt.Sprintf(", Amount: %f", *o.Amount)
	}

	if o.IsCoinbase != nil {
		output += fmt.Sprintf(", IsCoinbase: %t", *o.IsCoinbase)
	}

	return output
}
func (o *Output) SetDType() {
	o.DType = []string{DType}
}

func (o Output) ToUpdate() (op UpdateOutputData) {
	op.DType = o.DType
	op.Uid = o.Uid
	op.OutputIndex = o.OutputIndex
	op.InputIndex = o.InputIndex
	op.IsCoinbase = o.IsCoinbase
	op.TxType = o.TxType

	if o.Amount != nil {
		op.Amount = strconv.FormatFloat(*o.Amount, 'f', 8, 64)
	}

	return op
}

type UpdateOutputData struct {
	Uid         string   `json:"uid,omitempty"`
	OutputIndex *uint64  `json:"outputindex,omitempty"`
	InputIndex  *uint64  `json:"inputindex,omitempty"`
	TxType      string   `json:"txtype,omitempty"`
	Amount      string   `json:"amount,omitempty"`
	IsCoinbase  *bool    `json:"iscoinbase,omitempty"`
	DType       []string `json:"dgraph.type,omitempty"`
}

func (o UpdateOutputData) String() string {
	output := fmt.Sprintf("Uid: %s, Amount: %s", o.Uid, o.Amount)

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

func (o UpdateOutputData) ToOutput() (op Output, err error) {
	op.DType = o.DType
	op.Uid = o.Uid
	op.OutputIndex = o.OutputIndex
	op.InputIndex = o.InputIndex
	op.IsCoinbase = o.IsCoinbase
	op.TxType = o.TxType

	if o.Amount != "" {
		amount, err := strconv.ParseFloat(o.Amount, 64)
		if err != nil {
			return op, err
		}

		op.Amount = &amount
	}

	return op, err
}

type outputQuery struct {
	GetOutput []struct {
		Outputs []UpdateOutputData `json:"tx_outputs"`
	} `json:"getOutput"`
}

const (
	ErrorNotFound      = "output not found"
	ErrorMultipleFound = "found multiple outputs"
)

func (oq outputQuery) payload() (op Output, err error) {
	lenQ := len(oq.GetOutput)
	if lenQ == 0 {
		return op, errors.New(ErrorNotFound)
	}

	lenTx := len(oq.GetOutput[0].Outputs)
	if lenTx == 0 {
		return op, errors.New(ErrorNotFound)
	}

	if lenQ > 1 || lenTx > 1 {
		// found more than one output, which should not be possible
		return op, errors.New(ErrorMultipleFound)
	}

	return oq.GetOutput[0].Outputs[0].ToOutput()
}

type VerboseOutput struct {
	Uid               string
	OutputIndex       *uint64
	InputIndex        *uint64
	TxType            string
	Amount            string
	IsCoinbase        *bool
	OutputTransaction string
	InputTransaction  string
	Addresses         []string
}
