package transaction

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// Upserts the given heuristic
func UpsertHeuristic(c *dgo.Dgraph, h Heuristic) (err error) {
	// delete potential already existing edges of h
	if err = deleteHeuristicEdges(c, h); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	h.SetDType()
	h.Uid = "uid(h)"
	h.TxUid = "uid(tx)"

	pb, err := json.Marshal(h)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	query := `
		query Q($txhash: string, $type: string) {
			tx as var(func: eq(txhash, $txhash)){
				h as ~h_transaction@filter(eq(type,$type))
			}
		}
	`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$txhash": h.TxHash, "$type": h.HeuristicType},
		Mutations: []*api.Mutation{{
			SetJson: pb,
			Cond:    `@if(eq(len(tx), 1))`,
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
		query Q($txhash: string, $type: string) {
			var(func: eq(txhash, $txhash)){
				h as ~h_transaction@filter(eq(type,$type))
			}
		}
	`

	mu := &api.Mutation{}
	dgo.DeleteEdges(mu, "uid(h)", "results")

	req := &api.Request{
		Query:     query,
		Vars:      map[string]string{"$txhash": h.TxHash, "$type": h.HeuristicType},
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	if err = db.TxWithRetry(c, db.GetBackendContext(), req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return
}

// Returns all origins which are created after the specified date
func GetOriginsByDate(c *dgo.Dgraph, uid string, timestamp string) (origins []Transaction, err error) {
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
		origins = append(origins, Transaction{
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

// Returns input transactions of the given transaction
func GetInputTransactions(c *dgo.Dgraph, tx string) (inputTransactions []Transaction, err error) {
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
		inputTransactions = append(inputTransactions, Transaction{
			Uid:       t.Uid,
			Timestamp: t.Block[0].Timestamp,
			Outputs:   t.Outputs,
		})
	}

	return
}
