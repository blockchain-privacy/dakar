package transaction

import "fmt"

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

type HeuristicTransaction struct {
	// Uid holds the uid of the transaction
	Uid string `json:"uid,omitempty"`
	// Timestamp holds the timestamp of the corresponding block
	Timestamp string `json:"ts,omitempty"`
	// Address holds the address associated with this transaction.
	// Depending on the context this could be the address of an input or an output
	Address string
	// Outputs holds a slice of amounts.
	// Depending on the context this could be the input or the output amounts.
	Outputs []struct {
		Amount int64 `json:"amount,omitempty"`
	} `json:"tx_outputs,omitempty"`
}

func (h HeuristicTransaction) String() string {
	return fmt.Sprintf("Uid: %s, timestamp: %s, associated address: %s, output count: %d", h.Uid, h.Timestamp, h.Address, len(h.Outputs))
}
