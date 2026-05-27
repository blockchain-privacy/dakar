// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Time ────────────────────────────────────►
//
// ┌──────┐
// │Origin├──┐  O:2
// └──────┘  │ ┌──────┐
//           ├─┤Mixing├─┐   O:3  C
// ┌──────┐  │ └──────┘ │  ┌──────┐
// │Origin├──┘          ├──┤Mixing├─┐
// └──────┘     O:1     │  └──────┘ │
//             ┌──────┐ │           │
// ┌──────┐  ┌─┤Mixing├─┘   O:1     │
// │Origin├──┤ └──────┘    ┌──────┐ │
// └──────┘  │          ┌──┤Mixing├─┤
//           │  O:1     │  └──────┘ │
//           │ ┌──────┐ │           │
//           └─┤Mixing├─┘           │
// ┌──────┐    └──────┘             │
// │Origin├──┐                      │
// └──────┘  │                      │
//           │  O:2         O:2     │
// ┌──────┐  │ ┌──────┐    ┌──────┐ │
// │Origin├──┴─┤Mixing├────┤Mixing├─┤  O:6  C
// └──────┘    └──────┘    └──────┘ │ ┌──────┐
//              O:1         O:1     ├─┤Mixing│
// ┌──────┐    ┌──────┐    ┌──────┐ │ └──────┘
// │Origin├────┤Mixing├────┤Mixing├─┘
// └──────┘    └──────┘    └──────┘

// GetConnectedPrivacyTransactions gets the first numNodes classified transactions including their input transaction
// from the database.
func GetConnectedPrivacyTransactions(ctx context.Context, c external.Database, numNodes int, afterNode string,
	transactionType string) ([]ConnectedNode, error) {
	const query = `query Q($type:string,$first:int,$after:string){
				q(func: eq(Transaction.type,$type),first:$first,after:$after){
					uid
					Transaction.type
					block:~transactions{
						ts
					}
					i:tx_inputs{
						~addr_outputs{
							uid
						}
						~tx_outputs{
							uid
						}
					}
				}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$type": transactionType,
		"$first": strconv.Itoa(numNodes), "$after": afterNode})
	if err != nil {
		return nil, err
	}

	var r struct {
		Q []ConnectedNodeRequest `json:"q"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	connectedNodes := make([]ConnectedNode, len(r.Q))

	for i, connectedNode := range r.Q {
		node, conversionErr := connectedNode.toConnectedNode()
		if conversionErr != nil {
			return nil, conversionErr
		}

		connectedNodes[i] = *node
	}

	return connectedNodes, nil
}

// GetPrivacyTransactions gets the numNodes maxTx classified transactions from the database.
func GetPrivacyTransactions(ctx context.Context, c external.Database,
	numNodes int, afterNode string, transactionType string) ([]Node, error) {
	const query = `query Q($type:string,$first:int,$after:string){
				q(func: eq(Transaction.type,$type),first:$first,after:$after){
					uid
					Transaction.type
					block:~transactions{
						ts
					}
				}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$type": transactionType,
		"$first": strconv.Itoa(numNodes), "$after": afterNode})
	if err != nil {
		return nil, err
	}

	var r struct {
		Q []Node `json:"q"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Q, nil
}

// GetPrivacyTransactionsWithHash gets the numNodes maxTx classified transactions from the database.
// Includes the transaction hashes and input amounts connected .
func GetPrivacyTransactionsWithHash(ctx context.Context, c external.Database,
	numNodes int, afterNode string, transactionType string, connectionTransactionType string) ([]NodeWithHash, error) {
	const query = `query Q($type:string,$connectionType:string,$first:int,$after:string){
				q(func: eq(Transaction.type,$type),first:$first,after:$after){
					uid
					txhash
					Transaction.type
					block:~transactions{
						ts
					}
					tx_inputs@cascade@normalize{
						amount:amount
						~tx_outputs@filter(eq(Transaction.type,$connectionType)){
							block:~transactions{
								ts:ts
							}
						}
					}
					oc:count(tx_outputs)
				}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{
		"$type": transactionType, "$connectionType": connectionTransactionType,
		"$first": strconv.Itoa(numNodes), "$after": afterNode})
	if err != nil {
		return nil, err
	}

	var r struct {
		Q []NodeWithHash `json:"q"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Q, nil
}

// GetDashTransactionTypeCount gets the number of transaction per dash transaction type
func GetDashTransactionTypeCount(ctx context.Context, c external.Database) (mixingCount int,
	originCount int, ccCount int, cpCount int,
	destinationCount int, err error) {
	const query = `{
				mixing(func: eq(Transaction.type,"` + constants.TypeDashMixing + `")){
					count(uid)
				}

				origin(func: eq(Transaction.type,"` + constants.TypeDashOrigin + `")){
					count(uid)
				}

				destination(func: eq(Transaction.type,"` + constants.TypeDashDestination + `")){
					count(uid)
				}

				cc(func: eq(Transaction.type,"` + constants.TypeDashCC + `")){
					count(uid)
				}

				cp(func: eq(Transaction.type,"` + constants.TypeDashCP + `")){
					count(uid)
				}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return
	}

	var r struct {
		Mixing []struct {
			Count int `json:"count,omitempty"`
		} `json:"mixing,omitempty"`

		Origin []struct {
			Count int `json:"count,omitempty"`
		} `json:"origin,omitempty"`

		Destination []struct {
			Count int `json:"count,omitempty"`
		} `json:"destination,omitempty"`

		CC []struct {
			Count int `json:"count,omitempty"`
		} `json:"cc,omitempty"`
		CP []struct {
			Count int `json:"count,omitempty"`
		} `json:"cp,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Mixing) != 1 || len(r.Origin) != 1 || len(r.Destination) != 1 || len(r.CC) != 1 || len(r.CP) != 1 {
		err = serror.FromStr("invalid response from database")
		return
	}

	mixingCount = r.Mixing[0].Count
	originCount = r.Origin[0].Count
	destinationCount = r.Destination[0].Count
	ccCount = r.CC[0].Count
	cpCount = r.CP[0].Count

	return
}

// GetBTCTransactionTypeCount gets the number of transaction per bitcoin transaction type
func GetBTCTransactionTypeCount(ctx context.Context, c external.Database) (wasabi2MixingCount int,
	wasabi2OriginCount int, wasabi2DestinationCount int, whirlpoolMixingCount int,
	whirlpoolOriginCount int, whirlpoolDestinationCount int, err error) {
	const query = `{
				wasabi2Origin(func: eq(Transaction.type,"` + constants.TypeWasabi2Origin + `")){
					count(uid)
				}

				wasabi2Mixing(func: eq(Transaction.type,"` + constants.TypeWasabi2Mixing + `")){
					count(uid)
				}

				wasabi2Destination(func: eq(Transaction.type,"` + constants.TypeWasabi2Destination + `")){
					count(uid)
				}

				whirlpoolOrigin(func: eq(Transaction.type,"` + constants.TypeWhirlpoolOrigin + `")){
					count(uid)
				}

				whirlpoolMixing(func: eq(Transaction.type,"` + constants.TypeWhirlpoolMixing + `")){
					count(uid)
				}

				whirlpoolDestination(func: eq(Transaction.type,"` + constants.TypeWhirlpoolDestination + `")){
					count(uid)
				}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return
	}

	var r struct {
		Wasabi2Origin []struct {
			Count int `json:"count,omitempty"`
		} `json:"wasabi2Origin,omitempty"`
		Wasabi2Mixing []struct {
			Count int `json:"count,omitempty"`
		} `json:"wasabi2Mixing,omitempty"`
		Wasabi2Destination []struct {
			Count int `json:"count,omitempty"`
		} `json:"wasabi2Destination,omitempty"`
		WhirlpoolOrigin []struct {
			Count int `json:"count,omitempty"`
		} `json:"whirlpoolOrigin,omitempty"`
		WhirlpoolMixing []struct {
			Count int `json:"count,omitempty"`
		} `json:"whirlpoolMixing,omitempty"`
		WhirlpoolDestination []struct {
			Count int `json:"count,omitempty"`
		} `json:"whirlpoolDestination,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Wasabi2Mixing) != 1 || len(r.Wasabi2Origin) != 1 || len(r.Wasabi2Destination) != 1 ||
		len(r.WhirlpoolOrigin) != 1 || len(r.WhirlpoolMixing) != 1 || len(r.WhirlpoolDestination) != 1 {
		err = serror.FromStr("invalid response from database")
		return
	}

	wasabi2MixingCount = r.Wasabi2Mixing[0].Count
	wasabi2OriginCount = r.Wasabi2Origin[0].Count
	wasabi2DestinationCount = r.Wasabi2Destination[0].Count
	whirlpoolMixingCount = r.WhirlpoolMixing[0].Count
	whirlpoolOriginCount = r.WhirlpoolOrigin[0].Count
	whirlpoolDestinationCount = r.WhirlpoolDestination[0].Count

	return
}

// GetPrivacyTransactionsByBlock gets all destination transactions, mixing transactions and
// their connected transactions of the given blockHeight
func GetPrivacyTransactionsByBlock(ctx context.Context, c external.Database,
	blockHeight int64) ([]ConnectedNode, []Node, error) {
	const query = `query Q($bid: string) {
				b as var(func: eq(id,$bid))
				var(func: uid(b)){
					txs as transactions
				}
				# get mixing transactions
				mx as var(func: uid(txs))@filter(eq(Transaction.type,` + constants.AllMixingTypes + `)){
					tx_inputs{
						mxi as ~tx_outputs
					}
				}
				# get destination transactions
				dst as var(func: uid(txs))@filter(eq(Transaction.type,` + constants.AllDestinationTypes + `)){
					tx_inputs{
						dsti as ~tx_outputs
					}
				}
				
				connected(func: uid(mx,dst)){
					uid
					Transaction.type
					block:~transactions{
						ts
					}
					i:tx_inputs{
						~tx_outputs{
							uid:uid
						}
						~addr_outputs{
							uid:uid
						}
					}
				}

				single(func: uid(mxi,dsti)){
					uid
					Transaction.type
					block:~transactions{
						ts
					}
				}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query,
		map[string]string{"$bid": strconv.FormatInt(blockHeight, 10)})
	if err != nil {
		return nil, nil, err
	}

	var r struct {
		Connected []ConnectedNodeRequest `json:"connected"`
		Single    []Node                 `json:"single"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, nil, serror.New(err)
	}

	connectedNodes := make([]ConnectedNode, len(r.Connected))

	for i, connectedNode := range r.Connected {
		node, conversionErr := connectedNode.toConnectedNode()
		if conversionErr != nil {
			return nil, nil, conversionErr
		}

		connectedNodes[i] = *node
	}

	return connectedNodes, r.Single, nil
}

// GetTransactionTypeData returns timestamps when the transactions of the specified type occur.
// If the type is empty then all transaction types are considered.
func GetTransactionTypeData(ctx context.Context, c external.Database,
	transactionType string) (ts []time.Time, counts []int, err error) {
	query := `query Q($txType:string){
				q(func:eq(Transaction.type,$txType))@normalize{
					outputCount:count(tx_outputs)
					~transactions{
						ts:ts
					}
			  	}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query,
		map[string]string{"$txType": transactionType})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Query []struct {
			Timestamp   time.Time `json:"ts,omitempty"`
			OutputCount int       `json:"outputCount,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	ts = make([]time.Time, len(r.Query))
	counts = make([]int, len(r.Query))
	for i, q := range r.Query {
		ts[i] = q.Timestamp.UTC()
		counts[i] = q.OutputCount
	}

	return
}

// GetForwardLookupTransactions traverses forward in the transaction graph, starting with the transaction
// specified by startTxHash. This function is used to generate test data.
func GetForwardLookupTransactions(ctx context.Context, c external.Database, startTxHash string) (blocks []db.Block,
	addresses []db.Address, transactions []db.Transaction, err error) {
	const query = `query Q($txhash: string) {
				var(func: eq(txhash, $txhash))@recurse(depth:3){
					tx_outputs
					~tx_inputs@filter(has(Transaction.type))
					pt as txhash
				}

				# get input transactions of all transactions
				var(func: uid(pt)){
					tx_inputs {
						it as ~tx_outputs
					}
				}

				var(func: uid(pt)) {
					b as ~transactions
					i as tx_inputs
					o as tx_outputs
				}

				var(func: uid(o,i)){
					a as ~addr_outputs
				}

				shallow_txs(func: uid(it))@filter(not uid(pt)){
					uid
					tx_outputs{
						uid
					}
				}
				
				addresses(func: uid(a)){
					uid
					addresshash
					dgraph.type
					addr_outputs@filter(uid(o,i)){
						uid
					}
				}

				blocks(func: uid(b)){
					uid
					id
					ts
					blockhash
					dgraph.type
					prevblock {
						uid
						blockhash
						dgraph.type
					}
					transactions@filter(uid(pt)){
						uid
						txhash
						Transaction.type
						fee
						dgraph.type
						tx_outputs {
							...fOutput
						}
						tx_inputs {
							...fOutput
						}
					}
				}
			  }

				fragment fOutput {
					uid
					amount
					inputindex
					outputindex
					iscoinbase
					dgraph.type
				}`

	resp, err := db.QueryVarWithRetry(ctx, c, query,
		map[string]string{"$txhash": startTxHash})
	if err != nil {
		return
	}
	var r struct {
		Blocks              []db.Block       `json:"blocks,omitempty"`
		Addresses           []db.Address     `json:"addresses,omitempty"`
		ShallowTransactions []db.Transaction `json:"shallow_txs,omitempty"`
	}
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	for x := range r.Blocks {
		r.Blocks[x].UID = "_:" + r.Blocks[x].UID
		r.Blocks[x].PrevBlock.UID = "_:" + r.Blocks[x].PrevBlock.UID

		for i := range r.Blocks[x].Transactions {
			r.Blocks[x].Transactions[i].UID = "_:" + r.Blocks[x].Transactions[i].UID

			for y := range r.Blocks[x].Transactions[i].Outputs {
				r.Blocks[x].Transactions[i].Outputs[y].UID = "_:" + r.Blocks[x].Transactions[i].Outputs[y].UID
			}

			for y := range r.Blocks[x].Transactions[i].Inputs {
				r.Blocks[x].Transactions[i].Inputs[y].UID = "_:" + r.Blocks[x].Transactions[i].Inputs[y].UID
			}
		}
	}

	for i := range r.Addresses {
		r.Addresses[i].UID = "_:" + r.Addresses[i].UID

		for y := range r.Addresses[i].Outputs {
			r.Addresses[i].Outputs[y].UID = "_:" + r.Addresses[i].Outputs[y].UID
		}
	}

	for i := range r.ShallowTransactions {
		r.ShallowTransactions[i].UID = "_:" + r.ShallowTransactions[i].UID

		for y := range r.ShallowTransactions[i].Outputs {
			r.ShallowTransactions[i].Outputs[y].UID = "_:" + r.ShallowTransactions[i].Outputs[y].UID
		}
	}

	blocks = r.Blocks
	addresses = r.Addresses
	transactions = r.ShallowTransactions

	return
}

type SpenderTransaction struct {
	// Either Transaction or ClusterUID is set
	Transaction  db.Transaction
	ClusterUID   string
	ClusterSize  int
	Destinations []db.Transaction
}

// GetDestinationTransactionClusterSpenders returns all destination transactions which send funds to the same cluser
func GetDestinationTransactionClusterSpenders(ctx context.Context, c external.Database, transactionType string) (
	transactions []SpenderTransaction, globalDestinationCount int, includedDestinationCount int,
	globalClusterCount int, includedClusterCount int, err error) {
	query := `{
		destinations as var(func: eq(Transaction.type,"` + transactionType + `"))@cascade{
			~transactions@filter(gt(ts,"2018-01-01T00:00:00"))
		}
		
		q(func: uid(destinations)){
			uid
			txhash
			tx_outputs@normalize{
				~addr_outputs {
					~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
						clusterUID:uid
						clusterCount:Cluster.addressCount
					}
				}
			}
		}
	}`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return
	}
	var r struct {
		Transactions []struct {
			UID             string `json:"uid,omitempty"`
			TransactionHash string `json:"txhash,omitempty"`
			Clusters        []struct {
				UID          string `json:"clusterUID,omitempty"`
				ClusterCount int    `json:"clusterCount,omitempty"`
			} `json:"tx_outputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	clusterToDestinationTransactions := map[string]map[string]bool{}
	uidToTx := make(map[string]db.Transaction, len(r.Transactions))
	allClusters := map[string]bool{}
	includedClusters := map[string]bool{}
	includedTransactions := map[string]bool{}

	for _, tx := range r.Transactions {
		uidToTx[tx.UID] = db.Transaction{
			UID:  tx.UID,
			Hash: tx.TransactionHash,
		}

		for _, cluster := range tx.Clusters {
			allClusters[cluster.UID] = true
			if cluster.ClusterCount > 1000 {
				continue
			}

			if len(clusterToDestinationTransactions[cluster.UID]) == 0 {
				clusterToDestinationTransactions[cluster.UID] = make(map[string]bool)
			}

			clusterToDestinationTransactions[cluster.UID][tx.UID] = true
		}
	}

	for clusterUID, txUIDs := range clusterToDestinationTransactions {
		if len(txUIDs) < 2 {
			continue
		}
		includedClusters[clusterUID] = true
		destinations := make([]db.Transaction, len(txUIDs))
		i := 0
		for txUID := range txUIDs {
			includedTransactions[txUID] = true
			destinations[i] = uidToTx[txUID]
			i++
		}

		transactions = append(transactions, SpenderTransaction{
			ClusterUID:   clusterUID,
			Destinations: destinations,
		})
	}

	globalDestinationCount = len(r.Transactions)
	includedDestinationCount = len(includedTransactions)
	globalClusterCount = len(allClusters)
	includedClusterCount = len(includedClusters)

	return
}

// GetTransactionCountPerCluster returns the number of transactions this cluster has created
func GetTransactionCountPerCluster(ctx context.Context, c external.Database, clusterUID string) (int, int, error) {
	const query = `query Q($uid:string){
					var(func: uid($uid)){
						Cluster.addresses {
							addr_outputs{
								i as ~tx_inputs
								o as ~tx_outputs
							}
						}
					}
					
					q(func: uid(i)){
						count(uid)
					}

					x(func: uid(o)){
						count(uid)
					}
				}`

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$uid": clusterUID})
	if err != nil {
		return 0, 0, err
	}
	var r struct {
		Inputs []struct {
			Count int `json:"count,omitempty"`
		} `json:"q,omitempty"`
		Ouptuts []struct {
			Count int `json:"count,omitempty"`
		} `json:"x,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return 0, 0, serror.New(err)
	}

	if len(r.Inputs) != 1 || len(r.Ouptuts) != 1 {
		return 0, 0, serror.FromStr("invalid result")
	}

	return r.Inputs[0].Count, r.Ouptuts[0].Count, nil
}

// GetAllFMIClusters returns the uids of all FMI clusters
func GetAllFMIClusters(ctx context.Context, c external.Database) (uids []string, err error) {
	const query = `{
		q(func: type(Cluster))@filter(eq(Cluster.type, "fmi")){
			uid
		}
	}`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return
	}
	var r struct {
		Clusters []db.UIDNode `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	for _, a := range r.Clusters {
		uids = append(uids, a.UID)
	}
	return
}

// GetShortestTransactionPathAnyDirection returns the transactions of the shortest path between two transactions.
// anyDirection determines the search direction of the shortest transaction path query. Maximum depth is set to 20.
// True: Both inputs and outputs are traversed
// False: Only inputs are traversed
// withTransactionTypes determines if classified transactions should be considered when doing the shortest path lookup
func GetShortestTransactionPathAnyDirection(ctx context.Context, c external.Database, txFrom string, txTo string,
	withTransactionTypes bool, anyDirection bool) ([]db.FrontendTransaction, error) {
	/* Full query
	query Q($txFrom:string, $txTo:string){
					f as var(func: eq(txhash,$txFrom))
					t as var(func: eq(txhash,$txTo))
					path as shortest(from: uid(f), to: uid(t), depth: 20){
						tx_inputs
						~tx_outputs@filter(NOT has(Transaction.type)) tx_outputs ~tx_inputs@filter(NOT has(Transaction.type)) }
					path(func: uid(path))@normalize{
						txhash:txhash
						txtype:Transaction.type
						~transactions{
							bid:id
							bts:ts
							bhash:blockhash
						}
					}
				  }
	*/

	typeFlag := " " // spaces are needed

	if !withTransactionTypes {
		typeFlag = "@filter(NOT has(Transaction.type)) " // spaces are needed
	}

	var anyDirectionFlag string

	if anyDirection {
		anyDirectionFlag = "tx_outputs ~tx_inputs" + typeFlag
	}

	query := `query Q($txFrom:string, $txTo:string){
				f as var(func: eq(txhash,$txFrom))
				t as var(func: eq(txhash,$txTo))
				path as shortest(from: uid(f), to: uid(t), depth: 20){
					tx_inputs
					~tx_outputs` + typeFlag + anyDirectionFlag + `}
				path(func: uid(path))@normalize{
					txhash:txhash
					txtype:Transaction.type
					~transactions{
						bid:id
						bts:ts
						bhash:blockhash
					}
				}
			  }`

	resp, err := c.Query(ctx, query, map[string]string{"$txFrom": txFrom, "$txTo": txTo})
	if err != nil {
		// no type or var available for edge limit error, so we can only do a string comparision
		if isDeadlineExceeded(err) || strings.Contains(err.Error(), "Exceeded query edge limit") {
			return nil, nil
		}

		return nil, serror.New(err)
	}

	// json struct
	var r struct {
		Transactions []db.FrontendTransaction `json:"path,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Transactions, nil
}

// check if deadline was exceeded natively or via grpc
func isDeadlineExceeded(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return status.Code(err) == codes.DeadlineExceeded
}

// GetOutputCountsPerAddress returns the number of outputs controlled by each address,
// for all addresses part of the transactions specified by transactionType
func GetOutputCountsPerAddress(ctx context.Context, c external.Database, transactionType string) ([]AddressOutputCount, error) {
	query := `{
		t as var(func: eq(Transaction.type, "` + transactionType + `"))@cascade{
			~transactions@filter(gt(ts,"2018-01-01T00:00:00"))
		}
		
		var(func: uid(t)){
			tx_inputs {
				ia as ~addr_outputs
			}
		
			tx_outputs {
				oa as ~addr_outputs
			}
		}
		
		q(func: uid(ia,oa)) {
			addresshash
			outputCount:count(addr_outputs)
		}
	}`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return nil, err
	}
	var r struct {
		Counts []AddressOutputCount `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Counts, nil
}

// GetCollateralPaymentTimestamps returns for all collateral payment transations its timestamp and input timestamp
func GetCollateralPaymentTimestamps(ctx context.Context, c external.Database) ([]CollateralPaymentTimestamps, error) {
	query := `{
		t as var(func: eq(Transaction.type, "` + constants.TypeDashCP + `"))@cascade{
			~transactions@filter(gt(ts,"2018-01-01T00:00:00"))
		}
		
		q(func: uid(t))@normalize{
			txhash:txhash

			~transactions{
				ts:ts
			}
		
			tx_inputs {
				~tx_outputs{
					~transactions{
						input_ts:ts
					}
				}
			}
		}
	}`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return nil, err
	}
	var r struct {
		Counts []CollateralPaymentTimestamps `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Counts, nil
}
