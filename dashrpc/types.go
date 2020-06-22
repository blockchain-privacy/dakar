package dashrpc

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"math"
	"time"
)

// Database core data
const (
	ChainType_Block     = "blk" // Block hash -> Block
	ChainType_TxDetails = "txd" // TX hash -> TxDetails
	ChainType_AddrData  = "add" // add -> AddressData
	ChainType_Cluster   = "adc" // cluster ID -> ClusterData
)

// ChainItem represents a generic blockchain item
type ChainItem struct {
	ItemType string
	Data     []byte
}

type Transaction struct {
	Tx            TxDetails `json:"tx"`
	Bhash         string    `json:"bhash"`
	Bheight       uint64    `json:"bheight"`
	Bts           int64     `json:"bts"`
	Confirmations uint64    `json:"confirmations"`
	Version       int32     `json:"version"`
}

// Block represents a simple block
type Block struct {
	Hash          chainhash.Hash `json:"hash"`
	Id            uint64         `json:"id"`
	Timestamp     time.Time      `json:"ts"`
	PrevBlockHash chainhash.Hash `json:"prevblockhash"`
	NextBlockHash chainhash.Hash `json:"nextblockhash"`
	TxHashes      []string       `json:"txhashes"`
}

// BlkDetails represents transaction data
type BlkDetails struct {
	Hash          string    `json:"hash"`
	Id            uint64    `json:"id"`
	Timestamp     time.Time `json:"ts"`
	PrevBlockHash string    `json:"prevblockhash"`
	NextBlockHash string    `json:"nextblockhash"`
	TxHashes      []string  `json:"txhashes"`
}

func (blk BlkDetails) String() string {
	return fmt.Sprintf("hash: %s\nid:\n%timestamp:\n%v\n",
		blk.Hash, blk.Id, blk.Timestamp)
}

// TxOutput represents simple value transfer from/to address(es)
type TxOutput struct {
	TxHash string  `json:"txhash"`
	TxType string  `json:"txtype"`
	Amount float64 `json:"amount"`
	// For UTXO this will not be known yet
	Addresses []string `json:"addresses"`
	// For TXO this represents the index of the output in the transaction
	Index      int  `json:"index"`
	IsCoinbase bool `json:"iscoinbase"`
}

// ClusterData represents the cluster information
type ClusterData struct {
	Addresses []string `json:"addresses"` // addresses in this cluster
	Name      string   `json:"name"`      // this cluster name
	Heuristic string   `json:"heuristic"` // heuristic name and version used to generate this cluster
}

// AddressData represents the data associated with an address
type AddressData struct {
	Address  string     `json:"address"`  // the actual string address
	Clusters []string   `json:"clusters"` // Clusters with this address
	Txs      []TxOutput `json:"txs"`      // Tx with this address
}

func (tx TxOutput) String() string {
	return fmt.Sprintf("hash: %s\ntype: %s\namount: %f\nisCoinbase: %v\naddresses: %v",
		tx.TxHash, tx.TxType, tx.Amount, tx.IsCoinbase, tx.Addresses)
}

// TxDetails represents transaction data
type TxDetails struct {
	Hash      string     `json:"hash"`
	Inputs    []TxOutput `json:"inputs"`
	Outputs   []TxOutput `json:"outputs"`
	Timestamp int64      `json:"ts"`
}

func (tx TxDetails) String() string {
	return fmt.Sprintf("hash: %s\noutputs:\n%vinputs:\n%v\n",
		tx.Hash, tx.Outputs, tx.Inputs)
}

// IsCreateDenominations checks if the TX creates denominations
func (tx TxDetails) IsCreateDenominations() bool {
	denom := CountDenominations(tx.Outputs)
	return len(tx.Inputs) == 1 &&
		(denom[0] > 2 || denom[1] > 2 || denom[2] > 2)
}

func (tx TxDetails) IsPrivateSend() bool {
	denom := CountDenominations(tx.Inputs)
	return len(tx.Outputs) == 1 &&
		(denom[0] > 2 || denom[1] > 2 || denom[2] > 2)
}

// IsOneOrTwoOutputs checks if TX has only 1 or 2 outputs. Used for clustering.
func (tx TxDetails) IsOneOrTwoOutput() bool {
	return !tx.IsMixing() &&
		(len(tx.Outputs) == 2 || len(tx.Outputs) == 1)
}

// IsMixing checks if TX is mixing
func (tx TxDetails) IsMixing() bool {
	if len(tx.Inputs) != len(tx.Outputs) {
		return false
	}
	denomIn := CountDenominations(tx.Inputs)
	denomOut := CountDenominations(tx.Outputs)
	sum := 0
	for _, v := range denomIn {
		sum += v
	}
	if sum == 0 {
		return false
	}
	sum = 0
	for _, v := range denomIn {
		sum += v
	}
	if sum == 0 {
		return false
	}
	for i := range denomIn {
		if denomIn[i] != denomOut[i] {
			return false
		}
	}
	return true
}

func almostEqual(a, b float64) bool {
	var delta float64
	delta = 0.00001
	return math.Abs(a-b) <= delta
}

func CountDenominations(txs []TxOutput) []int {
	denominations := make([]int, 4)
	denominationsTypes := []float64{1.00001, 0.100001, 0.0100001, 0.00100001}

	for _, o := range txs {
	inner:
		for i, v := range denominationsTypes {
			if almostEqual(o.Amount, v) {
				denominations[i]++
				break inner
			}
		}
	}

	return denominations
}
