package main

import (
	cli "dashrpc/cmd/cliutil"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"io"
	"log"
	"os"
)

// setup cli
func getExplorerCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.BadgerDirectory, cli.TxSearch, cli.Logfile)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	return cliArgs, err
}

//
// Simple utility to browse/lookup the TXs from the badger database
//
// Work in Progress. NOT WORKING YET.
//
func main() {

	cliArgs, err := getExplorerCLIArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	// setup Logging
	if len(cliArgs.Logfile) > 0 {
		f, err := os.OpenFile(cliArgs.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Error opening log file", err)
			return
		}
		defer func() {
			err = f.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()
		log.SetPrefix("query ")
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions(cliArgs.BadgerDir)
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

	recordFile, err := os.Create("./results.csv")
	if err != nil {
		log.Println("Error while creating the file ::", err)
		return
	}

	// Initialize the writer
	writer := csv.NewWriter(recordFile)
	res := search(db, cliArgs.TxSearch, writer)
	if res == nil {
		log.Println("Result in NIL -- fix it.")
		return
	}

	writer.Flush()       // Writes the buffered data to the writer
	err = writer.Error() // Checks if any error occurred while writing
	if err != nil {
		log.Println("Error while writing to the file ::", err)
		return
	}
	err = recordFile.Close()
	if err != nil {
		log.Println("Error while closing the file ::", err)
		return
	}

	// fmt.Printf("%v\n\n", res)
	log.Printf("Final map has %v elements\n", len(res))
	log.Printf("\n")
}
