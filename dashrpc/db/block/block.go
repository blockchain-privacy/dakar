package block

import (
	"dashrpc/db"
	"encoding/json"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// gets block information from the database
func GetBlock(c *dgo.Dgraph, blockHash string) (blk Block, err error) {
	query := `query Q($hash: string) {
				q(func: eq(blockhash, $hash)){
					uid
					id
					ts
					blockhash
					prevblock { 
						uid
						blockhash
					}
					transactions{
						uid
						txhash
					}
				}
			  }
				`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(),
		query, map[string]string{"$hash": blockHash})

	if err != nil {
		return blk, err
	}
	var r blockQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return blk, err
	}

	return r.payload()
}

// upserts a block
func UpsertBlock(c *dgo.Dgraph, block Block) error {
	block.Uid = "uid(v)"
	block.SetDType()

	for i := range block.Transactions {
		block.Transactions[i].DType = []string{"Transaction"}
		for y := range block.Transactions[i].Inputs {
			block.Transactions[i].Inputs[y].SetDType()
		}
		for y := range block.Transactions[i].Outputs {
			block.Transactions[i].Outputs[y].SetDType()
		}
	}

	pb, err := json.Marshal(block)
	if err != nil {
		return err
	}

	query := `
		query Q($hash: string) {
			q(func: eq(blockhash, $hash)) {
				v as uid
			}
		}
	`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$hash": block.Hash},
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	_, err = c.NewTxn().Do(db.GetContext(), req)
	return err
}
