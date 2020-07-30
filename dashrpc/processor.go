package dashrpc

import (
	"dashrpc/btcjson"
	dbaddr "dashrpc/db/address"
	dbblk "dashrpc/db/block"
	dbop "dashrpc/db/output"
	dbtx "dashrpc/db/transaction"
	"dashrpc/rpcclient"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

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
	// handle input mappings
	if err = buildAddressMapping(inputs, txFromDB.Inputs, addrMap); err != nil {
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
func BuildTransactionMapping(dgraph *dgo.Dgraph, client *rpcclient.Client, processAddrs bool, txHashString string) (txDetails dbtx.Transaction, tMap TransactionMapping, err error) {
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
	if processAddrs {
		tMap = TransactionMapping{hash: txDetails.Hash, outputs: outputMappings}
	}

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

	//saveProcessingState(db, currentHash.String(), blockId)
	//DbIncrementBlockCount(db)
	block := dbblk.Block{
		Hash:      currentHash,
		Timestamp: timestamp,
		Id:        &blockId,
		PrevBlock: &dbblk.Block{
			Hash: prevBlockHash,
		},
		Transactions: transactions,
	}

	return dbblk.InsertBlock(dgraph, block)
}

// processes all the new blocks from a given hash down to the block that is already in DB
func ProcessNewBlocks(dgraph *dgo.Dgraph, client *rpcclient.Client,
	includeAddresses bool, startingBlockId uint64, stoppingBlockId uint64) error {

	//DbSetStatus(dgraph, DbBlockStatusProcessing)
	//err := DbSetUint64(dgraph, DbBlockStopBlockId, stoppingBlockId)
	//if err != nil {
	//	log.Printf("Error: failed to save stopBlockID\n%v\n", err)
	//}

	startHashObj, err := client.GetBlockHash(int64(startingBlockId))
	if err != nil {
		log.Printf("we have problem with GetBlockHash() %s\n", err.Error())
		return err
	}

	blkCounter := 0
	txCounter := 0

	currentBlockId := startingBlockId
	currentBlockHash := startHashObj.String()

	// We will handle CTRL-C and CTRL-Z nicely
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	timerStart := time.Now()
	// Main loop
mainLoop:
	for {
		select {
		case <-c:
			log.Printf("\n### Block processing interrupted\n")
			break mainLoop
		default:
			// we do nothing
		}
		if b, err := dbblk.GetBlock(dgraph, currentBlockHash); err == nil && b.IsComplete() {
			// block already in database
			// todo check if continue instead of break is okay
			continue
		}

		// get block from RPC-Client
		startBlock, err := client.GetBlockVerbose(startHashObj)
		if err != nil {
			return err
		}

		var txMapping []TransactionMapping
		var transactions []dbtx.Transaction
		//fmt.Printf("%s;%d\n", currentBlockHash, len(startBlock.Tx))
		for _, t := range startBlock.Tx {
			newTx, tMap, err := BuildTransactionMapping(dgraph, client, includeAddresses, t)
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

			//DbIncrementGlobalTxCount(db)

			//if txCounter%5000 == 0 {
			//	log.Printf("%v * 5k TXs done. BlockId: %v, %v\n", txCounter/5000, currentBlockId, currentBlockHash)
			//	log.Printf("Block %s processed, tx count: %d\n", currentBlockId, txCounter)
			//}
		}

		// create new block
		ts := time.Unix(startBlock.Time, 0).Format(time.RFC3339)
		if err = ProcessBlock(dgraph, transactions, currentBlockHash, currentBlockId, ts, startBlock.PreviousHash); err != nil {
			log.Println("Error: we had problem processing the block")
			log.Printf("Hash: %s, BlockId: %d\n", currentBlockHash, currentBlockId)
			break
		}

		blkCounter++

		if includeAddresses {
			if err = processAddresses(dgraph, txMapping); err != nil {
				return err
			}
		}

		//saveProcessingState(db, currentBlockHash, block.Id-1)

		if currentBlockId == stoppingBlockId || startBlock.NextHash == "" {
			// finished
			//saveProcessingStateFinished(db)
			break
		}

		// set values for next round
		startHashObj, err = chainhash.NewHashFromStr(startBlock.NextHash)
		if err != nil {
			return err
		}

		currentBlockHash = startBlock.NextHash
		currentBlockId++

		//if blkCounter%20000 == 0 {
		//	log.Printf("%d * 20k blocks done\n", blkCounter/20000)
		//}

		if blkCounter%5 == 0 {
			log.Printf("%v ms/block\n", time.Since(timerStart).Milliseconds()/int64(blkCounter))
		}
	}

	elapsedTime := time.Since(timerStart)
	if blkCounter > 0 {
		log.Printf("Last Block Hash: %s, Id: %d\n", currentBlockHash, currentBlockId)
		log.Printf("Final Blocks count: %v\n", blkCounter)
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

// ProcessAddressClustering traverses the transactions from a given address and creates the cluster data in DB
// TODO currently there is no clustering, just a lookup for current TXs associated with a given address
func ProcessAddressClustering(db *badger.DB, startingAddr string) error {

	addrData := AddressData{}
	err := DbGetDataForAddress(db, startingAddr, &addrData)
	if err != nil {
		return err
	}
	txs := addrData.Txs

	for _, tx := range txs {
		log.Printf("TX %v -- hash: %v Amount: %f -- %v\n", tx.TxType, tx.TxHash, tx.Amount, tx.Addresses)
	}

	return nil
}

// todo implement for dgraph
// saveProcessingState saves the current processing state
func saveProcessingState(db *badger.DB, currentBlockHash string, currentBlockId uint64) {
	// Updating the upper range
	rangeUp := DbGetRangeUp(db)
	if currentBlockId > rangeUp {
		DbSetRangeUp(db, currentBlockId)
	}
	// Updating the lower range
	rangeDown := DbGetRangeDown(db)
	if rangeDown == 0 || currentBlockId < rangeDown { // lazy initialization problem
		DbSetRangeDown(db, currentBlockId)
	}

	err := DbSetUint64(db, DbBlockLastBlockId, currentBlockId)
	if err != nil {
		log.Printf("Error: error saving CurrentBlockHash state: %v\n", err)
	}
	err = DbSetString(db, DbBlockLastBlockHash, currentBlockHash)
	if err != nil {
		log.Printf("Error: error saving CurrentBlockID state: %v\n", err)
	}
}

// todo implement for dgraph
// resetting the processing state to default values
func saveProcessingStateFinished(db *badger.DB) {
	DbSetStatus(db, DbBlockStatusFinished)
}
