package db

import (
	"backend/cmd/cliutil"
	"backend/external"

	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

// SetupSchema installs a schema into dgraph
func SetupSchema(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			blockhash: string @index(hash) @upsert .
			txhash: string @index(hash) @upsert .
			addresshash: string @index(hash) @upsert .

			tx_inputs: [uid] @reverse .
			tx_outputs: [uid] @reverse .
			addr_outputs: [uid] @reverse .
			transactions: [uid] @reverse .
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
			isclassifying: bool .
			isclustering: bool .
			lastblockid: int .
			lowestblockid: int .
			lastclassifiedid: int .
			lastclusteredid: int .

			type: string @index(hash) .
			parameter: string .

			role_name: string @index(hash) .

			user_email: string @index(term, fulltext) .
			user_pwhash: string .
			user_roles: [uid] @reverse .
			user_created: dateTime @index(day) .
			user_modified: dateTime @index(day) .
			user_heuristics: [uid] @reverse .

			origin: uid @reverse .
			destinations: [uid] @reverse .

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

			type ClassifierStatus {
				isclassifying
				lastclassifiedid
			}

			type CMultiInputStatus {
				isclustering
				lastclusteredid
			}

			type TransactionHeuristic {
				type
				parameter
				h_transaction
				results
				ts
				parent_heuristic
			}

			type TransactionHeuristicResult {
				origin
				destinations
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
func IsSchemaSet(c external.Database) (exists bool, err error) {
	query := "schema(type: Block){}"
	ctx, cancel := GetBackendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, nil)
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

// AlterSchemaAddClassifier adds the new classifier status field
func AlterSchemaAddClassifier(c external.Database) error {
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

// AlterSchemaChangePrivacyTypePredicate adds the privacy type predicate
func AlterSchemaChangePrivacyTypePredicate(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			privacytype: int @index(int) .
		`,
	})
}

// DropAllPrivacyTypes drops the all data of the predicate privacytype
func DropAllPrivacyTypes(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "privacytype",
	})
}

// DropAllOrigins removes all data from the predicate origins
func DropAllOrigins(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "origins",
	})
}

// DropIsAnalyzing removes all data from the predicate isanalyzing
func DropIsAnalyzing(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "isanalyzing",
	})
}

// DropLastAnalysedID removes all data from the predicate lastanalysedid
func DropLastAnalysedID(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "lastanalysedid",
	})
}

// DropTypeAnalyzerStatus removes the type AnalyzerStatus
func DropTypeAnalyzerStatus(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropOp:    api.Operation_TYPE,
		DropValue: "AnalyzerStatus",
	})
}

// AlterSchemaSetTransactionType adds the privacytype predicate to the transaction type
func AlterSchemaSetTransactionType(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			type Transaction {
				txhash
				privacytype
				isrlookupdone
				fee
				<~transactions>
				tx_outputs
				tx_inputs
			}
		`,
	})
}

// AlterSchemaAddHeuristicResult adds the heuristic result type
func AlterSchemaAddHeuristicResult(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			origin: uid @reverse .
			destinations: [uid] @reverse .

			type TransactionHeuristicResult {
				origin
				destinations
			}
		`,
	})
}

// AlterSchemaAddMultiInputClusteringStatus adds the new multi-input clustering status field
func AlterSchemaAddMultiInputClusteringStatus(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			isclustering: bool .
			lastclusteredid: int .

			type CMultiInputStatus {
				isclustering
				lastclusteredid
			}
		`,
	})
}
