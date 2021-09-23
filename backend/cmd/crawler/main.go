package main

import (
	"backend/analytics"
	"backend/analytics/clustering"
	"backend/analytics/graph"
	heuristic "backend/analytics/heuristics/transaction"
	"backend/blockiterator"
	cli "backend/cmd/cliutil"
	"backend/db"
	dbus "backend/db/user"
	"backend/processor"
	"backend/server"

	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
)

// VersionString displays the version of the Crawler
const VersionString = "v1.0.0"

var thisLogger *log.Logger

func initLogger() {
	thisLogger = log.New(log.Writer(), "\033[0;31mcrawler\033[0m\t", log.Flags())
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

func info(v ...interface{}) {
	thisLogger.Println(v...)
}

func setCommandFlags(c *Commands) {
	flag.BoolVar(&c.ResetDB, "reset", false, "Remove all data from the database (default: false)")
	flag.BoolVar(&c.IgnoreSafeGuard, "ignoresafeguard", false, "Ignore the crawling safe guard (default: false)")
}

// The crawler for the system. It needs to be run prior to using any of the other
// commands that rely on the Dgraph DB to be pre-created.
//
// The crawler traverses the Dash blockchain and creates a Dgraph database entry for each transaction
// starting from a given block, and, working backwards, until a given stop block.
func main() {
	fmt.Println("Dakar", VersionString, "compiled with", runtime.Version())

	////// SET FLAGS //////

	var commands Commands
	setCommandFlags(&commands)

	defaultConfigName := "config.yml"
	var filePath string
	var createConfigFile bool
	cli.SetConfigFlags(defaultConfigName, &filePath, &createConfigFile)
	flag.Parse()

	////// CONFIGURATION FILE HANDLING //////

	if createConfigFile {
		fmt.Println("Generating configuration file ...")
		defConfig, configErr := getDefaultConfig()
		if configErr != nil {
			fmt.Println(configErr)
			return
		}

		err := cli.WriteConfig(defaultConfigName, defConfig)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("config file", defaultConfigName, "successfully created")
		return
	}

	var config Config
	if err := cli.ReadConfig(filePath, &config); err != nil {
		fmt.Println(err)
		return
	}

	if moduleErr := checkHTTPModuleConfig(config.Modules.HTTP); moduleErr != nil {
		fmt.Println(moduleErr)
		return
	}

	////// SETUP //////

	// setup Logging
	if f, err := cli.GetLogfile(config.Logfile); err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				fmt.Println(err)
			}
		}()
	} else if len(config.Logfile) > 0 {
		fmt.Println("Could not create logfile", config.Logfile)
		return
	}

	initAllLoggers()

	// select blockchain config
	var processorConfig processor.Config
	var analyserConfig analytics.Config
	if config.BlockchainMode == "Dash" {
		processorConfig = processor.NewDashConfig()
		analyserConfig = analytics.NewDashConfig()
	} else if config.BlockchainMode == "BTC" {
		processorConfig = processor.NewBitcoinConfig()
		analyserConfig = analytics.NewBitcoinConfig()
	} else if config.BlockchainMode == "Doge" {
		processorConfig = processor.NewDogecoinConfig()
		analyserConfig = analytics.NewDogecoinConfig()
	} else {
		fmt.Println("invalid blockchain mode selected")
		return
	}

	// disable the heuristic worker if it is disabled per configuration
	if !analyserConfig.IsHeuristicWorkerEnabled {
		config.Modules.Heuristics = false
	}

	// disable classifying if it is disabled per configuration
	if !analyserConfig.IsClassifyingEnabled {
		config.Modules.Classifier = false
	}

	// disable HMI clustering if it is disabled per configuration
	if !analyserConfig.IsHMIClusteringEnabled {
		config.Modules.Clustering.HMI = false
	}

	// disable FMI clustering if it is disabled per configuration
	if !analyserConfig.IsFMIClusteringEnabled {
		config.Modules.Clustering.FMI = false
	}

	info(processorConfig.BlockchainName, "mode")

	////// CONNECT TO DATABASE //////

	endpoint, err := cli.BuildEndpoint(config.Database.Host, config.Database.Port)
	if err != nil {
		info(err)
		return
	}

	// create dgraph client
	graphDB, c, err := db.CreateClient(endpoint)
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

	if commands.ResetDB {
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

	if !config.Modules.Classifier && !config.Modules.Crawler.Active &&
		!config.Modules.Clustering.HMI && !config.Modules.Clustering.FMI &&
		!config.Modules.HTTP.Active {
		log.Println("All modules are disabled. Exiting ...")
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

	if !commands.IgnoreSafeGuard {
		if ok, err := isCrawling(graphDB); err != nil {
			info(err)
			return
		} else if ok {
			info("Crawling process is already running. Use -ignoresafeguard to crawl despite this.")
			return
		}
	}

	////// CREATE ADMIN USER //////

	// create admin account if none is set
	if config.Modules.HTTP.Active {
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
				if len(config.Logfile) > 0 {
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

	////// CONNECT TO RPC //////

	// Set up the RPC connection, only if needed
	var client *rpcclient.Client
	var batchClient *rpcclient.Client
	if config.Modules.HTTP.Active || config.Modules.Crawler.Active {
		rpcEndpoint, err := cli.BuildEndpoint(config.RPC.Host, config.RPC.Port)
		if err != nil {
			info(err)
			return
		}

		connection := rpcclient.ConnConfig{
			Host:                rpcEndpoint,
			User:                config.RPC.User,
			Pass:                config.RPC.Password,
			DisableConnectOnNew: true,
			DisableTLS:          true,
			HTTPPostMode:        true,
		}

		client, err = rpcclient.New(&connection, nil)
		if err != nil {
			info(err)
			return
		}

		batchClient, err = rpcclient.NewBatch(&connection)
		if err != nil {
			info(err)
			return
		}

		// test if rpc client is active
		if !waitForRPCClient(client) {
			return
		}

		// test if batch rpc client is active
		if !waitForBatchRPCClient(batchClient) {
			return
		}
	}

	////// START MODULES //////

	// handle shutdown signals
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

	appContext, terminateApp := context.WithCancel(context.Background())

	// channels which are set to true as soon as the associated goroutine stops
	chCrawlingStopped := make(chan bool, 1)
	chClassifyingStopped := make(chan bool, 1)
	chHMIClusteringStopped := make(chan bool, 1)
	chFMIClusteringStopped := make(chan bool, 1)

	// the wait group which handles the modules of the crawler
	var wg sync.WaitGroup

	// activate crawler
	if config.Modules.Crawler.Active {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chCrawlingStopped <- true
			}()

			if processorErr := blockiterator.StartIteration(processor.NewCrawler(
				appContext, graphDB, client, batchClient, config.Modules.Crawler.InitialCacheSize,
				processorConfig)); processorErr != nil {
				info(processorErr)
			}
		}()
	}

	graphWrapper := graph.NewWrapper(appContext, graphDB)
	worker := heuristic.NewWorker(graphWrapper)
	var classifierStarted bool

	if config.Modules.HTTP.Active && config.Modules.Heuristics {
		// the classifier must be started after the in-memory graphs are loaded
		classifierStarted = true
		go func() {
			graphErr := graphWrapper.LoadGraphs()
			if graphErr != nil {
				info(graphErr)
				return
			}

			if config.Modules.Classifier {
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
	if config.Modules.Classifier && !classifierStarted {
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

	// activate HMI clustering
	if config.Modules.Clustering.HMI {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chHMIClusteringStopped <- true
			}()

			if clusteringErr := blockiterator.StartIteration(clustering.NewHierarchicalMultiInput(
				appContext, graphDB)); clusteringErr != nil {
				info(clusteringErr)
			}
		}()
	}

	// activate FMI clustering
	if config.Modules.Clustering.FMI {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chFMIClusteringStopped <- true
			}()

			if clusteringErr := blockiterator.StartIteration(clustering.NewFlatMultiInput(
				appContext, graphDB)); clusteringErr != nil {
				info(clusteringErr)
			}
		}()
	}

	// activate server
	var srv *http.Server
	if config.Modules.HTTP.Active {
		wg.Add(1)
		srv = server.StartServer(&wg, config.Modules.HTTP.Port, config.Modules.HTTP.BasicAuthUser,
			config.Modules.HTTP.BasicAuthPWHash, config.Modules.HTTP.TokenPublicKey, config.Modules.HTTP.TokenPrivateKey,
			graphDB, client, worker)
	}

	////// HANDLE SHUTDOWN //////

	var crawlerStopped = !config.Modules.Crawler.Active
	var classifierStopped = !config.Modules.Classifier
	var clusteringHMIStopped = !config.Modules.Clustering.HMI
	var clusteringFMIStopped = !config.Modules.Clustering.FMI
	var interrupted bool

	for !(interrupted || (crawlerStopped && classifierStopped && clusteringHMIStopped && clusteringFMIStopped)) {
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
		case <-chHMIClusteringStopped:
			terminateApp()
			clusteringHMIStopped = true
		case <-chFMIClusteringStopped:
			terminateApp()
			clusteringFMIStopped = true
		}
	}

	if config.Modules.HTTP.Active && crawlerStopped && classifierStopped &&
		clusteringHMIStopped && clusteringFMIStopped {
		// if the crawler, the classifier and clustering stopped working on their own accord,
		// the server is still active at this point

		<-chSignal
		shutdownServer(srv)
	}

	wg.Wait()
}
