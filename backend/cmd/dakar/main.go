// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"backend/analytics/classifier"
	"backend/analytics/clustering"
	"backend/analytics/graph"
	"backend/blockiterator"
	"backend/constants"
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

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/blockchain-privacy/gomisc/config"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// versionString displays the version of the Crawler
const versionString = "v1.0.0"

// name of the executable
const executableName = "crawler"

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "main"}, v...)...)
}

func warn(err error, v ...any) {
	serror.Log(slog.Default(), err, v...)
}

func setCommandFlags(c *Commands) {
	flag.BoolVar(&c.ResetDB, "reset", false, "Remove all data from the database (default: false)")
	flag.BoolVar(&c.IgnoreSafeGuard, "ignoresafeguard", false, "Ignore the crawling safe guard (default: false)")
	flag.BoolVar(&c.ShowVersion, "version", false, "Show version information (default: false)")
	flag.BoolVar(&c.UpgradeSchema, "upgradeschema", false, "Upgrade the database schema to the newest version (default: false)")
	flag.StringVar(&c.CPUProfilePath, "cpuprofile", "", "Path where the cpu profile should be stored (default: <empty>)")
}

type iteratorConfigurations struct {
	processor  processor.Config
	classifier classifier.Config
	clustering clustering.Config
	graph      graph.Config
}

// selectConfig returns iterator configurations based on the given blockchain mode.
func selectConfig(blockchainMode string) (*iteratorConfigurations, error) {
	switch blockchainMode {
	case constants.BlockchainModeDash:
		return &iteratorConfigurations{
			processor:  processor.NewDashConfig(),
			classifier: classifier.NewDashConfig(),
			graph:      graph.NewDashConfig(),
			clustering: clustering.NewDashConfig(),
		}, nil
	case constants.BlockchainModeBTC:
		return &iteratorConfigurations{
			processor:  processor.NewBitcoinConfig(),
			classifier: classifier.NewBTCConfig(),
			graph:      graph.NewBTCConfig(),
			clustering: clustering.NewBTCConfig(),
		}, nil
	default:
		return nil, serror.FromStr("invalid blockchain mode")
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

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	if err := database.DropAll(ctx); err != nil {
		return err
	}
	info("Dropped all data.")

	if err := db.SetupSchema(database); err != nil {
		return err
	}
	info("Successfully set up new schema.")

	if err := status.InitializeMeta(ctx, database, blockchainMode); err != nil {
		return err
	}
	info("Successfully initialized database")

	return nil
}

// connectBlockchainRPCClient connects to blockchain RPC client specified in the given configuration.
func connectBlockchainRPCClient(rpcConfig RPCConfig) (external.RPCClient, error) {
	client := jsonrpc.NewBlockchainClient(rpcConfig.Host, rpcConfig.User, rpcConfig.Password, nil)

	// test if rpc client is active
	if err := waitForRPCClient(client); err != nil {
		return nil, err
	}

	return client, nil
}

//	@title			Dakar API
//	@version		1.0.5
//	@description	This is the REST API for Dakar
//	@host			localhost
//	@BasePath		/api/v1
//
// This is the backend of Dakar. It crawls, classifies and clusters either the Dash
// or Bitcoin blockchain and exposes its data via a RESTful API.
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

	////// LOGGING //////

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	////// CONNECT TO DATABASE //////

	graphDB, err := external.CreateClient(newConfig.Database.Host)
	if err != nil {
		warn(err)
		return
	}
	defer graphDB.Close()

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

	if commands.UpgradeSchema {
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
			// only exit if crawler is supposed to be active
			if newConfig.Modules.Crawler.Active {
				warn(err)
				return
			}

			info("Unable to connect to RPC client. Continuing ...")
		}
	}

	////// START MODULES //////

	iterConfigs, err := selectConfig(newConfig.BlockchainMode)
	if err != nil {
		fmt.Printf("invalid blockchain mode: '%s', valid values are '%s' and '%s'\n",
			newConfig.BlockchainMode, constants.BlockchainModeDash, constants.BlockchainModeBTC)
		return
	}

	info("Blockchain mode: " + newConfig.BlockchainMode)

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
				newConfig.Modules.Crawler.InitialCacheSize, iterConfigs.processor)
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
			if graphErr := graphWrapper.LoadGraphs(iterConfigs.graph); graphErr != nil {
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
					cl := classifier.NewClassifier(appContext, graphDB, iterConfigs.classifier)
					cl.RegisterMetrics(prometheus.DefaultRegisterer)

					if classifierErr := blockiterator.StartIteration(cl,
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
			cl := classifier.NewClassifier(appContext, graphDB, iterConfigs.classifier)
			cl.RegisterMetrics(prometheus.DefaultRegisterer)
			if classifierErr := blockiterator.StartIteration(cl,
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
			fmi := clustering.NewFlatMultiInput(appContext, graphDB, iterConfigs.clustering)
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

	for !interrupted && (!crawlerStopped || !classifierStopped || !clusteringHMIStopped || !clusteringFMIStopped) {
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
		terminateApp()
		shutdownServer(apiHTTPServer)
		shutdownServer(userHTTPServer)
		shutdownServer(metricsHTTPServer)
	}

	wg.Wait()
}
