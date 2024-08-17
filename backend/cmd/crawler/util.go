package main

import (
	database "backend/db"
	"backend/db/status"
	"backend/external"
	"context"
	"errors"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type RPCConfig struct {
	Host     string `yaml:"host"`
	Port     uint   `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type DatabaseConfig struct {
	Host string `yaml:"host"`
	Port uint   `yaml:"port"`
}

type CrawlerModule struct {
	Active           bool  `yaml:"active"`
	InitialCacheSize int64 `yaml:"initialCacheSize"`
}

type FMIModule struct {
	Active    bool   `yaml:"active"`
	MaxBlocks uint64 `yaml:"maxBlocks"`
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

type ModulesConfig struct {
	HTTP       APIModule     `yaml:"api"`
	Metrics    MetricsModule `yaml:"metrics"`
	User       UserModule    `yaml:"user"`
	Crawler    CrawlerModule `yaml:"crawler"`
	FMI        FMIModule     `yaml:"fmi"`
	HMI        bool          `yaml:"hmi"`
	Classifier bool          `yaml:"classifier"`
	Heuristics bool          `yaml:"heuristics"`
}

type Config struct {
	Logfile string `yaml:"logfile"`
	// BlockchainMode controls various config parameters (see config.go).
	// Allowed values: "Dash" and "Bitcoin"
	BlockchainMode string         `yaml:"blockchainMode"`
	RPC            RPCConfig      `yaml:"rpc"`
	Database       DatabaseConfig `yaml:"database"`
	Modules        ModulesConfig  `yaml:"modules"`
}

var defaultConfig = Config{
	Logfile:        "dakar.log",
	BlockchainMode: "",
	RPC: RPCConfig{
		Host:     "0.0.0.0",
		Port:     9998,
		User:     "rpc1user",
		Password: "1234pass",
	},
	Database: DatabaseConfig{
		Host: "0.0.0.0",
		Port: 9080,
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
		Classifier: false,
		Heuristics: false,
		Crawler:    CrawlerModule{Active: true, InitialCacheSize: 25000},
		FMI: FMIModule{
			Active:    false,
			MaxBlocks: 10,
		},
		HMI: false,
	},
}

type Commands struct {
	ResetDB         bool
	IgnoreSafeGuard bool
	ShowVersion     bool
	CPUProfilePath  string
}

// checks if a crawling process is already running
func isCrawling(db external.Database) (bool, error) {
	dbStatus, err := status.GetCrawlerStatus(db)
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
	const maxRetries = 5
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := range maxRetries {
		_, err := client.GetBlockCount()
		if err == nil {
			if printedErrMessage {
				info("Successfully established connection to RPC client.")
			}
			return nil
		}

		if strings.Contains(err.Error(), "status code: 401") {
			return serror.FromFormat("Authentication error: %w", err)
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

// shutdownServer sends a shutdown signal to the server with a timout of 10 seconds
func shutdownServer(srv *http.Server) {
	if srv == nil {
		return
	}
	info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
func checkMeta(db external.Database, blockchainMode string) bool {
	meta, err := status.GetMeta(db)
	if err != nil {
		warn(err)
		return false
	}

	// check if the blockchain mode of database matches the blockchain mode of the configuration
	if meta.BlockchainMode != blockchainMode {
		info("Database is using a different blockchain mode than the "+executableName+" configuration. You likely are connecting to the wrong database.",
			"database blockchain mode", meta.BlockchainMode,
			"crawler "+executableName+" blockchain mode", blockchainMode)
		return false
	}

	if meta.SchemaVersion == nil {
		info("database schema version is not set")
		return false
	}

	// check if the database schema version matches the schema version of the executable
	if *meta.SchemaVersion != database.SchemaVersion {
		// The log message looks wrong, but is right ("executable schema version", database.SchemaVersion)
		info("Database is using a different schema version than executable. You may have to upgrade the database schema or use a different version of the executable.",
			"database schema version", *meta.SchemaVersion,
			"executable schema version", database.SchemaVersion)
		return false
	}

	return true
}
