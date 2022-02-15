package heuristics

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/db/analytics/clustering"
	dbtx "backend/db/transaction"
	"backend/external"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

	const query = `query Q($uuid:string, $uids:string, $type:string){
				h as var(func: uid($uids))@filter(uid_in(~User.heuristics,$uuid) AND eq(dgraph.type,$type)){
					hr as Heuristic.results
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uuid": userUID, "$uids": uidList, "$type": DType},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(hr) * * .\nuid(h) * * .\n<" + userUID + "> <User.heuristics> uid(h) ."),
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
		Query: "query Q($uuid:string){var(func: uid($uuid)){h as User.heuristics{hr as Heuristic.results}}}",
		Vars:  map[string]string{"$uuid": userUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(hr) * * .\nuid(h) * * .\n<" + userUID + "> <User.heuristics> uid(h) ."),
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
	query := `query Q($uuid:string, $hash:string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				# get all heuristic of that user and transaction
				var(func: uid($uuid)){
					h as User.heuristics@filter(uid_in(Heuristic.transaction, uid(tx))){
						hr as Heuristic.results
					}
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uuid": userUID, "$hash": txhash},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(hr) * * .\nuid(h) * * .\n<" + userUID + "> <User.heuristics> uid(h) ."),
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

// GetHeuristic gets heuristic information from the database
func GetHeuristic(c external.Database, heuristicUID string) (h Heuristic, err error) {
	query := `query Q($uid: string) {
				q(func: uid($uid)){
					uid
					Heuristic.type
					Heuristic.parameter
					Heuristic.results{
						HeuristicResult.origin {
							uid
						}
						HeuristicResult.destinations {
							uid
						}
					}
					Heuristic.ts
					Heuristic.transaction{
						uid
					}
					Heuristic.parent{
						uid
					}
					~Heuristic.parent{
						uid
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Second*20, query, map[string]string{"$uid": heuristicUID})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r struct {
		Heuristics []Heuristic `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Heuristics) != 1 {
		err = errors.New("invalid response from database")
		return
	}
	h = r.Heuristics[0]
	return
}

// GetHeuristicResults returns the connected transactions of heuristic
func GetHeuristicResults(c external.Database, heuristicUID string) (results []HeuristicTransaction, err error) {
	const query = `query Q($uid: string) {
				var (func: uid($uid)){
					Heuristic.results { x as HeuristicResult.origin }
				}
				
				q(func: uid(x)){
					uid
					tx_outputs{
						amount
					}
					tx_inputs(first:1)@normalize{
						~addr_outputs{
							addr_uid:uid
							~Cluster.addresses@filter(eq(Cluster.type,` + clustering.TypeFMI + `)){
								cluster_uid:uid
							}
						}
					}
			  	}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, string(query), map[string]string{"$uid": heuristicUID})

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

	for _, t := range r.Transaction {
		if len(t.Inputs) != 1 {
			err = fmt.Errorf("invalid cluster information for transaction %s", t.UID)
			return
		}

		var cUID ClusterUID

		// if the cluster is not set we use the address uid as the "cluster",
		//this can happen if the address has not been assigned a cluster yet
		if t.Inputs[0].Cluster == "" {
			cUID = ClusterUID(t.Inputs[0].Address)
		} else {
			cUID = ClusterUID(t.Inputs[0].Cluster)
		}

		results = append(results, HeuristicTransaction{
			UID:     t.UID,
			Cluster: cUID,
			Outputs: t.Outputs,
		})
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

// GetTransactionsWithOutputAmountAndCluster returns a slice of transactions.
// Each transaction contains its output amounts and the cluster of all inputs.
func GetTransactionsWithOutputAmountAndCluster(c external.Database, uids []string, userUID string,
	requestedClusterTypes []clustering.ClusterType) (origins []HeuristicTransaction, err error) {
	isSimpleClustering := requestedClusterTypes == nil // true -> only multi-input clusters will be used

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

	type mergedClusterItem struct {
		clusterHash string
		clusterUIDs map[string]bool
	}

	var mergedClusters []mergedClusterItem
	allClusters := make(map[string]bool)

	if !isSimpleClustering {
		// get all merged clusters
		for _, userCluster := range userClusterUIDs {
			// check if userCluster has already been found in a previous iteration
			if allClusters[userCluster] {
				continue
			}

			mergedClusterUIDS, err := clustering.GetRelatedClusters(c, userCluster, userUID, requestedClusterTypes)
			if err != nil {
				return nil, err
			}

			mergedCluster := make(map[string]bool)
			for _, mcu := range mergedClusterUIDS {
				mergedCluster[mcu] = true
				allClusters[mcu] = true
			}

			mergedClusters = append(mergedClusters, mergedClusterItem{clusterUIDs: mergedCluster})
		}
	}

	for _, o := range r.Origins {
		if o.Inputs == nil || o.Inputs[0].Address == nil {
			err = fmt.Errorf("invalid cluster information for transaction %s", o.UID)
			return
		}

		var cUID ClusterUID

		// if the cluster is not set we use the address uid as the "cluster",
		// this can happen if the address has not been assigned to a cluster yet
		if o.Inputs[0].Address[0].Cluster == nil {
			cUID = ClusterUID(o.Inputs[0].Address[0].UID)
		} else if isSimpleClustering {
			// must have a multi-input cluster UID
			cUID = ClusterUID(o.Inputs[0].Address[0].Cluster[0].UID)
		} else {
			// if this clusters uid does not appear in the merged user clusters,
			// then use the multi-input cluster UID
			firstClusterUId := o.Inputs[0].Address[0].Cluster[0].UID
			if !allClusters[firstClusterUId] {
				cUID = ClusterUID(firstClusterUId)
			} else {
				var found bool
				for i, mc := range mergedClusters {
					if _, ok := mc.clusterUIDs[firstClusterUId]; ok {

						// lazy creation of map hashes
						if mc.clusterHash == "" {
							mc.clusterHash = createKeyHash(mc.clusterUIDs)
							mergedClusters[i] = mc
						}

						cUID = ClusterUID(mc.clusterHash)
						found = true
						break
					}
				}
				if !found {
					return nil, errors.New("did not find cluster uid in merged cluster list")
				}
			}
		}

		origins = append(origins, HeuristicTransaction{
			UID:     o.UID,
			Cluster: cUID,
			Outputs: o.Outputs,
		})
	}

	return
}

func createKeyHash(someMap map[string]bool) string {
	var allKeys string
	for k := range someMap {
		allKeys += k
	}

	sha1Hash := sha1.New()
	sha1Hash.Write([]byte(allKeys))
	return base64.URLEncoding.EncodeToString(sha1Hash.Sum(nil))
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

	query := `query Q($hash:string, $uids:string, $type:string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				# filter and count
				q(func: uid($uids))@filter(uid_in(Heuristic.transaction, uid(tx)) AND eq(dgraph.type,$type)){
					count(uid)
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query,
		map[string]string{"$hash": txhash, "$uids": uidList, "$type": DType})
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

// GetBasicFrontendHeuristic returns all heuristics for a given transaction created by userUid. Basic information only
func GetBasicFrontendHeuristic(c external.Database, txHash string, userUID string) (heuristics []FrontendHeuristic, err error) {
	query := `query Q($hash: string, $uuid: string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				var(func: uid($uuid)){
					h as User.heuristics@filter(uid_in(Heuristic.transaction, uid(tx)))
				}
				
				q(func: uid(h)){
					uid
					ts:Heuristic.ts
					type:Heuristic.type
					parameter:Heuristic.parameter
					parent:Heuristic.parent{
						uid
					}
					children: ~Heuristic.parent{
						uid
					}
					num_results: count(Heuristic.results)
				}
			  }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash, "$uuid": userUID})
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

	if len(r.Heuristics) > 0 {
		heuristics = r.Heuristics
	}

	return
}

// GetFrontendHeuristicByUID the heuristic for the given heuristicUid
func GetFrontendHeuristicByUID(c external.Database, heuristicUID string, userUID string) (
	frontendHeuristic FrontendHeuristicShort, err error) {
	const query = `query Q($uid:string,$uuid:string){
				var(func:uid($uid))@cascade{
					~User.heuristics@filter(uid($uuid))
					r as Heuristic.results
				}

				q(func: uid(r)){
					origin:HeuristicResult.origin@normalize{
						txhash:txhash
						~transactions{
							ts:ts
						}
						tx_inputs(first:1)@normalize{
							~addr_outputs{
								addr as a:addresshash
								clusters as ~Cluster.addresses@filter(eq(Cluster.type,` + clustering.TypeFMI + `)){
									cuid:uid
								}
							}
						}
					}
					d:HeuristicResult.destinations{uid}
				}
				
				# get labels per cluster
				ca(func:uid(clusters))@cascade{
					uid
					cla:Cluster.addresses{
						attr:~Attribution.address{
							tag:Attribution.tag
							isPublic:Attribution.isPublic
						}
					}
				}

				# get labels per address, because not all addresses have an associated cluster
				aa(func:uid(addr))@cascade{
					a:addresshash
					attr:~Attribution.address{
						tag:Attribution.tag
						isPublic:Attribution.isPublic
					}
				}
			   }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, string(query), map[string]string{"$uid": heuristicUID, "$uuid": userUID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Results []struct {
			Origin []struct {
				Address     string `json:"a,omitempty"`
				Cluster     string `json:"cuid,omitempty"`
				Transaction string `json:"txhash,omitempty"`
				Timestamp   string `json:"ts,omitempty"`
			} `json:"origin,omitempty"`
			Destinations []struct {
				UID string `json:"uid,omitempty"`
			} `json:"d,omitempty"`
		} `json:"q,omitempty"`
		ClusterAttribution []struct {
			UID       string `json:"uid,omitempty"`
			Addresses []struct {
				Attributions []Attribution `json:"attr,omitempty"`
			} `json:"cla,omitempty"`
		} `json:"ca,omitempty"`
		AddressAttribution []struct {
			Address      string        `json:"a,omitempty"`
			Attributions []Attribution `json:"attr,omitempty"`
		} `json:"aa,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Results) == 0 {
		err = errors.New("invalid response from database")
		return
	}

	clusterDestinations := make(map[string]struct {
		destinations map[string]bool
		transactions []FrontendTransactionResult
	})

	for _, result := range r.Results {
		if len(result.Origin) != 1 {
			err = fmt.Errorf("invalid number of origins for heuristic %s", heuristicUID)
			return
		}

		if result.Origin[0].Cluster == "" && result.Origin[0].Address == "" {
			err = fmt.Errorf("invalid response from database for tx %s", result.Origin[0].Transaction)
			return
		}

		clusterID := result.Origin[0].Cluster
		if clusterID == "" {
			clusterID = result.Origin[0].Address
		}

		// get map item
		txMap := clusterDestinations[clusterID]
		txMap.transactions = append(txMap.transactions,
			FrontendTransactionResult{
				Hash:      result.Origin[0].Transaction,
				Address:   result.Origin[0].Address,
				Timestamp: result.Origin[0].Timestamp,
			})

		if txMap.destinations == nil {
			txMap.destinations = make(map[string]bool)
		}

		// add transaction uids
		for _, d := range result.Destinations {
			txMap.destinations[d.UID] = true
		}

		// set map item
		clusterDestinations[clusterID] = txMap
	}

	var results []FrontendHeuristicShortItem
	for k, v := range clusterDestinations {
		results = append(results, FrontendHeuristicShortItem{
			ClusterID:          k,
			CountForwardLookup: len(v.destinations),
			Transactions:       v.transactions,
		})
	}

	// gather address attributions
	var addressAttributions []AddressAttribution
	for _, aa := range r.AddressAttribution {
		addressAttributions = append(addressAttributions, AddressAttribution{
			Address:      aa.Address,
			Attributions: aa.Attributions,
		})
	}

	// gather cluster attributions
	var clusterAttributions []ClusterAttribution
	for _, ca := range r.ClusterAttribution {
		var attributions []Attribution
		for _, a := range ca.Addresses {
			attributions = append(attributions, a.Attributions...)
		}

		clusterAttributions = append(clusterAttributions, ClusterAttribution{
			UID:          ca.UID,
			Attributions: attributions,
		})
	}

	frontendHeuristic = FrontendHeuristicShort{
		UID:                 heuristicUID,
		ResultCount:         len(clusterDestinations),
		Results:             results,
		AddressAttributions: addressAttributions,
		ClusterAttributions: clusterAttributions,
	}

	return
}

// GetFrontendHeuristic returns all heuristics for a given transaction
func GetFrontendHeuristic(c external.Database, txHash string, userUID string) (completeHeuristic FrontendHeuristicComplete, err error) {
	query := `query Q($hash: string, $uuid: string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				var(func: uid($uuid)){
					h as User.heuristics@filter(uid_in(Heuristic.transaction, uid(tx)))
				}
				t(func: uid(tx))@normalize{
					uid:uid
					~transactions{
						ts:ts
					}
				}
				q(func: uid(h)){
					uid
					ts:Heuristic.ts
					type:Heuristic.type
					parameter:Heuristic.parameter
					parent:Heuristic.parent{
						uid
					}
					children: ~Heuristic.parent{
						uid
					}
					results:Heuristic.results{
						origin:HeuristicResult.origin@normalize{
							uid:uid
							txhash:txhash
							~transactions{
								ts:ts
							}
							tx_inputs{ 
								~addr_outputs{
									addresshash:addresshash
								}
							}
						}
						destinations:HeuristicResult.destinations@normalize{
							uid:uid
							txhash:txhash
							~transactions{
								ts:ts
							}
						}
					}
				}
			  }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash, "$uuid": userUID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Heuristics  []FrontendHeuristicResponse `json:"q,omitempty"`
		Transaction []FrontendHeuristicComplete `json:"t,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) != 1 || len(r.Heuristics) == 0 ||
		len(r.Transaction[0].UID) == 0 || len(r.Transaction[0].Timestamp) == 0 {
		err = errors.New("invalid response from database")
		return
	}

	completeHeuristic = r.Transaction[0]

	for _, h := range r.Heuristics {
		transactions := make(map[string]bool)

		var results []FrontendHeuristicResult
		for _, r := range h.Results {
			// only append a result once per transaction
			if transactions[r.Origin[0].UID] {
				continue
			}
			results = append(results, FrontendHeuristicResult{Origin: r.Origin[0], Destinations: r.Destinations})
			transactions[r.Origin[0].UID] = true
		}

		completeHeuristic.Heuristics = append(completeHeuristic.Heuristics, FrontendHeuristic{
			UID:             h.UID,
			Timestamp:       h.Timestamp,
			Type:            h.Type,
			Parameter:       h.Parameter,
			ParentHeuristic: h.ParentHeuristic,
			ChildHeuristics: h.ChildHeuristics,
			ResultCount:     h.ResultCount,
			Results:         results,
		})
	}

	return
}

// GetShortestTransactionPathAnyDirection returns the transactions of a shortest path between two transactions.
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
