package testhelper

import (
	"backend/external"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/integration/rpctest"
	"github.com/btcsuite/btcd/rpcclient"
	"google.golang.org/grpc"
	"log"
	"os"
	"testing"
)

const EnvCIFlag = "CI_ACTIVE"

type ContainerName string

const (
	ContainerNameStatus    = ContainerName("dgraph_status")
	ContainerNameUser      = ContainerName("dgraph_user")
	ContainerNameProcessor = ContainerName("dgraph_processor")
	ContainerNameAnalytics = ContainerName("dgraph_analytics")
	ContainerNameDB        = ContainerName("dgraph_db")
)

func IsCIActive() bool {
	return os.Getenv(EnvCIFlag) != ""
}

func SkipIfNotCI(t *testing.T) {
	if !IsCIActive() {
		t.Skip("skipping CI test")
	}
}

// RunDgraphTests connects to the given dgraph container and runs all tests
// packageDBHandle should be set to the global db interface handle of the package module.
func RunDgraphTests(m *testing.M, packageDBHandle *external.Database, containerName ContainerName) {
	if IsCIActive() {
		// create dgraph client
		graphDB, c, err := external.CreateClient(string(containerName) + ":9080")
		if err != nil {
			log.Panic(err)
			return
		}
		defer func(c *grpc.ClientConn) {
			err := c.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(c)

		if !external.WaitForDatabase(graphDB) {
			log.Panic("Could not connect to database", err)
			return
		}

		*packageDBHandle = graphDB
	}

	m.Run()
}

// RunDgraphTestsWithRPC runs all test with the given dgraph container and an in-memory RPC client.
// packageDBHandle should be set to the global db interface handle of the package module.
func RunDgraphTestsWithRPC(m *testing.M, packageDBHandle *external.Database, containerName ContainerName,
	client *rpcclient.Client, batchClient *rpcclient.Client) {
	if IsCIActive() {
		// create dgraph client
		graphDB, c, err := external.CreateClient(string(containerName) + ":9080")
		if err != nil {
			log.Panic(err)
			return
		}
		defer func(c *grpc.ClientConn) {
			err := c.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(c)

		if !external.WaitForDatabase(graphDB) {
			log.Panic("Could not connect to database", err)
			return
		}

		harness, err := rpctest.New(&chaincfg.SimNetParams, nil, []string{"--rejectnonstd"}, "")
		if err != nil {
			log.Panic("unable to create primary harness: ", err)
			return
		}

		defer func(harness *rpctest.Harness) {
			_ = harness.TearDown()
		}(harness)

		// Initialize the primary mining node with a chain of length 125,
		// providing 25 mature coinbases to allow spending from for testing
		// purposes.
		if err := harness.SetUp(true, 25); err != nil {
			log.Panic("unable to setup test chain: ", err)
		}

		*packageDBHandle = graphDB
		*client = *harness.Client
		*batchClient = *harness.BatchClient
	}

	m.Run()
}
