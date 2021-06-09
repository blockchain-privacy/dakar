package analytics

import (
	"backend/constants"
	"time"
)

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

type Node struct {
	UID         string                `json:"uid"`
	PrivacyType constants.PrivacyType `json:"privacytype"`
	Block       []struct {
		Ts time.Time `json:"ts"`
	} `json:"block"`
}

type AddressNode struct {
	UID    string `json:"uid"`
	Inputs []struct {
		UID string `json:"uid"`
	} `json:"i"`
}
