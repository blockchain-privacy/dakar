package analytics

import (
	"backend/db"
	"backend/external"
	"encoding/json"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"strconv"
	"time"
)

// GetUniqueAddressCountsPerBlock returns the number of unique addresses and clusters for the given day
// option == 1: count only output addresses and clusters
// option == 2: count only input addresses and clusters
// option == 3: count both input and output addresses and clusters
func GetUniqueAddressCountsPerBlock(c external.Database, date time.Time, option int) (addressCount uint64,
	clusterCount uint64, addressesWithClusterCount uint64, err error) {
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
		err = serror.NewStackErrorStr("invalid option")
		return
	}

	toDate := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999, date.Location())
	var query = fmt.Sprintf(`query Q($from:string,$to:string) {
					blocks as var(func: between(ts, $from, $to))@filter(type(Block))

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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$to": toDate.Format(time.RFC3339), "$from": date.Format(time.RFC3339)})
	if err != nil {
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
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.NewStackError(err)
		return
	}

	if len(r.AddressCount) != 1 || len(r.ClusterCount) != 1 || len(r.AddressesWithCluster) != 1 {
		err = serror.NewStackErrorStr("invalid response from database")
		return
	}

	addressCount = r.AddressCount[0].Count
	clusterCount = r.ClusterCount[0].Count
	addressesWithClusterCount = r.AddressesWithCluster[0].Count

	return
}

func BlockHeightToTimestamp(c external.Database, blockHeight uint64) (timestamp string, err error) {
	const query = `query Q($height:string) {
					q(func: eq(id, $height)){
						ts
					}
				  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$height": strconv.FormatUint(blockHeight, 10)})
	if err != nil {
		return
	}

	var r struct {
		Query []struct {
			Timestamp string `json:"ts,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.NewStackError(err)
		return
	}

	if len(r.Query) != 1 {
		err = serror.NewStackErrorStr("invalid response from database")
		return
	}

	timestamp = r.Query[0].Timestamp

	return
}
