package clustering

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/external"
	"errors"
	"github.com/dgraph-io/dgo/v210/protos/api"

	"encoding/json"
	"fmt"
	"strconv"
	"time"
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

// AddClusters adds the given clusters to the database
func AddClusters(c external.Database, clusters []Cluster) (map[string]string, error) {
	// validate data
	for _, cluster := range clusters {
		if cluster.Type == "" {
			return nil, errors.New("cluster type is not set")
		}

		if cluster.Transaction.Uid == "" {
			return nil, errors.New("cluster transaction is not set")
		}

		if len(cluster.Addresses) == 0 && len(cluster.Children) == 0 {
			return nil, errors.New("cluster has no child clusters and no addresses set")
		}
	}

	pb, err := json.Marshal(clusters)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return nil, err
	}

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}
	resp, err := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return resp.GetUids(), err
}

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
