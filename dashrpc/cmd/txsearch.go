package main

import (
	"dashrpc"
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

func searchTx(db *badger.DB, txHash string) (dashrpc.TxDetails, error) {
	tx := dashrpc.TxDetails{}
	hash := txHash
	for {
		err := dashrpc.DbGetTxDetails(db, hash, &tx)
		if err != nil {
			fmt.Println("Error", err)
		}
		if tx.IsCreateDenominationsTx() {
			return tx, nil
		}
		hash = tx.Inputs[0].TxHash
	}
	return tx, nil
}

//
// Simple utility to browse/lookup the TXs from the badger database
// Work in Progress. NOT WORKING YET.
//
func main() {
	badgerDir := flag.String("db", "/tmp/badger", "badger database location")
	txHash := flag.String("txhash", "", "Tx Hash")

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
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := searchTx(db, *txHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(dashrpc.CountDenominations(tx.Inputs))
	fmt.Println("Got TX\n\n%s", tx.String())

}

