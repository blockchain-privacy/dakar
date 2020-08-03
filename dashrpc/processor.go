package dashrpc

import (
	"dashrpc/btcjson"
	dbaddr "dashrpc/db/address"
	dbblk "dashrpc/db/block"
	dbop "dashrpc/db/output"
	dbstat "dashrpc/db/status"
	dbtx "dashrpc/db/transaction"
	"dashrpc/rpcclient"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const (
	VersionString = "v0.0.1"

	// average Dash block time
	blockTime = 2*time.Minute + 30*time.Second

	// time interval in which the processor checks if a new block is available
	newBlockIntervalTime = blockTime / 3
)

// holds the current state of the processing loop
type processingState struct {
	// current block id
	id uint64
	// current block hash
	hash string
	// current block hash as a chainhash.Hash
	chainHash *chainhash.Hash
}

func (p processingState) String() string {
	return fmt.Sprintf("Id: %d, Hash: %s", p.id, p.hash)
}

// increments the state for the next processing loop
func (p *processingState) increment(nextHash string) (err error) {
	p.chainHash, err = chainhash.NewHashFromStr(nextHash)
	if err != nil {
		return err
	}

	p.hash = nextHash
	p.id++
	return
}

// maps a address to one or more indexes of a transaction
type outputMapping struct {
	hash    string
	indexes []uint64
}

// maps a address to one or more indexes of a transaction
type TransactionMapping struct {
	hash    string
	inputs  map[string]outputMapping
	outputs map[string]outputMapping
}

// adds indexOutput to an existing outputMapping in mapping. If none exists it inserts a new mapping
func addOutputToMapping(mapping map[string]outputMapping, addr string, indexOutput uint64) map[string]outputMapping {
	if val, ok := mapping[addr]; ok {
		val.indexes = append(val.indexes, indexOutput)
		return mapping
	}

	mapping[addr] = outputMapping{
		hash:    addr,
		indexes: []uint64{indexOutput},
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

func buildAddressMapping(outMap map[string]outputMapping, outputs []dbop.Output, addrs *map[string]dbaddr.Address) error {
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

	return nil
}

func buildAddresses(dgraph *dgo.Dgraph, txHash string, inputs map[string]outputMapping,
	outputs map[string]outputMapping, addrMap *map[string]dbaddr.Address) error {
	txFromDB, err := dbtx.GetTransaction(dgraph, txHash)
	if err != nil {
		return err
	}

	// handle output mappings
	return buildAddressMapping(outputs, txFromDB.Outputs, addrMap)
}

// inserts mappings between addresses and outputs in database
func processAddresses(dgraph *dgo.Dgraph, transactionMappings []TransactionMapping) error {
	addrMap := make(map[string]dbaddr.Address)
	for _, mapping := range transactionMappings {
		if err := buildAddresses(dgraph, mapping.hash, mapping.inputs, mapping.outputs, &addrMap); err != nil {
			return err
		}
	}

	// map to slice
	var addrSlice []dbaddr.Address
	for _, a := range addrMap {
		addrSlice = append(addrSlice, a)
	}

	if _, err := dbaddr.UpsertAddresses(dgraph, addrSlice); err != nil {
		return err
	}

	return nil
}

// processes the transaction specified by 'txHashString'
// 'txDetails' is the created transaction
// 'tMap' is the transaction mapping between the transaction and its output, this needed for address processing
func BuildTransactionMapping(dgraph *dgo.Dgraph, client *rpcclient.Client, txHashString string) (txDetails dbtx.Transaction, tMap TransactionMapping, err error) {
	txHash, err := chainhash.NewHashFromStr(txHashString)
	if err != nil {
		log.Printf("Cannot convert string to Hash in BuildTransactionMapping(). String: %s", txHashString)
		return txDetails, tMap, err
	}

	tx, err := client.GetRawTransactionVerbose(txHash)
	if err != nil {
		log.Printf("Problems getting the RawTransaction from hash: %v\n", txHash)
		return txDetails, tMap, err
	}

	txDetails.Hash = tx.Txid

	isCoinbaseTransaction := false
	if len(tx.Vin) == 1 && tx.Vin[0].IsCoinBase() {
		isCoinbaseTransaction = true
	} else {
		// process inputs if transaction is not a coinbase transaction
		for i, d := range tx.Vin {
			if err = processTxVin(dgraph, &txDetails, d, uint64(i)); err != nil {
				log.Printf("Problems with processTxVin() call in ProcessBlock(): %s", err.Error())
				return txDetails, tMap, err
			}
		}
	}

	// process all outputs
	outputMappings := make(map[string]outputMapping)
	for _, d := range tx.Vout {
		uindex := uint64(d.N)
		amount := d.Value
		txDetails.Outputs = append(txDetails.Outputs, dbop.Output{
			IsCoinbase:  &isCoinbaseTransaction,
			Amount:      &amount,
			TxType:      d.ScriptPubKey.Type,
			OutputIndex: &uindex,
		})

		for _, e := range d.ScriptPubKey.Addresses {
			outputMappings = addOutputToMapping(outputMappings, e, uindex)
		}
	}

	// create transaction mapping for address processing later on
	tMap = TransactionMapping{hash: txDetails.Hash, outputs: outputMappings}

	return txDetails, tMap, err
}

// maps the input information to the output if it exists already in the database
func processTxVin(dgraph *dgo.Dgraph, details *dbtx.Transaction, vin btcjson.Vin, index uint64) error {
	if vin.IsCoinBase() {
		// coin base >>input<< does not hold any valuable information, therefore we do not include it in the database
		// we can recognize coinbase outputs by checking the number of connected transactions
		return nil
	}

	output, err := dbop.GetOutput(dgraph, vin.Txid, vin.Vout, false)
	if err != nil {
		// origin transaction of output does not exist in database, ignore input
		// this can happen if we process a transaction which uses an output of a transaction which is not included in our block range
		// e.g. our range is block 5 -- 15 and we process a transaction in block 10 which uses an output from a transaction in block 4
		if err.Error() == dbop.ErrorNotFound {
			return nil
		}
		return err
	}

	details.Inputs = append(details.Inputs, dbop.Output{
		Uid:        output.Uid,
		InputIndex: &index,
	})
	return nil
}

// builds a block with the provided arguments and inserts it in the database
func ProcessBlock(dgraph *dgo.Dgraph, transactions []dbtx.Transaction, currentHash string,
	blockId uint64, timestamp string, prevBlockHash string) error {

	block := dbblk.Block{
		Hash:      currentHash,
		Timestamp: timestamp,
		Id:        &blockId,
		PrevBlock: &dbblk.Block{
			Hash: prevBlockHash,
		},
		Transactions: transactions,
	}

	return dbblk.UpsertBlock(dgraph, block)
}

func getStartingId(dgraph *dgo.Dgraph, continuous bool, startBlockId uint64) (startId uint64, err error) {
	if !continuous {
		startId = startBlockId
		return
	}

	status, err := dbstat.GetVerbose(dgraph)
	if err != nil {
		if err.Error() == dbstat.ErrorLastBlockIdNotFound {
			// last block id is not set -> we start at the beginning of the chain
			startId = 1
			err = nil
			return
		}
		return
	}

	if status.LastBlockId != status.HighestBlockId {
		err = errors.New("last crawled block and highest block are not the same! Status: " + status.String())
		return
	}

	startId = status.LastBlockId

	return
}

func processingInterrupted() {
	log.Printf("### Block processing interrupted ###")

}

// wait for the next block
// if the interrupt receives a signal isInterrupt is true
// if the next block is available, currentBlock gets updated
func waitForNextBlock(client *rpcclient.Client, interrupt <-chan os.Signal, hashObj *chainhash.Hash) (currentBlock *btcjson.GetBlockVerboseResult, isInterrupt bool, err error) {
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
				return
			}
		}

		if currentBlock.NextHash != "" {
			break
		}
	}

	return
}

// creates the initial state of the processing loop
func getInitialState(dgraph *dgo.Dgraph, client *rpcclient.Client, continuous bool, startId uint64) (state processingState, err error) {
	if state.id, err = getStartingId(dgraph, continuous, startId); err != nil {
		return
	}

	if state.chainHash, err = client.GetBlockHash(int64(state.id)); err != nil {
		return
	}
	state.hash = state.chainHash.String()

	return
}

// processes all the new blocks from a given hash down to the block that is already in DB
func ProcessNewBlocks(dgraph *dgo.Dgraph, client *rpcclient.Client, continuous bool,
	startingBlockId uint64, stoppingBlockId uint64) error {

	state, err := getInitialState(dgraph, client, continuous, startingBlockId)
	if err != nil {
		return err
	}

	blkCounter := 0
	txCounter := 0

	log.Println("Starting crawling at", state)

	// We will handle CTRL-C and CTRL-Z nicely
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	timerStart := time.Now()
	// Main loop

	firstLoop := true
	var currentBlock *btcjson.GetBlockVerboseResult

mainLoop:
	for {
		select {
		case <-c:
			processingInterrupted()
			break mainLoop
		default:
			// we do nothing
		}

		if continuous && !firstLoop {
			// set values for this round
			if currentBlock.NextHash == "" {
				log.Println("Waiting for next block.", state)
				var isInterrupt bool
				// can not used short hand declaration, because it would mask currentBlock in the outer scope
				currentBlock, isInterrupt, err = waitForNextBlock(client, c, state.chainHash)
				if err != nil {
					return err
				}

				if isInterrupt {
					break mainLoop
				}

				log.Println("Found next block.", state)
			}
		}

		// if not the first round increment state
		if !firstLoop {
			if err = state.increment(currentBlock.NextHash); err != nil {
				return err
			}
		}

		// check for stop conditions if not stop
		if !continuous {
			// stoppingBlockId+1 <- +1 because we still need to process this round
			if state.id == stoppingBlockId+1 || (currentBlock != nil && currentBlock.NextHash == "") {
				// finished
				break
			}
		}

		firstLoop = false
		// get block from RPC-Client
		currentBlock, err = client.GetBlockVerbose(state.chainHash)
		if err != nil {
			return err
		}

		var txMapping []TransactionMapping
		var transactions []dbtx.Transaction

		for _, t := range currentBlock.Tx {
			newTx, tMap, err := BuildTransactionMapping(dgraph, client, t)
			if err != nil {
				log.Printf("DbGetBlock() failed in tx traversal. blkcount: %v, txcount: %v\n", blkCounter, txCounter)
				log.Printf("Error: %s\n", err.Error())
				log.Printf("Tx: %v\n", t)
				return err
			}

			txCounter++
			transactions = append(transactions, newTx)
			if tMap.hash != "" && (len(tMap.inputs) > 0 || len(tMap.outputs) > 0) {
				txMapping = append(txMapping, tMap)
			}

			//if txCounter%5000 == 0 {
			//	log.Printf("%v * 5k TXs done. BlockId: %v, %v\n", txCounter/5000, state.id, state.hash)
			//	log.Printf("Block %s processed, tx count: %d\n", state.id, txCounter)
			//}
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
		if b, err := dbblk.GetBlock(dgraph, state.hash); err != nil || !b.IsComplete() {
			// block is not yet in database -> create new block
			ts := time.Unix(currentBlock.Time, 0).Format(time.RFC3339)
			if err = ProcessBlock(dgraph, transactions, state.hash, state.id, ts, currentBlock.PreviousHash); err != nil {
				log.Println("Error: we had problem processing the block.", state)
				break
			}

			blkCounter++
		} else {
			// reduce txCounter as the currentBlock is not processed
			txCounter -= len(currentBlock.Tx)
		}

		if err = processAddresses(dgraph, txMapping); err != nil {
			return err
		}

		// save processing state
		if err = dbstat.SetLastBlockId(dgraph, state.id); err != nil {
			log.Printf("error saving state.id state: %v\n", err)
			return err
		}

		if blkCounter%10 == 0 {
			log.Printf("%v ms/block\n", time.Since(timerStart).Milliseconds()/int64(blkCounter))
		}
	}

	elapsedTime := time.Since(timerStart)
	if blkCounter > 0 {
		log.Println("Last Block:", state)
		log.Printf("New blocks inserted: %v\n", blkCounter)
		log.Printf("Final TX count: %v\n", txCounter)
		log.Printf("Elapsed time: %s\n", elapsedTime)
		log.Printf("Performance: %v ms/block", elapsedTime.Milliseconds()/int64(blkCounter))
	} else {
		log.Println("Processed no new blocks")
		log.Printf("Final TX count: %v\n", txCounter)
		log.Printf("Elapsed time: %s", elapsedTime)
	}

	return nil
}
