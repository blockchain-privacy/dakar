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
			outputindex: int .
			inputindex: int .
			txtype: string .
			amount: string .
			iscoinbase: bool .

			iscrawling: bool .
			

			type Block {
				blockhash
				id
				ts
				prevblock
				<~prevblock>
				transactions
			}

			type Transaction {
				txhash
				<~transactions>
				tx_outputs
				tx_inputs
			}

			type TxOutput {
				outputindex
				inputindex
				txtype
				amount
				iscoinbase
				<~tx_inputs>
				<~tx_outputs>
				<~addr_outputs>
			}
			
			type Address {
				addresshash
				addr_outputs
			}

			type Status {
				iscrawling
			}
		`,
	})
}
