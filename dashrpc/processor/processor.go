package processor

import (
	"context"
	"dashrpc/cmd/cliutil"
	dbaddr "dashrpc/db/address"
	dbblk "dashrpc/db/block"
	dbop "dashrpc/db/output"
	dbstat "dashrpc/db/status"
	dbtx "dashrpc/db/transaction"
	"dashrpc/rpcclient"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func info(v ...interface{}) {
	log.SetPrefix("\033[0;35mprocess\u001B[0m\t")
	log.Println(v)
	log.SetPrefix("")
}

const isDash = true

const (
	// blockTime is the average Dash block time
	blockTime = 2*time.Minute + 30*time.Second

	// newBlockIntervalTime is the time interval in which the processor checks if a new block is available
	newBlockIntervalTime = blockTime / 3

	// forkRangeLimit is the number of blocks which the RPC client must
	// be ahead of the crawler, for the crawler to include new blocks
	// in the database. This is done so potential chain forks/reordering do not need to be handled
	forkRangeLimit = 2000
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
	return fmt.Sprintf("Id: %d, Hash: %s", p.id, p.hash)
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

// maps a address to one or more indexes of a transaction
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

// adds the given uids of outputs to the address specified by addr in addresses
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
		editAddress.Outputs = append(editAddress.Outputs, dbop.Output{Uid: uid})
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
					uids = append(uids, o.Uid)
				}
			}
		}
		*addrs = addOutputsToAddresses(*addrs, mapping.hash, uids)
	}

	return
}

func buildAddresses(dgraph *dgo.Dgraph, txHash string, outputs map[string]outputMapping,
	addrMap *map[string]dbaddr.Address) (err error) {
	txFromDB, err := dbtx.GetTransaction(dgraph, txHash)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// handle output mappings
	buildAddressMapping(outputs, txFromDB.Outputs, addrMap)
	return
}

// inserts mappings between addresses and outputs in database
func processAddresses(dgraph *dgo.Dgraph, transactionMappings []TransactionMapping) (err error) {
	addrMap := make(map[string]dbaddr.Address)
	for _, mapping := range transactionMappings {
		if err = buildAddresses(dgraph, mapping.hash, mapping.outputs, &addrMap); err != nil {
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

// creates a named uid, parsable by dgraph
func createOutputUid(transaction string, outputId uint32) string {
	return "_:" + transaction + strconv.FormatUint(uint64(outputId), 10)
}

// processes the transaction specified by 'txHashString'
// 'txDetails' is the created transaction
// 'tMap' is the transaction mapping between the transaction and its output, this needed for address processing
func BuildTransactionMapping(dgraph *dgo.Dgraph, rawTransaction btcjson.TxRawResult,
	txHashMap map[string]btcjson.TxRawResult, isContinuous bool) (txDetails dbtx.Transaction, tMap TransactionMapping, err error) {
	txDetails.Hash = rawTransaction.Txid

	isCoinbaseTransaction := false
	if len(rawTransaction.Vin) == 1 && rawTransaction.Vin[0].IsCoinBase() {
		isCoinbaseTransaction = true
	} else {
		// process inputs if transaction is not a coinbase transaction
		for i, d := range rawTransaction.Vin {
			if err = processTxVin(dgraph, &txDetails, d, uint32(i), txHashMap); err != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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

		txDetails.Outputs = append(txDetails.Outputs, dbop.Output{
			Uid:         createOutputUid(rawTransaction.Txid, index),
			IsCoinbase:  &isCoinbaseTransaction,
			Amount:      &intAmount,
			TxType:      d.ScriptPubKey.Type,
			OutputIndex: &index,
		})

		if d.ScriptPubKey.Type == "pubkey" {
			asms := strings.Split(d.ScriptPubKey.Asm, " ")

			decodeString, decodeErr := hex.DecodeString(asms[0])
			if decodeErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), decodeErr)
				return
			}
			cfg := chaincfg.MainNetParams

			// for DASH address creation
			if isDash {
				cfg.PubKeyHashAddrID = 0x4c
			}

			address, addressConversionErr := btcutil.NewAddressPubKey(decodeString, &cfg)
			if addressConversionErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), addressConversionErr)
				return
			}

			pubkeyAddress := address.EncodeAddress()
			if d.ScriptPubKey.Addresses == nil {
				outputMappings = addOutputToMapping(outputMappings, pubkeyAddress, index)
			} else {
				for _, e := range d.ScriptPubKey.Addresses {
					outputMappings = addOutputToMapping(outputMappings, e, index)
					if e != pubkeyAddress {
						info("pubkey address mismatch in tx", txDetails.Hash,
							"pubkey decoded address:", pubkeyAddress,
							"rpc address:", e)
					}
				}
			}
		} else {
			for _, e := range d.ScriptPubKey.Addresses {
				outputMappings = addOutputToMapping(outputMappings, e, index)
			}
		}
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

// maps the input information to the output if it exists already in the database
func processTxVin(dgraph *dgo.Dgraph, details *dbtx.Transaction, vin btcjson.Vin, index uint32, txHashMap map[string]btcjson.TxRawResult) error {
	if vin.IsCoinBase() {
		// coin base >>input<< does not hold any valuable information, therefore we do not include it in the database
		// we can recognize coinbase outputs by checking the number of connected transactions
		return nil
	}

	refOutput := dbop.Output{
		InputIndex: &index,
	}

	if v, ok := txHashMap[vin.Txid]; ok {
		refOutput.Uid = createOutputUid(vin.Txid, vin.Vout)
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
		refOutput.Uid = output.Uid
	}

	details.Inputs = append(details.Inputs, refOutput)
	return nil
}

// builds a block with the provided arguments and inserts it in the database
func ProcessBlock(dgraph *dgo.Dgraph, transactions []dbtx.Transaction, currentHash string,
	blockId uint64, timestamp string, prevBlockHash string) (err error) {

	block := dbblk.Block{
		Hash:      currentHash,
		Timestamp: timestamp,
		Id:        &blockId,
		PrevBlock: &dbblk.Block{
			Hash: prevBlockHash,
		},
		Transactions: transactions,
	}

	err = dbblk.UpsertBlock(dgraph, block)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}
	return
}

var errorBlockIdsDoNotMatch = errors.New("block id of last crawled block and highest found block do not match")

// Gets the block id from which the crawling will be resumed. If no crawling has
// happened yet, the block id is set to 1.
func getStartingId(dgraph *dgo.Dgraph) (startId uint64, err error) {
	status, err := dbstat.GetCrawlerStatus(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if status.LastBlockId == nil {
		// last block id is not set -> we start at the beginning of the chain
		startId = 1
		return
	}

	highestBlockId, err := dbstat.GetHighestBlockId(dgraph)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if *status.LastBlockId != highestBlockId {
		err = errorBlockIdsDoNotMatch
	}

	startId = *status.LastBlockId

	return
}

func processingInterrupted() {
	info("### Block processing interrupted ###")
}

// wait for the next block
// if the interrupt receives a signal isInterrupt is true
// if the next block is available, currentBlock gets updated
func waitForNextRPCBlock(client *rpcclient.Client, interrupt <-chan struct{}, hashObj *chainhash.Hash,
	rpcNumBlocks uint64) (currentBlock *btcjson.GetBlockVerboseResult, isInterrupt bool, err error) {
	ticker := time.NewTicker(newBlockIntervalTime)
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
func getRPCNumberOfBlocks(client *rpcclient.Client) (uint64, error) {
	rpcInfo, err := client.GetBlockChainInfo()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	if rpcInfo.Blocks < 0 {
		return 0, errors.New("error RPC client block count is negativ")
	}

	return uint64(rpcInfo.Blocks), nil
}

// getInitialState creates the initial state of the processing loop
func getInitialState(dgraph *dgo.Dgraph, client *rpcclient.Client, continuous bool, startId uint64) (state crawlerProcessingState, err error) {
	if state.id, err = getStartingId(dgraph); err != nil {
		if !errors.Is(err, errorBlockIdsDoNotMatch) {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
		info(errorBlockIdsDoNotMatch.Error(), "continuing...")
	}

	if !continuous && startId > state.id {
		state.id = startId
	}

	if state.chainHash, err = client.GetBlockHash(int64(state.id)); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	state.hash = state.chainHash.String()

	return
}

// processes all blocks from startingBlockId to stoppingBlockId
func ProcessBlockRange(ctx context.Context, dgraph *dgo.Dgraph, client *rpcclient.Client,
	startingBlockId uint64, stoppingBlockId uint64) error {

	if err := dbstat.SetCrawling(dgraph, true); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	defer func() {
		if err := dbstat.SetCrawling(dgraph, false); err != nil {
			info("could not set crawling status:", err)
			return
		}
	}()

	state, err := getInitialState(dgraph, client, false, startingBlockId)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	status, err := dbstat.GetCrawlerStatus(dgraph)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	var isEmptyDatabase bool
	if status.LastBlockId == nil {
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
			err := ProcessRound(dgraph, client, state, currentBlock, isEmptyDatabase, false); err == nil {
			blkCounter += rBlockCounter
			txCounter += rTransactionCounter
			isEmptyDatabase = false
		} else {
			return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		}

		if state.id >= stoppingBlockId || currentBlock.NextHash == "" {
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

// prints the given metrics
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

// processes all blocks provided by the RPC client continuously
func ProcessBlocksContinuously(ctx context.Context, dgraph *dgo.Dgraph, client *rpcclient.Client) error {

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
	if status.LastBlockId == nil {
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
			if currentBlock.NextHash == "" || numBlocks < state.id+forkRangeLimit {
				info("Waiting for next block.", state)
				var isInterrupt bool
				// can not used short hand declaration, because it would mask currentBlock in the outer scope
				currentBlock, isInterrupt, err = waitForNextRPCBlock(client, ctx.Done(), state.chainHash, numBlocks)
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
			isEmptyDatabase, true); processErr == nil {
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

// creates a hash map of btcjson.TxRawResult
func createTransactionHashmap(client *rpcclient.Client, transactions []string) (map[string]btcjson.TxRawResult, error) {
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
func ProcessRound(dgraph *dgo.Dgraph, client *rpcclient.Client, state crawlerProcessingState,
	block *btcjson.GetBlockVerboseResult, setLowestId bool, isContinuous bool) (blkCounter uint64, txCounter uint64, err error) {
	var txMapping []TransactionMapping
	var transactions []dbtx.Transaction

	txHashMap, err := createTransactionHashmap(client, block.Tx)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, t := range txHashMap {
		var newTx dbtx.Transaction
		var tMap TransactionMapping

		newTx, tMap, err = BuildTransactionMapping(dgraph, t, txHashMap, isContinuous)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
	// the database a new block is created shallow blocks get created when a crawling process gets
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
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}

		blkCounter++
	} else {
		// reset txCounter as the block is not processed
		txCounter = 0
	}

	if err = processAddresses(dgraph, txMapping); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// save processing state
	if setLowestId {
		if err = dbstat.SetCrawlerStatus(dgraph, dbstat.CrawlerStatus{LastBlockId: &state.id,
			LowestBlockId: &state.id}); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	} else {
		if err = dbstat.SetLastBlockId(dgraph, state.id); err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		}
	}

	return
}
