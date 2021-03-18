package external

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

type DashWrapper struct {
	rpc *rpcclient.Client
}

func NewDashWrapper(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (DashWrapper, error) {
	newClient, err := rpcclient.New(config, ntfnHandlers)
	if err != nil {
		return DashWrapper{}, err
	}
	return DashWrapper{rpc: newClient}, nil
}

// convertGetBlockVerboseResult converts from its corresponding the library type to the wrapper type
func convertGetBlockVerboseResult(result *btcjson.GetBlockVerboseResult) *GetBlockVerboseResult {
	newResult := GetBlockVerboseResult{
		Hash:          result.Hash,
		Confirmations: result.Confirmations,
		StrippedSize:  result.StrippedSize,
		Size:          result.Size,
		Weight:        result.Weight,
		Height:        result.Height,
		Version:       result.Version,
		VersionHex:    result.VersionHex,
		MerkleRoot:    result.MerkleRoot,
		Tx:            result.Tx,
		Time:          result.Time,
		Nonce:         result.Nonce,
		Bits:          result.Bits,
		Difficulty:    result.Difficulty,
		PreviousHash:  result.PreviousHash,
		NextHash:      result.NextHash,
	}

	for _, tx := range result.RawTx {
		newResult.RawTx = append(newResult.RawTx, convertTxRawResult(tx))
	}

	return &newResult
}

// convertTxRawResult converts from its corresponding the library type to the wrapper type
func convertTxRawResult(result btcjson.TxRawResult) TxRawResult {
	newTx := TxRawResult{
		Hex:           result.Hex,
		Txid:          result.Txid,
		Hash:          result.Hash,
		Size:          result.Size,
		Vsize:         result.Vsize,
		Weight:        result.Weight,
		Version:       result.Version,
		LockTime:      result.LockTime,
		BlockHash:     result.BlockHash,
		Confirmations: result.Confirmations,
		Time:          result.Time,
		Blocktime:     result.Blocktime,
	}

	if len(result.Vin) > 0 {
		for _, in := range result.Vin {

			var sig *ScriptSig = nil

			if in.ScriptSig != nil {
				sigTemp := ScriptSig(*in.ScriptSig)
				sig = &sigTemp
			}

			newVin := Vin{
				Coinbase:  in.Coinbase,
				Txid:      in.Txid,
				Vout:      in.Vout,
				ScriptSig: sig,
				Sequence:  in.Sequence,
				Witness:   in.Witness,
			}

			newTx.Vin = append(newTx.Vin, newVin)
		}
	}

	if len(result.Vout) > 0 {
		for _, out := range result.Vout {
			newVout := Vout{
				Value:        out.Value,
				N:            out.N,
				ScriptPubKey: ScriptPubKeyResult(out.ScriptPubKey),
			}

			newTx.Vout = append(newTx.Vout, newVout)
		}
	}

	return newTx
}

func convertGetBlockChainInfoResult(info *btcjson.GetBlockChainInfoResult) *GetBlockChainInfoResult {
	return &GetBlockChainInfoResult{
		Chain:                info.Chain,
		Blocks:               info.Blocks,
		Headers:              info.Headers,
		BestBlockHash:        info.BestBlockHash,
		Difficulty:           info.Difficulty,
		MedianTime:           info.MedianTime,
		VerificationProgress: info.VerificationProgress,
		Pruned:               info.Pruned,
		PruneHeight:          info.PruneHeight,
		ChainWork:            info.ChainWork,
	}
}

func (d DashWrapper) GetBlockCount() (int64, error) {
	return d.rpc.GetBlockCount()
}

func (d DashWrapper) GetBlockVerbose(blockHash *Hash) (*GetBlockVerboseResult, error) {
	hash := chainhash.Hash(*blockHash)
	result, err := d.rpc.GetBlockVerbose(&hash)
	if err != nil {
		return &GetBlockVerboseResult{}, err
	}

	return convertGetBlockVerboseResult(result), nil
}

func (d DashWrapper) GetBlockChainInfo() (*GetBlockChainInfoResult, error) {
	info, err := d.rpc.GetBlockChainInfo()
	if err != nil {
		return &GetBlockChainInfoResult{}, err
	}

	return convertGetBlockChainInfoResult(info), nil
}

func (d DashWrapper) GetBlockHash(blockHeight int64) (*Hash, error) {
	hash, err := d.rpc.GetBlockHash(blockHeight)
	if err != nil {
		return &Hash{}, err
	}
	newHash := Hash(*hash)
	return &newHash, nil
}

func (d DashWrapper) GetRawTransactionVerbose(txHash *Hash) (*TxRawResult, error) {
	hash := chainhash.Hash(*txHash)
	tx, err := d.rpc.GetRawTransactionVerbose(&hash)
	if err != nil {
		return &TxRawResult{}, err
	}
	convertedResult := convertTxRawResult(*tx)
	return &convertedResult, nil
}

func (d DashWrapper) GetBlockChainInfoAsync() FutureGetBlockChainInfoResult {
	info := d.rpc.GetBlockChainInfoAsync()

	return FutureGetBlockChainInfoResult{
		dash: &info,
	}
}
