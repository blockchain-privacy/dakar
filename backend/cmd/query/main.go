package main

import (
	heuristic "backend/analytics/heuristics/transaction"
	cli "backend/cmd/cliutil"
	"backend/db"

	"errors"
	"flag"
	"fmt"
	"log"
)

var thisLogger *log.Logger

func initLogger() {
	thisLogger = log.New(log.Writer(), "\033[0;31mquery\033[0m\t", log.Flags())
}
func info(v ...interface{}) {
	thisLogger.Println(v)
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
		if f, err := cli.GetLogfile(cliArgs.Logfile); err == nil {
			defer func() {
				if err = f.Close(); err != nil {
					fmt.Println(err)
				}
			}()
		}
	}

	initLogger()

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
		//interestTransaction := "6c3786e2a7b10319b2613236c3f5dbe0179d28e626989279b10c774c8bafeba1"

		// test transactions upper part
		//testOneSource := "cfb95252da737464c9b37fcd294e1b19b3903b84ae7dd7a045bcb9765c0fb570"

		//test transactions lower part
		//twoSources := "78d7d55ecd30c78ea91bffdff536e9c4476d44aa1e2d874663cfba3a547a0eef"
		//fourSources := "cc48f524a5201715428d25dc79a362a5a0fb21747370f224ca5cd2dc1e616862"
		//threeSources := "f0db46cc9ca20502bd8265df9b201b38337511d825a3ffe93bdb708ddbc85b01"
		oneSource := "cdfa16675b1320f84d4bb3569e295cb00bdb2372967eba475785f582a01de05b"

		for i := 2; i < 3; i++ {
			hours := uint32(i * 24)

			typeHX := heuristic.BuildExecutor(heuristic.NewDenominationTypeHeuristic())
			matchHX := heuristic.BuildExecutor(heuristic.NewPerfectMatchHeuristic())
			amountHX := heuristic.BuildExecutor(heuristic.NewAmountHeuristic(), matchHX, typeHX)
			oneSourceHX := heuristic.BuildExecutor(heuristic.NewOneSourceHeuristic(hours), amountHX)

			if err := oneSourceHX.RunAsync(dgraph, oneSource, ""); err != nil {
				log.Println(err)
				return
			}
		}

		//if err := oneSourceHX.Run(dgraph, oneSource, ""); err != nil {
		//	log.Println(err)
		//	return
		//}
		//
		//log.Println("----------")

		//log.Println("----------")
		//if err := oneSourceHX.Run(dgraph, threeSources, ""); err != nil {
		//	log.Println(err)
		//	return
		//}
		//log.Println("----------")
		//if err := oneSourceHX.Run(dgraph, fourSources, ""); err != nil {
		//	log.Println(err)
		//	return
		//}

		// heuristic evaluation
		//cHeuristic, err := transaction.GetFrontendHeuristic(dgraph, oneSource)
		//if err != nil {
		//	log.Println(err)
		//	return
		//}
		//log.Println("Analysed transactions:", cHeuristic.Uid, cHeuristic.Timestamp)
		//
		//inputTransactions, err := transaction.GetInputTransactions(dgraph, interestTransaction)
		//if err != nil {
		//	log.Println(err)
		//	return
		//}
		//
		//for _, it := range inputTransactions {
		//	log.Println(it, it.Timestamp)
		//}
		//
		//for _, h := range cHeuristic.Heuristics {
		//
		//	for _, r := range h.Results {
		//		pathLen, err := transaction.GetShortestPathLength(dgraph, cHeuristic.Uid, r.Uid)
		//		if err != nil {
		//			log.Println(err)
		//			return
		//		}
		//		log.Println("path length:", pathLen)
		//	}
		//}

	} else if len(cliArgs.ClusterAddr) > 0 {
		log.Println("Clustering is not yet implemented")
	}
}
