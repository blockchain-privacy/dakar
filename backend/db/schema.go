package db

import (
	"backend/cmd/cliutil"

	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// SetupSchema installs a schema into dgraph
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
			fee: int .
			iscoinbase: bool .
			privacytype: int @index(int) .
			isrlookupdone: bool @index(bool) .
			keyasm: string @index(term) .
			sigasm: string .
			keyhex: string .
			sighex: string .

			iscrawling: bool .
			isanalyzing: bool .
			isclassifying: bool .
			lastblockid: int .
			lowestblockid: int .
			lastanalysedid: int . 
			lastclassifiedid: int .

			type: string @index(hash) .
			parameter: string .

			role_name: string @index(hash) .

			user_email: string @index(term, fulltext) .
			user_pwhash: string .
			user_roles: [uid] @reverse .
			user_created: dateTime @index(day) .
			user_modified: dateTime @index(day) .
			user_heuristics: [uid] @reverse .

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
				keyasm
				sigasm
				keyhex
				sighex
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

			type ClassifierStatus {
				isclassifying
				lastclassifiedid
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
				user_heuristics
			}
		`,
	})
}

// IsSchemaSet checks if a schema is set
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

func AlterSchemaAddClassifier(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			isclassifying: bool .
			lastclassifiedid: int .

			type ClassifierStatus {
				isclassifying
				lastclassifiedid
			}
		`,
	})
}

func AlterSchemaChangePrivacyTypePredicate(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			privacytype: int @index(int) .
		`,
	})
}

func DropAllPrivacyTypes(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "privacytype",
	})
}

func DropAllOrigins(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "origins",
	})
}

func AlterSchemaAddOriginsPredicate(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			origins: [uid] @count @reverse .
		`,
	})
}
