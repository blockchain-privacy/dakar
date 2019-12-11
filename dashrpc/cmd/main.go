package main

import (
	"dashrpc"
	"dashrpc/rpcclient"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/dgraph-io/badger"
)

//
// DashRPC client traverses the Dash blockchain and creates a Badger database entry for each transaction
// starting from a given block, and, working backwards
//
func main() {
	fmt.Println("Go DashRPC client, v0.0.1")

	badgerDir := flag.String("db", "/tmp/badger", "badger database location")
	rpcUser := flag.String("rpcuser", "rpc1user", "Dash RPC user")
	rpcPassword := flag.String("rpcpassword", "1234pass", "Dash RPC password")
	startBlockID := flag.String("start", "", "Start Block Id")
	stopBlockID := flag.String("stop", "", "Stop Block Id")
	startBlockHash := flag.String("hash", "", "Start Block Hash")

	flag.Parse()

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       "0.0.0.0:9998",
		User:       *rpcUser,
		Pass:       *rpcPassword,
		DisableTLS: true,
	}
	var client, err = rpcclient.New(&conn)
	if err != nil {
		fmt.Printf("we have a problem: %s\n", err.Error())
	}

	// Setup the Badger DB connection
	opts := badger.DefaultOptions(*badgerDir)
	opts.WithNumVersionsToKeep(1)
	opts.WithSyncWrites(false)
	// not needed opts.WithValueDir(*badgerDir)
	// not needed opts.WithDir(*badgerDir)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	count, err := client.GetBlockCount()
	if err != nil {
		fmt.Printf("\nError: problem with count() %s\n", err.Error())
		return
	}

	fmt.Printf("Current block count in the chain: %v\n", count)

	/*
		err = db.View(func(txn *badger.Txn) error {
			count, err := txn.Get([]byte(dashrpc.DB_BLOCK_COUNT))
			if err != nil {
				fmt.Printf("Problems with txn.Get(DB_BLOCK_COUNT: %s\n", err.Error())
			}
			fmt.Printf("Current block count in DB: %v\n", count)
			return nil
		})
	*/

	if *startBlockID == "" || *startBlockHash == "" {
		fmt.Println("\nMissing block ID or block Hash. They need to match!")
		return
	}

	startingBlockId, err := strconv.ParseUint(*startBlockID, 10, 64)
	if err != nil {
		fmt.Printf("\nError %v\n", err.Error())
		return
	}
	startingBlockHash := *startBlockHash

	stopingBlockId, err := strconv.ParseUint(*stopBlockID, 10, 64)
	if err != nil {
		fmt.Printf("\nError %v\n", err.Error())
		return
	}
	//startingBlockId := uint64(1060000)
	//startingBlockHash := "00000000000000132447e6bac9fe0d7d756851450eab29358787dc05d809bf07"

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

	dashrpc.ProcessNewBlocks(db, client, startingBlockHash, startingBlockId, stopingBlockId)

}
