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
		return nil, cliutil.NewStackError(err)
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
			return nil, cliutil.NewStackError(conversionErr)
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
		return nil, cliutil.NewStackError(err)
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
		err = cliutil.NewStackError(err)
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
		return nil, nil, cliutil.NewStackError(err)
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
			return nil, nil, cliutil.NewStackError(conversionErr)
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
		err = cliutil.NewStackError(err)
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
