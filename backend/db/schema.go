package db

import (
	"backend/cmd/cliutil"

	"context"
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
			origins: [uid] @count @reverse .
			results: [uid] @count @reverse .
			prevblock: uid @reverse .
			h_transaction: uid @reverse .
			parent_heuristic: [uid] @reverse .
			
			id: int @index(int) .
			ts: dateTime @index(day) .
			outputindex: int .
			inputindex: int .
			txtype: string .
			amount: int .
			script: string @index(term, fulltext) .
			fee: int .
			iscoinbase: bool .
			privacytype: string @index(hash) .
			isrlookupdone: bool @index(bool) .

			iscrawling: bool .
			isanalyzing: bool .
			lastblockid: int .
			lowestblockid: int .
			lastanalysedid: int . 

			type: string @index(hash) .
			parameter: string .

			role_name: string @index(hash) .

			user_email: string @index(term, fulltext) .
			user_pwhash: string .
			user_roles: [uid] @reverse .
			user_created: dateTime @index(day) .
			user_modified: dateTime @index(day) .

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
				isrlookupdone
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
				script
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

			type TransactionHeuristic {
				type
				parameter
				h_transaction
				results
				ts
				parent_heuristic
			}

			type Role {
				role_name
			}

			type User {
				user_email
				user_pwhash
				user_roles
				user_created
				user_modified
			}
		`,
	})
}

// checks if a schema is set
func IsSchemaSet(c *dgo.Dgraph) (exists bool, err error) {
	query := "schema(type: Block){}"
	ctx, cancel := GetBackendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().Query(ctx, query)
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

func AlterSchemaAddUsers(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			role_name: string @index(hash) .

			user_email: string @index(term, fulltext) .
			user_pwhash: string .
			user_roles: [uid] @reverse .
			user_created: dateTime @index(day) .
			user_modified: dateTime @index(day) .

			type Role {
				role_name
			}

			type User {
				user_email
				user_pwhash
				user_roles
				user_created
				user_modified
			}
		`,
	})
}

func AlterSchemaAddReverseLookupDoneFlag(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			isrlookupdone: bool @index(bool) .

			type Transaction {
				txhash
				privacytype
				isrlookupdone
				fee
				origins
				<~origins>
				<~transactions>
				tx_outputs
				tx_inputs
			}
		`,
	})
}
