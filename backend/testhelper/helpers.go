package testhelper

import (
	"backend/external"
	"backend/jsonrpc"
	"context"
	_ "embed"
	"github.com/dgraph-io/dgo/v250"
	"github.com/dgraph-io/dgo/v250/protos/api"
	"log"
	"os"
	"sync/atomic"
	"testing"
)

type ContainerName string

const (
	EnvDBTests               = "DB_TESTS"
	EnvRPCTests              = "RPC_TESTS"
	EnvDBHostname            = "DB_HOSTNAME"
	EnvRPCHostname           = "RPC_HOSTNAME"
	UseClassifierFile        = "classifier"
	UseBlockFile             = "block"
	UsePrivacyFile           = "privacy"
	UseBTCPrivacyFile        = "btc_privacy"
	ClassifierFileFirstBlock = 1557775
	ClassifierFileLastBlock  = 1557830
	BlockFileFirstBlock      = 60000
	BlockFileLastBlock       = 60020
)

// BlockFile contains Dash blocks from height 60000 to 60020.
// This file includes block, transaction, address and cluster data.
//
//go:embed blocks_60000_60020.json
var BlockFile []byte

// ClassifierFile contains Dash blocks from height 1557775 to 1557780.
// This file includes block, transaction, and address data.
//
//go:embed blocks_1557775_1557830.json
var ClassifierFile []byte

// PrivacyFile contains a small transaction graph created by traversing forward beginning with tx
// 452f795486980ef698fe652b56597eef3e7f6ad155cb0c9f1de21254d9bd9b0e
//
//go:embed privacy_transactions.json
var PrivacyFile []byte

// BTCPrivacyFile contains a bitcoin blocks between 573945 574040
//
//go:embed btc_privacy_transactions.json
var BTCPrivacyFile []byte

type TestDB struct {
	DB          external.Database
	IsDirty     atomic.Bool
	FileNameKey string
}

func (t *TestDB) Mutate(ctx context.Context, req *api.Request) (*api.Response, error) {
	t.IsDirty.Store(true)
	return t.DB.Mutate(ctx, req)
}

func (t *TestDB) Query(ctx context.Context, q string, vars map[string]string) (*api.Response, error) {
	return t.DB.Query(ctx, q, vars)
}

// NewTxn creates a new transaction.
func (t *TestDB) NewTxn() *dgo.Txn {
	t.IsDirty.Store(true)
	return t.DB.NewTxn()
}

// Close shutdown down all the connections to the Dgraph Cluster.
func (t *TestDB) Close() {
	t.DB.Close()
}

// DropAllNamespaces removes all data including schemas of all namespaces
func (t *TestDB) DropAllNamespaces(ctx context.Context) error {
	t.IsDirty.Store(true)
	return t.DB.DropAllNamespaces(ctx)
}

// DropPredicate dops the predicate of the specified namespace
func (t *TestDB) DropPredicate(ctx context.Context, nsName, predicate string) error {
	t.IsDirty.Store(true)
	return t.DB.DropPredicate(ctx, nsName, predicate)
}

// SetSchema sets the schema of the specified namespace
func (t *TestDB) SetSchema(ctx context.Context, nsName string, schema string) error {
	t.IsDirty.Store(true)
	return t.DB.SetSchema(ctx, nsName, schema)
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
func SkipIfNoDB(t testing.TB) {
	if !DoDBTests() {
		t.SkipNow()
	}
}

func SkipIfNoRPC(t testing.TB) {
	if !DoRPCTests() {
		t.SkipNow()
	}
}

// setupGraphDB connects the given handel to the database.
// The returned GRPC connection must be closed by the caller, after all tests are done.
func setupGraphDB(packageDBHandle *TestDB) {
	dbName, ok := GetDBName()
	if !ok {
		log.Fatal("environment variable " + EnvDBHostname + " is not set")
	}
	// create dgraph client
	graphDB, err := external.CreateClient(dbName + ":9080")
	if err != nil {
		log.Panic(err)
		return
	}

	if !external.WaitForDatabase(graphDB) {
		log.Panic("Could not connect to database", err)
		return
	}

	packageDBHandle.DB = graphDB
	packageDBHandle.IsDirty.Store(true)
}

func setupRPC(client *jsonrpc.BlockchainClient) {
	rpcHostname, ok := GetRPCName()
	if !ok {
		log.Panic("environment variable " + EnvRPCHostname + " is not set")
		return
	}

	rpcClient := jsonrpc.NewBlockchainClient(rpcHostname+":8131", "rpc1user", "1234pass", nil)

	*client = *rpcClient

	if err := setupRPCTest(client, 5); err != nil {
		log.Panic("Could not setup RPC test", err)
		return
	}
}

func setupRPCTest(client *jsonrpc.BlockchainClient, numBlocks int) error {
	// wallet might already exist -> ignore error
	_, _ = client.CreateWallet("testwallet")
	// wallet might already be loaded -> ignore error
	_, _ = client.LoadWallet("testwallet")

	generateToAddress, err := client.GetNewAddress()
	if err != nil {
		return err
	}

	_, err = client.GenerateToAddress(numBlocks, generateToAddress)
	return err
}

// RunDgraphTests connects to the given dgraph container and runs all tests
// packageDBHandle should be set to the global db interface handle of the package module.
func RunDgraphTests(m *testing.M, packageDBHandle *TestDB) {
	if DoDBTests() {
		setupGraphDB(packageDBHandle)
		defer packageDBHandle.Close()
	}

	m.Run()
}

// RunDgraphAndRPCTests runs all tests.
// packageDBHandle should be set to the global db interface handle of the package module.
func RunDgraphAndRPCTests(m *testing.M, packageDBHandle *TestDB, client *jsonrpc.BlockchainClient) {
	if DoDBTests() {
		setupGraphDB(packageDBHandle)
		defer packageDBHandle.Close()
	}

	if DoRPCTests() {
		setupRPC(client)
	}

	m.Run()
}

func GetPointer[number any](n number) *number {
	return &n
}
