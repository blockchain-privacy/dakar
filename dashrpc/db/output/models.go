package output

type Output struct {
	Uid        string   `json:"uid,omitempty"`
	Index      *uint64  `json:"index,omitempty"`
	TxType     string   `json:"txtype,omitempty"`
	Amount     *float64 `json:"amount,omitempty"`
	IsCoinbase *bool    `json:"iscoinbase,omitempty"`
	DType      []string `json:"dgraph.type,omitempty"`
}

type outputQuery struct {
	GetOutput []struct {
		Transaction []struct {
			Outputs []Output `json:"tx_outputs"`
		} `json:"transaction"`
	} `json:"getOutput"`
}
