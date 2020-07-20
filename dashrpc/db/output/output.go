package output

type Output struct {
	Uid        string   `json:"uid,omitempty"`
	Index      string   `json:"index,omitempty"`
	TxType     string   `json:"txtype,omitempty"`
	Amount     string   `json:"amount,omitempty"`
	IsCoinbase string   `json:"iscoinbase,omitempty"`
	DType      []string `json:"dgraph.type,omitempty"`
}
