package block

import (
	"dashrpc/db"
	"encoding/json"
	"errors"
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

// gets verbose block information from the database
func GetVerboseBlock(c *dgo.Dgraph, blockHash string) (block VerboseBlock, err error) {
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
					nextblock: ~prevblock { 
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
		return
	}

	// json struct
	var r struct {
		Blocks []struct {
			Uid       string `json:"uid,omitempty"`
			Id        uint64 `json:"id,omitempty"`
			Ts        string `json:"ts,omitempty"`
			Hash      string `json:"blockhash,omitempty"`
			PrevBlock struct {
				Uid  string `json:"uid,omitempty"`
				Hash string `json:"blockhash,omitempty"`
			} `json:"prevblock,omitempty"`
			NextBlock []struct {
				Uid  string `json:"uid,omitempty"`
				Hash string `json:"blockhash,omitempty"`
			} `json:"nextblock,omitempty"`
			Transactions []struct {
				Uid  string `json:"uid,omitempty"`
				Hash string `json:"txhash,omitempty"`
			} `json:"transactions,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.Blocks) != 1 || len(r.Blocks[0].NextBlock) > 1 {
		err = errors.New("invalid length of of property in verbose query")
		return
	}

	b := r.Blocks[0]

	var txHashes []string
	for _, e := range b.Transactions {
		txHashes = append(txHashes, e.Hash)
	}

	block = VerboseBlock{
		Uid:           b.Uid,
		Hash:          b.Hash,
		Id:            b.Id,
		Timestamp:     b.Ts,
		PrevBlockHash: b.PrevBlock.Hash,
		Transactions:  txHashes,
	}

	if len(b.NextBlock) == 1 {
		block.NextBlockHash = b.NextBlock[0].Hash
	}

	return
}

// upserts a block and the prevBlock relation
func UpsertBlock(c *dgo.Dgraph, block Block) error {
	block.Uid = "uid(v)"
	block.PrevBlock.Uid = "uid(x)"
	block.SetDType()
	block.PrevBlock.SetDType()

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
		query Q($currentHash: string, $prevHash: string) {
			current(func: eq(blockhash, $currentHash)) {
				v as uid
			}
			previous(func: eq(blockhash, $prevHash)) {
				x as uid
			}
		}
	`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$currentHash": block.Hash, "$prevHash": block.PrevBlock.Hash},
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	_, err = c.NewTxn().Do(db.GetContext(), req)
	return err
}

// inserts a block, the prevBlock relation is done via an upsert
func InsertBlock(c *dgo.Dgraph, block Block) error {
	block.PrevBlock.Uid = "uid(v)"
	block.PrevBlock.SetDType()
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
		Vars:  map[string]string{"$hash": block.PrevBlock.Hash},
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	_, err = c.NewTxn().Do(db.GetContext(), req)
	return err
}

// gets the number of blocks in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
