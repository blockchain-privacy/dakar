package main

import (
	"dashrpc"
	"dashrpc/rpcclient"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/dgraph-io/badger"
	"log"
)

func main() {
	fmt.Println("Go DashRPC client, v.0.0.1")

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       "0.0.0.0:9998",
		User:       "rpc1user",
		Pass:       "1234pass",
		DisableTLS: true,
	}
	var client, err = rpcclient.New(&conn)
	if err != nil {
		fmt.Printf("we have a problem: %s\n", err.Error())
	}

	// Setup the Badger DB connection
	opts := badger.DefaultOptions
	opts.NumVersionsToKeep = 1
	opts.SyncWrites = false
	opts.ValueDir = "/tmp/research/badger"
	opts.Dir = "/tmp/research/badger"
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	count, err := client.GetBlockCount()
	if err != nil {
		fmt.Printf("we have problem with count() %s\n", err.Error())
	}
	fmt.Printf("Current block count in the chain is: %v\n", count)

	err = db.View(func(txn *badger.Txn) error {
		count, err := txn.Get([]byte(dashrpc.DB_BLOCK_COUNT))
		if err != nil {
			fmt.Printf("Problems with txn.Get(DB_BLOCK_COUNT: %s\n", err.Error())
		}
		fmt.Printf("Current block count in DB: %v\n", count)
		return nil
	})

	// hardcoded starting point.
	// we will go back, until we re-connect with the DB
	//
	// 2019-05-05 19:22
	// Block: 1065229
	// 0000000000000015b42d1e661ccffac1128a0fde14ae6ec5ed78f7b16a04820c
	//
	//blockHash := "0000000000000015b42d1e661ccffac1128a0fde14ae6ec5ed78f7b16a04820c"

	//
	// Appeared in Dash 126744 (2014-08-28 19:47:52)
	// startingBlockHash := "00000000000d0b8cd2507d6ea244bc7109ff9c979a8653617caaff6df848452d"

	// 50000 block
	// startingBlockHash := "00000000000fa6230896498b3cc6f1015456b4512452ead9979f6b43ca0a74dc"

	// 50 block
	startingBlockHash := "00000f106b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"
	//dashrpc.ProcessNewBlocks(db, client, startingBlockHash)
	block := dashrpc.Block{}
	err = dashrpc.DbGetBlock(db, startingBlockHash, &block)
	if err == nil {
		// we have processed the block already, we are done.
		fmt.Printf("we have found the block in DB! %v\n", block)
	}

	h, err1 := chainhash.NewHashFromStr(startingBlockHash)
	fmt.Printf("string hash: %s, value hash: %v", startingBlockHash, h)
	startBlock, err := client.GetBlock(h)
	if err1 != nil || err != nil {
		fmt.Printf("we have problem with getBlock() %s -- %s\n", err1.Error(), err.Error())
	}
	var lastBlockHash chainhash.Hash
	blockNew := dashrpc.Block{}
	err = dashrpc.ProcessBlock(db, startBlock, *h, lastBlockHash, &blockNew)
	fmt.Printf("Processed block is: %v", blockNew)

	err = dashrpc.DbGetBlock(db, startingBlockHash, &block)
	if err == nil {
		fmt.Printf("got block after save %v\n", block)
	} else {
		fmt.Printf("Block not saved, problem with hashes!!! %s\n", err.Error())
	}

}
