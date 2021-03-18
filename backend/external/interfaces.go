package external

import (
	"errors"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	utilrpc "github.com/test/test/rpcclient"
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
	GetBlockVerbose(blockHash *Hash) (*GetBlockVerboseResult, error)
	// GetBlockChainInfo returns information related to the processing state of
	// various chain-specific details such as the current difficulty from the tip
	// of the main chain.
	GetBlockChainInfo() (*GetBlockChainInfoResult, error)
	// GetBlockHash returns the hash of the block in the best block chain at the
	// given height.
	GetBlockHash(blockHeight int64) (*Hash, error)
	// GetRawTransactionVerbose returns information about a transaction given
	// its hash.
	GetRawTransactionVerbose(txHash *Hash) (*TxRawResult, error)
	// GetBlockChainInfoAsync returns an instance of a type that can be used to get
	// the result of the RPC at some future time by invoking the Receive function
	// on the returned instance.
	GetBlockChainInfoAsync() FutureGetBlockChainInfoResult
}

const HashSize = 32

type Hash [HashSize]byte

func (h Hash) String() string {
	chainHash := chainhash.Hash(h)
	return chainHash.String()
}

type ScriptPubKeyResult struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex,omitempty"`
	ReqSigs   int32    `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

type Vout struct {
	Value        float64            `json:"value"`
	N            uint32             `json:"n"`
	ScriptPubKey ScriptPubKeyResult `json:"scriptPubKey"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type Vin struct {
	Coinbase  string     `json:"coinbase"`
	Txid      string     `json:"txid"`
	Vout      uint32     `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig"`
	Sequence  uint32     `json:"sequence"`
	Witness   []string   `json:"txinwitness"`
}

// IsCoinBase returns a bool to show if a Vin is a Coinbase one or not.
// Copied from btcutil
func (v *Vin) IsCoinBase() bool {
	return len(v.Coinbase) > 0
}

type TxRawResult struct {
	Hex           string `json:"hex"`
	Txid          string `json:"txid"`
	Hash          string `json:"hash,omitempty"`
	Size          int32  `json:"size,omitempty"`
	Vsize         int32  `json:"vsize,omitempty"`
	Weight        int32  `json:"weight,omitempty"`
	Version       int16  `json:"version"`
	LockTime      uint32 `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash,omitempty"`
	Confirmations uint64 `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	Blocktime     int64  `json:"blocktime,omitempty"`
}

type GetBlockVerboseResult struct {
	Hash          string        `json:"hash"`
	Confirmations int64         `json:"confirmations"`
	StrippedSize  int32         `json:"strippedsize"`
	Size          int32         `json:"size"`
	Weight        int32         `json:"weight"`
	Height        int64         `json:"height"`
	Version       int32         `json:"version"`
	VersionHex    string        `json:"versionHex"`
	MerkleRoot    string        `json:"merkleroot"`
	Tx            []string      `json:"tx,omitempty"`
	RawTx         []TxRawResult `json:"rawtx,omitempty"` // Note: this field is always empty when verbose != 2.
	Time          int64         `json:"time"`
	Nonce         uint32        `json:"nonce"`
	Bits          string        `json:"bits"`
	Difficulty    float64       `json:"difficulty"`
	PreviousHash  string        `json:"previousblockhash"`
	NextHash      string        `json:"nextblockhash,omitempty"`
}

type FutureGetBlockChainInfoResult struct {
	dash *rpcclient.FutureGetBlockChainInfoResult
	util *utilrpc.FutureGetBlockChainInfoResult
}

func (r FutureGetBlockChainInfoResult) Receive() (*GetBlockChainInfoResult, error) {
	if r.dash != nil {
		result, err := r.dash.Receive()
		if err != nil {
			return nil, err
		}
		return convertGetBlockChainInfoResult(result), nil
	}

	if r.util != nil {
		result, err := r.util.Receive()
		if err != nil {
			return nil, err
		}
		return convertGetBlockChainInfoResult(result), nil
	}
	return nil, errors.New("error no FutureGetBlockChainInfoResult available")
}

type GetBlockChainInfoResult struct {
	Chain                string  `json:"chain"`
	Blocks               int32   `json:"blocks"`
	Headers              int32   `json:"headers"`
	BestBlockHash        string  `json:"bestblockhash"`
	Difficulty           float64 `json:"difficulty"`
	MedianTime           int64   `json:"mediantime"`
	VerificationProgress float64 `json:"verificationprogress,omitempty"`
	Pruned               bool    `json:"pruned"`
	PruneHeight          int32   `json:"pruneheight,omitempty"`
	ChainWork            string  `json:"chainwork,omitempty"`
}

func NewHashFromStr(hash string) (*Hash, error) {
	newHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, err
	}

	hashBytes := Hash(*newHash)
	return &hashBytes, nil
}
