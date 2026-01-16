// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package crawler

import (
	"backend/db"
	"backend/external"
	"context"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

type outputCache struct {
	c map[string]map[int32]db.Output
}

// newUTXOCache loads the unspent transaction outputs from the last initialLoadSize blocks
func newUTXOCache(ctx context.Context, dgraph external.Database, mostRecentBlockID int64, initialLoadSize int64) (*outputCache, error) {
	if initialLoadSize == 0 {
		return &outputCache{c: make(map[string]map[int32]db.Output)}, nil
	}

	fromBlock := int64(1)
	if initialLoadSize <= mostRecentBlockID {
		fromBlock = mostRecentBlockID - (initialLoadSize - 1)
	}

	// load blocks in batches from db
	const steps = 100
	var transactions []db.Transaction
	stop := false
	for i := fromBlock; !stop; i += steps {
		to := i + steps - 1
		if to >= mostRecentBlockID {
			to = mostRecentBlockID
			stop = true
		}

		stepTransactions, err := db.GetOutputs(ctx, dgraph, i, to)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, stepTransactions...)
	}

	cache := outputCache{c: make(map[string]map[int32]db.Output)}

	for _, t := range transactions {
		if len(t.Outputs) == 0 {
			continue
		}
		var utxos []db.Output

		for _, o := range t.Outputs {
			if o.InputIndex == nil {
				utxos = append(utxos, o)
			}
		}

		if len(utxos) > 0 {
			if err := cache.setOutputs(t.Hash, utxos); err != nil {
				return nil, err
			}
		}
	}

	return &cache, nil
}

// newOutputCache returns an empty output cache
func newOutputCache() *outputCache {
	return &outputCache{c: make(map[string]map[int32]db.Output)}
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
func (u *outputCache) setOutputs(txHash string, outputs []db.Output) error {
	if len(outputs) == 0 {
		return serror.FromFormat("tried to set zero outputs for transaction %s", txHash)
	}

	if txHash == "" {
		return serror.FromStr("transaction hash is empty")
	}

	if _, ok := u.c[txHash]; ok {
		return nil
	}
	outputMap := make(map[int32]db.Output)
	for _, o := range outputs {
		if o.OutputIndex == nil {
			return serror.FromFormat("output index is not set for tx %s", txHash)
		}
		outputMap[*o.OutputIndex] = o
	}

	u.c[txHash] = outputMap
	return nil
}

// getOutput returns specified output
func (u *outputCache) getOutput(txHash string, outputIndex int32) *db.Output {
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

// deleteOutput returns the output specified output and deletes it afterward
func (u *outputCache) getAndEvictOutput(txHash string, outputIndex int32) *db.Output {
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
