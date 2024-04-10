package main

import (
	"backend/cmd/cliutil"
	database "backend/db"
	"backend/db/status"
	"backend/external"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	ory "github.com/ory/kratos-client-go"
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

type APIModule struct {
	Active               bool   `yaml:"active"`
	Port                 uint   `yaml:"port"`
	KratosPublicEndpoint string `yaml:"kratosPublicEndpoint"`
	KratosAdminEndpoint  string `yaml:"kratosAdminEndpoint"`
}

type MetricsModule struct {
	Active bool `yaml:"active"`
	Port   uint `yaml:"port"`
}

type ModulesConfig struct {
	HTTP       APIModule     `yaml:"api"`
	Metrics    MetricsModule `yaml:"metrics"`
	Crawler    CrawlerModule `yaml:"crawler"`
	Clustering ClusterModule `yaml:"clustering"`
	Classifier bool          `yaml:"classifier"`
	Heuristics bool          `yaml:"heuristics"`
}

type Config struct {
	Logfile  string         `yaml:"logfile"`
	RPC      RPCConfig      `yaml:"rpc"`
	Database DatabaseConfig `yaml:"database"`
	Modules  ModulesConfig  `yaml:"modules"`
}

var defaultConfig = Config{
	Logfile: "dakar.log",
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
			Active:               true,
			Port:                 8081,
			KratosPublicEndpoint: "http://localhost:4433",
			KratosAdminEndpoint:  "http://localhost:4434",
		},
		Metrics: MetricsModule{
			Active: true,
			Port:   8481,
		},
		Classifier: false,
		Heuristics: false,
		Crawler:    CrawlerModule{Active: true, InitialCacheSize: 25000},
		Clustering: ClusterModule{HMI: false, FMI: false},
	},
}

// checkAPIModuleConfig returns an error if the given http module has invalid values
func checkAPIModuleConfig(c APIModule) error {
	if c.KratosPublicEndpoint == "" || c.KratosAdminEndpoint == "" {
		return cliutil.NewStackErrorStr("http module config invalid, not all fields are filled")
	}

	return nil
}

type Commands struct {
	ResetDB         bool
	IgnoreSafeGuard bool
	ShowVersion     bool
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
		return true, cliutil.NewStackErrorStr("was not able to get crawling status successfully")
	}

	return *dbStatus.IsCrawling, nil
}

// waitForRPCClient waits until the RPC client is ready to receive requests
func waitForRPCClient(client external.RPCClient) bool {
	const maxRetries = 5
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := range maxRetries {
		_, err := client.GetBlockCount()
		if err == nil {
			if printedErrMessage {
				info("Successfully established connection to RPC client.")
			}
			return true
		}

		if strings.Contains(err.Error(), "status code: 401") {
			warn(cliutil.NewStackErrorf("Authentication error: %w", err))
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
		warn(cliutil.NewStackErrorf("Server was shutdown and returned error: %w", err))
	}
}

// printVersion prints the version of the application and build information
func printVersion() {
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
func checkMeta(db external.Database) bool {
	meta, err := status.GetMeta(db)
	if err != nil {
		warn(err)
		return false
	}

	// check if the blockchain mode of database matches the blockchain mode of the executable
	if meta.BlockchainMode != blockchainMode {
		info("Database is using a different blockchain mode than the executable. You likely used the wrong executable or connected to the wrong database.",
			"database blockchain mode", meta.BlockchainMode,
			"executable blockchain mode", blockchainMode)
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

// newKratosClient creates a new kratos client
func newKratosClient(endpoint string) (*ory.APIClient, error) {
	if endpoint == "" {
		return nil, cliutil.NewStackErrorf("endpoint is invalid: %s", endpoint)
	}

	cj, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: endpoint}}

	conf.HTTPClient = &http.Client{Jar: cj}

	return ory.NewAPIClient(conf), nil
}

// isKratosAlive returns true if a successful connection to kratos has been established
func isKratosAlive(auth *ory.APIClient) bool {
	if auth == nil || auth.MetadataAPI == nil {
		return false
	}

	// check if endpoint is alive
	ctx1, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	_, resp, err := auth.MetadataAPI.IsAlive(ctx1).Execute()
	if resp != nil {
		if err := resp.Body.Close(); err != nil {
			warn(cliutil.NewStackError(err))
		}
	}
	return err == nil
}

// waitForKratos waits until kratos is ready to receive requests
func waitForKratos(auth *ory.APIClient) bool {
	const maxRetries = 20
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := range maxRetries {
		if isKratosAlive(auth) {
			if printedErrMessage {
				fmt.Println("Successfully established connection to kratos")
			}
			return true
		}

		if !printedErrMessage {
			fmt.Println("Waiting for kratos")
			printedErrMessage = true
		}

		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return false
}

// getKratosClient returns a public (first) and admin (second) handle to an ory kratos instance.
// Also checks if the connections are alive.
func getKratosClient(publicEndpoint string, adminEndpoint string) (*ory.APIClient, *ory.APIClient, error) {
	auth, err := newKratosClient(publicEndpoint)
	if err != nil {
		return nil, nil, err
	}

	adminAuth, err := newKratosClient(adminEndpoint)
	if err != nil {
		return nil, nil, err
	}

	// check if public endpoint is alive
	if !waitForKratos(auth) {
		return nil, nil, cliutil.NewStackErrorStr("kratos public endpoint is not ready to receive requests")
	}

	// check if public endpoint is alive
	if !waitForKratos(adminAuth) {
		return nil, nil, cliutil.NewStackErrorStr("kratos admin endpoint is not ready to receive requests")
	}

	return auth, adminAuth, nil
}
