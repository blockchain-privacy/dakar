package main

import (
	"backend/analytics"
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/status"
	"backend/processor"

	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"

	"github.com/dgraph-io/dgo/v2"
)

// VersionString displays the version of the Crawler
const VersionString = "v1.0.0"

func info(v ...interface{}) {
	log.SetPrefix("\033[0;31mcrawler\033[0m\t")
	log.Println(v...)
	log.SetPrefix("")
}

func getCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.Continuous, cli.ResetDB, cli.RpcUser, cli.RpcPassword, cli.StartBlockID,
		cli.StopBlockID, cli.IsPrintStatus, cli.RpcHost, cli.RpcPort, cli.Logfile, cli.IgnoreSafeguard,
		cli.DisableHttpServer, cli.DisableAnalyzer, cli.DisableCrawler, cli.HttpServerPort, cli.DBPort,
		cli.DBHost, cli.BTC, cli.Dash, cli.Doge)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	if numSelected := cli.NumBlockchainSelected(cliArgs); numSelected != 1 {
		flag.PrintDefaults()
		if numSelected == 0 {
			err = errors.New(fmt.Sprintln("Select a blockchain (-dash, -btc or -doge)."))
		} else {
			err = errors.New(fmt.Sprintln("Number of blockchains selected:", numSelected,
				"Only one selected blockchain is allowed"))
		}

		return
	}

	if !cliArgs.DisableCrawler && !cliArgs.Continuous && (cliArgs.StartBlockID == 0 || cliArgs.StopBlockID == 0) {
		flag.PrintDefaults()
		err = errors.New("select crawling mode")
		return
	}

	if cliArgs.DisableCrawler && (cliArgs.StartBlockID > 0 || cliArgs.StopBlockID > 0 || cliArgs.Continuous) {
		flag.PrintDefaults()
		err = errors.New("enable crawler to use additional arguments")
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
	fmt.Printf("Dakar %s\nBlock crawler\n\n", VersionString)
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

	// select blockchain config
	var processorConfig processor.Config
	if cliArgs.Dash {
		processorConfig = processor.NewDashConfig()
	} else if cliArgs.BTC {
		processorConfig = processor.NewBitcoinConfig()
	} else if cliArgs.Doge {
		processorConfig = processor.NewDogecoinConfig()
	} else {
		fmt.Println("invalid blockchain selected")
		return
	}

	info(processorConfig.BlockchainName, "mode active")

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

		if strings.TrimSpace(strings.ToLower(userAnswer)) != "yes" {
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

	if cliArgs.DisableAnalyzer && cliArgs.DisableCrawler && cliArgs.DisableHttpServer {
		return
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

	// Setup the RPC connection, only if needed
	var client *rpcclient.Client
	if !cliArgs.DisableHttpServer || !cliArgs.DisableCrawler {
		client, err = rpcclient.New(&rpcclient.ConnConfig{
			Host:         cliArgs.RpcEndpoint,
			User:         cliArgs.RpcUser,
			Pass:         cliArgs.RpcPassword,
			DisableTLS:   true,
			HTTPPostMode: true,
		}, nil)
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
	}

	// We will handle CTRL-C and CTRL-Z nicely
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

	crawlerContext, cancelCrawler := context.WithCancel(context.Background())
	analyzerContext, cancelAnalyzer := context.WithCancel(context.Background())

	chCrawlingStopped := make(chan bool, 1)
	chAnalyzingStopped := make(chan bool, 1)

	// the waitgroup which handles the modules of the crawler
	var wg sync.WaitGroup

	// activate crawler
	if !cliArgs.DisableCrawler {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chCrawlingStopped <- true
			}()
			if cliArgs.Continuous {
				err = processor.ProcessBlocksContinuously(crawlerContext, dgraph, client, processorConfig)
			} else {
				err = processor.ProcessBlockRange(crawlerContext, dgraph, client, cliArgs.StartBlockID,
					cliArgs.StopBlockID, processorConfig)
			}

			if err != nil {
				info(err)
			}
		}()
	}

	// activate analyzer
	if !cliArgs.DisableAnalyzer {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chAnalyzingStopped <- true
			}()

			if err := analytics.StartPost(analyzerContext, dgraph); err != nil {
				info(err)
			}
		}()
	}

	// activate server
	var srv Server
	if !cliArgs.DisableHttpServer {
		wg.Add(1)
		srv = createServer(&wg, cliArgs.HttpServerPort, dgraph, client)
	}

	var crawlerStopped bool
	var analyzerStopped bool
	var interrupted bool

	for !(interrupted || (crawlerStopped && analyzerStopped)) {
		select {
		case <-chSignal:
			interrupted = true
			cancelCrawler()
			cancelAnalyzer()
			srv.shutdownServer()
		case <-chCrawlingStopped:
			cancelCrawler()
			crawlerStopped = true
		case <-chAnalyzingStopped:
			cancelAnalyzer()
			analyzerStopped = true
		}
	}

	if !cliArgs.DisableHttpServer && crawlerStopped && analyzerStopped {
		// if the crawler and analyzer stopped working on there own accord, the server is still active at this point
		select {
		case <-chSignal:
			srv.shutdownServer()
		}
	}

	wg.Wait()
}
