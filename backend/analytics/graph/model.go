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

type addressNode struct {
	id int64
}

func (a addressNode) ID() int64      { return a.id }
func (a addressNode) String() string { return toHex(a.id) }
