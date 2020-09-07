package db

import (
	"context"
	"dashrpc/cmd/cliutil"
	"encoding/json"
	"fmt"
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
			origins: [uid] @reverse .
			prevblock: uid @reverse .
			
			id: int @index(int) .
			ts: dateTime @index(day) .
			outputindex: int .
			inputindex: int .
			txtype: string .
			amount: int .
			fee: int .
			iscoinbase: bool .
			#todo check if index is actually helpful
			privacytype: string @index(hash) .

			iscrawling: bool .
			isanalyzing: bool .
			lastblockid: int .
			lowestblockid: int .
			lastanalysedid: int . 

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
				privacytype
				fee
				origins
				<~origins>
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

			type CrawlerStatus {
				iscrawling
				lastblockid
				lowestblockid
			}

			type AnalyzerStatus {
				isanalyzing
				lastanalysedid
			}
		`,
	})
}

// checks if a schema is set
func IsSchemaSet(c *dgo.Dgraph) (exists bool, err error) {
	query := "schema(type: Block){}"

	resp, err := c.NewReadOnlyTxn().Query(GetBackendContext(), query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Types) == 1 && len(r.Types[0].Fields) > 0 {
		exists = true
	}

	return
}
