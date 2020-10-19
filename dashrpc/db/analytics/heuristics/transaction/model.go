package transaction

const DType = "TransactionHeuristic"

type DummyOrigin struct {
	Uid string `json:"uid,omitempty"`
}

type Heuristic struct {
	Uid           string        `json:"uid,omitempty"`
	HeuristicType string        `json:"type,omitempty"`
	TxUid         string        `json:"h_transaction,omitempty"`
	Origins       []DummyOrigin `json:"results,omitempty"`
	DType         []string      `json:"dgraph.type,omitempty"`
	// only included for finding the tx uid in the upsert step
	TxHash string `json:"-"`
}

func (h *Heuristic) SetDType() {
	h.DType = []string{DType}
}

type InputTransaction struct {
	Uid       string `json:"uid,omitempty"`
	Timestamp string `json:"ts,omitempty"`
}
