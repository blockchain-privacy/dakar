// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// TransactionDType is the Dgraph database type for the Transaction type
const TransactionDType = "Transaction"

// Transaction is the database representation of a blockchain transaction
type Transaction struct {
	UID     string   `json:"uid,omitempty"`
	Type    string   `json:"Transaction.type,omitempty"`
	Fee     *int64   `json:"fee,omitempty"`
	Outputs []Output `json:"tx_outputs,omitempty"`
	Inputs  []Output `json:"tx_inputs,omitempty"`
	Hash    string   `json:"txhash,omitempty"`
	DType   []string `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for Dgraph type recognition
func (t *Transaction) SetDType() {
	t.DType = []string{TransactionDType}
}

// CalculateTransactionFee sets the transaction fee based
// on the cumulative amount of inputs and outputs
func (t *Transaction) CalculateTransactionFee() (err error) {
	var amountInputs int64
	for _, e := range t.Inputs {
		if e.Amount == nil {
			return serror.FromStr("amount is not set")
		}
		amountInputs += *e.Amount
	}

	var amountOutputs int64
	for _, e := range t.Outputs {
		if e.Amount == nil {
			return serror.FromStr("amount is not set")
		}
		amountOutputs += *e.Amount
	}

	t.Fee = new(amountInputs - amountOutputs)

	return
}

type transactionQuery struct {
	Q []Transaction `json:"q"`
}

// FrontendTransactionOutput holds the output data which is exposed to the frontend
type FrontendTransactionOutput struct {
	Amount      *int64 `json:"amount"`
	InputIndex  *int32 `json:"inputindex,omitempty"`
	OutputIndex *int32 `json:"outputindex,omitempty"`
	IsCoinbase  bool   `json:"iscoinbase"`
	AddressHash string `json:"addresshash"`

	// This is data from either the transaction where this output is generated or spent
	TransactionType string `json:"txtype,omitempty"`
	Hash            string `json:"txhash,omitempty"`
	BlockTimestamp  string `json:"ts,omitempty"`

	// set to true if the output should be highlighted on the frontend
	Highlight *bool `json:"highlight,omitempty"`
}

// FrontendTransaction holds the transaction data which is exposed to the frontend
type FrontendTransaction struct {
	Hash           string                      `json:"txhash,omitempty"`
	BlockHash      string                      `json:"bhash,omitempty"`
	Fee            int64                       `json:"fee"`
	Type           string                      `json:"txtype,omitempty"`
	BlockID        int64                       `json:"bid"`
	BlockTimestamp string                      `json:"bts,omitempty"`
	Outputs        []FrontendTransactionOutput `json:"outputs,omitempty"`
	Inputs         []FrontendTransactionOutput `json:"inputs,omitempty"`
}

func (f FrontendTransaction) String() string {
	return fmt.Sprintf("Hash: %s, BlockHash: %s, BlockID: %d, "+
		"Fee: %d, type: %s, BlockTimestamp: %s, Output Count: %d, Input Count: %d",
		f.Hash, f.BlockHash, f.BlockID, f.Fee, f.Type, f.BlockTimestamp, len(f.Outputs), len(f.Inputs))
}

const FrontendTransactionFragments = `
				fragment fOutputTransaction {
					txhash:txhash
					txtype:Transaction.type
					~transactions{
						ts:ts
					}
				}
				
				fragment fOutput {
					amount: amount
					inputindex: inputindex
					iscoinbase: iscoinbase
					~addr_outputs{
						addresshash: addresshash
					}
				}`

type OutputTransactionMapping struct {
	Hash    string   `json:"txhash,omitempty"`
	Outputs []Output `json:"tx_outputs,omitempty"`
}

// GetTransactionsOutputs returns all outputs of each given transaction
func GetTransactionsOutputs(ctx context.Context, c external.Database, transactionHashes []string) (
	transaction []OutputTransactionMapping, err error) {
	if len(transactionHashes) == 0 {
		return nil, serror.New(ErrEmptyRequestArgument)
	}

	for _, t := range transactionHashes {
		if !isValidQueryInput(t) {
			return nil, serror.FromStr("invalid transaction hash")
		}
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

	resp, err := QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return
	}
	var r struct {
		Transactions []OutputTransactionMapping `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(transactionHashes) != len(r.Transactions) {
		err = serror.FromStr("number of returned transactions does not match number of requested transactions")
		return
	}

	return r.Transactions, nil
}

// GetTransactionsByBlock returns the transaction contained in the requested block.
// If transactionTypeFilter is not nil, only transactions with the specified type are returned.
func GetTransactionsByBlock(ctx context.Context, c external.Database, fromBlockID int64,
	toBlockID int64, transactionTypeFilter []string) (transactions []Transaction, err error) {
	var typeFilter string
	if len(transactionTypeFilter) > 0 {
		for _, t := range transactionTypeFilter {
			if typeFilter != "" {
				typeFilter += ","
			}

			typeFilter += "\"" + t + "\""
		}

		typeFilter = "@filter(eq(Transaction.type," + typeFilter + "))"
	}

	query := `query Q($from:int,$to:int) {
				var(func: between(id, $from, $to)){
					txs as transactions` + typeFilter + `
				}

				q(func: uid(txs)){
					uid
					txhash
					fee
					Transaction.type
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

	resp, err := QueryVarWithRetry(ctx, c, query,
		map[string]string{"$from": strconv.FormatInt(fromBlockID, 10),
			"$to": strconv.FormatInt(toBlockID, 10)})

	if err != nil {
		return
	}

	var r transactionQuery
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Q) == 0 {
		err = serror.NewWithContext(ErrTransactionNotFound, "block from", fromBlockID, "block to", toBlockID)
		return
	}

	transactions = r.Q

	return
}

// GetTransaction returns the transaction specified by the transaction hash
func GetTransaction(ctx context.Context, c external.Database, txHash string) (transaction Transaction, err error) {
	if txHash == "" {
		err = serror.New(ErrEmptyRequestArgument)
		return
	}

	const query = `query Q($txhash:string) {
				q(func: eq(txhash,$txhash)){
					uid
					txhash
					fee
					Transaction.type
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

	resp, err := QueryVarWithRetry(ctx, c, query, map[string]string{"$txhash": txHash})
	if err != nil {
		return
	}

	var r transactionQuery
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Q) == 0 {
		err = serror.New(ErrTransactionNotFound)
		return
	}

	transaction = r.Q[0]

	return
}

// GetInputOutputAddressCounts returns the number of distinct addresses associated
// with the inputs and outputs of the transaction uid
func GetInputOutputAddressCounts(ctx context.Context, c external.Database,
	uid string) (inputCount int, outputcount int, err error) {
	if uid == "" {
		err = serror.New(ErrEmptyRequestArgument)
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

	resp, err := QueryVarWithRetry(ctx, c, query, map[string]string{"$uid": uid})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Input []struct {
			Count int `json:"count,omitempty"`
		} `json:"input,omitempty"`
		Output []struct {
			Count int `json:"count,omitempty"`
		} `json:"output,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Input) == 0 || len(r.Output) == 0 {
		err = serror.New(ErrTransactionNotFound)
		return
	}

	if len(r.Input) > 1 || len(r.Output) > 1 {
		err = serror.New(errInvalidResult)
		return
	}

	inputCount = r.Input[0].Count
	outputcount = r.Output[0].Count

	return
}

// GetOutputAddressCounts returns the number of distinct addresses associated
// with the outputs of the transaction uid
func GetOutputAddressCounts(ctx context.Context, c external.Database,
	uid string) (outputcount int, err error) {
	if uid == "" {
		err = serror.New(ErrEmptyRequestArgument)
		return
	}

	const query = `query Q($uid: string){
				var(func: uid($uid)){
					tx_outputs {
						~addr_outputs{
							oa as addresshash
						}
					}
				}
				output(func: uid(oa)){
					count(uid)
				}
			   }`

	resp, err := QueryVarWithRetry(ctx, c, query, map[string]string{"$uid": uid})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Output []struct {
			Count int `json:"count,omitempty"`
		} `json:"output,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Output) == 0 {
		err = serror.New(ErrTransactionNotFound)
		return
	}

	if len(r.Output) > 1 {
		err = serror.New(errInvalidResult)
		return
	}

	outputcount = r.Output[0].Count

	return
}

// GetFrontendTransaction gets transaction information for the frontend
func GetFrontendTransaction(ctx context.Context, c external.Database, txHash string) (transactions []FrontendTransaction, err error) {
	if txHash == "" {
		err = serror.New(ErrEmptyRequestArgument)
		return
	}
	const query = `query Q($hash: string){
				q(func: eq(txhash,$hash)){
					txhash
					Transaction.type
					fee
					inputs: tx_inputs@normalize{
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

	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			Hash    string                      `json:"txhash,omitempty"`
			Type    string                      `json:"Transaction.type,omitempty"`
			Fee     *int64                      `json:"fee,omitempty"`
			Outputs []FrontendTransactionOutput `json:"outputs,omitempty"`
			Inputs  []FrontendTransactionOutput `json:"inputs,omitempty"`
			Block   []struct {
				Hash string `json:"blockhash,omitempty"`
				TS   string `json:"ts,omitempty"`
				ID   int64  `json:"id,omitempty"`
			} `json:"block,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Transaction) == 0 {
		err = serror.New(ErrTransactionNotFound)
		return
	}

	for _, t := range r.Transaction {
		if len(t.Block) == 0 || len(t.Block) != 1 {
			err = serror.New(errInvalidResult)
			return
		}

		// t.Fee should never be nil, but just in case
		fee := int64(-1)
		if t.Fee != nil {
			fee = *t.Fee
		}

		transactions = append(transactions, FrontendTransaction{
			Hash:           t.Hash,
			Type:           t.Type,
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

type AmountTransaction struct {
	Hash         string `json:"txhash,omitempty"`
	Fee          *int64 `json:"fee,omitempty"`
	Type         string `json:"txtype,omitempty"`
	Timestamp    string `json:"ts,omitempty"`
	InputAmount  *int64 `json:"inputAmount,omitempty"`
	OutputAmount *int64 `json:"outputAmount,omitempty"`
}

// GetFrontendTransactionAmounts returns summed up amount values per transaction
func GetFrontendTransactionAmounts(ctx context.Context, c external.Database, txUids []string) (txs []AmountTransaction, err error) {
	if len(txUids) == 0 {
		err = serror.New(ErrEmptyRequestArgument)
		return
	}

	const query = `query Q($uids:string){
			t as var(func: uid($uids)){
				tx_outputs{
					outputAmount as amount
				}
				
				tx_inputs{
					inputAmount as amount
				}
				
				outputSum as sum(val(outputAmount))
				inputSum as sum(val(inputAmount))
			}
			
			q(func: uid(t)){
				txhash
				Transaction.type
				fee
				~transactions{
					ts
				}
				inputAmount:val(inputSum)
				outputAmount:val(outputSum)
			}
		}`

	resp, err := c.Query(ctx, query, map[string]string{"$uids": CreateCommaArray(txUids)})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	var r struct {
		Transactions []struct {
			Hash  string `json:"txhash,omitempty"`
			Fee   *int64 `json:"fee,omitempty"`
			Type  string `json:"Transaction.type,omitempty"`
			Block []struct {
				Timestamp string `json:"ts,omitempty"`
			} `json:"~transactions,omitempty"`
			InputAmount  *int64 `json:"inputAmount,omitempty"`
			OutputAmount *int64 `json:"outputAmount,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}
	txs = make([]AmountTransaction, len(r.Transactions))
	for i, tx := range r.Transactions {
		if len(tx.Block) != 1 {
			err = serror.FromFormat("multiple blocks returned for transaction %s", tx.Hash)
			return
		}
		txs[i] = AmountTransaction{
			Hash:         tx.Hash,
			Fee:          tx.Fee,
			Type:         tx.Type,
			Timestamp:    tx.Block[0].Timestamp,
			InputAmount:  tx.InputAmount,
			OutputAmount: tx.OutputAmount,
		}
	}

	return
}

// GetTransactionUIDMapping returns for each transaction a mapping between transaction UID and transaction hash
func GetTransactionUIDMapping(ctx context.Context, c external.Database, txUids []string) (txs []Transaction, err error) {
	if len(txUids) == 0 {
		err = serror.New(ErrEmptyRequestArgument)
		return
	}

	const query = `query Q($uids:string){
				txs as var(func: uid($uids))
				q(func: uid(txs)){
					uid
					txhash
				}
			  }`

	resp, err := c.Query(ctx, query, map[string]string{"$uids": CreateCommaArray(txUids)})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	var r struct {
		Transactions []Transaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	txs = r.Transactions

	return
}

// GetTransactionBlockID gets the block id of the transaction. If there exist multiple transactions
// with the same hash (e.g. in Bitcoin) the highest blockId is returned
func GetTransactionBlockID(ctx context.Context, c external.Database, txHash string) (blockID int64, err error) {
	if txHash == "" {
		err = serror.New(ErrEmptyRequestArgument)
		return
	}

	query := `query Q($hash: string){
				q(func: eq(txhash, $hash))@normalize{
					~transactions {
						id:id
					}
			  	}
			   }`

	resp, err := c.Query(ctx, query, map[string]string{"$hash": txHash})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			ID int64 `json:"id,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Transaction) == 0 {
		err = serror.New(ErrTransactionNotFound)
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
func UpdateTransactions(ctx context.Context, c external.Database, transactions []Transaction) error {
	if len(transactions) == 0 {
		return serror.New(ErrEmptyRequestArgument)
	}

	for _, tx := range transactions {
		if tx.UID == "" {
			return serror.FromStr("uid is not set for transaction " + tx.Hash)
		}
	}

	pb, err := json.Marshal(transactions)
	if err != nil {
		return serror.New(err)
	}

	return MutationWithRetry(ctx, c, &api.Request{Mutations: []*api.Mutation{{SetJson: pb}}, CommitNow: true})
}

// GetTransactionUID returns the uid of the given transaction
func GetTransactionUID(ctx context.Context, c external.Database, txHash string) (uid string, err error) {
	if txHash == "" {
		return "", serror.New(ErrEmptyRequestArgument)
	}

	const query = `query Q($tx:string) {
					q(func: eq(txhash, $tx)){uid}
				   }`

	resp, err := QueryVarWithRetry(ctx, c, query, map[string]string{"$tx": txHash})
	if err != nil {
		return "", err
	}

	var r struct {
		Q []struct {
			UID string `json:"uid"`
		} `json:"q"`
	}
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Q) == 0 {
		err = serror.New(ErrTransactionNotFound)
		return
	}

	uid = r.Q[0].UID

	return
}

// GetOutputs returns the transaction outputs of the given block range
func GetOutputs(ctx context.Context, c external.Database,
	fromBlockID int64, toBlockID int64) (transactions []Transaction, err error) {
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

	resp, err := QueryVarWithRetry(ctx, c, query,
		map[string]string{"$id1": strconv.FormatInt(fromBlockID, 10),
			"$id2": strconv.FormatInt(toBlockID, 10)})
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Transactions []Transaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	transactions = r.Transactions

	return
}

type OutputCount struct {
	Hash        string `json:"txhash,omitempty"`
	InputCount  int    `json:"inputCount,omitempty"`
	OutputCount int    `json:"outputCount,omitempty"`
}

// GetTransactionOutputCounts returns every transaction in the specified block range with its input and output counts.
// If excludeTransactionType is not empty, transactions matching the given type will be excluded.
func GetTransactionOutputCounts(ctx context.Context, c external.Database,
	fromBlockID int64, toBlockID int64, excludeTransactionType string) ([]OutputCount, error) {
	var filter string
	if excludeTransactionType != "" {
		filter = "@filter(not eq(Transaction.type,\"" + excludeTransactionType + "\"))"
	}

	query := `query Q($id1:int,$id2:int){
				var(func: between(id,$id1, $id2)){
					t as transactions` + filter + `
				}

				q(func:uid(t)){
					txhash
					inputCount:count(tx_inputs)
					outputCount:count(tx_outputs)
					
			  	}
			  }`

	resp, err := QueryVarWithRetry(ctx, c, query,
		map[string]string{"$id1": strconv.FormatInt(fromBlockID, 10), "$id2": strconv.FormatInt(toBlockID, 10)})
	if err != nil {
		return nil, err
	}

	// json struct
	var r struct {
		Query []OutputCount `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Query, nil
}
