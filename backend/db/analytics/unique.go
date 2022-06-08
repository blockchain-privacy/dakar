package analytics

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// GetUniqueAddressCountsPerBlock returns the number of unique addresses contained in the given block
// option == 1: count only output addresses and clusters
// option == 2: count only input addresses and clusters
// option == 3: count both input and output addresses and clusters
func GetUniqueAddressCountsPerBlock(c external.Database, blockID uint64, option int) (addressCount uint64,
	clusterCount uint64, addressesWithClusterCount uint64, timestamp string, err error) {
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
		err = errors.New("invalid option")
		return
	}

	var query = fmt.Sprintf(`query Q($block:string) {
					var(func: eq(id,$block)){
						t as ts 
						transactions {
							%s
						}
					}

					var(func: eq(id,$block)){
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
					
					ts(func: uid(t)){
						ts:val(t)
					}
				  }`, addressSelector, clusterSelector, addressCountVariables,
		clusterCountVariables, addressesWithClusterVariables)

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$block": strconv.FormatUint(blockID, 10)})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		AddressCount []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"address_count,omitempty"`
		ClusterCount []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"cluster_count,omitempty"`
		AddressesWithCluster []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"addresses_with_clusters,omitempty"`
		Timestamp []struct {
			TS string `json:"ts,omitempty"`
		} `json:"ts,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.AddressCount) != 1 || len(r.ClusterCount) != 1 || len(r.AddressesWithCluster) != 1 || len(r.Timestamp) != 1 {
		err = errors.New("invalid response from database")
		return
	}

	addressCount = r.AddressCount[0].Count
	clusterCount = r.ClusterCount[0].Count
	addressesWithClusterCount = r.AddressesWithCluster[0].Count
	timestamp = r.Timestamp[0].TS

	return
}
