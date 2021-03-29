package transaction

import (
	"backend/cmd/cliutil"
	"backend/db"
	"errors"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"strconv"
	"time"

	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
)

// GetTransaction gets transaction information from the database
func GetTransaction(c *dgo.Dgraph, txHash string, blockHash string) (transaction Transaction, err error) {
	query := `query Q($tx:string,$block:string) {
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Second*20, query,
		map[string]string{"$tx": txHash, "$block": blockHash})

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

// GetTransaction gets transaction information from the database by block id
func GetTransactionByBlock(c *dgo.Dgraph, blockId uint64) (transactions []Transaction, err error) {
	query := `query Q($block:string) {
				var(func: eq(id, $block)){
					txs as transactions
				}

				q(func: uid(txs)){
					uid
					txhash
					fee
					privacytype
					tx_inputs{
						uid
						amount
						inputindex
						outputindex
					}
					tx_outputs{
						uid
						amount
						inputindex
						outputindex
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query,
		map[string]string{"$block": strconv.FormatUint(blockId, 10)})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r transactionQuery
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Q) == 0 {
		err = ErrorTransactionNotFound
		return
	}

	transactions = r.Q

	return
}

// GetOutputAddressCounts returns the number of distinct addresses associated with the inputs and outputs of the transaction uid
func GetOutputAddressCounts(c *dgo.Dgraph, uid string) (inputCount uint32, outputcount uint32, err error) {
	query := `query Q($uid: string){
				var(func: uid($uid)){
					tx_inputs {
						~addr_outputs{
							ia as addresshash
						}
					}
					tx_outputs {
						~addr_outputs{
							oa as addresshash
						}
					}
				}
				input(func: uid(ia)){
					count(uid)
				}
				output(func: uid(oa)){
					count(uid)
				}
			   }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*1, query, map[string]string{"$uid": uid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Input []struct {
			Count uint32 `json:"count,omitempty"`
		} `json:"input,omitempty"`
		Output []struct {
			Count uint32 `json:"count,omitempty"`
		} `json:"output,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Input) == 0 || len(r.Output) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorTransactionNotFound)
		return
	}

	if len(r.Input) > 1 || len(r.Output) > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidResult)
		return
	}

	inputCount = r.Input[0].Count
	outputcount = r.Output[0].Count

	return
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
						keyasm: keyasm
						sigasm: sigasm
						~addr_outputs{
							addresshash: addresshash
						}
					}
					outputs: tx_outputs @normalize{
						amount: amount
						outputindex: outputindex
						inputindex: inputindex
						iscoinbase: iscoinbase
						keyasm: keyasm
						sigasm: sigasm
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

// UpdateTransactions sends the given transaction updates to the database.
// The transaction uids must be set.
func UpdateTransactions(c *dgo.Dgraph, transactions []Transaction) error {
	for _, tx := range transactions {
		if len(tx.Uid) == 0 {
			return errors.New("error uid is not set for transaction")
		}
	}

	pb, err := json.Marshal(transactions)
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

// GetOutputTransactions returns the output transactions of the given transactions until the given block height
func GetOutputTransactions(c *dgo.Dgraph, txUids []string,
	blockHeight uint64) (outputTransactions []Transaction, err error) {
	uidList := db.CreateUidList(txUids)

	query := `query Q($uids: string, $bid: string){
				var(func: eq(id,$bid)){t as ts}
				var (func: uid($uids)){
					tx_outputs{
						v as ~tx_outputs@cascade{
							~transaction@filter(le(ts,val(t)))
						}
					}
				}

				q(func: uid(v)){
					uid
					txhash
					fee
					privacytype
					tx_inputs{
						uid
						amount
						inputindex
						outputindex
					}
					tx_outputs{
						uid
						amount
						inputindex
						outputindex
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query,
		map[string]string{"$uids": uidList, "$bid": strconv.FormatUint(blockHeight, 10)})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r transactionQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}
