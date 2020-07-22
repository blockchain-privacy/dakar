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

type outputMapping struct {
	address string
	indexes []uint64
}

// adds indexOutput to an existing outputMapping in mapping. If none exists it inserts a new mapping
func addOutputToMapping(mapping []outputMapping, addr string, indexOutput uint64) []outputMapping {

	for i := range mapping {
		if mapping[i].address == addr {
			mapping[i].indexes = append(mapping[i].indexes, indexOutput)
			return mapping
		}
	}

	return append(mapping, outputMapping{
		address: addr,
		indexes: []uint64{indexOutput},
	})

}

func addOutputUidsToAddresses(addresses []dbaddr.Address, addr string, uids []string) []dbaddr.Address {

	for i := range addresses {
		if addresses[i].Hash == addr {
			for _, uid := range uids {
				addresses[i].Outputs = append(addresses[i].Outputs, dbop.Output{Uid: uid})
			}
			return addresses
		}
	}

	// create new address if none was found
	newAddr := dbaddr.Address{Hash: addr}
	for _, uid := range uids {
		newAddr.Outputs = append(newAddr.Outputs, dbop.Output{Uid: uid})
	}

	return append(addresses, newAddr)

}

func buildAddressMapping(outMap []outputMapping, outputs []dbop.Output, addrs *[]dbaddr.Address) error {
	for _, mapping := range outMap {
		var uids []string
		for _, idx := range mapping.indexes {
			for _, input := range outputs {
				if *input.Index == idx {
					uids = append(uids, input.Uid)
				}
			}
		}
		*addrs = addOutputUidsToAddresses(*addrs, mapping.address, uids)
	}

	return nil
}

func ProcessTx(dgraph *dgo.Dgraph, client *rpcclient.Client,
	processAddresses bool, txHashString string) error {

	if t, err := dbtx.GetTransaction(dgraph, txHashString); err == nil && t.IsComplete() {
		// we already have it in the system, we do nothing
		return nil
	}

	txHash, err := chainhash.NewHashFromStr(txHashString)
	if err != nil {
		fmt.Printf("Cannot convert string to Hash in ProcessTx(). String: %s", txHashString)
		return err
	}

	tx, err := client.GetRawTransactionVerbose(txHash)
	if err != nil {
		fmt.Printf("Problems getting the RawTransaction from hash: %v\n", txHash)
		return err
	}

	txDetails := dbtx.Transaction{Hash: tx.Txid}

	var inputMappings []outputMapping
	for index, d := range tx.Vin {
		uindex := uint64(index)
		iMapping, err := processTxVin(client, &txDetails, d, uindex)
		if err != nil {
			fmt.Printf("Problems with processTxVin() call in ProcessBlock(): %s", err.Error())
			continue
		}
		inputMappings = append(inputMappings, iMapping...)
	}

	var outputMappings []outputMapping

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

	if _, err = dbtx.UpsertTransaction(dgraph, &txDetails); err != nil {
		fmt.Printf("Problems updating transaction: %v\n", txDetails)
		return err
	}

	if !processAddresses || (inputMappings == nil && outputMappings == nil) {
		// no addresses to process
		return nil
	}

	txFromDB, err := dbtx.GetTransaction(dgraph, txDetails.Hash)
	if err != nil {
		return err
	}

	var addrSlice []dbaddr.Address

	// handle input mappings
	if err = buildAddressMapping(inputMappings, txFromDB.Inputs, &addrSlice); err != nil {
		fmt.Printf("Problems creating input address mapping\n")
		return err
	}

	// handle output mappings
	if err = buildAddressMapping(outputMappings, txFromDB.Outputs, &addrSlice); err != nil {
		fmt.Printf("Problems creating output address mapping\n")
		return err
	}
	if addrSlice != nil {
		if _, err = dbaddr.UpsertAddresses(dgraph, addrSlice); err != nil {
			fmt.Printf("Problems updating addresses: %v\n", txDetails)
			return err
		}
	}

	return nil
}

func processTxVin(client *rpcclient.Client, details *dbtx.Transaction,
	vin btcjson.Vin, index uint64) (mapping []outputMapping, err error) {
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

// ProcessBlock is traversing the blockchain backwards adding all unknown yet blocks.
// It stops on error or on first already known block
func ProcessBlock(dgraph *dgo.Dgraph, txHashes []string, currentHash string,
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
	}

	for _, tx := range txHashes {
		block.Transactions = append(block.Transactions, dbtx.Transaction{Hash: tx})
	}

	if err := dbblk.UpsertBlock(dgraph, block); err != nil {
		fmt.Printf("Saving block gave error %s\n", err.Error())
		return err
	}

	return nil
}

// ProcessNewBlocks process all the new blocks from a given hash down to the block that is already in DB
func ProcessNewBlocks(dgraphDb *dgo.Dgraph,
	client *rpcclient.Client,
	processAddresses bool,
	startingBlockHash string,
	startingBlockId uint64,
	stoppingBlockId uint64) error {

	timerStart := time.Now()
	//DbSetStatus(dgraphDb, DbBlockStatusProcessing)
	//err := DbSetUint64(dgraphDb, DbBlockStopBlockId, stoppingBlockId)
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

		if b, err := dbblk.GetBlock(dgraphDb, blockHash); err == nil && b.IsComplete() {
			// block already in database
			// todo check if continue instead of break is okay
			continue
		}

		startBlock, err := client.GetBlock(startHashObj)
		if err != nil {
			fmt.Printf("Problem with getBlock() %s\n", err.Error())
			break
		}

		txHashes, err := startBlock.TxHashes()
		if err != nil {
			fmt.Printf("we have problem with TxHashes() %s\n", err.Error())
		}

		// create transaction slice
		var txList []string
		for _, tx := range txHashes {
			txList = append(txList, tx.String())
		}

		ts := startBlock.Header.Timestamp.Format(time.RFC3339)
		previousBlockHash := startBlock.Header.PrevBlock.String()

		if err = ProcessBlock(dgraphDb, txList, blockHash, startingBlockId, ts, previousBlockHash); err != nil {
			fmt.Println("Error: we had problem processing the block")
			fmt.Printf("Hash: %s, BlockId: %d\n", blockHash, startingBlockId)
			break
		}

		for _, t := range txList {
			err = ProcessTx(dgraphDb, client, processAddresses, t)
			if err != nil {
				fmt.Printf("DbGetBlock() failed in tx traversal. blkcount: %v, txcount: %v\n", blkCounter, txCounter)
				fmt.Printf("Error: %s\n", err.Error())
				fmt.Printf("Tx: %v\n", t)
				break
			}
			txCounter++
			//DbIncrementGlobalTxCount(db)

			//if txCounter%5000 == 0 {
			//	fmt.Printf("%v * 5k TXs done. BlockId: %v, %v\n", txCounter/5000, startingBlockId, blockHash)
			//	fmt.Printf("Block %s processed, tx count: %d\n", startingBlockId, txCounter)
			//}
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
