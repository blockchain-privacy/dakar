package analytics

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
	dbtx "dashrpc/db/transaction"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"log"
)

// Searches for all potential origins. The returned string slice contains the uids of the found transactions
func AnalyzeOrigins(c *dgo.Dgraph, transactionHash string) (origins []string, err error) {
	query := `query Q($hash: string) {
				tx as var(func: eq(txhash, $hash))
	
				var(func: uid(tx))@recurse{
					tx_inputs
					v as ~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
				}

				q(func: uid(v))@filter(eq(privacytype,"origin")){
					uid
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$hash": transactionHash})
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
					v%d as ~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
				}

				o%d as var(func: uid(v%d))@filter(eq(privacytype,"origin"))`, i, i, i, i, i, i)

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
	if err = db.TxWithRetry(c, db.GetBackendContext(), req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// Searches for all potential origins and sets them.
func AnalyzeAndSetOrigins(c *dgo.Dgraph, txUid string) (err error) {
	query := `query Q($uid: string) {
				u as var(func: uid($uid))
				var(func: uid(u))@recurse{
					tx_inputs
					v as ~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
				}

				o as var(func: uid(v))@filter(eq(privacytype,"origin"))
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

	if txErr := db.TxWithRetry(c, db.GetBackendContext(), req); txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	return
}

// Gets all direct origin transactions and the accumulated origins of all direct mixing and destination transactions
func GetAccumulatedOrigins(c *dgo.Dgraph, transactionHash string) (origins []string, err error) {
	query := `query Q($hash: string) {
				tx as var(func: eq(txhash, $hash))
				
				var(func: uid(tx)){
					tx_inputs{
						~tx_outputs@filter(eq(privacytype,"origin")){
							u as uid
						}
					}
				}
				
				var(func: uid(tx)){
					tx_inputs{
						~tx_outputs@filter(eq(privacytype,["mixing","destination"])){
							o as origins
						}
					}
				}
				
				q(func: uid(u,o)){
					uid 
				}
			}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$hash": transactionHash})
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

// builds the request for the IRTL function. The request includes the mapped variables, query and mutations
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
						u%d as ~tx_outputs@filter(eq(privacytype,"origin"))
					}
				}
				
				var(func: uid($uid%d)){
					tx_inputs{
						~tx_outputs@filter(eq(privacytype,["mixing","origin"])){
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
	if err = db.TxWithRetry(c, db.GetBackendContext(), req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// gets all origin uids of the transaction specified by txHash
func GetOrigins(c *dgo.Dgraph, txHash string) (origins []string, err error) {
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash))@normalize{
					origins{
						uid: uid
					}
				}
			  }`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetFrontendContext(), query, map[string]string{"$hash": txHash})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Q []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Q) == 0 {
		err = errors.New("invalid response from database")
		return
	}

	for _, o := range r.Q {
		origins = append(origins, o.Uid)
	}

	return
}

// gets all origin uids of the transaction specified by txHash
func GetOrigins(c *dgo.Dgraph, txHash string) (origins []string, err error) {
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash))@normalize{
					origins{
						uid: uid
					}
				}
			  }`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetFrontendContext(), query, map[string]string{"$hash": txHash})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Q []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Q) == 0 {
		err = errors.New("invalid response from database")
		return
	}

	for _, o := range r.Q {
		origins = append(origins, o.Uid)
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$hash": txHash})
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

// Searches for all potential origins. The returned string slice contains the uids of the found transactions
func GetPaths(c *dgo.Dgraph, transactionHash string) (paths []TransactionPath,
	transactions map[string]dbtx.FrontendTransaction, err error) {
	query := `query Q($hash: string) {
				tx as var(func: eq(txhash, $hash))
	
				q(func: uid(tx))@recurse{
					tx_inputs
					~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
					txs as txhash
					privacytype
				}
				x(func: uid(txs))@normalize{
					txhash: txhash
					~transactions{
						bhash: blockhash
						bid: id
						bts: ts
					}
				}
			  }`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetBackendContext(), query, map[string]string{"$hash": transactionHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction          []transaction              `json:"q,omitempty"`
		FrontendTransactions []dbtx.FrontendTransaction `json:"x,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) != 1 || len(r.Transaction[0].Inputs) == 0 || len(r.FrontendTransactions) == 0 {
		err = errors.New("invalid response from database")
		return
	}

	paths = getTransactionsPaths(r.Transaction[0].Inputs)

	transactions = make(map[string]dbtx.FrontendTransaction)
	for _, t := range r.FrontendTransactions {
		transactions[t.Hash] = t
	}

	return
}

// gets all uids of the transactions which produce the inputs for the transactions included in the block specified by blockUid
func GetNotAnalyzedInputTransactionsPerBlock(c *dgo.Dgraph, blockUid string) (inputTransactions []string, err error) {
	query := `query Q($uid: string){
				var(func: uid($uid)){
					transactions@filter(eq(privacytype,"destination")){
						tx_inputs{
							v as ~tx_outputs@filter(eq(privacytype, ["mixing", "origin"]) AND eq(count(origins),0))
						}
					} 
				}
				
				q(func: uid(v)){
					uid
				}
			   }`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetBackendContext(), query, map[string]string{"$uid": blockUid})
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
func GetNotAnalyzedInputTransactions(c *dgo.Dgraph, txUid string) (inputTransactions []string, err error) {
	query := `query Q($uid: string){
				var(func: uid($uid)){
					tx_inputs{
						v as ~tx_outputs@filter(eq(privacytype, ["mixing", "origin"]) AND eq(count(origins),0))
					}
				}
				
				q(func: uid(v)){
					uid
				}
			   }`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetBackendContext(), query, map[string]string{"$uid": txUid})
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

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetBackendContext(), query, map[string]string{"$uid": blockUid})
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

func GetPrivacyTransactions(c *dgo.Dgraph, transactionHash string) (transactions []transaction, err error) {
	query := `query Q($hash: string) {
				q(func: has(privacytype)){
					txhash
				}
			  }`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetBackendContext(), query, map[string]string{"$hash": transactionHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []transaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	transactions = r.Transaction

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

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetBackendContext(), query, map[string]string{"$hash": transactionHash})
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

// starting at inputs get all possible paths to PrivateSend origins
func getTransactionsPaths(inputs []input) (filteredPaths []TransactionPath) {
	var paths []TransactionPath
	findPaths(inputs, &paths, nil)
	filteredPaths = filterPaths(paths)
	return
}

// saves all paths in inputs to paths; path holds the path of the curren recursion
func findPaths(inputs []input, paths *[]TransactionPath, path TransactionPath) {
	for _, i := range inputs {
		// transaction slice ALWAYS exists and has ALWAYS one element
		tx := i.Transaction[0]

		// copy path
		newPath := make(TransactionPath, len(path))
		copy(newPath, path)

		// add new PathElement
		isOrigin := false
		if tx.PrivacyType == dbtx.PrivacyOrigin {
			isOrigin = true
		}
		newPath = append(newPath, PathElement{
			IsOrigin: isOrigin,
			Hash:     tx.Hash,
		})

		// end of path
		if tx.Inputs == nil {
			*paths = append(*paths, newPath)
			continue
		}

		findPaths(tx.Inputs, paths, newPath)
	}
}
