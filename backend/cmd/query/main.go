package main

import (
	"backend/analytics"
	"backend/analytics/graph"
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/clustering"
	"backend/external"
	"backend/processor"
	"backend/server"
	"backend/worker"
	"flag"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"io"
	"log"
	"log/slog"
	"os"
	"time"
)

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

	thisLogger = slog.With(slog.String("module", "query"))

	analytics.InitLogger()
	db.InitLogger()
	processor.InitLogger()
	server.InitLogger()
	worker.InitLogger()
}

func info(msg string, v ...any) {
	thisLogger.Info(msg, v...)
}

func warn(err error, v ...any) {
	serror.LogError(thisLogger, err, v...)
}

type UniqueAddressesModule struct {
	Active bool `yaml:"active"`
	Option int  `yaml:"option"`
}

type PrivacyChartModule struct {
	Active    bool   `yaml:"active"`
	Directory string `yaml:"directory"`
}

type ExclusionSimulationModule struct {
	Active            bool   `yaml:"active"`
	UserUID           string `yaml:"userUID"`
	LookBackTimeHours int    `yaml:"lookBackTimeHours"`
	NodeID            string `yaml:"nodeID"`
}

type ExportReverseLookupModule struct {
	Active            bool   `yaml:"active"`
	LookBackTimeHours int    `yaml:"lookBackTimeHours"`
	NodeID            string `yaml:"nodeID"`
}

type TimestampAnalyticsModule struct {
	ExportDestinationTransactions bool                      `yaml:"exportDestinationTransactions"`
	ExportMixingTransactions      bool                      `yaml:"exportMixingTransactions"`
	ExportReverseLookup           ExportReverseLookupModule `yaml:"exportReverseLookup"`
}

type OriginGapModule struct {
	Active      bool   `yaml:"active"`
	Filename    string `yaml:"filename"`
	MinGapHours int    `yaml:"minGapHours"`
}

type ExportBlocksModule struct {
	Active     bool   `yaml:"active"`
	Filename   string `yaml:"filename"`
	StartBlock int    `yaml:"startBlock"`
	EndBlock   int    `yaml:"endBlock"`
}

type DestinationCountModule struct {
	Active   bool   `yaml:"active"`
	Filename string `yaml:"filename"`
}

type ExportPrivacyTransactionsModule struct {
	Active           bool   `yaml:"active"`
	Filename         string `yaml:"filename"`
	StartTransaction string `yaml:"startTransaction"`
}

type ExportClusterActivityModule struct {
	Active   bool   `yaml:"active"`
	Filename string `yaml:"filename"`
}

type Config struct {
	Logfile                   string                          `yaml:"logfile"`
	DBHost                    string                          `yaml:"host"`
	DBPort                    uint                            `yaml:"port"`
	PrivacyCharts             PrivacyChartModule              `yaml:"privacyCharts"`
	UniqueAddresses           UniqueAddressesModule           `yaml:"uniqueAddresses"`
	TimestampAnalytics        TimestampAnalyticsModule        `yaml:"timestampAnalytics"`
	ExclusionSimulations      ExclusionSimulationModule       `yaml:"exclusionSimulations"`
	OriginGap                 OriginGapModule                 `yaml:"originGap"`
	ExportBlocks              ExportBlocksModule              `yaml:"exportBlocks"`
	ExportPrivacyTransactions ExportPrivacyTransactionsModule `yaml:"exportPrivacyTransactions"`
	DestinationCount          DestinationCountModule          `yaml:"destinationCount"`
	ExportClusterActivity     ExportClusterActivityModule     `yaml:"exportClusterActivity"`
}

var defaultConfig = Config{
	Logfile: "",
	DBHost:  "0.0.0.0",
	DBPort:  9080,
	PrivacyCharts: PrivacyChartModule{
		Active:    false,
		Directory: "",
	},
	UniqueAddresses: UniqueAddressesModule{
		Active: false,
		Option: 0,
	},
	TimestampAnalytics: TimestampAnalyticsModule{
		ExportDestinationTransactions: false,
		ExportMixingTransactions:      false,
		ExportReverseLookup: ExportReverseLookupModule{
			Active:            false,
			LookBackTimeHours: 0,
			NodeID:            "",
		},
	},
	ExclusionSimulations: ExclusionSimulationModule{
		Active:            false,
		UserUID:           "",
		LookBackTimeHours: 0,
		NodeID:            "",
	},
	OriginGap: OriginGapModule{
		Active:      false,
		Filename:    "",
		MinGapHours: 0,
	},
	ExportBlocks: ExportBlocksModule{
		Active:     false,
		Filename:   "",
		StartBlock: 0,
		EndBlock:   0,
	},
	ExportPrivacyTransactions: ExportPrivacyTransactionsModule{
		Active:           false,
		Filename:         "",
		StartTransaction: "",
	},
	DestinationCount: DestinationCountModule{
		Active:   false,
		Filename: "",
	},
	ExportClusterActivity: ExportClusterActivityModule{
		Active:   false,
		Filename: "",
	},
}

//nolint:gocyclo
func main() {
	////// SET FLAGS //////

	defaultConfigName := "config.yml"
	var filePath string
	var createConfigFile bool
	cli.SetConfigFlags(defaultConfigName, &filePath, &createConfigFile)
	flag.Parse()

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

	// setup Logging
	var f *os.File
	if len(config.Logfile) > 0 {
		var err error
		if f, err = cli.GetLogfile(config.Logfile); err == nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.SetOutput(io.MultiWriter(os.Stdout, f))
			defer func() {
				if err = f.Close(); err != nil {
					fmt.Println(err)
				}
			}()
		} else {
			fmt.Println("error setting up log file")
		}
	}

	initAllLoggers(f)

	endpoint, err := cli.BuildEndpoint(config.DBHost, config.DBPort)
	if err != nil {
		warn(err)
		return
	}

	// create dgraph client
	dgraph, c, err := external.CreateClient(endpoint)
	if err != nil {
		warn(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			warn(err)
		}
	}()

	if config.PrivacyCharts.Active {
		exportTransactionData(dgraph, config.PrivacyCharts.Directory)
	}

	if config.UniqueAddresses.Active {
		doUniqueAddressAnalysis(dgraph, config.UniqueAddresses.Option, "uniqueAddresses")
	}

	var g *graph.ReversibleGraph

	if config.TimestampAnalytics.ExportMixingTransactions ||
		config.TimestampAnalytics.ExportDestinationTransactions ||
		config.TimestampAnalytics.ExportReverseLookup.Active ||
		config.ExclusionSimulations.Active ||
		config.OriginGap.Active ||
		config.DestinationCount.Active {
		g, err = graph.LoadTransactionGraph(dgraph, 0)
		if err != nil {
			warn(err)
			return
		}
	}

	if config.TimestampAnalytics.ExportDestinationTransactions {
		doDestinationTimestampAnalysis(g)
	}

	if config.TimestampAnalytics.ExportMixingTransactions {
		exportMixingTimestamps(g, true)
	}

	if config.TimestampAnalytics.ExportReverseLookup.Active {
		exportReverseLookup(g, config.TimestampAnalytics.ExportReverseLookup.NodeID,
			config.TimestampAnalytics.ExportReverseLookup.LookBackTimeHours,
			nil, false, false)
	}

	if config.ExclusionSimulations.Active {
		doSimulation(dgraph, g, config.ExclusionSimulations.NodeID,
			config.ExclusionSimulations.UserUID, config.ExclusionSimulations.LookBackTimeHours)
	}

	if config.OriginGap.Active {
		doOriginGapAnalysis(g, time.Hour*time.Duration(config.OriginGap.MinGapHours), config.OriginGap.Filename)
	}

	if config.ExportBlocks.Active {
		doExportBlocks(dgraph, config.ExportBlocks.Filename, config.ExportBlocks.StartBlock, config.ExportBlocks.EndBlock)
	}

	if config.ExportPrivacyTransactions.Active {
		doExportPrivacyTransactions(dgraph, config.ExportPrivacyTransactions.Filename,
			config.ExportPrivacyTransactions.StartTransaction)
	}

	if config.DestinationCount.Active {
		doDestinationCountAnalysis(dgraph, g, config.DestinationCount.Filename)
	}

	if config.ExportClusterActivity.Active {
		doExportClusterActivity(dgraph, config.ExportClusterActivity.Filename)
	}
}

func getBlockRange(dgraph external.Database, firstBlock int, lastBlock int) ([]db.Block, error) {
	numBlocks := lastBlock - firstBlock
	if numBlocks <= 0 {
		return nil, nil
	}

	blocks := make([]db.Block, numBlocks+1)

	for i := firstBlock; i <= lastBlock; i++ {
		block, err := db.GetFullBlock(dgraph, i, true)
		if err != nil {
			return nil, err
		}
		blocks[i-firstBlock] = block
	}

	return blocks, nil
}

func getAddressRange(dgraph external.Database, firstBlock int, lastBlock int) ([]db.Address, error) {
	numBlocks := lastBlock - firstBlock
	if numBlocks <= 0 {
		return nil, nil
	}

	return db.GetAddressesByBlockRange(dgraph, firstBlock, lastBlock, true)
}

func getClusterRange(dgraph external.Database, firstBlock int, lastBlock int) ([]clustering.Cluster, error) {
	numBlocks := lastBlock - firstBlock
	if numBlocks <= 0 {
		return nil, nil
	}

	return clustering.GetClustersByBlockRange(dgraph, firstBlock, lastBlock, true)
}
