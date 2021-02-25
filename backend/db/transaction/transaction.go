package transaction

import (
	"backend/cmd/cliutil"
	"backend/db"

	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
)

// GetTransaction gets transaction information from the database
func GetTransaction(c *dgo.Dgraph, txHash string, blockHash string) (transaction Transaction, err error) {
	query := `query Q($tx: string,$block:string) {
				blk as var(func: eq(blockhash, $block))

				q(func: eq(txhash, $tx))@filter(uid_in(~transactions,uid(blk))){
					uid
					txhash
					privacytype
					fee
					tx_inputs{
						uid
						amount
						inputindex
						outputindex
						iscoinbase
						txtype
					}
					tx_outputs{
						uid
						amount
						inputindex
						outputindex
						iscoinbase
						txtype
					}
				}
			  }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$tx": txHash, "$block": blockHash})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r transactionQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// GetFrontendTransaction gets transaction information for the frontend
func GetFrontendTransaction(c *dgo.Dgraph, txHash string) (transactions []FrontendTransaction, err error) {
	query := `query Q($hash: string){
				q(func: eq(txhash, $hash)){
					txhash
					privacytype
					fee
					inputs: tx_inputs @normalize{
						amount: amount
						inputindex: inputindex
						iscoinbase: iscoinbase
						~addr_outputs{
							addresshash: addresshash
						}
					}
					outputs: tx_outputs @normalize{
						amount: amount
						outputindex: outputindex
						inputindex: inputindex
						iscoinbase: iscoinbase
						~addr_outputs{
							addresshash: addresshash
						}
					}
					block: ~transactions {
						blockhash
						ts
						id
					}
					origincount: count(origins)
			  	}
			   }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$hash": txHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			Hash        string           `json:"txhash,omitempty"`
			PrivacyType string           `json:"privacytype,omitempty"`
			Fee         *int64           `json:"fee,omitempty"`
			OriginCount *uint64          `json:"origincount,omitempty"`
			Outputs     []FrontendOutput `json:"outputs,omitempty"`
			Inputs      []FrontendOutput `json:"inputs,omitempty"`
			Block       []struct {
				Hash string `json:"blockhash,omitempty"`
				Ts   string `json:"ts,omitempty"`
				Id   uint64 `json:"id,omitempty"`
			} `json:"block,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorTransactionNotFound)
		return
	}

	for _, t := range r.Transaction {
		if len(t.Block) == 0 || len(t.Block) != 1 || t.OriginCount == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidResult)
			return
		}

		// t.Fee can be nil in case we do -start to -stop crawling
		fee := int64(-1)
		if t.Fee != nil {
			fee = *t.Fee
		}

		transactions = append(transactions, FrontendTransaction{
			Hash:           t.Hash,
			PrivacyType:    t.PrivacyType,
			Fee:            fee,
			OriginCount:    *t.OriginCount,
			BlockHash:      t.Block[0].Hash,
			BlockId:        t.Block[0].Id,
			BlockTimestamp: t.Block[0].Ts,
			Outputs:        t.Outputs,
			Inputs:         t.Inputs,
		})
	}

	return
}

// GetTransactionBlockId gets the block id of the transaction. If there exist multiple transactions
// with the same hash (e.g. in Bitcoin) the highest blockId is returned
func GetTransactionBlockId(c *dgo.Dgraph, txHash string) (blockId uint64, err error) {
	query := `query Q($hash: string){
				q(func: eq(txhash, $hash))@normalize{
					~transactions {
						id:id
					}
			  	}
			   }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$hash": txHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			Id uint64 `json:"id,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorTransactionNotFound)
		return
	}

	for _, tx := range r.Transaction {
		if tx.Id > blockId {
			blockId = tx.Id
		}
	}

	return
}

// gets the number of transactions in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
