package graph

import (
	"backend/constants"
	"gonum.org/v1/gonum/graph"

	"time"
)

type transactionNode struct {
	ts          time.Time
	id          int64
	privacyType constants.PrivacyType
}

func (n transactionNode) ID() int64      { return n.id }
func (n transactionNode) String() string { return toHex(n.id) }

// addressEdge is a graph edge containing the uid of an address.
type addressEdge struct {
	F, T        graph.Node
	addressUIDs []int64
}

// From returns the from-node of the edge.
func (a addressEdge) From() graph.Node { return a.F }

// To returns the to-node of the edge.
func (a addressEdge) To() graph.Node { return a.T }

// ReversedEdge returns a new Edge with the F and T fields
// swapped.
func (a addressEdge) ReversedEdge() graph.Edge { return addressEdge{F: a.T, T: a.F} }
