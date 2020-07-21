package block

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// gets block information from the database
func GetBlock(c *dgo.Dgraph, blockHash string) (Block, error) {

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

	resp, err := tx.QueryWithVars(context.Background(),
		query, map[string]string{"$hash": blockHash})

	if err != nil {
		return Block{}, err
	}
	var r blockQuery
	err = json.Unmarshal(resp.Json, &r)

	if err != nil {
		return Block{}, err
	}

	lenQ := len(r.Q)

	if lenQ == 0 {
		return Block{}, errors.New("no blocks found")
	}

	block := r.Q[0]
	if lenQ > 1 {
		// found more than one block, which should not be possible
		return block, errors.New("found more than one block")
	}

	return block, nil
}

// checks if the given block has all attributes filled
func isBlockComplete(blk Block) bool {
	return blk.Uid != "" && blk.Hash != "" && blk.Id != "" && blk.Timestamp != "" ||
		blk.DType != nil && blk.Transactions != nil && blk.PrevBlock != nil
}

// gets block information from the database and checks if it is complete
func GetCompleteBlock(c *dgo.Dgraph, blockHash string) error {
	block, err := GetBlock(c, blockHash)
	if err != nil {
		return err
	}

	if !isBlockComplete(block) {
		return errors.New("block not complete")
	}

	return nil
}

// upserts a block
func UpsertBlock(c *dgo.Dgraph, block UpsertBlockData) error {
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
		Vars:      map[string]string{"$hash": *block.Hash},
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	_, err = c.NewTxn().Do(context.Background(), req)
	return err
}
