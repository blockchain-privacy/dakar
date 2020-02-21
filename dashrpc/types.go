package dashrpc

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"math"
	"time"
)

const ChainType_Block = "blk" // Block hash -> Block
const ChainType_TxDetails = "txd" // TX hash -> TxDetails

const ChainType_AddrOutputs = "ado" // addr -> []TxOutputs
const ChainType_AddrCluster = "adc" // C_addr -> cluster ID
const ChainType_Cluster = "clu" // cluster ID -> []string (addresses)

const Prefix_AddrCluster = "C_"

// not used ATM const DB_BLOCK_COUNT = "DB_BLOCK_COUNT"

// ChainItem represents a generic blockchain item
type ChainItem struct {
	ItemType string
	Data     []byte
}

// Block represents a simple block
type Block struct {
	Hash          chainhash.Hash
	Id            uint64
	Timestamp     time.Time
	PrevBlockHash chainhash.Hash
	NextBlockHash chainhash.Hash
	TxHashes      []string
}

// TxOutput represents simple value transfer from/to address(es)
type TxOutput struct {
	TxHash string
	TxType string
	Amount float64
	// For UTXO this will not be known
	Addresses []string
	// For TXO this represents the index of the output in the transaction
	Index      int
	IsCoinbase bool
}

func (tx TxOutput) String() string {
	return fmt.Sprintf("hash: %s\ntype: %s\namount: %f\nisCoinbase: %v\naddresses: %v",
		tx.TxHash, tx.TxType, tx.Amount, tx.IsCoinbase, tx.Addresses)
}

// TxDetails represents transaction data
type TxDetails struct {
	Hash      string
	Inputs    []TxOutput
	Outputs   []TxOutput
	Timestamp int64
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
