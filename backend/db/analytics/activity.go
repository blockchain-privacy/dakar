package analytics

import (
	"backend/constants"
	"backend/db"
	"backend/external"
	"github.com/qrest/gomisc/serror"

	"encoding/json"
	"fmt"
	"time"
)

// Functions for detecting mixing activity

// GetMixingActivity returns all privacy transactions directly connected
// to the cluster (of the given address) and directly connected to all collateral
// transactions of the cluster.
// If isClusterLookup is false, only the given address and its connected transactions will be considered.
func GetMixingActivity(c external.Database, addressHash string, isClusterLookup bool) ([]MixingActivity, error) {
	var clusterID, clusterQuery string

	if isClusterLookup {
		clusterID = ",ca"
		clusterQuery = `var(func: uid(addr)){
							~Cluster.addresses@filter(eq(Cluster.type, "fmi")){ca as Cluster.addresses}
						}`
	}

	query := fmt.Sprintf(`query Q($address: string)
				{
					addr as var(func: eq(addresshash,$address))
					
					# conditional cluster lookup0
					%s
					var(func: uid(addr%s)){
						addr_outputs {
							t1 as ~tx_inputs
							t2 as ~tx_outputs
						}
					}
					
					not_mixing as var(func: uid(t1,t2))@filter(between(privacytype,`+
		constants.StrPrivacyDestinationFirst+","+constants.StrPrivacyCollateralPaymentLast+`))
					
					var(func: uid(not_mixing))@recurse{
						tx_outputs
						c_not_mixing as ~tx_inputs@filter(between(privacytype,`+
		constants.StrPrivacyDestinationFirst+","+constants.StrPrivacyCollateralPaymentLast+`))
					}
					
					all_not_mixing as var(func: uid(not_mixing, c_not_mixing))
					
					var(func: uid(all_not_mixing)){
						tx_outputs {
							mixing as ~tx_inputs@filter(between(privacytype,0,`+constants.StrPrivacyMixingLast+`))
						}
					}
					
					transactions as var(func: uid(all_not_mixing, mixing))
					
					q(func: uid(transactions)){
						txhash
						privacytype
						block:~transactions{
						ts
						}
						input_txs:tx_inputs @normalize{
							~tx_outputs@filter(uid(transactions)){
								txhash:txhash
							}
						}
					}
			  	}`, clusterQuery, clusterID)

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{"$address": addressHash})
	if err != nil {
		return nil, err
	}

	var r struct {
		Q []MixingActivity `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, serror.NewStackError(err)
	}

	// filter duplicate transaction hashes (due to one hash per output)
	activities := make([]MixingActivity, len(r.Q))
	for i, ma := range r.Q {
		newActivity := MixingActivity{
			TransactionHash: ma.TransactionHash,
			PrivacyType:     ma.PrivacyType,
			Block:           ma.Block,
		}

		transactions := make(map[string]bool)

		for _, t := range ma.InputTransactions {
			transactions[t.TransactionHash] = true
		}

		for k := range transactions {
			newActivity.InputTransactions = append(newActivity.InputTransactions, struct {
				TransactionHash string `json:"txhash"`
			}{TransactionHash: k})
		}

		activities[i] = newActivity
	}

	return activities, nil
}
