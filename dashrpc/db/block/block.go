package block

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// gets block information from the database
func GetBlock(c *dgo.Dgraph, blockHash string, block *Block) error {

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
					~prevblock { 
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
	vars := make(map[string]string)
	vars["$hash"] = blockHash
	resp, err := tx.QueryWithVars(context.Background(), query, vars)

	if err != nil {
		return err
	}
	var r blockQuery
	err = json.Unmarshal(resp.Json, &r)

	if err != nil {
		return err
	}

	lenQ := len(r.Q)

	if lenQ == 0 {
		return errors.New("no blocks found")
	}

	*block = r.Q[0]
	if lenQ > 1 {
		// found more than one block, which should not be possible
		return errors.New("found more than one block")
	}

	return nil
}

// checks if the given block has all attributes filled
func isBlockComplete(blk Block) bool {
	return blk.Uid != "" && blk.Hash != "" && blk.Id != "" && blk.Timestamp != "" ||
		blk.DType != nil && blk.Transactions != nil && blk.PrevBlock != nil
}

// gets block information from the database and checks if it is complete
func GetCompleteBlock(c *dgo.Dgraph, blockHash string, block *Block) error {
	if err := GetBlock(c, blockHash, block); err != nil {
		return err
	}

	if !isBlockComplete(*block) {
		return errors.New("block not complete")
	}

	return nil
}

// upserts a block
func UpsertBlock(c *dgo.Dgraph, block *Block) (*api.Response, error) {
	(*block).Uid = "uid(v)"
	(*block).DType = []string{"Block"}
	pb, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}

	query := `
		query Q($hash: string) {
			q(func: eq(blockhash, $hash)) {
				v as uid
			}
		}
	`
	vars := make(map[string]string)
	vars["$hash"] = block.Hash

	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Vars:      vars,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	res, err := c.NewTxn().Do(context.Background(), req)
	return res, err
}
