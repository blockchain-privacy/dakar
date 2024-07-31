package heuristics

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	"backend/external"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"sort"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v230/protos/api"
)

var (
	errInvalidDatabaseResponse = errors.New("error invalid response")
	// ErrNoMutationHappened is returned if no mutation occurred
	ErrNoMutationHappened = errors.New("no mutation happened")
)

// InsertHeuristic inserts the given heuristic
func InsertHeuristic(c external.Database, h *Heuristic, userUID string, workspaceUID string) (insertUID string, err error) {
	h.SetDType()

	const newHeuristicDummyUID = "new_h"
	h.UID = "_:" + newHeuristicDummyUID

	var query string

	// if TxHash is not empty we have to search for the transaction uid
	if h.TxHash != "" {
		h.Transaction.UID = "uid(tx)"
		query = `query Q($txhash: string, $userUID: string, $workspaceUID: string) {
					tx as var(func: eq(txhash, $txhash))
					var(func: uid($userUID))@filter(type(User)){
						w as User.workspaces@filter(uid($workspaceUID))
					}
				  }`
	}

	type dummyWorkspace struct {
		UID        string      `json:"uid,omitempty"`
		Heuristics []Heuristic `json:"Workspace.heuristics,omitempty"`
	}

	// set cluster height to 0, to force an update of the corresponding workspace
	pb, err := json.Marshal(dummyWorkspace{UID: workspaceUID, Heuristics: []Heuristic{*h}})
	if err != nil {
		err = serror.New(err)
		return
	}

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$txhash": h.TxHash, "$userUID": userUID, "$workspaceUID": workspaceUID},
		Mutations: []*api.Mutation{{
			Cond:    "@if(gt(len(w), 0))",
			SetJson: pb,
		}},
		CommitNow: true,
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if err != nil {
		return
	}

	insertUID, ok := resp.GetUids()[newHeuristicDummyUID]
	if !ok {
		err = serror.FromStr("no new heuristic created")
		return
	}

	return
}

// DeleteUserHeuristics deletes all given heuristic uids of a user
func DeleteUserHeuristics(c external.Database, uids []string, userUID string, workspaceUID string) error {
	const query = `
		query Q($userUID:string,$heuristicUIDs:string,$workspaceUID:string){
			var(func: uid($userUID)){
				User.workspaces@filter(uid($workspaceUID)){
					h as Workspace.heuristics@filter(uid($heuristicUIDs)){
						hc as Heuristic.clusters{
							hr as HeuristicCluster.results
						}
					}
				}
			}
		}`

	req := &api.Request{
		Query: query,
		Vars: map[string]string{"$userUID": userUID,
			"$heuristicUIDs": db.CreateCommaArray(uids), "$workspaceUID": workspaceUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte(` uid(hr) * * .
								uid(hc) * * .
								uid(h) * * .
								<` + workspaceUID + "> <Workspace.heuristics> uid(h) ."),
		}},
		CommitNow: true,
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if err != nil {
		return err
	}

	if v, ok := resp.Metrics.NumUids["mutation_cost"]; !ok || v == 0 {
		return serror.New(ErrNoMutationHappened)
	}

	return nil
}

// GetHeuristicResults returns the connected transactions of heuristic
func GetHeuristicResults(c external.Database, heuristicUID string) (results []HeuristicTransaction,
	attributionMap map[ClusterUID][]string, err error) {
	const query = `query Q($uid:string) {
				var (func: uid($uid)){ x as Heuristic.clusters }
				
				q(func: uid(x)){
					uid
					HeuristicCluster.results{
						HeuristicResult.origin{
							uid
							tx_outputs{
								amount
							}
						}
					}
					HeuristicCluster.attributions{
						uid
					}
			  	}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$uid": heuristicUID})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Clusters []struct {
			UID     ClusterUID `json:"uid,omitempty"`
			Results []struct {
				Origin struct {
					UID     string            `json:"uid,omitempty"`
					Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
				} `json:"HeuristicResult.origin,omitempty"`
			} `json:"HeuristicCluster.results,omitempty"`
			Attributions []db.UIDNode `json:"HeuristicCluster.attributions,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	attributionMap = make(map[ClusterUID][]string)
	var clusterCounter int64
	for _, cluster := range r.Clusters {
		thisClusterID := ClusterUID(strconv.FormatInt(clusterCounter, 10))
		for _, result := range cluster.Results {
			results = append(results, HeuristicTransaction{
				UID:     result.Origin.UID,
				Cluster: thisClusterID,
				Outputs: result.Origin.Outputs,
			})
		}

		for _, attr := range cluster.Attributions {
			attributionMap[thisClusterID] = append(attributionMap[thisClusterID], attr.UID)
		}

		clusterCounter++
	}

	return
}

// GetInputTransactions returns the input mixing transactions of the given transaction.
func GetInputTransactions(c external.Database, tx string) (inputTransactions []HeuristicTransaction, err error) {
	query := `query Q($txhash: string){
				var (func: eq(txhash,$txhash)){
					tx_inputs{
						v as ~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast + `))
					}
				}
				
				q(func: uid(v)){
					uid
					tx_outputs@normalize{
						amount:amount
						~tx_inputs{
							input_tx:txhash
						}
					}
					~transactions{
						ts
					}
				}
				}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$txhash": tx})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			UID     string            `json:"uid,omitempty"`
			Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
			Inputs  []struct {
				Address string `json:"addr_uid,omitempty"`
				Cluster string `json:"cluster_uid,omitempty"`
			} `json:"tx_inputs,omitempty"`
			Block []struct {
				Timestamp time.Time `json:"ts,omitempty"`
			} `json:"~transactions,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Transaction) == 0 {
		err = serror.New(errInvalidDatabaseResponse)
		return
	}

	for _, t := range r.Transaction {
		if len(t.Block) != 1 || len(t.Outputs) == 0 {
			err = serror.New(errInvalidDatabaseResponse)
			return
		}
		inputTransactions = append(inputTransactions, HeuristicTransaction{
			UID:       t.UID,
			Timestamp: t.Block[0].Timestamp,
			Outputs:   t.Outputs,
		})
	}

	return
}

// GetTransactionsWithOutputAmountAndCluster returns a slice of transactions and used attributions per cluster.
// Each transaction contains its output amounts and the cluster of all inputs.
// If no attributions were used or found the returned map is nil.
func GetTransactionsWithOutputAmountAndCluster(c external.Database, uids []string, userUID string,
	requestedClusterTypes []clustering.ClusterType) (origins []HeuristicTransaction,
	attributionMapping map[ClusterUID][]string, err error) {
	isSimpleClustering := len(requestedClusterTypes) == 0 // true -> only multi-input clusters will be used

	// get user clusters if necessary
	var userClusterUIDs []string
	if !isSimpleClustering {
		userClusterUIDs, err = clustering.GetUserClustersUIDs(c, userUID, requestedClusterTypes)
		if err != nil {
			return
		}

		// if the user does not have defined any custom clusters,
		// then the request can be treated like multi-input only
		if userClusterUIDs == nil {
			isSimpleClustering = true
		}
	}

	/// build query

	var usedClusterTypes string
	if !isSimpleClustering {
		for i, ct := range requestedClusterTypes {
			usedClusterTypes += string(ct)

			if i+1 < len(requestedClusterTypes) {
				usedClusterTypes += ","
			}
		}
		usedClusterTypes = " or (eq(Cluster.type, " + usedClusterTypes + ")  and uid_in(Cluster.user," + userUID + "))"
	}

	query := fmt.Sprintf(`query Q($uids:string){
				q(func: uid($uids)){
					uid
					tx_outputs{
						amount
					}
					tx_inputs(first:1){
						~addr_outputs{
							uid
							~Cluster.addresses@filter(eq(Cluster.type,%s)%s){
								uid
							}
						}
					}
			   	}
			  }`, string(clustering.TypeFMI), usedClusterTypes)

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$uids": db.CreateCommaArray(uids)})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Origins []struct {
			UID     string            `json:"uid,omitempty"`
			Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
			Inputs  []struct {
				Address []struct {
					UID     string       `json:"uid,omitempty"`
					Cluster []db.UIDNode `json:"~Cluster.addresses,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_inputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	var superClusters []mergedClusterItem
	allClusters := make(map[string]bool)

	if !isSimpleClustering {
		// get all merged clusters
		for _, userCluster := range userClusterUIDs {
			// check if userCluster has already been found in a previous iteration
			if allClusters[userCluster] {
				continue
			}

			mergedClusterUIDS, relatedErr := clustering.GetRelatedClusters(c, userCluster, userUID, requestedClusterTypes)
			if relatedErr != nil {
				err = relatedErr
				return
			}

			mergedCluster := make(map[string]bool)
			for _, mcu := range mergedClusterUIDS {
				mergedCluster[mcu] = true
				allClusters[mcu] = true
			}

			superClusters = append(superClusters, mergedClusterItem{clusterUIDs: mergedCluster})
		}
	}

	attributions, attributionErr := attribution.GetAttributionsPerCluster(c, userUID, requestedClusterTypes)
	if attributionErr != nil {
		err = attributionErr
		return
	}

	type usedCluster struct{ superCluster map[string]bool }

	// allUsedClusters holds all cluster IDs which are used by the generated HeuristicTransactions below
	allUsedClusters := make(map[string]usedCluster)

	for _, o := range r.Origins {
		if o.Inputs == nil || o.Inputs[0].Address == nil {
			err = serror.FromFormat("invalid cluster information for transaction %s", o.UID)
			return
		}

		var cUID ClusterUID

		// If the cluster is not set, we use the address UID as the "cluster".
		// This can happen if the address has not been assigned to a cluster yet.
		if o.Inputs[0].Address[0].Cluster == nil {
			id := o.Inputs[0].Address[0].UID
			cUID = ClusterUID(id)
			allUsedClusters[id] = usedCluster{superCluster: nil}
		} else if firstClusterUID := o.Inputs[0].Address[0].Cluster[0].UID; isSimpleClustering {
			// must have a multi-input cluster UID
			cUID = ClusterUID(firstClusterUID)
			allUsedClusters[firstClusterUID] = usedCluster{superCluster: nil}
		} else {
			// If this cluster's UID does not appear in the merged user clusters,
			// then use the multi-input cluster UID.
			if !allClusters[firstClusterUID] {
				cUID = ClusterUID(firstClusterUID)
				allUsedClusters[firstClusterUID] = usedCluster{superCluster: nil}
			} else {
				var superCluster map[string]bool
				cUID, superCluster, err = getClusterUIDFromMergedClusters(superClusters, firstClusterUID)
				if err != nil {
					return
				}
				allUsedClusters[string(cUID)] = usedCluster{superCluster: superCluster}
			}
		}

		origins = append(origins, HeuristicTransaction{
			UID:     o.UID,
			Cluster: cUID,
			Outputs: o.Outputs,
		})
	}

	if attributions == nil {
		return
	}

	attributionMapping = make(map[ClusterUID][]string)
	for clusterID, v := range allUsedClusters {
		// no super clusters, so either a simple address or a multi-input cluster
		if v.superCluster == nil {
			if attr, ok := attributions[clusterID]; ok {
				attributionMapping[ClusterUID(clusterID)] = attr
			}
		} else {
			for cluster := range v.superCluster {
				if attr, ok := attributions[cluster]; ok {
					attributionMapping[ClusterUID(clusterID)] = attr
				}
			}
		}
	}

	if len(attributionMapping) == 0 {
		attributionMapping = nil
	}

	return
}

// getClusterUIDFromMergedClusters searches for clusterUID in mergedClusters
// and returns a hash of the merged clusters if found. In case the uid is not found, an error is returned.
func getClusterUIDFromMergedClusters(mergedClusters []mergedClusterItem,
	clusterUID string) (ClusterUID, map[string]bool, error) {
	for i, mc := range mergedClusters {
		if _, ok := mc.clusterUIDs[clusterUID]; ok {
			// lazy creation of map hashes
			if mc.clusterHash == "" {
				mc.clusterHash = createKeyHash(mc.clusterUIDs)
				mergedClusters[i] = mc
			}

			return ClusterUID(mc.clusterHash), mc.clusterUIDs, nil
		}
	}

	return "", nil, serror.FromStr("did not find cluster uid in merged cluster list")
}

// createKeyHash creates from the keys of the map a unique string.
// Maps with the same keys but in different order create the same output.
func createKeyHash(someMap map[string]bool) string {
	// catch both nil Maps and empty maps
	if len(someMap) == 0 {
		return ""
	}

	// sort elements so a consistent hash can be generated
	keys := cliutil.GetMapKeys(someMap)
	sort.Strings(keys)

	var allKeys []byte
	for _, k := range keys {
		allKeys = append(allKeys, []byte(k)...)
	}

	sha256Hash := sha256.New()
	sha256Hash.Write(allKeys)
	return base64.URLEncoding.EncodeToString(sha256Hash.Sum(nil))
}

// GetTransactionsWithInputAmount returns a slice of transactions. Each transaction contains its input amounts.
func GetTransactionsWithInputAmount(c external.Database, uids []string) (origins []HeuristicTransaction, err error) {
	query := `query Q($uids:string){
				q(func: uid($uids)){
					uid
					tx_inputs{
						amount
					}
				}
			   }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$uids": db.CreateCommaArray(uids)})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Origins []struct {
			UID     string            `json:"uid,omitempty"`
			Outputs []HeuristicOutput `json:"tx_inputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	for _, o := range r.Origins {
		origins = append(origins, HeuristicTransaction{
			UID:     o.UID,
			Outputs: o.Outputs,
		})
	}

	return
}

// GetInputAmounts gets the amounts of the inputs
func GetInputAmounts(c external.Database, tx string) (transaction HeuristicTransaction, err error) {
	query := `query Q($txhash: string){
				q(func: eq(txhash,$txhash)){
					uid
					tx_inputs{
						amount
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$txhash": tx})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			UID     string            `json:"uid,omitempty"`
			Outputs []HeuristicOutput `json:"tx_inputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Transaction) != 1 {
		err = serror.New(errInvalidDatabaseResponse)
		return
	}

	t := r.Transaction[0]
	transaction = HeuristicTransaction{
		UID:     t.UID,
		Outputs: t.Outputs,
	}

	return
}

// GetFrontendHeuristicByUID returns the heuristic for the given heuristicUID, which was created by userUID
func GetFrontendHeuristicByUID(ctx context.Context, c external.Database,
	heuristicUID string, userUID string, workspaceUID string) (frontendHeuristic FrontendHeuristicShort, err error) {
	const query = `query Q($heuristicUID:string,$userUID:string,$workspaceUID:string){
				var(func: uid($userUID)){
					User.workspaces@filter(uid($workspaceUID)){
						Workspace.heuristics@filter(uid($heuristicUID)){
							c as Heuristic.clusters
						}
					}
				}

				q(func: uid(c)){
					HeuristicCluster.results{
						HeuristicResult.origin@normalize{
							txhash:txhash
							~transactions{
								ts:ts
							}
						}
						HeuristicResult.destinations{
							uid
						}
					}
					HeuristicCluster.attributions {
						tag:Attribution.tag
						isPublic:Attribution.isPublic
					}
				}
			   }`

	resp, err := c.Query(ctx, query, map[string]string{"$heuristicUID": heuristicUID,
		"$userUID": userUID, "$workspaceUID": workspaceUID})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	var r struct {
		Clusters []struct {
			Results []struct {
				// Origin must be declared as an array because in the query @normalize is used
				Origin       []FrontendTransactionResult `json:"HeuristicResult.origin,omitempty"`
				Destinations []db.UIDNode                `json:"HeuristicResult.destinations,omitempty"`
			} `json:"HeuristicCluster.results,omitempty"`
			Attributions []Attribution `json:"HeuristicCluster.attributions,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	for _, cluster := range r.Clusters {
		var origins []FrontendTransactionResult
		destinationMap := make(map[string]bool)
		for _, result := range cluster.Results {
			if len(result.Origin) != 1 {
				err = serror.FromStr("invalid response from database")
				return
			}

			result.Origin[0].DestinationCount = len(result.Destinations)

			origins = append(origins, result.Origin[0])

			// collect destinations of all results in map
			for _, destination := range result.Destinations {
				destinationMap[destination.UID] = true
			}
		}

		frontendHeuristic.Clusters = append(frontendHeuristic.Clusters, FrontendHeuristicCluster{
			Transactions: origins,
			Attributions: cluster.Attributions,
		})
	}

	frontendHeuristic.UID = heuristicUID

	return
}
