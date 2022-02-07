package db

import (
	"backend/external"

	"context"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

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

// DropTypeOldClusterStatusStatus removes the old cluster status type
func DropTypeOldClusterStatusStatus(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropOp:    api.Operation_TYPE,
		DropValue: "CMultiInputStatus",
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

// AlterSchemaAddHierarchicalMultiInputClusteringStatus adds the new multi-input clustering status field
func AlterSchemaAddHierarchicalMultiInputClusteringStatus(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			isclustering: bool .
			lastclusteredid: int .

			# Clustering Hierarchical Multi Input Status
			type CHMIStatus {
				isclustering
				lastclusteredid
			}
		`,
	})
}

// AlterSchemaAddFlatMultiInputClusteringStatus adds the new multi-input clustering status field
func AlterSchemaAddFlatMultiInputClusteringStatus(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			isclustering: bool .
			lastclusteredid: int .
			
			# Clustering Flat Multi Input Status
			type CFMIStatus {
				isclustering
				lastclusteredid
			}
		`,
	})
}

// AlterSchemaAddMultiInputClusterType adds the new multi-input cluster type
func AlterSchemaAddMultiInputClusterType(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			cluster_type: string @index(hash) . # the cluster type
			cluster_transaction: uid @reverse . # the transaction which contains the address because of which the cluster was created
			cluster_addresses: [uid] @reverse . # all direct addresses, these occur in cluster_transaction
			cluster_children: [uid] @reverse . # all direct child clusters
			cluster_address_count: int . # number of connected addresses connected to this cluster (including child clusters)

			type Cluster {
				cluster_type
				cluster_transaction
				cluster_addresses
				cluster_children
				cluster_address_count
			}
		`,
	})
}

// AlterSchemaRemoveLowestBlockIdFromCrawlerStatus removes the lowestblock id from the crawler status
func AlterSchemaRemoveLowestBlockIdFromCrawlerStatus(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			type CrawlerStatus {
				iscrawling
				lastblockid
			}
		`,
	})
}

// DropLowestBlockId removes all data from the predicate lowestblockid
func DropLowestBlockId(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "lowestblockid",
	})
}

// AlterSchemaAddUserToCluster adds the user predicate to the cluster type
func AlterSchemaAddUserToCluster(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			cluster_user: uid @reverse . # the user which created the cluster
			cluster_ts: dateTime @index(day).
			
			type Cluster {
				cluster_type
				cluster_transaction
				cluster_addresses
				cluster_children
				cluster_address_count
				cluster_user
				cluster_ts
			}
		`,
	})
}

// AlterSchemaAddAttribution adds the attribution type
func AlterSchemaAddAttribution(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			attribution_user: uid @reverse . # the user which created the attribution
			attribution_tag: string @index(term, trigram) . # tag of the attribution
			attribution_address: uid @reverse . # the attribution's address
			attribution_ts: dateTime @index(day). # creation time of the attribution
			attribution_description: string @index(term, trigram) . # description of the attribution
			attribution_source: string @index(term, trigram) . # source (link) of the attribution
			attribution_category: string @index(term, trigram) . # category (e.g. exchange) of the attribution
			attribution_ispublic: bool @index(bool) . # set to true if the attribution is available to everyone

			type Attribution {
				attribution_user
				attribution_tag
				attribution_address
				attribution_ts
				attribution_description
				attribution_source
				attribution_category
				attribution_ispublic
			}
		`,
	})
}

// DropAllHeuristicPredicates drops all data of the heuristic predicates
func DropAllHeuristicPredicates(c external.Database) error {
	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "origin",
	}); err != nil {
		return err
	}

	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "destinations",
	})
}

// DropTypeTransactionHeuristicResult removes the type TransactionHeuristicResult
func DropTypeTransactionHeuristicResult(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropOp:    api.Operation_TYPE,
		DropValue: "TransactionHeuristicResult",
	})
}

// AlterSchemaAddNewHeuristicResult changes the heuristic result type
func AlterSchemaAddNewHeuristicResult(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			HeuristicResult.origin: uid @reverse .
			HeuristicResult.destinations: [uid] @reverse .
			HeuristicResult.cluster: string .

			type HeuristicResult {
				HeuristicResult.origin
				HeuristicResult.destinations
				HeuristicResult.cluster
			}
		`,
	})
}

// DropAllHeuristicPredicates2 drops all data of the heuristic predicates
func DropAllHeuristicPredicates2(c external.Database) error {
	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "type",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "parameter",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "h_transaction",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "results",
	}); err != nil {
		return err
	}

	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "parent_heuristic",
	})
}

// DropTypeTransactionHeuristic removes the type TransactionHeuristic
func DropTypeTransactionHeuristic(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropOp:    api.Operation_TYPE,
		DropValue: "TransactionHeuristic",
	})
}

// AlterSchemaAddHeuristic changes adds the new heuristic type
func AlterSchemaAddHeuristic(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			Heuristic.type: string @index(hash) .
			Heuristic.parameter: string .
			Heuristic.transaction: uid @reverse .
			Heuristic.results: [uid] @count @reverse .
			Heuristic.parent: [uid] @reverse .
			Heuristic.ts: dateTime @index(day) .

			type Heuristic {
				Heuristic.type
				Heuristic.parameter
				Heuristic.transaction
				Heuristic.results
				Heuristic.ts
				Heuristic.parent
			}
		`,
	})
}

// DropAllIsRLookupDone drops the predicate isrlookupdone
func DropAllIsRLookupDone(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		DropAttr: "isrlookupdone",
	})
}

// AlterSchemaUpdateTransaction updates the transaction type
func AlterSchemaUpdateTransaction(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
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
		`,
	})
}
