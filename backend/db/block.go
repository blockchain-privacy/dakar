// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"backend/external"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// blockDType is the dgraph database type for the Block type
const blockDType = "Block"

// Block is the database representation of a block
type Block struct {
	UID          string        `json:"uid,omitempty"`
	Hash         string        `json:"blockhash,omitempty"`
	ID           *int64        `json:"id,omitempty"`
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
		output += ", PrevBlockHash: " + b.PrevBlock.Hash
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
	ID               int64                 `json:"id,omitempty"`
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
		err = serror.FromStr("no blocks found")
		return
	} else if lenQ > 1 {
		// found more than one block, which should not be possible
		err = serror.FromStr("found more than one block")
		return
	}
	blk = bq.Q[0]
	return
}

// GetBlock gets block information from the database
func GetBlock(ctx context.Context, c external.Database, blockHash string) (blk Block, err error) {
	if blockHash == "" {
		err = serror.New(ErrEmptyRequestArgument)
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

	resp, err := QueryVarWithRetry(ctx, c, query, map[string]string{"$hash": blockHash})
	if err != nil {
		return
	}

	var r blockQuery
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	return r.payload()
}

// GetFullBlock gets a full block from the database
func GetFullBlock(ctx context.Context, c external.Database, id int, convertUIDs bool) (blk Block, err error) {
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
						Transaction.type
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
					dgraph.type
				}`

	resp, err := QueryVarWithRetry(ctx, c, query, map[string]string{"$blockID": strconv.Itoa(id)})
	if err != nil {
		return
	}
	var r blockQuery

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	block, err := r.payload()
	if err != nil {
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
func GetFrontendBlock(ctx context.Context, c external.Database, blockHash string, offset int) (block FrontendBlock, err error) {
	if blockHash == "" {
		err = serror.New(ErrEmptyRequestArgument)
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
					Transaction.type
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

	resp, err := c.Query(ctx, query, map[string]string{"$ident": blockHash})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	var r struct {
		Blocks       []FrontendBlock `json:"q,omitempty"`
		Transactions []struct {
			Hash    string                      `json:"txhash,omitempty"`
			Type    string                      `json:"Transaction.type,omitempty"`
			Fee     *int64                      `json:"fee,omitempty"`
			Outputs []FrontendTransactionOutput `json:"outputs,omitempty"`
			Inputs  []FrontendTransactionOutput `json:"inputs,omitempty"`
		} `json:"x,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}

	if len(r.Blocks) == 0 {
		err = serror.New(ErrBlockNotFound)
		return
	} else if len(r.Blocks) != 1 {
		err = serror.New(errInvalidResult)
		return
	}

	block = r.Blocks[0]

	for _, t := range r.Transactions {
		// t.Fee should never be nil, but just in case
		fee := int64(-1)
		if t.Fee != nil {
			fee = *t.Fee
		}

		block.Transactions = append(block.Transactions, FrontendTransaction{
			Hash:           t.Hash,
			Type:           t.Type,
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
func UpsertBlock(ctx context.Context, c external.Database, block Block) error {
	if block.PrevBlock == nil {
		return serror.FromFormat("previous block reference is nil: %v", block)
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
		return serror.New(err)
	}

	query := `query Q($currentHash:string,$prevHash:string){
				current(func: eq(blockhash,$currentHash)){
					v as uid
				}
				previous(func: eq(blockhash,$prevHash)){
					x as uid
				}
			  }`

	return MutationWithRetry(ctx, c, &api.Request{
		Query: query,
		Vars:  map[string]string{"$currentHash": block.Hash, "$prevHash": block.PrevBlock.Hash},
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	})
}

// InsertArbitraryJSON insert the given JSON into the database. No client-side checks are performed.
func InsertArbitraryJSON(ctx context.Context, c external.Database, data []byte) error {
	if len(data) == 0 {
		return serror.New(ErrEmptyRequestArgument)
	}

	return MutationWithRetry(ctx, c, &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: data,
		}},
		CommitNow: true,
	})
}
