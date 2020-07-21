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
			blockhash: string @index(hash) @upsert .
			txhash: string @index(hash) @upsert .
			addresshash: string @index(hash) @upsert .

			tx_inputs: [uid] @reverse .
			tx_outputs: [uid] @reverse .
			addr_outputs: [uid] @reverse .
			transactions: [uid] @reverse .
			prevblock: uid @reverse .
			
			id: int .
			ts: dateTime .
			index: int .
			txtype: string .
			amount: float .
			iscoinbase: bool .
			

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
