package testhelper

import (
	"backend/external"
	"context"
	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
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

type TestDB struct {
	DB            external.Database
	IsDirty       bool
	BlockFileName string
}

func (t *TestDB) Mutate(ctx context.Context, req *api.Request) (*api.Response, error) {
	t.IsDirty = true
	return t.DB.Mutate(ctx, req)
}

func (t *TestDB) Query(ctx context.Context, q string, vars map[string]string) (*api.Response, error) {
	return t.DB.Query(ctx, q, vars)
}

func (t *TestDB) Alter(ctx context.Context, op *api.Operation) error {
	t.IsDirty = true
	return t.DB.Alter(ctx, op)
}

// NewTxn creates a new transaction.
func (t *TestDB) NewTxn() *dgo.Txn {
	t.IsDirty = true
	return t.DB.NewTxn()
}

func IsCIActive() bool {
	return os.Getenv(EnvCIFlag) != ""
}

func SkipIfNotCI(t *testing.T) {
	if !IsCIActive() {
		t.SkipNow()
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
