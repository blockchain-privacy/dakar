package main

import (
	"backend/analytics"
	"backend/analytics/graph"
	heuristic "backend/analytics/heuristics/transaction"
	"backend/blockiterator"
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/status"
	dbus "backend/db/user"
	"backend/external"
	"backend/processor"
	"backend/server"
	"backend/user"

	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/rpcclient"

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
	cliArgs, err = cli.BuildArgs(cli.Continuous, cli.ResetDB, cli.RPCUser, cli.RPCPassword, cli.StartBlockID,
		cli.StopBlockID, cli.IsPrintStatus, cli.RPCHost, cli.RPCPort, cli.Logfile, cli.IgnoreSafeguard,
		cli.DisableHTTPServer, cli.DisableHeuristics, cli.DisableCrawler, cli.DisableClassifier,
		cli.HTTPServerPort, cli.DBPort, cli.DBHost, cli.BTC, cli.Dash, cli.Doge)

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
func isCrawling(db external.Database) (bool, error) {
	dbStatus, err := status.GetCrawlerStatus(db)
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
func waitForRPCClient(client external.RPCClient) bool {
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
func waitForDatabase(db external.Database) bool {
	const maxRetries = 5
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := 0; i < maxRetries; i++ {
		if status.IsConnectionEstablished(db) {
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

// shutdownServer sends a shutdown signal to the server with a timout of 10 seconds
func shutdownServer(srv *http.Server) {
	if srv == nil {
		return
	}
	info("### Shutting down server###")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		info("Server was shutdown and returned error:", err)
	}
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
	} else if len(cliArgs.Logfile) > 0 {
		fmt.Println("Could not create logfile", cliArgs.Logfile)
		return
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

	// disable the heuristic worker if it is disabled per configuration
	if !analyserConfig.IsHeuristicWorkerEnabled {
		cliArgs.DisableHeuristics = true
	}

	// disable classifying if it is disabled per configuration
	if !analyserConfig.IsClassifyingEnabled {
		cliArgs.DisableClassifier = true
	}

	info(processorConfig.BlockchainName, "mode active")

	// create dgraph client
	graphDB, c, err := db.CreateClient(cliArgs.DBEndpoint)
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
	if !waitForDatabase(graphDB) {
		return
	}

	if cliArgs.IsPrintStatus {
		status.PrintStatus(graphDB)
		return
	}

	// check if signing keys and basic auth are set
	if !cliArgs.DisableHTTPServer {
		_, _, keyErr := server.GetSigningKeysFromEnv()
		if keyErr != nil {
			info("error getting signing keys. Set the following environment variables:",
				server.SigningPubkeyEnvironmentField, server.SigningPrivkeyEnvironmentField)

			publicKey, privateKey, genErr := ed25519.GenerateKey(nil)
			if genErr != nil {
				return
			}

			info("Generated new key pair:\npublic key:", hex.EncodeToString(publicKey), "\nprivate key:", hex.EncodeToString(privateKey))
			return
		}

		_, _, authErr := server.GetBasicAuthCredentialsFromEnv()
		if authErr != nil {
			info("error getting basic auth credentials. Set the following environment variables:",
				server.BasicAuthUserEnvironmentField, server.BasicAuthPasswordHashEnvironmentField)

			password, pwErr := user.GenerateRandomPassword()
			if pwErr != nil {
				return
			}

			pwHash, pwErr := user.GeneratePasswordHash(user.DefaultPasswordConfig, password)
			if pwErr != nil {
				return
			}

			info("Generated new basic auth pair:\nuser: dakar", "\npassword:", password, "\npassword hash:", pwHash)
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

		err = db.DropAll(graphDB)
		if err != nil {
			info(err)
			return
		}
		info("Dropped all data.")
		err = db.SetupSchema(graphDB)
		if err != nil {
			info(err)
			return
		}
		info("Successfully set up new schema.")
	}

	if cliArgs.DisableClassifier && cliArgs.DisableCrawler && cliArgs.DisableHTTPServer {
		return
	}

	// check if schema exists
	if isSet, err := db.IsSchemaSet(graphDB); err != nil {
		info(err)
		return
	} else if !isSet {
		info("Schema is not set. Use -reset to create a new schema.")
		return
	}

	if !cliArgs.IgnoreSafeguard {
		if ok, err := isCrawling(graphDB); err != nil {
			info(err)
			return
		} else if ok {
			info("Crawling process is already running. Use -ignoresafeguard to crawl despite this.")
			return
		}
	}

	// create admin account if none is set
	if !cliArgs.DisableHTTPServer {
		// check if users already exist
		_, userErr := dbus.GetUsers(graphDB)
		if userErr != nil {
			// no users exists -> create admin user
			if errors.Is(userErr, dbus.ErrorUsersNotFound) {
				adminEmail := "admin@dakar.null"
				pw, userCreationError := dbus.CreateAdminUser(graphDB, adminEmail)
				if userCreationError != nil {
					info(err)
					return
				}
				// do not log
				fmt.Println("New admin user created. Email:", adminEmail, "Pw:", pw)
				if len(cliArgs.Logfile) > 0 {
					fmt.Println("Write the credentials down, they will not be written to the log file.")
				} else {
					fmt.Println("Write the credentials down.")
				}
			} else {
				info(userErr)
				return
			}
		}
	}

	// Setup the RPC connection, only if needed
	var client *rpcclient.Client
	if !cliArgs.DisableHTTPServer || !cliArgs.DisableCrawler {
		client, err = rpcclient.New(&rpcclient.ConnConfig{
			Host:         cliArgs.RPCEndpoint,
			User:         cliArgs.RPCUser,
			Pass:         cliArgs.RPCPassword,
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

	// handle shutdown signals
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

	appContext, terminateApp := context.WithCancel(context.Background())

	// channels which are set to true as soon as the associated goroutine stops
	chCrawlingStopped := make(chan bool, 1)
	chClassifyingStopped := make(chan bool, 1)

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

			if processorErr := blockiterator.StartIteration(processor.NewCrawler(
				appContext, graphDB, client, processorConfig)); processorErr != nil {
				info(processorErr)
			}
		}()

		//wg.Add(1)
		//go func() {
		//	defer wg.Done()
		//	defer func() {
		//		chCrawlingStopped <- true
		//	}()
		//	if cliArgs.Continuous {
		//		err = processor.ProcessBlocksContinuously(appContext, graphDB, client, processorConfig)
		//	} else {
		//		err = processor.ProcessBlockRange(appContext, graphDB, client, cliArgs.StartBlockID,
		//			cliArgs.StopBlockID, processorConfig)
		//	}
		//
		//	if err != nil {
		//		info(err)
		//	}
		//}()
	}

	graphWrapper := graph.NewWrapper(appContext, graphDB)
	worker := heuristic.NewWorker(graphWrapper)
	var classifierStarted bool
	if !cliArgs.DisableHTTPServer && !cliArgs.DisableHeuristics {
		// the classifier must be started after the in-memory graphs are loaded
		classifierStarted = true
		go func() {
			graphErr := graphWrapper.LoadGraphs()
			if graphErr != nil {
				info(graphErr)
				return
			}

			if !cliArgs.DisableClassifier {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if iterErr := blockiterator.StartIteration(graphWrapper); iterErr != nil {
						info(iterErr)
					}
				}()
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer func() {
						chClassifyingStopped <- true
					}()

					if classifierErr := blockiterator.StartIteration(analytics.NewClassifier(
						appContext, graphDB, analyserConfig)); classifierErr != nil {
						info(classifierErr)
					}
				}()
			}
		}()

		if ok := worker.Start(appContext, graphDB); !ok {
			info("could not start worker")
			return
		}
	}

	// activate classifier
	if !cliArgs.DisableClassifier && !classifierStarted {
		// in-memory graphs are not loaded -> start classifier
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chClassifyingStopped <- true
			}()

			if classifierErr := blockiterator.StartIteration(analytics.NewClassifier(
				appContext, graphDB, analyserConfig)); classifierErr != nil {
				info(classifierErr)
			}
		}()
	}

	// activate server
	var srv *http.Server
	if !cliArgs.DisableHTTPServer {
		wg.Add(1)
		srv = server.StartServer(&wg, cliArgs.HTTPServerPort, graphDB, client, worker)
	}

	var crawlerStopped bool
	var classifierStopped bool
	var interrupted bool

	for !(interrupted || (crawlerStopped && classifierStopped)) {
		select {
		case <-chSignal:
			interrupted = true
			terminateApp()
			shutdownServer(srv)
		case <-chCrawlingStopped:
			terminateApp()
			crawlerStopped = true
		case <-chClassifyingStopped:
			terminateApp()
			classifierStopped = true
		}
	}

	if !cliArgs.DisableHTTPServer && crawlerStopped && classifierStopped {
		// if the crawler and classifier stopped working on their own accord,
		// the server is still active at this point

		<-chSignal
		shutdownServer(srv)
	}

	wg.Wait()
}
