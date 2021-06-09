package block

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"time"

	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

// GetBlock gets block information from the database
func GetBlock(c *external.GraphDB, blockHash string) (blk Block, err error) {
	query := `query Q($hash: string) {
				q(func: eq(blockhash, $hash)){
					uid
					id
					ts
					blockhash
					dgraph.type
					prevblock { 
						uid
						blockhash
					}
					transactions{
						uid
						txhash
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query, map[string]string{"$hash": blockHash})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r blockQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

func isBlockIdentifier(field string) bool {
	_, err := strconv.Atoi(field)
	return err == nil
}

// GetFrontendBlock gets verbose block information from the database
func GetFrontendBlock(c *external.GraphDB, blockHash string) (block FrontendBlock, err error) {
	searchProperty := "blockhash"
	if isBlockIdentifier(blockHash) {
		searchProperty = "id"
	}

	query := fmt.Sprintf(`query Q($ident: string){
				v as q(func: eq(%s, $ident))@normalize{
					id: id
					ts: ts
					blockhash: blockhash
					prevblock { 
						prevblockhash: blockhash
					}
					nextblock: ~prevblock { 
						nextblockhash: blockhash
					}
				}
				x(func: uid(v)) @normalize{
					transactions{
						tx: txhash
					}
				}
			  }`, searchProperty)
	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$ident": blockHash})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct

	var r struct {
		Blocks       []FrontendBlock `json:"q,omitempty"`
		Transactions []struct {
			Hash string `json:"tx,omitempty"`
		} `json:"x,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.Blocks) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorBlockNotFound)
		return
	} else if len(r.Blocks) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidResult)
		return
	}

	block = r.Blocks[0]

	for _, e := range r.Transactions {
		block.Transactions = append(block.Transactions, e.Hash)
	}

	return
}

// UpsertBlock upserts a block and the prevBlock relation
func UpsertBlock(c *external.GraphDB, block Block) error {
	block.UID = "uid(v)"
	block.PrevBlock.UID = "uid(x)"
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return err
	}

	query := `query Q($currentHash:string,$prevHash:string){
				current(func: eq(blockhash,$currentHash)){
					v as uid
				}
				previous(func: eq(blockhash,$prevHash)){
					x as uid
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$currentHash": block.Hash, "$prevHash": block.PrevBlock.Hash},
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}
	if err = db.TxWithRetry(c, time.Minute*2, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
}

// GetCount gets the number of blocks in the database
func GetCount(c *external.GraphDB) (uint64, error) {
	return db.GetCount(c, DType)
}
