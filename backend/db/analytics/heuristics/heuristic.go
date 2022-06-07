package heuristics

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/db/analytics/clustering"
	dbtx "backend/db/transaction"
	"backend/external"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

var (
	errInvalidDatabaseResponse = errors.New("error invalid response")
	// ErrNoMutationHappened is returned if no mutation occurred
	ErrNoMutationHappened = errors.New("no mutation happened")
)

// InsertHeuristic inserts the given heuristic
func InsertHeuristic(c external.Database, h Heuristic, userUID string) (insertUID string, err error) {
	h.SetDType()
	h.Timestamp = time.Now().UTC().Format(time.RFC3339)

	const newHeuristicDummyUID = "new_h"
	h.UID = "_:" + newHeuristicDummyUID

	var query string

	// if TxHash is not empty we have to search for the transaction uid
	if h.TxHash != "" {
		h.Transaction.UID = "uid(tx)"
		query = `query Q($txhash: string) {
					tx as var(func: eq(txhash, $txhash))
				  }`
	}

	type dummyUser struct {
		UID        string      `json:"uid,omitempty"`
		Heuristics []Heuristic `json:"User.heuristics,omitempty"`
	}

	pb, err := json.Marshal(dummyUser{UID: userUID, Heuristics: []Heuristic{h}})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$txhash": h.TxHash},
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	insertUID, ok := resp.GetUids()[newHeuristicDummyUID]
	if !ok {
		err = errors.New(fmt.Sprintln("no new heuristic created"))
		return
	}

	return
}

// DeleteUserHeuristics deletes all given heuristic uids of a user
func DeleteUserHeuristics(c external.Database, uids []string, userUID string) (err error) {
	uidList := db.CreateCommaArray(uids)

	const query = `query Q($user:string, $uids:string){
				h as var(func: uid($uids))@filter(uid_in(~User.heuristics,$user) AND eq(dgraph.type,` + DType + `)){
					hc as Heuristic.clusters{
							hr as HeuristicCluster.results
					}
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$user": userUID, "$uids": uidList},
		Mutations: []*api.Mutation{{
			DelNquads: []byte(` uid(hr) * * .
								uid(hc) * * .
								uid(h) * * .
								<` + userUID + "> <User.heuristics> uid(h) ."),
		}},
		CommitNow: true,
	}

	if txErr := db.TxWithRetry(c, time.Minute*5, req); txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}
	return
}

// DeleteAllUserHeuristics deletes all heuristics of a user
func DeleteAllUserHeuristics(c external.Database, userUID string) (err error) {
	req := &api.Request{
		Query: `query Q($user:string){
				var(func: uid($user)){
					h as User.heuristics{
						hc as Heuristic.clusters{
							hr as HeuristicCluster.results
						}
					}
				}
}`,
		Vars: map[string]string{"$user": userUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte(` uid(hr) * * .
								uid(hc) * * .
								uid(h) * * .
								<` + userUID + "> <User.heuristics> uid(h) ."),
		}},
		CommitNow: true,
	}

	resp, txErr := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	if v, ok := resp.Metrics.NumUids["mutation_cost"]; !ok || v == 0 {
		return ErrNoMutationHappened
	}

	return
}

// DeleteAllUserTxHeuristics deletes all heuristics of a user of a particular transaction
func DeleteAllUserTxHeuristics(c external.Database, txhash string, userUID string) (err error) {
	query := `query Q($user:string, $hash:string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				# get all heuristic of that user and transaction
				var(func: uid($user)){
					h as User.heuristics@filter(uid_in(Heuristic.transaction, uid(tx))){
						hc as Heuristic.clusters{
							hr as HeuristicCluster.results
						}
					}
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$user": userUID, "$hash": txhash},
		Mutations: []*api.Mutation{{
			DelNquads: []byte(` uid(hr) * * .
								uid(hc) * * .
								uid(h) * * .
								<` + userUID + "> <User.heuristics> uid(h) ."),
		}},
		CommitNow: true,
	}

	resp, txErr := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	if v, ok := resp.Metrics.NumUids["mutation_cost"]; !ok || v == 0 {
		return ErrNoMutationHappened
	}

	return
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Clusters []queryHeuristicClusters `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []queryHeuristicTransaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidDatabaseResponse)
		return
	}

	for _, t := range r.Transaction {
		if len(t.Block) != 1 || len(t.Outputs) == 0 {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidDatabaseResponse)
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
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Origins []struct {
			UID     string            `json:"uid,omitempty"`
			Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
			Inputs  []struct {
				Address []struct {
					UID     string `json:"uid,omitempty"`
					Cluster []struct {
						UID string `json:"uid,omitempty"`
					} `json:"~Cluster.addresses,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_inputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), relatedErr)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), attributionErr)
		return
	}

	type usedCluster struct{ superCluster map[string]bool }

	// allUsedClusters holds all cluster IDs which are used by the generated HeuristicTransactions below
	allUsedClusters := make(map[string]usedCluster)

	for _, o := range r.Origins {
		if o.Inputs == nil || o.Inputs[0].Address == nil {
			err = fmt.Errorf("invalid cluster information for transaction %s", o.UID)
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
					err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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

	return "", nil, errors.New("did not find cluster uid in merged cluster list")
}

// createKeyHash creates from the keys of the map a unique string.
// Maps with the same keys but in different order create the same output.
func createKeyHash(someMap map[string]bool) string {
	// catch both nil Maps and empty maps
	if len(someMap) == 0 {
		return ""
	}

	// sort elements so a consistent hash can be generated
	keys := make([]string, 0, len(someMap))
	for k := range someMap {
		keys = append(keys, k)
	}

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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Origins []queryHeuristicTransactionInputs `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidDatabaseResponse)
		return
	}

	t := r.Transaction[0]
	transaction = HeuristicTransaction{
		UID:     t.UID,
		Outputs: t.Outputs,
	}

	return
}

// DoesHeuristicUIDExist checks if the given heuristic uids exist. All heuristics must belong to the same transaction
func DoesHeuristicUIDExist(c external.Database, txhash string, uids []string) (allExist bool, err error) {
	uidList := db.CreateCommaArray(uids)

	const query = `query Q($hash:string, $uids:string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				# filter and count
				q(func: uid($uids))@filter(uid_in(Heuristic.transaction, uid(tx)) AND eq(dgraph.type,` + DType + `)){
					count(uid)
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query,
		map[string]string{"$hash": txhash, "$uids": uidList})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Count []struct {
			Number int `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Count) == 0 || len(r.Count) > 1 {
		err = errors.New("error invalid response from database")
		return
	} else if r.Count[0].Number != len(uids) {
		err = errors.New("error received number of uids does not match")
		return
	}

	allExist = true
	return
}

// GetBasicFrontendHeuristic returns all heuristics for a given transaction created by userUid. Basic information only.
func GetBasicFrontendHeuristic(c external.Database, txHash string, userUID string) (
	heuristics []FrontendHeuristic, err error) {
	const query = `query Q($hash:string, $user:string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				var(func: uid($user)){
					h as User.heuristics@filter(uid_in(Heuristic.transaction, uid(tx)))
				}
				
				q(func: uid(h)){
					uid
					ts:Heuristic.ts
					type:Heuristic.type
					parameter:Heuristic.parameter
					clusterTypes:Heuristic.clusterTypes
					excludeAddresses:Heuristic.excludeAddresses
					parent:Heuristic.parent{
						uid
					}
					children:~Heuristic.parent{
						uid
					}
					clusterCount: count(Heuristic.clusters)
				}
			  }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash, "$user": userUID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Heuristics []FrontendHeuristic `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	heuristics = r.Heuristics

	return
}

// GetFrontendHeuristicByUID returns the heuristic for the given heuristicUID, which was created by userUID
func GetFrontendHeuristicByUID(c external.Database, heuristicUID string, userUID string) (
	frontendHeuristic FrontendHeuristicShort, err error) {
	const query = `query Q($uid:string,$user:string){
				var(func:uid($uid))@cascade{
					~User.heuristics@filter(uid($user))
					c as Heuristic.clusters
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

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$uid": heuristicUID, "$user": userUID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Clusters []struct {
			Results []struct {
				// Origin must be declared as an array because in the query @normalize is used
				Origin       []FrontendTransactionResult `json:"HeuristicResult.origin,omitempty"`
				Destinations []struct {
					UID string `json:"uid,omitempty"`
				} `json:"HeuristicResult.destinations,omitempty"`
			} `json:"HeuristicCluster.results,omitempty"`
			Attributions []Attribution `json:"HeuristicCluster.attributions,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if heuristicUID == "" {
		err = errors.New("empty response from database")
		return
	}

	frontendHeuristic.UID = heuristicUID

	for _, cluster := range r.Clusters {
		var origins []FrontendTransactionResult
		destinationMap := make(map[string]bool)
		for _, result := range cluster.Results {
			if len(result.Origin) != 1 {
				err = errors.New("invalid response from database")
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

	return
}

// GetShortestTransactionPathAnyDirection returns the transactions of the shortest path between two transactions.
// anyDirection determines the search direction of the shortest transaction path query
// True: Both inputs and outputs are traversed
// False: Only inputs are traversed
// withPrivacyTransactions determines if privacy transactions should be considered when doing the shortest path lookup
func GetShortestTransactionPathAnyDirection(c external.Database, txFrom string, txTo string,
	withPrivacyTransactions bool, anyDirection bool) (txs []dbtx.FrontendTransaction, err error) {
	/* Full query
	query Q($txFrom:string, $txTo:string){
					f as var(func: eq(txhash,$txFrom))
					t as var(func: eq(txhash,$txTo))
					path as shortest(from: uid(f), to: uid(t)){
						tx_inputs
						~tx_outputs@filter(NOT has(privacytype)) tx_outputs ~tx_inputs@filter(NOT has(privacytype)) }
					path(func: uid(path))@normalize{
						txhash:txhash
						privacytype:privacytype
						~transactions{
							bid:id
							bts:ts
							bhash:blockhash
						}
					}
				  }
	*/

	privacyFlag := " " // spaces are needed

	if !withPrivacyTransactions {
		privacyFlag = "@filter(NOT has(privacytype)) " // spaces are needed
	}

	var anyDirectionFlag string

	if anyDirection {
		anyDirectionFlag = "tx_outputs ~tx_inputs" + privacyFlag
	}

	query := `query Q($txFrom:string, $txTo:string){
				f as var(func: eq(txhash,$txFrom))
				t as var(func: eq(txhash,$txTo))
				path as shortest(from: uid(f), to: uid(t)){
					tx_inputs
					~tx_outputs` + privacyFlag + anyDirectionFlag + `}
				path(func: uid(path))@normalize{
					txhash:txhash
					privacytype:privacytype
					~transactions{
						bid:id
						bts:ts
						bhash:blockhash
					}
				}
			  }`

	// without retry, as this request can easily time out
	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$txFrom": txFrom, "$txTo": txTo})
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
		err = nil
		return
	}

	// json struct
	var r struct {
		Transactions []dbtx.FrontendTransaction `json:"path,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	txs = r.Transactions

	return
}

// GetHeuristicListByUser returns all transactions for which the given user has created heuristics
func GetHeuristicListByUser(c external.Database, userUID string) (frontendHeuristic []HeuristicListItem, err error) {
	query := `query Q($uuid:string){
				# get transaction
				var(func: uid($uuid)){
					User.heuristics{
						tx as Heuristic.transaction
					}
				}
				# get count
				var(func: uid(tx)){
					c as count(~Heuristic.transaction)@filter(uid_in(~User.heuristics,$uuid))
				}
				# get time
				var(func: uid(tx)){
					~Heuristic.transaction@filter(uid_in(~User.heuristics,$uuid)){
						t as Heuristic.ts
					}
					max_time as  max(val(t))
				}
				# output
				q(func: uid(tx)){
					txhash
					h_count: val(c)
					mod_time: val(max_time)
				}
			   }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$uuid": userUID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Items []HeuristicListItem `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	frontendHeuristic = r.Items

	return
}
