// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics/clustering"
	"gitlab.com/blockchain-privacy/dakar/db/status"
	"gitlab.com/blockchain-privacy/dakar/external"
	"log/slog"
	"os"
	"strconv"
	"time"

	"gitlab.com/blockchain-privacy/gomisc/config"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "main"}, v...)...)
}

func warn(err error, v ...any) {
	serror.Log(slog.Default(), err, v...)
}

type UniqueAddressesModule struct {
	Active bool `yaml:"active"`
	Option int  `yaml:"option"`
}

type PrivacyChartModule struct {
	Active    bool   `yaml:"active"`
	Directory string `yaml:"directory"`
}

type StatsModule struct {
	Active                 bool   `yaml:"active"`
	Filename               string `yaml:"filename"`
	ExcludeTransactionType string `yaml:"excludeTransactionType"`
}

type TimestampAnalyticsModule struct {
	Active          bool   `yaml:"active"`
	Filename        string `yaml:"filename"`
	TransactionType string `yaml:"transactionType"`
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

type ExportTransactionsModule struct {
	Active           bool     `yaml:"active"`
	Filename         string   `yaml:"filename"`
	StartBlock       int64    `yaml:"startBlock"`
	EndBlock         int64    `yaml:"endBlock"`
	TransactionTypes []string `yaml:"transactionTypes"`
}

type DestinationCountModule struct {
	Active          bool   `yaml:"active"`
	Filename        string `yaml:"filename"`
	TransactionType string `yaml:"transactionType"`
}

type ExportPrivacyGraphModule struct {
	Active           bool   `yaml:"active"`
	Filename         string `yaml:"filename"`
	StartTransaction string `yaml:"startTransaction"`
}

type ExportClusterActivityModule struct {
	Active   bool   `yaml:"active"`
	Filename string `yaml:"filename"`
}

type TagPackInserterModule struct {
	Active    bool   `yaml:"active"`
	Directory string `yaml:"directory"`
}

type CollateralsModule struct {
	Active     bool   `yaml:"active"`
	CpFilename string `yaml:"cpFilename"`
	CcFilename string `yaml:"ccFilename"`
}

type CollateralGapModule struct {
	Active   bool   `yaml:"active"`
	Filename string `yaml:"filename"`
}

type HeuristicAnalysisModule struct {
	Active          bool   `yaml:"active"`
	Filename        string `yaml:"filename"`
	TransactionType string `yaml:"transactionType"`
}

type Config struct {
	DBHost                string                      `yaml:"host"`
	PrivacyCharts         PrivacyChartModule          `yaml:"privacyCharts"`
	UniqueAddresses       UniqueAddressesModule       `yaml:"uniqueAddresses"`
	TimestampAnalytics    TimestampAnalyticsModule    `yaml:"timestampAnalytics"`
	OriginGap             OriginGapModule             `yaml:"originGap"`
	ExportBlocks          ExportBlocksModule          `yaml:"exportBlocks"`
	ExportTransactions    ExportTransactionsModule    `yaml:"exportTransactions"`
	ExportPrivacyGraph    ExportPrivacyGraphModule    `yaml:"exportPrivacyGraph"`
	DestinationCount      DestinationCountModule      `yaml:"destinationCount"`
	ExportClusterActivity ExportClusterActivityModule `yaml:"exportClusterActivity"`
	TagPackInserter       TagPackInserterModule       `yaml:"tagPackInserter"`
	Stats                 StatsModule                 `yaml:"stats"`
	Collaterals           CollateralsModule           `yaml:"collaterals"`
	CollateralGap         CollateralGapModule         `yaml:"collateralGap"`
	HeuristicAnalysis     HeuristicAnalysisModule     `yaml:"heuristicAnalysis"`
}

var defaultConfig = Config{
	DBHost: "0.0.0.0:9080",
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

	var cfg Config
	if err := config.ReadConfig(filePath, &cfg); err != nil {
		fmt.Println(err)
		return
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// create dgraph client
	dgraph, err := external.CreateClient(cfg.DBHost)
	if err != nil {
		warn(err)
		return
	}
	defer dgraph.Close()

	ctx := context.Background()
	if cfg.PrivacyCharts.Active {
		exportTransactionData(ctx, dgraph, cfg.PrivacyCharts.Directory)
	}

	if cfg.UniqueAddresses.Active {
		doUniqueAddressAnalysis(ctx, dgraph, cfg.UniqueAddresses.Option, "uniqueAddresses")
	}

	var g *graph.ReversibleGraph

	if cfg.TimestampAnalytics.Active ||
		cfg.OriginGap.Active ||
		cfg.DestinationCount.Active ||
		cfg.HeuristicAnalysis.Active {
		meta, err := status.GetMeta(ctx, dgraph)
		if err != nil {
			warn(err)
			return
		}

		var graphConfig graph.Config
		switch meta.BlockchainMode {
		case constants.BlockchainModeDash:
			graphConfig = graph.NewDashConfig()
		case constants.BlockchainModeBTC:
			graphConfig = graph.NewBTCConfig()
		default:
			warn(serror.FromStrWithContext("invalid blockchain mode", "mode", meta.BlockchainMode))
			return
		}

		numTxToLoad := 0

		if graphLimit, ok := os.LookupEnv("DEV_GRAPH_LIMIT"); ok {
			numGraphLimit, err := strconv.Atoi(graphLimit)
			if err != nil {
				warn(serror.FromFormat("DEV_GRAPH_LIMIT is not a number: %w", err))
				return
			}
			if numGraphLimit < 0 {
				info("DEV_GRAPH_LIMIT environment variable is negative. Exiting ...")
				return
			} else if numGraphLimit == 0 {
				info("DEV_GRAPH_LIMIT environment variable is set to zero. Ignoring ...")
			} else {
				info(fmt.Sprintf("DEV_GRAPH_LIMIT environment variable is set. Limiting in-memory mixing graph to %d transactions", numGraphLimit))
				numTxToLoad = numGraphLimit
			}
		}

		g, err = graph.LoadTransactionGraph(ctx, graphConfig, dgraph, numTxToLoad)
		if err != nil && !errors.Is(err, graph.ErrDBContainsNoClassifiedTransactions) {
			warn(err)
			return
		}
	}

	if cfg.TimestampAnalytics.Active {
		doDestinationTimestampAnalysis(g, cfg.TimestampAnalytics.TransactionType, cfg.TimestampAnalytics.Filename)
	}

	if cfg.OriginGap.Active {
		doOriginGapAnalysis(g, time.Hour*time.Duration(cfg.OriginGap.MinGapHours), cfg.OriginGap.Filename)
	}

	if cfg.ExportBlocks.Active {
		doExportBlocks(ctx, dgraph, cfg.ExportBlocks.Filename,
			cfg.ExportBlocks.StartBlock, cfg.ExportBlocks.EndBlock)
	}

	if cfg.ExportTransactions.Active {
		doExportTransactions(ctx, dgraph, cfg.ExportTransactions.Filename,
			cfg.ExportTransactions.StartBlock, cfg.ExportTransactions.EndBlock,
			cfg.ExportTransactions.TransactionTypes)
	}

	if cfg.ExportPrivacyGraph.Active {
		doExportPrivacyGraph(ctx, dgraph, cfg.ExportPrivacyGraph.Filename,
			cfg.ExportPrivacyGraph.StartTransaction)
	}

	if cfg.DestinationCount.Active {
		doDestinationCountAnalysis(ctx, dgraph, g, cfg.DestinationCount.Filename, cfg.DestinationCount.TransactionType)
	}

	if cfg.ExportClusterActivity.Active {
		doExportClusterActivity(ctx, dgraph, cfg.ExportClusterActivity.Filename)
	}

	if cfg.Stats.Active {
		doStats(ctx, dgraph, cfg.Stats.Filename, cfg.Stats.ExcludeTransactionType)
	}

	if cfg.TagPackInserter.Active {
		doInsertTagPacks(ctx, dgraph, cfg.TagPackInserter.Directory)
	}

	if cfg.Collaterals.Active {
		doCollateralAnalysis(ctx, dgraph, cfg.Collaterals.CpFilename, cfg.Collaterals.CcFilename)
	}

	if cfg.CollateralGap.Active {
		doCollateralGapAnalysis(ctx, dgraph, cfg.CollateralGap.Filename)
	}

	if cfg.HeuristicAnalysis.Active {
		doHeuristicAnalysis(ctx, dgraph, g, cfg.HeuristicAnalysis.Filename, cfg.HeuristicAnalysis.TransactionType)
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
