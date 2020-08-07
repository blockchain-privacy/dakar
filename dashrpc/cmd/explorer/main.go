package main

import (
	"dashrpc"
	cli "dashrpc/cmd/cliutil"
	"dashrpc/db"
	dbstat "dashrpc/db/status"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// setup cli
func getExplorerCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.RpcUser, cli.RpcPassword, cli.RpcHost,
		cli.RpcPort, cli.Logfile, cli.IsPrintStatus, cli.ExplorerServerPort)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	return cliArgs, err
}

// Simple web-based utility to browse/lookup tx, block, address and meta data from the badger database
// It provides the API through HTTP
// Work in Progress.
func main() {
	fmt.Printf("Go DashRPC client  %s\nBlock explorer\n\n", dashrpc.VersionString)

	cliArgs, err := getExplorerCLIArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	// setup Logging
	if f, err := cli.GetLogfile(cliArgs.Logfile, "explorer"); err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				fmt.Println(err)
			}
		}()
	}

	// create dgraph client
	dgraph, c, err := db.CreateDefaultClient()
	if err != nil {
		log.Print(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			log.Println(err)
		}
	}()

	// check if schema exists
	if isSet, err := db.IsSchemaSet(dgraph); err != nil {
		log.Println(err)
		return
	} else if !isSet {
		log.Println("Schema is not set.")
		return
	}

	if cliArgs.IsPrintStatus {
		dbstat.PrintStatus(dgraph)
		return
	}

	// setup REST API
	setupHandlers(dgraph)

	// start the server
	log.Printf("Starting server at endpoint http://localhost:%d\n", cliArgs.ExplorerServerPort)
	log.Fatal(http.ListenAndServe(":"+strconv.FormatUint(uint64(cliArgs.ExplorerServerPort), 10), nil))
}
