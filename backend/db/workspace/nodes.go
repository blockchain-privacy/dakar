package workspace

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"encoding/json"
	"slices"
	"time"
)

// GetFMIClustersByAddress returns uid mappings between addresses and their associated FMI cluster
func GetFMIClustersByAddress(c external.Database, addresses []string) (map[string]string, error) {
	if len(addresses) == 0 {
		return nil, cliutil.NewStackError(db.ErrEmptyRequestArgument)
	}
	// todo check if this function is still needed
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

	const query = `query Q($uids:string,$user:string){
					# input uids
					uids as var(func: uid($uids))
					
					# transaction uids
					t as var(func: uid(uids))@filter(type("Transaction"))
					
					# heuristic uids
					var(func: uid($user)){
						h as User.heuristics@filter(uid_in(Heuristic.transaction, uid(t)))
					}

					# find fmi cluster for each address
					address_cluster(func: uid(uids))@filter(type("Address")){
						uid
						c as cluster:~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
							uid
						}
					}
					
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
						~Heuristic.transaction@filter(uid(h)){
								uid
								ts:Heuristic.ts
								type:Heuristic.type
								parameter:Heuristic.parameter
								clusterTypes:Heuristic.clusterTypes
								excludeAddresses:Heuristic.excludeAddresses
								excludeSpendingGaps:Heuristic.excludeSpendingGaps
								parent:Heuristic.parent{
									uid
								}
								children:~Heuristic.parent{
									uid
								}
								clusterCount: count(Heuristic.clusters)
								Heuristic.clusters{
									HeuristicCluster.results{
										HeuristicResult.origin@filter(uid(t)){
											uid
										}
										HeuristicResult.destinations@filter(uid(t)){
											uid
										}
									}
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

	transactions, clusters, heuristicNodes := parseConnectionResult(r)

	connections = slices.Concat(transactions, clusters)

	return
}

// parseConnectionResult parses the result of a connection request and returns the resulting connections
//
//nolint:gocyclo
func parseConnectionResult(r connectionRequest) (transactions []NodeConnections, clusters []NodeConnections, heuristics []FrontendGraphNode) {
	// clusterToAddress contains the mapping of flat multi-input clusters to their addresses.
	// This map is used to replace the uid of clusters with the uid of addresse.
	// This is done because we ultimatly want to store the address uids, not the cluster uids as they are not static.
	clusterToAddress := map[string]string{}
	for _, address := range r.AddressClusters {
		if len(address.Cluster) != 1 {
			continue
		}

		clusterToAddress[address.Cluster[0].UID] = address.UID
	}

	// txToHeuristic contains the mapping of transaction to its directly connected heuristics (root heuristics).
	// This map is used to add the contained heuristic uids as children to their corresponding transaction.
	txToHeuristic := map[string][]string{}
	for _, heuristicTransaction := range r.Heuristics {
		var rootHeuristics []string
		for _, h := range heuristicTransaction.Heuristics {
			// no parent -> root heuristic
			if len(h.ParentHeuristic) == 0 {
				rootHeuristics = append(rootHeuristics, h.UID)
			}

			children := make([]string, len(h.ChildHeuristics))
			for i, c := range h.ChildHeuristics {
				children[i] = c.UID
			}

			// add connections between heuristics and their found origins
			for _, cluster := range h.Clusters {
				for _, result := range cluster.Results {
					children = append(children, result.Origin.UID)
					for _, destination := range result.Destinations {
						children = append(children, destination.UID)
					}
				}
			}

			heuristics = append(heuristics, FrontendGraphNode{
				UID:                 h.UID,
				Type:                "heuristic",
				Children:            children,
				HeuristicType:       h.Type,
				Parameter:           h.Parameter,
				ExcludeAddresses:    &h.ExcludeAddresses,
				ExcludeSpendingGaps: &h.ExcludeSpendingGaps,
				ClusterTypes:        h.ClusterTypes,
				ClusterCount:        h.ClusterCount,
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
					// find corresponding address UID and set it connected to this transaction
					if addressUID, ok := clusterToAddress[cluster.UID]; ok {
						children[addressUID] = true
					}
				}
			}
		}

		for _, inputs := range queryTx.Inputs {
			for _, outputTx := range inputs.OutputTransactions {
				children[outputTx.UID] = true
			}

			for _, address := range inputs.Addresses {
				for _, cluster := range address.Clusters {
					// find corresponding address UID and set it connected to this transaction
					if addressUID, ok := clusterToAddress[cluster.UID]; ok {
						children[addressUID] = true
					}
				}
			}
		}

		connectedTransactions[queryTx.UID] = NodeConnections{
			UID:      queryTx.UID,
			Children: cliutil.GetMapKeys(children),
		}
	}

	transactions = make([]NodeConnections, 0, len(connectedTransactions))
	for _, v := range connectedTransactions {
		transactions = append(transactions, v)
	}

	for _, cluster := range r.ClusterClusters {
		thisClusterAddressUID, ok := clusterToAddress[cluster.UID]
		if !ok {
			continue
		}

		clusterUIDs := map[string]bool{}
		for _, address := range cluster.Addresses {
			for _, output := range address.Outputs {
				for _, outputCluster := range output.OutputClusters {
					// find corresponding address UID and set it connected to this transaction
					if addressUID, ok := clusterToAddress[outputCluster.UID]; ok {
						clusterUIDs[addressUID] = true
					}
				}
				for _, inputCluster := range output.InputClusters {
					// find corresponding address UID and set it connected to this transaction
					if addressUID, ok := clusterToAddress[inputCluster.UID]; ok {
						clusterUIDs[addressUID] = true
					}
				}
			}
		}
		clusters = append(clusters, NodeConnections{UID: thisClusterAddressUID, Children: cliutil.GetMapKeys(clusterUIDs)})
	}

	return
}

// SearchForNode returns the uid which matches to the given query. In case the query is an address
// which is connected to clusters, they are returned instead.
func SearchForNode(c external.Database, nodeQuery string, userUID string) (node *GraphNode, err error) {
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
							~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
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
		node = &GraphNode{UID: tx.UID, Type: "transaction", TransactionHash: nodeQuery, PrivacyType: tx.PrivacyType}
		return
	}

	if len(r.Address) > 0 {
		addr := r.Address[0]
		if len(addr.Clusters) != 1 {
			return nil, cliutil.NewStackErrorStr("address has no cluster attached")
		}

		node = &GraphNode{UID: addr.UID, Type: "cluster", AddressHash: nodeQuery, ClusterType: addr.Clusters[0].Type}
		return
	}

	return nil, nil
}
