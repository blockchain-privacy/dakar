package transaction

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"

	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

// GetTransaction gets transaction information from the database.
// Use this function if duplicate transaction hashes can not be tolerated.
func GetTransaction(c external.Database, txHash string, blockHash string) (transaction Transaction, err error) {
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query,
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

// GetTransactionsOutputs returns all outputs of each given transaction
func GetTransactionsOutputs(c external.Database, transactionHashes []string) (transaction []OutputTransactionMapping, err error) {
	query := `{
				q(func: eq(txhash,` + db.CreateUIDList(transactionHashes) + `)){
					txhash
					tx_outputs{
						uid
						amount
						outputindex
					}
				}
			  }`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*5, query)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r struct {
		Transactions []OutputTransactionMapping `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(transactionHashes) != len(r.Transactions) {
		err = errors.New("number of returned transaction does not match number of requested transactions")
		return
	}

	return r.Transactions, nil
}

// GetTransactionByBlock gets transaction information from the database by block id
func GetTransactionByBlock(c external.Database, blockID uint64) (transactions []Transaction, err error) {
	const query = `query Q($block:string) {
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$block": strconv.FormatUint(blockID, 10)})
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
		err = fmt.Errorf("%s: %w", ErrorTransactionNotFound, fmt.Errorf("block: %d", blockID))
		return
	}

	transactions = r.Q

	return
}

// GetOutputAddressCounts returns the number of distinct addresses associated with the inputs and outputs of the transaction uid
func GetOutputAddressCounts(c external.Database, uid string) (inputCount uint32, outputcount uint32, err error) {
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorInvalidResult)
		return
	}

	inputCount = r.Input[0].Count
	outputcount = r.Output[0].Count

	return
}

// GetFrontendTransaction gets transaction information for the frontend
func GetFrontendTransaction(c external.Database, txHash string) (transactions []FrontendTransaction, err error) {
	const query = `query Q($hash: string){
				q(func: eq(txhash,$hash)){
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
					block: ~transactions {
						blockhash
						ts
						id
					}
				}
			  }` + FrontendTransactionFragments

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			Hash        string           `json:"txhash,omitempty"`
			PrivacyType *int64           `json:"privacytype,omitempty"`
			Fee         *int64           `json:"fee,omitempty"`
			Outputs     []FrontendOutput `json:"outputs,omitempty"`
			Inputs      []FrontendOutput `json:"inputs,omitempty"`
			Block       []struct {
				Hash string `json:"blockhash,omitempty"`
				Ts   string `json:"ts,omitempty"`
				ID   uint64 `json:"id,omitempty"`
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
		if len(t.Block) == 0 || len(t.Block) != 1 {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorInvalidResult)
			return
		}

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

		transactions = append(transactions, FrontendTransaction{
			Hash:           t.Hash,
			PrivacyType:    pType,
			Fee:            fee,
			BlockHash:      t.Block[0].Hash,
			BlockID:        t.Block[0].ID,
			BlockTimestamp: t.Block[0].Ts,
			Outputs:        t.Outputs,
			Inputs:         t.Inputs,
		})
	}

	return
}

// GetFrontendTransactionsByUID returns the FrontendTransaction's specified by uid
func GetFrontendTransactionsByUID(c external.Database, txUids []string) (txs []FrontendTransaction, err error) {
	const query = `query Q($uids:string){
				txs as var(func: uid($uids))
				q(func: uid(txs))@normalize{
					txhash:txhash
					privacytype:privacytype
					~transactions{
						bid:id
						bts:ts
						bhash:blockhash
					}
				}
			  }`

	// without retry, as this request can easily time out
	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$uids": db.CreateUIDList(txUids)})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transactions []FrontendTransaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	txs = r.Transactions

	return
}

// GetTransactionBlockID gets the block id of the transaction. If there exist multiple transactions
// with the same hash (e.g. in Bitcoin) the highest blockId is returned
func GetTransactionBlockID(c external.Database, txHash string) (blockID uint64, err error) {
	query := `query Q($hash: string){
				q(func: eq(txhash, $hash))@normalize{
					~transactions {
						id:id
					}
			  	}
			   }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			ID uint64 `json:"id,omitempty"`
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
		if tx.ID > blockID {
			blockID = tx.ID
		}
	}

	return
}

// GetCount gets the number of transactions in the database
func GetCount(c external.Database) (uint64, error) {
	return db.GetCount(c, DType)
}

// UpdateTransactions sends the given transaction updates to the database.
// The transaction uids must be set.
func UpdateTransactions(c external.Database, transactions []Transaction) error {
	for _, tx := range transactions {
		if len(tx.UID) == 0 {
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

// GetTransactionUID returns the uid of the given transaction
func GetTransactionUID(c external.Database, txHash string) (uid string, err error) {
	query := `query Q($tx:string) {
				q(func: eq(txhash, $tx)){
					uid
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Second*20, query,
		map[string]string{"$tx": txHash})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Q []struct {
			UID string `json:"uid"`
		} `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Q) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorTransactionNotFound)
		return
	}

	uid = r.Q[0].UID

	return
}

// GetOutputs returns the transaction outputs of the given block range
func GetOutputs(c external.Database, fromBlockID int64, toBlockID int64) (transactions []Transaction, err error) {
	const query = `query Q($id1:int,$id2:int){
					var(func: between(id,$id1, $id2)){
						t as transactions
					}
					
					q(func: uid(t)){
						txhash
						tx_outputs{
							uid
							outputindex
							inputindex
							amount
						}
					}
				}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*10, query, map[string]string{"$id1": strconv.FormatInt(fromBlockID, 10),
		"$id2": strconv.FormatInt(toBlockID, 10)})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transactions []Transaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	transactions = r.Transactions

	return
}
