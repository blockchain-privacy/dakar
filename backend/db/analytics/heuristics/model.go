package heuristics

import (
	"backend/db"
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

// HeuristicResult holds one result (origin) of a heuristic and
// optionally the results of a forward lookup (destinations)
type HeuristicResult struct {
	Origin       db.UIDNode   `json:"HeuristicResult.origin,omitempty"`
	Destinations []db.UIDNode `json:"HeuristicResult.destinations,omitempty"`
	DType        []string     `json:"dgraph.type,omitempty"`
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
	UID                 string             `json:"uid,omitempty"`
	HeuristicType       string             `json:"Heuristic.type,omitempty"`
	Parameter           string             `json:"Heuristic.parameter,omitempty"`
	ClusterTypes        []string           `json:"Heuristic.clusterTypes,omitempty"`
	ExcludeAddresses    *bool              `json:"Heuristic.excludeAddresses"`
	ExcludeSpendingGaps *bool              `json:"Heuristic.excludeSpendingGaps"`
	UserUID             string             `json:"~User.heuristics,omitempty"`
	WorkspaceUID        string             `json:"~Workspace.heuristics,omitempty"`
	Transaction         db.UIDNode         `json:"Heuristic.transaction,omitempty"`
	Timestamp           string             `json:"Heuristic.ts,omitempty"`
	ParentHeuristic     []Heuristic        `json:"Heuristic.parent,omitempty"`
	ChildHeuristics     []Heuristic        `json:"~Heuristic.parent,omitempty"`
	Clusters            []HeuristicCluster `json:"Heuristic.clusters,omitempty"`

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
	return fmt.Sprintf("UID: %s, timestamp: %s, associated cluster: %s, output count: %d",
		h.UID, h.Timestamp, h.Cluster, len(h.Outputs))
}

// todo change to Options
type Config struct {
	Parameter string `json:"parameter,omitempty"`
	// ClusterTypes are used to cluster the results of the heuristic.
	// If cluster types are set to nil, the result will not be clustered.
	// If multiple cluster types are set, then the consolidation of these clusters will be used.
	ClusterTypes []clustering.ClusterType `json:"clusterTypes,omitempty"`
	// ExcludeAddresses controls whether certain addresses should be excluded from the lookups
	ExcludeAddresses bool `json:"excludeAddresses"`
	// ExcludeSpendingGaps controls whether mixing outputs with a spending gap should be traversed
	ExcludeSpendingGaps bool   `json:"excludeSpendingGaps"`
	TransactionHash     string `json:"transactionHash,omitempty"`
	// UserUID is the UID of the user who created this heuristic
	UserUID string `json:"-"`
}

func (c Config) String() string {
	return fmt.Sprintf("Transaction Hash: %s, Parameter: %s, cluster type: %v, exclude addresses: %v, exclude spending gaps: %v",
		c.TransactionHash, c.Parameter, c.ClusterTypes, c.ExcludeAddresses, c.ExcludeSpendingGaps)
}

// DatabaseHeuristicRequest holds all heuristic data which is set by the user
type DatabaseHeuristicRequest struct {
	UID                string  `json:"uid,omitempty"`
	Type               string  `json:"type,omitempty"`
	ParentHeuristicUID string  `json:"parentUID,omitempty"`
	Configuration      *Config `json:"config"`
}

type FrontendTransactionResult struct {
	Timestamp        string `json:"ts,omitempty"`
	Hash             string `json:"txhash,omitempty"`
	DestinationCount int    `json:"destinationCount,omitempty"`
}

// FrontendHeuristicCluster holds the results counts of a heuristic per cluster
type FrontendHeuristicCluster struct {
	Transactions []FrontendTransactionResult `json:"txs,omitempty"`
	Attributions []Attribution               `json:"attributions,omitempty"`
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

type mergedClusterItem struct {
	clusterHash string
	clusterUIDs map[string]bool
}
