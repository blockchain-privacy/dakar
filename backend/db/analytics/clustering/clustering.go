package clustering

import (
	"backend/constants"
	"backend/db"
	"backend/external"
	"github.com/qrest/gomisc/serror"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
)

// GetInputAddressesByBlock gets all input addresses per transaction by block id.
// The size of the returned slice can be zero in case the only transaction contained
// in the block is the coinbase transaction (no inputs) or all transaction are filtered out (mixing transactions).
func GetInputAddressesByBlock(c external.Database, blockID uint64, clusterType ClusterType) (transactions []TransactionWithAddressClusters, err error) {
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
					clusters: ~Cluster.addresses@filter(eq(Cluster.type,$ctype))@cascade(uid){
						uid
						Cluster.addressCount
						parents:~Cluster.children{
							uid
						}
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$block": strconv.FormatUint(blockID, 10), "$ctype": string(clusterType)})
	if err != nil {
		return
	}

	var r struct {
		TransactionToAddresses []TransactionWithAddresses `json:"q,omitempty"`
		AddressToClusters      []AddressWithClusters      `json:"x,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.TransactionToAddresses) == 0 {
		return
	}

	// create address to cluster lookup map
	addressToCluster := make(map[string][]ClusterWithParent)
	for _, ac := range r.AddressToClusters {
		addressToCluster[ac.UID] = ac.Clusters
	}

	// merge the two returned arrays
	for _, t := range r.TransactionToAddresses {
		// new transaction
		tx := TransactionWithAddressClusters{UID: t.UID}

		for _, a := range t.Addresses {
			ca := AddressWithClusters{UID: a.UID}

			if _, ok := addressToCluster[a.UID]; ok {
				ca.Clusters = append(ca.Clusters, addressToCluster[a.UID]...)
			}

			tx.Addresses = append(tx.Addresses, ca)
		}

		transactions = append(transactions, tx)
	}

	return
}

// GetAddressesByBlock gets all addresses per transaction by block id.
func GetAddressesByBlock(c external.Database, fromBlockID uint64, toBlockID uint64,
	clusterType ClusterType) (transactions []TransactionWithInputOutputAddressCluster, err error) {
	const query = `query Q($from:string,$to:string,$ctype:string) {
				var(func: between(id, $from, $to)){
					txs as transactions
				}

				q(func: uid(txs)){
					uid
					privacytype
					input_addr:tx_inputs@normalize{
						~addr_outputs{
							a as uid:uid
						}
					}
					output_addr:tx_outputs@normalize{
						~addr_outputs{
							b as uid:uid
						}
					}
				}

				x(func: uid(a,b))@cascade{
					uid
					clusters: ~Cluster.addresses@filter(eq(Cluster.type,$ctype)){
						uid
						Cluster.addressCount
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$from": strconv.FormatUint(fromBlockID, 10),
			"$to": strconv.FormatUint(toBlockID, 10), "$ctype": string(clusterType)})
	if err != nil {
		return
	}

	var r struct {
		TransactionToAddresses []TransactionWithInputOutputAddresses `json:"q,omitempty"`
		AddressToClusters      []AddressWithClusters                 `json:"x,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.TransactionToAddresses) == 0 {
		return
	}

	// create address to cluster lookup map
	addressToCluster := make(map[string]BasicCluster)
	for _, ac := range r.AddressToClusters {
		if len(ac.Clusters) > 1 {
			err = serror.FromFormat("address with more than one FMI Cluster found: %s", ac.UID)
			return
		}

		addressToCluster[ac.UID] = BasicCluster{
			UID:          ac.Clusters[0].UID,
			AddressCount: ac.Clusters[0].AddressCount,
		}
	}

	// merge the two returned arrays
	for _, t := range r.TransactionToAddresses {
		// new transaction
		tx := TransactionWithInputOutputAddressCluster{UID: t.UID, PrivacyType: t.PrivacyType}

		for _, a := range t.InputAddresses {
			ca := AddressWithCluster{UID: a.UID}

			if addressCluster, ok := addressToCluster[a.UID]; ok {
				ca.Cluster = &addressCluster
			}

			tx.InputAddresses = append(tx.InputAddresses, ca)
		}

		for _, a := range t.OutputAddresses {
			ca := AddressWithCluster{UID: a.UID}

			if addressCluster, ok := addressToCluster[a.UID]; ok {
				ca.Cluster = &addressCluster
			}

			tx.OutputAddresses = append(tx.OutputAddresses, ca)
		}

		transactions = append(transactions, tx)
	}

	return
}

// AddCustomClusters adds the given clusters to the database
func AddCustomClusters(ctx context.Context, c external.Database, clusters []CustomCluster) error {
	// validate data
	for _, cluster := range clusters {
		if cluster.Type == "" {
			return serror.FromStr("cluster type is not set")
		}

		if len(cluster.Addresses) == 0 {
			return serror.FromStr("cluster no addresses set")
		}
	}

	pb, err := json.Marshal(clusters)
	if err != nil {
		return serror.New(err)
	}

	_, err = c.Mutate(ctx, &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	})
	if err != nil {
		return serror.New(err)
	}

	return nil
}

// AddClusters adds the given clusters to the database
func AddClusters(c external.Database, clusters []Cluster, checkTx bool) error {
	// validate data
	for _, cluster := range clusters {
		if cluster.Type == "" {
			return serror.FromStr("cluster type is not set")
		}

		if checkTx && cluster.Transaction.UID == "" {
			return serror.FromStr("cluster transaction is not set")
		}

		if len(cluster.Addresses) == 0 && len(cluster.Children) == 0 {
			return serror.FromStr("cluster has no child clusters and no addresses set")
		}
	}

	pb, err := json.Marshal(clusters)
	if err != nil {
		return serror.New(err)
	}

	return db.TxWithRetry(c, time.Minute*5, &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	})
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
	clusters := make([]Cluster, len(operations))
	for i, o := range operations {
		clusters[i] = o.NewCluster
	}

	pb, err := json.Marshal(clusters)
	if err != nil {
		return serror.New(err)
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
		query += "var(func:uid(" + db.CreateCommaList(o.OldClusters) + ")){a" + index + " as Cluster.addresses}\n"
		setNquads += "<" + o.NewCluster.UID + "> <Cluster.addresses> uid(a" + index + ") .\n"

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
		return err
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
		return err
	}

	// step 3: delete all merged clusters
	req = &api.Request{
		Mutations: []*api.Mutation{{
			DelNquads: []byte(delNquads),
		}},
		CommitNow: true,
	}

	return db.ExistingTxWithRetry(txn, time.Minute*5, req)
}

// GetHierarchicalClusterRoot returns the root of the cluster tree clusterUID is part of
func GetHierarchicalClusterRoot(c external.Database, clusterUID string) (rootCluster ClusterWithParent, err error) {
	const query = `query Q($uid:string) {
				var(func: uid($uid))@recurse{
					c as ~Cluster.children
				}
				
				q(func: uid(c))@filter(eq(count(~Cluster.children),0)){
					uid
					Cluster.addressCount
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$uid": clusterUID})
	if err != nil {
		return
	}

	var r struct {
		Root []ClusterWithParent `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Root) != 1 {
		err = serror.FromFormat("invalid number of roots returned: %v", r.Root)
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

	return `q(func: uid(c))@filter(not has(Cluster.user) or uid_in(Cluster.user,$user)){
				uid
				Cluster.type
				Cluster.addressCount
				Cluster.transaction@normalize{
					txhash:txhash
					~transactions{
						bhash:blockhash
						bid:id
						ts:ts
					}
				}
				Cluster.addresses` + limiter + `{
					addresshash
					output_count: count(addr_outputs)
					spent_output_count: count(addr_outputs@filter(has(~tx_inputs)))
				}
			}
			tags(func: uid(c))@filter(not eq(Cluster.type,` + string(TypeHMI) + `)){
				uid
				tags:Cluster.addresses@normalize {
					~Attribution.address@filter(eq(Attribution.isPublic,true) or uid_in(Attribution.user,$user)) {
						tag:Attribution.tag
						isPublic:Attribution.isPublic
					}
				}
			}`
}

// responseToFrontendClusters combines the results of a cluster request to frontend clusters
func responseToFrontendClusters(clusters []FrontendClusterRequest, clusterTags []ClusterTags) (
	frontendClusters []FrontendCluster, err error) {
	tagMap := make(map[string][]Attribution)

	for _, c := range clusterTags {
		tagMap[c.UID] = c.Attributions
	}

	for _, cluster := range clusters {
		if len(cluster.Transaction) > 1 {
			err = serror.FromFormat("invalid transaction count: %d", len(cluster.Transaction))
			return
		}

		frontendCluster := FrontendCluster{
			Type:         cluster.Type,
			AddressCount: cluster.AddressCount,
			Addresses:    cluster.Addresses,
			Attributions: tagMap[cluster.UID],
		}

		// uid is only needed for deleting custom clusters
		if cluster.Type == "custom" {
			frontendCluster.UID = cluster.UID
		}

		// Transaction can be not set if the cluster was created by a user
		if cluster.Transaction != nil {
			frontendCluster.TransactionHash = cluster.Transaction[0].TransactionHash
			frontendCluster.BlockID = cluster.Transaction[0].BlockID
			frontendCluster.BlockHash = cluster.Transaction[0].BlockHash
			frontendCluster.Timestamp = cluster.Transaction[0].Timestamp
		}

		frontendClusters = append(frontendClusters, frontendCluster)
	}

	return
}

// GetClusters returns cluster information for all clusters (except hmi clusters) associated with addressHash
func GetClusters(ctx context.Context, c external.Database, addressHash string,
	maxAddresses int, userID string) ([]FrontendCluster, error) {
	query := `query Q($addressHash:string,$user:string) {
				var(func:eq(addresshash,$addressHash)){
					c as ~Cluster.addresses@filter(not eq(Cluster.type,` + string(TypeHMI) + `))
				}` + getClusterQuery(maxAddresses) + "}"

	resp, err := c.Query(ctx, query, map[string]string{"$addressHash": addressHash, "$user": userID})
	if err != nil {
		return nil, serror.New(err)
	}

	var r struct {
		Clusters    []FrontendClusterRequest `json:"q,omitempty"`
		ClusterTags []ClusterTags            `json:"tags,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, serror.New(err)
	}

	return responseToFrontendClusters(r.Clusters, r.ClusterTags)
}

// GetHMIClusters returns all connected hierarchical multi-input cluster to the
// given address and the uid of the cluster directly connected to the address
func GetHMIClusters(ctx context.Context, c external.Database,
	addressHash string) (addressCluster string, clusters []FrontendHMICluster, err error) {
	const query = string(`query Q($addressHash:string) {
							var(func: eq(addresshash,$addressHash)){
								hmi as ~Cluster.addresses@filter(eq(Cluster.type,` + TypeHMI + `))
							}
							
							var(func: uid(hmi))@recurse{
								s as Cluster.children
								v as ~Cluster.children
							}

							x(func: uid(hmi)){
								uid
							}
							
							q(func: uid(s,v)){
								uid
								Cluster.addressCount
								Cluster.transaction{
									txhash
								}
								Cluster.children{
									uid
								}
								~Cluster.children{
									uid
								}
							}
						  }`)

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$addressHash": addressHash})
	if err != nil {
		err = serror.New(err)
		return
	}

	var r struct {
		Clusters []struct {
			UID          string `json:"uid,omitempty"`
			AddressCount int    `json:"Cluster.addressCount,omitempty"`
			Transaction  struct {
				TxHash string `json:"txhash,omitempty"`
			} `json:"Cluster.transaction,omitempty"`
			Children []SubCluster `json:"Cluster.children,omitempty"`
			Parent   []SubCluster `json:"~Cluster.children,omitempty"`
		} `json:"q,omitempty"`
		AddressCluster []struct {
			UID string `json:"uid,omitempty"`
		} `json:"x,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.AddressCluster) == 0 {
		// no clusters found
		return
	}

	if len(r.AddressCluster) > 1 {
		err = serror.FromStr("too many clusters associated with address")
		return
	}

	addressCluster = r.AddressCluster[0].UID

	for _, cluster := range r.Clusters {
		if len(cluster.Parent) > 1 {
			err = serror.FromFormat("cluster %s has multiple parents: %v", cluster.UID, cluster.Parent)
			return
		}

		var parentUID string
		if len(cluster.Parent) == 1 {
			parentUID = cluster.Parent[0].UID
		}

		var childClusters []string
		for _, child := range cluster.Children {
			childClusters = append(childClusters, child.UID)
		}

		clusters = append(clusters, FrontendHMICluster{
			UID:             cluster.UID,
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
					c as ~Cluster.user
				}

				q(func: uid(c)){
					uid
					Cluster.ts
					Cluster.addressCount
					Cluster.addresses(first:10){
						addresshash
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$user": userID})
	if err != nil {
		return
	}

	var r struct {
		Clusters []struct {
			UID          string `json:"uid,omitempty"`
			Timestamp    string `json:"Cluster.ts,omitempty"`
			AddressCount int64  `json:"Cluster.addressCount,omitempty"`
			Addresses    []struct {
				Hash string `json:"addresshash,omitempty"`
			} `json:"Cluster.addresses,omitempty"`
		} `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	for _, cluster := range r.Clusters {
		var addresses []string
		for _, a := range cluster.Addresses {
			addresses = append(addresses, a.Hash)
		}
		clusters = append(clusters, FrontendUserCluster{
			UID:          cluster.UID,
			Timestamp:    cluster.Timestamp,
			AddressCount: cluster.AddressCount,
			Addresses:    addresses,
		})
	}

	return
}

// GetUserClustersUIDs returns all UIDs of clusters of a user
func GetUserClustersUIDs(c external.Database, userID string, clusterTypeFilter []ClusterType) (clusters []string, err error) {
	var filter string
	if len(clusterTypeFilter) > 0 {
		for i, ct := range clusterTypeFilter {
			filter += string(ct)

			if i+1 < len(clusterTypeFilter) {
				filter += ","
			}
		}

		filter = "@filter(eq(Cluster.type," + filter + "))"
	}

	query := fmt.Sprintf(`query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					c as ~Cluster.user%s
				}

				q(func: uid(c)){
					uid
				}
			  }`, filter)

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$user": userID})
	if err != nil {
		return
	}

	var r struct {
		Clusters []struct {
			UID string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	for _, cluster := range r.Clusters {
		clusters = append(clusters, cluster.UID)
	}

	return
}

// DeleteCluster deletes the given cluster
func DeleteCluster(c external.Database, userID string, clusterUID string) error {
	req := &api.Request{
		Query: `query Q($user:string,$cluster:string) {
				var(func:uid($user))@filter(type(User)){
					c as ~Cluster.user@filter(uid($cluster))
				}
			  }`,
		Vars: map[string]string{"$user": userID, "$cluster": clusterUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(c) * * ."),
		}},
		CommitNow: true,
	}
	resp, err := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if err != nil {
		return err
	}

	if resp.GetMetrics().NumUids["mutation_cost"] == 0 {
		return serror.FromStr("nothing was deleted")
	}

	return nil
}

// DeleteAllClusters deletes all clusters of a given user
func DeleteAllClusters(ctx context.Context, c external.Database, userID string) error {
	req := &api.Request{
		Query: `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					c as ~Cluster.user
				}
			  }`,
		Vars: map[string]string{"$user": userID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(c) * * ."),
		}},
		CommitNow: true,
	}

	return db.MutationWithRetry(ctx, c, req)
}

// GetRelatedClusters returns the UIDs of clusters which can be reached from the given cluster
func GetRelatedClusters(c external.Database, clusterUID string, userUID string, clusterTypeFilter []ClusterType) (clusters []string, err error) {
	if clusterTypeFilter == nil {
		err = serror.FromStr("no cluster types passed to function")
		return
	}

	var filter string
	for i, ct := range clusterTypeFilter {
		filter += string(ct)

		if i+1 < len(clusterTypeFilter) {
			filter += ","
		}
	}

	query := fmt.Sprintf(`query Q($user:string,$cluster:string) {
					var(func: uid($cluster))@recurse{
    					Cluster.addresses
    					c as ~Cluster.addresses@filter(eq(Cluster.type, `+string(TypeFMI)+`) or (eq(Cluster.type,%s) and uid_in(Cluster.user,$user)))
  					}

  					q(func: uid(c)) {
    					uid
					}
				  }`, filter)

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$user": userUID, "$cluster": clusterUID})
	if err != nil {
		return
	}

	var r struct {
		Clusters []struct {
			UID string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	for _, cluster := range r.Clusters {
		clusters = append(clusters, cluster.UID)
	}

	return
}

// GetClusterAddressCount returns the number of addresses the cluster of the given address is connected to
func GetClusterAddressCount(c external.Database, addressHash string) (addressCount int64, err error) {
	const query = `query Q($addressHash:string){
				q(func:eq(addresshash,$addressHash))@normalize{
					~Cluster.addresses@filter(eq(Cluster.type,` + string(TypeFMI) + `)){
						count:Cluster.addressCount
					}
				}
			   }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$addressHash": addressHash})
	if err != nil {
		return
	}

	var r struct {
		Count []struct {
			Count int64 `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Count) == 0 {
		return 0, nil
	}

	if len(r.Count) != 1 {
		return 0, serror.FromStr("invalid response from database")
	}

	addressCount = r.Count[0].Count

	return
}

// DeleteAllFMIClusters deletes all flat multi-input clusters of a given user
func DeleteAllFMIClusters(c external.Database) error {
	const query = `{
				 q(func: eq(Cluster.type, "fmi")){
					count(uid)
				  }
				}`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Second*20, query)
	if err != nil {
		return err
	}

	var r struct {
		Count []struct {
			Count int64 `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return err
	}

	deletionRounds := r.Count[0].Count / 10000
	for range deletionRounds + 1 {
		err := deleteTenThousandFMIClusters(c)
		if err != nil {
			return err
		}
	}

	return err
}

func deleteTenThousandFMIClusters(c external.Database) error {
	req := &api.Request{
		Query: `{
				var(func:eq(Cluster.type, "fmi"), first: 10000){
					c as uid
				}
			  }`,
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(c) * * ."),
		}},
		CommitNow: true,
	}
	_, err := db.TxWithRetryAndResponse(c, time.Minute*15, req)

	return err
}

// GetClustersByBlockRange returns all cluster-address mappings of the given block range
func GetClustersByBlockRange(c external.Database, blockHeightStart int, blockHeightEnd int,
	convertUIDs bool) (clusters []Cluster, err error) {
	const query = `query Q($start: string,$end: string) {
				var(func: between(id,$start,$end)) {
					transactions {
						o as tx_outputs
						i as tx_inputs
					}
				}
				
				var(func: uid(o,i)){
					a as ~addr_outputs{
						c as ~Cluster.addresses
					}
				}
				
				q(func: uid(c)){
					uid 
					Cluster.type
					Cluster.addressCount
					Cluster.transaction {
						uid
					}
					dgraph.type
					Cluster.addresses@filter(uid(a)){
						uid
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*10, query,
		map[string]string{"$start": strconv.Itoa(blockHeightStart), "$end": strconv.Itoa(blockHeightEnd)})
	if err != nil {
		return
	}

	var r struct {
		Clusters []Cluster `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if convertUIDs {
		for i := range r.Clusters {
			r.Clusters[i].UID = "_:" + r.Clusters[i].UID

			for y := range r.Clusters[i].Addresses {
				r.Clusters[i].Addresses[y].UID = "_:" + r.Clusters[i].Addresses[y].UID
			}

			r.Clusters[i].Transaction.UID = "_:" + r.Clusters[i].Transaction.UID
		}
	}

	clusters = r.Clusters

	return
}
