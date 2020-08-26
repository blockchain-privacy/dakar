package main

import (
	cli "dashrpc/cmd/cliutil"
	"dashrpc/db"
	"dashrpc/db/analytics"
	dbtx "dashrpc/db/transaction"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

const privateSendFilename = "result.csv"

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

// Simple utility to browse/lookup the TXs from the database
func main() {

	cliArgs, err := getExplorerCLIArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	// setup Logging
	if f, err := cli.GetLogfile(cliArgs.Logfile); err != nil {
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

		err := analytics.GetOrigins(dgraph, cliArgs.TxSearch, uint(4))
		if err != nil {
			log.Println(err)
			return
		}

		os.Exit(0)
		err, res := transactionSearch(dgraph, cliArgs.TxSearch, "./"+privateSendFilename)

		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Final map has", len(res), "elements")
	} else if len(cliArgs.TxInfo) > 0 {

		transaction, err := dbtx.GetTransaction(dgraph, cliArgs.TxInfo)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Tx IsPrivacyOrigin:", transaction.IsPrivacyOrigin())
		if transaction.IsPrivacyOrigin() {
			log.Println("Denominations:", transaction.CountOutputDenominations())
		}
		log.Println("Tx isMixingTransaction:", transaction.IsMixing())
		if transaction.IsMixing() {
			log.Println("Denominations on outputs:", transaction.CountOutputDenominations())
			log.Println("Denominations on inputs:", transaction.CountInputDenominations())
		}
		log.Println("Tx IsPrivacyDestination:", transaction.IsPrivacyDestination())
		if transaction.IsPrivacyDestination() {
			log.Println("Denominations on inputs:", transaction.CountInputDenominations())
		}
	} else if len(cliArgs.ClusterAddr) > 0 {
		log.Println("Clustering is not yet implemented")
	}
}
