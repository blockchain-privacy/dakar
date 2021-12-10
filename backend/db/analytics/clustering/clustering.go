package clustering

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/external"

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
)

// GetInputAddressesByBlock gets all input addresses per transaction by block id.
// The size of the returned slice can be zero in case the only transaction contained
// in the block is the coinbase transaction (no inputs) or all transaction are filtered out (mixing transactions).
func GetInputAddressesByBlock(c external.Database, blockID uint64, clusterType ClusterType) (transactions []ClusterTransaction, err error) {
	const query = `query Q($block:string,$ctype:string) {
				var(func: eq(id, $block)){
					# do not consider mixing transaction
					txs as transactions@filter(not between(privacytype,0,` + constants.StrPrivacyMixingLast + `))
				}

				q(func: uid(txs))@filter(gt(count(tx_inputs),1))@cascade{
					uid
					addr:tx_inputs@normalize{
						# cascade(uid) so addresses without associated clusters are still returned
						~addr_outputs{
							a as uid:uid
						}
					}
				}

				x(func: uid(a))@cascade{
					uid
					clusters: ~cluster_addresses@filter(eq(cluster_type,$ctype))@cascade(uid){
						uid
						cluster_address_count
						parents:~cluster_children{
							uid
						}
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$block": strconv.FormatUint(blockID, 10), "$ctype": string(clusterType)})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		TransactionToAddresses []TransactionWithAddresses `json:"q,omitempty"`
		AddressToClusters      []ClusterAddress           `json:"x,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.TransactionToAddresses) == 0 {
		return
	}

	// create address to cluster lookup map
	addressToCluster := make(map[string][]ClusterWithParent)
	for _, ac := range r.AddressToClusters {
		addressToCluster[ac.Uid] = ac.Clusters
	}

	// merge the two returned arrays
	for _, t := range r.TransactionToAddresses {
		// new transaction
		tx := ClusterTransaction{Uid: t.Uid}

		for _, a := range t.Addresses {
			ca := ClusterAddress{Uid: a.Uid}

			if _, ok := addressToCluster[a.Uid]; ok {
				ca.Clusters = append(ca.Clusters, addressToCluster[a.Uid]...)
			}

			tx.Addresses = append(tx.Addresses, ca)
		}

		transactions = append(transactions, tx)
	}

	return
}

// AddCustomClusters adds the given clusters to the database
func AddCustomClusters(c external.Database, clusters []CustomCluster) error {
	// validate data
	for _, cluster := range clusters {
		if cluster.Type == "" {
			return errors.New("cluster type is not set")
		}

		if len(cluster.Addresses) == 0 {
			return errors.New("cluster no addresses set")
		}
	}

	pb, err := json.Marshal(clusters)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return err
	}

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}
	err = db.TxWithRetry(c, time.Minute*5, req)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
}

// AddClusters adds the given clusters to the database
func AddClusters(c external.Database, clusters []Cluster, checkTx bool) error {
	// validate data
	for _, cluster := range clusters {
		if cluster.Type == "" {
			return errors.New("cluster type is not set")
		}

		if checkTx && cluster.Transaction.Uid == "" {
			return errors.New("cluster transaction is not set")
		}

		if len(cluster.Addresses) == 0 && len(cluster.Children) == 0 {
			return errors.New("cluster has no child clusters and no addresses set")
		}
	}

	pb, err := json.Marshal(clusters)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return err
	}

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}
	err = db.TxWithRetry(c, time.Minute*5, req)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
}

type DBOperation struct {
	NewCluster  Cluster
	OldClusters []string
}

// ProcessClusterOperations performs the given operations
func ProcessClusterOperations(c external.Database, operations []DBOperation) error {
	txn := c.NewTxn()
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute*5)
	defer func(txn *dgo.Txn, ctx context.Context) {
		err := txn.Discard(ctx)
		if err != nil {
			log.Println("error while discarding transaction:", err)
		}
	}(txn, ctx)
	defer cancelFunc()

	// step 1: set new clusters and add new addresses to existing clusters
	var clusters []Cluster
	for _, o := range operations {
		clusters = append(clusters, o.NewCluster)
	}

	pb, err := json.Marshal(clusters)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return err
	}

	// step 2: get all addresses of clusters which will be deleted and add them to the clusters from step 1

	// build query and create nquads
	var setNquads string
	var delNquads string
	query := "{\n"
	for i, o := range operations {
		if len(o.OldClusters) == 0 {
			continue
		}
		index := strconv.Itoa(i)
		query += "var(func:uid(" + db.CreateUIDEnum(o.OldClusters) + ")){a" + index + " as cluster_addresses}\n"
		setNquads += "<" + o.NewCluster.Uid + "> <cluster_addresses> uid(a" + index + ") .\n"

		for _, oc := range o.OldClusters {
			delNquads += "<" + oc + "> * * .\n"
		}
	}
	query += "}"

	existClusterMerges := setNquads != ""

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: !existClusterMerges,
	}
	err = db.ExistingTxWithRetry(txn, time.Minute*5, req)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if !existClusterMerges {
		return nil
	}

	req = &api.Request{
		Query: query,
		Mutations: []*api.Mutation{{
			SetNquads: []byte(setNquads),
		}},
	}
	err = db.ExistingTxWithRetry(txn, time.Minute*5, req)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// step 3: delete all merged clusters
	req = &api.Request{
		Mutations: []*api.Mutation{{
			DelNquads: []byte(delNquads),
		}},
		CommitNow: true,
	}
	err = db.ExistingTxWithRetry(txn, time.Minute*5, req)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
}

// GetHierarchicalClusterRoot returns the root of the cluster tree clusterUID is part of
func GetHierarchicalClusterRoot(c external.Database, clusterUID string) (rootCluster ClusterWithParent, err error) {
	const query = `query Q($uid:string) {
				var(func: uid($uid))@recurse{
					c as ~cluster_children
				}
				
				q(func: uid(c))@filter(eq(count(~cluster_children),0)){
					uid
					cluster_address_count
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$uid": clusterUID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Root []ClusterWithParent `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Root) != 1 {
		err = fmt.Errorf("invalid number of roots returned: %v", r.Root)
		return
	}

	rootCluster = r.Root[0]
	return
}

func getClusterQuery(maxAddresses int) string {
	var limiter string

	if maxAddresses > 0 {
		limiter = "(first:" + strconv.Itoa(maxAddresses) + ")"
	}

	return `q(func: uid(c)){
				cluster_type
				cluster_address_count
				cluster_transaction@normalize{
					txhash:txhash
					~transactions{
						bhash:blockhash
						bid:id
						ts:ts
					}
				}
				cluster_addresses` + limiter + `{
					addresshash
					output_count: count(addr_outputs)
					spent_output_count: count(addr_outputs@filter(has(~tx_inputs)))
				}
		}`
}

// GetClusters returns cluster information for all clusters (except hmi clusters) associated with addressHash
func GetClusters(c external.Database, addressHash string, maxAddresses int) (clusters []FrontendCluster, err error) {
	query := `query Q($addressHash:string) {
				var(func:eq(addresshash,$addressHash)){
					c as ~cluster_addresses
				}
				` + getClusterQuery(maxAddresses) + "}"

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$addressHash": addressHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Clusters []FrontendClusterRequest `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, cluster := range r.Clusters {
		if len(cluster.Transaction) > 1 {
			err = fmt.Errorf("invalid transaction count: %d", len(cluster.Transaction))
			return
		}

		frontendCluster := FrontendCluster{
			Type:         cluster.Type,
			AddressCount: cluster.AddressCount,
			Addresses:    cluster.Addresses,
		}

		// Transaction can be not set if the cluster was created by a user
		if cluster.Transaction != nil {
			frontendCluster.TransactionHash = cluster.Transaction[0].TransactionHash
			frontendCluster.BlockID = cluster.Transaction[0].BlockID
			frontendCluster.BlockHash = cluster.Transaction[0].BlockHash
			frontendCluster.Timestamp = cluster.Transaction[0].Timestamp
		}

		clusters = append(clusters, frontendCluster)
	}

	return
}

// GetCommonClusters returns cluster information for all clusters (except hmi clusters)
// shared by addressHash1 and addressHash2
func GetCommonClusters(c external.Database, addressHash1 string, addressHash2 string, maxAddresses int) (clusters []FrontendCluster,
	err error) {
	query := `query Q($a1:string,$a2:string) {
				var(func:eq(addresshash,$a1)){
					c1 as ~cluster_addresses@filter(not eq(cluster_type,` + string(TypeHMI) + `))
				}

				var(func:eq(addresshash,$a2)){
					c as ~cluster_addresses@filter(not eq(cluster_type,` + string(TypeHMI) + `) and uid(c1))
				}
				` + getClusterQuery(maxAddresses) + "}"

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$a1": addressHash1, "$a2": addressHash2})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Clusters []FrontendClusterRequest `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, cluster := range r.Clusters {
		if len(cluster.Transaction) != 1 {
			err = fmt.Errorf("invalid transaction count: %d", len(cluster.Transaction))
			return
		}
		clusters = append(clusters, FrontendCluster{
			Type:            cluster.Type,
			AddressCount:    cluster.AddressCount,
			TransactionHash: cluster.Transaction[0].TransactionHash,
			BlockID:         cluster.Transaction[0].BlockID,
			BlockHash:       cluster.Transaction[0].BlockHash,
			Timestamp:       cluster.Transaction[0].Timestamp,
			Addresses:       cluster.Addresses,
		})

	}

	return
}

// GetHMIClusters returns all connected hierarchical multi-input cluster to the
// given address and the uid of the cluster directly connected to the address
func GetHMIClusters(c external.Database, addressHash string) (addressCluster string, clusters []FrontendHMICluster, err error) {
	const query = string(`query Q($addressHash:string) {
							var(func: eq(addresshash,$addressHash)){
								hmi as ~cluster_addresses@filter(eq(cluster_type,` + TypeHMI + `))
							}
							
							var(func: uid(hmi))@recurse{
								s as cluster_children
								v as ~cluster_children
							}

							x(func: uid(hmi)){
								uid
							}
							
							q(func: uid(s,v)){
								uid
								cluster_address_count
								cluster_transaction{
									txhash
								}
								cluster_children{
									uid
								}
								~cluster_children{
									uid
								}
							}
						  }`)

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$addressHash": addressHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Clusters []struct {
			Uid          string `json:"uid,omitempty"`
			AddressCount int    `json:"cluster_address_count,omitempty"`
			Transaction  struct {
				TxHash string `json:"txhash,omitempty"`
			} `json:"cluster_transaction,omitempty"`
			Children []SubCluster `json:"cluster_children,omitempty"`
			Parent   []SubCluster `json:"~cluster_children,omitempty"`
		} `json:"q,omitempty"`
		AddressCluster []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"x,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.AddressCluster) == 0 {
		// no clusters found
		return
	}

	if len(r.AddressCluster) > 1 {
		err = errors.New("too many clusters associated with address")
		return
	}

	addressCluster = r.AddressCluster[0].Uid

	for _, cluster := range r.Clusters {
		if len(cluster.Parent) > 1 {
			err = fmt.Errorf("cluster %s has multiple parents: %v", cluster.Uid, cluster.Parent)
			return
		}

		var parentUID string
		if len(cluster.Parent) == 1 {
			parentUID = cluster.Parent[0].Uid
		}

		var childClusters []string
		for _, child := range cluster.Children {
			childClusters = append(childClusters, child.Uid)
		}

		clusters = append(clusters, FrontendHMICluster{
			Uid:             cluster.Uid,
			AddressCount:    cluster.AddressCount,
			TransactionHash: cluster.Transaction.TxHash,
			Parent:          parentUID,
			Children:        childClusters,
		})
	}

	return
}

// GetUserClusters returns all clusters of a user
func GetUserClusters(c external.Database, userID string) (clusters []FrontendUserCluster, err error) {
	const query = `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					c as ~cluster_user
				}

				q(func: uid(c)){
					uid
					cluster_addresses {
						addresshash
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$user": userID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Clusters []struct {
			Uid              string `json:"uid,omitempty"`
			ClusterAddresses []struct {
				Hash string `json:"addresshash,omitempty"`
			} `json:"cluster_addresses,omitempty"`
		} `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, cluster := range r.Clusters {
		var addresses []string
		for _, a := range cluster.ClusterAddresses {
			addresses = append(addresses, a.Hash)
		}
		clusters = append(clusters, FrontendUserCluster{
			Uid:       cluster.Uid,
			Addresses: addresses,
		})
	}

	return
}

// DeleteCluster deletes the given cluster
func DeleteCluster(c external.Database, userID string, clusterUID string) (err error) {
	req := &api.Request{
		Query: `query Q($user:string,$cluster:string) {
				var(func:uid($user))@filter(type(User)){
					c as ~cluster_user@filter(uid($cluster))
				}
			  }`,
		Vars: map[string]string{"$user": userID, "$cluster": clusterUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(c) * * ."),
		}},
		CommitNow: true,
	}
	resp, txErr := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	if resp.GetMetrics().NumUids["mutation_cost"] == 0 {
		return errors.New("nothing was deleted")
	}

	return
}

// DeleteAllClusters deletes all clusters of a given user
func DeleteAllClusters(c external.Database, userID string) (err error) {
	req := &api.Request{
		Query: `query Q($user:string,$cluster:string) {
				var(func:uid($user))@filter(type(User)){
					c as ~cluster_user
				}
			  }`,
		Vars: map[string]string{"$user": userID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(c) * * ."),
		}},
		CommitNow: true,
	}
	resp, txErr := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	if resp.GetMetrics().NumUids["mutation_cost"] == 0 {
		return errors.New("nothing was deleted")
	}

	return
}
