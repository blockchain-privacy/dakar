package transaction

import (
	"backend/cmd/cliutil"
	"backend/db"

	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
)

// GetTransaction gets transaction information from the database
func GetTransaction(c *dgo.Dgraph, txHash string) (transaction Transaction, err error) {
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash)){
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
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$hash": txHash})

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
func GetFrontendTransaction(c *dgo.Dgraph, txHash string) (transaction FrontendTransaction, err error) {
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

	if len(r.Transaction) == 0 || len(r.Transaction[0].Block) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorTransactionNotFound)
		return
	} else if len(r.Transaction) != 1 || len(r.Transaction[0].Block) != 1 || r.Transaction[0].OriginCount == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidResult)
		return
	}

	t := r.Transaction[0]

	// t.Fee can be nil in case we do -start to -stop crawling
	fee := int64(-1)
	if t.Fee != nil {
		fee = *t.Fee
	}

	transaction = FrontendTransaction{
		Hash:           t.Hash,
		PrivacyType:    t.PrivacyType,
		Fee:            fee,
		OriginCount:    *t.OriginCount,
		BlockHash:      t.Block[0].Hash,
		BlockId:        t.Block[0].Id,
		BlockTimestamp: t.Block[0].Ts,
		Outputs:        t.Outputs,
		Inputs:         t.Inputs,
	}

	return
}

// GetTransactionBlockId gets the block id of the transaction
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
	} else if len(r.Transaction) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidResult)
		return
	}

	blockId = r.Transaction[0].Id

	return
}

// gets the number of transactions in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
