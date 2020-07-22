package transaction

import (
	"context"
	"encoding/json"
	"errors"
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
	err = json.Unmarshal(resp.Json, &r)

	if err != nil {
		return transaction, err
	}

	lenQ := len(r.Q)

	if lenQ == 0 {
		err = errors.New("no transactions found")
		return transaction, err
	}

	transaction, err = r.Q[0].toTransaction()
	if err != nil {
		return transaction, err
	}

	if lenQ > 1 {
		// found more than one transaction, which should not be possible
		err = errors.New("found more than one transaction")
		return transaction, err
	}

	return transaction, err
}

// gets transaction information from the database and checks if it is complete
func GetCompleteTransaction(c *dgo.Dgraph, txHash string) (transaction Transaction, err error) {
	if transaction, err := GetTransaction(c, txHash); err != nil {
		return transaction, err
	}

	if !transaction.isTransactionComplete() {
		err = errors.New("transaction not complete")
		return transaction, err
	}

	return transaction, err
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
