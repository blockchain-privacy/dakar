package output

import "strconv"

type Output struct {
	Uid        string   `json:"uid,omitempty"`
	Index      *uint64  `json:"index,omitempty"`
	TxType     string   `json:"txtype,omitempty"`
	Amount     *float64 `json:"amount,omitempty"`
	IsCoinbase *bool    `json:"iscoinbase,omitempty"`
	DType      []string `json:"dgraph.type,omitempty"`
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
