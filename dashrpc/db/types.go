package db

import (
	"context"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// Install a schema into dgraph.
func SetupSchema(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			blockhash: string @index(exact) @upsert .
			txhash: string @index(exact) @upsert .
			addresshash: string @index(exact) @upsert .
			tx_inputs: [uid] @reverse .
			tx_outputs: [uid] @reverse .
			addr_outputs: [uid] @reverse .
			prevblock: uid @reverse .
			transactions: [uid] @reverse .
			id: string .
			ts: dateTime .
			index: string .
			txtype: string .
			amount: string .
			iscoinbase: string .
			

			type Block {
				blockhash
				id
				ts
				prevblock
				transactions
			}

			type Transaction {
				txhash
				tx_outputs
				tx_inputs
			}

			type TxOutput {
				index
				txtype
				amount
				iscoinbase
			}
			
			type Address {
				addresshash
				addr_outputs
			}
		`,
	})
}

type Block struct {
	Uid          string         `json:"uid,omitempty"`
	Hash         string         `json:"blockhash,omitempty"`
	Id           string         `json:"id,omitempty"`
	Timestamp    string         `json:"ts,omitempty"`
	PrevBlock    *Block         `json:"prevblock,omitempty"`
	Transactions []*Transaction `json:"transactions,omitempty"`
	DType        []string       `json:"dgraph.type,omitempty"`
}

type Transaction struct {
	Uid     string     `json:"uid,omitempty"`
	Outputs []TxOutput `json:"tx_outputs,omitempty"`
	Inputs  []TxOutput `json:"tx_inputs,omitempty"`
	Hash    string     `json:"txhash,omitempty"`
	DType   []string   `json:"dgraph.type,omitempty"`
}

type TxOutput struct {
	Uid        string   `json:"uid,omitempty"`
	Index      string   `json:"index,omitempty"`
	TxType     string   `json:"txtype,omitempty"`
	Amount     string   `json:"amount,omitempty"`
	IsCoinbase string   `json:"iscoinbase,omitempty"`
	DType      []string `json:"dgraph.type,omitempty"`
}

type Address struct {
	Uid     string     `json:"uid,omitempty"`
	Hash    string     `json:"addresshash,omitempty"`
	Outputs []TxOutput `json:"addr_outputs,omitempty"`
	DType   []string   `json:"dgraph.type,omitempty"`
}

type blockQuery struct {
	Q []Block `json:"q"`
}

type transactionQuery struct {
	Q []Transaction `json:"q"`
}

type addressQuery struct {
	Q []Address `json:"q"`
}
