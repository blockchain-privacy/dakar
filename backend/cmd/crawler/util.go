package main

import (
	"backend/db/status"
	"backend/external"

	"context"
	"errors"
	"log"
	"net/http"
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

type ClusterModule struct {
	HMI bool `yaml:"hmi"`
	FMI bool `yaml:"fmi"`
}

type HTTPModule struct {
	Active bool `yaml:"active"`
	Port   uint `yaml:"port"`
}

type ModulesConfig struct {
	HTTP       HTTPModule    `yaml:"http"`
	Classifier bool          `yaml:"classifier"`
	Heuristics bool          `yaml:"heuristics"`
	Crawler    CrawlerModule `yaml:"crawler"`
	Clustering ClusterModule `yaml:"clustering"`
}

type Config struct {
	BlockchainMode string         `yaml:"blockchainMode"`
	Logfile        string         `yaml:"logfile"`
	RPC            RPCConfig      `yaml:"rpc"`
	Database       DatabaseConfig `yaml:"database"`
	Modules        ModulesConfig  `yaml:"modules"`
}

var defaultConfig = Config{
	BlockchainMode: "Dash",
	Logfile:        "",
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
		HTTP:       HTTPModule{Active: true, Port: 8081},
		Classifier: false,
		Heuristics: false,
		Crawler:    CrawlerModule{Active: true, InitialCacheSize: 25000},
		Clustering: ClusterModule{HMI: false, FMI: false},
	},
}

type Commands struct {
	ResetDB         bool
	IgnoreSafeGuard bool
}

// checks if a crawling process is already running
func isCrawling(db external.Database) (bool, error) {
	dbStatus, err := status.GetCrawlerStatus(db)
	if err != nil {
		// no status information found -> database is completely new
		// and thus no crawling is happening right now
		if errors.Is(err, status.ErrorStatusNotFound) {
			return false, nil
		}

		return true, err
	} else if dbStatus.IsCrawling == nil {
		return true, errors.New("was not able to get crawling status successfully")
	}

	return *dbStatus.IsCrawling, nil
}

// waitForRPCClient waits until the RPC client is ready to receive requests
func waitForRPCClient(client external.RPCClient) bool {
	const maxRetries = 5
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := 0; i < maxRetries; i++ {
		_, err := client.GetBlockCount()
		if err == nil {
			if printedErrMessage {
				info("Successfully established connection to RPC client.")
			}
			return true
		}

		if strings.Contains(err.Error(), "status code: 401") {
			info("Authentication error:", err)
			return false
		}

		if !printedErrMessage {
			info("Waiting for RPC client to start")
			printedErrMessage = true
		}

		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}
	info("RPC client is not ready to receive requests.")
	return false
}

// waitForBatchRPCClient waits until the batch RPC client is ready to receive requests
func waitForBatchRPCClient(client external.BatchRPCClient) bool {
	const maxRetries = 5
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := 0; i < maxRetries; i++ {
		result := client.GetBlockCountAsync()
		err := client.Send()
		if err != nil {
			log.Fatal(err)
		}
		_, err = result.Receive()
		if err == nil {
			if printedErrMessage {
				info("Successfully established connection to RPC client.")
			}
			return true
		}

		if strings.Contains(err.Error(), "status code: 401") {
			info("Authentication error:", err)
			return false
		}

		if !printedErrMessage {
			info("Waiting for RPC client to start")
			printedErrMessage = true
		}

		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}
	info("RPC client is not ready to receive requests.")
	return false
}

// waitForDatabase waits until the database is ready to receive requests
func waitForDatabase(db external.Database) bool {
	const maxRetries = 20
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := 0; i < maxRetries; i++ {
		if status.IsConnectionEstablished(db) {
			if printedErrMessage {
				info("Successfully established connection to database.")
			}
			return true
		}

		if !printedErrMessage {
			info("Waiting for database")
			printedErrMessage = true
		}

		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	info("Database is not ready to receive requests.")

	return false
}

// shutdownServer sends a shutdown signal to the server with a timout of 10 seconds
func shutdownServer(srv *http.Server) {
	if srv == nil {
		return
	}
	info("### Shutting down server###")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		info("Server was shutdown and returned error:", err)
	}
}
