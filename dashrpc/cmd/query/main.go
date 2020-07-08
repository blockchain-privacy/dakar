package main

import (
	"dashrpc"
	cli "dashrpc/cmd/cliutil"
	"errors"
	"flag"
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"io"
	"log"
	"os"
)

// setup cli
func getExplorerCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.BadgerDirectory, cli.TxSearch, cli.Logfile, cli.TxInfo, cli.ClusterAddr)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	if len(cliArgs.TxInfo) == 0 && len(cliArgs.TxSearch) == 0 && len(cliArgs.ClusterAddr) == 0 {
		flag.PrintDefaults()
		return cliArgs, errors.New("provide one input hash")
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

	// set maximum number of memtables to 50 (default: 5)
	opts.WithNumMemtables(50)

	// set maximum size of LSM table to 512 MB (default: 64 MB)
	opts.WithMaxTableSize(512 << 20)

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}

	// close database when done
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	if len(cliArgs.TxSearch) > 0 {
		err, res := transactionSearch(db, cliArgs.TxSearch, "./result.csv")

		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Final map has %v elements\n", len(res))
	} else if len(cliArgs.TxInfo) > 0 {
		var txDetails dashrpc.TxDetails
		err = dashrpc.DbGetTxDetails(db, cliArgs.TxInfo, &txDetails)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Tx isCreateDenominations %v\n", txDetails.IsCreateDenominations())
		if txDetails.IsCreateDenominations() {
			log.Printf("Denominations: %v\n", dashrpc.CountDenominations(txDetails.Outputs))
		}
		log.Printf("Tx isMixingTransaction %v\n", txDetails.IsMixing())
		if txDetails.IsMixing() {
			log.Printf("Denominations on outputs: %v\n", dashrpc.CountDenominations(txDetails.Outputs))
			log.Printf("Denominations on inputs: %v\n", dashrpc.CountDenominations(txDetails.Inputs))
		}
		log.Printf("Tx isPrivateSend %v\n", txDetails.IsPrivateSend())
		if txDetails.IsPrivateSend() {
			log.Printf("Denominations on inputs: %v\n", dashrpc.CountDenominations(txDetails.Inputs))
		}
	} else if len(cliArgs.ClusterAddr) > 0 {
		if err := dashrpc.ProcessAddressClustering(db, cliArgs.ClusterAddr); err != nil {
			log.Println(err)
			return
		}
	}
}
