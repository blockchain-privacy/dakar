// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"backend/external"
	"context"
	"encoding/json"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// SchemaVersion is the version of the schema defined in SetupSchema.
// If SchemaVersion is higher than Meta.schemaVersion (which is stored in the db, see schema below),
// then a database upgrade is required.
// Use status.SetSchemaVersion to increase the schema version directly,
// or apply upgrades with upgrades.UpgradeDatabase which sets the version appropriately.
const SchemaVersion = 16

// SetupSchema installs a schema into Dgraph
func SetupSchema(c external.Database) error {
	return c.SetSchema(context.Background(), `
			Meta.creationTime: dateTime . # The time when this database was initialized
			Meta.blockchainMode: string . # The blockchain mode of this database.
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
			fee: int .
			tx_inputs: [uid] @reverse .
			tx_outputs: [uid] @reverse .
			Transaction.type: string @index(hash) .

			type Transaction {
				txhash
				Transaction.type
				fee
				tx_outputs
				tx_inputs
			}

			outputindex: int .
			inputindex: int .
			amount: int .
			iscoinbase: bool .

			type Output {
				outputindex
				inputindex
				amount
				iscoinbase
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

			HeuristicCluster.results: [uid] @reverse .
			HeuristicCluster.attributions: [uid] @reverse .

			type HeuristicCluster {
				HeuristicCluster.results
				HeuristicCluster.attributions
			}

			User.workspaces: [uid] @reverse .

			type User {
				User.workspaces
			}
			
			Workspace.name: string . # the workspace name
			Workspace.ts: dateTime @index(day) . # modification date of the workspace
			Workspace.state: string . # JSON encoded state of the workspace
			Workspace.clusterHeight: int . # last clustered block at which this workspace was updated
			Workspace.selectors: [uid] @reverse . # selectors which are managed by this workspace

			type Workspace {
				Workspace.name
				Workspace.ts
				Workspace.state
				Workspace.clusterHeight
				Workspace.selectors
			}
			
			Selector.created: dateTime @index(day) .  # creation date of the selector
			Selector.modified: dateTime @index(day) .  # modification date of the selector
			Selector.type: string @index(hash) . # type of the selector
			Selector.status: string @index(hash) . # status of the selector (waiting, error, success)
			Selector.errorCode: string . # reason for the error status, if any
			Selector.parent: uid @reverse . # parent node from which a selector can use data
			Selector.options: string . # JSON encoded options of the selector
			Selector.results: [uid] @reverse . # results of the selector
			Selector.totalResultCount: int . # number of results found by the selector (can be higher than number of stored results)

			type Selector {
				Selector.created
				Selector.modified
				Selector.type
				Selector.status
				Selector.errorCode
				Selector.parent
				Selector.options
				Selector.results
				Selector.totalResultCount
			}

			Cluster.type: string . # the cluster type
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
		`)
}

// IsSchemaSet checks if a schema is set
func IsSchemaSet(c external.Database) (exists bool, err error) {
	ctx, cancel := GetTaskContext()
	defer cancel()

	// we check if the type Block exists,
	// as this is the most basic type and thus should exist in all schemas
	resp, err := c.Query(ctx, "schema(type: Block){}", nil)
	if err != nil {
		err = serror.New(err)
		return
	}

	var r struct {
		Types []struct {
			Fields []struct {
				Name string `json:"name,omitempty"`
			} `json:"fields,omitempty"`
		} `json:"types,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Types) == 1 && len(r.Types[0].Fields) > 0 {
		exists = true
	}

	return
}
