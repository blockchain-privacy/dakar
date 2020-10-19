package transaction

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
	"encoding/json"
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
func GetOriginsByDate(c *dgo.Dgraph, uid string, timestamp string) (uids []string, err error) {
	query := `
			query Q($uid: string,$ts: string){
					var (func: uid($uid))@cascade{
						v as origins{
							~transactions@filter(gt(ts,$ts))
						}
					}
					
					q(func: uid(v)){
						uid
					}
				}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$uid": uid, "$ts": timestamp})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), "error invalid response")
		return
	}

	for _, u := range r.Transaction {
		uids = append(uids, u.Uid)
	}

	return
}

// Returns input transactions of the given transaction
func GetInputTransactions(c *dgo.Dgraph, tx string) (inputTransactions []InputTransaction, err error) {
	query := `
			query Q($txhash: string){
				var (func: eq(txhash,$txhash)){
					tx_inputs{
						v as ~tx_outputs
					}
				}
				
				q(func: uid(v))@normalize{
					uid
					~transactions{
						ts:ts
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
		Transaction []InputTransaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), "error invalid response")
		return
	}

	inputTransactions = r.Transaction

	return
}
