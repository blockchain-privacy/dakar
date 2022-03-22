package block

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/transaction"
	"backend/external"
	"time"

	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

// GetBlock gets block information from the database
func GetBlock(c external.Database, blockHash string) (blk Block, err error) {
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*20, query, map[string]string{"$hash": blockHash})

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

// isBlockIdentifier returns true if field contains a number (block id)
func isBlockIdentifier(field string) bool {
	_, err := strconv.Atoi(field)
	return err == nil
}

// GetFrontendBlock gets verbose block information from the database
func GetFrontendBlock(c external.Database, blockHash string, offset int) (block FrontendBlock, err error) {
	searchProperty := "blockhash"
	if isBlockIdentifier(blockHash) {
		searchProperty = "id"
	}

	query := fmt.Sprintf(`query Q($ident: string){
				q(func: eq(%s, $ident))@normalize{
					id: id
					ts: ts
					blockhash: blockhash
					prevblock { 
						prevblockhash: blockhash
					}
					nextblock: ~prevblock { 
						nextblockhash: blockhash
					}
					txcount: count(transactions)
					t as transactions
				}
				x(func: uid(t), first: 10, offset: %d){
					txhash
					privacytype
					fee
					inputs: tx_inputs @normalize{
						...fOutput
						~tx_outputs {
							...fOutputTransaction
						}
					}
					outputs: tx_outputs @normalize{
						outputindex: outputindex
						...fOutput
						~tx_inputs{
							...fOutputTransaction
						}
					}
				}
			  } %s`, searchProperty, offset, transaction.FrontendTransactionFragments)
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
			Hash        string                       `json:"txhash,omitempty"`
			PrivacyType *int64                       `json:"privacytype,omitempty"`
			Fee         *int64                       `json:"fee,omitempty"`
			Outputs     []transaction.FrontendOutput `json:"outputs,omitempty"`
			Inputs      []transaction.FrontendOutput `json:"inputs,omitempty"`
		} `json:"x,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.Blocks) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrBlockNotFound)
		return
	} else if len(r.Blocks) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrInvalidResult)
		return
	}

	block = r.Blocks[0]

	for _, t := range r.Transactions {
		// t.Fee can be nil in case we do -start to -stop crawling
		fee := int64(-1)
		if t.Fee != nil {
			fee = *t.Fee
		}

		// t.PrivacyType can be nil
		pType := int64(-1)
		if t.PrivacyType != nil {
			pType = *t.PrivacyType
		}

		block.Transactions = append(block.Transactions, transaction.FrontendTransaction{
			Hash:           t.Hash,
			PrivacyType:    pType,
			Fee:            fee,
			BlockHash:      block.Hash,
			BlockID:        block.ID,
			BlockTimestamp: block.Timestamp,
			Outputs:        t.Outputs,
			Inputs:         t.Inputs,
		})
	}

	return
}

// UpsertBlock upserts a block and the prevBlock relation
func UpsertBlock(c external.Database, block Block) error {
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
	if err = db.TxWithRetry(c, time.Minute*15, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
}
