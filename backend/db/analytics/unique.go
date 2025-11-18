// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"backend/db"
	"backend/external"
	"context"
	"encoding/json"
	"fmt"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"strconv"
	"time"
)

// GetUniqueAddressCountsPerBlock returns the number of unique addresses and clusters for the given day
// option == 1: count only output addresses and clusters
// option == 2: count only input addresses and clusters
// option == 3: count both input and output addresses and clusters
func GetUniqueAddressCountsPerBlock(ctx context.Context, c external.Database,
	date time.Time, option int) (addressCount int,
	clusterCount int, addressesWithClusterCount int, err error) {
	const outputAddressQuery = "tx_outputs { oa as ~addr_outputs}"
	const outputAddressVariable = "oa"
	const outputClusterVariable = "oc"
	const outputAddressesWithClusterVariable = "oawc"
	const outputClusterQuery = `tx_outputs {
									oawc as ~addr_outputs@cascade{
										oc as ~Cluster.addresses@filter(eq(Cluster.type, "fmi"))
									}
								}`

	const inputAddressQuery = "tx_inputs { ia as ~addr_outputs}"
	const inputAddressVariable = "ia"
	const inputClusterVariable = "ic"
	const inputAddressesWithClusterVariable = "iawc"
	const inputClusterQuery = `tx_inputs {
									iawc as ~addr_outputs@cascade{
										ic as ~Cluster.addresses@filter(eq(Cluster.type, "fmi"))
									}
								}`

	var addressSelector string
	var clusterSelector string
	var addressCountVariables string
	var clusterCountVariables string
	var addressesWithClusterVariables string
	switch option {
	case 1:
		addressSelector = outputAddressQuery
		clusterSelector = outputClusterQuery
		addressCountVariables = outputAddressVariable
		clusterCountVariables = outputClusterVariable
		addressesWithClusterVariables = outputAddressesWithClusterVariable
	case 2:
		addressSelector = inputAddressQuery
		clusterSelector = inputClusterQuery
		addressCountVariables = inputAddressVariable
		clusterCountVariables = inputClusterVariable
		addressesWithClusterVariables = inputAddressesWithClusterVariable
	case 3:
		addressSelector = outputAddressQuery + " \n " + inputAddressQuery
		clusterSelector = outputClusterQuery + " \n " + inputClusterQuery
		addressCountVariables = outputAddressVariable + ", " + inputAddressVariable
		clusterCountVariables = outputClusterVariable + ", " + inputClusterVariable
		addressesWithClusterVariables = outputAddressesWithClusterVariable + ", " + inputAddressesWithClusterVariable
	default:
		err = serror.FromStr("invalid option")
		return
	}

	toDate := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999, date.Location())
	var query = fmt.Sprintf(`query Q($from:string,$to:string) {
					blocks as var(func: between(ts, $from, $to))@filter(has(blockhash))

					var(func: uid(blocks)){
						transactions {
							%s
						}
					}

					var(func: uid(blocks)){
						transactions {
							%s
						}
					}
					
					address_count(func: uid(%s)){
						count(uid)
					}

					cluster_count(func: uid(%s)){
						count(uid)
					}

					addresses_with_clusters(func: uid(%s)){
						count(uid)
					}
				  }`, addressSelector, clusterSelector, addressCountVariables,
		clusterCountVariables, addressesWithClusterVariables)

	resp, err := db.QueryVarWithRetry(ctx, c, query,
		map[string]string{"$to": toDate.Format(time.RFC3339), "$from": date.Format(time.RFC3339)})
	if err != nil {
		return
	}

	var r struct {
		AddressCount []struct {
			Count int `json:"count,omitempty"`
		} `json:"address_count,omitempty"`
		ClusterCount []struct {
			Count int `json:"count,omitempty"`
		} `json:"cluster_count,omitempty"`
		AddressesWithCluster []struct {
			Count int `json:"count,omitempty"`
		} `json:"addresses_with_clusters,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.AddressCount) != 1 || len(r.ClusterCount) != 1 || len(r.AddressesWithCluster) != 1 {
		err = serror.FromStr("invalid response from database")
		return
	}

	addressCount = r.AddressCount[0].Count
	clusterCount = r.ClusterCount[0].Count
	addressesWithClusterCount = r.AddressesWithCluster[0].Count

	return
}

func BlockHeightToTimestamp(ctx context.Context, c external.Database, blockHeight int64) (timestamp string, err error) {
	const query = `query Q($height:string) {
					q(func: eq(id, $height)){
						ts
					}
				  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query,
		map[string]string{"$height": strconv.FormatInt(blockHeight, 10)})
	if err != nil {
		return
	}

	var r struct {
		Query []struct {
			Timestamp string `json:"ts,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Query) != 1 {
		err = serror.FromStr("invalid response from database")
		return
	}

	timestamp = r.Query[0].Timestamp

	return
}
