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
func UpsertHeuristic(c *dgo.Dgraph, h Heuristic) (err error) {
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

	if err = db.TxWithRetry(c, db.GetBackendContext(), req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
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
			  }
				`

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
		Origins []struct {
			Uid       string `json:"uid,omitempty"`
			Timestamp string `json:"ts,omitempty"`
			Inputs    []struct {
				AddressHash string `json:"addresshash,omitempty"`
			} `json:"tx_inputs,omitempty"`
			Outputs []struct {
				Amount int64 `json:"amount,omitempty"`
			} `json:"tx_outputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Origins) == 0 {
		err = errors.New("error invalid response")
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
			Uid:       o.Uid,
			Timestamp: o.Timestamp,
			Address:   o.Inputs[0].AddressHash,
			Outputs:   o.Outputs,
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
					tx_outputs{
						amount
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
		Transaction []struct {
			Uid     string `json:"uid,omitempty"`
			Outputs []struct {
				Amount int64 `json:"amount,omitempty"`
			} `json:"tx_outputs,omitempty"`
			Block []struct {
				Timestamp string `json:"ts,omitempty"`
			} `json:"~transactions,omitempty"`
		} `json:"q,omitempty"`
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
			Uid       string `json:"uid,omitempty"`
			Timestamp string `json:"ts,omitempty"`
			Address   string
			Outputs   []struct {
				Amount int64 `json:"amount,omitempty"`
			} `json:"tx_inputs,omitempty"`
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
		Uid:       t.Uid,
		Timestamp: t.Timestamp,
		Address:   t.Address,
		Outputs:   t.Outputs,
	}

	return
}
