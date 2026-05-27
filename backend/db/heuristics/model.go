package heuristics

import (
	"fmt"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics/attribution"
	"time"
)

type mergedClusterItem struct {
	clusterHash string
	clusterUIDs map[string]bool
}

// ClusterDType is the dgraph database type for the HeuristicCluster type
const ClusterDType = "HeuristicCluster"

// HeuristicCluster holds a set of results (origins) of a heuristic
// which belong to the same cluster (or merged cluster) and its attributions
type HeuristicCluster struct {
	Results      []db.UIDNode              `json:"HeuristicCluster.results,omitempty"`
	Attributions []attribution.Attribution `json:"HeuristicCluster.attributions,omitempty"`
	DType        []string                  `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (c *HeuristicCluster) SetDType() {
	c.DType = []string{ClusterDType}
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
	Cluster ClusterUID `json:"-"`
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
