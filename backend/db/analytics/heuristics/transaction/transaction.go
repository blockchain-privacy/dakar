package transaction

import (
	"backend/cmd/cliutil"
	"backend/db"
	dbtx "backend/db/transaction"
	"backend/external"

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

var (
	errInvalidDatabaseResponse = errors.New("error invalid response")
	ErrNoMutationHappened      = errors.New("no mutation happened")
)

// InsertHeuristic inserts the given heuristic
func InsertHeuristic(c *external.GraphDB, h Heuristic, userUid string) (insertUid string, err error) {
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

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
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

// DeleteUserHeuristics deletes all given heuristic uids of a user
func DeleteUserHeuristics(c *external.GraphDB, uids []string, userUid string) (err error) {
	uidList := db.CreateUidList(uids)

	query := "query Q($uuid:string, $uids:string, $type:string){h as var(func: uid($uids))" +
		"@filter(uid_in(~user_heuristics,$uuid) AND eq(dgraph.type,$type))}"

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uuid": userUid, "$uids": uidList, "$type": DType},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(h) * * .\n<" + userUid + "> <user_heuristics> uid(h) ."),
		}},
		CommitNow: true,
	}

	if txErr := db.TxWithRetry(c, time.Minute*5, req); txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}
	return
}

// DeleteAllUserHeuristics deletes all heuristics of a user
func DeleteAllUserHeuristics(c *external.GraphDB, userUid string) (err error) {
	req := &api.Request{
		Query: "query Q($uuid:string){var(func: uid($uuid)){h as user_heuristics}}",
		Vars:  map[string]string{"$uuid": userUid},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(h) * * .\n<" + userUid + "> <user_heuristics> uid(h) ."),
		}},
		CommitNow: true,
	}

	resp, txErr := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	if v, ok := resp.Metrics.NumUids["mutation_cost"]; !ok || (ok && v == 0) {
		return ErrNoMutationHappened
	}

	return
}

// DeleteAllUserTxHeuristics deletes all heuristics of a user for a particular transaction
func DeleteAllUserTxHeuristics(c *external.GraphDB, txhash string, userUid string) (err error) {
	query := `query Q($uuid:string, $hash:string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				# get all heuristic of that user and transaction
				var(func: uid($uuid)){
					h as user_heuristics@filter(uid_in(h_transaction, uid(tx)))
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uuid": userUid, "$hash": txhash},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(h) * * .\n<" + userUid + "> <user_heuristics> uid(h) ."),
		}},
		CommitNow: true,
	}

	resp, txErr := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	if v, ok := resp.Metrics.NumUids["mutation_cost"]; !ok || (ok && v == 0) {
		return ErrNoMutationHappened
	}

	return
}

// GetHeuristic gets heuristic information from the database
func GetHeuristic(c *external.GraphDB, heuristicUid string) (h Heuristic, err error) {
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Second*20, query, map[string]string{"$uid": heuristicUid})

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

func GetHeuristicResults(c *external.GraphDB, heuristicUid string) (results []HeuristicTransaction, err error) {
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
							uid:uid
						}
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$uid": heuristicUid})

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
		results = append(results, HeuristicTransaction{
			Uid:       t.Uid,
			Addresses: getInputAddresses(t.Inputs),
			Outputs:   t.Outputs,
		})
	}

	return
}

// getInputAddresses returns all addresses without duplicates
func getInputAddresses(inputs []HeuristicInput) []string {
	addressMap := make(map[string]bool)
	for _, i := range inputs {
		addressMap[i.AddressUid] = true
	}

	var addresses []string
	for k := range addressMap {
		addresses = append(addresses, k)
	}

	return addresses
}

// GetInputTransactions returns the input transactions of the given transaction
func GetInputTransactions(c *external.GraphDB, tx string) (inputTransactions []HeuristicTransaction, err error) {
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$txhash": tx})
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

// GetTransactionsWithOutputAmountAndInputAddresses returns a slice of transactions.
// Each transaction contains its output amounts and the addresses of all inputs.
func GetTransactionsWithOutputAmountAndInputAddresses(c *external.GraphDB, uids []string) (origins []HeuristicTransaction, err error) {
	query := `query Q($uids:string){
				q(func: uid($uids)){
					uid
					tx_outputs{
						amount
					}
					tx_inputs@normalize{
						~addr_outputs{
							uid:uid
						}
					}
				}
			   }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$uids": db.CreateUidList(uids)})
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
		origins = append(origins, HeuristicTransaction{
			Uid:       o.Uid,
			Addresses: getInputAddresses(o.Inputs),
			Outputs:   o.Outputs,
		})
	}

	return
}

// GetInputAmounts gets the amounts of the inputs
func GetInputAmounts(c *external.GraphDB, tx string) (transaction HeuristicTransaction, err error) {
	query := `query Q($txhash: string){
				q(func: eq(txhash,$txhash)){
					uid
					tx_inputs{
						amount
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$txhash": tx})
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

// DoesHeuristicUidExist checks if the given heuristic uids exist. All heuristics must belong to the same transaction
func DoesHeuristicUidExist(c *external.GraphDB, txhash string, uids []string) (allExist bool, err error) {
	uidList := db.CreateUidList(uids)

	query := `query Q($hash:string, $uids:string, $type:string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				# filter and count
				q(func: uid($uids))@filter(uid_in(h_transaction, uid(tx)) AND eq(dgraph.type,$type)){
					count(uid)
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query,
		map[string]string{"$hash": txhash, "$uids": uidList, "$type": DType})
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
func GetBasicFrontendHeuristic(c *external.GraphDB, txHash string, userUid string) (heuristics []FrontendHeuristic, err error) {
	query := `query Q($hash: string, $uuid: string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				var(func: uid($uuid)){
					h as user_heuristics@filter(uid_in(h_transaction, uid(tx)))
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
	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash, "$uuid": userUid})
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
func GetFrontendHeuristicByUid(c *external.GraphDB, heuristicUid string, userUid string) (
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
	resp, err := c.Query(ctx, query, map[string]string{"$uid": heuristicUid, "$uuid": userUid})
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

	type mapKey struct {
		txHash  string
		address string
	}

	results := r.Heuristics[0].Results
	var filteredResults []FrontendHeuristicResult
	txAddressMap := make(map[mapKey]bool)

	for _, result := range results {
		k := mapKey{
			txHash:  result.TxHash,
			address: result.AddressHash,
		}

		// check if the address and tx combination already exists
		if txAddressMap[k] {
			continue
		}

		txAddressMap[k] = true
		filteredResults = append(filteredResults, result)
	}

	r.Heuristics[0].Results = filteredResults

	frontendHeuristic = r.Heuristics[0]

	return
}

// GetFrontendHeuristic returns all heuristics for a given transaction
func GetFrontendHeuristic(c *external.GraphDB, txHash string, userUid string) (completeHeuristic FrontendHeuristicComplete, err error) {
	query := `query Q($hash: string, $uuid: string){
				# get tx uid
				tx as var(func: eq(txhash, $hash))
				var(func: uid($uuid)){
					h as user_heuristics@filter(uid_in(h_transaction, uid(tx)))
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
	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash, "$uuid": userUid})
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

// GetShortestTransactionPathAnyDirection returns the transactions of a shortest path between two transactions.
// anyDirection determines the search direction of the shortest transaction path query
// True: Both inputs and outputs are traversed
// False: Only inputs are traversed
// withPrivacyTransactions determines if privacy transactions should be considered when doing the shortest path lookup
func GetShortestTransactionPathAnyDirection(c *external.GraphDB, txFrom string, txTo string,
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
	resp, err := c.Query(ctx, query, map[string]string{"$txFrom": txFrom, "$txTo": txTo})
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
func GetHeuristicListByUser(c *external.GraphDB, userUid string) (frontendHeuristic []HeuristicListItem, err error) {
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$uuid": userUid})
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
