package analytics

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Origin struct {
	Uid            string
	Hash           string
	BlockHash      string
	BlockId        uint64
	BlockTimestamp string
}

func (o Origin) String() string {
	return fmt.Sprintf("Uid: %s, Hash: %s, block hash: %s, block id: %d, timestamp: %s",
		o.Uid, o.Hash, o.BlockHash, o.BlockId, o.BlockTimestamp)
}

type input struct {
	Transaction []transaction `json:"~tx_outputs,omitempty"`
}
type transaction struct {
	Hash        string  `json:"txhash,omitempty"`
	PrivacyType string  `json:"privacytype,omitempty"`
	Inputs      []input `json:"tx_inputs,omitempty"`
}

type PathElement struct {
	IsOrigin bool
	Hash     string
}

type TransactionPath []PathElement

func (p PathElement) equal(a PathElement) bool {
	return p.Hash == a.Hash
}

func (thisPath TransactionPath) equal(otherPath TransactionPath) bool {
	if len(thisPath) != len(otherPath) {
		return false
	}
	for i := range thisPath {
		if !thisPath[i].equal(otherPath[i]) {
			return false
		}
	}
	return true
}

// removes all mixing transactions on the tail of the path
func (thisPath TransactionPath) cutTail() TransactionPath {
	newPath := make(TransactionPath, len(thisPath))
	if len(thisPath) == 0 {
		return newPath
	}

	copy(newPath, thisPath)

	for i := len(thisPath) - 1; i >= 0; i-- {
		if thisPath[i].IsOrigin {
			break
		}

		newPath = newPath[:len(newPath)-1]
	}

	return newPath
}

func (thisPath TransactionPath) hash() string {
	var buffer bytes.Buffer
	for _, p := range thisPath {
		buffer.WriteString(p.Hash)
	}
	shaSum := sha256.Sum256([]byte(buffer.String()))

	return hex.EncodeToString(shaSum[:])
}
