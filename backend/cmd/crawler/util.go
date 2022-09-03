package main

import (
	database "backend/db"
	"backend/db/status"
	"backend/external"
	"backend/password"
	"fmt"
	ory "github.com/ory/kratos-client-go"
	"net/http/cookiejar"
	"runtime"
	"runtime/debug"

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
	Active               bool   `yaml:"active"`
	Port                 uint   `yaml:"port"`
	BasicAuthUser        string `yaml:"basicAuthUser"`
	BasicAuthPWHash      string `yaml:"basicAuthPWHash"`
	KratosPublicEndpoint string `yaml:"kratosPublicEndpoint"`
	KratosAdminEndpoint  string `yaml:"kratosAdminEndpoint"`
}

type ModulesConfig struct {
	HTTP       HTTPModule    `yaml:"http"`
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
		HTTP: HTTPModule{
			Active:               true,
			Port:                 8081,
			BasicAuthUser:        "dakar",
			BasicAuthPWHash:      "",
			KratosPublicEndpoint: "http://localhost:4433",
			KratosAdminEndpoint:  "http://localhost:4434",
		},
		Classifier: false,
		Heuristics: false,
		Crawler:    CrawlerModule{Active: true, InitialCacheSize: 25000},
		Clustering: ClusterModule{HMI: false, FMI: false},
	},
}

func getDefaultConfig() (Config, error) {
	passwd, pwErr := password.GenerateRandomPassword()
	if pwErr != nil {
		return Config{}, pwErr
	}

	pwHash, pwErr := password.GeneratePasswordHash(password.DefaultPasswordConfig, passwd)
	if pwErr != nil {
		return Config{}, pwErr
	}

	defaultConfig.Modules.HTTP.BasicAuthPWHash = pwHash

	fmt.Println("Generated new basic auth pair:\nuser: dakar", "\npassword:", passwd)
	fmt.Println("Save the password, it will not be written in the config file.")

	return defaultConfig, nil
}

// checkHTTPModuleConfig returns an error if the given http module has invalid values
func checkHTTPModuleConfig(c HTTPModule) error {
	if c.BasicAuthUser == "" || c.BasicAuthPWHash == "" ||
		c.KratosPublicEndpoint == "" || c.KratosAdminEndpoint == "" {
		return errors.New("http module config invalid, not all fields are filled")
	}

	parts := strings.Split(c.BasicAuthPWHash, "$")
	if len(parts) != 6 {
		return errors.New("basic auth password hash is invalid")
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
		info(err)
		return false
	}

	// check if the blockchain mode of database matches the blockchain mode of the executable
	if meta.BlockchainMode != blockchainMode {
		info("Database is using a different blockchain mode than the executable")
		info("Database blockchain mode:", meta.BlockchainMode)
		info("Executable blockchain mode:", blockchainMode)
		info("You likely used the wrong executable or connected to the wrong database")
		return false
	}

	if meta.SchemaVersion == nil {
		info("database schema version is not set")
		return false
	}

	// check if the database schema version matches the schema version of the executable
	if *meta.SchemaVersion != database.SchemaVersion {
		info("Database is using a different schema version than executable")
		info("Database schema version:", *meta.SchemaVersion)
		info("Executable schema version:", database.SchemaVersion)
		info("You may have to upgrade the database schema or use a different version of the executable")
		return false
	}

	return true
}

// newKratosClient creates a new kratos client
func newKratosClient(endpoint string) (*ory.APIClient, error) {
	cj, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: endpoint}}

	conf.HTTPClient = &http.Client{Jar: cj}

	return ory.NewAPIClient(conf), nil
}

// getKratosClient returns a public (first) and admin (second) handle to an ory kratos instance
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
	ctx1, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	_, resp1, err := auth.MetadataApi.IsAlive(ctx1).Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("kratos public endpoint is not alive: %w", err)
	}
	defer resp1.Body.Close()

	// check if admin endpoint is alive
	ctx2, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	_, resp2, err := adminAuth.MetadataApi.IsAlive(ctx2).Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("kratos admin endpoint is not alive: %w", err)
	}
	defer resp2.Body.Close()

	return auth, adminAuth, nil
}
