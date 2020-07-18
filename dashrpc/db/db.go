package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

func SetupSchema(c *dgo.Dgraph) error {
	// Install a schema into dgraph. Accounts have a `name` and a `balance`.
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			blockhash: string @index(exact) @upsert .
			txhash: string @index(exact) @upsert .
			addresshash: string @index(exact) @upsert .
			tx_inputs: [uid] @reverse .
			tx_outputs: [uid] @reverse .
			addr_outputs: [uid] @reverse .
			prevblock: uid @reverse .
			transactions: [uid] @reverse .
			id: string .
			ts: dateTime .
			index: string .
			txtype: string .
			amount: string .
			iscoinbase: string .
			

			type Block {
				blockhash
				id
				ts
				prevblock
				transactions
			}

			type Transaction {
				txhash
				ts
				tx_outputs
				tx_inputs
			}

			type TxOutput {
				index
				txtype
				amount
				iscoinbase
			}
			
			type Address {
				addresshash
				addr_outputs
			}
		`,
	})
}

// drops ALL data from the database, schema included
func DropAll(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		DropOp: api.Operation_ALL,
	})
}

func NewClient() (*dgo.Dgraph, error) {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	return dgo.NewDgraphClient(api.NewDgraphClient(d)), nil
}

// gets block information from the database
func GetBlock(c *dgo.Dgraph, blockHash string, block *Block) error {

	tx := c.NewReadOnlyTxn()
	query := `query Q($hash: string) {
				q(func: eq(blockhash, $hash)){
					uid
					id
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
			  }
				`
	vars := make(map[string]string)
	vars["$hash"] = blockHash
	resp, err := tx.QueryWithVars(context.Background(), query, vars)
	if err != nil {
		return err
	}
	var r blockQuery
	err = json.Unmarshal(resp.Json, &r)

	if err != nil {
		return err
	}

	lenQ := len(r.Q)

	if lenQ == 0 {
		return errors.New("no blocks found")
	}

	*block = r.Q[0]
	if lenQ > 1 {
		// found more than one block, which should not be possible
		return errors.New("found more than one block")
	}

	return nil
}

func GetCompleteBlock(c *dgo.Dgraph, blockHash string, block *Block) error {
	if err := GetBlock(c, blockHash, block); err != nil {
		return err
	}

	blk := *block
	if blk.Uid == "" || blk.Hash == "" || blk.Id == "" || blk.Timestamp == "" ||
		blk.DType == nil || blk.Transactions == nil || blk.PrevBlock == nil {
		return errors.New("block not complete")
	}

	return nil
}

func UpdateBlock(c *dgo.Dgraph, block *Block) (*api.Response, error) {
	// While setting an object if a struct has a Uid then its properties in the graph are updated
	// else a new node is created.
	// In the example below new nodes for Alice, Bob and Charlie and school are created (since they
	// dont have a Uid).

	(*block).Uid = "uid(v)"
	(*block).DType = []string{"Block"}
	pb, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		{
			q(func: eq(blockhash, "%s")) {
				...fragmentA
			}
		}

		fragment fragmentA {
			v as uid
		}
	`, block.Hash)

	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	res, err := c.NewTxn().Do(context.Background(), req)
	return res, err
}

// gets transaction information from the database
func GetTransaction(c *dgo.Dgraph, txHash string, transaction *Transaction) error {

	tx := c.NewReadOnlyTxn()
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash)){
					uid
					txhash
					ts
					tx_inputs{
						uid
						amount
						index
						iscoinbase
						txtype
					}
					tx_outputs{
						uid
						amount
						index
						iscoinbase
						txtype
					}
				}
			  }
				`
	vars := make(map[string]string)
	vars["$hash"] = txHash
	resp, err := tx.QueryWithVars(context.Background(), query, vars)
	if err != nil {
		return err
	}
	var r transactionQuery
	err = json.Unmarshal(resp.Json, &r)

	if err != nil {
		return err
	}

	lenQ := len(r.Q)

	if lenQ == 0 {
		return errors.New("no transactions found")
	}

	*transaction = r.Q[0]
	if lenQ > 1 {
		// found more than one transaction, which should not be possible
		return errors.New("found more than one transaction")
	}

	return nil
}

func GetCompleteTransaction(c *dgo.Dgraph, txHash string, transaction *Transaction) error {
	if err := GetTransaction(c, txHash, transaction); err != nil {
		return err
	}

	tx := *transaction
	if tx.Uid == "" || tx.Hash == "" || tx.DType == nil {
		return errors.New("transaction not complete")
	}

	return nil
}

func UpdateTransaction(c *dgo.Dgraph, transaction *Transaction) (*api.Response, error) {
	// variable for upsert
	(*transaction).Uid = "uid(v)"

	// set DType
	(*transaction).DType = []string{"Transaction"}

	inputs := (*transaction).Inputs
	outputs := (*transaction).Outputs

	for i := range inputs {
		inputs[i].DType = []string{"TxOutput"}
	}

	for i := range outputs {
		outputs[i].DType = []string{"TxOutput"}
	}

	// create json
	pb, err := json.Marshal(transaction)
	if err != nil {
		return nil, err
	}

	// build upsert
	query := fmt.Sprintf(`
		{
			q(func: eq(txhash, "%s")) {
				...fragmentA
			}
		}

		fragment fragmentA {
			v as uid
		}
	`, transaction.Hash)
	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	// commit transaction
	res, err := c.NewTxn().Do(context.Background(), req)
	return res, err
}

// gets address information from the database
func GetAddress(c *dgo.Dgraph, txHash string, address *Address) error {

	tx := c.NewReadOnlyTxn()
	query := `query Q($hash: string) {
				q(func: eq(addresshash, $hash)){
					uid
					addresshash
					tx_outputs{
						uid
						amount
						index
						iscoinbase
						txtype
					}
					tx_inputs{
						uid
						amount
						index
						iscoinbase
						txtype
					}
				}
			  }
				`
	vars := make(map[string]string)
	vars["$hash"] = txHash
	resp, err := tx.QueryWithVars(context.Background(), query, vars)
	if err != nil {
		return err
	}
	var r addressQuery
	err = json.Unmarshal(resp.Json, &r)

	if err != nil {
		return err
	}

	lenQ := len(r.Q)

	if lenQ == 0 {
		return errors.New("no addresses found")
	}

	*address = r.Q[0]
	if lenQ > 1 {
		// found more than one address, which should not be possible
		return errors.New("found more than one address")
	}

	return nil
}

func UpdateAddress(c *dgo.Dgraph, address *Address) (*api.Response, error) {
	// While setting an object if a struct has a Uid then its properties in the graph are updated
	// else a new node is created.
	// In the example below new nodes for Alice, Bob and Charlie and school are created (since they
	// dont have a Uid).

	(*address).Uid = "uid(v)"
	(*address).DType = []string{"Address"}
	pb, err := json.Marshal(address)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		{
			q(func: eq(addresshash, "%s")) {
				...fragmentA
			}
		}

		fragment fragmentA {
			v as uid
		}
	`, address.Hash)

	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	res, err := c.NewTxn().Do(context.Background(), req)
	return res, err
}

// todo implement bulk edit
func BulkUpdateAddresses(c *dgo.Dgraph, addresses []Address) (*api.Response, error) {
	if addresses == nil {
		return nil, errors.New("got null pointer for addresses")
	}

	query := "{\n"

	// set uid for all addresses and build query
	for i := range addresses {
		addresses[i].Uid = fmt.Sprintf("uid(a%d)", i)
		query += fmt.Sprintf("a%d as var(func: eq(addresshash, \"%s\"))\n", i, addresses[i].Hash)
	}

	query += "}"

	pb, err := json.Marshal(addresses)
	if err != nil {
		return nil, err
	}

	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	res, err := c.NewTxn().Do(context.Background(), req)

	if err != nil {
		return res, err
	}

	return res, err
}
