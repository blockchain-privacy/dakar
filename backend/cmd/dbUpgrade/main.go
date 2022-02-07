package main

import (
	cli "backend/cmd/cliutil"
	"backend/db"
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

type Config struct {
	Logfile string `yaml:"logfile"`
	Host    string `yaml:"host"`
	Port    uint   `yaml:"port"`
}

var defaultConfig = Config{
	Logfile: "",
	Host:    "0.0.0.0",
	Port:    9080,
}

// Simple utility to browse/lookup the TXs from the database
func main() {
	////// SET FLAGS //////

	defaultConfigName := "config.yml"
	var filePath string
	var createConfigFile bool
	cli.SetConfigFlags(defaultConfigName, &filePath, &createConfigFile)
	flag.Parse()

	////// CONFIGURATION FILE HANDLING //////

	if createConfigFile {
		fmt.Println("Generating configuration file ...")

		err := cli.WriteConfig(defaultConfigName, defaultConfig)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("config file", defaultConfigName, "successfully created")
		return
	}

	var config Config
	if err := cli.ReadConfig(filePath, &config); err != nil {
		fmt.Println(err)
		return
	}

	// setup Logging
	if len(config.Logfile) > 0 {
		if f, err := cli.GetLogfile(config.Logfile); err == nil {
			defer func() {
				if err = f.Close(); err != nil {
					fmt.Println(err)
				}
			}()
		}
	}

	initLogger()

	endpoint, err := cli.BuildEndpoint(config.Host, config.Port)
	if err != nil {
		info(err)
		return
	}

	// create dgraph client
	dgraph, c, err := db.CreateClient(endpoint)
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

	//info("removing lowest block id from type starting ...")
	//if err := db.AlterSchemaRemoveLowestBlockIdFromCrawlerStatus(dgraph); err != nil {
	//	info(err)
	//}
	//info("removing lowest block id from type starting  done")
	//
	//info("drop lowest block id predicate starting ...")
	//if err := db.DropLowestBlockId(dgraph); err != nil {
	//	info(err)
	//}
	//info("drop lowest block id predicate done")

	//info("add user predicate starting ...")
	//if err := db.AlterSchemaAddUserToCluster(dgraph); err != nil {
	//	info(err)
	//}
	//info("add user predicate done")

	//info("add attribution type starting ...")
	//if err := db.AlterSchemaAddAttribution(dgraph); err != nil {
	//	info(err)
	//}
	//info("add attribution type done")

	///

	//info("drop heuristic predicates starting ...")
	//if err := db.DropAllHeuristicPredicates(dgraph); err != nil {
	//	info(err)
	//}
	//info("drop heuristic predicates done")
	//
	//info("drop type TransactionHeuristicResult starting ...")
	//if err := db.DropTypeTransactionHeuristicResult(dgraph); err != nil {
	//	info(err)
	//}
	//info("drop type TransactionHeuristicResult done")
	//
	//info("add type HeuristicResult starting ...")
	//if err := db.AlterSchemaAddNewHeuristicResult(dgraph); err != nil {
	//	info(err)
	//}
	//info("drop type HeuristicResult done")

	/////

	info("drop heuristic predicates 2 starting ...")
	if err := db.DropAllHeuristicPredicates2(dgraph); err != nil {
		info(err)
	}
	info("drop heuristic predicates 2 done")

	info("drop type TransactionHeuristic starting ...")
	if err := db.DropTypeTransactionHeuristic(dgraph); err != nil {
		info(err)
	}
	info("drop type TransactionHeuristic done")

	info("add type Heuristic starting ...")
	if err := db.AlterSchemaAddHeuristic(dgraph); err != nil {
		info(err)
	}
	info("drop type Heuristic done")

}
