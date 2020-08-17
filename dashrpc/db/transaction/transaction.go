package transaction

import (
	"dashrpc/cmd/cliutil"
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
					isprivatesend
					iscreatedenominations
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r transactionQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// gets transaction information for the frontend
func GetFrontendTransaction(c *dgo.Dgraph, txHash string) (transaction FrontendTransaction, err error) {
	query := `query Q($hash: string){
				q(func: eq(txhash, $hash)){
					txhash
					isprivatesend
					iscreatedenominations
					inputs: tx_inputs @normalize{
						amount: amount
						inputindex: inputindex
						iscoinbase: iscoinbase
						~addr_outputs{
							addresshash: addresshash
						}
					}
					outputs: tx_outputs @normalize{
						amount: amount
						outputindex: outputindex
						inputindex: inputindex
						iscoinbase: iscoinbase
						~addr_outputs{
							addresshash: addresshash
						}
					}
					block: ~transactions {
						blockhash
						ts
						id
					}
			  	}
			   }`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(), query, map[string]string{"$hash": txHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			Hash                  string           `json:"txhash,omitempty"`
			IsPrivateSend         bool             `json:"isprivatesend,omitempty"`
			IsCreateDenominations bool             `json:"iscreatedenominations,omitempty"`
			Outputs               []FrontendOutput `json:"outputs,omitempty"`
			Inputs                []FrontendOutput `json:"inputs,omitempty"`
			Block                 []struct {
				Hash string `json:"blockhash,omitempty"`
				Ts   string `json:"ts,omitempty"`
				Id   uint64 `json:"id,omitempty"`
			} `json:"block,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 || len(r.Transaction[0].Block) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorTransactionNotFound)
		return
	} else if len(r.Transaction) != 1 || len(r.Transaction[0].Block) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidResult)
		return
	}

	t := r.Transaction[0]

	transaction = FrontendTransaction{
		Hash:                  t.Hash,
		IsPrivateSend:         t.IsPrivateSend,
		IsCreateDenominations: t.IsCreateDenominations,
		BlockHash:             t.Block[0].Hash,
		BlockId:               t.Block[0].Id,
		BlockTimestamp:        t.Block[0].Ts,
		Outputs:               t.Outputs,
		Inputs:                t.Inputs,
	}

	return
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
	pb, err := json.Marshal(transaction)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	return res, err
}

// gets the number of transactions in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
