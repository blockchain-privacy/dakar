package external

import (
	"context"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
)

// This package defines interfaces for external dependencies,
// so they can be mocked in testing.

// RPCClient defines the methods which rpcclient of btcsuite implements.
// Descriptions are copied from there.
type RPCClient interface {
	// GetBlockCount returns the number of blocks in the longest blockchain.
	GetBlockCount() (int64, error)
	// GetBlockVerbose returns a data structure from the server with information
	// about a block given its hash.
	GetBlockVerbose(blockHash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error)
	// GetBlockChainInfo returns information related to the processing state of
	// various chain-specific details such as the current difficulty from the tip
	// of the main chain.
	GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error)
	// GetBlockHash returns the hash of the block in the best blockchain at the
	// given height.
	GetBlockHash(blockHeight int64) (*chainhash.Hash, error)
	// GetRawTransactionVerbose returns information about a transaction given
	// its hash.
	GetRawTransactionVerbose(txHash *chainhash.Hash) (*btcjson.TxRawResult, error)
	// GetBlockChainInfoAsync returns an instance of a type that can be used to get
	// the result of the RPC at some future time by invoking the Receive function
	// on the returned instance.
	GetBlockChainInfoAsync() rpcclient.FutureGetBlockChainInfoResult
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
