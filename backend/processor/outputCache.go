package processor

import (
	"backend/cmd/cliutil"
	dbop "backend/db/output"
	dbtx "backend/db/transaction"
	"backend/external"
	"errors"
	"fmt"
)

const initialLoadSize = 10

type outputCache struct {
	c map[string]map[uint32]dbop.Output
}

// newUTXOCache loads the unspent transaction outputs from the last initialLoadSize blocks
func newUTXOCache(dgraph external.Database, mostRecentBlockID int64) (*outputCache, error) {
	fromBlock := mostRecentBlockID - initialLoadSize
	if fromBlock <= 0 {
		fromBlock = 1
	}

	transactions, err := dbtx.GetOutputs(dgraph, fromBlock, mostRecentBlockID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	cache := outputCache{c: make(map[string]map[uint32]dbop.Output)}

	for _, t := range transactions {
		if len(t.Outputs) == 0 {
			continue
		}
		var utxos []dbop.Output

		for _, o := range t.Outputs {
			if o.InputIndex == nil {
				utxos = append(utxos, o)
			}
		}

		if len(utxos) > 0 {
			if err := cache.setOutputs(t.Hash, utxos); err != nil {
				return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}
	}

	return &cache, nil
}

// newOutputCache returns an empty output cache
func newOutputCache() *outputCache {
	return &outputCache{c: make(map[string]map[uint32]dbop.Output)}
}

// getOutputCounts returns the number of outputs in the cache
func (u *outputCache) getOutputCounts() int {
	var numOutputs int

	for _, v := range u.c {
		numOutputs += len(v)
	}

	return numOutputs
}

// setOutputs sets the outputs for the specified transaction hash.
func (u *outputCache) setOutputs(txHash string, outputs []dbop.Output) error {
	if len(outputs) == 0 {
		return fmt.Errorf("tried to set zero outputs for transaction %s", txHash)
	}

	if txHash == "" {
		return errors.New("transaction hash is empty")
	}

	if _, ok := u.c[txHash]; ok {
		return nil
		//return fmt.Errorf("transaction %s does already exist in cache", txHash)
	}
	outputMap := make(map[uint32]dbop.Output)
	for _, o := range outputs {
		if o.OutputIndex == nil {
			return fmt.Errorf("output index is not set for tx %s", txHash)
		}
		outputMap[*o.OutputIndex] = o
	}

	u.c[txHash] = outputMap
	return nil
}

// getOutput returns specified output
func (u *outputCache) getOutput(txHash string, outputIndex uint32) *dbop.Output {
	t, ok := u.c[txHash]
	if !ok {
		return nil
	}
	output, ok := t[outputIndex]
	if !ok {
		return nil
	}
	return &output
}

// deleteOutput deletes the specified output
func (u *outputCache) deleteOutput(txHash string, outputIndex uint32) {
	t, ok := u.c[txHash]
	if !ok {
		return
	}

	delete(t, outputIndex)

	// if transaction has no more unspent outputs, then remove the transaction reference
	if len(t) == 0 {
		delete(u.c, txHash)
	}
}

// deleteOutput returns the output specified output and deletes it afterwards
func (u *outputCache) getAndEvictOutput(txHash string, outputIndex uint32) *dbop.Output {
	t, ok := u.c[txHash]
	if !ok {
		return nil
	}
	output, ok := t[outputIndex]
	if !ok {
		return nil
	}

	delete(t, outputIndex)

	// if transaction has no more unspent outputs, then remove the transaction reference
	if len(t) == 0 {
		delete(u.c, txHash)
	}

	return &output
}
