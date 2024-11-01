package main

import (
	"backend/analytics"
	"backend/analytics/graph"
	"backend/db"
	"backend/db/analytics/clustering"
	"backend/external"
	"backend/processor"
	"backend/server"
	"backend/workspace"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/qrest/gomisc/config"
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
	workspace.InitLogger()
}

func info(msg string, v ...any) {
	thisLogger.Info(msg, v...)
}

func warn(err error, v ...any) {
	serror.Log(thisLogger, err, v...)
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
	DBHost:  "0.0.0.0:9080",
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

	// setup Logging
	var f *os.File
	if len(newConfig.Logfile) > 0 {
		var err error
		if f, err = config.GetLogfile(newConfig.Logfile); err == nil {
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

	// create dgraph client
	dgraph, c, err := external.CreateClient(newConfig.DBHost)
	if err != nil {
		warn(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			warn(err)
		}
	}()

	ctx, cancel := db.GetLongTaskContext()
	defer cancel()

	if newConfig.PrivacyCharts.Active {
		exportTransactionData(ctx, dgraph, newConfig.PrivacyCharts.Directory)
	}

	if newConfig.UniqueAddresses.Active {
		doUniqueAddressAnalysis(ctx, dgraph, newConfig.UniqueAddresses.Option, "uniqueAddresses")
	}

	var g *graph.ReversibleGraph

	if newConfig.TimestampAnalytics.ExportMixingTransactions ||
		newConfig.TimestampAnalytics.ExportDestinationTransactions ||
		newConfig.TimestampAnalytics.ExportReverseLookup.Active ||
		newConfig.ExclusionSimulations.Active ||
		newConfig.OriginGap.Active ||
		newConfig.DestinationCount.Active {
		g, err = graph.LoadTransactionGraph(ctx, graph.NewDashConfig(), dgraph, 0)
		if err != nil && !errors.Is(err, graph.ErrDBContainsNoPrivacyTransactions) {
			warn(err)
			return
		}
	}

	if newConfig.TimestampAnalytics.ExportDestinationTransactions {
		doDestinationTimestampAnalysis(g)
	}

	if newConfig.TimestampAnalytics.ExportMixingTransactions {
		exportMixingTimestamps(g, true)
	}

	if newConfig.TimestampAnalytics.ExportReverseLookup.Active {
		exportReverseLookup(g, newConfig.TimestampAnalytics.ExportReverseLookup.NodeID,
			newConfig.TimestampAnalytics.ExportReverseLookup.LookBackTimeHours,
			nil, false, false)
	}

	if newConfig.ExclusionSimulations.Active {
		doSimulation(ctx, dgraph, g, newConfig.ExclusionSimulations.NodeID,
			newConfig.ExclusionSimulations.UserUID, newConfig.ExclusionSimulations.LookBackTimeHours)
	}

	if newConfig.OriginGap.Active {
		doOriginGapAnalysis(g, time.Hour*time.Duration(newConfig.OriginGap.MinGapHours), newConfig.OriginGap.Filename)
	}

	if newConfig.ExportBlocks.Active {
		doExportBlocks(ctx, dgraph, newConfig.ExportBlocks.Filename,
			newConfig.ExportBlocks.StartBlock, newConfig.ExportBlocks.EndBlock)
	}

	if newConfig.ExportPrivacyTransactions.Active {
		doExportPrivacyTransactions(ctx, dgraph, newConfig.ExportPrivacyTransactions.Filename,
			newConfig.ExportPrivacyTransactions.StartTransaction)
	}

	if newConfig.DestinationCount.Active {
		doDestinationCountAnalysis(ctx, dgraph, g, newConfig.DestinationCount.Filename)
	}

	if newConfig.ExportClusterActivity.Active {
		doExportClusterActivity(ctx, dgraph, newConfig.ExportClusterActivity.Filename)
	}
}

func getBlockRange(ctx context.Context, dgraph external.Database, firstBlock int, lastBlock int) ([]db.Block, error) {
	numBlocks := lastBlock - firstBlock
	if numBlocks <= 0 {
		return nil, nil
	}

	blocks := make([]db.Block, numBlocks+1)

	for i := firstBlock; i <= lastBlock; i++ {
		block, err := db.GetFullBlock(ctx, dgraph, i, true)
		if err != nil {
			return nil, err
		}
		blocks[i-firstBlock] = block
	}

	return blocks, nil
}

func getAddressRange(ctx context.Context, dgraph external.Database, firstBlock int, lastBlock int) ([]db.Address, error) {
	numBlocks := lastBlock - firstBlock
	if numBlocks <= 0 {
		return nil, nil
	}

	return db.GetAddressesByBlockRange(ctx, dgraph, firstBlock, lastBlock, true)
}

func getClusterRange(ctx context.Context, dgraph external.Database, firstBlock int, lastBlock int) ([]clustering.Cluster, error) {
	numBlocks := lastBlock - firstBlock
	if numBlocks <= 0 {
		return nil, nil
	}

	return clustering.GetClustersByBlockRange(ctx, dgraph, firstBlock, lastBlock, true)
}
