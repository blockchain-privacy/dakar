package analytics

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/external"

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
							~cluster_addresses@filter(eq(cluster_type, "fmi")){ca as cluster_addresses}
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
					
					q(func: uid(all_not_mixing, mixing))@normalize{
						privacytype:privacytype
						~transactions{
							ts:ts
						}
					}
			  	}`, clusterQuery, clusterID)

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{"$address": addressHash})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	var r struct {
		Q []MixingActivity `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return r.Q, nil
}
