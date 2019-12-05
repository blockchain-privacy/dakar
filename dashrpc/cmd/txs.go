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
		for _,t := range txDetails.Inputs {
			txH = t.TxHash

		}
		rounds++
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
	opts.NumMemtables = 50
	opts.MaxTableSize = 512 << 20
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	findTx(db, *txHash)
}
