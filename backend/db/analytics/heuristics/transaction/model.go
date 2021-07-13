package transaction

import (
	"fmt"
	"time"
)

// DType is the dgraph database type for the TransactionHeuristic type
const DType = "TransactionHeuristic"
const ResultDType = "TransactionHeuristicResult"

// DummyNode holds the uid of a database node
type DummyNode struct {
	UID string `json:"uid,omitempty"`
}

// HeuristicResult holds one result (origin) of a heuristic and
// optionally the results of a forward lookup (destinations)
type HeuristicResult struct {
	Origin       DummyNode   `json:"origin,omitempty"`
	Destinations []DummyNode `json:"destinations,omitempty"`
	DType        []string    `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (r *HeuristicResult) SetDType() {
	r.DType = []string{ResultDType}
}

// Heuristic is the database type representation of a heuristic
type Heuristic struct {
	UID           string `json:"uid,omitempty"`
	HeuristicType string `json:"type,omitempty"`
	Parameter     string `json:"parameter,omitempty"`
	UserUID       string `json:"~user_heuristics,omitempty"`
	Transaction   struct {
		UID string `json:"uid,omitempty"`
	} `json:"h_transaction,omitempty"`
	Timestamp       string            `json:"ts,omitempty"`
	ParentHeuristic []Heuristic       `json:"parent_heuristic,omitempty"`
	ChildHeuristics []Heuristic       `json:"~parent_heuristic,omitempty"`
	Results         []HeuristicResult `json:"results,omitempty"`

	DType []string `json:"dgraph.type,omitempty"`
	// only included for finding the tx uid in the upsert step
	TxHash string `json:"-"`
}

// SetDType sets the DType for dgraph type recognition
func (h *Heuristic) SetDType() {
	h.DType = []string{DType}
}

// HeuristicTransaction holds all data a heuristic needs for a specific transaction
type HeuristicTransaction struct {
	// UID holds the uid of the transaction
	UID string `json:"uid,omitempty"`
	// Timestamp holds the timestamp of the corresponding block
	Timestamp time.Time `json:"ts,omitempty"`
	// Addresses holds the address associated with this transaction.
	// Depending on the context this could be the address of an input or an output
	Addresses []string
	// Outputs holds a slice of amounts.
	// Depending on the context this could be the input or the output amounts.
	Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
}

// HeuristicOutput holds all data a heuristic needs for a specific output
type HeuristicOutput struct {
	Amount           int64  `json:"amount,omitempty"`
	InputTransaction string `json:"input_tx,omitempty"`
}

func (h HeuristicTransaction) String() string {
	return fmt.Sprintf("UID: %s, timestamp: %s, associated address: %s, output count: %d", h.UID, h.Timestamp, h.Addresses, len(h.Outputs))
}

// HeuristicInput only holds the uid of the input address
type HeuristicInput struct {
	AddressUID string `json:"uid,omitempty"`
}

type queryHeuristicTransaction struct {
	UID     string            `json:"uid,omitempty"`
	Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
	Inputs  []HeuristicInput  `json:"tx_inputs,omitempty"`
	Block   []struct {
		Timestamp time.Time `json:"ts,omitempty"`
	} `json:"~transactions,omitempty"`
}

// FrontendHeuristicComplete holds all heuristic tree data which is exposed to the frontend
type FrontendHeuristicComplete struct {
	UID        string              `json:"uid,omitempty"`
	Timestamp  string              `json:"ts,omitempty"`
	Heuristics []FrontendHeuristic `json:"~h_transaction,omitempty"`
}

// String returns the string representation of a FrontendHeuristicComplete object
func (f FrontendHeuristicComplete) String() string {
	return fmt.Sprintf("UID:%s, timestamp:%s, heuristic count:%d", f.UID, f.Timestamp, len(f.Heuristics))
}

// FrontendHeuristicResult holds heuristic result data which is exposed to the frontend
type FrontendHeuristicResult struct {
	UID         string `json:"uid,omitempty"`
	Timestamp   string `json:"ts,omitempty"`
	AddressHash string `json:"addresshash,omitempty"`
	TxHash      string `json:"txhash,omitempty"`
}

// FrontendHeuristic holds all heuristic data which is exposed to the frontend
type FrontendHeuristic struct {
	UID             string                    `json:"uid,omitempty"`
	Timestamp       string                    `json:"ts,omitempty"`
	Type            string                    `json:"type,omitempty"`
	Parameter       string                    `json:"parameter,omitempty"`
	ParentHeuristic []Heuristic               `json:"parent_heuristic,omitempty"`
	ChildHeuristics []Heuristic               `json:"children,omitempty"`
	ResultCount     int                       `json:"num_results,omitempty"`
	Results         []FrontendHeuristicResult `json:"results,omitempty"`
}

// ShortestTransactionPathRequest holds all configuration data for a shortest transaction search request
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

// HeuristicListItem holds data for an item in the heuristic list of a user
type HeuristicListItem struct {
	Transaction      string `json:"txhash,omitempty"`
	LastModification string `json:"mod_time,omitempty"`
	HeuristicCount   uint64 `json:"h_count,omitempty"`
}

// DeleteHeuristicRequest holds configuration data about whether all heuristics should be deleted or only a specific one
type DeleteHeuristicRequest struct {
	DeleteAll       bool   `json:"delete_all"`
	TransactionHash string `json:"tx_hash,omitempty"`
}
