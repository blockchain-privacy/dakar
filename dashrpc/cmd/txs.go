package main

import (
	"dashrpc"
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

//
// Simple utility to browse/lookup the TXs from the badger database
// Work in Progress. NOT WORKING YET.
//
func main() {
/*
	rpcUser := flag.String("rpcuser", "rpc1user", "Dash RPC user")
	rpcPassword := flag.String("rpcpassword", "1234pass", "Dash RPC password")

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

	info, err := client.GetBlockChainInfo()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Blocks: ", info.Blocks)
	currentBlock, _ := client.GetBlockCount()
	hash, _ := client.GetBlockHash(currentBlock)
	startBlock, err := client.GetBlock(hash)
	if err != nil {
		fmt.Println(err)
		return
	}
	count := 0
	for {
		count++
		startBlock, err = client.GetBlock(&startBlock.Header.PrevBlock)
		if err != nil {
			fmt.Println(err)
		}
		if count % 100 == 0 {
			fmt.Print(".")
		}
	}
*/

	badgerDir := flag.String("db", "/tmp/badger", "badger database location")
	txHash := flag.String("tx", "2ff76b332b9cb1d22b3eb56e4328964c2e9d69e5e67fd74071cd2823fd7fbdab", "tx to be processed")

	flag.Parse()
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions(*badgerDir)
	// in badger 1.6.0 this is not needed to set explicit anymore
	// opts.Dir = *badgerDir
	// opts.ValueDir = *badgerDir
	opts.WithNumMemtables(50)
	opts.WithMaxTableSize(512 << 20)
	db, err := badger.Open(opts)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func () {
		err := db.Sync()
		//err := db.Close()
		if err != nil {
			fmt.Println("Error closing the Badger DB")
			fmt.Println(err)
		}
	}()

	var txDetails dashrpc.TxDetails
	err = dashrpc.DbGetTxDetails(db, *txHash, &txDetails)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tx isCreateDenominations %v\n", txDetails.IsCreateDenominations())
	if txDetails.IsCreateDenominations() {
		fmt.Printf("Denominations: %v\n", dashrpc.CountDenominations(txDetails.Outputs))
	}
	fmt.Printf("Tx isMixingTransaction %v\n", txDetails.IsMixing())
	if txDetails.IsMixing() {
		fmt.Printf("Denominations on outputs: %v\n", dashrpc.CountDenominations(txDetails.Outputs))
		fmt.Printf("Denominations on inputs: %v\n", dashrpc.CountDenominations(txDetails.Inputs))
	}
	fmt.Printf("Tx isPrivateSend %v\n", txDetails.IsPrivateSend())
	if txDetails.IsPrivateSend() {
		fmt.Printf("Denominations on inputs: %v\n", dashrpc.CountDenominations(txDetails.Inputs))
	}
	//fmt.Printf("Tx details:\n%v\n", txDetails)

}
