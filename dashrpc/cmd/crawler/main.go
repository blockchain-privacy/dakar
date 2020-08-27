package main

import (
	"context"
	"dashrpc/analytics"
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

func info(v ...interface{}) {
	log.SetPrefix("\033[0;31mcrawler\033[0m\t")
	log.Println(v)
	log.SetPrefix("")
}

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
	dbStatus, err := status.GetCrawlerStatus(dgraph)
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
	if f, err := cli.GetLogfile(cliArgs.Logfile); err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				fmt.Println(err)
			}
		}()
	}

	// create dgraph client
	dgraph, c, err := db.CreateDefaultClient()
	if err != nil {
		info(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			info(err)
		}
	}()

	if cliArgs.IsPrintStatus {
		status.PrintStatus(dgraph)
		return
	}

	if cliArgs.ResetDB {
		// get confirmation for database deletion
		var userAnswer string
		info("All data in the database will we deleted! Do you want to continue (yes/no)?")
		if _, err := fmt.Scanln(&userAnswer); err != nil {
			info(err)
			return
		}

		if userAnswer != "yes" {
			info("Exiting program. Database has not been changed.")
			return
		}

		err = db.DropAll(dgraph)
		if err != nil {
			info(err)
			return
		}
		info("dropped all data")
		err = db.SetupSchema(dgraph)
		if err != nil {
			info(err)
			return
		}
		info("setup new schema")
	}

	// check if schema exists
	if isSet, err := db.IsSchemaSet(dgraph); err != nil {
		info(err)
		return
	} else if !isSet {
		info("Schema is not set. Use -reset to create a new schema.")
		return
	}

	if !cliArgs.IgnoreSafeguard {
		if ok, err := isCrawling(dgraph); err != nil {
			info(err)
			return
		} else if ok {
			info("Crawling process is already running. Use -ignoresafeguard to crawl despite this.")
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
		info(err)
		return
	}

	count, err := client.GetBlockCount()
	if err != nil {
		info("Error: problem with count()", err.Error())
		return
	}
	info("Current block count in the chain of the RPC client:", count)

	// We will handle CTRL-C and CTRL-Z nicely
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

	ctx, cancelFunc := context.WithCancel(context.Background())

	chCrawlingStopped := make(chan bool, 1)
	chAnalyzingStopped := make(chan bool, 1)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			chCrawlingStopped <- true
		}()
		if cliArgs.Continuous {
			err = processor.ProcessBlocksContinuously(ctx, dgraph, client)
		} else {
			err = processor.ProcessBlockRange(ctx, dgraph, client, cliArgs.StartBlockID, cliArgs.StopBlockID)
		}

		if err != nil {
			info(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			chAnalyzingStopped <- true
		}()

		if err := analytics.StartPost(ctx, dgraph); err != nil {
			info(err)
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
	case <-chCrawlingStopped:
		cancelFunc()
		stoppedWorking = true
	case <-chAnalyzingStopped:
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
