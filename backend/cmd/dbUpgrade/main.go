package main

import (
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/heuristics/transaction"
	"flag"
	"fmt"
	"log"
)

var thisLogger *log.Logger

func initLogger() {
	thisLogger = log.New(log.Writer(), "\033[0;31mdbup\033[0m\t", log.Flags())
}

func info(v ...interface{}) {
	thisLogger.Println(v...)
}

// setup cli
func getExplorerCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.Logfile, cli.DBPort, cli.DBHost)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
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

	isSet, err := db.IsSchemaSet(dgraph)
	if err != nil {
		info(err)
		return
	}

	if !isSet {
		info("Schema is not set")
		return
	}

	// Enable for BTC db upgrade -- START

	//info("classifier migration starting ...")
	//if err := db.AlterSchemaAddClassifier(dgraph); err != nil {
	//	info(err)
	//}
	//info("classifier migration done")
	//
	//info("privacytype deletion starting ...")
	//if err := db.DropAllPrivacyTypes(dgraph); err != nil {
	//	info(err)
	//}
	//info("privacytype deletion done")
	//
	//info("privacytype migration starting ...")
	//if err := db.AlterSchemaChangePrivacyTypePredicate(dgraph); err != nil {
	//	info(err)
	//}
	//info("privacytype migration done")
	//
	//// ---------------- ORIGINS
	//
	//info("origins deletion starting ...")
	//if err := db.DropAllOrigins(dgraph); err != nil {
	//	info(err)
	//}
	//info("origins deletion done")
	//
	//info("setting transaction type starting ...")
	//if err := db.AlterSchemaSetTransactionType(dgraph); err != nil {
	//	info(err)
	//}
	//info("setting transaction type done")
	//
	//// ---------------- ANALYZER
	//
	//info("lastanalysedid deletion starting ...")
	//if err := db.DropLastAnalysedID(dgraph); err != nil {
	//	info(err)
	//}
	//info("lastanalysedid deletion done")
	//
	//info("isanalyzing deletion starting ...")
	//if err := db.DropIsAnalyzing(dgraph); err != nil {
	//	info(err)
	//}
	//info("isanalyzing migration done")
	//
	//info("type AnalyzerStatus deletion starting ...")
	//if err := db.DropTypeAnalyzerStatus(dgraph); err != nil {
	//	info(err)
	//}
	//info("type AnalyzerStatus deletion done")

	// Enable for BTC db upgrade -- END

	info("type adding HeuristicResult starting ...")
	if err := db.AlterSchemaAddHeuristicResult(dgraph); err != nil {
		info(err)
	}
	info("type adding HeuristicResult starting done")

	info("deletion of all heuristics starting ...")
	if err := transaction.DeleteAllHeuristics(dgraph); err != nil {
		info(err)
	}
	info("deletion of all heuristics done")

	info("type adding CMultiInputStatus starting ...")
	if err := db.AlterSchemaAddMultiInputClusteringStatus(dgraph); err != nil {
		info(err)
	}
	info("type adding CMultiInputStatus starting done")

	info("type adding Cluster starting ...")
	if err := db.AlterSchemaAddMultiInputClusterType(dgraph); err != nil {
		info(err)
	}
	info("type adding Cluster starting done")
}
