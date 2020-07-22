package output

import (
	"errors"
	"fmt"
	"strconv"
)

type Output struct {
	Uid        string   `json:"uid,omitempty"`
	Index      *uint64  `json:"index,omitempty"`
	TxType     string   `json:"txtype,omitempty"`
	Amount     *float64 `json:"amount,omitempty"`
	IsCoinbase *bool    `json:"iscoinbase,omitempty"`
	DType      []string `json:"dgraph.type,omitempty"`
}

func (o Output) String() string {
	output := fmt.Sprintf("Uid: %s", o.Uid)

	if o.Index != nil {
		output += fmt.Sprintf(", Index: %d", *o.Index)
	}

	if o.Amount != nil {
		output += fmt.Sprintf(", Amount: %f", *o.Amount)
	}

	if o.IsCoinbase != nil {
		output += fmt.Sprintf(", IsCoinbase: %t", *o.IsCoinbase)
	}

	return output
}

func (o Output) ToUpdate() (op UpdateOutputData) {
	op.DType = o.DType
	op.Uid = o.Uid
	op.Index = o.Index
	op.IsCoinbase = o.IsCoinbase
	op.TxType = o.TxType

	if o.Amount != nil {
		op.Amount = strconv.FormatFloat(*o.Amount, 'f', 8, 64)
	}

	return op
}

type UpdateOutputData struct {
	Uid        string   `json:"uid,omitempty"`
	Index      *uint64  `json:"index,omitempty"`
	TxType     string   `json:"txtype,omitempty"`
	Amount     string   `json:"amount,omitempty"`
	IsCoinbase *bool    `json:"iscoinbase,omitempty"`
	DType      []string `json:"dgraph.type,omitempty"`
}

func (o UpdateOutputData) String() string {
	output := fmt.Sprintf("Uid: %s, Amount: %s", o.Uid, o.Amount)

	if o.Index != nil {
		output += fmt.Sprintf(", Index: %d", *o.Index)
	}

	if o.IsCoinbase != nil {
		output += fmt.Sprintf(", IsCoinbase: %t", *o.IsCoinbase)
	}

	return output
}

func (o UpdateOutputData) ToOutput() (op Output, err error) {
	op.DType = o.DType
	op.Uid = o.Uid
	op.Index = o.Index
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
		Transaction []struct {
			Outputs []Output `json:"tx_outputs"`
		} `json:"transaction"`
	} `json:"getOutput"`
}

func (oq outputQuery) payload() (op Output, err error) {
	lenQ := len(oq.GetOutput)
	if lenQ == 0 {
		return op, errors.New("no output found")
	}

	lenTx := len(oq.GetOutput[0].Transaction)
	if lenTx == 0 {
		return op, errors.New("no output found")
	}

	lenO := len(oq.GetOutput[0].Transaction[0].Outputs)
	if lenO == 0 {
		return op, errors.New("no output found")
	}

	if lenQ > 1 || lenTx > 1 || lenO > 1 {
		// found more than one output, which should not be possible
		return op, errors.New("found more than one output")
	}

	op = oq.GetOutput[0].Transaction[0].Outputs[0]
	return op, err
}
