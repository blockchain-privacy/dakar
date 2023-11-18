package analytics

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/external"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
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

// GetConnectedPrivacyTransactions gets the first numNodes privacy transactions including their input transaction
// from the database.
func GetConnectedPrivacyTransactions(c external.Database, numNodes int, offsetNodes int,
	privacyRangeFirst constants.PrivacyType, privacyRangeLast constants.PrivacyType) ([]ConnectedNode, error) {
	query := fmt.Sprintf(`{
				q(func: between(privacytype,`+
		strconv.Itoa(int(privacyRangeFirst))+","+strconv.Itoa(int(privacyRangeLast))+`), first:%d, offset:%d ){
					uid
					privacytype
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
			  }`, numNodes, offsetNodes)

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*2, query)
	if err != nil {
		return nil, err
	}

	var r struct {
		Q []ConnectedNodeRequest `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, cliutil.NewStackError(err)
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

// GetPrivacyTransactions gets the numNodes maxTx privacy transactions from the database.
func GetPrivacyTransactions(c external.Database, numNodes int, offsetNodes int, privacyRangeFirst constants.PrivacyType,
	privacyRangeLast constants.PrivacyType) ([]Node, error) {
	query := fmt.Sprintf(`{
				q(func: between(privacytype,`+
		strconv.Itoa(int(privacyRangeFirst))+","+strconv.Itoa(int(privacyRangeLast))+`), first:%d, offset:%d ){
					uid
					privacytype
					block:~transactions{
						ts
					}
				}
			  }`, numNodes, offsetNodes)

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*2, query)
	if err != nil {
		return nil, err
	}

	var r struct {
		Q []Node `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, cliutil.NewStackError(err)
	}

	return r.Q, nil
}

// GetMixingTransactions gets the first numNodes mixing transactions including their input transactions
// from the database. If maxTx is equal to 0, all mixing transaction are returned.
func GetMixingTransactions(c external.Database, numNodes int, offsetNodes int) ([]ConnectedNode, error) {
	return GetConnectedPrivacyTransactions(c, numNodes, offsetNodes, 0, constants.PrivacyMixingLast)
}

// GetDestinationTransactions gets the first numNodes destination transactions including their input transactions
// from the database. If maxTx is equal to 0, all destination transaction are returned.
func GetDestinationTransactions(c external.Database, numNodes int, offsetNodes int) ([]ConnectedNode, error) {
	return GetConnectedPrivacyTransactions(c, numNodes, offsetNodes, constants.PrivacyDestinationFirst,
		constants.PrivacyDestinationLast)
}

// GetOriginTransactions gets the first numNodes origin transactions from the database.
// If maxTx is equal to 0, all origin transaction are returned.
func GetOriginTransactions(c external.Database, numNodes int, offsetNodes int) ([]Node, error) {
	return GetPrivacyTransactions(c, numNodes, offsetNodes, constants.PrivacyOriginFirst, constants.PrivacyOriginLast)
}

// GetCollateralCreationTransactions gets the numNodes maxTx cc transactions from the database.
// If maxTx is equal to 0, all cc transaction are returned.
func GetCollateralCreationTransactions(c external.Database, numNodes int, offsetNodes int) ([]Node, error) {
	return GetPrivacyTransactions(c, numNodes, offsetNodes, constants.PrivacyCollateralCreationFirst,
		constants.PrivacyCollateralCreationLast)
}

// GetPrivacyTransactionCount gets the number of transaction per privacy type
func GetPrivacyTransactionCount(c external.Database) (mixingCount int, originCount int, ccCount int,
	destinationCount int, err error) {
	const query = `{
				mixing(func: between(privacytype,0,` + constants.StrPrivacyMixingLast + `)){
					count(uid)
				}

				origin(func: between(privacytype,` +
		constants.StrPrivacyOriginFirst + "," + constants.StrPrivacyOriginLast + `)){
					count(uid)
				}

				destination(func: between(privacytype,` +
		constants.StrPrivacyDestinationFirst + "," + constants.StrPrivacyDestinationLast + `)){
					count(uid)
				}

				cc(func: between(privacytype,` +
		constants.StrPrivacyCollateralCreationFirst + "," + constants.StrPrivacyCollateralCreationLast + `)){
					count(uid)
				}
			  }`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*2, query)
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
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	if len(r.Mixing) != 1 || len(r.Origin) != 1 || len(r.Destination) != 1 || len(r.CC) != 1 {
		err = cliutil.NewStackErrorStr("invalid response from database")
		return
	}

	mixingCount = r.Mixing[0].Count
	originCount = r.Origin[0].Count
	destinationCount = r.Destination[0].Count
	ccCount = r.CC[0].Count

	return
}

// GetPrivacyTransactionsByBlock gets all destination transactions, mixing transactions and
// their connected transactions of the given blockHeight
func GetPrivacyTransactionsByBlock(c external.Database, blockHeight uint64) ([]ConnectedNode, []Node, error) {
	const query = `query Q($bid: string) {
				b as var(func: eq(id,$bid))
				var(func: uid(b)){
					txs as transactions
				}
				# get mixing transactions
				mx as var(func: uid(txs))@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast + `)){
					tx_inputs{
						mxi as ~tx_outputs
					}
				}
				# get destination transactions
				dst as var(func: uid(txs))@filter(between(privacytype,` + constants.StrPrivacyDestinationFirst + "," +
		constants.StrPrivacyDestinationLast + `)){
					tx_inputs{
						dsti as ~tx_outputs
					}
				}
				
				connected(func: uid(mx,dst)){
					uid
					privacytype
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
					privacytype
					block:~transactions{
						ts
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query,
		map[string]string{"$bid": strconv.FormatUint(blockHeight, 10)})
	if err != nil {
		return nil, nil, err
	}

	var r struct {
		Connected []ConnectedNodeRequest `json:"connected"`
		Single    []Node                 `json:"single"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, nil, cliutil.NewStackError(err)
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

// GetPrivacyTypeData returns timestamps when the transactions of the specified privacyType occur.
// If the string is empty then all privacy transactions are considered.
func GetPrivacyTypeData(c external.Database, startRange string, stopRange string) (ts []time.Time, err error) {
	const query = `query Q($ge:string,$le:string){
				q(func:between(privacytype,$ge,$le))@normalize{
					~transactions{
						ts:ts
					}
			  	}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query,
		map[string]string{"$ge": startRange, "$le": stopRange})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Query []struct {
			Timestamp time.Time `json:"ts,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	for _, q := range r.Query {
		ts = append(ts, q.Timestamp.UTC())
	}

	return
}

// GetForwardLookupTransactions traverses forward in the transaction graph, starting with the transaction
// specified by startTxHash. This function is used to generate test data.
func GetForwardLookupTransactions(c external.Database, startTxHash string) (blocks []db.Block,
	addresses []db.Address, transactions []db.Transaction, err error) {
	const query = `query Q($txhash: string) {
				var(func: eq(txhash, $txhash))@recurse(depth:3){
					tx_outputs
					~tx_inputs@filter(has(privacytype))
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
						privacytype
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*20, query,
		map[string]string{"$txhash": startTxHash})
	if err != nil {
		return
	}
	var r struct {
		Blocks              []db.Block       `json:"blocks,omitempty"`
		Addresses           []db.Address     `json:"addresses,omitempty"`
		ShallowTransactions []db.Transaction `json:"shallow_txs,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
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
	TxHash       string
	ClusterSize  int
	Destinations []string
}

// GetDestinationTransactionSpenders returns all transactions which spend at least one output of a destination transaction
func GetDestinationTransactionSpenders(c external.Database) (transactions []SpenderTransaction, err error) {
	const query = `{
		destinations as var(func: between(privacytype,100,199))@cascade{
			~transactions@filter(gt(ts,"2018-01-01T00:00:00"))
		}

		var(func: uid(destinations)){
			tx_outputs{
				using_dst as ~tx_inputs
			}
		}
		
		q(func: uid(using_dst)){
			uid
			txhash
			tx_inputs@normalize{
				~tx_outputs@filter(between(privacytype,100,199)){
					txhash:txhash
				}
				~addr_outputs {
					~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
						clusterCount:Cluster.addressCount
					}
				}
			}
		}
	}`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*20, query)
	if err != nil {
		return
	}
	var r struct {
		Transactions []struct {
			UID               string `json:"uid,omitempty"`
			TransactionHash   string `json:"txhash,omitempty"`
			InputTransactions []struct {
				TransactionHash string `json:"txhash,omitempty"`
				ClusterCount    int    `json:"clusterCount,omitempty"`
			} `json:"tx_inputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	spentDestinationTransactions := map[string]int{}
	usingDestinationTransactionsCount := 0
	excludedBecauseOfClusterSize := 0
	for _, tx := range r.Transactions {
		txCount := len(tx.InputTransactions)
		if txCount < 2 {
			continue
		}

		uniqueTxs := map[string]int{}
		var clusterCount int
		for _, it := range tx.InputTransactions {
			if it.TransactionHash == "" {
				// filter out artifacts from normalization
				continue
			}
			uniqueTxs[it.TransactionHash] = it.ClusterCount
			clusterCount = it.ClusterCount
		}

		if len(uniqueTxs) < 2 {
			continue
		}

		if clusterCount > 1000 {
			excludedBecauseOfClusterSize++
			continue
		}

		for k, v := range uniqueTxs {
			spentDestinationTransactions[k] = v
		}
		usingDestinationTransactionsCount++

		transactions = append(transactions, SpenderTransaction{
			TxHash:       tx.TransactionHash,
			ClusterSize:  clusterCount,
			Destinations: cliutil.GetMapKeys(uniqueTxs),
		})
	}

	fmt.Println("spent destination transactions", len(spentDestinationTransactions))
	fmt.Println("excluded because of cluster size", excludedBecauseOfClusterSize)
	fmt.Println("using destination transactions count", usingDestinationTransactionsCount)
	return
}
