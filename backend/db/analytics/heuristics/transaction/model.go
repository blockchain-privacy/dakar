package transaction

import (
	"fmt"
	"time"
)

const DType = "TransactionHeuristic"

type DummyOrigin struct {
	Uid string `json:"uid,omitempty"`
}

type Heuristic struct {
	Uid           string `json:"uid,omitempty"`
	HeuristicType string `json:"type,omitempty"`
	Parameter     string `json:"parameter,omitempty"`
	UserUid       string `json:"~user_heuristics,omitempty"`
	Transaction   struct {
		Uid string `json:"uid,omitempty"`
	} `json:"h_transaction,omitempty"`
	Timestamp       string        `json:"ts,omitempty"`
	ParentHeuristic []Heuristic   `json:"parent_heuristic,omitempty"`
	ChildHeuristics []Heuristic   `json:"~parent_heuristic,omitempty"`
	Origins         []DummyOrigin `json:"results,omitempty"`

	DType []string `json:"dgraph.type,omitempty"`
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
	Timestamp time.Time `json:"ts,omitempty"`
	// Addresses holds the address associated with this transaction.
	// Depending on the context this could be the address of an input or an output
	Addresses []string
	// Outputs holds a slice of amounts.
	// Depending on the context this could be the input or the output amounts.
	Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
}

type HeuristicOutput struct {
	Amount           int64  `json:"amount,omitempty"`
	InputTransaction string `json:"input_tx,omitempty"`
}

func (h HeuristicTransaction) String() string {
	return fmt.Sprintf("Uid: %s, timestamp: %s, associated address: %s, output count: %d", h.Uid, h.Timestamp, h.Addresses, len(h.Outputs))
}

type HeuristicInput struct {
	AddressHash string `json:"addresshash,omitempty"`
}

type queryHeuristicTransaction struct {
	Uid     string            `json:"uid,omitempty"`
	Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
	Inputs  []HeuristicInput  `json:"tx_inputs,omitempty"`
	Block   []struct {
		Timestamp time.Time `json:"ts,omitempty"`
	} `json:"~transactions,omitempty"`
}

type FrontendHeuristicComplete struct {
	Uid        string              `json:"uid,omitempty"`
	Timestamp  string              `json:"ts,omitempty"`
	Heuristics []FrontendHeuristic `json:"~h_transaction,omitempty"`
}

func (f FrontendHeuristicComplete) String() string {
	return fmt.Sprintf("Uid:%s, timestamp:%s, heuristic count:%d", f.Uid, f.Timestamp, len(f.Heuristics))
}

type FrontendHeuristicResult struct {
	Uid         string `json:"uid,omitempty"`
	Timestamp   string `json:"ts,omitempty"`
	AddressHash string `json:"addresshash,omitempty"`
	TxHash      string `json:"txhash,omitempty"`
}

type FrontendHeuristic struct {
	Uid             string                    `json:"uid,omitempty"`
	Timestamp       string                    `json:"ts,omitempty"`
	Type            string                    `json:"type,omitempty"`
	Parameter       string                    `json:"parameter,omitempty"`
	ParentHeuristic []Heuristic               `json:"parent_heuristic,omitempty"`
	ChildHeuristics []Heuristic               `json:"children,omitempty"`
	ResultCount     int                       `json:"num_results,omitempty"`
	Results         []FrontendHeuristicResult `json:"results,omitempty"`
}

type ShortestTransactionPathRequest struct {
	// From is the starting point of the shortest path lookup
	From string `json:"from,omitempty"`
	// To is the end point of the shortest path lookup
	To string `json:"to,omitempty"`
	// IncludePrivacyTransactions determines if privacy transactions
	// should be considered when doing the shortest path lookup
	IncludePrivacyTransactions bool `json:"includePrivacyTransactions"`
	// AnyDirection determines the search direction of the shortest transaction path query
	// True: Both inputs and outputs are traversed
	// False: Only inputs are traversed
	AnyDirection bool `json:"anyDirection"`
}

type HeuristicListItem struct {
	Transaction      string `json:"txhash,omitempty"`
	LastModification string `json:"mod_time,omitempty"`
	HeuristicCount   uint64 `json:"h_count,omitempty"`
}

type DeleteHeuristicRequest struct {
	DeleteAll       bool   `json:"delete_all"`
	TransactionHash string `json:"tx_hash,omitempty"`
}
