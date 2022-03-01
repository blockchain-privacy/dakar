package db

import (
	"backend/cmd/cliutil"
	"backend/external"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"time"
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

func copyRoles(c external.Database) error {
	const query = `{
				v as var(func: type(Role)){
					roleValue as role_name
				}
			  }`

	req := &api.Request{
		Query: query,
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(v), 0))",
			SetNquads: []byte("uid(v) <Role.name> val(roleValue) ."),
		}},
		CommitNow: true,
	}
	if _, err := c.Mutate(context.Background(), req); err != nil {
		return err
	}

	return nil
}

// MigrateRole migrates the role predicates to the new dot notation
func MigrateRole(c external.Database) error {
	// add new empty predicate
	if err := c.Alter(context.Background(), &api.Operation{
		Schema: `
			role_name: string @index(hash) .
			Role.name: string @index(hash) .

			type Role {
				Role.name
				role_name
			}
		`,
	}); err != nil {
		return err
	}

	// copy data to new predicate
	if err := copyRoles(c); err != nil {
		return err
	}

	// drop old predicate
	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "role_name",
	}); err != nil {
		return err
	}

	// update type definition
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			Role.name: string @index(hash) .

			type Role {
				Role.name
			}
		`,
	})
}

func copyUserValuePredicates(c external.Database) error {
	const query = `{
				v as var(func: type(User)){
					emailVal as user_email
					pwVal as user_pwhash
					createdVal as user_created
					modifiedVal as user_modified
				}
			  }`

	req := &api.Request{
		Query: query,
		Mutations: []*api.Mutation{{
			Cond: "@if(gt(len(v), 0))",
			SetNquads: []byte("uid(v) <User.email> val(emailVal) .\n" +
				"uid(v) <User.pwhash> val(pwVal) .\n" +
				"uid(v) <User.created> val(createdVal) .\n" +
				"uid(v) <User.modified> val(modifiedVal) ."),
		}},
		CommitNow: true,
	}
	if _, err := c.Mutate(context.Background(), req); err != nil {
		return err
	}

	return nil
}

func copyUserUidPredicates(c external.Database) error {
	query := `{
				q(func: type(User)){
					uid
				}
			  }`

	resp, err := ReadOnlyTxWithRetry(c, time.Minute*2, query)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// json struct

	var r struct {
		Users []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return err
	}

	if len(r.Users) == 0 {
		return errors.New("no users")
	}

	for _, u := range r.Users {
		err := copyOneUserUidPredicates(c, u.Uid)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyOneUserUidPredicates(c external.Database, userUid string) error {
	const query = `query Q($uid:string){
				v as var(func: uid($uid)){
					roleVal as user_roles
					heuristicsVal as user_heuristics
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uid": userUid},
		Mutations: []*api.Mutation{
			{
				Cond:      "@if(gt(len(v), 0) and gt(len(roleVal), 0))",
				SetNquads: []byte("uid(v) <User.roles> uid(roleVal) ."),
			},
			{
				Cond:      "@if(gt(len(v), 0) and gt(len(heuristicsVal), 0))",
				SetNquads: []byte("uid(v) <User.heuristics> uid(heuristicsVal) ."),
			},
		},
		CommitNow: true,
	}
	if _, err := c.Mutate(context.Background(), req); err != nil {
		return err
	}

	return nil
}

// MigrateUser migrates the User predicates to the new dot notation
func MigrateUser(c external.Database) error {
	// add new empty predicate
	if err := c.Alter(context.Background(), &api.Operation{
		Schema: `
			user_email: string @index(term, fulltext) .
			user_pwhash: string .
			user_roles: [uid] @reverse .
			user_created: dateTime @index(day) .
			user_modified: dateTime @index(day) .
			user_heuristics: [uid] @reverse .

			User.email: string @index(term, fulltext) .
			User.pwhash: string .
			User.roles: [uid] @reverse .
			User.created: dateTime @index(day) .
			User.modified: dateTime @index(day) .
			User.heuristics: [uid] @reverse .

			type User {
				user_email
				user_pwhash
				user_roles
				user_created
				user_modified
				user_heuristics

				User.email
				User.pwhash
				User.roles
				User.created
				User.modified
				User.heuristics
			}

		`,
	}); err != nil {
		return err
	}

	// copy value data to new predicate
	if err := copyUserValuePredicates(c); err != nil {
		return err
	}

	// copy uid data to new predicate
	if err := copyUserUidPredicates(c); err != nil {
		return err
	}

	// drop old predicate
	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "user_email",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "user_pwhash",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "user_roles",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "user_created",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "user_modified",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "user_heuristics",
	}); err != nil {
		return err
	}

	// update type definition
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			User.email: string @index(term, fulltext) .
			User.pwhash: string .
			User.roles: [uid] @reverse .
			User.created: dateTime @index(day) .
			User.modified: dateTime @index(day) .
			User.heuristics: [uid] @reverse .
	
			type User {
				User.email
				User.pwhash
				User.roles
				User.created
				User.modified
				User.heuristics
			}
		`,
	})
}

func copyOneCluster(c external.Database, clusterUid string) error {
	const query = `query Q($uid:string){
				v as var(func: uid($uid)){
					typeVal as cluster_type
					txVal as cluster_transaction
					addrVal as cluster_addresses
					childVal as cluster_children
					countVal as cluster_address_count
					userVal as cluster_user
					tsVal as cluster_ts
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uid": clusterUid},
		Mutations: []*api.Mutation{
			{
				Cond:      "@if(gt(len(v), 0) and gt(len(typeVal), 0))",
				SetNquads: []byte("uid(v) <Cluster.type> val(typeVal) ."),
			},
			{
				Cond:      "@if(gt(len(v), 0) and gt(len(txVal), 0))",
				SetNquads: []byte("uid(v) <Cluster.transaction> uid(txVal) ."),
			},
			{
				Cond:      "@if(gt(len(v), 0) and gt(len(addrVal), 0))",
				SetNquads: []byte("uid(v) <Cluster.addresses> uid(addrVal) ."),
			},
			{
				Cond:      "@if(gt(len(v), 0) and gt(len(childVal), 0))",
				SetNquads: []byte("uid(v) <Cluster.children> uid(childVal) ."),
			},
			{
				Cond:      "@if(gt(len(v), 0) and gt(len(countVal), 0))",
				SetNquads: []byte("uid(v) <Cluster.addressCount> val(countVal) ."),
			},
			{
				Cond:      "@if(gt(len(v), 0) and gt(len(userVal), 0))",
				SetNquads: []byte("uid(v) <Cluster.user> uid(userVal) ."),
			},
			{
				Cond:      "@if(gt(len(v), 0) and gt(len(tsVal), 0))",
				SetNquads: []byte("uid(v) <Cluster.ts> val(tsVal) ."),
			},
		},
		CommitNow: true,
	}
	if _, err := c.Mutate(context.Background(), req); err != nil {
		return err
	}

	return nil
}

func copyCluster(c external.Database) error {
	query := `{
				q(func: type(Cluster))@filter(not has(Cluster.type)){
					uid
				}
			  }`

	resp, err := ReadOnlyTxWithRetry(c, time.Minute*2, query)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// json struct
	var r struct {
		Clusters []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return err
	}

	if len(r.Clusters) == 0 {
		return errors.New("no clusterse")
	}

	info("migrating", len(r.Clusters), "clusters")

	for _, cluster := range r.Clusters {
		if err := copyOneCluster(c, cluster.Uid); err != nil {
			return err
		}
	}

	return nil
}

// MigrateCluster migrates the Cluster predicates to the new dot notation
func MigrateCluster(c external.Database) error {
	// add new empty predicate
	if err := c.Alter(context.Background(), &api.Operation{
		Schema: `
			Cluster.type: string @index(hash) . # the cluster type
			Cluster.transaction: uid @reverse . # the transaction which contains the address because of which the cluster was created
			Cluster.addresses: [uid] @reverse . # all direct addresses, these occur in cluster_transaction
			Cluster.children: [uid] @reverse . # all direct child clusters
			Cluster.user: uid @reverse . # the user which created the cluster
			Cluster.addressCount: int . # number of connected addresses connected to this cluster (including child clusters)
			Cluster.ts: dateTime @index(day). # when the cluster was created, should only be used for custom clusters as for other clusters the creation time can be derived from the connected tx
	
			cluster_type: string @index(hash) . # the cluster type
			cluster_transaction: uid @reverse . # the transaction which contains the address because of which the cluster was created
			cluster_addresses: [uid] @reverse . # all direct addresses, these occur in cluster_transaction
			cluster_children: [uid] @reverse . # all direct child clusters
			cluster_user: uid @reverse . # the user which created the cluster
			cluster_address_count: int . # number of connected addresses connected to this cluster (including child clusters)
			cluster_ts: dateTime @index(day). # when the cluster was created, should only be used for custom clusters as for other clusters the creation time can be derived from the connected tx
	
			type Cluster {
				Cluster.type
				Cluster.transaction
				Cluster.addresses
				Cluster.children
				Cluster.addressCount
				Cluster.user
				Cluster.ts
	
				cluster_type
				cluster_transaction
				cluster_addresses
				cluster_children
				cluster_address_count
				cluster_user
				cluster_ts
			}
		`,
	}); err != nil {
		return err
	}

	// copy value data to new predicate
	if err := copyCluster(c); err != nil {
		return err
	}

	// drop old predicate
	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "cluster_type",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "cluster_transaction",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "cluster_addresses",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "cluster_children",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "cluster_address_count",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "cluster_user",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "cluster_ts",
	}); err != nil {
		return err
	}

	// update type definition
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
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
		`,
	})
}

// MigrateHeuristics migrates the heuristics predicates to the new structure (including attributions and clusters)
func MigrateHeuristics(c external.Database) error {
	// drop old predicate
	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "Heuristic.results",
	}); err != nil {
		return err
	}

	if err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "HeuristicResult.cluster",
	}); err != nil {
		return err
	}

	// update type definition
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			Heuristic.type: string @index(hash) .
			Heuristic.parameter: string .
			Heuristic.transaction: uid @reverse .
			Heuristic.clusters: [uid] @count @reverse .
			Heuristic.parent: [uid] @reverse .
			Heuristic.ts: dateTime @index(day) .

			type Heuristic {
				Heuristic.type
				Heuristic.parameter
				Heuristic.transaction
				Heuristic.clusters
				Heuristic.ts
				Heuristic.parent
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
		`,
	})
}

// AlterSchemaAddHeuristicProperties changes adds the new heuristic type
func AlterSchemaAddHeuristicProperties(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
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
		`,
	})
}

// AlterSchemaAddAddressExclusion changes adds the new addressExclusions predicate
func AlterSchemaAddAddressExclusion(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
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
		`,
	})
}
