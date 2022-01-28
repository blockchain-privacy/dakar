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
			id: int @index(int) .
			ts: dateTime @index(day) .
			prevblock: uid @reverse .
			transactions: [uid] @reverse .

			type Block {
				blockhash
				id
				ts
				prevblock
				<~prevblock>
				transactions
			}

			txhash: string @index(hash) @upsert .
			privacytype: int @index(int) .
			isrlookupdone: bool @index(bool) .
			fee: int .
			tx_inputs: [uid] @reverse .
			tx_outputs: [uid] @reverse .

			type Transaction {
				txhash
				privacytype
				isrlookupdone
				fee
				<~transactions>
				tx_outputs
				tx_inputs
			}

			outputindex: int .
			inputindex: int .
			txtype: string .
			amount: int .
			iscoinbase: bool .
			keyasm: string @index(term) .
			sigasm: string .
			keyhex: string .
			sighex: string .

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

			addresshash: string @index(hash) @upsert .
			addr_outputs: [uid] @reverse .

			type Address {
				addresshash
				addr_outputs
			}

			iscrawling: bool .
			lastblockid: int .

			type CrawlerStatus {
				iscrawling
				lastblockid
			}

			isclassifying: bool .
			lastclassifiedid: int .

			type ClassifierStatus {
				isclassifying
				lastclassifiedid
			}

			isclustering: bool .
			lastclusteredid: int .

			# Clustering Hierarchical Multi Input Status
			type CHMIStatus {
				isclustering
				lastclusteredid
			}

			# Clustering Flat Multi Input Status
			type CFMIStatus {
				isclustering
				lastclusteredid
			}

			type: string @index(hash) .
			parameter: string .
			h_transaction: uid @reverse .
			results: [uid] @count @reverse .
			parent_heuristic: [uid] @reverse .

			type TransactionHeuristic {
				type
				parameter
				h_transaction
				results
				ts
				parent_heuristic
			}

			origin: uid @reverse .
			destinations: [uid] @reverse .

			type TransactionHeuristicResult {
				origin
				destinations
			}

			role_name: string @index(hash) .

			type Role {
				role_name
			}

			user_email: string @index(term, fulltext) .
			user_pwhash: string .
			user_roles: [uid] @reverse .
			user_created: dateTime @index(day) .
			user_modified: dateTime @index(day) .
			user_heuristics: [uid] @reverse .

			type User {
				user_email
				user_pwhash
				user_roles
				user_created
				user_modified
				user_heuristics
			}

			cluster_type: string @index(hash) . # the cluster type
			cluster_transaction: uid @reverse . # the transaction which contains the address because of which the cluster was created
			cluster_addresses: [uid] @reverse . # all direct addresses, these occur in cluster_transaction
			cluster_children: [uid] @reverse . # all direct child clusters
			cluster_user: uid @reverse . # the user which created the cluster
			cluster_address_count: int . # number of connected addresses connected to this cluster (including child clusters)
			cluster_ts: dateTime @index(day). # when the cluster was created, should only be used for custom clusters as for other clusters the creation time can be derived from the connected tx
			
			type Cluster {
				cluster_type
				cluster_transaction
				cluster_addresses
				cluster_children
				cluster_address_count
				cluster_user
				cluster_ts
			}

			attribution_user: uid @reverse . # the user which created the attribution
			attribution_tag: string @index(term, trigram) . # tag of the attribution
			attribution_address: uid @reverse . # the attribution's address
			attribution_ts: dateTime @index(day). # creation time of the attribution
			attribution_description: string @index(term, trigram) . # description of the attribution
			attribution_source: string @index(term, trigram) . # source (link) of the attribution
			attribution_category: string @index(term, trigram) . # category (e.g. exchange) of the attribution

			type Attribution {
				attribution_user
				attribution_tag
				attribution_address
				attribution_ts
				attribution_description
				attribution_source
				attribution_category
			}
		`,
	})
}

// IsSchemaSet checks if a schema is set
func IsSchemaSet(c external.Database) (exists bool, err error) {
	ctx, cancel := GetBackendContext()
	defer cancel()

	// we check if the type Block exists,
	// as this is the most basic type and thus should exist in all schemas
	resp, err := c.Query(ctx, "schema(type: Block){}", nil)
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
