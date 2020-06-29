package main

import (
	"dashrpc"
	cli "dashrpc/cmd/cliutil"
	"dashrpc/rpcclient"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

// setup cli
func getExplorerCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.BadgerDirectory, cli.RpcUser, cli.RpcPassword, cli.RpcHost,
		cli.RpcPort, cli.Logfile, cli.IsPrintStatus, cli.ExplorerServerPort)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	return cliArgs, err
}

// Simple web-based utility to browse/lookup the TXs from the badger database
// It provides the API through HTTP
// Work in Progress. NOT WORKING YET.
func main() {
	fmt.Printf("Go DashRPC client  %s\nBlock explorer\n\n", dashrpc.VersionString)

	cliArgs, err := getExplorerCLIArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	// setup Logging
	if len(cliArgs.Logfile) > 0 {
		f, err := os.OpenFile(cliArgs.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Error opening log file", err)
			return
		}
		defer func() {
			err = f.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()
		log.SetPrefix("explorer ")
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	db := dashrpc.SetupBadgerDB(cliArgs.BadgerDir)
	defer func() {
		e := db.Close()
		if e != nil { /* ignore */
		}
	}()

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       cliArgs.RpcEndpoint,
		User:       cliArgs.RpcUser,
		Pass:       cliArgs.RpcPassword,
		DisableTLS: true,
	}
	client, err := rpcclient.New(&conn)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	dbBlockCount := dashrpc.DbGetBlockCount(db)
	dbTxCount := dashrpc.DbGetGlobalTxCount(db)
	log.Printf("DB block count: %v  TX count: %v\n", dbBlockCount, dbTxCount)
	if cliArgs.IsPrintStatus {
		dashrpc.PrintStatus(db)
		return
	}

	dbStatus := dashrpc.DbGetStatus(db)
	log.Printf("DB status: %s\n", dbStatus)

	// API end points

	http.HandleFunc(getRouteTransaction(), handlerTxDetails(db, client))
	http.HandleFunc(getRouteAddress(), handlerAddressDetails(db, client))
	http.HandleFunc(getRouteBlock(), handlerBlockDetails(db, client))
	http.HandleFunc(getRouteMeta(), handlerMeta(db, client))
	http.HandleFunc(getRouteRoot(), handlerRoot)

	// start the server
	log.Println("Starting server on port", cliArgs.ExplorerServerPort)
	log.Fatal(http.ListenAndServe(":"+strconv.FormatUint(uint64(cliArgs.ExplorerServerPort), 10), nil))
}
