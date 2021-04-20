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
	thisLogger.Println(v)
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

	info("classifier migration starting ...")
	if err := db.AlterSchemaAddClassifier(dgraph); err != nil {
		info(err)
		return
	}
	info("classifier migration done")

	info("privacytype deletion starting ...")
	if err := db.DropAllPrivacyTypes(dgraph); err != nil {
		info(err)
		return
	}
	info("privacytype deletion done")

	info("privacytype migration starting ...")
	if err := db.AlterSchemaChangePrivacyTypePredicate(dgraph); err != nil {
		info(err)
		return
	}
	info("privacytype migration done")

	//info("origins deletion starting ...")
	//if err := db.DropAllOrigins(dgraph); err != nil {
	//	info(err)
	//	return
	//}
	//info("origins deletion done")
	//
	//info("origins migration starting ...")
	//if err := db.AlterSchemaAddOriginsPredicate(dgraph); err != nil {
	//	info(err)
	//	return
	//}
	//info("origins migration done")

	//info("checkpoint migration starting ...")
	//if err := db.AlterSchemaAddCheckpoints(dgraph); err != nil {
	//	info(err)
	//	return
	//}
	//info("checkpoint migration done")
}
