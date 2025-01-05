package analytics

import (
	"github.com/qrest/gomisc/serror"
	"time"
)

// ConnectedNodeRequest is the request for ConnectedNode
type ConnectedNodeRequest struct {
	UID             string `json:"uid"`
	TransactionType string `json:"Transaction.type"`
	Block           []struct {
		TS time.Time `json:"ts"`
	} `json:"block"`
	Inputs []struct {
		Addresses []struct {
			UID string `json:"uid"`
		} `json:"~addr_outputs,omitempty"`
		InputTransactions []struct {
			UID string `json:"uid"`
		} `json:"~tx_outputs,omitempty"`
	} `json:"i"`
}

func (c ConnectedNodeRequest) toConnectedNode() (*ConnectedNode, error) {
	if len(c.Block) != 1 {
		return nil, serror.FromStrWithContext("invalid block count",
			"transaction uid", c.UID, "block count", len(c.Block))
	}

	node := ConnectedNode{
		UID:  c.UID,
		Type: c.TransactionType,
		TS:   c.Block[0].TS,
	}

	for _, i := range c.Inputs {
		if len(i.Addresses) != 1 {
			return nil, serror.FromStrWithContext("input does not have exactly one address",
				"transaction uid", c.UID, "address count", len(i.Addresses))
		}

		if len(i.InputTransactions) != 1 {
			return nil, serror.FromStrWithContext("input does not have exactly one input transaction",
				"transaction uid", c.UID, "input transaction count", len(i.InputTransactions))
		}

		node.Inputs = append(node.Inputs, struct {
			Address          string
			InputTransaction string
		}{
			Address:          i.Addresses[0].UID,
			InputTransaction: i.InputTransactions[0].UID,
		})
	}

	return &node, nil
}

// ConnectedNode holds data for the current node and all connections on the input side
type ConnectedNode struct {
	UID    string
	Type   string
	TS     time.Time
	Inputs []struct {
		Address          string
		InputTransaction string
	}
}

// Node holds data for the current node
type Node struct {
	UID             string `json:"uid"`
	TransactionType string `json:"Transaction.type"`
	Block           []struct {
		TS time.Time `json:"ts"`
	} `json:"block"`
}

// AddressNode can hold data for an address or transaction
type AddressNode struct {
	UID    string `json:"uid"`
	Inputs []struct {
		UID string `json:"uid"`
	} `json:"i"`
}

// MixingActivity contains the timestamp and type of a classified transaction
type MixingActivity struct {
	TransactionHash string `json:"txhash"`
	TransactionType string `json:"txtype,omitempty"`
	Block           []struct {
		BlockTimestamp string `json:"ts,omitempty"`
	} `json:"block,omitempty"`
	InputTransactions []struct {
		TransactionHash string `json:"txhash"`
	} `json:"input_txs,omitempty"`
}
