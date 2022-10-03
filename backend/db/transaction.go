package db

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/external"

	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

// transactionDType is the dgraph database type for the Transaction type
const transactionDType = "Transaction"

// Transaction is the database representation of a blockchain transaction
type Transaction struct {
	UID         string                 `json:"uid,omitempty"`
	PrivacyType *constants.PrivacyType `json:"privacytype,omitempty"`
	Fee         *int64                 `json:"fee,omitempty"`
	Outputs     []Output               `json:"tx_outputs,omitempty"`
	Inputs      []Output               `json:"tx_inputs,omitempty"`
	Hash        string                 `json:"txhash,omitempty"`
	DType       []string               `json:"dgraph.type,omitempty"`
}

func (t *Transaction) String() string {
	output := fmt.Sprintf("UID: %s, Hash: %s", t.UID, t.Hash)

	if t.PrivacyType != nil {
		output += fmt.Sprintf(", Privacy type: %d", *t.PrivacyType)
	}

	if t.Fee != nil {
		output += fmt.Sprintf(", Fee: %d", *t.Fee)
	}

	if t.Outputs != nil {
		output += fmt.Sprintf(", Output count: %d", len(t.Outputs))
	}

	if t.Inputs != nil {
		output += fmt.Sprintf(", Input count: %d", len(t.Inputs))
	}

	return output
}

// SetDType sets the DType for dgraph type recognition
func (t *Transaction) SetDType() {
	t.DType = []string{transactionDType}
}

// CalculateTransactionFee sets the transaction fee based
// on the cumulative amount of inputs and outputs
func (t *Transaction) CalculateTransactionFee() (err error) {
	var amountInputs int64
	for _, e := range t.Inputs {
		if e.Amount == nil {
			return errors.New("amount is not set")
		}
		amountInputs += *e.Amount
	}

	var amountOutputs int64
	for _, e := range t.Outputs {
		if e.Amount == nil {
			return errors.New("amount is not set")
		}
		amountOutputs += *e.Amount
	}

	fee := amountInputs - amountOutputs
	t.Fee = &fee

	return
}

// IsMixingTransaction evaluates the privacy type of the transaction and
// returns true if the transaction is a mixing transaction. Inputs and Outputs are not checked
func (t *Transaction) IsMixingTransaction() bool {
	return t.PrivacyType != nil && t.PrivacyType.IsMixing()
}

// IsDestinationTransaction evaluates the privacy type of the transaction and
// returns true if the transaction is a destination transaction. Inputs and Outputs are not checked
func (t *Transaction) IsDestinationTransaction() bool {
	return t.PrivacyType != nil && t.PrivacyType.IsDestination()
}

type transactionQuery struct {
	Q []Transaction `json:"q"`
}

// FrontendTransactionOutput holds the output data which is exposed to the frontend
type FrontendTransactionOutput struct {
	Amount      *int64  `json:"amount"`
	InputIndex  *uint32 `json:"inputindex,omitempty"`
	OutputIndex *uint32 `json:"outputindex,omitempty"`
	IsCoinbase  bool    `json:"iscoinbase"`
	AddressHash string  `json:"addresshash"`
	SigAsm      string  `json:"sigasm,omitempty"`
	KeyAsm      string  `json:"keyasm,omitempty"`

	// This is data from either the transaction where this output is generated or spent
	PrivacyType    int64  `json:"privacytype,omitempty"`
	Hash           string `json:"txhash,omitempty"`
	BlockTimestamp string `json:"ts,omitempty"`
}

// FrontendTransaction holds the transaction data which is exposed to the frontend
type FrontendTransaction struct {
	Hash           string                      `json:"txhash,omitempty"`
	BlockHash      string                      `json:"bhash,omitempty"`
	Fee            int64                       `json:"fee"`
	PrivacyType    int64                       `json:"privacytype,omitempty"`
	BlockID        uint64                      `json:"bid"`
	BlockTimestamp string                      `json:"bts,omitempty"`
	Outputs        []FrontendTransactionOutput `json:"outputs,omitempty"`
	Inputs         []FrontendTransactionOutput `json:"inputs,omitempty"`
}

func (f FrontendTransaction) String() string {
	return fmt.Sprintf("Hash: %s, BlockHash: %s, BlockID: %d, "+
		"Fee: %d, Privacy type: %d, BlockTimestamp: %s, Output Count: %d, Input Count: %d",
		f.Hash, f.BlockHash, f.BlockID, f.Fee, f.PrivacyType, f.BlockTimestamp, len(f.Outputs), len(f.Inputs))
}

const FrontendTransactionFragments = `
				fragment fOutputTransaction {
					txhash:txhash
					privacytype:privacytype
					~transactions{
						ts:ts
					}
				}
				
				fragment fOutput {
					amount: amount
					inputindex: inputindex
					iscoinbase: iscoinbase
					keyasm: keyasm
					sigasm: sigasm
					~addr_outputs{
						addresshash: addresshash
					}
				}`

type OutputTransactionMapping struct {
	Hash    string   `json:"txhash,omitempty"`
	Outputs []Output `json:"tx_outputs,omitempty"`
}

// GetTransactionsOutputs returns all outputs of each given transaction
func GetTransactionsOutputs(c external.Database, transactionHashes []string) (
	transaction []OutputTransactionMapping, err error) {
	if len(transactionHashes) == 0 {
		return nil, errEmptyRequestArgument
	}

	query := `{
				q(func: eq(txhash,` + CreateCommaArray(transactionHashes) + `)){
					txhash
					tx_outputs{
						uid
						amount
						outputindex
					}
				}
			  }`

	resp, err := ReadOnlyTxWithRetry(c, time.Minute*10, query)

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
		err = errors.New("number of returned transactions does not match number of requested transactions")
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

	resp, err := ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
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
		err = fmt.Errorf("%s: %w", ErrTransactionNotFound, fmt.Errorf("block: %d", blockID))
		return
	}

	transactions = r.Q

	return
}

// GetTransaction returns the transaction specified by the transaction hash
func GetTransaction(c external.Database, txHash string) (transaction Transaction, err error) {
	if txHash == "" {
		err = errEmptyRequestArgument
		return
	}

	const query = `query Q($txhash:string) {
				q(func: eq(txhash,$txhash)){
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

	resp, err := ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$txhash": txHash})
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
		err = ErrTransactionNotFound
		return
	}

	transaction = r.Q[0]

	return
}

// GetOutputAddressCounts returns the number of distinct addresses associated
// with the inputs and outputs of the transaction uid
func GetOutputAddressCounts(c external.Database, uid string) (inputCount uint32, outputcount uint32, err error) {
	if uid == "" {
		err = errEmptyRequestArgument
		return
	}

	const query = `query Q($uid: string){
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

	resp, err := ReadOnlyTxVarWithRetry(c, time.Minute*1, query, map[string]string{"$uid": uid})
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrTransactionNotFound)
		return
	}

	if len(r.Input) > 1 || len(r.Output) > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidResult)
		return
	}

	inputCount = r.Input[0].Count
	outputcount = r.Output[0].Count

	return
}

// GetFrontendTransaction gets transaction information for the frontend
func GetFrontendTransaction(c external.Database, txHash string) (transactions []FrontendTransaction, err error) {
	if txHash == "" {
		err = errEmptyRequestArgument
		return
	}
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

	ctx, cancel := GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			Hash        string                      `json:"txhash,omitempty"`
			PrivacyType *int64                      `json:"privacytype,omitempty"`
			Fee         *int64                      `json:"fee,omitempty"`
			Outputs     []FrontendTransactionOutput `json:"outputs,omitempty"`
			Inputs      []FrontendTransactionOutput `json:"inputs,omitempty"`
			Block       []struct {
				Hash string `json:"blockhash,omitempty"`
				TS   string `json:"ts,omitempty"`
				ID   uint64 `json:"id,omitempty"`
			} `json:"block,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrTransactionNotFound)
		return
	}

	for _, t := range r.Transaction {
		if len(t.Block) == 0 || len(t.Block) != 1 {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidResult)
			return
		}

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

		transactions = append(transactions, FrontendTransaction{
			Hash:           t.Hash,
			PrivacyType:    pType,
			Fee:            fee,
			BlockHash:      t.Block[0].Hash,
			BlockID:        t.Block[0].ID,
			BlockTimestamp: t.Block[0].TS,
			Outputs:        t.Outputs,
			Inputs:         t.Inputs,
		})
	}

	return
}

// GetFrontendTransactionsByUID returns the FrontendTransaction's specified by uid
func GetFrontendTransactionsByUID(c external.Database, txUids []string) (txs []FrontendTransaction, err error) {
	if len(txUids) == 0 {
		err = errEmptyRequestArgument
		return
	}

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
	ctx, cancel := GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, map[string]string{"$uids": CreateCommaArray(txUids)})
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
	if txHash == "" {
		err = errEmptyRequestArgument
		return
	}

	query := `query Q($hash: string){
				q(func: eq(txhash, $hash))@normalize{
					~transactions {
						id:id
					}
			  	}
			   }`

	ctx, cancel := GetFrontendContext()
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrTransactionNotFound)
		return
	}

	for _, tx := range r.Transaction {
		if tx.ID > blockID {
			blockID = tx.ID
		}
	}

	return
}

// UpdateTransactions sends the given transaction updates to the database.
// The transaction uids must be set.
func UpdateTransactions(c external.Database, transactions []Transaction) error {
	if len(transactions) == 0 {
		return errEmptyRequestArgument
	}

	for _, tx := range transactions {
		if tx.UID == "" {
			return errors.New("uid is not set for transaction " + tx.String())
		}
	}

	pb, err := json.Marshal(transactions)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return err
	}

	req := &api.Request{Mutations: []*api.Mutation{{SetJson: pb}}, CommitNow: true}

	if err = TxWithRetry(c, time.Minute*5, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
}

// GetTransactionUID returns the uid of the given transaction
func GetTransactionUID(c external.Database, txHash string) (uid string, err error) {
	if txHash == "" {
		return "", errEmptyRequestArgument
	}

	const query = `query Q($tx:string) {
				q(func: eq(txhash, $tx)){
					uid
				}
			  }`

	resp, err := ReadOnlyTxVarWithRetry(c, time.Second*20, query, map[string]string{"$tx": txHash})

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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrTransactionNotFound)
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

	resp, err := ReadOnlyTxVarWithRetry(c, time.Minute*20, query,
		map[string]string{"$id1": strconv.FormatInt(fromBlockID, 10),
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
