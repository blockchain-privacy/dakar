package block

import (
	"backend/cmd/cliutil"
	"backend/db"
	"time"

	"encoding/json"
	"fmt"
	"strconv"

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

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$hash": blockHash})

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

// gets block information from the database
func GetBlockById(c *dgo.Dgraph, blockId uint64) (blk Block, err error) {
	query := `query Q($id: string) {
				q(func: eq(id, $id)){
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
			  }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query,
		map[string]string{"$id": strconv.FormatUint(blockId, 10)})

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
	if err != nil {
		return false
	}
	return true
}

// gets verbose block information from the database
func GetFrontendBlock(c *dgo.Dgraph, blockHash string) (block FrontendBlock, err error) {
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
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$ident": blockHash})

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

// updates a block
func UpdateBlock(c *dgo.Dgraph, block Block) error {
	pb, err := json.Marshal(block)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return err
	}

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	if err = db.TxWithRetry(c, time.Minute*5, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
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
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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

	if err = db.TxWithRetry(c, time.Minute*5, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
}

// gets the number of blocks in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
