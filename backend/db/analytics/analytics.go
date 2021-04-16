package analytics

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"strconv"
	"time"

	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// SameRequestMutationLimit is the maximum number of origins a reverse
// lookup can produce, while getting inserted into the db in the same request
const SameRequestMutationLimit = 2000

// AnalyzeOrigins searches for all potential origins. The returned string slice contains the uids of the found transactions
// GET part of AnalyzeAndSetOrigins
func AnalyzeOrigins(c *dgo.Dgraph, txUid string) (origins []string, err error) {
	query := `query Q($uid: string) {
				var(func: uid($uid))@recurse{
					tx_inputs
					v as ~tx_outputs@filter(between(privacytype,0,` + constants.StrPrivacyMixingLast + `))
				}

				var(func: uid(v,$uid)){
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

// gets the number of origins of a transaction
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

// gets all uids of the transactions which produce the inputs for the transactions included in the block specified by blockUid
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
