package workspace

import (
	"backend/cmd/cliutil"
	"backend/db"
	dbHeuristic "backend/db/analytics/heuristics"
	"backend/external"
	"context"
	"encoding/json"
	"github.com/qrest/gomisc/serror"
	"slices"
	"time"
)

// GetWorkspaceConnections returns all connections between the given UIDs, and all connected heuristics
func GetWorkspaceConnections(c external.Database, uids []string, userUID string, workspaceUID string) (
	connections []NodeConnections, heuristicNodes []Node, selectorNodes []Node, clusterHeight int64, err error) {
	result, err := getWorkspaceConnectionsRaw(c, uids, userUID, workspaceUID)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	transactions, clusters, heuristicNodes, selectorNodes, clusterHeight, err := parseConnectionResult(result)
	if err != nil {
		return
	}

	connections = append(transactions, clusters...)

	return
}

// getWorkspaceConnectionsRaw returns all connections between the given UIDs, and all connected heuristics in unparsed form
func getWorkspaceConnectionsRaw(c external.Database, uids []string, userUID string, workspaceUID string) (
	*connectionRequest, error) {
	// one uid is still okay, because it could a be destination transaction with connected heuristics
	if len(uids) == 0 {
		return nil, serror.New(db.ErrEmptyRequestArgument)
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
					
					# selector uids
					var(func: uid($userUID)){
						User.workspaces@filter(uid($workspaceUID)){
							s as Workspace.selectors
						}
					}

					# split selectors by type
					heuristicSelectors as var(func: uid(s))@filter(eq(Selector.type, ` + TypeHeuristic + `))
					propSelectors as var(func: uid(s))@filter(not eq(Selector.type, ` + TypeHeuristic + `))

					# find fmi cluster for each address
					address_cluster(func: uid(uids))@filter(has(addresshash)){
						uid
						c as cluster:~Cluster.addresses@filter(eq(Cluster.type, "fmi")){uid}
					}

					heuristic_clusters(func: uid(heuristicSelectors)){
						uid
						results: Selector.results{
							HeuristicCluster.results{
								...fTxCluster
							}
						}
					}

					selector_clusters(func: uid(propSelectors)){
						uid
						results: Selector.results{
							...fTxCluster
						}
					}
					
					transactions(func: uid(t)){
						uid
						tx_outputs{
							~tx_inputs@filter(uid(t)){uid}
							...fGetCluster
						}
						tx_inputs {
							~tx_outputs@filter(uid(t)){uid}
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
									...fTxCluster
								}
							}
						}
					}

					selectors(func: uid(propSelectors)){
						uid
						created: Selector.created
						modified: Selector.modified
						type: Selector.type
						status: Selector.status
						parent: Selector.parent {uid}
						children: ~Selector.parent{uid}
						options: Selector.options
						resultCount: count(Selector.results)
						results: Selector.results {uid}
					}

					heuristics(func: uid(heuristicSelectors)){
						uid
						created: Selector.created
						modified: Selector.modified
						type: Selector.type
						status: Selector.status
						parent: Selector.parent {uid}
						children: ~Selector.parent{uid}
						options: Selector.options
						resultCount: count(Selector.results)
						results: Selector.results{
							HeuristicCluster.results@filter(uid(t)){uid}
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
				fragment fTxCluster {
					tx_inputs(first:1){
						...fGetCluster
					}
				}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{
		"$uids": db.CreateCommaArray(uids), "$userUID": userUID, "$workspaceUID": workspaceUID})
	if err != nil {
		return nil, err
	}

	// json struct
	var r connectionRequest

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, serror.New(err)
	}

	return &r, nil
}

// parseConnectionResult parses the result of a connection request and returns the resulting connections
//
//nolint:gocyclo
func parseConnectionResult(r *connectionRequest) (transactions []NodeConnections, clusters []NodeConnections,
	heuristics []Node, selectorNodes []Node, clusterHeight int64, err error) {
	if len(r.ClusterHeight) > 1 {
		err = serror.FromFormat("invalid number of cluster height results: %d", len(r.ClusterHeight))
		return
	}

	if len(r.ClusterHeight) == 0 {
		// cluster height not set yet
		clusterHeight = 0
	} else if r.ClusterHeight[0].LastClusteredID == nil {
		err = serror.FromStr("null pointer received for last clustered ID")
		return
	} else {
		clusterHeight = *r.ClusterHeight[0].LastClusteredID
	}

	// clusterToAddress contains the mapping of flat multi-input clusters to their addresses.
	// This map is used to replace the uid of clusters with the uid of addresses.
	// This is done because we ultimatly want to store the address uids, not the cluster uids as they are not static.
	clusterToAddress := map[string]string{}
	for _, address := range r.AddressClusters {
		if len(address.Cluster) != 1 {
			err = serror.FromFormat("address not attached to FMI-cluster: %s", address)
			return
		}

		clusterToAddress[address.Cluster[0].UID] = address.UID
	}

	heuristicToClusters := map[string]map[string]bool{}
	for _, heuristic := range r.HeuristicClusters {
		heuristicClusters := map[string]bool{}
		for _, heuristicCluster := range heuristic.Clusters {
			for _, result := range heuristicCluster.Results {
				for _, input := range result.Inputs {
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

		heuristicToClusters[heuristic.UID] = heuristicClusters
	}

	// parentToHeuristic contains the mapping of parents to its directly
	// connected heuristics. This map is used to add the contained heuristic
	// uids as children to their corresponding transaction (if its parent is a transaction).
	parentToHeuristic := map[string][]string{}
	for _, h := range r.Heuristics {
		if h.Parent != nil && h.Parent.UID != "" {
			parentToHeuristic[h.Parent.UID] = append(parentToHeuristic[h.Parent.UID], h.UID)
		}

		children := make([]string, len(h.Children))
		for i, c := range h.Children {
			children[i] = c.UID
		}

		// add connections between heuristics and their found origins
		for _, cluster := range h.Clusters {
			for _, result := range cluster.Results {
				children = append(children, result.UID)
			}
		}

		// add cluster reachable from this heuristic as children
		for heuristicClusters := range heuristicToClusters[h.UID] {
			children = append(children, heuristicClusters)
		}

		var opt dbHeuristic.Options
		if err = json.Unmarshal([]byte(h.Options), &opt); err != nil {
			err = serror.NewWithContext(err, "opt", h.Options)
			return
		}

		heuristics = append(heuristics, Node{
			UID:                 h.UID,
			Type:                "selector",
			Children:            children,
			SelectorType:        h.Type,
			SelectorStatus:      h.Status,
			SelectorResultCount: h.ResultCount,
			SelectorCreated:     h.Created,
			SelectorModified:    h.Modified,
			HeuristicOptions:    &opt,
		})
	}

	selectorToClusters := map[string]map[string]bool{}
	for _, s := range r.SelectorClusters {
		selectorCluster := map[string]bool{}
		for _, result := range s.Results {
			for _, input := range result.Inputs {
				for _, address := range input.Addresses {
					for _, cluster := range address.Clusters {
						// find corresponding address UID and set it connected to this transaction
						if addressUID, ok := clusterToAddress[cluster.UID]; ok {
							selectorCluster[addressUID] = true
						}
					}
				}
			}
		}

		selectorToClusters[s.UID] = selectorCluster
	}

	for _, s := range r.Selectors {
		children := make([]string, len(s.Children))
		for i, c := range s.Children {
			children[i] = c.UID
		}

		// add connections between selectors and their results
		for _, result := range s.Results {
			children = append(children, result.UID)
		}

		// add cluster reachable from this heuristic as children
		for selectorClusters := range selectorToClusters[s.UID] {
			children = append(children, selectorClusters)
		}

		var opt Options
		if err = json.Unmarshal([]byte(s.Options), &opt); err != nil {
			err = serror.NewWithContext(err, "opt", s.Options)
			return
		}

		heuristics = append(heuristics, Node{
			UID:                 s.UID,
			Type:                "selector",
			Children:            children,
			SelectorType:        s.Type,
			SelectorStatus:      s.Status,
			SelectorResultCount: s.ResultCount,
			SelectorCreated:     s.Created,
			SelectorModified:    s.Modified,
			SelectorOptions:     &opt,
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
		if rootHeuristics, ok := parentToHeuristic[queryTx.UID]; ok {
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
						// add current cluster as child of cluster
						parentCluster, ok := connectedClusters[addressUID]
						if !ok {
							parentCluster = NodeConnectionsMap{UID: addressUID, children: map[string]bool{}}
						}

						parentCluster.children[thisClusterAddressUID] = true
						connectedClusters[addressUID] = parentCluster
					}
				}
				for _, inputCluster := range output.InputClusters {
					// find corresponding address UID and set it connected to this transaction
					if addressUID, ok := clusterToAddress[inputCluster.UID]; ok {
						// add cluster as child of current cluster
						cc.children[addressUID] = true
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

// GetConnectionClusterToCluster return the transaction UIDs which connect two clusters.
// The provided UIDs must be of addresses of the respective clusters.
func GetConnectionClusterToCluster(ctx context.Context, c external.Database, firstUID string, secondUID string) (
	frontendTransactions []db.AmountTransaction, err error) {
	const query = `query Q($first:string,$second:string){
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
							uid
							tx_outputs{
								~addr_outputs{
									~Cluster.addresses@filter(uid(c2)){
										uid
									}
								}
							}
						}
						~tx_outputs@cascade{
							uid
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

	resp, err := c.Query(ctx, query, map[string]string{"$first": firstUID, "$second": secondUID})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	// while the query also returns the cluster uids, only the transaction hashes are collected
	var r struct {
		ClusterClusters []struct {
			Addresses []struct {
				Outputs []struct {
					InputClusters  []db.UIDNode `json:"~tx_inputs,omitempty"`
					OutputClusters []db.UIDNode `json:"~tx_outputs,omitempty"`
				} `json:"addr_outputs,omitempty"`
			} `json:"Cluster.addresses,omitempty"`
		} `json:"cluster_clusters,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.ClusterClusters) != 1 {
		err = serror.FromFormat("invalid number of clusters returned: %d", len(r.ClusterClusters))
		return
	}
	transactionMap := map[string]bool{}
	for _, addresses := range r.ClusterClusters[0].Addresses {
		for _, outputs := range addresses.Outputs {
			for _, txs := range outputs.InputClusters {
				transactionMap[txs.UID] = true
			}
			for _, txs := range outputs.OutputClusters {
				transactionMap[txs.UID] = true
			}
		}
	}

	if len(transactionMap) > 0 {
		frontendTransactions, err = db.GetFrontendTransactionAmounts(ctx, c, cliutil.GetMapKeys(transactionMap))
		if err != nil {
			return
		}
	}

	return
}

// GetConnectionClusterToHeuristic returns the transaction UIDs which connects a cluster to an heuristic.
// The provided cluster UID must be of a cluster address.
func GetConnectionClusterToHeuristic(ctx context.Context, c external.Database, clusterUID string,
	heuristicUID string, userUID string, workspaceUID string) (frontendTransactions []db.AmountTransaction, err error) {
	const query = `query Q($cluster:string,$heuristic:string,$userUID:string,$workspaceUID:string){
			# heuristic uids
			var(func: uid($userUID)){
				User.workspaces@filter(uid($workspaceUID)){
					h as Workspace.selectors@filter(uid($heuristic))
				}
			}
			
			# find fmi cluster for address
			var(func: uid($cluster))@filter(has(addresshash)){
				c as cluster:~Cluster.addresses@filter(eq(Cluster.type, "fmi"))
			}
			
			heuristic_clusters(func: uid(h)){
				Heuristic.clusters{
					HeuristicCluster.results@cascade{
						uid
						tx_inputs(first:1){
							...fGetCluster
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

	resp, err := c.Query(ctx, query, map[string]string{"$cluster": clusterUID,
		"$heuristic": heuristicUID, "$userUID": userUID, "$workspaceUID": workspaceUID})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	// while the query also returns the cluster uids, only the transaction hashes are collected
	var r struct {
		HeuristicClusters []struct {
			Clusters []struct {
				Results []db.UIDNode `json:"HeuristicCluster.results,omitempty"`
			} `json:"Heuristic.clusters,omitempty"`
		} `json:"heuristic_clusters,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.HeuristicClusters) != 1 {
		err = serror.FromFormat("invalid number of heuristic results returned: %d", len(r.HeuristicClusters))
		return
	}

	if len(r.HeuristicClusters[0].Clusters) != 1 {
		err = serror.FromFormat("invalid number of cluster results returned: %d", len(r.HeuristicClusters[0].Clusters))
		return
	}

	transactionMap := map[string]bool{}
	for _, results := range r.HeuristicClusters[0].Clusters[0].Results {
		transactionMap[results.UID] = true
	}

	if len(transactionMap) > 0 {
		frontendTransactions, err = db.GetFrontendTransactionAmounts(ctx, c, cliutil.GetMapKeys(transactionMap))
		if err != nil {
			return
		}
	}

	return
}

// SearchForNode returns the uid which matches to the given query. In case the query is an address
// which is connected to clusters, they are returned instead.
func SearchForNode(ctx context.Context, c external.Database, nodeQuery string, userUID string) (node *Node, err error) {
	if nodeQuery == "" || userUID == "" {
		err = serror.New(db.ErrEmptyRequestArgument)
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

	resp, err := c.Query(ctx, query, map[string]string{"$query": nodeQuery, "$user": userUID})
	if err != nil {
		err = serror.New(err)
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
		err = serror.New(err)
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
			return nil, serror.FromStr("address has no cluster attached")
		}

		node = &Node{UID: addr.UID, Type: NodeTypeCluster, AddressHash: nodeQuery, ClusterType: addr.Clusters[0].Type}
		return
	}

	return nil, nil
}

// GetConnectionClusterToTransaction returns the given transaction, with each output
// having a flag if it belongs to one of the cluster's addresses.
func GetConnectionClusterToTransaction(ctx context.Context, c external.Database, clusterUID string,
	transactionUID string) (frontendTransactions []db.FrontendTransaction, err error) {
	const query = `query Q($transaction:string,$address:string){
					# find fmi cluster for address
					var(func: uid($address))@filter(has(addresshash)){
						uid
						c as cluster:~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
							uid
						}
					}
				
					q(func: uid($transaction)){
						txhash
						tx_outputs@cascade{
							...fGetCluster
						}
						tx_inputs@cascade{
							...fGetCluster
						}
					}
				}
				
				fragment fGetCluster {
					~addr_outputs{
						addresshash
						~Cluster.addresses@filter(uid(c)){
							uid
						}
					}
				}`

	resp, err := c.Query(ctx, query, map[string]string{"$transaction": transactionUID, "$address": clusterUID})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	// while the query also returns the cluster uids, only the address hashes are collected
	var r struct {
		Transactions []struct {
			TransactionHash string `json:"txhash"`
			Inputs          []struct {
				Adddresses []struct {
					AddressHash string `json:"addresshash,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_inputs,omitempty"`
			Outputs []struct {
				Adddresses []struct {
					AddressHash string `json:"addresshash,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_outputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Transactions) < 1 || len(r.Transactions) > 2 {
		err = serror.FromFormat("invalid number of transactions returned: %d", len(r.Transactions))
		return
	}

	addressMap := map[string]bool{}
	for _, transaction := range r.Transactions {
		for _, outputs := range transaction.Outputs {
			for _, address := range outputs.Adddresses {
				addressMap[address.AddressHash] = true
			}
		}

		for _, inputs := range transaction.Inputs {
			for _, address := range inputs.Adddresses {
				addressMap[address.AddressHash] = true
			}
		}
	}

	if len(addressMap) > 0 && r.Transactions[0].TransactionHash != "" {
		frontendTransactions, err = db.GetFrontendTransaction(ctx, c, r.Transactions[0].TransactionHash)
		if err != nil {
			return
		}
		yes := true
		for i, transaction := range frontendTransactions {
			for y, inputs := range transaction.Inputs {
				if addressMap[inputs.AddressHash] {
					frontendTransactions[i].Inputs[y].Highlight = &yes
				}
			}

			for y, output := range transaction.Outputs {
				if addressMap[output.AddressHash] {
					frontendTransactions[i].Outputs[y].Highlight = &yes
				}
			}
		}
	}

	return
}

// GetConnectionHeuristicToTransaction returns the given transaction, with each output
// having a flag if it belongs to one of the heuristic's clusters addresses.
func GetConnectionHeuristicToTransaction(ctx context.Context, c external.Database, heuristicUID string,
	transactionUID string, userUID string, workspaceUID string) (frontendTransactions []db.FrontendTransaction, err error) {
	const query = `query Q($transaction:string,$heuristic:string,$userUID:string,$workspaceUID:string){
			# heuristic uids
			var(func: uid($userUID)){
				User.workspaces@filter(uid($workspaceUID)){
					h as Workspace.selectors@filter(uid($heuristic))
				}
			}

			t as var(func: uid($transaction))

			# get all cluster of heuristic
			var(func: uid(h)){
				Heuristic.clusters{
					HeuristicCluster.results@filter(uid(t)){
						tx_inputs(first:1){
							~addr_outputs{
								c as ~Cluster.addresses@filter(eq(Cluster.type, "fmi"))
							}
						}
					}
				}
			}

			q(func: uid(t)){
				txhash
				tx_outputs@cascade{
					...fGetCluster
				}
				tx_inputs@cascade{
					...fGetCluster
				}
			}
		}

		fragment fGetCluster {
			~addr_outputs{
				addresshash
				~Cluster.addresses@filter(uid(c)){
					uid
				}
			}
		}`

	resp, err := c.Query(ctx, query, map[string]string{"$transaction": transactionUID,
		"$heuristic": heuristicUID, "$userUID": userUID, "$workspaceUID": workspaceUID})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	// while the query also returns the cluster uids, only the address hashes are collected
	var r struct {
		Transactions []struct {
			TransactionHash string `json:"txhash"`
			Inputs          []struct {
				Adddresses []struct {
					AddressHash string `json:"addresshash,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_inputs,omitempty"`
			Outputs []struct {
				Adddresses []struct {
					AddressHash string `json:"addresshash,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_outputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Transactions) < 1 || len(r.Transactions) > 2 {
		err = serror.FromFormat("invalid number of transactions returned: %d", len(r.Transactions))
		return
	}

	addressMap := map[string]bool{}
	for _, transaction := range r.Transactions {
		for _, outputs := range transaction.Outputs {
			for _, address := range outputs.Adddresses {
				addressMap[address.AddressHash] = true
			}
		}

		for _, inputs := range transaction.Inputs {
			for _, address := range inputs.Adddresses {
				addressMap[address.AddressHash] = true
			}
		}
	}

	if len(addressMap) > 0 && r.Transactions[0].TransactionHash != "" {
		frontendTransactions, err = db.GetFrontendTransaction(ctx, c, r.Transactions[0].TransactionHash)
		if err != nil {
			return
		}
		yes := true
		for i, transaction := range frontendTransactions {
			for y, inputs := range transaction.Inputs {
				if addressMap[inputs.AddressHash] {
					frontendTransactions[i].Inputs[y].Highlight = &yes
				}
			}

			for y, output := range transaction.Outputs {
				if addressMap[output.AddressHash] {
					frontendTransactions[i].Outputs[y].Highlight = &yes
				}
			}
		}
	}

	return
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
