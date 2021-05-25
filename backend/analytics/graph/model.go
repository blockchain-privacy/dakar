package graph

import (
	"backend/constants"

	"time"
)

type transactionNode struct {
	ts          time.Time
	id          int64
	privacyType constants.PrivacyType
}

func (n transactionNode) ID() int64      { return n.id }
func (n transactionNode) String() string { return toHex(n.id) }

type addressGraphNode struct {
	id        int64
	isAddress bool
}

func (a addressGraphNode) ID() int64      { return a.id }
func (a addressGraphNode) String() string { return toHex(a.id) }
