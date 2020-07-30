package transaction

import (
	"dashrpc/db"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// gets transaction information from the database
func GetTransaction(c *dgo.Dgraph, txHash string) (transaction Transaction, err error) {
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash)){
					uid
					txhash
					tx_inputs{
						uid
						amount
						inputindex
						outputindex
						iscoinbase
						txtype
					}
					tx_outputs{
						uid
						amount
						inputindex
						outputindex
						iscoinbase
						txtype
					}
				}
			  }
				`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(), query, map[string]string{"$hash": txHash})
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
	transaction.SetDType()

	inputs := (*transaction).Inputs
	outputs := (*transaction).Outputs

	for i := range inputs {
		inputs[i].SetDType()
	}

	for i := range outputs {
		outputs[i].SetDType()
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
	res, err := c.NewTxn().Do(db.GetContext(), req)
	return res, err
}

// gets the number of transactions in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
