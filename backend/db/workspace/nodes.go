package workspace

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"encoding/json"
	"slices"
	"time"
)

// GetWorkspaceConnections returns all connections between the given UIDs, and all connected heuristics
func GetWorkspaceConnections(c external.Database, uids []string, userUID string, workspaceUID string) (
	connections []NodeConnections, heuristicNodes []Node, clusterHeight int64, err error) {
	// one uid is still okay, because it could a be destination transaction with connected heuristics
	if len(uids) == 0 {
		err = cliutil.NewStackError(db.ErrEmptyRequestArgument)
		return
	}

	// todo: in block 'transactions' only select first input when searching for clusters (for performance)
	const query = `query Q($uids:string,$userUID:string,$workspaceUID:string){
					# get cluster height
					cluster_height(func: type(CFMIStatus)){
						lastclusteredid
					}

					# input uids
					uids as var(func: uid($uids))
					
					# transaction uids
					t as var(func: uid(uids))@filter(has(txhash))
					
					# heuristic uids
					var(func: uid($userUID)){
						User.workspaces@filter(uid($workspaceUID)){
							h as Workspace.heuristics
						}
					}

					heuristic_clusters(func: uid(h)){
						uid
						Heuristic.clusters{
							HeuristicCluster.results{
								HeuristicResult.destinations{
									...fGetHeuristicCluster
								}
								HeuristicResult.origin{
									...fGetHeuristicCluster
								}
							}
						}
					}

					# find fmi cluster for each address
					address_cluster(func: uid(uids))@filter(has(addresshash)){
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
									tx_inputs(first:1){
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

					heuristics(func: uid(h)){
						uid
						transaction:Heuristic.transaction{
							uid
						}
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
				
				fragment fGetCluster {
					~addr_outputs{
						~Cluster.addresses@filter(uid(c)){
							uid:uid
						}
					}
				}
				
				# only select the transaction creator
				fragment fGetHeuristicCluster {
					tx_inputs(first:1){
						...fGetCluster
					}
				}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{
		"$uids": db.CreateCommaArray(uids), "$userUID": userUID, "$workspaceUID": workspaceUID})
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

	transactions, clusters, heuristicNodes, clusterHeight, err := parseConnectionResult(r)
	if err != nil {
		return
	}

	connections = append(transactions, clusters...)

	return
}

// parseConnectionResult parses the result of a connection request and returns the resulting connections
//
//nolint:gocyclo
func parseConnectionResult(r connectionRequest) (transactions []NodeConnections, clusters []NodeConnections,
	heuristics []Node, clusterHeight int64, err error) {
	if len(r.ClusterHeight) != 1 {
		err = cliutil.NewStackErrorf("invalid number of cluster information: %d", len(r.ClusterHeight))
		return
	}

	if r.ClusterHeight[0].LastClusteredID == nil {
		err = cliutil.NewStackErrorStr("null pointer received for last clustered ID")
		return
	}

	clusterHeight = *r.ClusterHeight[0].LastClusteredID

	// clusterToAddress contains the mapping of flat multi-input clusters to their addresses.
	// This map is used to replace the uid of clusters with the uid of addresses.
	// This is done because we ultimatly want to store the address uids, not the cluster uids as they are not static.
	clusterToAddress := map[string]string{}
	for _, address := range r.AddressClusters {
		if len(address.Cluster) != 1 {
			err = cliutil.NewStackErrorf("address not attached to FMI-cluster: %s", address)
			return
		}

		clusterToAddress[address.Cluster[0].UID] = address.UID
	}

	heuristicToClusters := map[string]map[string]bool{}
	for _, heuristic := range r.HeuristicClusters {
		heuristicClusters := map[string]bool{}
		for _, heuristicCluster := range heuristic.Clusters {
			for _, result := range heuristicCluster.Results {
				for _, tx := range append(result.Destinations, result.Origin) {
					for _, input := range tx.Inputs {
						for _, address := range input.Addresses {
							for _, cluster := range address.Clusters {
								// find corresponding address UID and set it connected to this transaction
								if addressUID, ok := clusterToAddress[cluster.UID]; ok {
									heuristicClusters[addressUID] = true
								}
							}
						}
					}
				}
			}
		}

		heuristicToClusters[heuristic.UID] = heuristicClusters
	}

	// txToHeuristic contains the mapping of transaction to its directly connected heuristics (root heuristics).
	// This map is used to add the contained heuristic uids as children to their corresponding transaction.
	txToHeuristic := map[string][]string{}
	for _, h := range r.Heuristics {
		// no parent -> root heuristic
		if len(h.ParentHeuristic) == 0 {
			txToHeuristic[h.Transaction.UID] = append(txToHeuristic[h.Transaction.UID], h.UID)
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

		// add cluster reachable from this heuristic as children
		for heuristicClusters := range heuristicToClusters[h.UID] {
			children = append(children, heuristicClusters)
		}

		heuristics = append(heuristics, Node{
			UID:                 h.UID,
			Type:                "heuristic",
			Children:            children,
			HeuristicType:       h.Type,
			Parameter:           h.Parameter,
			ExcludeAddresses:    &h.ExcludeAddresses,    // #nosec G601, false positive as of go1.22
			ExcludeSpendingGaps: &h.ExcludeSpendingGaps, // #nosec G601, false positive as of go1.22
			ClusterTypes:        h.ClusterTypes,
			ClusterCount:        h.ClusterCount,
			Timestamp:           h.Timestamp,
		})
	}

	connectedTransactions := map[string]NodeConnectionsMap{}
	connectedClusters := map[string]NodeConnectionsMap{}
	for _, queryTx := range r.Transactions {
		ct, ok := connectedTransactions[queryTx.UID]
		if !ok {
			ct = NodeConnectionsMap{UID: queryTx.UID, children: map[string]bool{}}
		}

		// add root heuristics to transaction if available
		if rootHeuristics, ok := txToHeuristic[queryTx.UID]; ok {
			for _, h := range rootHeuristics {
				ct.children[h] = true
			}
		}

		for _, output := range queryTx.Outputs {
			for _, inputTx := range output.InputTransactions {
				ct.children[inputTx.UID] = true
			}

			for _, address := range output.Addresses {
				for _, cluster := range address.Clusters {
					// find corresponding address UID and set it connected to this transaction
					if addressUID, ok := clusterToAddress[cluster.UID]; ok {
						ct.children[addressUID] = true
					}
				}
			}
		}

		for _, inputs := range queryTx.Inputs {
			for _, outputTx := range inputs.OutputTransactions {
				// add this transaction as child of the input transaction
				parentTransaction, ok := connectedTransactions[outputTx.UID]
				if !ok {
					parentTransaction = NodeConnectionsMap{UID: outputTx.UID, children: map[string]bool{}}
				}

				parentTransaction.children[queryTx.UID] = true
				connectedTransactions[outputTx.UID] = parentTransaction
			}

			for _, address := range inputs.Addresses {
				for _, cluster := range address.Clusters {
					// find corresponding address UID and set it connected to this transaction
					if addressUID, ok := clusterToAddress[cluster.UID]; ok {
						// add this transaction as child of the connected cluster
						parentCluster, ok := connectedClusters[addressUID]
						if !ok {
							parentCluster = NodeConnectionsMap{UID: addressUID, children: map[string]bool{}}
						}

						parentCluster.children[queryTx.UID] = true
						connectedClusters[addressUID] = parentCluster
					}
				}
			}
		}

		connectedTransactions[queryTx.UID] = ct
	}

	i := 0
	transactions = make([]NodeConnections, len(connectedTransactions))
	for _, ct := range connectedTransactions {
		transactions[i] = NodeConnections{
			UID:      ct.UID,
			Children: cliutil.GetMapKeys(ct.children),
		}
		i++
	}

	for _, cluster := range r.ClusterClusters {
		thisClusterAddressUID, ok := clusterToAddress[cluster.UID]
		if !ok {
			continue
		}

		cc, ok := connectedClusters[thisClusterAddressUID]
		if !ok {
			cc = NodeConnectionsMap{UID: thisClusterAddressUID, children: map[string]bool{}}
		}

		for _, address := range cluster.Addresses {
			for _, output := range address.Outputs {
				for _, outputCluster := range output.OutputClusters {
					// find corresponding address UID and set it connected to this transaction
					if addressUID, ok := clusterToAddress[outputCluster.UID]; ok {
						// add cluster as child of current cluster
						cc.children[addressUID] = true
					}
				}
				for _, inputCluster := range output.InputClusters {
					// find corresponding address UID and set it connected to this transaction
					if addressUID, ok := clusterToAddress[inputCluster.UID]; ok {
						// add current cluster as child of cluster
						parentCluster, ok := connectedClusters[addressUID]
						if !ok {
							parentCluster = NodeConnectionsMap{UID: addressUID, children: map[string]bool{}}
						}

						parentCluster.children[thisClusterAddressUID] = true
						connectedClusters[addressUID] = parentCluster
					}
				}
			}
		}

		connectedClusters[thisClusterAddressUID] = cc
	}

	i = 0
	clusters = make([]NodeConnections, len(connectedClusters))
	for _, cc := range connectedClusters {
		clusters[i] = NodeConnections{
			UID:      cc.UID,
			Children: cliutil.GetMapKeys(cc.children),
		}
		i++
	}

	return
}

// GetConnectionClusterToCluster return the transactions which connect two clusters.
// The provided UIDs must be of addresses of the respective clusters.
func GetConnectionClusterToCluster(c external.Database, firstUID string, secondUID string, userUID string, workspaceUID string) (
	transactions []string, err error) {
	const query = `query Q($first:string,$second:string,$userUID:string,$workspaceUID:string){
			# find fmi cluster for first address
			var(func: uid($first))@filter(has(addresshash)){
				uid
				c1 as cluster:~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
					uid
				}
			}
			
			# find fmi cluster for second address
			var(func: uid($second))@filter(has(addresshash)){
				uid
				c2 as cluster:~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
					uid
				}
			}
			
			
			cluster_clusters(func: uid(c1))@ignorereflex{
				Cluster.addresses{
					addr_outputs {
						~tx_inputs@cascade{
							txhash
							tx_outputs{
								~addr_outputs{
									~Cluster.addresses@filter(uid(c2)){
										uid
									}
								}
							}
						}
						~tx_outputs@cascade{
							txhash:txhash
							tx_inputs(first:1){
								~addr_outputs{
									~Cluster.addresses@filter(uid(c2)){
										uid
									}
								}
							}
						}
					}
				}
			}
}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{"$first": firstUID,
		"$second": secondUID, "$userUID": userUID, "$workspaceUID": workspaceUID})
	if err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	// json struct
	// while the query also returns the cluster uids, only the transaction hashes are collected
	var r struct {
		ClusterClusters []struct {
			Addresses []struct {
				Outputs []struct {
					InputClusters []struct {
						TransactionHash string `json:"txhash,omitempty"`
					} `json:"~tx_inputs,omitempty"`
					OutputClusters []struct {
						TransactionHash string `json:"txhash,omitempty"`
					} `json:"~tx_outputs,omitempty"`
				} `json:"addr_outputs,omitempty"`
			} `json:"Cluster.addresses,omitempty"`
		} `json:"cluster_clusters,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	if len(r.ClusterClusters) != 1 {
		err = cliutil.NewStackErrorf("invalid number of clusters returned: %d", len(r.ClusterClusters))
		return
	}
	transactionMap := map[string]bool{}
	for _, addresses := range r.ClusterClusters[0].Addresses {
		for _, outputs := range addresses.Outputs {
			for _, txs := range outputs.InputClusters {
				transactionMap[txs.TransactionHash] = true
			}
			for _, txs := range outputs.OutputClusters {
				transactionMap[txs.TransactionHash] = true
			}
		}
	}

	transactions = cliutil.GetMapKeys(transactionMap)

	return
}

// GetConnectionClusterToHeuristic return the transactions which connects a cluster to an heuristic.
// The provided cluster UID must be of an address of the cluster.
func GetConnectionClusterToHeuristic(c external.Database, clusterUID string, heuristicUID string, userUID string,
	workspaceUID string) (transactions []string, err error) {
	const query = `query Q($cluster:string,$heuristic:string,$userUID:string,$workspaceUID:string){
			# heuristic uids
			var(func: uid($userUID)){
				User.workspaces@filter(uid($workspaceUID)){
					h as Workspace.heuristics@filter(uid($heuristic))
				}
			}
			
			# find fmi cluster for address
			var(func: uid($cluster))@filter(has(addresshash)){
				c as cluster:~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
					uid
				}
			}
			
			heuristic_clusters(func: uid(h)){
				Heuristic.clusters{
					HeuristicCluster.results{
						# todo show only transaction which connects to cluster
						HeuristicResult.destinations@cascade{
							txhash
							tx_inputs(first:1){
								...fGetCluster
							}
						}
						HeuristicResult.origin@cascade{
							txhash
							tx_inputs(first:1){
								...fGetCluster
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{"$cluster": clusterUID,
		"$heuristic": heuristicUID, "$userUID": userUID, "$workspaceUID": workspaceUID})
	if err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	// json struct
	// while the query also returns the cluster uids, only the transaction hashes are collected
	var r struct {
		HeuristicClusters []struct {
			Clusters []struct {
				Results []struct {
					Destinations []struct {
						TransactionHash string `json:"txhash,omitempty"`
					} `json:"HeuristicResult.destinations,omitempty"`
					Origin struct {
						TransactionHash string `json:"txhash,omitempty"`
					} `json:"HeuristicResult.origin,omitempty"`
				} `json:"HeuristicCluster.results,omitempty"`
			} `json:"Heuristic.clusters,omitempty"`
		} `json:"heuristic_clusters,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	if len(r.HeuristicClusters) != 1 {
		err = cliutil.NewStackErrorf("invalid number of heuristic results returned: %d", len(r.HeuristicClusters))
		return
	}

	if len(r.HeuristicClusters[0].Clusters) != 1 {
		err = cliutil.NewStackErrorf("invalid number of cluster results returned: %d", len(r.HeuristicClusters[0].Clusters))
		return
	}

	transactionMap := map[string]bool{}
	for _, results := range r.HeuristicClusters[0].Clusters[0].Results {
		for _, destination := range results.Destinations {
			transactionMap[destination.TransactionHash] = true
		}
		transactionMap[results.Origin.TransactionHash] = true
	}

	transactions = cliutil.GetMapKeys(transactionMap)

	return
}

// SearchForNode returns the uid which matches to the given query. In case the query is an address
// which is connected to clusters, they are returned instead.
func SearchForNode(c external.Database, nodeQuery string, userUID string) (node *Node, err error) {
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
		node = &Node{UID: tx.UID, Type: NodeTypeTransaction, TransactionHash: nodeQuery, PrivacyType: tx.PrivacyType}
		return
	}

	if len(r.Address) > 0 {
		addr := r.Address[0]
		if len(addr.Clusters) != 1 {
			return nil, cliutil.NewStackErrorStr("address has no cluster attached")
		}

		node = &Node{UID: addr.UID, Type: NodeTypeCluster, AddressHash: nodeQuery, ClusterType: addr.Clusters[0].Type}
		return
	}

	return nil, nil
}

// FindDescendantHeuristicUIDs returns the given node uid and all node uids which can
// be found by recursively traversing their children. Only heuristics are considered.
func FindDescendantHeuristicUIDs(nodes map[string]Node, nodeUID string) []string {
	var descendants []string

	n, ok := nodes[nodeUID]
	if !ok || n.Type != NodeTypeHeuristic {
		return descendants
	}

	descendants = append(descendants, n.UID)

	for _, childNode := range n.Children {
		descendants = append(descendants, FindDescendantHeuristicUIDs(nodes, childNode)...)
	}
	return descendants
}

// DeleteNodes returns a new slice which contains nodes which do not have an UID contained in uids
func DeleteNodes(nodes []Node, uids []string) []Node {
	return slices.DeleteFunc(nodes, func(node Node) bool {
		return slices.Contains(uids, node.UID)
	})
}
