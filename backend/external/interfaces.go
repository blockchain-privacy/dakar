package external

import (
	"backend/jsonrpc"
	"context"
	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
)

// This package defines interfaces for external dependencies,
// so they can be mocked in testing.

// RPCClient defines the methods which rpcclient of btcsuite implements.
// Descriptions are copied from there.
type RPCClient interface {
	// GetBlockCount returns the number of blocks in the longest blockchain.
	GetBlockCount() (int64, error)
	// GetBlockVerbose returns a data structure from the server with information about a block given its hash.
	GetBlockVerbose(blockHash string) (*jsonrpc.GetBlockVerboseResult, error)
	// GetBlockHash returns the hash of the block in the best blockchain at the given height.
	GetBlockHash(blockHeight int64) (string, error)
	// GetRawTransactionVerbose returns information about a transaction given its hash.
	GetRawTransactionVerbose(txHash string) (*jsonrpc.TxRawResult, error)
	// GetRawTransactionVerboseBatch returns several GetRawTransactionVerbose in one batch
	GetRawTransactionVerboseBatch(txs []string) ([]*jsonrpc.TxRawResult, error)
	// Generate mines blocks immediately. This is an non-standard RPC used for testing with BTCD
	Generate(numBlocks int) ([]string, error)
}

// Database defines the methods which Dgraph of Dgo implements.
// Descriptions are copied from there.
type Database interface {
	// Mutate executes a query followed by one or more than one mutation.
	Mutate(ctx context.Context, req *api.Request) (*api.Response, error)
	// Query but allows a variable map to be used.
	// This can provide safety against injection attacks.
	Query(ctx context.Context, q string, vars map[string]string) (*api.Response, error)
	// Alter can be used to do the following by setting various fields of api.Operation:
	//   1. Modify the schema.
	//   2. Drop a predicate.
	//   3. Drop the database.
	Alter(ctx context.Context, op *api.Operation) error
	// NewTxn creates a new transaction.
	NewTxn() *dgo.Txn
}
