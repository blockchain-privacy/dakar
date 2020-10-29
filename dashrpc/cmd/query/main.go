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
		// test transactions upper part
		//testOneSource := "cfb95252da737464c9b37fcd294e1b19b3903b84ae7dd7a045bcb9765c0fb570"

		// heuristic test
		//if err := heuristic.Exec(dgraph, interestTransaction, "",
		//	heuristic.NewTimeConstraintHeuristic(2*24)); err != nil {
		//	log.Println(err)
		//	return
		//}

		//if err := heuristic.Exec(dgraph, interestTransaction, "",
		//	heuristic.NewTimeAmountConstraintHeuristic(2*24)); err != nil {
		//	log.Println(err)
		//	return
		//}

		//if err := heuristic.Exec(dgraph, interestTransaction, "0x2eff90",
		//	heuristic.NewOmniSourceHeuristic()); err != nil {
		//	log.Println(err)
		//	return
		//}

		// interest transaction
		//interestTransaction := "6c3786e2a7b10319b2613236c3f5dbe0179d28e626989279b10c774c8bafeba1"
		//if err := heuristic.Exec(dgraph, interestTransaction, "",
		//	heuristic.NewAllHeuristics(7*24)); err != nil {
		//	log.Println(err)
		//	return
		//}

		//test transactions lower part
		//twoSources := "78d7d55ecd30c78ea91bffdff536e9c4476d44aa1e2d874663cfba3a547a0eef"
		//fourSources := "cc48f524a5201715428d25dc79a362a5a0fb21747370f224ca5cd2dc1e616862"
		//threeSources := "f0db46cc9ca20502bd8265df9b201b38337511d825a3ffe93bdb708ddbc85b01"
		oneSource := "cdfa16675b1320f84d4bb3569e295cb00bdb2372967eba475785f582a01de05b"
		hours := uint32(7 * 24)

		if _, err := heuristic.Exec(dgraph, oneSource, "",
			heuristic.NewAllHeuristics(hours)); err != nil {
			log.Println(err)
			return
		}

		//if _, err := heuristic.Exec(dgraph, twoSources, "",
		//	heuristic.NewAllHeuristics(hours)); err != nil {
		//	log.Println(err)
		//	return
		//}
		//
		//if _, err := heuristic.Exec(dgraph, threeSources, "",
		//	heuristic.NewAllHeuristics(hours)); err != nil {
		//	log.Println(err)
		//	return
		//}
		//
		//if _, err := heuristic.Exec(dgraph, fourSources, "",
		//	heuristic.NewAllHeuristics(hours)); err != nil {
		//	log.Println(err)
		//	return
		//}

		/////////////////////// Executor test start

		typeHX := heuristic.BuildExecutor(heuristic.NewDenominationTypeHeuristic())
		matchHX := heuristic.BuildExecutor(heuristic.NewPerfectMatchHeuristic())
		amountHX := heuristic.BuildExecutor(heuristic.NewAmountHeuristic(), matchHX, typeHX)
		oneSourceHX := heuristic.BuildExecutor(heuristic.NewOneSourceHeuristic(hours), amountHX)

		if err := oneSourceHX.Run(dgraph, oneSource, ""); err != nil {
			log.Println(err)
			return
		}

		/////////////////////// Executor test end

		//if err := heuristic.Exec(dgraph, testOneSource, "",
		//	heuristic.NewOneSourceHeuristic(3*24)); err != nil {
		//	log.Println(err)
		//	return
		//}

		//if err := heuristic.Exec(dgraph, testOneSource, "0x37b1e6",
		//	heuristic.NewAmountHeuristic()); err != nil {
		//	log.Println(err)
		//	return
		//}

		//if err := heuristic.Exec(dgraph, testOneSource, "0x37b1e4",
		//	heuristic.NewDenominationTypeHeuristic()); err != nil {
		//	log.Println(err)
		//	return
		//}

		//if err := heuristic.Exec(dgraph, testOneSource, "0x37b1e6",
		//	heuristic.NewPerfectMatchHeuristic()); err != nil {
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
