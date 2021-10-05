package analytics

import (
	"backend/constants"
	"time"
)

// ConnectedNode holds data for the current node and all connections on the input side
type ConnectedNode struct {
	UID         string                `json:"uid"`
	PrivacyType constants.PrivacyType `json:"privacytype"`
	Block       []struct {
		Ts time.Time `json:"ts"`
	} `json:"block"`
	Inputs []struct {
		UID string `json:"uid"`
	} `json:"i"`
}

// Node holds data for the current node
type Node struct {
	UID         string                `json:"uid"`
	PrivacyType constants.PrivacyType `json:"privacytype"`
	Block       []struct {
		Ts time.Time `json:"ts"`
	} `json:"block"`
}

// AddressNode can hold data for an address or transaction
type AddressNode struct {
	UID    string `json:"uid"`
	Inputs []struct {
		UID string `json:"uid"`
	} `json:"i"`
}

// MixingActivity contains the timestamp and privacytype of a privacy transaction
type MixingActivity struct {
	PrivacyType    int64  `json:"privacytype,omitempty"`
	BlockTimestamp string `json:"ts,omitempty"`
}
