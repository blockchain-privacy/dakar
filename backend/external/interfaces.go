package external

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

// This package defines interfaces for external dependencies
// so they can be mocked in testing.

// RPCClient defines the methods which rpcclient of btcsuite implements.
// Descriptions are copied from there.
type RPCClient interface {
	// GetBlockCount returns the number of blocks in the longest block chain.
	GetBlockCount() (int64, error)
	// GetBlockVerbose returns a data structure from the server with information
	// about a block given its hash.
	GetBlockVerbose(blockHash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error)
	// GetBlockChainInfo returns information related to the processing state of
	// various chain-specific details such as the current difficulty from the tip
	// of the main chain.
	GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error)
	// GetBlockHash returns the hash of the block in the best block chain at the
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
