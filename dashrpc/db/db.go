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
			id: int .
			ts: dateTime .
			nextblock: uid .
			prevblock: uid .
			transactions: [uid] .
			Dtype: [string] .
			
			type Block {
				blockhash
				id
				ts
				nextblock
				prevblock
				transactions
			 }
			type Transaction {
				txhash
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
						blockhash
					}
					txhashes{
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
	if lenQ > 0 {
		*block = r.Q[0]
		if lenQ > 1 {
			err = errors.New("found more than one block")
		}
	}

	return err
}

func InsertBlock(c *dgo.Dgraph, block *Block) (*api.Response, error) {
	// While setting an object if a struct has a Uid then its properties in the graph are updated
	// else a new node is created.
	// In the example below new nodes for Alice, Bob and Charlie and school are created (since they
	// dont have a Uid).

	(*block).Uid = "uid(v)"

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
