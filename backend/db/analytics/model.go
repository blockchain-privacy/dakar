package analytics

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

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

// returns true if both PathElement's have the same transaction hash
func (p PathElement) equal(a PathElement) bool {
	return p.Hash == a.Hash
}

// returns true if all PathElement's of thisPath are equal to and have
// the same order to all PathElement's of otherPath
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

// creates a sha256 hash from all transaction hashes in thisPath
func (thisPath TransactionPath) hash() string {
	var buffer bytes.Buffer
	for _, p := range thisPath {
		buffer.WriteString(p.Hash)
	}
	shaSum := sha256.Sum256([]byte(buffer.String()))

	return hex.EncodeToString(shaSum[:])
}

type ConnectedNode struct {
	Uid   string `json:"uid"`
	Block []struct {
		Ts time.Time `json:"ts"`
	} `json:"block"`
	Inputs []struct {
		Uid string `json:"uid"`
	} `json:"i"`
}

type Node struct {
	Uid   string `json:"uid"`
	Block []struct {
		Ts time.Time `json:"ts"`
	} `json:"block"`
}
