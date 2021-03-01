package main

import (
	"backend/analytics"
	heuristic "backend/analytics/heuristics/transaction"
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/status"
	dbus "backend/db/user"
	"backend/processor"
	"backend/server"
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/rpcclient"

	"github.com/dgraph-io/dgo/v2"

	"golang.org/x/crypto/ed25519"
)

// VersionString displays the version of the Crawler
const VersionString = "v1.0.0"

var thisLogger *log.Logger

func initLogger() {
	thisLogger = log.New(log.Writer(), "\033[0;31mcrawler\033[0m\t", log.Flags())
}

func info(v ...interface{}) {
	thisLogger.Println(v...)
}

func initAllLoggers() {
	initLogger()

	writer := log.Writer()
	flags := log.Flags()

	analytics.InitLogger(writer, flags)
	db.InitLogger(writer, flags)
	processor.InitLogger(writer, flags)
	server.InitLogger(writer, flags)
	heuristic.InitLogger(writer, flags)
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

// waitForRPCClient waits until the RPC client is ready to receive requests
func waitForRPCClient(client *rpcclient.Client) bool {
	const maxRetries = 5
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := 0; i < maxRetries; i++ {
		_, err := client.GetBlockCount()
		if err == nil {
			if printedErrMessage {
				info("Successfully established connection to RPC client.")
			}
			return true
		}

		if !printedErrMessage {
			info("Waiting for RPC client to start")
			printedErrMessage = true
		}

		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}
	info("RPC client is not ready to receive requests.")
	return false
}

// waitForDatabase waits until the database is ready to receive requests
func waitForDatabase(dgraph *dgo.Dgraph) bool {
	const maxRetries = 5
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := 0; i < maxRetries; i++ {
		if status.IsConnectionEstablished(dgraph) {
			if printedErrMessage {
				info("Successfully established connection to database.")
			}
			return true
		}

		if !printedErrMessage {
			info("Waiting for database")
			printedErrMessage = true
		}

		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	info("Database is not ready to receive requests.")

	return false
}

// The crawler for the system. It needs to be run prior to using any of the other
// commands that rely on the Dgraph DB to be pre-created.
//
// The crawler traverses the Dash blockchain and creates a Dgraph database entry for each transaction
// starting from a given block, and, working backwards, until a given stop block.
func main() {
	fmt.Println("Dakar", VersionString)
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

	initAllLoggers()

	// select blockchain config
	var processorConfig processor.Config
	var analyserConfig analytics.Config
	if cliArgs.Dash {
		processorConfig = processor.NewDashConfig()
		analyserConfig = analytics.NewDashConfig()
	} else if cliArgs.BTC {
		processorConfig = processor.NewBitcoinConfig()
		analyserConfig = analytics.NewBitcoinConfig()
	} else if cliArgs.Doge {
		processorConfig = processor.NewDogecoinConfig()
		analyserConfig = analytics.NewDogecoinConfig()
	} else {
		fmt.Println("invalid blockchain mode selected")
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

	// test if database is active
	if !waitForDatabase(dgraph) {
		return
	}

	if cliArgs.IsPrintStatus {
		status.PrintStatus(dgraph)
		return
	}

	// check if signing keys are set
	if !cliArgs.DisableHttpServer {
		_, _, keyErr := server.GetSigningKeysFromEnv()
		if keyErr != nil {
			info("error getting signing keys. Set the following environment variables:",
				server.SigningPubkeyEnvironmentField, server.SigningPrivkeyEnvironmentField, keyErr)

			publicKey, privateKey, err := ed25519.GenerateKey(nil)
			if err != nil {
				return
			}

			info("Generated new key pair:\npublic key:", hex.EncodeToString(publicKey), "\nprivate key:", hex.EncodeToString(privateKey))
			return
		}
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
		info("Dropped all data.")
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

	// create admin account if none is set
	if !cliArgs.DisableHttpServer {
		// check if users already exist
		_, userErr := dbus.GetUsers(dgraph)
		if userErr != nil {
			// no users exists -> create admin user
			if errors.Is(userErr, dbus.ErrorUsersNotFound) {
				adminEmail := "admin@dakar.null"
				pw, userCreationError := dbus.CreateAdminUser(dgraph, adminEmail)
				if userCreationError != nil {
					info(err)
					return
				}
				// do not log
				fmt.Println("New admin user created. Email:", adminEmail, "Pw:", pw)
				fmt.Println("Write the credentials down, this message is not logged.")
			} else {
				info(userErr)
				return
			}
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

		// test if rpc client is active
		if !waitForRPCClient(client) {
			return
		}
	}

	// We will handle CTRL-C and CTRL-Z nicely
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

	crawlerContext, cancelCrawler := context.WithCancel(context.Background())
	analyzerContext, cancelAnalyzer := context.WithCancel(context.Background())

	chCrawlingStopped := make(chan bool, 1)
	chAnalyzingStopped := make(chan bool, 1)

	// the wait group which handles the modules of the crawler
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
	if !cliArgs.DisableAnalyzer && analyserConfig.IsAnalysingEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chAnalyzingStopped <- true
			}()

			if analyserErr := analytics.StartAnalysis(analyzerContext, dgraph, analyserConfig); analyserErr != nil {
				info(analyserErr)
			}
		}()
	}

	// activate server
	var srv server.Server
	if !cliArgs.DisableHttpServer {
		wg.Add(1)
		srv = server.CreateServer(&wg, cliArgs.HttpServerPort, dgraph, client)
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
			srv.ShutdownServer()
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
			srv.ShutdownServer()
		}
	}

	wg.Wait()
}
