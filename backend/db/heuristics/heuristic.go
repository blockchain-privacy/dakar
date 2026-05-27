// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package heuristics

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/dakar/cmd/cliutil"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics/clustering"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

var errInvalidDatabaseResponse = errors.New("error invalid response")

// GetHeuristicTransactions returns the connected transactions of heuristic.
// Only outputs connected to transactions with the given transaction types are included.
func GetHeuristicTransactions(ctx context.Context, c external.Database, heuristicUID string,
	allowedTransactionType string) (results []HeuristicTransaction, attributionMap map[ClusterUID][]string, err error) {
	query := `query Q($uid:string,$type:string) {
				var (func: uid($uid)){ x as Selector.results }
				
				q(func: uid(x)){
					uid
					HeuristicCluster.results{
						uid
						tx_outputs@cascade{
							amount
							~tx_inputs@filter(eq(Transaction.type,$type))
						}
					}
					HeuristicCluster.attributions{
						uid
					}
			  	}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$uid": heuristicUID, "$type": allowedTransactionType})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Clusters []struct {
			UID     ClusterUID `json:"uid,omitempty"`
			Results []struct {
				UID     string            `json:"uid,omitempty"`
				Outputs []HeuristicOutput `json:"tx_outputs,omitempty"`
			} `json:"HeuristicCluster.results,omitempty"`
			Attributions []db.UIDNode `json:"HeuristicCluster.attributions,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	attributionMap = make(map[ClusterUID][]string)
	var clusterCounter int64
	for _, cluster := range r.Clusters {
		thisClusterID := ClusterUID(strconv.FormatInt(clusterCounter, 10))
		for _, result := range cluster.Results {
			results = append(results, HeuristicTransaction{
				UID:     result.UID,
				Cluster: thisClusterID,
				Outputs: result.Outputs,
			})
		}

		for _, attr := range cluster.Attributions {
			attributionMap[thisClusterID] = append(attributionMap[thisClusterID], attr.UID)
		}

		clusterCounter++
	}

	return
}

// GetHeuristicTransactionsOutputs returns the requested transaction with their output amounts
func GetHeuristicTransactionsOutputs(ctx context.Context, c external.Database, txUIDs []string,
	allowedTransactionType string) ([]HeuristicTransaction, error) {
	query := `query Q($txUIDs:string,$type:string) {
				q(func: uid($txUIDs)){
					uid
					tx_outputs@cascade{
						amount
						~tx_inputs@filter(eq(Transaction.type,$type))
					}
			  	}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$txUIDs": db.CreateCommaArray(txUIDs),
		"$type": allowedTransactionType})
	if err != nil {
		return nil, err
	}

	// json struct
	var r struct {
		Transactions []HeuristicTransaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Transactions, nil
}

// GetInputTransactions returns the input mixing transactions of the given transaction.
func GetInputTransactions(ctx context.Context, c external.Database,
	tx string, allowedTransactionType string) (inputTransactions []HeuristicTransaction, err error) {
	query := `query Q($txhash: string){
				var (func: eq(txhash,$txhash)){
					tx_inputs{
						v as ~tx_outputs@filter(eq(Transaction.type,"` + allowedTransactionType + `"))
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

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$txhash": tx})
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

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
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

// GetInputTransaction gets for given transaction hash the HeuristicTransaction
func GetInputTransaction(ctx context.Context, c external.Database, txhash string) (*HeuristicTransaction, error) {
	query := `query Q($txhash: string){
				q(func: eq(txhash,$txhash)){
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

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$txhash": txhash})
	if err != nil {
		return nil, err
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

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	if len(r.Transaction) != 1 {
		return nil, serror.New(errInvalidDatabaseResponse)
	}

	if len(r.Transaction[0].Block) != 1 || len(r.Transaction[0].Outputs) == 0 {
		return nil, serror.New(errInvalidDatabaseResponse)
	}

	return &HeuristicTransaction{
		UID:       r.Transaction[0].UID,
		Timestamp: r.Transaction[0].Block[0].Timestamp,
		Outputs:   r.Transaction[0].Outputs}, nil
}

// GetTransactionsWithOutputAmountAndCluster returns a slice of transactions and used attributions per cluster.
// Each transaction contains its output amounts and the clusters of all inputs.
// If no attributions were used or found the returned map is nil.
func GetTransactionsWithOutputAmountAndCluster(ctx context.Context, c external.Database, uids []string,
	userUID string, requestedClusterTypes []clustering.ClusterType, attributions map[string][]string,
	allowedTransactionType string) (origins []HeuristicTransaction, attributionMapping map[ClusterUID][]string, err error) {
	if allowedTransactionType == "" {
		return nil, nil, serror.FromStr("received empty transaction type")
	}

	isSimpleClustering := len(requestedClusterTypes) == 0 // true -> only multi-input clusters will be used

	// get user clusters if necessary
	var userClusterUIDs []string
	if !isSimpleClustering {
		userClusterUIDs, err = clustering.GetUserClustersUIDs(ctx, c, userUID, requestedClusterTypes)
		if err != nil {
			return
		}

		// if the user does not have defined any custom clusters,
		// then the request can be treated like multi-input only
		if userClusterUIDs == nil {
			isSimpleClustering = true
		}
	}

	query := `query Q($uids:string,$txType:string,$clusterType:string){
				q(func: uid($uids)){
					uid
					tx_outputs@cascade{
						amount
						~tx_inputs@filter(eq(Transaction.type,$txType))
					}
					tx_inputs(first:1){
						~addr_outputs{
							uid
							~Cluster.addresses@filter(eq(Cluster.type,$clusterType)){
								uid
							}
						}
					}
					~transactions{
						id
					}
			   	}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query,
		map[string]string{"$uids": db.CreateCommaArray(uids), "$txType": allowedTransactionType,
			"$clusterType": string(clustering.TypeFMI)})
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
			Block []struct {
				ID *int `json:"id,omitempty"`
			} `json:"~transactions,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
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

			mergedClusterUIDS, relatedErr := clustering.GetRelatedClusters(ctx, c, userCluster,
				userUID, requestedClusterTypes)
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

	type usedCluster struct{ superCluster map[string]bool }

	// holds all cluster IDs which are used by the generated HeuristicTransactions below
	allUsedClusters := make(map[string]usedCluster)
	for _, o := range r.Origins {
		var clusterUID string
		if o.Inputs == nil && o.Block != nil && o.Block[0].ID != nil {
			// coinbase transaction
			clusterUID = "coinbase_" + strconv.Itoa(*o.Block[0].ID)
		} else if o.Inputs != nil && o.Inputs[0].Address != nil && o.Inputs[0].Address[0].Cluster != nil {
			clusterUID = o.Inputs[0].Address[0].Cluster[0].UID
		} else {
			err = serror.FromFormat("invalid cluster information for transaction %s", o.UID)
			return
		}

		var cUID ClusterUID

		if isSimpleClustering || !allClusters[clusterUID] {
			// address of origin is only associated with multi-input clusters
			cUID = ClusterUID(clusterUID)
			allUsedClusters[clusterUID] = usedCluster{superCluster: nil}
		} else {
			var superCluster map[string]bool
			cUID, superCluster, err = getClusterUIDFromMergedClusters(superClusters, clusterUID)
			if err != nil {
				return
			}
			allUsedClusters[string(cUID)] = usedCluster{superCluster: superCluster}
		}

		origins = append(origins, HeuristicTransaction{
			UID:     o.UID,
			Cluster: cUID,
			Outputs: o.Outputs,
		})
	}

	if len(attributions) == 0 {
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
func GetTransactionsWithInputAmount(ctx context.Context, c external.Database,
	uids []string) (origins []HeuristicTransaction, err error) {
	query := `query Q($uids:string){
				q(func: uid($uids)){
					uid
					tx_inputs{
						amount
					}
				}
			   }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$uids": db.CreateCommaArray(uids)})
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

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
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

// GetInputAmounts gets the amounts of the inputs.
// Only inputs produced by transactions with the given transaction types are included.
func GetInputAmounts(ctx context.Context, c external.Database, tx string,
	allowedTransactionType string) (transaction HeuristicTransaction, err error) {
	query := `query Q($txhash:string,$type:string){
				q(func: eq(txhash,$txhash)){
					uid
					tx_inputs@cascade{
						amount
						~tx_outputs@filter(eq(Transaction.type,$type))
					}
				}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$txhash": tx, "$type": allowedTransactionType})
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

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
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

// GetNodeType returns the transaction type or heuristic type of the given node UID
func GetNodeType(ctx context.Context, c external.Database, uid string) (string, string, error) {
	if uid == "" {
		return "", "", serror.New(db.ErrEmptyRequestArgument)
	}

	const query = `query Q($uid:string){
				q(func: uid($uid)){
					Transaction.type
					Selector.options
					Selector.type
				}
			  }`

	resp, err := c.Query(ctx, query, map[string]string{"$uid": uid})
	if err != nil {
		return "", "", serror.New(err)
	}

	// json struct
	var r struct {
		Type []struct {
			TransactionType string `json:"Transaction.type,omitempty"`
			SelectorType    string `json:"Selector.type,omitempty"`
			SelectorOptions string `json:"Selector.options,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", "", serror.New(err)
	}

	if len(r.Type) != 1 {
		return "", "", serror.FromStrWithContext("invalid response", "response", r.Type)
	}

	var heuristicType string
	if r.Type[0].SelectorType == constants.TypeHeuristic {
		// try to extract heuristic options
		var opt struct {
			// Type is the type of the heuristic
			Type string `json:"type,omitempty"`
		}
		if err := json.Unmarshal([]byte(r.Type[0].SelectorOptions), &opt); err != nil {
			return "", "", serror.New(err)
		}
		heuristicType = opt.Type
	}

	return r.Type[0].TransactionType, heuristicType, nil
}
