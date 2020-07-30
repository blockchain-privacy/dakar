package main

import (
	cli "dashrpc/cmd/cliutil"
	"dashrpc/db"
	"errors"
	"flag"
	"fmt"
	"log"
)

// setup cli
func getExplorerCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.TxSearch, cli.Logfile, cli.TxInfo, cli.ClusterAddr)

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
	if f, err := cli.GetLogfile(cliArgs.Logfile, "query"); err != nil {
		defer func() {
			if err = f.Close(); err != nil {
				fmt.Println(err)
			}
		}()
	}

	dgraph, c, err := db.CreateDefaultClient()
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		if err = c.Close(); err != nil {
			log.Println(err)
		}
	}()

	if len(cliArgs.TxSearch) > 0 {
		err, res := transactionSearch(dgraph, cliArgs.TxSearch, "./result.csv")

		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Final map has %v elements\n", len(res))
	}
	//else if len(cliArgs.TxInfo) > 0 {
	//	var txDetails dashrpc.TxDetails
	//	err = dashrpc.DbGetTxDetails(dgraph, cliArgs.TxInfo, &txDetails)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	log.Printf("Tx isCreateDenominations %v\n", txDetails.IsCreateDenominations())
	//	if txDetails.IsCreateDenominations() {
	//		log.Printf("Denominations: %v\n", dashrpc.CountDenominations(txDetails.Outputs))
	//	}
	//	log.Printf("Tx isMixingTransaction %v\n", txDetails.IsMixing())
	//	if txDetails.IsMixing() {
	//		log.Printf("Denominations on outputs: %v\n", dashrpc.CountDenominations(txDetails.Outputs))
	//		log.Printf("Denominations on inputs: %v\n", dashrpc.CountDenominations(txDetails.Inputs))
	//	}
	//	log.Printf("Tx isPrivateSend %v\n", txDetails.IsPrivateSend())
	//	if txDetails.IsPrivateSend() {
	//		log.Printf("Denominations on inputs: %v\n", dashrpc.CountDenominations(txDetails.Inputs))
	//	}
	//} else if len(cliArgs.ClusterAddr) > 0 {
	//	if err := dashrpc.ProcessAddressClustering(dgraph, cliArgs.ClusterAddr); err != nil {
	//		log.Println(err)
	//		return
	//	}
	//}
}
