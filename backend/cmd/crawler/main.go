package main

import (
	"backend/analytics"
	"backend/analytics/clustering"
	"backend/analytics/graph"
	heuristic "backend/analytics/heuristics"
	"backend/blockiterator"
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/status"
	dbus "backend/db/user"
	"backend/external"
	"backend/processor"
	"backend/server"
	"context"
	"errors"
	"flag"
	"fmt"
	ory "github.com/ory/kratos-client-go"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
)

// versionString displays the version of the Crawler
const versionString = "v1.0.0"

// blockchainMode is empty by default and should be set at compile time via the ldflags options.
// blockchainMode controls various config parameters (see config.go).
// Allowed values: "Dash" "Bitcoin" "Doge"
var blockchainMode = ""

var thisLogger *slog.Logger

func initAllLoggers(fileHandle *os.File) {
	var outputWriter io.Writer
	if fileHandle != nil {
		outputWriter = io.MultiWriter(fileHandle, os.Stdout)
	} else {
		outputWriter = os.Stdout
	}

	logger := slog.New(slog.NewTextHandler(outputWriter, nil))
	slog.SetDefault(logger)

	thisLogger = slog.With(slog.String("module", "crawler"))

	analytics.InitLogger()
	db.InitLogger()
	processor.InitLogger()
	server.InitLogger()
	heuristic.InitLogger()
}

func info(msg string, v ...any) {
	thisLogger.Info(msg, v...)
}

func warn(err error, v ...any) {
	cli.LogError(thisLogger, err, v...)
}

func setCommandFlags(c *Commands) {
	flag.BoolVar(&c.ResetDB, "reset", false, "Remove all data from the database (default: false)")
	flag.BoolVar(&c.IgnoreSafeGuard, "ignoresafeguard", false, "Ignore the crawling safe guard (default: false)")
	flag.BoolVar(&c.ShowVersion, "version", false, "Show version information")
}

// selectConfig returns processor and analytics configurations based on the given blockchain mode.
func selectConfig(blockchainMode string) (processor.Config, analytics.Config, error) {
	switch blockchainMode {
	case "Dash":
		return processor.NewDashConfig(), analytics.NewDashConfig(), nil
	case "Bitcoin":
		return processor.NewBitcoinConfig(), analytics.NewBitcoinConfig(), nil
	case "Doge":
		return processor.NewDogecoinConfig(), analytics.NewDogecoinConfig(), nil
	default:
		return processor.Config{}, analytics.Config{}, cli.NewStackErrorStr("invalid blockchain mode")
	}
}

// disableModules disables modules in config based on analyserConfig
func disableModules(analyserConfig analytics.Config, config *Config) {
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
}

// resetDatabaseDialog asks the user if the database should be reset and performs the reset if necessary.
// Returns false if the program should be shutdown.
func resetDatabaseDialog(database external.Database) bool {
	// get confirmation for database deletion
	var userAnswer string
	info("All data in the database will we deleted! Do you want to continue (yes/no)?")
	if _, err := fmt.Scanln(&userAnswer); err != nil {
		warn(err)
		return false
	}

	if strings.TrimSpace(strings.ToLower(userAnswer)) != "yes" {
		info("Exiting program. Database was not modified.")
		return false
	}

	if err := db.DropAll(database); err != nil {
		warn(err)
		return false
	}
	info("Dropped all data.")

	if err := db.SetupSchema(database); err != nil {
		warn(err)
		return false
	}
	info("Successfully set up new schema.")

	if err := status.InitializeMeta(database, blockchainMode); err != nil {
		warn(err)
		return false
	}
	info("Successfully initialized database")

	return true
}

// createAdminUser creates an admin user if no user exist in the database and
// prints the credentials to stdout
func createAdminUser(database external.Database, adminAuth *ory.APIClient) error {
	// check if users already exist
	_, userErr := dbus.GetUsers(database)
	if userErr == nil {
		return nil
	}

	// no users exists -> create admin user
	if errors.Is(userErr, dbus.ErrorUsersNotFound) {
		adminEmail := "admin@dakar.null"
		pw, userCreationError := dbus.CreateAdminUser(database, adminAuth, adminEmail)
		if userCreationError != nil {
			return userCreationError
		}
		// do not log
		fmt.Println("New admin user created. Email:", adminEmail, "Pw:", pw)
		fmt.Println("Write the credentials down.")
	} else {
		return userErr
	}

	return nil
}

// connectBlockchainRPCClient connects to blockchain RPC client specified in the given configuration.
func connectBlockchainRPCClient(rpcConfig RPCConfig) (*rpcclient.Client, *rpcclient.Client, error) {
	rpcEndpoint, err := cli.BuildEndpoint(rpcConfig.Host, rpcConfig.Port)
	if err != nil {
		return nil, nil, err
	}

	connection := rpcclient.ConnConfig{
		Host:                rpcEndpoint,
		User:                rpcConfig.User,
		Pass:                rpcConfig.Password,
		DisableConnectOnNew: true,
		DisableTLS:          true,
		HTTPPostMode:        true,
	}

	client, err := rpcclient.New(&connection, nil)
	if err != nil {
		return nil, nil, err
	}

	batchClient, err := rpcclient.NewBatch(&connection)
	if err != nil {
		return nil, nil, err
	}

	// test if rpc client is active
	if !waitForRPCClient(client) {
		// no error text, was already handled in function above
		return nil, nil, cli.NewStackErrorStr("")
	}

	// test if batch rpc client is active
	if !waitForBatchRPCClient(batchClient) {
		// no error text, was already handled in function above
		return nil, nil, cli.NewStackErrorStr("")
	}

	return client, batchClient, nil
}

//	@title			Dakar API
//	@version		1.0.3
//	@description	This is the REST API for Dakar
//	@host			localhost
//	@BasePath		/api/v1
//
// The crawler for the system. It needs to be run prior to using any of the other
// commands that rely on the Dgraph DB to be pre-created.
//
// The crawler traverses the Dash blockchain and creates a Dgraph database entry for each transaction
// starting from a given block, and, working backwards, until a given stop block.
//
// nolint:gocyclo
func main() {
	////// SET FLAGS //////

	var commands Commands
	setCommandFlags(&commands)

	const defaultConfigName = "config.yml"
	var filePath string
	var createConfigFile bool
	cli.SetConfigFlags(defaultConfigName, &filePath, &createConfigFile)
	flag.Parse()

	////// PRINT VERSION //////

	if commands.ShowVersion {
		printVersion()
		return
	}

	////// CONFIGURATION FILE HANDLING //////

	if createConfigFile {
		fmt.Println("Generating configuration file ...")
		err := cli.WriteConfig(defaultConfigName, defaultConfig)
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

	if moduleErr := checkAPIModuleConfig(config.Modules.HTTP); moduleErr != nil {
		fmt.Println(moduleErr)
		return
	}

	////// SETUP //////

	// setup Logging
	f, err := cli.GetLogfile(config.Logfile)
	if err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				fmt.Println(err)
			}
		}()
	} else if len(config.Logfile) > 0 {
		fmt.Println("Could not create logfile", config.Logfile)
		return
	}

	initAllLoggers(f)

	processorConfig, analyserConfig, err := selectConfig(blockchainMode)
	if err != nil {
		fmt.Println("invalid blockchain mode selected: '" + blockchainMode + "'")
		fmt.Println("the blockchain mode has to be set at compile time via the ldflags option.")
		fmt.Println("example: go build -ldflags \"-X main.blockchainMode=Dash\" .")
		return
	}

	disableModules(analyserConfig, &config)

	info("Blockchain mode: " + processorConfig.BlockchainName)

	////// CONNECT TO DATABASE //////

	endpoint, err := cli.BuildEndpoint(config.Database.Host, config.Database.Port)
	if err != nil {
		warn(err)
		return
	}

	// create dgraph client
	graphDB, c, err := external.CreateClient(endpoint)
	if err != nil {
		warn(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			warn(err)
		}
	}()

	// test if database is active
	if !external.WaitForDatabase(graphDB) {
		info("could not connect to database")
		return
	}

	if commands.ResetDB {
		if !resetDatabaseDialog(graphDB) {
			return
		}
	}

	// exit if no module is active (excluding the metrics module)
	if !config.Modules.Classifier && !config.Modules.Crawler.Active &&
		!config.Modules.Clustering.HMI && !config.Modules.Clustering.FMI &&
		!config.Modules.HTTP.Active {
		log.Println("All modules are disabled. Exiting ...")
		return
	}

	// check if schema exists
	if isSet, err := db.IsSchemaSet(graphDB); err != nil {
		warn(err)
		return
	} else if !isSet {
		info("Schema is not set. Use -reset to create a new schema.")
		return
	}

	if !checkMeta(graphDB) {
		return
	}

	if !commands.IgnoreSafeGuard {
		if ok, err := isCrawling(graphDB); err != nil {
			warn(err)
			return
		} else if ok {
			info("Crawling process is already running. Use -ignoresafeguard to crawl despite this.")
			return
		}
	}

	////// CONNECT TO KRATOS //////

	auth, adminAuth, err := getKratosClient(config.Modules.HTTP.KratosPublicEndpoint,
		config.Modules.HTTP.KratosAdminEndpoint)
	if err != nil {
		warn(err)
		return
	}

	////// CREATE ADMIN USER //////

	// create admin account if none is set
	if config.Modules.HTTP.Active {
		if err := createAdminUser(graphDB, adminAuth); err != nil {
			warn(err)
			return
		}
	}

	////// CONNECT TO RPC //////

	// Set up the RPC connection, only if needed
	var client *rpcclient.Client
	var batchClient *rpcclient.Client
	if config.Modules.HTTP.Active || config.Modules.Crawler.Active {
		client, batchClient, err = connectBlockchainRPCClient(config.RPC)
		if err != nil {
			warn(err)
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
				warn(processorErr)
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
				warn(graphErr)
				return
			}

			if config.Modules.Classifier {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if iterErr := blockiterator.StartIteration(graphWrapper); iterErr != nil {
						warn(iterErr)
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
						warn(classifierErr)
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
				warn(classifierErr)
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
				warn(clusteringErr)
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
				warn(clusteringErr)
			}
		}()
	}

	// start api endpoint
	var apiHTTPServer *http.Server
	if config.Modules.HTTP.Active {
		apiServer, serverErr := server.NewServer(graphDB, adminAuth, auth, client, worker)
		if serverErr != nil {
			warn(serverErr)
		}

		wg.Add(1)

		apiHTTPServer = apiServer.StartServer(&wg, config.Modules.HTTP.Port)
	}

	// start metrics endpoint
	var metricsHTTPServer *http.Server
	if config.Modules.Metrics.Active {
		wg.Add(1)
		metricsHTTPServer = server.StartMetrics(&wg, config.Modules.Metrics.Port)
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
			shutdownServer(apiHTTPServer)
			shutdownServer(metricsHTTPServer)
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
		shutdownServer(apiHTTPServer)
		shutdownServer(metricsHTTPServer)
	}

	wg.Wait()
}
