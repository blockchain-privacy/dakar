package workspace

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"encoding/json"
)

type NodeConnections struct {
	UID          string   `json:"uid,omitempty"`
	Clusters     []string `json:"clusters,omitempty"`
	Transactions []string `json:"transactions,omitempty"`
}

// GetWorkspaceConnections returns all connections between the given UIDs
func GetWorkspaceConnections(c external.Database, uids []string) (transactions []NodeConnections, clusters []NodeConnections, err error) {
	// need at least two uids to find connections
	if len(uids) < 2 {
		err = cliutil.NewStackError(db.ErrEmptyRequestArgument)
		return
	}
	const query = `query Q($uids:string){
			# input uids
			uids as var(func: uid($uids))
			
			# transaction uids
			t as var(func: uid(uids))@filter(type("Transaction"))
			
			# flat cluster uids
			c as var(func: uid(uids))@filter(type("Cluster") and not eq(Cluster.type, "hmi"))
			
			transactions(func: uid(t)){
				uid
				tx_outputs{
					~tx_inputs@filter(uid(t)){
						uid
					}
					~addr_outputs{
						~Cluster.addresses@filter(uid(c)){
							uid
						}
					}
				}
			
				tx_inputs {
					~tx_outputs@filter(uid(t)){
						uid
					}
					~addr_outputs{
						~Cluster.addresses@filter(uid(c)){
							uid
						}
					}
				}
			}
			
			cluster_transactions(func: uid(c)){
				uid
				Cluster.addresses {
					addr_outputs {
						~tx_inputs@filter(uid(t)){
							uid
						}
						~tx_outputs@filter(uid(t)){
							uid
						}
					}
				}
			}

			cluster_clusters(func: uid(c))@ignorereflex{
				uid
				Cluster.addresses {
					addr_outputs {
						~tx_inputs@normalize{
							tx_outputs{
								~addr_outputs{
									~Cluster.addresses@filter(uid(c)){
										uid:uid
									}
								}
							}
						}
						~tx_outputs@normalize{
							tx_inputs{
								~addr_outputs{
									~Cluster.addresses@filter(uid(c)){
										uid:uid
									}
								}
							}
						}
					}
				}
			}
		}`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$uids": db.CreateCommaArray(uids)})
	if err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	// json struct
	var r struct {
		Transactions []struct {
			UID     string `json:"uid,omitempty"`
			Outputs []struct {
				InputTransactions []struct {
					UID string `json:"uid,omitempty"`
				} `json:"~tx_inputs,omitempty"`
				Addresses []struct {
					Clusters []struct {
						UID string `json:"uid,omitempty"`
					} `json:"~Cluster.addresses,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_outputs,omitempty"`
			Inputs []struct {
				OutputTransactions []struct {
					UID string `json:"uid,omitempty"`
				} `json:"~tx_outputs,omitempty"`
				Addresses []struct {
					Clusters []struct {
						UID string `json:"uid,omitempty"`
					} `json:"~Cluster.addresses,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_inputs,omitempty"`
		} `json:"transactions,omitempty"`
		ClusterTransactions []struct {
			UID       string `json:"uid,omitempty"`
			Addresses []struct {
				Outputs []struct {
					InputTransactions []struct {
						UID string `json:"uid,omitempty"`
					} `json:"~tx_inputs,omitempty"`
					OutputTransactions []struct {
						UID string `json:"uid,omitempty"`
					} `json:"~tx_outputs,omitempty"`
				} `json:"addr_outputs,omitempty"`
			} `json:"Cluster.addresses,omitempty"`
		} `json:"cluster_transactions,omitempty"`
		ClusterClusters []struct {
			UID       string `json:"uid,omitempty"`
			Addresses []struct {
				Outputs []struct {
					InputClusters []struct {
						UID string `json:"uid,omitempty"`
					} `json:"~tx_inputs,omitempty"`
					OutputClusters []struct {
						UID string `json:"uid,omitempty"`
					} `json:"~tx_outputs,omitempty"`
				} `json:"addr_outputs,omitempty"`
			} `json:"Cluster.addresses,omitempty"`
		} `json:"cluster_clusters,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	connectedClusters := map[string]NodeConnections{}

	for _, ct := range r.ClusterTransactions {
		txUIDs := map[string]bool{}
		for _, address := range ct.Addresses {
			for _, output := range address.Outputs {
				for _, it := range output.InputTransactions {
					txUIDs[it.UID] = true
				}
				for _, ot := range output.OutputTransactions {
					txUIDs[ot.UID] = true
				}
			}
		}

		// at this point each cluster is new, no need to update
		connectedClusters[ct.UID] = NodeConnections{UID: ct.UID, Transactions: cliutil.GetMapKeys(txUIDs)}
	}

	for _, cluster := range r.ClusterClusters {
		clusterUIDs := map[string]bool{}
		for _, address := range cluster.Addresses {
			for _, output := range address.Outputs {
				for _, outputCluster := range output.OutputClusters {
					clusterUIDs[outputCluster.UID] = true
				}
				for _, inputCluster := range output.InputClusters {
					clusterUIDs[inputCluster.UID] = true
				}
			}
		}
		nodeConnection := connectedClusters[cluster.UID]
		nodeConnection.Clusters = cliutil.GetMapKeys(clusterUIDs)
		connectedClusters[cluster.UID] = nodeConnection
	}

	for _, v := range connectedClusters {
		clusters = append(clusters, v)
	}

	for _, queryTx := range r.Transactions {
		clusterUIDs := map[string]bool{}
		transactionUIDs := map[string]bool{}

		for _, output := range queryTx.Outputs {
			for _, inputTx := range output.InputTransactions {
				transactionUIDs[inputTx.UID] = true
			}

			for _, address := range output.Addresses {
				for _, cluster := range address.Clusters {
					clusterUIDs[cluster.UID] = true
				}
			}
		}

		for _, inputs := range queryTx.Inputs {
			for _, outputTx := range inputs.OutputTransactions {
				transactionUIDs[outputTx.UID] = true
			}

			for _, address := range inputs.Addresses {
				for _, cluster := range address.Clusters {
					clusterUIDs[cluster.UID] = true
				}
			}
		}

		transactions = append(transactions, NodeConnections{
			UID:          queryTx.UID,
			Transactions: cliutil.GetMapKeys(transactionUIDs),
			Clusters:     cliutil.GetMapKeys(clusterUIDs),
		})
	}

	return
}
