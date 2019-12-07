package main

import (
	"dashrpc"
	"flag"
	"github.com/dgraph-io/badger"
	"log"
)

func findTx(db *badger.DB, txHash string) {

	rounds := 0
	inputs := 0
	txH := txHash
	var txDetails dashrpc.TxDetails
	for rounds < 16 {
		err := dashrpc.DbGetTxDetails(db, txH, &txDetails)
		if err != nil {
			log.Fatal(err)
		}
		inputs += len(txDetails.Inputs)
		for _, t := range txDetails.Inputs {
			txH = t.TxHash

		}
		rounds++
	}

}

//
// Simple utility to browse/lookup the TXs from the badger database
// Work in Progress. NOT WORKING YET.
//
func main() {
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
		log.Fatal(err)
	}
	defer db.Close()

	findTx(db, *txHash)
}
