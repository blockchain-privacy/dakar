// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package crawler

import (
	"backend/cmd/cliutil"
	"backend/db"
	dbstat "backend/db/status"
	"backend/external"
	"backend/jsonrpc"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "crawler"}, v...)...)
}

func warn(err error, v ...any) {
	serror.Log(slog.Default(), err, v...)
}

// holds the current state of the crawling processing loop
type crawlerState struct {
	// current block id
	id int64
	// top is the last seen highest block id
	top int64
	// current block hash
	hash string

	incremented bool
}

func (p *crawlerState) String() string {
	return fmt.Sprintf("ID: %d, Hash: %s", p.id, p.hash)
}

// increments the state for the next processing loop
func (p *crawlerState) increment(nextHash string) (err error) {
	p.incremented = false

	if nextHash == "" {
		return
	}

	p.hash = nextHash
	p.id++
	p.incremented = true

	return
}

// maps an address to one or more indexes of a transaction
type outputMapping struct {
	hash    string
	indexes []int32
}

// transactionMapping maps an address to one or more indexes of a transaction
type transactionMapping struct {
	hash    string
	outputs map[string]outputMapping
}

// adds indexOutput to an existing outputMapping in mapping. If none exists it inserts a new mapping
func addOutputToMapping(mapping map[string]outputMapping, addr string, indexOutput int32) map[string]outputMapping {
	if val, ok := mapping[addr]; ok {
		val.indexes = append(val.indexes, indexOutput)
		mapping[addr] = val
		return mapping
	}

	mapping[addr] = outputMapping{
		hash:    addr,
		indexes: []int32{indexOutput},
	}

	return mapping
}

// addOutputsToAddresses adds the given uids of outputs to the address specified by addr in addresses
// addr is inserted into addresses if it does not yet exist
func addOutputsToAddresses(addresses map[string]db.Address, addr string, uids []string) {
	var (
		editAddress db.Address
		ok          bool
	)

	if editAddress, ok = addresses[addr]; !ok {
		// new address -> set hash
		editAddress.Hash = addr
	}

	// add new outputs
	for _, uid := range uids {
		editAddress.Outputs = append(editAddress.Outputs, db.Output{UID: uid})
	}

	// save in map
	addresses[addr] = editAddress
}

func buildAddresses(mutex sync.Locker, cache *outputCache, txHash string, outputs map[string]outputMapping,
	addrMap map[string]db.Address) error {
	if cache == nil {
		return serror.FromStr("cache is not set")
	}

	if txHash == "" {
		return serror.FromStr("transaction hash is empty")
	}

	for _, mapping := range outputs {
		var uids []string
		for _, idx := range mapping.indexes {
			output := cache.getOutput(txHash, idx)

			if output == nil {
				return serror.FromFormat("requested output not found in cache: hash: %s index: %d", txHash, idx)
			}

			uids = append(uids, output.UID)
		}

		mutex.Lock()
		addOutputsToAddresses(addrMap, mapping.hash, uids)
		mutex.Unlock()
	}

	return nil
}

// processAddresses inserts mappings between addresses and outputs in database
func processAddresses(ctx context.Context, dgraph external.Database, cache *outputCache,
	transactionMappings []transactionMapping) error {
	if len(transactionMappings) == 0 {
		return nil
	}

	if cache == nil {
		return serror.FromStr("cache is not set")
	}

	addrMap := make(map[string]db.Address)
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var err error
	for _, mapping := range transactionMappings {
		wg.Add(1)
		go func(hash string, outputs map[string]outputMapping) {
			defer wg.Done()
			if err = buildAddresses(&mutex, cache, hash, outputs, addrMap); err != nil {
				return
			}
		}(mapping.hash, mapping.outputs)
	}

	wg.Wait()

	// check error from wait group
	if err != nil {
		return err
	}

	return db.UpsertAddresses(ctx, dgraph, cliutil.GetMapValues(addrMap))
}

// createOutputUID creates a named uid, parsable by dgraph
func createOutputUID(transaction string, outputID int32) string {
	return "_:" + transaction + strconv.FormatInt(int64(outputID), 10)
}

// newAmount multiplies the given float times 1e8 and returns an integer
func newAmount(f float64) (int64, error) {
	// The amount is only considered invalid if it cannot be represented
	// as an integer type. This may happen if f is NaN or +-Infinity.
	switch {
	case math.IsNaN(f):
		fallthrough
	case math.IsInf(f, 1):
		fallthrough
	case math.IsInf(f, -1):
		return 0, errors.New("invalid amount")
	}

	f *= 1e8

	if f < 0 {
		return int64(f - 0.5), nil
	}
	return int64(f + 0.5), nil
}

// getOutputAddress returns the address associated with the given output.
// If no address can be found, an empty address is returned.
func getOutputAddress(pubKey *jsonrpc.ScriptPubKeyResult, pubKeyHashAddrID byte) (string, error) {
	if pubKey == nil {
		return "", serror.FromStr("received nil ScriptPubKeyResult")
	}

	if pubKey.Address != "" {
		return pubKey.Address, nil
	}

	// try to extract addresses
	if pubKey.Type != "nulldata" && pubKey.Type != "nonstandard" {
		decodeString, err := hex.DecodeString(pubKey.Hex)
		if err != nil {
			return "", serror.New(err)
		}

		cfg := chaincfg.MainNetParams
		cfg.PubKeyHashAddrID = pubKeyHashAddrID
		_, addresses, _, err := txscript.ExtractPkScriptAddrs(decodeString, &cfg)
		if err != nil {
			return "", serror.New(err)
		}

		if len(addresses) > 0 {
			// use first address, ignore others
			return addresses[0].EncodeAddress(), nil
		}
	}

	return "", nil
}

// buildTransactionMapping processes given transaction.
// arguments:
// - rawTransaction: the transaction which is being processed
// - txHashMap: maps transaction hashes to transactions
// - externalOutputs: mapping between transaction hashes to mapping of indexes to transaction outputs
// returns:
// - txDetails: the created transaction
// - tMap: the transaction mapping between the transaction and its output, this needed for address processing
func buildTransactionMapping(rawTransaction jsonrpc.TxRawResult,
	txHashMap map[string]jsonrpc.TxRawResult, externalOutputs map[string]map[int32]db.Output,
	config Config, cache *outputCache) (txDetails db.Transaction, tMap transactionMapping, err error) {
	txDetails.Hash = rawTransaction.Txid

	var isCoinbaseTransaction bool
	if len(rawTransaction.Vin) == 1 && rawTransaction.Vin[0].IsCoinBase() {
		isCoinbaseTransaction = true
	} else {
		// process inputs if transaction is not a coinbase transaction
		i := int32(0)
		for _, d := range rawTransaction.Vin {
			if processErr := processTxVin(&txDetails, externalOutputs, d, i, txHashMap, cache); processErr != nil {
				err = processErr
				return
			}
			i++
		}
	}

	var foundAllInputs bool
	if !isCoinbaseTransaction {
		if len(rawTransaction.Vin) == len(txDetails.Inputs) {
			foundAllInputs = true
		} else {
			err = serror.FromFormat("not all inputs where found in transaction %s", rawTransaction.Txid)
			return
		}
	} else {
		// no fees for coinbase transactions
		txDetails.Fee = new(int64(0))
	}

	// process all outputs
	outputMappings := make(map[string]outputMapping)
	for _, d := range rawTransaction.Vout {
		intAmount, valErr := newAmount(d.Value)
		if valErr != nil {
			err = serror.New(valErr)
			return
		}
		index := d.N

		address, outputErr := getOutputAddress(&d.ScriptPubKey, config.PubKeyHashAddrID)
		if outputErr != nil {
			err = outputErr
			return
		}

		if address != "" {
			outputMappings = addOutputToMapping(outputMappings, address, index)
		}

		// create new output
		txDetails.Outputs = append(txDetails.Outputs, db.Output{
			UID:         createOutputUID(rawTransaction.Txid, index),
			IsCoinbase:  &isCoinbaseTransaction,
			Amount:      &intAmount,
			OutputIndex: &index,
		})
	}

	// if all inputs are available the transaction fee gets calculated
	if foundAllInputs {
		if err = txDetails.CalculateTransactionFee(); err != nil {
			err = serror.New(err)
			return
		}
	}

	// create transaction mapping for address processing later on
	tMap = transactionMapping{hash: txDetails.Hash, outputs: outputMappings}

	return
}

// filterExternalOutputs returns all inputs for which the outputs need to be loaded from the database
func filterExternalOutputs(txHashMap map[string]jsonrpc.TxRawResult, cache *outputCache) map[string][]int32 {
	externalOutputs := make(map[string][]int32)

	for _, t := range txHashMap {
		for _, vin := range t.Vin {
			if vin.IsCoinBase() {
				// coin base >>input<< does not hold any valuable information, therefore we do not include it in the database
				// we can recognize coinbase outputs by checking the number of connected transactions
				continue
			}

			if _, ok := txHashMap[vin.Txid]; !ok && cache.getOutput(vin.Txid, vin.Vout) == nil {
				externalOutputs[vin.Txid] = append(externalOutputs[vin.Txid], vin.Vout)
			}
		}
	}

	return externalOutputs
}

// processTxVin maps the input information to the output if it exists already in the database
func processTxVin(details *db.Transaction, externalOutputs map[string]map[int32]db.Output,
	vin jsonrpc.Vin, index int32, txHashMap map[string]jsonrpc.TxRawResult, cache *outputCache) error {
	if vin.IsCoinBase() {
		// coin base >>input<< does not hold any valuable information, therefore we do not include it in the database
		// we can recognize coinbase outputs by checking the number of connected transactions
		return nil
	}

	refOutput := db.Output{InputIndex: &index}

	if v, ok := txHashMap[vin.Txid]; ok {
		refOutput.UID = createOutputUID(vin.Txid, vin.Vout)
		intAmount, err := newAmount(v.Vout[vin.Vout].Value)
		if err != nil {
			return serror.New(err)
		}
		refOutput.Amount = &intAmount
	} else if o := cache.getAndEvictOutput(vin.Txid, vin.Vout); o != nil {
		refOutput.Amount = o.Amount
		refOutput.UID = o.UID
	} else {
		t, ok := externalOutputs[vin.Txid]
		if !ok {
			return serror.FromFormat("tx %s does not exist in external cache", vin.Txid)
		}

		o, ok := t[vin.Vout]
		if !ok {
			return serror.FromFormat("tx %s - outputindex %d does not exist in external cache", vin.Txid, vin.Vout)
		}

		refOutput.Amount = o.Amount
		refOutput.UID = o.UID
	}

	details.Inputs = append(details.Inputs, refOutput)
	return nil
}

var errBlockIDsDoNotMatch = errors.New("block id of last crawled block and highest found block do not match")

// getStartingID gets the block id from which the crawling will be resumed. If no crawling has
// happened yet, the block id is set to 1.
func getStartingID(ctx context.Context, dgraph external.Database) (startID int64, err error) {
	status, err := dbstat.GetCrawlerStatus(ctx, dgraph)
	if err != nil {
		return
	}

	if status.LastBlockID == nil {
		// last block id is not set -> we start at the beginning of the chain
		startID = 1
		return
	}

	highestBlockID, err := dbstat.GetHighestBlockID(ctx, dgraph)
	if err != nil {
		return
	}

	if *status.LastBlockID != highestBlockID {
		err = errBlockIDsDoNotMatch
	}

	startID = *status.LastBlockID

	return
}

func processingInterrupted() {
	info("Block processing interrupted")
}

// waitForNextRPCBlock waits for the next block. If the interrupt receives a signal isInterrupt is true.
// If the next block is available, currentBlock gets updated.
func waitForNextRPCBlock(client external.RPCClient, interrupt <-chan struct{}, hashObj string,
	rpcNumBlocks int64, config Config) (currentBlock *jsonrpc.GetBlockVerboseResult, isInterrupt bool, err error) {
	if hashObj == "" {
		err = serror.FromStr("blockhash is nil")
		return
	}

	ticker := time.Tick(config.NewBlockIntervalTime)
	for {
		select {
		case <-interrupt:
			processingInterrupted()
			isInterrupt = true
			return
		case <-ticker:
			currentBlock, err = client.GetBlockVerbose(hashObj)
			if err != nil {
				return
			}
		}

		numBlocks, rpcErr := getRPCNumberOfBlocks(client)
		if rpcErr != nil {
			err = rpcErr
			return
		}
		// check if block is available and if it is an actual new block
		if currentBlock.NextHash != "" && numBlocks > rpcNumBlocks {
			break
		}
	}

	return
}

// getRPCNumberOfBlocks returns the number of blocks currently in the chain of the RPC client
func getRPCNumberOfBlocks(client external.RPCClient) (int64, error) {
	blocksCount, err := client.GetBlockCount()
	if err != nil {
		return 0, err
	}

	if blocksCount < 0 {
		return 0, serror.FromStr("error RPC client block count is negative")
	}

	return blocksCount, nil
}

// getInitialState creates the initial state of the processing loop
func getInitialState(ctx context.Context, dgraph external.Database, client external.RPCClient) (state crawlerState, err error) {
	if state.id, err = getStartingID(ctx, dgraph); err != nil {
		if !errors.Is(err, errBlockIDsDoNotMatch) {
			return
		}
		warn(serror.FromFormat("%w, continuing ...", errBlockIDsDoNotMatch))
	}

	if state.hash, err = client.GetBlockHash(state.id); err != nil {
		return
	}

	// get RPC client block count
	numBlocks, err := getRPCNumberOfBlocks(client)
	if err != nil {
		return
	}

	state.top = numBlocks

	return
}

// getExternalOutputs returns a mapping between transaction hashes and a mapping of indexes to transaction outputs
func getExternalOutputs(ctx context.Context, dgraph external.Database,
	outputs map[string][]int32) (map[string]map[int32]db.Output, error) {
	if len(outputs) == 0 {
		return map[string]map[int32]db.Output{}, nil
	}

	transactionsOutputs, err := db.GetTransactionsOutputs(ctx, dgraph, cliutil.GetMapKeys(outputs))
	if err != nil {
		return nil, err
	}

	returnMap := make(map[string]map[int32]db.Output)

	for _, t := range transactionsOutputs {
		indexes := outputs[t.Hash]

		for _, i := range indexes {
			for _, o := range t.Outputs {
				if o.OutputIndex == nil {
					return nil, serror.FromFormat("output index was not set for tx %s", t.Hash)
				}
				if *o.OutputIndex == i {
					// add index mapping
					indexMap := returnMap[t.Hash]
					if indexMap == nil {
						indexMap = make(map[int32]db.Output)
					}

					indexMap[i] = o
					returnMap[t.Hash] = indexMap
				}
			}
		}
	}

	return returnMap, nil
}

// processRound process the given block. That includes the insertion of the block,
// its transaction, the outputs of all transaction and the mapping between outputs and addresses
func processRound(ctx context.Context, dgraph external.Database, state crawlerState,
	block *jsonrpc.GetBlockVerboseResult, txMap map[string]jsonrpc.TxRawResult, config Config, cache *outputCache) (
	blkCounter int64, txCounter int64, err error) {
	externalOutputs, err := getExternalOutputs(ctx, dgraph, filterExternalOutputs(txMap, cache))
	if err != nil {
		return 0, 0, err
	}

	var txMapping []transactionMapping
	transactions := make([]db.Transaction, 0, len(txMap))
	for _, t := range txMap {
		newTx, tMap, buildErr := buildTransactionMapping(t, txMap, externalOutputs, config, cache)
		if buildErr != nil {
			err = serror.AddContext(err, "state", state.String())
			return
		}

		txCounter++
		transactions = append(transactions, newTx)
		if tMap.hash != "" && len(tMap.outputs) > 0 {
			txMapping = append(txMapping, tMap)
		}
	}

	// sanity check for number of transactions
	if len(transactions) != len(block.Tx) {
		err = serror.FromFormat("wrong number of transactions in block: %s", block.Hash)
		return
	}

	// if the current block is not yet in the database or if only a shallow block exist in
	// the database a new block is created. Shallow blocks get created when a crawling process gets
	// started for the first time. Each block creation connects the current block with the previous block.
	// In the case of the first block, a previous block does not exist, thus a shallow block is created.
	// This check is relatively late in the processing loop. The reason for this is, that even if the
	// block already exists, the address mapping might not exist. This is the case if after block
	// creation the crawling process is aborted. So the address mapping must be created either way.
	// Address mappings are upserted in the worst case with same mappings as already included in the database,
	// so there is no damage done if we upsert the same mapping twice.
	var b db.Block
	if b, err = db.GetBlock(ctx, dgraph, state.hash); err != nil || !b.IsComplete() {
		// block is not yet in database -> create new block
		if err = db.UpsertBlock(ctx, dgraph, db.Block{
			Hash:      state.hash,
			Timestamp: time.Unix(block.Time, 0).Format(time.RFC3339),
			ID:        &state.id,
			PrevBlock: &db.Block{
				Hash: block.PreviousHash,
			},
			Transactions: transactions,
		}); err != nil {
			err = serror.AddContext(err, "state", state.String())
			return
		}

		blkCounter++
	} else {
		// reset txCounter as the block is not processed
		txCounter = 0
	}

	transactionOutputs, err := db.GetOutputs(ctx, dgraph, state.id, state.id)
	if err != nil {
		return
	}

	allOutputsCache := newOutputCache()
	for _, t := range transactionOutputs {
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
			// this cache only gets UTXOs
			if setErr := cache.setOutputs(t.Hash, utxos); setErr != nil {
				err = setErr
				return
			}
		}

		// this cache gets all outputs
		if setErr := allOutputsCache.setOutputs(t.Hash, t.Outputs); setErr != nil {
			err = setErr
			return
		}
	}

	if err = processAddresses(ctx, dgraph, allOutputsCache, txMapping); err != nil {
		err = serror.AddContext(err, "state", state.String())
		return
	}

	// save processing state
	if err = dbstat.SetLastBlockID(ctx, dgraph, state.id); err != nil {
		err = serror.AddContext(err, "state", state.String())
		return
	}

	return
}
