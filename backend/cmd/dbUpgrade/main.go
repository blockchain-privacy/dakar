package main

import (
	cli "backend/cmd/cliutil"
	"backend/db"
	"backend/db/status"
	"backend/external"
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

	info("dropping roles starting ...")
	err = db.AlterSchemaDropRoles(dgraph)
	if err != nil {
		warn(err)
		return
	}
	info("dropping roles finished")

	info("altering user starting ...")
	err = db.AlterSchemaDropUserPredicates(dgraph)
	if err != nil {
		warn(err)
		return
	}
	info("altering user finished")

	info("increasing schema version ...")
	err = status.SetSchemaVersion(dgraph, 3)
	if err != nil {
		warn(err)
		return
	}
	info("increased schema version")
}
