package analytics

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	dbtx "backend/db/transaction"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

const (
	// SameRequestMutationLimit is the maximum number of origins a reverse
	// lookup can produce, while getting inserted into the db in the same request
	SameRequestMutationLimit = 2000
	// StrSameRequestMutationLimit is the string representation of SameRequestMutationLimit
	StrSameRequestMutationLimit = "2000"

	// MinCheckPointSize is the number of origins a transaction has to be connected to become
	// a reverse lookup checkpoint
	MinCheckPointSize = 100
	// StrMinCheckPointSize is the string representation of MinCheckPointSize
	StrMinCheckPointSize = "100"

	// MaxCheckPointSize is the number of origins a transaction has to be connected to become
	// a reverse lookup checkpoint
	MaxCheckPointSize = 1000
	// StrMaxCheckPointSize is the string representation of MaxCheckPointSize
	StrMaxCheckPointSize = "1000"
)

// AnalyzeOriginsV2 searches for all potential origins of a transaction by traversing connected mixing
// transactions. In case more than MinCheckPointSize origins are found, the origins are connected
// to the transaction. These transaction are called checkpoints. Checkpoints are not traversed in further lookups.
// If checkpoints are found in a backward traversal and the analysed transaction is becoming a checkpoint, the
// checkpoints and their checkpoints are connected to the newly created checkpoint. See following ASCII graph:
// The transaction at the bottom with three origins (O:3) is a checkpoint if the minimum Checkpoint size is 3. It has three
// origins by itself and is also connected to another checkpoint with three origins.
//
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
// │Origin├──┴─┤Mixing├────┤Mixing├─┤  O:3  C
// └──────┘    └──────┘    └──────┘ │ ┌──────┐
//              O:1         O:1     ├─┤Mixing│
// ┌──────┐    ┌──────┐    ┌──────┐ │ └──────┘
// │Origin├────┤Mixing├────┤Mixing├─┘
// └──────┘    └──────┘    └──────┘
func AnalyzeOriginsV2(c *dgo.Dgraph, txUid string) (numOrigins uint64, numDirectCheckpoints uint64,
	numIndirectCheckpoints uint64, err error) {
	const query = `query Q($uid: string) {
				var(func: uid($uid))@recurse{
					tx_inputs
					v as ~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast +
		`) AND not has(origins))
				}

				var(func: uid(v,$uid)){
					tx_inputs{
						dc as ~tx_outputs@filter(has(origins)){
							dco as origins
							c as checkpoints{
								co as origins
							}
						}
					}
				}

				dcoo as var(func: uid(dco, co))

				var(func: uid(v,$uid)){
					tx_inputs{
						f as ~tx_outputs@filter(between(privacytype,` + constants.StrPrivacyOriginFirst + "," +
		constants.StrPrivacyOriginLast + `) AND NOT uid(dcoo))
					}
				}

				dcc as var(func: uid(dc,c))

				q(func: uid(f)){
					count(uid)
				}
				x(func: uid(dc)){
					count(uid)
				}
				y(func: uid(c)){
					count(uid)
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uid": txUid},
		Mutations: []*api.Mutation{
			{
				Cond:      "@if(ge(len(f)," + StrMinCheckPointSize + "))",
				SetNquads: []byte("<" + txUid + "> <origins> uid(f) ."),
			},
			{
				Cond:      "@if(ge(len(f)," + StrMinCheckPointSize + ") AND gt(len(dcc),0))",
				SetNquads: []byte("<" + txUid + "> <checkpoints> uid(dcc) ."),
			},
		},
		CommitNow: true,
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Origins []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"q,omitempty"`
		DirectCheckpoints []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"x,omitempty"`
		IndirectCheckpoints []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"y,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Origins) != 1 || len(r.DirectCheckpoints) != 1 || len(r.IndirectCheckpoints) != 1 {
		err = errors.New("error invalid number of counts")
		return
	}

	numOrigins = r.Origins[0].Count
	numDirectCheckpoints = r.DirectCheckpoints[0].Count
	numIndirectCheckpoints = r.IndirectCheckpoints[0].Count
	return
}

// AnalyzeOrigins searches for all potential origins. The returned string slice contains the uids of the found transactions
// GET part of AnalyzeAndSetOrigins
func AnalyzeOrigins(c *dgo.Dgraph, txUid string) (origins []string, err error) {
	query := `query Q($uid: string) {
				y as var(func: uid($uid))
				var(func: uid(y))@recurse{
					tx_inputs
					v as ~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast + `))
				}
				
				var(func: uid(v,y)){
					tx_inputs{
						f as ~tx_outputs@filter(between(privacytype,` + constants.StrPrivacyOriginFirst + "," +
		constants.StrPrivacyOriginLast + `))
					}
				}

				q(func: uid(f)){
					uid
				}
			  }`

	nQuadString := fmt.Sprintf("<%s> <origins> uid(f) .\n<%s> <isrlookupdone> \"true\" .", txUid, txUid)

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uid": txUid},
		Mutations: []*api.Mutation{{
			Cond:      "@if(lt(len(f)," + strconv.Itoa(SameRequestMutationLimit) + ") AND gt(len(f), 0))",
			SetNquads: []byte(nQuadString),
		}},
		CommitNow: true,
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, uid := range r.Transaction {
		origins = append(origins, uid.Uid)
	}

	return
}

// AnalyzeOriginsTest searches for all potential origins. The returned string slice contains the uids of the found transactions
// GET part of AnalyzeAndSetOrigins
func AnalyzeOriginsTest(c *dgo.Dgraph, txUid string) (origins []string, err error) {
	query := `query Q($uid: string) {
				y as var(func: uid($uid))
				var(func: uid(y))@recurse{
					tx_inputs
					v as ~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast + `))
				}
				
				var(func: uid(v,y)){
					tx_inputs{
						f as ~tx_outputs@filter(between(privacytype,` + constants.StrPrivacyOriginFirst + "," +
		constants.StrPrivacyOriginLast + `))
					}
				}

				q(func: uid(f)){
					uid
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*10, query, map[string]string{"$uid": txUid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, uid := range r.Transaction {
		origins = append(origins, uid.Uid)
	}

	return
}

// AnalyzeDestinationsTest searches for all potential origins. The returned string slice contains the uids of the found transactions
// GET part of AnalyzeAndSetOrigins
func AnalyzeDestinationsTest(c *dgo.Dgraph, txUid string) (origins []string, err error) {
	query := `query Q($uid: string) {
				y as var(func: uid($uid))
				var(func: uid(y))@recurse{
					tx_outputs
					v as ~tx_inputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast + `))
				}
				
				var(func: uid(v,y)){
					tx_outputs{
						f as ~tx_outputs@filter(between(privacytype,` + constants.StrPrivacyDestinationFirst + "," +
		constants.StrPrivacyDestinationLast + `))
					}
				}

				q(func: uid(f)){
					uid
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*10, query, map[string]string{"$uid": txUid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, uid := range r.Transaction {
		origins = append(origins, uid.Uid)
	}

	return
}

func buildAnalyzeAndSetOriginsRequest(transactionUids []string) *api.Request {
	queryHeader := "query Q("
	var query string
	queryVars := make(map[string]string)

	var mutations []*api.Mutation

	// build query
	for i, u := range transactionUids {

		mutations = append(mutations, &api.Mutation{
			Cond:      fmt.Sprintf("@if(gt(len(o%d), 0))", i),
			SetNquads: []byte(fmt.Sprintf("uid(u%d) <origins> uid(o%d) .", i, i)),
		})

		queryVars[fmt.Sprintf("$uid%d", i)] = u
		queryHeader += fmt.Sprintf("$uid%d:string", i)

		query += fmt.Sprintf(`
				u%d as var(func: uid($uid%d))
				var(func: uid(u%d))@recurse{
					tx_inputs
					# mixing or origin
					v%d as ~tx_outputs@filter(between(privacytype,0,`+constants.StrPrivacyMixingLast+
			`) or between(privacytype,`+constants.StrPrivacyOriginFirst+","+constants.StrPrivacyOriginLast+`))
				}

				o%d as var(func: uid(v%d))@filter(between(privacytype,`+constants.StrPrivacyOriginFirst+","+
			constants.StrPrivacyOriginLast+`))`, i, i, i, i, i, i)

		if i+1 < len(transactionUids) {
			queryHeader += ","
		}
	}

	queryHeader += "){"

	return &api.Request{
		Query:     queryHeader + query + "}",
		Vars:      queryVars,
		Mutations: mutations,
		CommitNow: true,
	}
}

func PartialReverseLookup(c *dgo.Dgraph, transactionUids []string) (err error) {
	if len(transactionUids) == 0 {
		return
	}

	req := buildAnalyzeAndSetOriginsRequest(transactionUids)
	if err = db.TxWithRetry(c, time.Minute*10, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// AnalyzeAndSetOrigins searches for all potential origins and sets them.
// Non-batched version of PartialReverseLookup
func AnalyzeAndSetOrigins(c *dgo.Dgraph, txUid string) (err error) {
	const query = `query Q($uid: string) {
				u as var(func: uid($uid))
				var(func: uid(u))@recurse{
					tx_inputs
					v as ~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast +
		`) or between(privacytype,` + constants.StrPrivacyOriginFirst + "," + constants.StrPrivacyOriginLast + `))
				}

				o as var(func: uid(v))@filter(between(privacytype,` + constants.StrPrivacyOriginFirst + "," +
		constants.StrPrivacyOriginLast + `))
			  }`

	mu := &api.Mutation{
		Cond:      "@if(gt(len(o), 0))",
		SetNquads: []byte("uid(u) <origins> uid(o) ."),
	}
	req := &api.Request{
		Query:     query,
		Vars:      map[string]string{"$uid": txUid},
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	if txErr := db.TxWithRetry(c, time.Minute*20, req); txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	return
}

// GetAccumulatedOrigins gets all direct origin transactions and the accumulated
// origins of all direct mixing and destination transactions
func GetAccumulatedOrigins(c *dgo.Dgraph, uid string) (origins []string, err error) {
	const query = `query Q($uid: string) {
				var(func: uid($uid)){
					tx_inputs{
						u as ~tx_outputs@filter(between(privacytype,` + constants.StrPrivacyOriginFirst + "," +
		constants.StrPrivacyOriginLast + `))
					}
				}
				
				var(func: uid($uid)){
					tx_inputs{
						~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast +
		`) or between(privacytype,` + constants.StrPrivacyOriginFirst + "," + constants.StrPrivacyOriginLast + `)){
							o as origins
						}
					}
				}
				
				q(func: uid(u,o)){
					uid 
				}
			}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$uid": uid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, uid := range r.Transaction {
		origins = append(origins, uid.Uid)
	}

	return
}

// buildIRTLRequest builds the request for the IRTL function. The request includes the mapped variables, query and mutations
func buildIRTLRequest(transactionUids map[string]bool) *api.Request {
	queryHeader := "query Q("
	var query string
	queryVars := make(map[string]string)

	var mutations []*api.Mutation

	// build query
	i := 0
	for u := range transactionUids {

		mutations = append(mutations, &api.Mutation{
			Cond:      fmt.Sprintf("@if(gt(len(f%d), 0))", i),
			SetNquads: []byte(fmt.Sprintf("uid(tx%d) <origins> uid(f%d) .", i, i)),
		})

		queryVars[fmt.Sprintf("$uid%d", i)] = u
		queryHeader += fmt.Sprintf("$uid%d:string", i)

		query += fmt.Sprintf(`
				tx%d as var(func: uid($uid%d)){
					tx_inputs{
						u%d as ~tx_outputs@filter(between(privacytype,`+constants.StrPrivacyOriginFirst+","+
			constants.StrPrivacyOriginLast+`))
					}
				}
				
				var(func: uid($uid%d)){
					tx_inputs{
						~tx_outputs@filter(between(privacytype,0,`+constants.StrPrivacyMixingLast+
			`) or between(privacytype,`+constants.StrPrivacyOriginFirst+","+constants.StrPrivacyOriginLast+`)){
							o%d as origins
						}
					}
				}
				
				f%d as var(func: uid(o%d,u%d))`, i, i, i, i, i, i, i, i)

		if i+1 < len(transactionUids) {
			queryHeader += ","
		}
		i++
	}

	queryHeader += "){"

	return &api.Request{
		Query:     queryHeader + query + "}",
		Vars:      queryVars,
		Mutations: mutations,
		CommitNow: true,
	}
}

// IRTL stands for "Incremental Reverse Transaction Lookup". It sets all direct origin transactions
// and the accumulated origins of all direct mixing and destination transactions
func IRTL(c *dgo.Dgraph, transactionUids map[string]bool) (err error) {
	if len(transactionUids) == 0 {
		return
	}

	req := buildIRTLRequest(transactionUids)
	if err = db.TxWithRetry(c, time.Minute*20, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// ConnectDirectNeighbours sets all direct origin transactions and the
// accumulated origins of all direct mixing and destination transactions
// non-batched version of IRTL
func ConnectDirectNeighbours(c *dgo.Dgraph, transactionUid string) (err error) {
	queryVars := map[string]string{"$uid": transactionUid}

	const query = `query Q($uid:string){
				tx as var(func: uid($uid)){
					tx_inputs{
						u as ~tx_outputs@filter(between(privacytype,` + constants.StrPrivacyOriginFirst + "," +
		constants.StrPrivacyOriginLast + `))
					}
				}
				
				var(func: uid($uid)){
					tx_inputs{
						~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast +
		`) or between(privacytype,` + constants.StrPrivacyOriginFirst + "," + constants.StrPrivacyOriginLast + `)){
							o as origins
						}
					}
				}
				
				f as var(func: uid(o,u))
			   }`

	req := &api.Request{
		Query: query,
		Vars:  queryVars,
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(f), 0))",
			SetNquads: []byte("uid(tx) <origins> uid(f) ."),
		}},
		CommitNow: true,
	}

	if err = db.TxWithRetry(c, time.Minute*10, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// SetOrigins sets all originUids as origins of txUid.
// If isDone is true, the isrlookupdone flag for this transaction will be set
func SetOrigins(c *dgo.Dgraph, txUid string, originUids []string, isDone bool) (err error) {
	if len(originUids) == 0 {
		err = errors.New("tried to set origins with length of 0")
		return
	}

	var nQuadString string
	nQuadPart := "<" + txUid + "> <origins> <"
	for _, o := range originUids {
		nQuadString += nQuadPart + o + "> .\n"
	}

	if isDone {
		nQuadString += "<" + txUid + "> <isrlookupdone> \"true\" ."
	}

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetNquads: []byte(nQuadString),
		}},
		CommitNow: true,
	}

	if err = db.TxWithRetry(c, time.Minute*5, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// GetOriginCount gets the number of origins of a transaction
func GetOriginCount(c *dgo.Dgraph, txHash string) (originCount int, err error) {
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash)){
					count(origins)
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{"$hash": txHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Q []struct {
			OriginCount int `json:"count(origins),omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Q) != 1 {
		err = errors.New("invalid response from database")
		return
	}
	originCount = r.Q[0].OriginCount
	return
}

// GetNotAnalyzedInputTransactionsPerBlock gets all uids of the transactions which produce
// the inputs for the transactions included in the block specified by blockUid
func GetNotAnalyzedInputTransactionsPerBlock(c *dgo.Dgraph, blockUid string) (inputTransactions []string, err error) {
	const query = `query Q($uid: string){
				var(func: uid($uid)){
					transactions@filter(between(privacytype,` + constants.StrPrivacyDestinationFirst + "," +
		constants.StrPrivacyDestinationLast + `)){
						tx_inputs{
							v as ~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast +
		`) AND NOT eq(isrlookupdone, true))
						}
					} 
				}
				
				q(func: uid(v)){
					uid
				}
			   }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$uid": blockUid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 {
		return
	}

	for _, u := range r.Transaction {
		inputTransactions = append(inputTransactions, u.Uid)
	}

	return
}

// GetNotAnalyzedInputTransactionsPerTx gets all uids of the transactions which produce the
// inputs for the transactions included in the block specified by blockUid
func GetNotAnalyzedInputTransactionsPerTx(c *dgo.Dgraph, txUid string) (inputTransactions []string, err error) {
	query := `query Q($uid: string){
				var(func: uid($uid)){
					tx_inputs{
						v as ~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast +
		`) AND NOT eq(isrlookupdone, true))
					}
				}
				
				q(func: uid(v)){
					uid
				}
			   }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$uid": txUid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 {
		return
	}

	for _, u := range r.Transaction {
		inputTransactions = append(inputTransactions, u.Uid)
	}

	return
}

// GetInputTransactions gets all uids of the transactions which produce the inputs for the transactions
// included in the block specified by blockUid
func GetInputTransactions(c *dgo.Dgraph, blockUid string) (inputTransactions map[string]map[string]bool, err error) {
	query := `query Q($uid: string) {
				q(func: uid($uid)){
					transactions@filter(has(privacytype)){
						uid
						input_transaction: tx_inputs@normalize{
							~tx_outputs{
								uid: uid
							}
						}
					}
				}
			  }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$uid": blockUid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Block []struct {
			Transaction []struct {
				Uid              string `json:"uid,omitempty"`
				InputTransaction []struct {
					Uid string `json:"uid,omitempty"`
				} `json:"input_transaction,omitempty"`
			} `json:"transactions,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Block) == 0 {
		return
	} else if len(r.Block) != 1 || len(r.Block[0].Transaction) == 0 {
		err = errors.New("invalid response from database")
		return
	}

	inputTransactions = make(map[string]map[string]bool)

	for _, t := range r.Block[0].Transaction {
		inputTransactions[t.Uid] = make(map[string]bool)
		for _, iT := range t.InputTransaction {
			inputTransactions[t.Uid][iT.Uid] = true
		}
	}

	return
}

func SameBlockTest(c *dgo.Dgraph, transactionHash string) (err error) {
	query := `query Q($hash: string) {
				s as var(func: eq(txhash, $hash)){
					block as ~transactions
				}
				
				q(func: uid(s)){
					tx_inputs {
						~tx_outputs@filter(has(privacytype)){
							~transactions@filter(uid(block)){
								id
							}
						}
					}
				}
			}`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$hash": transactionHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(resp.Json) != 8 {
		log.Println(transactionHash, string(resp.Json))
	}

	return
}

//
func filterPaths(paths []TransactionPath) (filteredPaths []TransactionPath) {
	var tailCutPaths []TransactionPath
	// holds for each 'end origin transaction hash' the shortest to number elements
	lenPathMap := make(map[string]int)

	for _, p := range paths {
		// remove 'mixing' transactions at tail
		cutPath := p.cutTail()
		if len(cutPath) == 0 {
			continue
		}

		thisHash := p[len(p)-1].Hash
		prevLen := lenPathMap[thisHash]

		if prevLen != 0 && prevLen < len(p) {
			continue
		}
		lenPathMap[thisHash] = len(p)

		tailCutPaths = append(tailCutPaths, p)
	}

	// post processing of paths
	pathMap := make(map[string]bool)
	for _, p := range tailCutPaths {

		// filter out paths longer than shortest path found for origin
		originHash := p[len(p)-1].Hash
		if lenPathMap[originHash] < len(p) {
			continue
		}

		pathHash := p.hash()

		// if path is already in map then continue
		if pathMap[pathHash] {
			continue
		}
		pathMap[pathHash] = true
		filteredPaths = append(filteredPaths, p)
	}

	return
}

// GetMixingTransactionsByBlock gets mixing transactions from the database by block id
func GetMixingTransactionsByBlock(c *dgo.Dgraph, blockId uint64) (transactions []dbtx.Transaction, err error) {
	const query = `query Q($block:string) {
				var(func: eq(id, $block)){
					txs as transactions@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast +
		`) and not has(origins))
				}

				q(func: uid(txs)){
					uid
					privacytype
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query,
		map[string]string{"$block": strconv.FormatUint(blockId, 10)})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Q []dbtx.Transaction `json:"q"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	transactions = r.Q

	return
}

// GetConnectedPrivacyTransactions gets the first maxTx privacy transactions including their input transaction
// from the database. If maxTx is equal to 0, all mixing transaction are returned.
func GetConnectedPrivacyTransactions(c *dgo.Dgraph, maxTx int, privacyRangeFirst int, privacyRangeLast int) ([]ConnectedNode, error) {
	step := 50000
	max := maxTx
	if max == 0 {
		max = 3300000
	}

	if step > maxTx {
		step = maxTx
	}

	queryString := `{
				q(func: between(privacytype,` + strconv.Itoa(privacyRangeFirst) + "," + strconv.Itoa(privacyRangeLast) + `), first:%d, offset:%d ){
					uid
					block:~transactions{
						ts
					}
					i:tx_inputs@normalize{
						~tx_outputs{
							uid:uid
						}
					}
				}
			  }`

	var nodes []ConnectedNode
	for i := 0; i < max; i = i + step {
		query := fmt.Sprintf(queryString, step, i)

		resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*2, query)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		var r struct {
			Q []ConnectedNode `json:"q"`
		}

		if err = json.Unmarshal(resp.Json, &r); err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		nodes = append(nodes, r.Q...)
	}

	return nodes, nil
}

// GetPrivacyTransactions gets the first maxTx privacy transactions from the database.
// If maxTx is equal to 0, all origin transaction are returned.
func GetPrivacyTransactions(c *dgo.Dgraph, maxTx int, privacyRangeFirst int, privacyRangeLast int) ([]Node, error) {
	step := 50000
	max := maxTx
	if max == 0 {
		max = 3300000
	}

	if step > maxTx {
		step = maxTx
	}
	queryString := `{
				q(func: between(privacytype,` +
		strconv.Itoa(privacyRangeFirst) + "," + strconv.Itoa(privacyRangeLast) + `), first:%d, offset:%d ){
					uid
					block:~transactions{
						ts
					}
				}
			  }`
	var nodes []Node
	for i := 0; i < max; i = i + step {
		query := fmt.Sprintf(queryString, step, i)

		resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*2, query)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		var r struct {
			Q []Node `json:"q"`
		}

		if err = json.Unmarshal(resp.Json, &r); err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		nodes = append(nodes, r.Q...)
	}

	return nodes, nil
}

// GetMixingTransactions gets the first maxTx mixing transactions including their input transactions
// from the database. If maxTx is equal to 0, all mixing transaction are returned.
func GetMixingTransactions(c *dgo.Dgraph, maxTx int) ([]ConnectedNode, error) {
	return GetConnectedPrivacyTransactions(c, maxTx, 0, constants.PrivacyMixingLast)
}

// GetDestinationTransactions gets the first maxTx destination transactions including their input transactions
// from the database. If maxTx is equal to 0, all destination transaction are returned.
func GetDestinationTransactions(c *dgo.Dgraph, maxTx int) ([]ConnectedNode, error) {
	return GetConnectedPrivacyTransactions(c, maxTx, constants.PrivacyDestinationFirst,
		constants.PrivacyDestinationLast)
}

// GetOriginTransactions gets the first maxTx origin transactions from the database.
// If maxTx is equal to 0, all origin transaction are returned.
func GetOriginTransactions(c *dgo.Dgraph, maxTx int) ([]Node, error) {
	return GetPrivacyTransactions(c, maxTx, constants.PrivacyOriginFirst, constants.PrivacyOriginLast)
}

// GetCollateralCreationTransactions gets the first maxTx cc transactions from the database.
// If maxTx is equal to 0, all cc transaction are returned.
func GetCollateralCreationTransactions(c *dgo.Dgraph, maxTx int) ([]Node, error) {
	return GetPrivacyTransactions(c, maxTx, constants.PrivacyCollateralCreationFirst,
		constants.PrivacyCollateralCreationLast)
}
