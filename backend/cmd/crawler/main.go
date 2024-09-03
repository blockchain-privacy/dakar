package main

import (
	"backend/analytics"
	"backend/analytics/clustering"
	"backend/analytics/graph"
	"backend/blockiterator"
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/status"
	"backend/db/upgrades"
	"backend/external"
	"backend/jsonrpc"
	"backend/processor"
	"backend/server"
	"backend/userserver"
	"backend/workspace"
	"context"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qrest/gomisc/config"
	"github.com/qrest/gomisc/serror"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"
)

// versionString displays the version of the Crawler
const versionString = "v1.0.0"

// name of the executable
const executableName = "crawler"

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
	userserver.InitLogger()
	workspace.InitLogger()
	upgrades.InitLogger()
}

func info(msg string, v ...any) {
	thisLogger.Info(msg, v...)
}

func warn(err error, v ...any) {
	serror.Log(thisLogger, err, v...)
}

func setCommandFlags(c *Commands) {
	flag.BoolVar(&c.ResetDB, "reset", false, "Remove all data from the database (default: false)")
	flag.BoolVar(&c.IgnoreSafeGuard, "ignoresafeguard", false, "Ignore the crawling safe guard (default: false)")
	flag.BoolVar(&c.ShowVersion, "version", false, "Show version information (default: false)")
	flag.BoolVar(&c.UpgradeDatabase, "upgradedatabase", false, "Upgrade the database schema to the newest version (default: false)")
	flag.StringVar(&c.CPUProfilePath, "cpuprofile", "", "Path where the cpu profile should be stored (default: <empty>)")
}

// selectConfig returns processor and analytics configurations based on the given blockchain mode.
func selectConfig(blockchainMode string) (processor.Config, analytics.Config, error) {
	switch blockchainMode {
	case "Dash":
		return processor.NewDashConfig(), analytics.NewDashConfig(), nil
	case "Bitcoin":
		return processor.NewBitcoinConfig(), analytics.NewBitcoinConfig(), nil
	default:
		return processor.Config{}, analytics.Config{}, serror.FromStr("invalid blockchain mode")
	}
}

// disableModules disables modules in config based on analyserConfig
func disableModules(analyserConfig analytics.Config, config *Config) {
	if !analyserConfig.IsHeuristicWorkerEnabled {
		config.Modules.Heuristics = false
	}

	// disable classifying if it is disabled per configuration
	if !analyserConfig.IsClassifyingEnabled {
		config.Modules.Classifier.Active = false
	}

	// disable HMI clustering if it is disabled per configuration
	if !analyserConfig.IsHMIClusteringEnabled {
		config.Modules.HMI = false
	}

	// disable FMI clustering if it is disabled per configuration
	if !analyserConfig.IsFMIClusteringEnabled {
		config.Modules.FMI.Active = false
	}
}

// resetDatabaseDialog asks the user if the database should be reset and performs the reset if necessary.
// Returns false if the program should be shutdown.
func resetDatabaseDialog(database external.Database, blockchainMode string) error {
	// get confirmation for database deletion
	var userAnswer string
	info("All data in the database will we deleted! Do you want to continue (yes/no)?")
	if _, err := fmt.Scanln(&userAnswer); err != nil {
		return err
	}

	if strings.TrimSpace(strings.ToLower(userAnswer)) != "yes" {
		info("Exiting program. Database was not modified.")
		return nil
	}

	if err := db.DropAll(database); err != nil {
		return err
	}
	info("Dropped all data.")

	if err := db.SetupSchema(database); err != nil {
		return err
	}
	info("Successfully set up new schema.")

	if err := status.InitializeMeta(database, blockchainMode); err != nil {
		return err
	}
	info("Successfully initialized database")

	return nil
}

// connectBlockchainRPCClient connects to blockchain RPC client specified in the given configuration.
func connectBlockchainRPCClient(rpcConfig RPCConfig) (external.RPCClient, error) {
	rpcEndpoint, err := cli.BuildEndpoint(rpcConfig.Host, rpcConfig.Port)
	if err != nil {
		return nil, err
	}

	client := jsonrpc.NewBlockchainClient(rpcEndpoint, rpcConfig.User, rpcConfig.Password, nil)

	// test if rpc client is active
	if err := waitForRPCClient(client); err != nil {
		return nil, err
	}

	return client, nil
}

//	@title			Dakar API
//	@version		1.0.4
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
	config.SetConfigFlags(defaultConfigName, &filePath, &createConfigFile)
	flag.Parse()

	////// CONFIGURATION FILE HANDLING //////

	if createConfigFile {
		fmt.Println("Generating configuration file ...")
		err := config.WriteConfig(defaultConfigName, defaultConfig)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("config file", defaultConfigName, "successfully created")
		return
	}

	var newConfig Config
	if err := config.ReadConfig(filePath, &newConfig); err != nil {
		fmt.Println(err)
		return
	}

	////// PRINT VERSION //////

	if commands.ShowVersion {
		printVersion(newConfig.BlockchainMode)
		return
	}

	////// CPU PROFILING //////

	if commands.CPUProfilePath != "" {
		f, err := os.Create(commands.CPUProfilePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		if err = pprof.StartCPUProfile(f); err != nil {
			fmt.Println(err)
			return
		}
		defer pprof.StopCPUProfile()
	}

	////// SETUP //////

	// setup Logging
	f, err := config.GetLogfile(newConfig.Logfile)
	if err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				fmt.Println(err)
			}
		}()
	} else if len(newConfig.Logfile) > 0 {
		fmt.Println("Could not create logfile", newConfig.Logfile)
		return
	}

	initAllLoggers(f)

	processorConfig, analyserConfig, err := selectConfig(newConfig.BlockchainMode)
	if err != nil {
		fmt.Printf("invalid blockchain mode: '%s', valid values are 'Dash' and 'Bitcoin'\n", newConfig.BlockchainMode)
		return
	}

	disableModules(analyserConfig, &newConfig)

	info("Blockchain mode: " + processorConfig.BlockchainName)

	////// CONNECT TO DATABASE //////

	endpoint, err := cli.BuildEndpoint(newConfig.Database.Host, newConfig.Database.Port)
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
		if err = resetDatabaseDialog(graphDB, newConfig.BlockchainMode); err != nil {
			warn(err)
		}
		return
	}

	if commands.UpgradeDatabase {
		if err = upgrades.UpgradeDatabase(graphDB); err != nil {
			warn(err)
		}
		return
	}

	// exit if no module is active (excluding the metrics module)
	if !newConfig.Modules.Classifier.Active && !newConfig.Modules.Crawler.Active &&
		!newConfig.Modules.HMI && !newConfig.Modules.FMI.Active &&
		!newConfig.Modules.HTTP.Active {
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

	if !checkMeta(graphDB, newConfig.BlockchainMode) {
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

	////// CONNECT TO RPC //////

	// Set up the RPC connection, only if needed
	var client external.RPCClient
	if newConfig.Modules.HTTP.Active || newConfig.Modules.Crawler.Active {
		client, err = connectBlockchainRPCClient(newConfig.RPC)
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
	if newConfig.Modules.Crawler.Active {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chCrawlingStopped <- true
			}()

			crawler := processor.NewCrawler(appContext, graphDB, client,
				newConfig.Modules.Crawler.InitialCacheSize, processorConfig)
			crawler.RegisterMetrics(prometheus.DefaultRegisterer)
			if processorErr := blockiterator.StartIteration(crawler, 0, nil); processorErr != nil {
				warn(processorErr)
			}
		}()
	}

	workspaceMutex := workspace.NewMutex()
	graphWrapper := graph.NewWrapper(appContext, graphDB)
	graphWrapper.RegisterMetrics(prometheus.DefaultRegisterer)
	w := workspace.NewWorker(workspaceMutex, graphDB, graphWrapper)
	w.RegisterMetrics(prometheus.DefaultRegisterer)

	var classifierStarted bool

	if newConfig.Modules.HTTP.Active && newConfig.Modules.Heuristics {
		// the classifier must be started after the in-memory graphs are loaded
		classifierStarted = true
		go func() {
			if graphErr := graphWrapper.LoadGraphs(); graphErr != nil {
				warn(graphErr)
				return
			}

			if newConfig.Modules.Classifier.Active {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if iterErr := blockiterator.StartIteration(graphWrapper, 0, nil); iterErr != nil {
						warn(iterErr)
					}
				}()
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer func() {
						chClassifyingStopped <- true
					}()
					classifier := analytics.NewClassifier(appContext, graphDB, analyserConfig)
					classifier.RegisterMetrics(prometheus.DefaultRegisterer)

					if classifierErr := blockiterator.StartIteration(classifier,
						time.Second*time.Duration(newConfig.Modules.Classifier.TargetDuration),
						nil); classifierErr != nil {
						warn(classifierErr)
					}
				}()
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			w.Start(appContext)
		}()
	}

	// activate classifier
	if newConfig.Modules.Classifier.Active && !classifierStarted {
		// in-memory graphs are not loaded -> start classifier
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chClassifyingStopped <- true
			}()
			classifier := analytics.NewClassifier(appContext, graphDB, analyserConfig)
			classifier.RegisterMetrics(prometheus.DefaultRegisterer)
			if classifierErr := blockiterator.StartIteration(classifier,
				time.Second*time.Duration(newConfig.Modules.Classifier.TargetDuration),
				nil); classifierErr != nil {
				warn(classifierErr)
			}
		}()
	}

	// activate HMI clustering
	if newConfig.Modules.HMI {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chHMIClusteringStopped <- true
			}()

			hmi := clustering.NewHierarchicalMultiInput(appContext, graphDB)
			hmi.RegisterMetrics(prometheus.DefaultRegisterer)
			if clusteringErr := blockiterator.StartIteration(hmi, 0, nil); clusteringErr != nil {
				warn(clusteringErr)
			}
		}()
	}

	// activate FMI clustering
	if newConfig.Modules.FMI.Active {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				chFMIClusteringStopped <- true
			}()
			fmi := clustering.NewFlatMultiInput(appContext, graphDB)
			fmi.RegisterMetrics(prometheus.DefaultRegisterer)
			if clusteringErr := blockiterator.StartIteration(fmi,
				time.Second*time.Duration(newConfig.Modules.FMI.TargetDuration),
				nil); clusteringErr != nil {
				warn(clusteringErr)
			}
		}()
	}

	// start api endpoint
	var apiHTTPServer *http.Server
	if newConfig.Modules.HTTP.Active {
		apiServer, serverErr := server.NewServer(workspaceMutex, graphDB, client, w, graphWrapper)
		if serverErr != nil {
			warn(serverErr)
		}

		wg.Add(1)
		apiHTTPServer = apiServer.StartServer(&wg, newConfig.Modules.HTTP.Port)
	}
	// start user api endpoint
	var userHTTPServer *http.Server
	if newConfig.Modules.HTTP.Active {
		wg.Add(1)
		userHTTPServer = userserver.NewServer(graphDB).StartServer(&wg, newConfig.Modules.User.Port)
	}

	// start metrics endpoint
	var metricsHTTPServer *http.Server
	if newConfig.Modules.Metrics.Active {
		wg.Add(1)
		metricsHTTPServer = server.StartMetrics(&wg, newConfig.Modules.Metrics.Port)
	}

	////// HANDLE SHUTDOWN //////

	var crawlerStopped = !newConfig.Modules.Crawler.Active
	var classifierStopped = !newConfig.Modules.Classifier.Active
	var clusteringHMIStopped = !newConfig.Modules.HMI
	var clusteringFMIStopped = !newConfig.Modules.FMI.Active
	var interrupted bool

	for !(interrupted || (crawlerStopped && classifierStopped && clusteringHMIStopped && clusteringFMIStopped)) {
		select {
		case <-chSignal:
			interrupted = true
			terminateApp()
			shutdownServer(apiHTTPServer)
			shutdownServer(userHTTPServer)
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

	if newConfig.Modules.HTTP.Active && crawlerStopped && classifierStopped &&
		clusteringHMIStopped && clusteringFMIStopped {
		// if the crawler, the classifier and clustering stopped working on their own accord,
		// the server is still active at this point

		<-chSignal
		shutdownServer(apiHTTPServer)
		shutdownServer(userHTTPServer)
		shutdownServer(metricsHTTPServer)
	}

	wg.Wait()
}
