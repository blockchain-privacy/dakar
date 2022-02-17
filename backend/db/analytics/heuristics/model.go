package heuristics

import (
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"fmt"
	"time"
)

// DType is the dgraph database type for the Heuristic type
const DType = "Heuristic"

// ResultDType is the dgraph database type for the HeuristicResult type
const ResultDType = "HeuristicResult"

// ClusterDType is the dgraph database type for the HeuristicCluster type
const ClusterDType = "HeuristicCluster"

// DummyNode holds the uid of a database node
type DummyNode struct {
	UID string `json:"uid,omitempty"`
}

// HeuristicResult holds one result (origin) of a heuristic and
// optionally the results of a forward lookup (destinations)
type HeuristicResult struct {
	Origin       DummyNode   `json:"HeuristicResult.origin,omitempty"`
	Destinations []DummyNode `json:"HeuristicResult.destinations,omitempty"`
	DType        []string    `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (r *HeuristicResult) SetDType() {
	r.DType = []string{ResultDType}
}

// HeuristicCluster holds a set of results (origins) of a heuristic
// which belong to the same cluster (or merged cluster) and its attributions
type HeuristicCluster struct {
	Results      []HeuristicResult         `json:"HeuristicCluster.results,omitempty"`
	Attributions []attribution.Attribution `json:"HeuristicCluster.attributions,omitempty"`
	DType        []string                  `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (c *HeuristicCluster) SetDType() {
	c.DType = []string{ClusterDType}
}

// Heuristic is the database type representation of a heuristic
type Heuristic struct {
	UID           string `json:"uid,omitempty"`
	HeuristicType string `json:"Heuristic.type,omitempty"`
	Parameter     string `json:"Heuristic.parameter,omitempty"`
	UserUID       string `json:"~User.heuristics,omitempty"`
	Transaction   struct {
		UID string `json:"uid,omitempty"`
	} `json:"Heuristic.transaction,omitempty"`
	Timestamp       string             `json:"Heuristic.ts,omitempty"`
	ParentHeuristic []Heuristic        `json:"Heuristic.parent,omitempty"`
	ChildHeuristics []Heuristic        `json:"~Heuristic.parent,omitempty"`
	Clusters        []HeuristicCluster `json:"Heuristic.clusters,omitempty"`

	DType []string `json:"dgraph.type,omitempty"`
	// only included for finding the tx uid in the upsert step
	TxHash string `json:"-"`
}

// SetDType sets the DType for dgraph type recognition
func (h *Heuristic) SetDType() {
	h.DType = []string{DType}
}

type ClusterUID string

// HeuristicTransaction holds all data a heuristic needs for a specific transaction
type HeuristicTransaction struct {
	// UID holds the uid of the transaction
	UID string `json:"uid,omitempty"`
	// Timestamp holds the timestamp of the corresponding block
	Timestamp time.Time `json:"ts,omitempty"`
	// Cluster holds the cluster associated with this transaction.
	// Depending on the context this could be the cluster of an input or an output
	Cluster ClusterUID
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
	return fmt.Sprintf("UID: %s, timestamp: %s, associated cluster: %s, output count: %d", h.UID, h.Timestamp, h.Cluster, len(h.Outputs))
}

// HeuristicInput only holds the uid of the input address
type HeuristicInput struct {
	Address string `json:"addr_uid,omitempty"`
	Cluster string `json:"cluster_uid,omitempty"`
}

type queryHeuristicTransaction struct {
	UID     string            `json:"uid,omitempty"`
	Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
	Inputs  []HeuristicInput  `json:"tx_inputs,omitempty"`
	Block   []struct {
		Timestamp time.Time `json:"ts,omitempty"`
	} `json:"~transactions,omitempty"`
}

type queryHeuristicClusters struct {
	UID     ClusterUID `json:"uid,omitempty"`
	Results []struct {
		Origin struct {
			UID     string            `json:"uid,omitempty"`
			Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
		} `json:"HeuristicResult.origin,omitempty"`
	} `json:"HeuristicCluster.results,omitempty"`
	Attributions []struct {
		UID string `json:"uid,omitempty"`
	} `json:"HeuristicCluster.attributions,omitempty"`
}

type queryHeuristicTransactionInputs struct {
	UID     string            `json:"uid,omitempty"`
	Outputs []HeuristicOutput `json:"tx_inputs,omitempty"`
}

// FrontendHeuristicComplete holds all heuristic tree data which is exposed to the frontend
type FrontendHeuristicComplete struct {
	UID        string              `json:"uid,omitempty"`
	Timestamp  string              `json:"ts,omitempty"`
	Heuristics []FrontendHeuristic `json:"~Heuristic.transaction,omitempty"`
}

// String returns the string representation of a FrontendHeuristicComplete object
func (f FrontendHeuristicComplete) String() string {
	return fmt.Sprintf("UID:%s, timestamp:%s, heuristic count:%d", f.UID, f.Timestamp, len(f.Heuristics))
}

// FrontendHeuristicResult holds heuristic result data which is exposed to the frontend
type FrontendHeuristicResult struct {
	Origin struct {
		UID         string `json:"uid,omitempty"`
		Timestamp   string `json:"ts,omitempty"`
		AddressHash string `json:"addresshash,omitempty"`
		TxHash      string `json:"txhash,omitempty"`
	} `json:"origin,omitempty"`
	Destinations []struct {
		UID       string `json:"uid,omitempty"`
		Timestamp string `json:"ts,omitempty"`
		TxHash    string `json:"txhash,omitempty"`
	} `json:"destinations,omitempty"`
}

// FrontendHeuristicRequest holds all heuristic data which is set by the user
type FrontendHeuristicRequest struct {
	UID             string                   `json:"uid,omitempty"`
	Type            string                   `json:"type,omitempty"`
	Parameter       string                   `json:"parameter,omitempty"`
	ParentHeuristic []Heuristic              `json:"parent,omitempty"`
	ChildHeuristics []Heuristic              `json:"children,omitempty"`
	ClusterTypes    []clustering.ClusterType `json:"clusterTypes,omitempty"`
}

// FrontendHeuristic holds all heuristic data which is exposed to the frontend
type FrontendHeuristic struct {
	UID             string                    `json:"uid,omitempty"`
	Timestamp       string                    `json:"ts,omitempty"`
	Type            string                    `json:"type,omitempty"`
	Parameter       string                    `json:"parameter,omitempty"`
	ParentHeuristic []Heuristic               `json:"parent,omitempty"`
	ChildHeuristics []Heuristic               `json:"children,omitempty"`
	ClusterCount    int                       `json:"clusterCount,omitempty"`
	Results         []FrontendHeuristicResult `json:"results,omitempty"`
}

// FrontendHeuristicResponse holds all heuristic data of a heuristic frontend response
type FrontendHeuristicResponse struct {
	UID             string      `json:"uid,omitempty"`
	Timestamp       string      `json:"ts,omitempty"`
	Type            string      `json:"type,omitempty"`
	Parameter       string      `json:"parameter,omitempty"`
	ParentHeuristic []Heuristic `json:"parent,omitempty"`
	ChildHeuristics []Heuristic `json:"children,omitempty"`
	ClusterCount    int         `json:"clusterCount,omitempty"`
	Results         []struct {
		Origin []struct {
			UID         string `json:"uid,omitempty"`
			Timestamp   string `json:"ts,omitempty"`
			AddressHash string `json:"addresshash,omitempty"`
			TxHash      string `json:"txhash,omitempty"`
		} `json:"origin,omitempty"`
		Destinations []struct {
			UID       string `json:"uid,omitempty"`
			Timestamp string `json:"ts,omitempty"`
			TxHash    string `json:"txhash,omitempty"`
		} `json:"destinations,omitempty"`
	} `json:"results,omitempty"`
}

type FrontendTransactionResult struct {
	Timestamp string `json:"ts,omitempty"`
	Hash      string `json:"txhash,omitempty"`
}

// FrontendHeuristicCluster holds the results counts of a heuristic per cluster
type FrontendHeuristicCluster struct {
	Transactions       []FrontendTransactionResult `json:"txs,omitempty"`
	Attributions       []Attribution               `json:"attributions,omitempty"`
	CountForwardLookup int                         `json:"count,omitempty"`
}

type Attribution struct {
	Tag      string `json:"tag,omitempty"`
	IsPublic bool   `json:"isPublic"`
}

// FrontendHeuristicShort holds all result counts of a heuristic
type FrontendHeuristicShort struct {
	UID      string                     `json:"uid,omitempty"`
	Clusters []FrontendHeuristicCluster `json:"clusters,omitempty"`
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

type mergedClusterItem struct {
	clusterHash string
	clusterUIDs map[string]bool
}
