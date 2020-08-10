package db

import (
	"context"
	"encoding/json"
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
			
			id: int @index(int) .
			ts: dateTime .
			outputindex: int .
			inputindex: int .
			txtype: string .
			amount: string .
			iscoinbase: bool .

			iscrawling: bool .
			lastblockid: int .
			lowestblockid: int .

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

			type Output {
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
				lastblockid
				lowestblockid
			}
		`,
	})
}

// checks if a schema is set
func IsSchemaSet(c *dgo.Dgraph) (exists bool, err error) {
	query := "schema(type: Block){}"

	resp, err := c.NewReadOnlyTxn().Query(GetContext(), query)

	if err != nil {
		return
	}

	var r struct {
		Types []struct {
			Fields []struct {
				Name string `json:"name,omitempty"`
			} `json:"fields,omitempty"`
		} `json:"types,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.Types) == 1 && len(r.Types[0].Fields) > 0 {
		exists = true
	}

	return
}
