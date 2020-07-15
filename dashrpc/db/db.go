package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"log"
	"time"
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
			txhashes: [uid] .
			Dtype: [string] .
			
			type Block {
				blockhash
				id
				ts
				nextblock
				prevblock
				txhashes
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
func GetBlock(c *dgo.Dgraph, blockHash string, block *DatabaseBlock) error {

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

	var txs []*DataBaseTransaction

	for _, e := range block.TxHashes {
		txs = append(txs, &DataBaseTransaction{
			Hash: e,
		})
	}

	pb, err := json.Marshal(DatabaseBlock{
		Uid:       "uid(v)",
		Hash:      block.Hash.String(),
		Id:        block.Id,
		Timestamp: block.Timestamp.Format(time.RFC3339),
		//NextBlock: &DatabaseBlock{
		//	Hash: block.NextBlockHash.String(),
		//},
		PrevBlock: &DatabaseBlock{
			Hash: block.PrevBlockHash.String(),
		},
		Transactions: txs,
		DType:        block.DType,
	})
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
			blockhash
			prevblock
			nextblock
		}
	`, block.Hash.String())

	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	// Update email only if matching uid found.
	res, err := c.NewTxn().Do(context.Background(), req)
	return res, err
}

func InsertTestData(c *dgo.Dgraph) (*api.Response, error) {
	// While setting an object if a struct has a Uid then its properties in the graph are updated
	// else a new node is created.
	// In the example below new nodes for Alice, Bob and Charlie and school are created (since they
	// dont have a Uid).

	hash, err := chainhash.NewHashFromStr("0000000000000001f2d21729627da94388739075ab220a8f179d2c5f38e78df1")
	if err != nil {
		return nil, err
	}

	tx := Transaction{
		Uid:   "_:test",
		Hash:  "3476d5be24076b8be345c553e298a013a4b5df309ad5f0809efa699373c3bd36",
		DType: []string{"Transaction"},
		Block: Block{
			Hash:  *hash,
			DType: []string{"Block"},
		},
	}

	pb, err := json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}
	return c.NewTxn().Mutate(context.Background(), &api.Mutation{
		SetJson:   pb,
		CommitNow: true,
	})
}

func test() {
	c, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}

	err = DropAll(c)
	if err != nil {
		log.Fatal(err)
	}

	err = SetupSchema(c)
	if err != nil {
		log.Fatal(err)
	}
}
