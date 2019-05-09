package main

import (
	"dashrpc"
	"dashrpc/rpcclient"
	"fmt"
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
	opts.ValueDir = "/mnt/dash_badger/data"
	opts.Dir = 		"/mnt/dash_badger/data"
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

	//
	//  2019-04-26 06:40
	//
	startingBlockId := uint64(1060000)
	startingBlockHash := "00000000000000132447e6bac9fe0d7d756851450eab29358787dc05d809bf07"


	// 2019-05-05 19:22
	// Block: 1065229
	// 0000000000000015b42d1e661ccffac1128a0fde14ae6ec5ed78f7b16a04820c
	//
	// startingBlockId := 1065229
	// startingBlockHash := "0000000000000015b42d1e661ccffac1128a0fde14ae6ec5ed78f7b16a04820c"

	//
	// Appeared in Dash 126744 (2014-08-28 19:47:52)
	// startingBlockHash := "00000000000d0b8cd2507d6ea244bc7109ff9c979a8653617caaff6df848452d"

	// startingBlockId := 50000
	// startingBlockHash := "00000000000fa6230896498b3cc6f1015456b4512452ead9979f6b43ca0a74dc"

	// 50 block
	// startingBlockHash := "00000f106b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"

	// 100 block
	// startingBlockHash := "00000fcef4b9e3b5aa2371dc7f310a8cc2e27171121d656e77f59464e7c0d400"


	dashrpc.ProcessNewBlocks(db, client, startingBlockHash, startingBlockId)

/*
	block := dashrpc.Block{}
	err = dashrpc.DbGetBlock(db, startingBlockHash, &block)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("Block 50 %v\n", block)

	err = dashrpc.DbGetBlock(db, "0000055d088fd066987aa49312cb75646ff033ac45792b6eb8112f162bd19868", &block)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("Block 49 %v\n", block)

	dashrpc.DbGetBlock(db, "000002ba37383d225302973113377b6d7ab36e60a9bdd03377ddf84d928a043d", &block)
	fmt.Printf("Block 73 %v\n", block)
*/
}