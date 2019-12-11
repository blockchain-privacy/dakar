package main

import (
	"dashrpc"
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

func traverseTx(db *badger.DB, blockHash string) error {
	/*
	return db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				fmt.Printf("key=%s, value=%s\n", k, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	*/
	countBlocks := 0
	countTx := 0
	for {
		block := dashrpc.Block{}
		err := dashrpc.DbGetBlock(db, blockHash, &block)
		if err != nil {
			fmt.Printf("Couldn't find the block with hash %s\n", blockHash)
			break
		}
		countBlocks++
		countTx += len(block.TxHashes)
		blockHash = block.PrevBlockHash.String()
	}

	fmt.Printf("Got %d blocks and %d txs.", countBlocks, countTx)
	return nil
}

//
// Simple utility to browse/lookup the TXs from the badger database
// Work in Progress. NOT WORKING YET.
//
func main() {
	badgerDir := flag.String("db", "/tmp/badger", "badger database location")
	startBlockHash := flag.String("blockhash", "", "Start Block Hash")

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

	traverseTx(db, *startBlockHash)
}
