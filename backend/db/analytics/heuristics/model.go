// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"fmt"
	"time"
)

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

type Options struct {
	// Type is the type of the heuristic
	Type      string `json:"type,omitempty"`
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

func (o Options) IsValid(_ bool) bool {
	return o.Type != "" && o.TransactionHash != ""
}

func (o Options) String() string {
	return fmt.Sprintf("Transaction Hash: %s, Parameter: %s, cluster type: %v, exclude addresses: %v, exclude spending gaps: %v",
		o.TransactionHash, o.Parameter, o.ClusterTypes, o.ExcludeAddresses, o.ExcludeSpendingGaps)
}

// DatabaseHeuristicRequest holds all heuristic data which is set by the user
type DatabaseHeuristicRequest struct {
	UID                string   `json:"uid,omitempty"`
	Type               string   `json:"type,omitempty"`
	ParentHeuristicUID string   `json:"parentUID,omitempty"`
	Configuration      *Options `json:"config"`
}

type mergedClusterItem struct {
	clusterHash string
	clusterUIDs map[string]bool
}
