package main

import (
	"backend/analytics"
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/clustering"
	"backend/db/status"
	"backend/external"
	"backend/processor"
	"backend/server"
	"backend/worker"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var thisLogger *slog.Logger

func initLogger(fileHandle *os.File) {
	var outputWriter io.Writer
	if fileHandle != nil {
		outputWriter = io.MultiWriter(fileHandle, os.Stdout)
	} else {
		outputWriter = os.Stdout
	}

	logger := slog.New(slog.NewTextHandler(outputWriter, nil))
	slog.SetDefault(logger)

	thisLogger = slog.With(slog.String("module", "dbUpgrade"))

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
	cli.LogError(thisLogger, err, v...)
}

type Config struct {
	Logfile string `yaml:"logfile"`
	Host    string `yaml:"host"`
	Port    uint   `yaml:"port"`
}

var defaultConfig = Config{
	Logfile: "",
	Host:    "0.0.0.0",
	Port:    9080,
}

// Simple utility to browse/lookup the TXs from the database
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

	initLogger(f)

	endpoint, err := cli.BuildEndpoint(config.Host, config.Port)
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

	isSet, err := db.IsSchemaSet(dgraph)
	if err != nil {
		warn(err)
		return
	}

	if !isSet {
		info("Schema is not set")
		return
	}

	info("adding workspaces starting ...")
	err = db.AlterSchemaAddWorkspaces(dgraph)
	if err != nil {
		warn(err)
		return
	}
	info("adding workspaces finished")

	info("increasing schema version ...")
	err = status.SetSchemaVersion(dgraph, 4)
	if err != nil {
		warn(err)
		return
	}
	info("increased schema version")

	info("deleting all FMI clusters ...")
	err = clustering.DeleteAllFMIClusters(dgraph)
	if err != nil {
		warn(err)
		return
	}
	info("deleted all FMI clusters")

	info("resetting FMI cluster status ...")
	zero := uint64(0)
	err = status.SetClusteringFMIStatus(dgraph, status.ClusteringFlatMultiInputStatus{
		LastClusteredBlockID: &zero,
	})
	if err != nil {
		warn(err)
		return
	}
	info("reset FMI cluster status")

	info("dropping hex starting ...")
	err = db.DropPredicateHex(dgraph)
	if err != nil {
		warn(err)
		return
	}
	info("dropping hex finished")

	info("removing hex starting ...")
	err = db.AlterSchemaRemoveHex(dgraph)
	if err != nil {
		warn(err)
		return
	}
	info("removing hex finished")

	info("increasing schema version ...")
	err = status.SetSchemaVersion(dgraph, 5)
	if err != nil {
		warn(err)
		return
	}
	info("increased schema version")
}
