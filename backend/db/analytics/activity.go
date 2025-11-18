// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"backend/constants"
	"backend/external"
	"context"
	"encoding/json"
	"fmt"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// Functions for detecting mixing activity

// GetMixingActivity returns all classified transactions directly connected
// to the cluster (of the given address) and directly connected to all collateral
// transactions of the cluster.
// If isClusterLookup is false, only the given address and its connected transactions will be considered.
func GetMixingActivity(ctx context.Context, c external.Database,
	addressHash string, isClusterLookup bool) ([]MixingActivity, error) {
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
					
					not_mixing as var(func: uid(t1,t2))@filter(has(Transaction.type) and not eq(Transaction.type,`+constants.AllMixingTypes+`))
					
					# outputs of destinations do not all belong to the user
					dst as var(func: uid(not_mixing))@filter(eq(Transaction.type,`+constants.AllDestinationTypes+`))

					# outputs of wasabi 2.0 origins do not all belong to the user
					wasabi_2_origins as var(func: uid(not_mixing))@filter(eq(Transaction.type,"`+constants.TypeWasabi2Origin+`"))


					# all origins except wasabi 2.0 and collateral transactions
					origins_and_collaterals as var(func: uid(not_mixing))@filter(not uid(dst) and not uid(wasabi_2_origins))

					var(func: uid(origins_and_collaterals))@recurse{
						tx_outputs
						recurse_origins_and_collaterals as ~tx_inputs@filter(has(Transaction.type) and not eq(Transaction.type,`+constants.AllMixingTypes+`))
					}
					
					
					all_not_mixing_not_dst as var(func: uid(origins_and_collaterals, recurse_origins_and_collaterals, wasabi_2_origins))

					var(func: uid(all_not_mixing_not_dst)){
						tx_outputs {
							mixing1 as ~tx_inputs@filter(eq(Transaction.type,`+constants.AllMixingTypes+`))
						}
					}

					var(func: uid(dst)){
						tx_inputs {
							mixing2 as ~tx_outputs@filter(eq(Transaction.type,`+constants.AllMixingTypes+`))
						}
					}

					transactions as var(func: uid(all_not_mixing_not_dst, mixing1, mixing2, dst))
					
					q(func: uid(transactions)){
						txhash
						txtype:Transaction.type
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

	resp, err := c.Query(ctx, query, map[string]string{"$address": addressHash})
	if err != nil {
		return nil, serror.New(err)
	}

	var r struct {
		Q []MixingActivity `json:"q"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	// filter duplicate transaction hashes (due to one hash per output)
	activities := make([]MixingActivity, len(r.Q))
	for i, ma := range r.Q {
		newActivity := MixingActivity{
			TransactionHash: ma.TransactionHash,
			TransactionType: ma.TransactionType,
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
