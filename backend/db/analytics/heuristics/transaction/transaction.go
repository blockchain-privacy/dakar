package transaction

import (
	"backend/cmd/cliutil"
	"backend/db"
	dbtx "backend/db/transaction"

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

var (
	errInvalidDatabaseResponse = errors.New("error invalid response")
)

// CopyHeuristicTree copies the complete heuristic tree starting at rootHeuristicUid.
// The heuristic results stay the same.
func CopyHeuristicTree(c *dgo.Dgraph, rootHeuristicUid string) (err error) {
	hTree, treeErr := GetHeuristicTree(c, rootHeuristicUid)
	if treeErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), treeErr)
		return
	}

	// create map to slice
	heuristicMap := make(map[string]Heuristic)
	for i, h := range hTree {
		uid := h.Uid

		// save changes
		h.Uid = "_:" + strconv.Itoa(i)
		hTree[i] = h
		heuristicMap[uid] = h
	}

	for i, h := range hTree {
		if len(h.ParentHeuristic) > 0 {
			parent, ok := heuristicMap[h.ParentHeuristic[0].Uid]
			if !ok {
				return errors.New("error heuristic uid not found")
			}
			h.ParentHeuristic[0].Uid = parent.Uid
			// save change
			hTree[i] = h
		}
	}

	if insertionError := insertHeuristics(c, hTree); insertionError != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), insertionError)
		return
	}

	return
}

// GetRootUids returns the root heuristic uids of the provided heuristics
func GetRootUids(c *dgo.Dgraph, uids []string) (roots []string, err error) {
	// This query gets built for multiple uids
	// {var(func: uid(0x42ae75,0x42ae76,0x42ae77))@recurse{
	//		v as uid
	//			parent_heuristic
	//		}
	//
	//		q(func: uid(v))@filter(not has(parent_heuristic)){
	//			uid
	//		}
	// }

	var queryPart string

	for i, uid := range uids {
		queryPart += uid
		if i+1 < len(uids) {
			queryPart += ","
		}
	}

	query := "{var(func: uid(" + queryPart + `))@recurse{
					v as uid
				parent_heuristic
				}
				
				q(func: uid(v))@filter(not has(parent_heuristic)){
					uid
				}
			}`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxWithRetry(c, ctx, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Roots []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Roots) == 0 {
		err = errors.New("error no heuristic tree root found")
		return
	}

	for _, u := range r.Roots {
		roots = append(roots, u.Uid)
	}

	return
}

// insertHeuristics inserts all given heuristics into the database
func insertHeuristics(c *dgo.Dgraph, heuristics []Heuristic) (err error) {
	for i, h := range heuristics {
		h.SetDType()
		heuristics[i] = h
	}

	pb, err := json.Marshal(heuristics)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}
	ctx, cancel := db.GetBackendContext()
	defer cancel()
	_, err = db.TxWithRetryAndResponse(c, ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// InsertHeuristic inserts the given heuristic
func InsertHeuristic(c *dgo.Dgraph, h Heuristic, userUid string) (insertUid string, err error) {
	h.SetDType()
	h.Timestamp = time.Now().UTC().Format(time.RFC3339)

	var query string

	// if TxHash is not empty we have to search for the transaction uid
	if h.TxHash != "" {
		h.Transaction.Uid = "uid(tx)"
		query = `query Q($txhash: string) {
					tx as var(func: eq(txhash, $txhash))
				  }`
	}

	type dummyUser struct {
		Uid        string      `json:"uid,omitempty"`
		Heuristics []Heuristic `json:"user_heuristics,omitempty"`
	}

	pb, err := json.Marshal(dummyUser{Uid: userUid, Heuristics: []Heuristic{h}})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$txhash": h.TxHash},
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.TxWithRetryAndResponse(c, ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	uids := resp.GetUids()
	if len(uids) != 1 {
		err = errors.New(fmt.Sprintln("invalid number of heuristics inserted. Heuristic count:",
			len(resp.GetUids())))
		return
	}

	// uids has only one element, so insertUid is only set once
	for _, u := range uids {
		insertUid = u
	}

	return
}

// DeleteHeuristics deletes all given uids
func DeleteHeuristics(c *dgo.Dgraph, uids []string, userUid string) (err error) {
	// This query gets built for multiple uids
	// {h as var(func: uid(0xca67913))@filter(eq(dgraph.type,TransactionHeuristic) AND eq(~user_heuristic,[0xc6f5e08]))}

	var queryPart string

	for i, uid := range uids {
		queryPart += uid
		if i+1 < len(uids) {
			queryPart += ","
		}
	}

	query := "{h as var(func: uid(" + queryPart + "))@filter(eq(dgraph.type," + DType +
		"))@cascade{~user_heuristics@filter(uid(" + userUid + "))}}"

	req := &api.Request{
		Query: query,
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(h) * * .\n <" + userUid + "> <user_heuristics> uid(h) ."),
		}},
		CommitNow: true,
	}
	ctx, cancel := db.GetBackendContext()
	defer cancel()
	if txErr := db.TxWithRetry(c, ctx, req); txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}
	return
}

// GetHeuristic gets heuristic information from the database
func GetHeuristic(c *dgo.Dgraph, heuristicUid string) (h Heuristic, err error) {
	query := `query Q($uid: string) {
				q(func: uid($uid)){
					uid
					type
					parameter
					results{
						uid
					}
					ts
					h_transaction{
						uid
					}
					parent_heuristic{
						uid
					}
					~parent_heuristic{
						uid
					}
				}
			  }`
	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$uid": heuristicUid})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r struct {
		Heuristics []Heuristic `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Heuristics) != 1 {
		err = errors.New("invalid response from database")
		return
	}
	h = r.Heuristics[0]
	return
}

// GetHeuristicTree gets a complete heuristic tree descending from rootHeuristicUid
func GetHeuristicTree(c *dgo.Dgraph, rootHeuristicUid string) (h []Heuristic, err error) {
	query := `query Q($uid: string){
					var(func: uid($uid))@recurse{
						h_uid as uid
						~parent_heuristic
					}
					
					q(func: uid(h_uid)){
						uid
						type
						parameter
						h_transaction{
							uid
						}
						results{
							uid
						}
						ts
						parent_heuristic{
							uid
						}
					}
				  }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$uid": rootHeuristicUid})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r struct {
		Heuristics []Heuristic `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Heuristics) == 0 {
		err = errors.New("invalid response from database")
		return
	}
	h = r.Heuristics
	return
}

func GetHeuristicResults(c *dgo.Dgraph, heuristicUid string) (results []HeuristicTransaction, err error) {
	query := `query Q($uid: string) {
				var (func: uid($uid)){
					x as results
				}
				
				q(func: uid(x)){
					uid
					tx_outputs{
						amount
					}
					tx_inputs@normalize{
						~addr_outputs{
							addresshash:addresshash
						}
					}
				}
			  }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$uid": heuristicUid})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []queryHeuristicTransaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, t := range r.Transaction {
		if len(t.Inputs) != 1 {
			var addresses []string
			for _, a := range t.Inputs {
				addresses = append(addresses, a.AddressHash)
			}

			if !areALLAddressesEqual(addresses) {
				err = errors.New("error invalid response")
				return
			}
		}
		results = append(results, HeuristicTransaction{
			Uid:     t.Uid,
			Address: t.Inputs[0].AddressHash,
			Outputs: t.Outputs,
		})
	}

	return
}

// GetOriginsByDate returns all origins which are created after the specified date
func GetOriginsByDate(c *dgo.Dgraph, uid string, timestamp string) (origins []HeuristicTransaction, err error) {
	query := `query Q($uid: string,$ts: string){
				var (func: uid($uid))@cascade{
					v as origins{
						~transactions@filter(gt(ts,$ts))
					}
				}
				
				q(func: uid(v)){
					uid
					tx_outputs{
						amount
					}
					tx_inputs@normalize{
						~addr_outputs{
							addresshash:addresshash
						}
					}
				}
			   }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$uid": uid, "$ts": timestamp})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Origins []queryHeuristicTransaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, o := range r.Origins {
		if len(o.Inputs) != 1 {
			var addresses []string
			for _, a := range o.Inputs {
				addresses = append(addresses, a.AddressHash)
			}

			if !areALLAddressesEqual(addresses) {
				err = errors.New("error invalid response")
				return
			}
		}
		origins = append(origins, HeuristicTransaction{
			Uid:     o.Uid,
			Address: o.Inputs[0].AddressHash,
			Outputs: o.Outputs,
		})
	}

	return
}

// GetDestinationTxOrigins collects all previously found origins of the
// direct input transactions of the given destination transaction
func GetDestinationTxOrigins(c *dgo.Dgraph, txhash string) (origins []HeuristicTransaction, err error) {
	query := `query Q($txhash: string){
				tx as var (func: eq(txhash,$txhash))
				var (func: uid(tx)){
					tx_inputs {
						~tx_outputs {
							v as origins
						}
					}
				}

				var (func: uid(tx)){
					tx_inputs {
						c as ~tx_outputs@filter(eq(privacytype, "origin"))
					}
				}
				
				q(func: uid(v,c)){
					uid
					tx_outputs{
						amount
					}
					tx_inputs@normalize{
						~addr_outputs{
							addresshash:addresshash
						}
					}
				}
			   }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$txhash": txhash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Origins []queryHeuristicTransaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, o := range r.Origins {
		if len(o.Inputs) != 1 {
			var addresses []string
			for _, a := range o.Inputs {
				addresses = append(addresses, a.AddressHash)
			}

			if !areALLAddressesEqual(addresses) {
				err = errors.New("error invalid response")
				return
			}
		}
		origins = append(origins, HeuristicTransaction{
			Uid:     o.Uid,
			Address: o.Inputs[0].AddressHash,
			Outputs: o.Outputs,
		})
	}

	return
}

// return true if all addresses are equal
func areALLAddressesEqual(addresses []string) bool {
	if len(addresses) < 2 {
		return true
	}

	hashes := make(map[string]bool)

	for _, a := range addresses {
		hashes[a] = true
		if len(hashes) > 1 {
			return false
		}
	}

	return true
}

// Returns the input transactions of the given transaction
func GetInputTransactions(c *dgo.Dgraph, tx string) (inputTransactions []HeuristicTransaction, err error) {
	query := `query Q($txhash: string){
				var (func: eq(txhash,$txhash)){
					tx_inputs{
						v as ~tx_outputs
					}
				}
				
				q(func: uid(v)){
					uid
					tx_outputs@normalize{
						amount:amount
						~tx_inputs{
							input_tx:txhash
						}
					}
					~transactions{
						ts
					}
				}
				}`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$txhash": tx})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []queryHeuristicTransaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidDatabaseResponse)
		return
	}

	for _, t := range r.Transaction {
		if len(t.Block) != 1 || len(t.Outputs) == 0 {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidDatabaseResponse)
			return
		}
		inputTransactions = append(inputTransactions, HeuristicTransaction{
			Uid:       t.Uid,
			Timestamp: t.Block[0].Timestamp,
			Outputs:   t.Outputs,
		})
	}

	return
}

// get the amounts of the inputs
func GetInputAmounts(c *dgo.Dgraph, tx string) (transaction HeuristicTransaction, err error) {
	query := `query Q($txhash: string){
				q(func: eq(txhash,$txhash)){
					uid
					tx_inputs{
						amount
					}
				}
			  }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$txhash": tx})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			Uid     string            `json:"uid,omitempty"`
			Outputs []HeuristicOutput `json:"tx_inputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidDatabaseResponse)
		return
	}

	t := r.Transaction[0]
	transaction = HeuristicTransaction{
		Uid:     t.Uid,
		Outputs: t.Outputs,
	}

	return
}

// DoesHeuristicUidExist checks if the given uids exist
func DoesHeuristicUidExist(c *dgo.Dgraph, txhash string, uids []string) (allExist bool, err error) {
	// This query gets built for multiple uids
	//{
	//	x as var(func: uid(0x43239f,0x4323b4))@filter(eq(dgraph.type, "TransactionHeuristic"))@cascade{
	//		h_transaction@filter(eq(txhash, "cdfa16675b1320f84d4bb3569e295cb00bdb2372967eba475785f582a01de05b"))
	//	}
	//
	//	q(func: uid(x)){
	//		uid
	//	}
	//}

	var queryPart string

	for i, uid := range uids {
		queryPart += uid
		if i+1 < len(uids) {
			queryPart += ","
		}
	}

	query := "{x as var(func: uid(" + queryPart + `))@filter(eq(dgraph.type, "` + DType + `"))@cascade{
					h_transaction@filter(eq(txhash, "` + txhash + `"))
				}
				
				q(func: uid(x)){
					count(uid)
				}
			}`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxWithRetry(c, ctx, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Count []struct {
			Number int `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Count) == 0 || len(r.Count) > 1 {
		err = errors.New("error invalid response from database")
		return
	} else if r.Count[0].Number != len(uids) {
		err = errors.New("error received number of uids does not match")
		return
	}

	allExist = true
	return
}

// GetBasicFrontendHeuristic returns all heuristics for a given transaction created by userUid. Basic information only
func GetBasicFrontendHeuristic(c *dgo.Dgraph, txHash string, userUid string) (heuristics []FrontendHeuristic, err error) {
	query := `query Q($hash: string, $uuid: string){
				var(func: uid($uuid)){
					h as user_heuristics@cascade{
						h_transaction@filter(eq(txhash, $hash))
					}
				}
				
				q(func: uid(h)){
					uid
					ts
					type
					parameter
					parent_heuristic{
						uid
					}
					children: ~parent_heuristic{
						uid
					}
					num_results: count(results)
				}
			  }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$hash": txHash, "$uuid": userUid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Heuristics []FrontendHeuristic `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Heuristics) > 0 {
		heuristics = r.Heuristics
	}

	return
}

// GetFrontendHeuristicByUid the heuristic for the given heuristicUid
func GetFrontendHeuristicByUid(c *dgo.Dgraph, heuristicUid string, userUid string) (
	frontendHeuristic FrontendHeuristic, err error) {
	query := `query Q($uid: string, $uuid: string){
					q(func: uid($uid))@cascade{
						~user_heuristics@filter(uid($uuid))
						uid
						results@normalize{
							txhash:txhash
							~transactions{
								ts:ts
							}
							tx_inputs{ 
								~addr_outputs{
									addresshash:addresshash
								}
							}
						}
					}
				}`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$uid": heuristicUid, "$uuid": userUid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Heuristics []FrontendHeuristic `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Heuristics) != 1 {
		err = errors.New("invalid response from database")
		return
	}

	frontendHeuristic = r.Heuristics[0]

	return
}

// GetFrontendHeuristic returns all heuristics for a given transaction
func GetFrontendHeuristic(c *dgo.Dgraph, txHash string, userUid string) (completeHeuristic FrontendHeuristicComplete, err error) {
	query := `query Q($hash: string, $uuid: string){
				var(func: uid($uuid)){
					h as user_heuristics@cascade{
						tx as h_transaction@filter(eq(txhash,$hash))
					}
				}
				t(func: uid(tx))@normalize{
					uid:uid
					~transactions{
						ts:ts
					}
				}
				q(func: uid(h)){
					uid
					ts
					type
					parameter
					parent_heuristic{
						uid
					}
					children: ~parent_heuristic{
						uid
					}
					results@normalize{
						uid:uid
						txhash:txhash
						~transactions{
							ts:ts
						}
						tx_inputs{ 
							~addr_outputs{
								addresshash:addresshash
							}
						}
					}
				}
			  }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$hash": txHash, "$uuid": userUid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Heuristics  []FrontendHeuristic         `json:"q,omitempty"`
		Transaction []FrontendHeuristicComplete `json:"t,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) != 1 || len(r.Heuristics) == 0 ||
		len(r.Transaction[0].Uid) == 0 || len(r.Transaction[0].Timestamp) == 0 {
		err = errors.New("invalid response from database")
		return
	}

	completeHeuristic = r.Transaction[0]

	for i, h := range r.Heuristics {
		transactions := make(map[string]bool)

		var results []FrontendHeuristicResult
		for _, r := range h.Results {
			// only append a result once per transaction
			if transactions[r.Uid] {
				continue
			}
			results = append(results, r)
			transactions[r.Uid] = true
		}
		r.Heuristics[i].Results = results
	}

	completeHeuristic.Heuristics = r.Heuristics

	return
}

// GetShortestPathLength returns the number of transactions in the shortest path between the given
// transactions specified by fromUid and toUid. The returned number is the count of transactions in
// the path between fromUid and toUid. A shortest path like the following would return the number 3.
// Example path: fromUid -> tx1 -> output1 -> tx2 -> output2 -> tx3 -> output3 -> toUid
func GetShortestPathLength(c *dgo.Dgraph, fromUid string, toUid string) (pathLength int, err error) {
	query := fmt.Sprintf(`{
				shortest(from: %s, to: %s){
					tx_inputs
					~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
				}
			  }`, fromUid, toUid)

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().Query(ctx, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Path []struct {
			Weight float64 `json:"_weight_,omitempty"`
		} `json:"_path_,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Path) != 1 {
		err = errors.New("invalid response from database")
		return
	}

	pathLength = (int(r.Path[0].Weight) - 2) / 2

	return
}

// GetShortestTransactionPathAnyDirection returns the transactions of a shortest path between two transactions.
// anyDirection determines the search direction of the shortest transaction path query
// True: Both inputs and outputs are traversed
// False: Only inputs are traversed
// withPrivacyTransactions determines if privacy transactions should be considered when doing the shortest path lookup
func GetShortestTransactionPathAnyDirection(c *dgo.Dgraph, txFrom string, txTo string,
	withPrivacyTransactions bool, anyDirection bool) (txs []dbtx.FrontendTransaction, err error) {
	/* Full query
	query Q($txFrom:string, $txTo:string){
					f as var(func: eq(txhash,$txFrom))
					t as var(func: eq(txhash,$txTo))
					path as shortest(from: uid(f), to: uid(t)){
						tx_inputs
						~tx_outputs@filter(NOT has(privacytype)) tx_outputs ~tx_inputs@filter(NOT has(privacytype)) }
					path(func: uid(path))@normalize{
						txhash:txhash
						privacytype:privacytype
						~transactions{
							bid:id
							bts:ts
							bhash:blockhash
						}
					}
				  }
	*/

	privacyFlag := " " // spaces are needed

	if !withPrivacyTransactions {
		privacyFlag = "@filter(NOT has(privacytype)) " // spaces are needed
	}

	var anyDirectionFlag string

	if anyDirection {
		anyDirectionFlag = "tx_outputs ~tx_inputs" + privacyFlag
	}

	query := `query Q($txFrom:string, $txTo:string){
				f as var(func: eq(txhash,$txFrom))
				t as var(func: eq(txhash,$txTo))
				path as shortest(from: uid(f), to: uid(t)){
					tx_inputs
					~tx_outputs` + privacyFlag + anyDirectionFlag + `}
				path(func: uid(path))@normalize{
					txhash:txhash
					privacytype:privacytype
					~transactions{
						bid:id
						bts:ts
						bhash:blockhash
					}
				}
			  }`

	// without retry, as this request can easily timeout
	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$txFrom": txFrom, "$txTo": txTo})
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
		err = nil
		return
	}

	// json struct
	var r struct {
		Transactions []dbtx.FrontendTransaction `json:"path,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	txs = r.Transactions

	return
}

// GetHeuristicListByUser returns all transactions for which the given user has created heuristics
func GetHeuristicListByUser(c *dgo.Dgraph, userUid string) (frontendHeuristic []HeuristicListItem, err error) {
	query := `query Q($uuid:string){
				# get transaction
				var(func: uid($uuid)){
					user_heuristics{
						tx as h_transaction
					}
				}
				# get count
				var(func: uid(tx)){
					c as count(~h_transaction)@filter(uid_in(~user_heuristics,$uuid))
				}
				# get time
				var(func: uid(tx)){
					~h_transaction@filter(uid_in(~user_heuristics,$uuid)){
						t as ts
					}
					max_time as  max(val(t))
				}
				# output
				q(func: uid(tx)){
					txhash
					h_count: val(c)
					mod_time: val(max_time)
				}
			   }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$uuid": userUid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Items []HeuristicListItem `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	frontendHeuristic = r.Items

	return
}
