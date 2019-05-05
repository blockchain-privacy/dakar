package main

import (
	"dashrpc"
	"dashrpc/rpcclient"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/dgraph-io/badger"
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
	opts.SyncWrites = true
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
	blockHash := "00000000000d0b8cd2507d6ea244bc7109ff9c979a8653617caaff6df848452d"

	startHash, err := chainhash.NewHashFromStr(blockHash)
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
	}

	counter := 0
	lastBlockHash := ""
	// Main loop
	for {
		_, err := dashrpc.DbGetBlock(db, blockHash)
		if err == nil {
			// we have processed the block already, we are done.
			break
		}

		startBlock, err := client.GetBlock(startHash)
		if err != nil {
			fmt.Printf("we have problem with getBlock() %s\n", err.Error())
			break
		}
		block, err := dashrpc.ProcessBlock(db, startBlock, lastBlockHash)
		blockHash = block.PrevBlockHash
		lastBlockHash = block.Hash
		startHash, err = chainhash.NewHashFromStr(blockHash)
		if err != nil {
			fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
		}
		counter++
		if counter % 10000 == 0 {
			fmt.Printf("%vk blocks\n", (counter / 10000))
		}
	}

	fmt.Printf("Processed in total: %v blocks\n", counter)

}
