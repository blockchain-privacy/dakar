package main

import (
	cli "dashrpc/cmd/cliutil"
	"dashrpc/db"
	dban "dashrpc/db/analytics"
	dbtx "dashrpc/db/transaction"
	"errors"
	"flag"
	"fmt"
	"log"
)

func info(v ...interface{}) {
	log.SetPrefix("\033[0;31mquery\033[0m\t")
	log.Println(v)
	log.SetPrefix("")
}

// setup cli
func getExplorerCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.TxSearch, cli.Logfile, cli.TxInfo, cli.ClusterAddr, cli.DBPort, cli.DBHost)

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

	// create dgraph client
	dgraph, c, err := db.CreateClient(cliArgs.DBEndpoint)
	if err != nil {
		info(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			info(err)
		}
	}()

	if len(cliArgs.TxSearch) > 0 {
		bigTransaction := "62cd8a10d62c42fa786bb2d897f48499bcbb58ee697a5e7bc0fe48cdec081efc"
		//smallTransation := cliArgs.TxSearch

		transaction, err := dbtx.GetTransaction(dgraph, bigTransaction)
		if err != nil {
			info(err)
			return
		}

		origins, err := dban.GetOrigins(dgraph, bigTransaction)
		if err != nil {
			info(err)
			return
		}

		fmt.Println("number of origins:", len(origins))
		for _, o := range origins {
			dban.GetShortestPath(dgraph, transaction.Uid, o.Uid)
		}

		log.Println("Information is already available in the crawler tool. Query the database or use the frontend to find origins.")

	} else if len(cliArgs.ClusterAddr) > 0 {
		log.Println("Clustering is not yet implemented")
	}
}
