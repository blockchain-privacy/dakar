package main

import (
	"dashrpc"
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
)


//
// Simple utility to browse/lookup the TXs from the badger database
// Work in Progress. NOT WORKING YET.
//
func main() {
	badgerDir := flag.String("db", "/tmp/badger", "badger database location")
	addr := flag.String("addr", "Xh6DtjUxHdfSCvftfTvzwbWd8oT68vjb3v",
		"addr to build the cluster for")

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
	defer db.Close()

	dashrpc.ProcessAddressClustering(db, *addr)
}
