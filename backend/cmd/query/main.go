package main

import (
	"backend/analytics/graph"
	cli "backend/cmd/cliutil"
	"backend/db"
	"flag"
	"fmt"
	"log"
	"time"
)

var thisLogger *log.Logger

func initLogger() {
	thisLogger = log.New(log.Writer(), "\033[0;31mquery\033[0m\t", log.Flags())
}
func info(v ...interface{}) {
	thisLogger.Println(v...)
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

type Config struct {
	Logfile              string                    `yaml:"logfile"`
	DBHost               string                    `yaml:"host"`
	DBPort               uint                      `yaml:"port"`
	PrivacyCharts        PrivacyChartModule        `yaml:"privacyCharts"`
	UniqueAddresses      UniqueAddressesModule     `yaml:"uniqueAddresses"`
	TimestampAnalytics   TimestampAnalyticsModule  `yaml:"timestampAnalytics"`
	ExclusionSimulations ExclusionSimulationModule `yaml:"exclusionSimulations"`
	OriginGap            OriginGapModule           `yaml:"originGap"`
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
}

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
	if len(config.Logfile) > 0 {
		if f, err := cli.GetLogfile(config.Logfile); err == nil {
			defer func() {
				if err = f.Close(); err != nil {
					fmt.Println(err)
				}
			}()
		}
	}

	initLogger()

	endpoint, err := cli.BuildEndpoint(config.DBHost, config.DBPort)
	if err != nil {
		info(err)
		return
	}

	// create dgraph client
	dgraph, c, err := db.CreateClient(endpoint)
	if err != nil {
		info(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			info(err)
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
		config.OriginGap.Active {
		// todo set to zero
		g, err = graph.LoadTransactionGraph(dgraph, 0)
		if err != nil {
			info(err)
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
			config.TimestampAnalytics.ExportReverseLookup.LookBackTimeHours, nil)
	}

	if config.ExclusionSimulations.Active {
		doSimulation(dgraph, g, config.ExclusionSimulations.NodeID,
			config.ExclusionSimulations.UserUID, config.ExclusionSimulations.LookBackTimeHours)
	}

	if config.OriginGap.Active {
		doOriginGapAnalysis(g, time.Hour*time.Duration(config.OriginGap.MinGapHours), config.OriginGap.Filename)
	}
}
