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
