package transaction

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"time"
)

// Upserts the given heuristic
func UpsertHeuristic(c *dgo.Dgraph, h Heuristic) (insertUid string, err error) {
	// todo check if needed
	// delete potential already existing edges of h
	//if err = deleteHeuristicEdges(c, h); err != nil {
	//	err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	//	return
	//}

	h.SetDType()
	h.Timestamp = time.Now().UTC().Format(time.RFC3339)
	h.TxUid = "uid(tx)"

	pb, err := json.Marshal(h)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	query := `
		query Q($txhash: string, $huid: string) {
			tx as var(func: eq(txhash, $txhash))
		}
	`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$txhash": h.TxHash},
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}
	resp, err := db.TxWithRetryAndResponse(c, db.GetBackendContext(), req)
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

// Deletes all edges of the result predicate of h
func deleteHeuristicEdges(c *dgo.Dgraph, h Heuristic) (err error) {
	h.SetDType()
	h.Uid = "uid(h)"

	query := `
		query Q($txhash: string, $type: string, $parameter: string) {
			var(func: eq(txhash, $txhash)){
				h as ~h_transaction@filter(eq(type,$type) AND eq(parameter, $parameter))
			}
		}
	`

	mu := &api.Mutation{}
	dgo.DeleteEdges(mu, "uid(h)", "results")

	req := &api.Request{
		Query:     query,
		Vars:      map[string]string{"$txhash": h.TxHash, "$type": h.HeuristicType, "$parameter": h.Parameter},
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	if err = db.TxWithRetry(c, db.GetBackendContext(), req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return
}

// gets heuristic information from the database
func GetHeuristic(c *dgo.Dgraph, heuristicUid string) (h Heuristic, err error) {
	query := `query Q($uid: string) {
				q(func: uid($uid)){
					uid
					type
					parameter
					h_transaction
					results{
						uid
					}
					ts
					parent_heuristic{
						uid
					}
					~parent_heuristic{
						uid
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(),
		query, map[string]string{"$uid": heuristicUid})

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
			  }
				`

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(),
		query, map[string]string{"$uid": heuristicUid})

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

// Returns all origins which are created after the specified date
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$uid": uid, "$ts": timestamp})
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

// GetOriginsByUid searches for the transaction by uid and returns all found origins.
func GetOriginsByUid(c *dgo.Dgraph, uid string) (origins []HeuristicTransaction, err error) {
	query := `query Q($uid: string){
				var (func: uid($uid)){
					v as origins
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$uid": uid})
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

// GetOrigins searches for the transaction by transaction hash and returns all found origins.
func GetOrigins(c *dgo.Dgraph, txhash string) (origins []HeuristicTransaction, err error) {
	query := `query Q($txhash: string){
				var (func: eq(txhash,$txhash)){
					v as origins
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$txhash": txhash})
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$txhash": tx})
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), "error invalid response")
		return
	}

	for _, t := range r.Transaction {
		if len(t.Block) != 1 || len(t.Outputs) == 0 {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), "error invalid response")
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$txhash": tx})
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), "error invalid response")
		return
	}

	t := r.Transaction[0]
	transaction = HeuristicTransaction{
		Uid:     t.Uid,
		Outputs: t.Outputs,
	}

	return
}

// GetFrontendHeuristic returns all heuristics for a given transaction
func GetFrontendHeuristic(c *dgo.Dgraph, txHash string) (completeHeuristic FrontendHeuristicComplete, err error) {
	query := `query Q($hash: string){
					q(func: eq(txhash,$hash)){
						uid
						~transactions{
							ts
						}
						~h_transaction{
							ts
							type
							parameter
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
					}
				}`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetFrontendContext(), query, map[string]string{"$hash": txHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			Uid   string `json:"uid,omitempty"`
			Block []struct {
				Timestamp string `json:"ts,omitempty"`
			} `json:"~transactions,omitempty"`
			Heuristics []FrontendHeuristic `json:"~h_transaction,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) != 1 || len(r.Transaction[0].Block) != 1 {
		err = errors.New("invalid response from database")
		return
	}

	completeHeuristic = FrontendHeuristicComplete{
		Uid:        r.Transaction[0].Uid,
		Timestamp:  r.Transaction[0].Block[0].Timestamp,
		Heuristics: r.Transaction[0].Heuristics,
	}

	return
}

// GetShortestPathLength returns the number of transactions in the shortest path between the given
// transactions specified by fromUid and toUid. The returned number is the count of transactions in
// the path between fromUid and toUid. A shortest path lik the following would return the number 3.
// Example path: fromUid -> tx1 -> output1 -> tx2 -> output2 -> tx3 -> output3 -> toUid
func GetShortestPathLength(c *dgo.Dgraph, fromUid string, toUid string) (pathLength int, err error) {
	query := fmt.Sprintf(`{
				shortest(from: %s, to: %s){
					tx_inputs
					~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
				}
			  }`, fromUid, toUid)

	resp, err := c.NewReadOnlyTxn().Query(db.GetFrontendContext(), query)
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
