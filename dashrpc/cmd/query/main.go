package main

import (
	heuristic "dashrpc/analytics/heuristics/transaction"
	cli "dashrpc/cmd/cliutil"
	"dashrpc/db"
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
	if len(cliArgs.Logfile) > 0 {
		if f, err := cli.GetLogfile(cliArgs.Logfile); err != nil {
			defer func() {
				if err = f.Close(); err != nil {
					fmt.Println(err)
				}
			}()
		}
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
		// some test transactions
		//bigTransaction := "62cd8a10d62c42fa786bb2d897f48499bcbb58ee697a5e7bc0fe48cdec081efc"
		//biggerTransaction := "d3efe170dc1c1e8db2e8feb6fab76da2d9188176196d69b0055ca8e485233fd3"
		//tx := "7336d112b9a2b838ea6fcedb0d55345308952f4dc67a8ff76ff3eba179ed31d4"
		//other := "fdaad37eb8cd68291cc54089e63b75ad01df3c90d10fde26226875dbefc49cdb"
		//testOneSource := "cfb95252da737464c9b37fcd294e1b19b3903b84ae7dd7a045bcb9765c0fb570"
		//testTransaction2 := "78d7d55ecd30c78ea91bffdff536e9c4476d44aa1e2d874663cfba3a547a0eef"
		//testTransation := "cc48f524a5201715428d25dc79a362a5a0fb21747370f224ca5cd2dc1e616862"
		interestTransaction := "6c3786e2a7b10319b2613236c3f5dbe0179d28e626989279b10c774c8bafeba1"

		// heuristic test
		//if err := heuristic.Exec(dgraph, interestTransaction, "",
		//	heuristic.NewTimeConstraintHeuristic(2*24)); err != nil {
		//	log.Println(err)
		//	return
		//}

		if err := heuristic.Exec(dgraph, interestTransaction, "",
			heuristic.NewTimeAmountConstraintHeuristic(2*24)); err != nil {
			log.Println(err)
			return
		}

		//if err := heuristic.Exec(dgraph, interestTransaction, "",
		//	heuristic.NewAllHeuristics(20*24)); err != nil {
		//	log.Println(err)
		//	return
		//}

		//if err := heuristic.Exec(dgraph, interestTransaction, "", heuristic.NewInputAmountHeuristic()); err != nil {
		//	log.Println(err)
		//	return
		//}

		// origin test
		//origins, err := dban.AnalyzeOriginsAlt(dgraph, cliArgs.TxSearch)
		//if err != nil {
		//	return
		//}
		//
		//log.Println(len(origins))

	} else if len(cliArgs.ClusterAddr) > 0 {
		log.Println("Clustering is not yet implemented")
	}
}
