package main

import (
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics"
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
	cliArgs, err = cli.BuildArgs(cli.Logfile, cli.DBPort, cli.DBHost, cli.ChartDir)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	if len(cliArgs.ChartDir) == 0 {
		flag.PrintDefaults()
		return cliArgs, errors.New("specify output directory for charts")
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

	if len(cliArgs.ChartDir) > 0 {
		ts, dbErr := analytics.GetPrivacyTypeData(dgraph, "mixing")
		if dbErr != nil {
			info(err)
			return
		}

		info(ts[0])

		info(ts[len(ts)-1])
		//for _, t := range ts {
		//	info(t)
		//}
	}

}
