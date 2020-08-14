package main

import (
	"context"
	cli "dashrpc/cmd/cliutil"
	"dashrpc/db"
	"dashrpc/db/status"
	"dashrpc/processor"
	"dashrpc/rpcclient"
	"errors"
	"flag"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func getCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.Continuous, cli.ResetDB, cli.RpcUser, cli.RpcPassword, cli.StartBlockID,
		cli.StopBlockID, cli.IsPrintStatus, cli.RpcHost, cli.RpcPort, cli.Logfile, cli.IgnoreSafeguard, cli.StartHttpServer, cli.HttpServerPort)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	if !cliArgs.IsPrintStatus && !cliArgs.Continuous && (cliArgs.StartBlockID == 0 || cliArgs.StopBlockID == 0) {
		flag.PrintDefaults()
		err = errors.New("missing block information")
		return
	}

	if cliArgs.Continuous && (cliArgs.StartBlockID > 0 || cliArgs.StopBlockID > 0) {
		flag.PrintDefaults()
		err = errors.New("continuous syncing can not be used together with start and stop block id")
		return
	}

	// startBlockID must be smaller or equal than stopBlockID, as we go forward
	if cliArgs.StartBlockID > cliArgs.StopBlockID {
		flag.PrintDefaults()
		err = errors.New("start must be smaller or equal than stop")
		return
	}

	return
}

// checks if a crawling process is already running
func isCrawling(dgraph *dgo.Dgraph) (bool, error) {
	dbStatus, err := status.GetStatus(dgraph)
	if err != nil {
		// no status information found -> database is completely new
		// and thus no crawling is happening right now
		if errors.Is(err, status.ErrorStatusNotFound) {
			return false, nil
		}

		return true, err
	} else if dbStatus.IsCrawling == nil {
		return true, errors.New("was not able to get crawling status successfully")
	}

	return *dbStatus.IsCrawling, nil
}

// The crawler for the system. It needs to be run prior to using any of the other
// commands that rely on the Dgraph DB to be pre-created.
//
// The crawler traverses the Dash blockchain and creates a Dgraph database entry for each transaction
// starting from a given block, and, working backwards, until a given stop block.
func main() {
	fmt.Printf("Go DashRPC client  %s\nBlock crawler\n\n", processor.VersionString)
	cliArgs, err := getCLIArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	// setup Logging
	if f, err := cli.GetLogfile(cliArgs.Logfile, "crawler"); err == nil {
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

	if cliArgs.IsPrintStatus {
		status.PrintStatus(dgraph)
		return
	}

	if cliArgs.ResetDB {
		err = db.DropAll(dgraph)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("dropped all data")
		err = db.SetupSchema(dgraph)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("setup new schema")
	}

	// check if schema exists
	if isSet, err := db.IsSchemaSet(dgraph); err != nil {
		log.Println(err)
		return
	} else if !isSet {
		log.Println("Schema is not set. Use -reset to create a new schema.")
		return
	}

	if !cliArgs.IgnoreSafeguard {
		if ok, err := isCrawling(dgraph); err != nil {
			log.Println(err)
			return
		} else if ok {
			log.Println("Crawling process is already running. Use -ignoresafeguard to crawl despite this.")
			return
		}
	}

	// Setup the RPC connection
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:       cliArgs.RpcEndpoint,
		User:       cliArgs.RpcUser,
		Pass:       cliArgs.RpcPassword,
		DisableTLS: true,
	})
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	count, err := client.GetBlockCount()
	if err != nil {
		log.Printf("\nError: problem with count() %s\n", err.Error())
		return
	}
	log.Printf("Current block count in the chain of the RPC client: %v\n", count)

	// We will handle CTRL-C and CTRL-Z nicely
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

	ctx, cancelFunc := context.WithCancel(context.Background())

	chStopped := make(chan bool, 1)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			chStopped <- true
		}()
		if cliArgs.Continuous {
			err = processor.ProcessBlocksContinuously(ctx, dgraph, client)
		} else {
			err = processor.ProcessBlockRange(ctx, dgraph, client, cliArgs.StartBlockID, cliArgs.StopBlockID)
		}

		if err != nil {
			log.Println(err)
		}
	}()

	var srv *http.Server
	if cliArgs.StartHttpServer {
		wg.Add(1)
		srv = createServer(&wg, cliArgs.HttpServerPort, dgraph, client)
	}

	var stoppedWorking bool
	select {
	case <-chSignal:
		cancelFunc()
		shutdownServer(srv)
	case <-chStopped:
		cancelFunc()
		stoppedWorking = true
	}

	if cliArgs.StartHttpServer && stoppedWorking {
		// if the crawler stopped working on his own accord, the server is still active at this point
		select {
		case <-chSignal:
			shutdownServer(srv)
		}
	}

	wg.Wait()
}
