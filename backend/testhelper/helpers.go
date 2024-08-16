package testhelper

import (
	"backend/external"
	"context"
	_ "embed"
	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"google.golang.org/grpc"
	"log"
	"os"
	"testing"
)

type ContainerName string

const (
	EnvDBTests              = "DB_TESTS"
	EnvRPCTests             = "RPC_TESTS"
	EnvDBHostname           = "DB_HOSTNAME"
	EnvRPCHostname          = "RPC_HOSTNAME"
	UseClassifierFile       = "classifier"
	UseBlockFile            = "block"
	UsePrivacyFile          = "privacy"
	ClassifierFileLastBlock = 1557780
	BlockFileFirstBlock     = 60000
	BlockFileLastBlock      = 60020
)

// BlockFile contains Dash blocks from height 60000 to 60020.
// This file includes block, transaction, address and cluster data.
//
//go:embed blocks_60000_60020.json
var BlockFile []byte

// ClassifierFile contains Dash blocks from height 1557775 to 1557780.
// This file includes block, transaction, and address data.
//
//go:embed blocks_1557775_1557780.json
var ClassifierFile []byte

// PrivacyFile contains a small transaction graph created by traversing forward beginning with tx
// 452f795486980ef698fe652b56597eef3e7f6ad155cb0c9f1de21254d9bd9b0e
//
//go:embed privacy_transactions.json
var PrivacyFile []byte

type TestDB struct {
	DB          external.Database
	IsDirty     bool
	FileNameKey string
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

func DoDBTests() bool {
	_, ok := os.LookupEnv(EnvDBTests)
	return ok
}

func DoRPCTests() bool {
	_, ok := os.LookupEnv(EnvRPCTests)
	return ok
}

func GetDBName() (string, bool) {
	return os.LookupEnv(EnvDBHostname)
}

func GetRPCName() (string, bool) {
	return os.LookupEnv(EnvRPCHostname)
}
func SkipIfNoDB(t *testing.T) {
	if !DoDBTests() {
		t.SkipNow()
	}
}

func SkipIfNoRPC(t *testing.T) {
	if !DoRPCTests() {
		t.SkipNow()
	}
}

// RunDgraphTests connects to the given dgraph container and runs all tests
// packageDBHandle should be set to the global db interface handle of the package module.
func RunDgraphTests(m *testing.M, packageDBHandle *external.Database) {
	if DoDBTests() {
		dbName, ok := GetDBName()
		if !ok {
			log.Fatal("environment variable " + EnvDBHostname + " is not set")
		}
		// create dgraph client
		graphDB, c, err := external.CreateClient(dbName + ":9080")
		if err != nil {
			log.Panic(err)
			return
		}
		defer func(c *grpc.ClientConn) {
			err := c.Close()
			if err != nil {
				log.Panic(err)
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

func GetPointer[number any](n number) *number {
	return &n
}
