package block

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// gets block information from the database
func GetBlock(c *dgo.Dgraph, blockHash string) (blk Block, err error) {

	tx := c.NewReadOnlyTxn()
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

	resp, err := tx.QueryWithVars(context.Background(),
		query, map[string]string{"$hash": blockHash})

	if err != nil {
		return blk, err
	}
	var r blockQuery
	err = json.Unmarshal(resp.Json, &r)

	if err != nil {
		return blk, err
	}

	lenQ := len(r.Q)

	if lenQ == 0 {
		return blk, errors.New("no blocks found")
	}

	blk = r.Q[0]
	if lenQ > 1 {
		// found more than one block, which should not be possible
		return blk, errors.New("found more than one block")
	}

	return blk, nil
}

// gets block information from the database and checks if it is complete
func GetCompleteBlock(c *dgo.Dgraph, blockHash string) error {
	block, err := GetBlock(c, blockHash)
	if err != nil {
		return err
	}

	if !block.isComplete() {
		return errors.New("block not complete")
	}

	return nil
}

// upserts a block
func UpsertBlock(c *dgo.Dgraph, block Block) error {
	block.Uid = "uid(v)"
	block.DType = []string{"Block"}

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
	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Vars:      map[string]string{"$hash": block.Hash},
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	_, err = c.NewTxn().Do(context.Background(), req)
	return err
}
