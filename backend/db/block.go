package db

import (
	"backend/cmd/cliutil"
	"backend/external"

	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

// blockDType is the dgraph database type for the Block type
const blockDType = "Block"

// Block is the database representation of a block
type Block struct {
	UID          string        `json:"uid,omitempty"`
	Hash         string        `json:"blockhash,omitempty"`
	ID           *uint64       `json:"id,omitempty"`
	Timestamp    string        `json:"ts,omitempty"`
	PrevBlock    *Block        `json:"prevblock,omitempty"`
	Transactions []Transaction `json:"transactions,omitempty"`
	DType        []string      `json:"dgraph.type,omitempty"`
}

func (b *Block) String() string {
	output := fmt.Sprintf("UID: %s, Hash: %s, Timestamp: %s", b.UID, b.Hash, b.Timestamp)

	if b.ID != nil {
		output += fmt.Sprintf(", ID: %d", *b.ID)
	}

	if b.PrevBlock != nil {
		output += fmt.Sprintf(", PrevBlockHash: %s", b.PrevBlock.Hash)
	}

	if b.Transactions != nil {
		output += fmt.Sprintf(", TransactionCount: %d", len(b.Transactions))
	}

	return output
}

// SetDType sets the DType for dgraph type recognition
func (b *Block) SetDType() {
	b.DType = []string{blockDType}
}

// IsComplete checks if the given block has all attributes filled
func (b *Block) IsComplete() bool {
	return b.UID != "" && b.Hash != "" && b.ID != nil && b.Timestamp != "" &&
		b.DType != nil && b.Transactions != nil && b.PrevBlock != nil
}

// FrontendBlock holds all block data which is exposed to the frontend
type FrontendBlock struct {
	Hash             string                `json:"blockhash,omitempty"`
	ID               uint64                `json:"id,omitempty"`
	Timestamp        string                `json:"ts,omitempty"`
	PrevBlockHash    string                `json:"prevblockhash,omitempty"`
	NextBlockHash    string                `json:"nextblockhash,omitempty"`
	TransactionCount int                   `json:"txcount,omitempty"`
	Transactions     []FrontendTransaction `json:"transactions,omitempty"`
}

func (v FrontendBlock) String() string {
	output := fmt.Sprintf("ID: %d, Hash: %s, Timestamp: %s, "+
		"PrevBlockHash: %s, NextBlockHash: %s, transaction count: %d",
		v.ID, v.Hash, v.Timestamp, v.PrevBlockHash, v.NextBlockHash, len(v.Transactions))

	return output
}

type blockQuery struct {
	Q []Block `json:"q"`
}

func (bq blockQuery) payload() (blk Block, err error) {
	lenQ := len(bq.Q)

	if lenQ == 0 {
		err = errors.New("no blocks found")
		return
	} else if lenQ > 1 {
		// found more than one block, which should not be possible
		err = errors.New("found more than one block")
		return
	}
	blk = bq.Q[0]
	return
}

// GetBlock gets block information from the database
func GetBlock(c external.Database, blockHash string) (blk Block, err error) {
	if blockHash == "" {
		err = errEmptyRequestArgument
		return
	}

	const query = `query Q($hash: string) {
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

	resp, err := ReadOnlyTxVarWithRetry(c, time.Minute*20, query, map[string]string{"$hash": blockHash})

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

// GetFullBlock gets a full block from the database
func GetFullBlock(c external.Database, id int, convertUIDs bool) (blk Block, err error) {
	const query = `query Q($blockID: string) {
				q(func: eq(id, $blockID)){
					uid
					id
					ts
					blockhash
					dgraph.type
					prevblock {
						uid
						blockhash
						dgraph.type
					}
					transactions{
						uid
						txhash
						privacytype
						fee
						dgraph.type
						tx_outputs {
							...fOutput
						}
						tx_inputs {
							...fOutput
						}
					}
				}
			  }

				fragment fOutput {
					uid
					amount
					inputindex
					outputindex
					iscoinbase
					keyasm
					sigasm
					txtype
					sighex
					keyhex
					dgraph.type
				}`

	resp, err := ReadOnlyTxVarWithRetry(c, time.Minute*20, query,
		map[string]string{"$blockID": strconv.Itoa(id)})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r blockQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	block, err := r.payload()
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if convertUIDs {
		block.UID = "_:" + block.UID
		block.PrevBlock.UID = "_:" + block.PrevBlock.UID

		for i := range block.Transactions {
			block.Transactions[i].UID = "_:" + block.Transactions[i].UID

			for y := range block.Transactions[i].Outputs {
				block.Transactions[i].Outputs[y].UID = "_:" + block.Transactions[i].Outputs[y].UID
			}

			for y := range block.Transactions[i].Inputs {
				block.Transactions[i].Inputs[y].UID = "_:" + block.Transactions[i].Inputs[y].UID
			}
		}
	}
	blk = block

	return
}

// GetFrontendBlock gets verbose block information from the database
func GetFrontendBlock(c external.Database, blockHash string, offset int) (block FrontendBlock, err error) {
	if blockHash == "" {
		err = errEmptyRequestArgument
		return
	}

	// isBlockIdentifier returns true if field is an integer (block id)
	isBlockIdentifier := func(field string) bool {
		_, err := strconv.Atoi(field)
		return err == nil
	}

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
			  } %s`, searchProperty, offset, FrontendTransactionFragments)
	ctx, cancel := GetFrontendContext()
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
			Hash        string           `json:"txhash,omitempty"`
			PrivacyType *int64           `json:"privacytype,omitempty"`
			Fee         *int64           `json:"fee,omitempty"`
			Outputs     []FrontendOutput `json:"outputs,omitempty"`
			Inputs      []FrontendOutput `json:"inputs,omitempty"`
		} `json:"x,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.Blocks) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrBlockNotFound)
		return
	} else if len(r.Blocks) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidResult)
		return
	}

	block = r.Blocks[0]

	for _, t := range r.Transactions {
		// t.Fee should never be nil, but just in case
		fee := int64(-1)
		if t.Fee != nil {
			fee = *t.Fee
		}

		// t.PrivacyType can be nil
		pType := int64(-1)
		if t.PrivacyType != nil {
			pType = *t.PrivacyType
		}

		block.Transactions = append(block.Transactions, FrontendTransaction{
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
	if block.PrevBlock == nil {
		return fmt.Errorf("previous block reference is nil: %v", block)
	}
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
	if err = TxWithRetry(c, time.Minute*15, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
}

// InsertArbitraryJSON insert the given JSON into the database. No client-side checks are performed.
func InsertArbitraryJSON(c external.Database, data []byte) error {
	if len(data) == 0 {
		return errEmptyRequestArgument
	}

	if err := TxWithRetry(c, time.Minute*15, &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: data,
		}},
		CommitNow: true,
	}); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}
