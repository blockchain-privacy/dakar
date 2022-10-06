package processor

import (
	"backend/cmd/cliutil"
	"backend/db"
	dbstat "backend/db/status"
	"backend/external"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
)

const (
	// loggerPrefix is the prefix which is printed for each log message
	loggerPrefix = "\033[0;35mprocess\u001B[0m\t"
)

var thisLogger = log.New(log.Writer(), loggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	thisLogger = log.New(out, loggerPrefix, flag)
}

func info(v ...interface{}) {
	thisLogger.Println(v...)
}

// holds the current state of the crawling processing loop
type crawlerState struct {
	// current block id
	id uint64
	// top is the last seen highest block id
	top uint64
	// current block hash
	hash string
	// current block hash as a chainhash.Hash
	chainHash *chainhash.Hash

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

	p.chainHash, err = chainhash.NewHashFromStr(nextHash)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
	indexes []uint32
}

// transactionMapping maps an address to one or more indexes of a transaction
type transactionMapping struct {
	hash    string
	outputs map[string]outputMapping
}

// adds indexOutput to an existing outputMapping in mapping. If none exists it inserts a new mapping
func addOutputToMapping(mapping map[string]outputMapping, addr string, indexOutput uint32) map[string]outputMapping {
	if val, ok := mapping[addr]; ok {
		val.indexes = append(val.indexes, indexOutput)
		mapping[addr] = val
		return mapping
	}

	mapping[addr] = outputMapping{
		hash:    addr,
		indexes: []uint32{indexOutput},
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
		return errors.New("cache is not set")
	}

	for _, mapping := range outputs {
		var uids []string
		for _, idx := range mapping.indexes {
			output := cache.getOutput(txHash, idx)

			if output == nil {
				return errors.New("requested output not found in cache")
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
func processAddresses(dgraph external.Database, cache *outputCache,
	transactionMappings []transactionMapping) error {
	if len(transactionMappings) == 0 {
		return nil
	}

	if cache == nil {
		return errors.New("cache is not set")
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
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
				return
			}
		}(mapping.hash, mapping.outputs)
	}

	wg.Wait()

	// check error from wait group
	if err != nil {
		return err
	}

	// map to slice
	addrSlice := make([]db.Address, 0, len(addrMap))
	for _, a := range addrMap {
		addrSlice = append(addrSlice, a)
	}

	if err := db.UpsertAddresses(dgraph, addrSlice); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// createOutputUID creates a named uid, parsable by dgraph
func createOutputUID(transaction string, outputID uint32) string {
	return "_:" + transaction + strconv.FormatUint(uint64(outputID), 10)
}

// buildTransactionMapping processes the transaction specified by 'txHashString'
// 'txDetails' is the created transaction
// 'tMap' is the transaction mapping between the transaction and its output, this needed for address processing
func buildTransactionMapping(rawTransaction btcjson.TxRawResult,
	txHashMap map[string]btcjson.TxRawResult, externalOutputs map[string]map[uint32]db.Output,
	config Config, cache *outputCache) (txDetails db.Transaction, tMap transactionMapping, err error) {
	txDetails.Hash = rawTransaction.Txid

	var isCoinbaseTransaction bool
	if len(rawTransaction.Vin) == 1 && rawTransaction.Vin[0].IsCoinBase() {
		isCoinbaseTransaction = true
	} else {
		// process inputs if transaction is not a coinbase transaction
		for i, d := range rawTransaction.Vin {
			if processErr := processTxVin(&txDetails, externalOutputs, d, uint32(i), txHashMap, cache); processErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), processErr)
				return
			}
		}
	}

	var foundAllInputs bool
	if !isCoinbaseTransaction {
		if len(rawTransaction.Vin) == len(txDetails.Inputs) {
			foundAllInputs = true
		} else {
			err = fmt.Errorf("not all inputs where found in transaction %s", rawTransaction.Txid)
			return
		}
	} else {
		// no fees for coinbase transactions
		nullAmount := int64(0)
		txDetails.Fee = &nullAmount
	}

	// process all outputs
	outputMappings := make(map[string]outputMapping)
	for _, d := range rawTransaction.Vout {
		amt, valErr := btcutil.NewAmount(d.Value)
		intAmount := int64(amt)
		if valErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), valErr)
			return
		}
		index := d.N

		// check if addresses can be extracted
		if d.ScriptPubKey.Addresses == nil && d.ScriptPubKey.Type != "nulldata" && d.ScriptPubKey.Type != "nonstandard" {
			decodeString, decodingErr := hex.DecodeString(d.ScriptPubKey.Hex)
			if decodingErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), decodingErr)
				return
			}

			cfg := chaincfg.MainNetParams
			cfg.PubKeyHashAddrID = config.PubKeyHashAddrID
			_, addresses, _, extractionError := txscript.ExtractPkScriptAddrs(decodeString, &cfg)
			if extractionError != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), extractionError)
				return
			}

			for _, e := range addresses {
				d.ScriptPubKey.Addresses = append(d.ScriptPubKey.Addresses, e.EncodeAddress())
			}
		}

		for _, e := range d.ScriptPubKey.Addresses {
			outputMappings = addOutputToMapping(outputMappings, e, index)
		}

		// create new output
		txDetails.Outputs = append(txDetails.Outputs, db.Output{
			UID:         createOutputUID(rawTransaction.Txid, index),
			IsCoinbase:  &isCoinbaseTransaction,
			Amount:      &intAmount,
			TxType:      d.ScriptPubKey.Type,
			KeyAsm:      d.ScriptPubKey.Asm,
			KeyHex:      d.ScriptPubKey.Hex,
			OutputIndex: &index,
		})
	}

	// if all inputs are available the transaction fee gets calculated
	if foundAllInputs {
		if err = txDetails.CalculateTransactionFee(); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}

	// create transaction mapping for address processing later on
	tMap = transactionMapping{hash: txDetails.Hash, outputs: outputMappings}

	return
}

// filterExternalOutputs returns all inputs for which the outputs need to be loaded from the database
func filterExternalOutputs(txHashMap map[string]btcjson.TxRawResult, cache *outputCache) map[string][]uint32 {
	externalOutputs := make(map[string][]uint32)

	for _, t := range txHashMap {
		for _, vin := range t.Vin {
			if vin.IsCoinBase() {
				// coin base >>input<< does not hold any valuable information, therefore we do not include it in the database
				// we can recognize coinbase outputs by checking the number of connected transactions
				continue
			}

			if _, ok := txHashMap[vin.Txid]; !ok && cache.getOutput(vin.Txid, vin.Vout) == nil {
				ids := externalOutputs[vin.Txid]
				ids = append(ids, vin.Vout)
				externalOutputs[vin.Txid] = ids
			}
		}
	}

	return externalOutputs
}

// processTxVin maps the input information to the output if it exists already in the database
func processTxVin(details *db.Transaction, externalOutputs map[string]map[uint32]db.Output,
	vin btcjson.Vin, index uint32, txHashMap map[string]btcjson.TxRawResult, cache *outputCache) error {
	if vin.IsCoinBase() {
		// coin base >>input<< does not hold any valuable information, therefore we do not include it in the database
		// we can recognize coinbase outputs by checking the number of connected transactions
		return nil
	}

	refOutput := db.Output{
		InputIndex: &index,
		SigAsm:     vin.ScriptSig.Asm,
		SigHex:     vin.ScriptSig.Hex,
	}

	if v, ok := txHashMap[vin.Txid]; ok {
		refOutput.UID = createOutputUID(vin.Txid, vin.Vout)
		amt, err := btcutil.NewAmount(v.Vout[vin.Vout].Value)
		intAmount := int64(amt)
		if err != nil {
			return err
		}
		refOutput.Amount = &intAmount
	} else if o := cache.getAndEvictOutput(vin.Txid, vin.Vout); o != nil {
		refOutput.Amount = o.Amount
		refOutput.UID = o.UID
	} else {
		t, ok := externalOutputs[vin.Txid]
		if !ok {
			return fmt.Errorf("tx %s does not exist in external cache", vin.Txid)
		}

		o, ok := t[vin.Vout]
		if !ok {
			return fmt.Errorf("tx %s - outputindex %d does not exist in external cache", vin.Txid, vin.Vout)
		}

		refOutput.Amount = o.Amount
		refOutput.UID = o.UID
	}

	details.Inputs = append(details.Inputs, refOutput)
	return nil
}

// processBlock builds a block with the provided arguments and inserts it in the database
func processBlock(dgraph external.Database, transactions []db.Transaction, currentHash string,
	blockID uint64, timestamp string, prevBlockHash string) (err error) {
	if err = db.UpsertBlock(dgraph, db.Block{
		Hash:      currentHash,
		Timestamp: timestamp,
		ID:        &blockID,
		PrevBlock: &db.Block{
			Hash: prevBlockHash,
		},
		Transactions: transactions,
	}); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	return
}

var errBlockIdsDoNotMatch = errors.New("block id of last crawled block and highest found block do not match")

// getStartingID gets the block id from which the crawling will be resumed. If no crawling has
// happened yet, the block id is set to 1.
func getStartingID(dgraph external.Database) (startID uint64, err error) {
	status, err := dbstat.GetCrawlerStatus(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if status.LastBlockID == nil {
		// last block id is not set -> we start at the beginning of the chain
		startID = 1
		return
	}

	highestBlockID, err := dbstat.GetHighestBlockID(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if *status.LastBlockID != highestBlockID {
		err = errBlockIdsDoNotMatch
	}

	startID = *status.LastBlockID

	return
}

func processingInterrupted() {
	info("### Block processing interrupted ###")
}

// waitForNextRPCBlock waits for the next block. If the interrupt receives a signal isInterrupt is true.
// If the next block is available, currentBlock gets updated.
func waitForNextRPCBlock(client external.RPCClient, interrupt <-chan struct{}, hashObj *chainhash.Hash,
	rpcNumBlocks uint64, config Config) (currentBlock *btcjson.GetBlockVerboseResult, isInterrupt bool, err error) {
	if hashObj == nil {
		err = errors.New("blockhash is nil")
		return
	}

	ticker := time.NewTicker(config.NewBlockIntervalTime)
	defer ticker.Stop()
	for {
		select {
		case <-interrupt:
			processingInterrupted()
			isInterrupt = true
			return
		case <-ticker.C:
			currentBlock, err = client.GetBlockVerbose(hashObj)
			if err != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
				return
			}
		}

		numBlocks, rpcErr := getRPCNumberOfBlocks(client)
		if rpcErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), rpcErr)
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
func getRPCNumberOfBlocks(client external.RPCClient) (uint64, error) {
	rpcInfo, err := client.GetBlockChainInfo()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if rpcInfo.Blocks < 0 {
		return 0, errors.New("error RPC client block count is negative")
	}

	return uint64(rpcInfo.Blocks), nil
}

// getInitialState creates the initial state of the processing loop
func getInitialState(dgraph external.Database, client external.RPCClient) (state crawlerState, err error) {
	if state.id, err = getStartingID(dgraph); err != nil {
		if !errors.Is(err, errBlockIdsDoNotMatch) {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
		info(errBlockIdsDoNotMatch.Error(), "continuing...")
	}

	if state.chainHash, err = client.GetBlockHash(int64(state.id)); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	state.hash = state.chainHash.String()

	// get RPC client block count
	numBlocks, rpcErr := getRPCNumberOfBlocks(client)
	if rpcErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), rpcErr)
		return
	}

	state.top = numBlocks

	return
}

// createTransactionHashmap creates a hash map of btcjson.TxRawResult
func createTransactionHashmap(client external.BatchRPCClient,
	transactions []string) (map[string]btcjson.TxRawResult, error) {
	type txLookup struct {
		hash   string
		result rpcclient.FutureGetRawTransactionVerboseResult
		err    error
	}

	c := make(chan txLookup, 5)

	for _, t := range transactions {
		go func(t string, c chan txLookup) {
			l := txLookup{hash: t}

			txHash, err := chainhash.NewHashFromStr(t)
			if err != nil {
				l.err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
				c <- l
				return
			}
			futureResults := client.GetRawTransactionVerboseAsync(txHash)
			if err != nil {
				l.err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
				c <- l
				return
			}

			l.result = futureResults

			c <- l
		}(t, c)
	}

	// collect future results
	var futures []txLookup
	for i := 0; i < len(transactions); i++ {
		lookup := <-c
		if lookup.err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), lookup.err)
		}
		futures = append(futures, lookup)
	}

	// send batch request
	if err := client.Send(); err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	// collect results
	txs := make(map[string]btcjson.TxRawResult)
	for _, f := range futures {
		r, err := f.result.Receive()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		txs[f.hash] = *r
	}

	return txs, nil
}

func getExternalOutputs(dgraph external.Database,
	outputs map[string][]uint32) (map[string]map[uint32]db.Output, error) {
	if len(outputs) == 0 {
		return nil, nil
	}

	transactionHashes := make([]string, 0, len(outputs))
	for k := range outputs {
		transactionHashes = append(transactionHashes, k)
	}

	transactionsOutputs, err := db.GetTransactionsOutputs(dgraph, transactionHashes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	returnMap := make(map[string]map[uint32]db.Output)

	for _, t := range transactionsOutputs {
		indexes := outputs[t.Hash]

		for _, i := range indexes {
			for _, o := range t.Outputs {
				if o.OutputIndex == nil {
					return nil, fmt.Errorf("output index was not set for tx %s", t.Hash)
				}
				if *o.OutputIndex == i {
					// add index mapping
					indexMap := returnMap[t.Hash]
					if indexMap == nil {
						indexMap = make(map[uint32]db.Output)
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
func processRound(dgraph external.Database, batchRPC external.BatchRPCClient, state crawlerState,
	block *btcjson.GetBlockVerboseResult, config Config, cache *outputCache) (
	blkCounter int64, txCounter int64, err error) {
	var txMapping []transactionMapping

	txHashMap, err := createTransactionHashmap(batchRPC, block.Tx)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), err)
		return
	}

	externalOutputs, err := getExternalOutputs(dgraph, filterExternalOutputs(txHashMap, cache))
	if err != nil {
		return 0, 0, err
	}

	transactions := make([]db.Transaction, 0, len(txHashMap))
	for _, t := range txHashMap {
		newTx, tMap, buildErr := buildTransactionMapping(t, txHashMap, externalOutputs, config, cache)
		if buildErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), buildErr)
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
		err = fmt.Errorf("wrong number of transactions in block: %s", block.Hash)
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
	if b, err = db.GetBlock(dgraph, state.hash); err != nil || !b.IsComplete() {
		// block is not yet in database -> create new block
		ts := time.Unix(block.Time, 0).Format(time.RFC3339)
		if err = processBlock(dgraph, transactions, state.hash, state.id, ts, block.PreviousHash); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), err)
			return
		}

		blkCounter++
	} else {
		// reset txCounter as the block is not processed
		txCounter = 0
	}

	blockID := int64(state.id)
	transactionOutputs, err := db.GetOutputs(dgraph, blockID, blockID)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), setErr)
				return
			}
		}

		// this cache gets all outputs
		if setErr := allOutputsCache.setOutputs(t.Hash, t.Outputs); setErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), setErr)
			return
		}
	}

	if err = processAddresses(dgraph, allOutputsCache, txMapping); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), err)
		return
	}

	// save processing state
	if err = dbstat.SetLastBlockID(dgraph, state.id); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), err)
		return
	}

	return
}
