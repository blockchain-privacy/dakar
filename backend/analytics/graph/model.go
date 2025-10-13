// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"gonum.org/v1/gonum/graph"

	"time"
)

type TransactionNode struct {
	TS   time.Time
	id   int64
	Type string
}

func (n TransactionNode) ID() int64      { return n.id }
func (n TransactionNode) String() string { return ToHex(n.id) }

// AddressEdge is a graph edge containing the uid of an address.
type AddressEdge struct {
	F, T        graph.Node
	AddressUIDs []int64
}

// From returns the from-node of the edge.
func (a AddressEdge) From() graph.Node { return a.F }

// To returns the to-node of the edge.
func (a AddressEdge) To() graph.Node { return a.T }

// ReversedEdge returns a new Edge with the F and T fields
// swapped.
func (a AddressEdge) ReversedEdge() graph.Edge { return AddressEdge{F: a.T, T: a.F} }
