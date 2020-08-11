package main

import (
	"context"
	"dashrpc"
	cli "dashrpc/cmd/cliutil"
	"dashrpc/db"
	dbstat "dashrpc/db/status"
	"flag"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
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

// creates a http server on the given port
func createServer(port uint, dgraph *dgo.Dgraph) *http.Server {
	// setup REST API
	setupHandlers(dgraph)

	// create server
	srv := &http.Server{
		Addr: ":" + strconv.FormatUint(uint64(port), 10),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln("server error:", err)
		}
	}()

	log.Println("Starting server at endpoint http://localhost", srv.Addr)
	return srv
}

// sends a shutdown signal to the server with a timout of 5 seconds
func shutdownServer(server *http.Server) {
	log.Println("### Shutting down server###")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln("Server Shutdown Failed:", err)
	}
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

	// create server
	srv := createServer(cliArgs.ExplorerServerPort, dgraph)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-done:
		shutdownServer(srv)
	}
}
