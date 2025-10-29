// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package external

import (
	"backend/jsonrpc"
	"context"
	"time"

	"github.com/dgraph-io/dgo/v250"
	"github.com/dgraph-io/dgo/v250/protos/api"
)

// This package defines interfaces for external dependencies,
// so they can be mocked in testing.

// RPCClient defines the methods which rpcclient of btcsuite implements.
// Descriptions are copied from there.
type RPCClient interface {
	// SetTimeout sets the request timeout of the rpc client
	SetTimeout(timeout time.Duration)
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
	// GenerateToAddress mines blocks immediately. Only available on regtest mode
	GenerateToAddress(numBlocks int, address string) ([]string, error)
}

// Database defines the methods which Dgraph of Dgo implements.
// Descriptions are copied from there.
type Database interface {
	// Mutate executes a query followed by one or more than one mutation.
	Mutate(ctx context.Context, req *api.Request) (*api.Response, error)
	// Query but allows a variable map to be used.
	// This can provide safety against injection attacks.
	Query(ctx context.Context, q string, vars map[string]string) (*api.Response, error)
	// NewTxn creates a new transaction.
	NewTxn() *dgo.Txn
	// Close shutdown down all the connections to the Dgraph Cluster.
	Close()
	// DropAll resets the database
	DropAll(ctx context.Context) error
	// DropData drops all data of the database
	DropData(ctx context.Context) error
	// DropPredicate drops the predicate of the specified namespace
	DropPredicate(ctx context.Context, predicate string) error
	// DropNamespace drops all data of the namespace
	DropNamespace(ctx context.Context, nsID uint64) error
	// SetSchema sets the schema of the specified namespace
	SetSchema(ctx context.Context, schema string) error
	// CreateNamespace creates a new namespace and returns its ID
	CreateNamespace(ctx context.Context) (uint64, error)
}
