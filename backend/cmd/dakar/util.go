// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"backend/db"
	"backend/db/status"
	"backend/external"
	"errors"
	"fmt"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type RPCConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type DatabaseConfig struct {
	Host string `yaml:"host"`
}

type CrawlerModule struct {
	Active           bool  `yaml:"active"`
	InitialCacheSize int64 `yaml:"initialCacheSize"`
}

type APIModule struct {
	Active bool `yaml:"active"`
	Port   uint `yaml:"port"`
}

type MetricsModule struct {
	Active bool `yaml:"active"`
	Port   uint `yaml:"port"`
}

type UserModule struct {
	Active bool `yaml:"active"`
	Port   uint `yaml:"port"`
}

type Classifier struct {
	Active         bool `yaml:"active"`
	TargetDuration int  `yaml:"targetDuration"`
}

type FMIModule struct {
	Active         bool `yaml:"active"`
	TargetDuration int  `yaml:"targetDuration"`
}

type ModulesConfig struct {
	HTTP       APIModule     `yaml:"api"`
	Metrics    MetricsModule `yaml:"metrics"`
	User       UserModule    `yaml:"user"`
	Crawler    CrawlerModule `yaml:"crawler"`
	FMI        FMIModule     `yaml:"fmi"`
	Classifier Classifier    `yaml:"classifier"`
	HMI        bool          `yaml:"hmi"`
	Heuristics bool          `yaml:"heuristics"`
}

type Config struct {
	// BlockchainMode controls various config parameters (see config.go).
	// Allowed values: "Dash" and "Bitcoin"
	BlockchainMode string         `yaml:"blockchainMode"`
	RPC            RPCConfig      `yaml:"rpc"`
	Database       DatabaseConfig `yaml:"database"`
	Modules        ModulesConfig  `yaml:"modules"`
}

var defaultConfig = Config{
	BlockchainMode: "",
	RPC: RPCConfig{
		Host:     "0.0.0.0:9998",
		User:     "rpc1user",
		Password: "1234pass",
	},
	Database: DatabaseConfig{
		Host: "0.0.0.0:9080",
	},
	Modules: ModulesConfig{
		HTTP: APIModule{
			Active: true,
			Port:   8081,
		},
		User: UserModule{
			Active: true,
			Port:   8085,
		},
		Metrics: MetricsModule{
			Active: true,
			Port:   8481,
		},
		Classifier: Classifier{
			Active:         true,
			TargetDuration: 10,
		},
		FMI: FMIModule{
			Active:         true,
			TargetDuration: 10,
		},
		Heuristics: false,
		Crawler:    CrawlerModule{Active: true, InitialCacheSize: 25000},
		HMI:        false,
	},
}

type Commands struct {
	ResetDB         bool
	IgnoreSafeGuard bool
	ShowVersion     bool
	UpgradeSchema   bool
	CPUProfilePath  string
}

// checks if a crawling process is already running
func isCrawling(c external.Database) (bool, error) {
	// short timeout as this is early in the execution process
	ctx, cancel := db.GetShortTaskContext()
	defer cancel()
	dbStatus, err := status.GetCrawlerStatus(ctx, c)
	if err != nil {
		// no status information found -> database is completely new
		// and thus no crawling is happening right now
		if errors.Is(err, status.ErrStatusNotFound) {
			return false, nil
		}

		return true, err
	} else if dbStatus.IsCrawling == nil {
		return true, serror.FromStr("was not able to get crawling status successfully")
	}

	return *dbStatus.IsCrawling, nil
}

// waitForRPCClient waits until the RPC client is ready to receive requests
func waitForRPCClient(client external.RPCClient) error {
	const maxRetries = 10
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	// set short time out for testing the connection and reset at end of function
	client.SetTimeout(time.Second)
	defer client.SetTimeout(0)

	for i := range maxRetries {
		_, err := client.GetBlockCount()
		if err == nil {
			if printedErrMessage {
				info("Successfully established connection to RPC client.")
			}
			return nil
		}

		if strings.Contains(err.Error(), "status code: 401") {
			return err
		}

		if !printedErrMessage {
			info("Waiting for RPC client to start")
			printedErrMessage = true
		}

		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}
	return serror.FromStr("RPC client is not ready to receive requests")
}

// shutdownServer sends a shutdown signal to the server with a timeout of 10 seconds
func shutdownServer(srv *http.Server) {
	if srv == nil {
		return
	}
	info("Shutting down server")

	ctx, cancel := db.GetShortTaskContext()
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		warn(serror.FromFormat("Server was shutdown and returned error: %w", err))
	}
}

// printVersion prints the version of the application and build information
func printVersion(blockchainMode string) {
	fmt.Println("Dakar", versionString, "compiled with", runtime.Version())
	fmt.Println("Blockchain mode:", blockchainMode)
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		fmt.Println("Modules:")
		for _, i := range buildInfo.Deps {
			moduleName := i.Path + " " + i.Version

			if i.Replace != nil {
				moduleName += " replaced with " + i.Replace.Path + " " + i.Replace.Version
			}
			fmt.Println(moduleName)
		}
	}
}

// checkMeta returns true if the blockchain mode and the schema version of the database match with the executable.
func checkMeta(c external.Database, blockchainMode string) bool {
	// short timeout as this is early in the execution process
	ctx, cancel := db.GetShortTaskContext()
	defer cancel()
	meta, err := status.GetMeta(ctx, c)
	if err != nil {
		warn(err)
		return false
	}

	// check if the blockchain mode of database matches the blockchain mode of the configuration
	if meta.BlockchainMode != blockchainMode {
		info("Database is using a different blockchain mode than the "+executableName+" configuration. You likely are connecting to the wrong database.",
			"database blockchain mode", meta.BlockchainMode,
			executableName+" blockchain mode", blockchainMode)
		return false
	}

	if meta.SchemaVersion == nil {
		info("database schema version is not set")
		return false
	}

	// check if the database schema version matches the schema version of the executable
	if *meta.SchemaVersion != db.SchemaVersion {
		// The log message looks wrong, but is right ("executable schema version", database.SchemaVersion)
		info("Database is using a different schema version than the executable. You may have to upgrade the database schema (CLI option: -upgradeschema) or use a different version of the executable.",
			"database schema version", *meta.SchemaVersion,
			"executable schema version", db.SchemaVersion)
		return false
	}

	return true
}
