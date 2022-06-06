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
			Meta.creationTime: dateTime . # The time when this database was initialized
			Meta.blockchainMode: string . # The blockchain mode of this database. Possible values: Bitcoin, Dash, Doge
			Meta.schemaVersion: int . # The schema version of this database. 

			type Meta {
				Meta.creationTime
				Meta.blockchainMode
				Meta.schemaVersion
			}

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
			fee: int .
			tx_inputs: [uid] @reverse .
			tx_outputs: [uid] @reverse .

			type Transaction {
				txhash
				privacytype
				fee
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

			Heuristic.type: string @index(hash) .
			Heuristic.parameter: string .
			Heuristic.transaction: uid @reverse .
			Heuristic.clusters: [uid] @count @reverse .
			Heuristic.parent: [uid] @reverse .
			Heuristic.ts: dateTime @index(day) .
			Heuristic.clusterTypes: [string] .
			Heuristic.excludeAddresses: bool .

			type Heuristic {
				Heuristic.type
				Heuristic.parameter
				Heuristic.transaction
				Heuristic.clusters
				Heuristic.ts
				Heuristic.parent
				Heuristic.clusterTypes
				Heuristic.excludeAddresses
			}

			HeuristicResult.origin: uid @reverse .
			HeuristicResult.destinations: [uid] @reverse .

			type HeuristicResult {
				HeuristicResult.origin
				HeuristicResult.destinations
			}

			HeuristicCluster.results: [uid] @reverse .
			HeuristicCluster.attributions: [uid] @reverse .

			type HeuristicCluster {
				HeuristicCluster.results
				HeuristicCluster.attributions
			}

			Role.name: string @index(hash) .

			type Role {
				Role.name
			}

			User.email: string @index(term, fulltext) .
			User.pwhash: string .
			User.roles: [uid] @reverse .
			User.created: dateTime @index(day) .
			User.modified: dateTime @index(day) .
			User.heuristics: [uid] @reverse .
			User.addressExclusions: [uid] @count @reverse .
	
			type User {
				User.email
				User.pwhash
				User.roles
				User.created
				User.modified
				User.heuristics
				User.addressExclusions
			}

			Cluster.type: string @index(hash) . # the cluster type
			Cluster.transaction: uid @reverse . # the transaction which contains the address because of which the cluster was created
			Cluster.addresses: [uid] @reverse . # all direct addresses, these occur in cluster_transaction
			Cluster.children: [uid] @reverse . # all direct child clusters
			Cluster.user: uid @reverse . # the user which created the cluster
			Cluster.addressCount: int . # number of connected addresses connected to this cluster (including child clusters)
			Cluster.ts: dateTime @index(day). # when the cluster was created, should only be used for custom clusters as for other clusters the creation time can be derived from the connected tx
	
			type Cluster {
				Cluster.type
				Cluster.transaction
				Cluster.addresses
				Cluster.children
				Cluster.addressCount
				Cluster.user
				Cluster.ts
			}

			Attribution.user: uid @reverse . # the user which created the attribution
			Attribution.tag: string @index(term, trigram) . # tag of the attribution
			Attribution.address: uid @reverse . # the attribution's address
			Attribution.ts: dateTime @index(day). # creation time of the attribution
			Attribution.description: string @index(term, trigram) . # description of the attribution
			Attribution.source: string @index(term, trigram) . # source (link) of the attribution
			Attribution.category: string @index(term, trigram) . # category (e.g. exchange) of the attribution
			Attribution.isPublic: bool @index(bool) . # set to true if the attribution is available to everyone

			type Attribution {
				Attribution.user
				Attribution.tag
				Attribution.address
				Attribution.ts
				Attribution.description
				Attribution.source
				Attribution.category
				Attribution.isPublic
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
