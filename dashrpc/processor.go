package dashrpc

import (
	"dashrpc/btcjson"
	dbaddr "dashrpc/db/address"
	dbblk "dashrpc/db/block"
	dbop "dashrpc/db/output"
	dbtx "dashrpc/db/transaction"
	"dashrpc/rpcclient"
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/dgo/v2"
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
				if *o.Index == idx {
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

func BuildTransactionMapping(client *rpcclient.Client, processAddrs bool, txHashString string) (txDetails dbtx.Transaction, tMap TransactionMapping, err error) {
	txHash, err := chainhash.NewHashFromStr(txHashString)
	if err != nil {
		fmt.Printf("Cannot convert string to Hash in BuildTransactionMapping(). String: %s", txHashString)
		return txDetails, tMap, err
	}

	tx, err := client.GetRawTransactionVerbose(txHash)
	if err != nil {
		fmt.Printf("Problems getting the RawTransaction from hash: %v\n", txHash)
		return txDetails, tMap, err
	}

	txDetails.Hash = tx.Txid

	inputMappings := make(map[string]outputMapping)

	for index, d := range tx.Vin {
		uindex := uint64(index)
		inputMappings, err = processTxVin(client, &txDetails, d, uindex, inputMappings)
		if err != nil {
			fmt.Printf("Problems with processTxVin() call in ProcessBlock(): %s", err.Error())
			return txDetails, tMap, err
		}
	}

	outputMappings := make(map[string]outputMapping)

	for _, d := range tx.Vout {
		uindex := uint64(d.N)
		isCoinBase := false
		amount := d.Value
		txDetails.Outputs = append(txDetails.Outputs, dbop.Output{
			IsCoinbase: &isCoinBase,
			Amount:     &amount,
			TxType:     d.ScriptPubKey.Type,
			Index:      &uindex,
		})

		for _, e := range d.ScriptPubKey.Addresses {
			outputMappings = addOutputToMapping(outputMappings, e, uindex)
		}
	}

	if processAddrs {
		tMap = TransactionMapping{hash: txDetails.Hash, inputs: inputMappings, outputs: outputMappings}
	}

	return txDetails, tMap, err
}

func processTxVin(client *rpcclient.Client, details *dbtx.Transaction, vin btcjson.Vin,
	index uint64, srcMapping map[string]outputMapping) (mapping map[string]outputMapping, err error) {
	mapping = srcMapping
	isCoinbase := vin.IsCoinBase()
	out := dbop.Output{
		IsCoinbase: &isCoinbase,
		Index:      &index,
	}

	if vin.IsCoinBase() {
		details.Inputs = append(details.Inputs, out)
		return nil, nil
	}

	h, err := chainhash.NewHashFromStr(vin.Txid)
	if err != nil {
		fmt.Printf("Problems with converting str to hash in showTxVinDetails: %s", err.Error())
		return nil, err
	}

	tx, err := client.GetRawTransactionVerbose(h)
	if err != nil {
		fmt.Printf("Problems with getting Tx details: %s\nHash: %v\nVin: %v\n", h.String(), vin, err.Error())
		return nil, err
	}
	out.Amount = &tx.Vout[vin.Vout].Value

	for _, e := range tx.Vout[vin.Vout].ScriptPubKey.Addresses {
		mapping = addOutputToMapping(mapping, e, *out.Index)
	}

	details.Inputs = append(details.Inputs, out)
	return mapping, nil
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

	return dbblk.UpsertBlock(dgraph, block)
}

// processes all the new blocks from a given hash down to the block that is already in DB
func ProcessNewBlocks(dgraph *dgo.Dgraph,
	client *rpcclient.Client,
	includeAddresses bool,
	startingBlockHash string,
	startingBlockId uint64,
	stoppingBlockId uint64) error {

	timerStart := time.Now()
	//DbSetStatus(dgraph, DbBlockStatusProcessing)
	//err := DbSetUint64(dgraph, DbBlockStopBlockId, stoppingBlockId)
	//if err != nil {
	//	fmt.Printf("Error: failed to save stopBlockID\n%v\n", err)
	//}

	blkCounter := 0
	txCounter := 0

	blockHash := startingBlockHash
	startHashObj, err := chainhash.NewHashFromStr(blockHash)
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
		return err
	}

	var lastBlockHashObj *chainhash.Hash

	// We will handle CTRL-C and CTRL-Z nicely
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Main loop
mainLoop:
	for {
		select {
		case <-c:
			fmt.Printf("\n### Block processing interrupted\n\nLast processed block #%d - %s\n\n",
				startingBlockId, lastBlockHashObj)
			break mainLoop
		default:
			// we do nothing
		}

		if b, err := dbblk.GetBlock(dgraph, blockHash); err == nil && b.IsComplete() {
			// block already in database
			// todo check if continue instead of break is okay
			continue
		}

		// get block from RPC-Client
		startBlock, err := client.GetBlock(startHashObj)
		if err != nil {
			fmt.Printf("Problem with getBlock() %s\n", err.Error())
			break
		}

		txHashes, err := startBlock.TxHashes()
		if err != nil {
			fmt.Printf("we have problem with TxHashes() %s\n", err.Error())
		}

		var txMapping []TransactionMapping
		var transactions []dbtx.Transaction
		for _, t := range txHashes {
			newTx, tMap, err := BuildTransactionMapping(client, includeAddresses, t.String())
			if err != nil {
				fmt.Printf("DbGetBlock() failed in tx traversal. blkcount: %v, txcount: %v\n", blkCounter, txCounter)
				fmt.Printf("Error: %s\n", err.Error())
				fmt.Printf("Tx: %v\n", t)
				return err
			}

			txCounter++
			transactions = append(transactions, newTx)
			if tMap.hash != "" && (len(tMap.inputs) > 0 || len(tMap.outputs) > 0) {
				txMapping = append(txMapping, tMap)
			}

			//DbIncrementGlobalTxCount(db)

			//if txCounter%5000 == 0 {
			//	fmt.Printf("%v * 5k TXs done. BlockId: %v, %v\n", txCounter/5000, startingBlockId, blockHash)
			//	fmt.Printf("Block %s processed, tx count: %d\n", startingBlockId, txCounter)
			//}
		}

		// create new block
		ts := startBlock.Header.Timestamp.Format(time.RFC3339)
		previousBlockHash := startBlock.Header.PrevBlock.String()
		if err = ProcessBlock(dgraph, transactions, blockHash, startingBlockId, ts, previousBlockHash); err != nil {
			fmt.Println("Error: we had problem processing the block")
			fmt.Printf("Hash: %s, BlockId: %d\n", blockHash, startingBlockId)
			break
		}

		if includeAddresses {
			if err = processAddresses(dgraph, txMapping); err != nil {
				return err
			}
		}

		//saveProcessingState(db, blockHash, block.Id-1)

		if startingBlockId == 0 || startingBlockId == stoppingBlockId || blockHash == "0000000000000000000000000000000000000000000000000000000000000000" {
			// finished
			//saveProcessingStateFinished(db)
			break
		}

		// set values for next round
		startHashObj = &(startBlock.Header.PrevBlock)
		lastBlockHashObj = startHashObj
		blockHash = previousBlockHash

		startingBlockId--
		blkCounter++

		//if blkCounter%20000 == 0 {
		//	fmt.Printf("%d * 20k blocks done\n", blkCounter/20000)
		//}

		if blkCounter%5 == 0 {
			fmt.Printf("%v ms/block\n", time.Since(timerStart).Milliseconds()/int64(blkCounter))
		}
	}

	elapsedTime := time.Since(timerStart)
	if blkCounter > 0 {
		fmt.Printf("Final Blocks count: %v\n", blkCounter)
		fmt.Printf("Final TX count: %v\n", txCounter)
		fmt.Printf("Elapsed time: %s\nPerformance: %v ms/block\n\n", elapsedTime,
			elapsedTime.Milliseconds()/int64(blkCounter))
	} else {
		fmt.Println("Processed no new blocks")
		fmt.Printf("Final TX count: %v\n", txCounter)
		fmt.Printf("Elapsed time: %s\n\n", elapsedTime)
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
		fmt.Printf("TX %v -- hash: %v Amount: %f -- %v\n", tx.TxType, tx.TxHash, tx.Amount, tx.Addresses)
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
		fmt.Printf("Error: error saving CurrentBlockHash state: %v\n", err)
	}
	err = DbSetString(db, DbBlockLastBlockHash, currentBlockHash)
	if err != nil {
		fmt.Printf("Error: error saving CurrentBlockID state: %v\n", err)
	}
}

// todo implement for dgraph
// resetting the processing state to default values
func saveProcessingStateFinished(db *badger.DB) {
	DbSetStatus(db, DbBlockStatusFinished)
}
