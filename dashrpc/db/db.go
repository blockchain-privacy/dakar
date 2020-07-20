package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"strconv"
)

// drops ALL data from the database, schema included
func DropAll(c *dgo.Dgraph) error {
	return c.Alter(context.Background(), &api.Operation{
		DropOp: api.Operation_ALL,
	})
}

// create a new dgraph client connecting to the specified host and port
func CreateClient(host string, port uint) (*dgo.Dgraph, error) {
	d, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	return dgo.NewDgraphClient(api.NewDgraphClient(d)), nil
}

// create a new dgraph client with default connection values
func CreateDefaultClient() (*dgo.Dgraph, error) {
	return CreateClient("localhost", 9080)
}

// gets block information from the database
func GetBlock(c *dgo.Dgraph, blockHash string, block *Block) error {

	tx := c.NewReadOnlyTxn()
	query := `query Q($hash: string) {
				q(func: eq(blockhash, $hash)){
					uid
					id
					ts
					blockhash
					prevblock { 
						uid
						blockhash
					}
					~prevblock { 
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

// todo refactor "iscomplete" logic in own function
// gets block information from the database and checks if it is complete
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

// upserts a block
func UpdateBlock(c *dgo.Dgraph, block *Block) (*api.Response, error) {
	(*block).Uid = "uid(v)"
	(*block).DType = []string{"Block"}
	pb, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}

	query := `
		query Q($hash: string) {
			q(func: eq(blockhash, $hash)) {
				v as uid
			}
		}
	`
	vars := make(map[string]string)
	vars["$hash"] = block.Hash

	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Vars:      vars,
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

// todo refactor "iscomplete" logic in own function
// gets transaction information from the database and checks if it is complete
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

// upserts a transaction
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
					addr_outputs{
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

// upserts an address
func UpdateAddress(c *dgo.Dgraph, address *Address) (*api.Response, error) {
	(*address).Uid = "uid(v)"
	(*address).DType = []string{"Address"}
	pb, err := json.Marshal(address)
	if err != nil {
		return nil, err
	}

	query := `
		query Q($hash: string) {
			q(func: eq(addresshash, $hash)) {
				v as uid
			}
		}
	`

	vars := make(map[string]string)
	vars["$hash"] = address.Hash
	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Vars:      vars,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	res, err := c.NewTxn().Do(context.Background(), req)
	return res, err
}

// upserts addresses
func UpdateAddresses(c *dgo.Dgraph, addresses []Address) (*api.Response, error) {
	if addresses == nil {
		return nil, errors.New("got null pointer for addresses")
	}

	// the following block creates the query for 4 addresses the query looks like this:
	//		query Q($h0: string,$h1: string,$h2: string,$h3: string) {
	//		a0 as var(func: eq(addresshash, $h0))
	//		a1 as var(func: eq(addresshash, $h1))
	//		a2 as var(func: eq(addresshash, $h2))
	//		a3 as var(func: eq(addresshash, $h3))
	//		}
	// $h0 ... $h4 are needed to be later replaced. This prevents string injection

	vars := make(map[string]string)
	queryPrefix := "query Q("
	var query string
	// set uid for all addresses and build query
	for i := range addresses {
		queryPrefix += "$h" + strconv.Itoa(i) + ": string"

		if i+1 < len(addresses) {
			queryPrefix += ","
		}

		addresses[i].Uid = fmt.Sprintf("uid(a%d)", i)
		addresses[i].DType = []string{"Address"}
		query += fmt.Sprintf("a%d as var(func: eq(addresshash, $h%d))\n", i, i)
		vars["$h"+strconv.Itoa(i)] = addresses[i].Hash
	}

	queryPrefix += ") {\n"

	pb, err := json.Marshal(addresses)
	if err != nil {
		return nil, err
	}

	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     queryPrefix + query + "}",
		Vars:      vars,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	res, err := c.NewTxn().Do(context.Background(), req)

	if err != nil {
		return res, err
	}

	return res, err
}
