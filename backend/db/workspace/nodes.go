package workspace

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/heuristics"
	"backend/external"
	"encoding/json"
	"time"
)

// GetFMIClustersByAddress returns uid mappings between addresses and their associated FMI cluster
func GetFMIClustersByAddress(c external.Database, addresses []string) (map[string]string, error) {
	if len(addresses) == 0 {
		return nil, cliutil.NewStackError(db.ErrEmptyRequestArgument)
	}

	// todo check address input at some point, also handle merging of clusters in client data
	query := `query {
			q(func: eq(addresshash,` + db.CreateCommaList(addresses) + `)){
				addresshash
				~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
					uid
				}
			}
		}`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*2, query)
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	// json struct
	var r struct {
		Addresses []struct {
			Addresshash string `json:"addresshash,omitempty"`
			Cluster     []struct {
				UID string `json:"uid,omitempty"`
			} `json:"~Cluster.addresses,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, cliutil.NewStackError(err)
	}

	addressClusterMapping := map[string]string{}
	for _, address := range r.Addresses {
		if len(address.Cluster) == 1 {
			addressClusterMapping[address.Addresshash] = address.Cluster[0].UID
		}
	}

	return addressClusterMapping, nil
}

// GetWorkspaceConnections returns all connections between the given UIDs, and all connected heuristics
func GetWorkspaceConnections(c external.Database, uids []string, userUID string) (
	connections []NodeConnections, heuristicNodes []FrontendGraphNode, err error) {
	// need at least two uids to find connections
	if len(uids) < 2 {
		err = cliutil.NewStackError(db.ErrEmptyRequestArgument)
		return
	}

	// todo limit user for cluster lookup
	const query = `query Q($uids:string,$user:string){
					# input uids
					uids as var(func: uid($uids))
					
					# transaction uids
					t as var(func: uid(uids))@filter(type("Transaction"))
					
					# heuristic uids
					var(func: uid($user)){
						h as User.heuristics@filter(uid_in(Heuristic.transaction, uid(t)))
					}

					# flat cluster uids
					c as var(func: uid(uids))@filter(type("Cluster") and not eq(Cluster.type, "hmi"))
					
					# address uids
					a as var(func: uid(uids))@filter(type("Address"))
					
					transactions(func: uid(t)){
						uid
						tx_outputs{
							~tx_inputs@filter(uid(t)){
								uid
							}
							...fGetCluster
						}
						tx_inputs {
							~tx_outputs@filter(uid(t)){
								uid
							}
							...fGetCluster
						}
					}

					transaction_addresses(func: uid(t)){
						uid
						tx_outputs{
							~addr_outputs@filter(uid(a)){
								uid
							}
						}
						tx_inputs {
							~addr_outputs@filter(uid(a)){
								uid
							}
						}
					}
					
					address_clusters(func: uid(a)){
						uid
						...fCluster
					}
				
					address_addresses(func: uid(a))@ignorereflex{
						uid
						addr_outputs {
							~tx_inputs{
								tx_outputs{
									~addr_outputs@filter(uid(a)){
										uid
									}
								}
							}
							~tx_outputs{
								tx_inputs{
									~addr_outputs@filter(uid(a)){
										uid
									}
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

					heuristics(func: uid(t)){
						uid
						~Heuristic.transaction@filter(uid(h)){` + heuristics.QueryBasicHeuristicAttributes + `}
					}
				}

				fragment fCluster {
					addr_outputs {
						~tx_inputs@normalize{
							tx_outputs{
								...fGetCluster
							}
						}
						~tx_outputs@normalize{
							tx_inputs{
								...fGetCluster
							}
						}
					}
				}
				
				fragment fGetCluster {
					~addr_outputs{
						~Cluster.addresses@filter(uid(c)){
							uid:uid
						}
					}
				}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{
		"$uids": db.CreateCommaArray(uids), "$user": userUID})
	if err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	// json struct
	var r connectionRequest

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	transactions, clusters, addresses, heuristicNodes := parseConnectionResult(r)
	connections = append(transactions, append(addresses, clusters...)...)

	return
}

// parseConnectionResult parses the result of a connection request and returns the resulting connections
//
//nolint:gocyclo
func parseConnectionResult(r connectionRequest) (transactions []NodeConnections, clusters []NodeConnections,
	addresses []NodeConnections, heuristics []FrontendGraphNode) {
	// txToHeuristic contains the mapping of transaction to its directly connected heuristics (root heuristics).
	// This map is used to add the contained heuristic uids as children to their corresponding transaction.
	txToHeuristic := map[string][]string{}
	for _, heuristicTransaction := range r.HeuristicTransactions {
		var rootHeuristics []string
		for _, h := range heuristicTransaction.Heuristics {
			// todo review once go 1.22 is released
			tmpHeuristic := h
			// no parent -> root heuristic
			if len(h.ParentHeuristic) == 0 {
				rootHeuristics = append(rootHeuristics, h.UID)
			}

			children := make([]string, len(h.ChildHeuristics))
			for i, c := range h.ChildHeuristics {
				children[i] = c.UID
			}

			heuristics = append(heuristics, FrontendGraphNode{
				UID:                 h.UID,
				Type:                "heuristic",
				Children:            children,
				HeuristicType:       h.Type,
				Parameter:           h.Parameter,
				ExcludeAddresses:    &tmpHeuristic.ExcludeAddresses,
				ExcludeSpendingGaps: &tmpHeuristic.ExcludeSpendingGaps,
				ClusterTypes:        h.ClusterTypes,
				ClusterCount:        tmpHeuristic.ClusterCount,
				Timestamp:           h.Timestamp,
			})
		}
		txToHeuristic[heuristicTransaction.UID] = rootHeuristics
	}

	connectedTransactions := map[string]NodeConnections{}
	for _, queryTx := range r.Transactions {
		children := map[string]bool{}

		// add root heuristics to transaction if available
		if rootHeuristics, ok := txToHeuristic[queryTx.UID]; ok {
			for _, h := range rootHeuristics {
				children[h] = true
			}
		}

		for _, output := range queryTx.Outputs {
			for _, inputTx := range output.InputTransactions {
				children[inputTx.UID] = true
			}

			for _, address := range output.Addresses {
				for _, cluster := range address.Clusters {
					children[cluster.UID] = true
				}
			}
		}

		for _, inputs := range queryTx.Inputs {
			for _, outputTx := range inputs.OutputTransactions {
				children[outputTx.UID] = true
			}

			for _, address := range inputs.Addresses {
				for _, cluster := range address.Clusters {
					children[cluster.UID] = true
				}
			}
		}

		connectedTransactions[queryTx.UID] = NodeConnections{
			UID:      queryTx.UID,
			Children: cliutil.GetMapKeys(children),
		}
	}

	for _, queryTx := range r.TransactionAddresses {
		addressUIDs := map[string]bool{}

		for _, output := range queryTx.Outputs {
			for _, address := range output.Addresses {
				addressUIDs[address.UID] = true
			}
		}

		for _, inputs := range queryTx.Inputs {
			for _, address := range inputs.Addresses {
				addressUIDs[address.UID] = true
			}
		}

		cTx := connectedTransactions[queryTx.UID]
		cTx.Children = append(cTx.Children, cliutil.GetMapKeys(addressUIDs)...)
		connectedTransactions[queryTx.UID] = cTx
	}

	connectedAddresses := map[string]NodeConnections{}
	for _, a := range r.AddressClusters {
		clusterUIDs := map[string]bool{}
		for _, output := range a.AddressOutputs {
			for _, ot := range output.OutputTransaction {
				clusterUIDs[ot.ClusterUID] = true
			}
			for _, it := range output.InputTransaction {
				clusterUIDs[it.ClusterUID] = true
			}
		}
		connectedAddresses[a.UID] = NodeConnections{UID: a.UID, Children: cliutil.GetMapKeys(clusterUIDs)}
	}

	transactions = make([]NodeConnections, 0, len(connectedTransactions))
	for _, v := range connectedTransactions {
		transactions = append(transactions, v)
	}

	for _, a := range r.AddressAddresses {
		addressUIDs := map[string]bool{}
		for _, output := range a.AddressOutputs {
			for _, ot := range output.OutputTransaction {
				for _, i := range ot.Inputs {
					for _, address := range i.Addresses {
						addressUIDs[address.UID] = true
					}
				}
			}
			for _, it := range output.InputTransaction {
				for _, i := range it.Outputs {
					for _, address := range i.Addresses {
						addressUIDs[address.UID] = true
					}
				}
			}
		}

		address := connectedAddresses[a.UID]
		address.Children = append(address.Children, cliutil.GetMapKeys(addressUIDs)...)
		connectedAddresses[a.UID] = address
	}

	addresses = make([]NodeConnections, 0, len(connectedAddresses))
	for _, v := range connectedAddresses {
		addresses = append(addresses, v)
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
		clusters = append(clusters, NodeConnections{UID: cluster.UID, Children: cliutil.GetMapKeys(clusterUIDs)})
	}

	return
}

// SearchForNode returns the uid which matches to the given query. In case the query is an address
// which is connected to clusters, they are returned instead.
func SearchForNode(c external.Database, nodeQuery string, userUID string) (nodes []GraphNode, err error) {
	if nodeQuery == "" || userUID == "" {
		err = cliutil.NewStackError(db.ErrEmptyRequestArgument)
		return
	}
	const query = `query Q($query:string, $user:string){
						transaction(func: eq(txhash, $query)){
							uid
							privacytype
						}
						
						address(func: eq(addresshash, $query)){
							uid
							~Cluster.addresses@filter(eq(Cluster.type, "fmi") or (eq(Cluster.type, "custom") and uid_in(Cluster.user, $user))){
								uid
								Cluster.type
							}
						}
					}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{"$query": nodeQuery, "$user": userUID})
	if err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	// json struct
	var r struct {
		Transactions []struct {
			UID         string `json:"uid,omitempty"`
			Hash        string `json:"txhash,omitempty"`
			PrivacyType *int   `json:"privacytype,omitempty"`
		} `json:"transaction,omitempty"`
		Address []struct {
			UID      string `json:"uid,omitempty"`
			Clusters []struct {
				UID  string `json:"uid,omitempty"`
				Type string `json:"Cluster.type,omitempty"`
			} `json:"~Cluster.addresses,omitempty"`
		} `json:"address,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	if len(r.Transactions) > 0 {
		tx := r.Transactions[0]
		nodes = []GraphNode{{UID: tx.UID, Type: "transaction", TransactionHash: nodeQuery}}
		nodes[0].PrivacyType = tx.PrivacyType
		return
	}

	if len(r.Address) > 0 {
		addr := r.Address[0]
		if len(addr.Clusters) > 0 {
			nodes = make([]GraphNode, len(addr.Clusters))
			for i, cluster := range addr.Clusters {
				nodes[i] = GraphNode{
					UID:         cluster.UID,
					Type:        "cluster",
					AddressHash: nodeQuery,
					ClusterType: cluster.Type,
				}
			}
			return
		}

		nodes = []GraphNode{{
			UID:         addr.UID,
			Type:        "cluster",
			AddressHash: nodeQuery,
		}}
		return
	}

	return
}
