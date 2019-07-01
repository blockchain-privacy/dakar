package main

import (
	"dashrpc"
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

func findTx(db *badger.DB, txHash string) {

	var txDetails dashrpc.TxDetails
	err := dashrpc.DbGetTxDetails(db, txHash, &txDetails)

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("\nProcessing finished.")
	}
}

func main() {
	badgerDir := flag.String("db", "/tmp/badger", "badger database location")
	txHash := flag.String("tx", "2ff76b332b9cb1d22b3eb56e4328964c2e9d69e5e67fd74071cd2823fd7fbdab", "tx to be processed")

	flag.Parse()
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions
	opts.Dir = *badgerDir
	opts.ValueDir = *badgerDir
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	findTx(db, *txHash)
}
