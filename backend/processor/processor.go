package processor

import (
	"backend/cmd/cliutil"
	dbaddr "backend/db/address"
	dbblk "backend/db/block"
	dbop "backend/db/output"
	dbstat "backend/db/status"
	dbtx "backend/db/transaction"
	"backend/external"

	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

const (
	// loggerPrefix is the prefix which is printed for each log message
	loggerPrefix = "\033[0;35mprocess\u001B[0m\t"

	// addressInvalidPubkey is the string which gets used as an address hash
	// if its public key can not be decoded
	addressInvalidPubkey = "error_decode_pubkey"
)

var thisLogger = log.New(log.Writer(), loggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	thisLogger = log.New(out, loggerPrefix, flag)
}

func info(v ...interface{}) {
	thisLogger.Println(v...)
}

var (
	errAddressDecoding = errors.New("error while decoding address")
	errNoAddresses     = errors.New("error output has no addresses")
)

// holds the current state of the crawling processing loop
type crawlerProcessingState struct {
	// current block id
	id uint64
	// current block hash
	hash string
	// current block hash as a chainhash.Hash
	chainHash *chainhash.Hash
}

func (p crawlerProcessingState) String() string {
	return fmt.Sprintf("ID: %d, Hash: %s", p.id, p.hash)
}

// increments the state for the next processing loop
func (p *crawlerProcessingState) increment(nextHash string) (err error) {
	p.chainHash, err = chainhash.NewHashFromStr(nextHash)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	p.hash = nextHash
	p.id++
	return
}

// maps a address to one or more indexes of a transaction
type outputMapping struct {
	hash    string
	indexes []uint32
}

// TransactionMapping maps a address to one or more indexes of a transaction
type TransactionMapping struct {
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
func addOutputsToAddresses(addresses map[string]dbaddr.Address, addr string, uids []string) map[string]dbaddr.Address {
	var (
		editAddress dbaddr.Address
		ok          bool
	)

	if editAddress, ok = addresses[addr]; !ok {
		// new address -> set hash
		editAddress.Hash = addr
	}

	// add new outputs
	for _, uid := range uids {
		editAddress.Outputs = append(editAddress.Outputs, dbop.Output{UID: uid})
	}

	// save in map
	addresses[addr] = editAddress
	return addresses
}

func buildAddressMapping(outMap map[string]outputMapping, outputs []dbop.Output, addrs *map[string]dbaddr.Address) {
	for _, mapping := range outMap {
		var uids []string
		for _, idx := range mapping.indexes {
			for _, o := range outputs {
				if *o.OutputIndex == idx {
					uids = append(uids, o.UID)
				}
			}
		}
		*addrs = addOutputsToAddresses(*addrs, mapping.hash, uids)
	}
}

func buildAddresses(dgraph *external.GraphDB, txHash string, blockHash string, outputs map[string]outputMapping,
	addrMap *map[string]dbaddr.Address) (err error) {
	txFromDB, err := dbtx.GetTransaction(dgraph, txHash, blockHash)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// handle output mappings
	buildAddressMapping(outputs, txFromDB.Outputs, addrMap)
	return
}

// processAddresses inserts mappings between addresses and outputs in database
func processAddresses(dgraph *external.GraphDB, transactionMappings []TransactionMapping, blockHash string) (err error) {
	if len(transactionMappings) == 0 {
		return
	}

	addrMap := make(map[string]dbaddr.Address)
	for _, mapping := range transactionMappings {
		if err = buildAddresses(dgraph, mapping.hash, blockHash, mapping.outputs, &addrMap); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}

	// map to slice
	var addrSlice []dbaddr.Address
	for _, a := range addrMap {
		addrSlice = append(addrSlice, a)
	}

	if err = dbaddr.UpsertAddresses(dgraph, addrSlice); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return nil
}

// createOutputUID creates a named uid, parsable by dgraph
func createOutputUID(transaction string, outputID uint32) string {
	return "_:" + transaction + strconv.FormatUint(uint64(outputID), 10)
}

// decodeAddress tries to decode the address in asm
func decodeAddress(asm string, pubkeyPrefix byte) (address string, err error) {
	if len(asm) == 0 {
		err = errAddressDecoding
		return
	}

	// Alternatively use btcutil's ExtractPkScriptAddrs(); This has a lot of overhead as it
	// tries to detect the transaction type based on the script. At this point we already know
	// that it is pay-to-pubkey. Thus, parsing is done manually

	amsParts := strings.Split(asm, " ")
	if len(amsParts[0]) == 0 {
		return "", errors.New("error received invalid asm: " + asm)
	}

	decodeString, decodeErr := hex.DecodeString(amsParts[0])
	if decodeErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), decodeErr)
		return
	}
	cfg := chaincfg.MainNetParams

	cfg.PubKeyHashAddrID = pubkeyPrefix

	addr, addressConversionErr := btcutil.NewAddressPubKey(decodeString, &cfg)
	if addressConversionErr != nil {
		return addressInvalidPubkey, nil
	}
	address = addr.EncodeAddress()

	return
}

// BuildTransactionMapping processes the transaction specified by 'txHashString'
// 'txDetails' is the created transaction
// 'tMap' is the transaction mapping between the transaction and its output, this needed for address processing
func BuildTransactionMapping(dgraph *external.GraphDB, rawTransaction btcjson.TxRawResult,
	txHashMap map[string]btcjson.TxRawResult, isContinuous bool, config Config) (
	txDetails dbtx.Transaction, tMap TransactionMapping, err error) {
	txDetails.Hash = rawTransaction.Txid

	var isCoinbaseTransaction bool
	if len(rawTransaction.Vin) == 1 && rawTransaction.Vin[0].IsCoinBase() {
		isCoinbaseTransaction = true
	} else {
		// process inputs if transaction is not a coinbase transaction
		for i, d := range rawTransaction.Vin {
			if processErr := processTxVin(dgraph, &txDetails, d, uint32(i), txHashMap); processErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), processErr)
				return
			}
		}
	}

	var foundAllInputs bool
	if !isCoinbaseTransaction {
		if len(rawTransaction.Vin) == len(txDetails.Inputs) {
			foundAllInputs = true
		} else if isContinuous {
			// Only create error if this is a continuous crawl. If it is not a continuous crawl, missing inputs are
			// expected as we only consider outputs created in the given block range
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

		if d.ScriptPubKey.Type == "pubkey" {
			if d.ScriptPubKey.Addresses == nil {
				pubkeyAddress, decodeErr := decodeAddress(d.ScriptPubKey.Asm, config.PubKeyHashAddrID)
				if decodeErr != nil {
					err = fmt.Errorf("%s: %s", cliutil.ShowCallInfo(),
						fmt.Sprint(decodeErr, "hash", txDetails.Hash))
					return
				}

				// log if we get an invalid address; this can happen if an invalid public key has been provided
				if pubkeyAddress == addressInvalidPubkey {
					info("could not decode public key for tx", txDetails.Hash)
				}

				outputMappings = addOutputToMapping(outputMappings, pubkeyAddress, index)
			} else {
				for _, e := range d.ScriptPubKey.Addresses {
					outputMappings = addOutputToMapping(outputMappings, e, index)
				}
			}
		} else if d.ScriptPubKey.Addresses == nil &&
			d.ScriptPubKey.Type != "nulldata" && d.ScriptPubKey.Type != "nonstandard" {
			err = errNoAddresses
			return
		} else {
			for _, e := range d.ScriptPubKey.Addresses {
				outputMappings = addOutputToMapping(outputMappings, e, index)
			}
		}

		// create new output
		txDetails.Outputs = append(txDetails.Outputs, dbop.Output{
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
	tMap = TransactionMapping{hash: txDetails.Hash, outputs: outputMappings}

	return
}

// processTxVin maps the input information to the output if it exists already in the database
func processTxVin(dgraph *external.GraphDB, details *dbtx.Transaction, vin btcjson.Vin, index uint32, txHashMap map[string]btcjson.TxRawResult) error {
	if vin.IsCoinBase() {
		// coin base >>input<< does not hold any valuable information, therefore we do not include it in the database
		// we can recognize coinbase outputs by checking the number of connected transactions
		return nil
	}

	refOutput := dbop.Output{
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
	} else {
		output, err := dbop.GetOutput(dgraph, vin.Txid, vin.Vout, false)
		if err != nil {
			// origin transaction of output does not exist in database, ignore input
			// this can happen if we process a transaction which uses an output of a transaction which is not included in our block range
			// e.g. our range is block 5 -- 15 and we process a transaction in block 10 which uses an output from a transaction in block 4
			if errors.Is(err, dbop.ErrorNotFound) {
				return nil
			}

			return err
		}

		refOutput.Amount = output.Amount
		refOutput.UID = output.UID
	}

	details.Inputs = append(details.Inputs, refOutput)
	return nil
}

// ProcessBlock builds a block with the provided arguments and inserts it in the database
func ProcessBlock(dgraph *external.GraphDB, transactions []dbtx.Transaction, currentHash string,
	blockID uint64, timestamp string, prevBlockHash string) (err error) {

	if err = dbblk.UpsertBlock(dgraph, dbblk.Block{
		Hash:      currentHash,
		Timestamp: timestamp,
		ID:        &blockID,
		PrevBlock: &dbblk.Block{
			Hash: prevBlockHash,
		},
		Transactions: transactions,
	}); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	return
}

var errorBlockIdsDoNotMatch = errors.New("block id of last crawled block and highest found block do not match")

// getStartingID gets the block id from which the crawling will be resumed. If no crawling has
// happened yet, the block id is set to 1.
func getStartingID(dgraph *external.GraphDB) (startID uint64, err error) {
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
		err = errorBlockIdsDoNotMatch
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
		// check if block is available and and if it is an actual new block
		if currentBlock.NextHash != "" && numBlocks > rpcNumBlocks {
			break
		}
	}

	return
}

// getRPCNumberOfBlocks returns the number of blocks currently processed by the RPC client
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
func getInitialState(dgraph *external.GraphDB, client external.RPCClient, continuous bool, startID uint64) (state crawlerProcessingState, err error) {
	if state.id, err = getStartingID(dgraph); err != nil {
		if !errors.Is(err, errorBlockIdsDoNotMatch) {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
		info(errorBlockIdsDoNotMatch.Error(), "continuing...")
	}

	if !continuous && startID > state.id {
		state.id = startID
	}

	if state.chainHash, err = client.GetBlockHash(int64(state.id)); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	state.hash = state.chainHash.String()

	return
}

// ProcessBlockRange processes all blocks from startingBlockId to stoppingBlockId
func ProcessBlockRange(ctx context.Context, dgraph *external.GraphDB, client external.RPCClient,
	startingBlockID uint64, stoppingBlockID uint64, config Config) error {

	if err := dbstat.SetCrawling(dgraph, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	defer func() {
		if err := dbstat.SetCrawling(dgraph, false); err != nil {
			info("could not set crawling status:", err)
			return
		}
	}()

	state, err := getInitialState(dgraph, client, false, startingBlockID)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	status, err := dbstat.GetCrawlerStatus(dgraph)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	var isEmptyDatabase bool
	if status.LastBlockID == nil {
		isEmptyDatabase = true
	}

	var blkCounter uint64
	var txCounter uint64

	info("Starting crawling at", state)

	timerStart := time.Now()
	// Main loop

	var currentBlock *btcjson.GetBlockVerboseResult

mainLoop:
	for {
		select {
		case <-ctx.Done():
			processingInterrupted()
			break mainLoop
		default:
			// we do nothing
		}

		// get block from RPC-Client
		currentBlock, err = client.GetBlockVerbose(state.chainHash)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// do the actual processing and aggregate the resulting metrics
		if rBlockCounter, rTransactionCounter,
			err := ProcessRound(dgraph, client, state, currentBlock, isEmptyDatabase, false, config); err == nil {
			blkCounter += rBlockCounter
			txCounter += rTransactionCounter
			isEmptyDatabase = false
		} else {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if state.id >= stoppingBlockID || currentBlock.NextHash == "" {
			// finished
			break
		}

		if err = state.increment(currentBlock.NextHash); err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}
	}

	printMetrics(state, blkCounter, txCounter, time.Since(timerStart))

	return nil
}

// printMetrics prints the given metrics
func printMetrics(state crawlerProcessingState, blkCounter uint64, txCounter uint64, elapsedTime time.Duration) {
	if blkCounter > 0 {
		info("Last Block:", state)
		info("New blocks inserted:", blkCounter)
		info("Final TX count:", txCounter)
		info("Elapsed time:", elapsedTime)
		info("Performance:", elapsedTime.Milliseconds()/int64(blkCounter), "ms/block")
	} else {
		info("Processed no new blocks")
		info("Final TX count:", txCounter)
		info("Elapsed time", elapsedTime)
	}
}

// ProcessBlocksContinuously processes all blocks provided by the RPC client continuously
func ProcessBlocksContinuously(ctx context.Context, dgraph *external.GraphDB, client external.RPCClient, config Config) error {
	if err := dbstat.SetCrawling(dgraph, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	defer func() {
		if err := dbstat.SetCrawling(dgraph, false); err != nil {
			info("could not set crawling status:", err)
			return
		}
	}()

	state, err := getInitialState(dgraph, client, true, 0)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	status, err := dbstat.GetCrawlerStatus(dgraph)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	var isEmptyDatabase bool
	if status.LastBlockID == nil {
		isEmptyDatabase = true
	}

	var blkCounter uint64
	var txCounter uint64

	info("Starting crawling at", state)

	timerStart := time.Now()
	// Main loop

	firstLoop := true
	var currentBlock *btcjson.GetBlockVerboseResult

mainLoop:
	for {
		select {
		case <-ctx.Done():
			processingInterrupted()
			break mainLoop
		default:
			// we do nothing
		}

		if !firstLoop {
			// get RPC client block count
			numBlocks, rpcErr := getRPCNumberOfBlocks(client)
			if rpcErr != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), rpcErr)
			}
			// set values for this round
			if currentBlock.NextHash == "" || numBlocks < state.id+config.ForkRangeLimit {
				info("Waiting for next block.", state)
				var isInterrupt bool
				// can not used short hand declaration, because it would mask currentBlock in the outer scope
				currentBlock, isInterrupt, err = waitForNextRPCBlock(client, ctx.Done(), state.chainHash, numBlocks, config)
				if err != nil {
					return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
				}

				if isInterrupt {
					break mainLoop
				}

				info("Found next block. Old state:", state)
			}
		}

		// if not first round -> increment state
		if !firstLoop {
			if err = state.increment(currentBlock.NextHash); err != nil {
				return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			}
		}

		firstLoop = false
		// get block from RPC-Client
		currentBlock, err = client.GetBlockVerbose(state.chainHash)
		if err != nil {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		// do the actual processing and aggregate the resulting metrics
		if rBlockCounter, rTransactionCounter, processErr := ProcessRound(dgraph, client, state, currentBlock,
			isEmptyDatabase, true, config); processErr == nil {
			blkCounter += rBlockCounter
			txCounter += rTransactionCounter
			isEmptyDatabase = false
		} else {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), processErr)
			return err
		}
	}

	printMetrics(state, blkCounter, txCounter, time.Since(timerStart))

	return nil
}

// createTransactionHashmap creates a hash map of btcjson.TxRawResult
func createTransactionHashmap(client external.RPCClient, transactions []string) (map[string]btcjson.TxRawResult, error) {
	txs := make(map[string]btcjson.TxRawResult)
	for _, t := range transactions {
		txHash, err := chainhash.NewHashFromStr(t)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return txs, err
		}
		tx, err := client.GetRawTransactionVerbose(txHash)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return txs, err
		}
		txs[t] = *tx
	}

	return txs, nil
}

// ProcessRound process the given block. Hat includes the insertion of the block,
// its transaction, the outputs of all transaction and the mapping between outputs and addresses
func ProcessRound(dgraph *external.GraphDB, client external.RPCClient, state crawlerProcessingState,
	block *btcjson.GetBlockVerboseResult, setLowestID bool, isContinuous bool, config Config) (
	blkCounter uint64, txCounter uint64, err error) {
	var txMapping []TransactionMapping
	var transactions []dbtx.Transaction

	txHashMap, err := createTransactionHashmap(client, block.Tx)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), err)
		return
	}

	for _, t := range txHashMap {
		newTx, tMap, buildErr := BuildTransactionMapping(dgraph, t, txHashMap, isContinuous, config)
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
	var b dbblk.Block
	if b, err = dbblk.GetBlock(dgraph, state.hash); err != nil || !b.IsComplete() {
		// block is not yet in database -> create new block
		ts := time.Unix(block.Time, 0).Format(time.RFC3339)
		if err = ProcessBlock(dgraph, transactions, state.hash, state.id, ts, block.PreviousHash); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), err)
			return
		}

		blkCounter++
	} else {
		// reset txCounter as the block is not processed
		txCounter = 0
	}

	if err = processAddresses(dgraph, txMapping, state.hash); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), err)
		return
	}

	// save processing state
	if setLowestID {
		if err = dbstat.SetCrawlerStatus(dgraph, dbstat.CrawlerStatus{LastBlockID: &state.id,
			LowestBlockID: &state.id}); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), err)
			return
		}
	} else {
		if err = dbstat.SetLastBlockID(dgraph, state.id); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo()+state.String(), err)
			return
		}
	}

	return
}
