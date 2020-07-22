package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// gets transaction information from the database
func GetTransaction(c *dgo.Dgraph, txHash string) (transaction Transaction, err error) {

	tx := c.NewReadOnlyTxn()
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash)){
					uid
					txhash
					tx_inputs{
						uid
						amount
						index
						iscoinbase
						txtype
					}
					tx_outputs{
						uid
						amount
						index
						iscoinbase
						txtype
					}
				}
			  }
				`
	vars := make(map[string]string)
	vars["$hash"] = txHash
	resp, err := tx.QueryWithVars(context.Background(), query, vars)
	if err != nil {
		return transaction, err
	}
	var r transactionQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return transaction, err
	}

	return r.payload()
}

// upserts a transaction
func UpsertTransaction(c *dgo.Dgraph, transaction *Transaction) (*api.Response, error) {
	// variable for upsert
	(*transaction).Uid = "uid(v)"

	// set DType
	(*transaction).DType = []string{"Transaction"}

	inputs := (*transaction).Inputs
	outputs := (*transaction).Outputs

	for i := range inputs {
		inputs[i].DType = []string{"TxOutput"}
	}

	for i := range outputs {
		outputs[i].DType = []string{"TxOutput"}
	}

	// create json
	pb, err := json.Marshal(transaction.toUpdate())
	if err != nil {
		return nil, err
	}

	// build upsert
	query := fmt.Sprintf(`
		{
			q(func: eq(txhash, "%s")) {
				v as uid
			}
		}
	`, transaction.Hash)

	req := &api.Request{
		Query: query,
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	// commit transaction
	res, err := c.NewTxn().Do(context.Background(), req)
	return res, err
}
