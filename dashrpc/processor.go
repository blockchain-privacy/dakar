package dashrpc

import (
	"dashrpc/btcjson"
	"dashrpc/db"
	"dashrpc/rpcclient"
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/dgo/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// ProcessTx process transaction, and the Vout and Vin records
func ProcessTx(db *badger.DB, client *rpcclient.Client, processAddresses bool, txHashString string) error {
	txDetails := TxDetails{}
	err := DbGetTxDetails(db, txHashString, &txDetails)
	if err == nil {
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

	txDetails.Hash = tx.Txid
	txDetails.Timestamp = tx.Time
	for index, d := range tx.Vin {
		err := processTxVin(db, client, processAddresses, &txDetails, d, index)
		if err != nil {
			fmt.Printf("Problems with processTxVin() call in ProcessBlock(): %s", err.Error())
		}
	}

	for index, d := range tx.Vout {
		err := processTxVout(&txDetails, d, index)
		if err != nil {
			fmt.Printf("Problems with processTxVout() call in ProcessBlock(): %s", err.Error())
		}
	}

	return DbSetTxDetails(db, txDetails)
}

func ProcessTx2(dgraph *dgo.Dgraph, client *rpcclient.Client,
	processAddresses bool, txHashString string, blockUid string) error {
	txDetails := db.Transaction{}

	err := db.GetCompleteTransaction(dgraph, txHashString, &txDetails)
	if err == nil {
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

	txDetails.Hash = tx.Txid

	txDetails.Timestamp = time.Unix(tx.Time, 0).Format(time.RFC3339)
	for index, d := range tx.Vin {
		err := processTxVin2(dgraph, client, processAddresses, &txDetails, d, index)
		if err != nil {
			fmt.Printf("Problems with processTxVin() call in ProcessBlock(): %s", err.Error())
		}
	}

	for index, d := range tx.Vout {
		//err := processTxVout2(&txDetails, d, index)

		txDetails.Outputs = append(txDetails.Outputs, db.TxOutput{
			IsCoinbase: false,
			Amount:     d.Value,
			TxType:     d.ScriptPubKey.Type,
			Index:      index,
		})

		//
		//if err != nil {
		//	fmt.Printf("Problems with processTxVout() call in ProcessBlock(): %s", err.Error())
		//}
	}

	// todo: process addresses after in- and outputs

	_, err = db.UpdateTransaction(dgraph, &txDetails)

	return err
}

func processTxVin2(dgraph *dgo.Dgraph, client *rpcclient.Client, processAddresses bool,
	details *db.Transaction, vin btcjson.Vin, index int) error {
	out := db.TxOutput{
		IsCoinbase: vin.IsCoinBase(),
		Index:      index,
	}

	if out.IsCoinbase {
		details.Inputs = append(details.Inputs, out)
		return nil
	}

	h, err := chainhash.NewHashFromStr(vin.Txid)
	if err != nil {
		fmt.Printf("Problems with converting str to hash in showTxVinDetails: %s", err.Error())
		return err
	}

	tx, err := client.GetRawTransactionVerbose(h)
	if err != nil {
		fmt.Printf("Problems with getting Tx details: %s\nHash: %v\nVin: %v\n", h.String(), vin, err.Error())
		return err
	}

	out.Amount = tx.Vout[vin.Vout].Value

	//for _, e := range tx.Vout[vin.Vout].ScriptPubKey.Addresses {
	//	out.Addresses = append(out.Addresses, db.Address{Hash: e})
	//}
	//
	//if processAddresses {
	//	// Let's associate the address with Tx
	//	for _, addr := range out.Addresses {
	//		err = DbAddTxToAddress2(dgraph, addr.Hash, out)
	//		if err != nil {
	//			// TODO for performance reasons, unspent TXs that cannot be linked with address are ignored
	//			// fmt.Printf("Problem adding Tx to address, %v. Error: %s", out, err.Error())
	//
	//		}
	//	}
	//}

	details.Inputs = append(details.Inputs, out)
	return nil
}

func processTxVin(db *badger.DB, client *rpcclient.Client, processAddresses bool, details *TxDetails, vin btcjson.Vin, index int) error {
	out := TxOutput{}
	out.IsCoinbase = vin.IsCoinBase()
	out.TxHash = vin.Txid
	if out.IsCoinbase {
		details.Inputs = append(details.Inputs, out)
		return nil
	}

	h, err := chainhash.NewHashFromStr(vin.Txid)
	if err != nil {
		fmt.Printf("Problems with converting str to hash in showTxVinDetails: %s", err.Error())
		return err
	}

	tx, err := client.GetRawTransactionVerbose(h)
	if err != nil {
		fmt.Printf("Problems with getting Tx details: %s\nHash: %v\nVin: %v\n", h.String(), vin, err.Error())
		return err
	}

	out.Amount = tx.Vout[vin.Vout].Value
	out.Addresses = tx.Vout[vin.Vout].ScriptPubKey.Addresses

	if processAddresses {
		// Let's associate the address with Tx
		for _, addr := range out.Addresses {
			err = DbAddTxToAddress(db, addr, out)
			if err != nil {
				// TODO for performance reasons, unspent TXs that cannot be linked with address are ignored
				// fmt.Printf("Problem adding Tx to address, %v. Error: %s", out, err.Error())

			}
		}
	}

	out.Index = index

	details.Inputs = append(details.Inputs, out)
	return nil
}

//func processTxVout2(details *db.Transaction, vout btcjson.Vout, index int) error {
//
//	//for _, e := range vout.ScriptPubKey.Addresses {
//	//	out.Addresses = append(out.Addresses, db.Address{Hash: e})
//	//}
//
//	details.Outputs = append(details.Outputs, db.TxOutput{
//		IsCoinbase: false,
//		TxHash:     details.Hash,
//		Amount:     vout.Value,
//		TxType:     vout.ScriptPubKey.Type,
//		Index:      index,
//	})
//	return nil
//}

func processTxVout(details *TxDetails, vout btcjson.Vout, index int) error {
	out := TxOutput{}

	out.IsCoinbase = false
	out.TxHash = details.Hash
	out.Amount = vout.Value
	out.Addresses = vout.ScriptPubKey.Addresses
	out.TxType = vout.ScriptPubKey.Type
	out.Index = index

	details.Outputs = append(details.Outputs, out)
	return nil
}

// ProcessBlock is traversing the blockchain backwards adding all unknown yet blocks.
// It stops on error or on first already known block
func ProcessBlock2(dgraph *dgo.Dgraph, startBlock *wire.MsgBlock, currentHash chainhash.Hash,
	nextBlock chainhash.Hash, blockId uint64, block *db.Block) error {
	txHashes, err := startBlock.TxHashes()
	if err != nil {
		fmt.Printf("we have problem with TxHashes() %s\n", err.Error())
	}

	//saveProcessingState(db, currentHash.String(), blockId)
	//DbIncrementBlockCount(db)

	block.Hash = currentHash.String()
	block.PrevBlock = &db.Block{
		Hash: startBlock.Header.PrevBlock.String(),
	}
	//block.NextBlock = &db.Block{Hash: nextBlock.String()}
	block.Timestamp = startBlock.Header.Timestamp.Format(time.RFC3339)
	block.Id = blockId

	for _, tx := range txHashes {
		block.Transactions = append(block.Transactions, &db.Transaction{Hash: tx.String()})
	}

	_, err = db.UpdateBlock(dgraph, block)
	if err != nil {
		fmt.Printf("Saving block gave error %s\n", err.Error())
		return err
	}

	return nil
}

// ProcessBlock is traversing the blockchain backwards adding all unknown yet blocks.
// It stops on error or on first already known block
func ProcessBlock(db *badger.DB, startBlock *wire.MsgBlock, currentHash chainhash.Hash,
	nextBlock chainhash.Hash, blockId uint64, block *Block) error {
	txHashes, err := startBlock.TxHashes()
	if err != nil {
		fmt.Printf("we have problem with TxHashes() %s\n", err.Error())
	}

	saveProcessingState(db, currentHash.String(), blockId)
	DbIncrementBlockCount(db)

	block.Hash = currentHash
	block.PrevBlockHash = startBlock.Header.PrevBlock
	block.NextBlockHash = nextBlock
	block.Timestamp = startBlock.Header.Timestamp
	block.Id = blockId

	var txHashStrings []string

	for _, tx := range txHashes {
		txHashStrings = append(txHashStrings, tx.String())
	}
	block.TxHashes = txHashStrings

	err = DbSetBlock(db, *block)
	if err != nil {
		fmt.Printf("Saving block gave error %s\n", err.Error())
		return err
	}

	return nil
}

// ProcessNewBlocks process all the new blocks from a given hash down to the block that is already in DB
func ProcessNewBlocks2(dgraphDb *dgo.Dgraph,
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
	startHash, err := chainhash.NewHashFromStr(blockHash)
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
		return err
	}

	var lastBlockHash *chainhash.Hash

	// We will handle CTRL-C and CTRL-Z nicely
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Main loop
mainLoop:
	for {
		select {
		case <-c:
			fmt.Printf("\n### Block processing interrupted\n\nLast processed block #%d - %s\n\n",
				startingBlockId, lastBlockHash)
			break mainLoop
		default:
			// we do nothing
		}

		block := db.Block{}

		err := db.GetCompleteBlock(dgraphDb, blockHash, &block)
		if err == nil {
			// we already have this block, we need to update the NextBlockHash
			//if lastBlockHash != nil {
			//	block.NextBlockHash = *lastBlockHash
			//	err = DbSetBlock(dgraphDb, block)
			//	if err != nil {
			//		fmt.Printf("Error saving last known Block with updated NextBlockHash")
			//	}
			//}
			break
		}

		startBlock, err := client.GetBlock(startHash)
		if lastBlockHash == nil {
			lastBlockHash = &chainhash.Hash{} // placeholder for the last block
		}

		if err != nil {
			fmt.Printf("Problem with getBlock() %s\n", err.Error())
			break
		}
		err = ProcessBlock2(dgraphDb, startBlock, *startHash, *lastBlockHash, startingBlockId, &block)
		if err != nil {
			fmt.Printf("Error: we had problem processing the block\n%v\n", block)
			fmt.Printf("Hash: %v, BlockId: %d\n", *startHash, startingBlockId)
			break
		}

		startHash, err = chainhash.NewHashFromStr(block.PrevBlock.Hash)
		if err != nil {
			fmt.Printf("Error: we had problem converting hash to chainhash\n%v\n", block.PrevBlock.Hash)
			break
		}

		if startingBlockId == 0 || block.Id == stoppingBlockId {
			break
		}
		startingBlockId--
		lastBlockHash, err = chainhash.NewHashFromStr(block.Hash)
		if err != nil {
			fmt.Printf("Error: we had problem converting hash to chainhash\n%v\n", block.Hash)
			break
		}
		blockHash = startHash.String()

		blkCounter++
		if blkCounter%20000 == 0 {
			fmt.Printf("%v * 20k blocks done\n", blkCounter/20000)
		}
	}

	fmt.Printf("Processed in total: %v blocks\n", blkCounter)

	blockHash = startingBlockHash

	for {
		block := db.Block{}
		err := db.GetBlock(dgraphDb, blockHash, &block)
		if err != nil {
			fmt.Printf("DbGetBlock() failed. blkcount: %v, txcount: %v\n", blkCounter, txCounter)
			fmt.Printf("Block: %v\n", block)
			break
		}

		txs := block.Transactions
		for _, t := range txs {
			err = ProcessTx2(dgraphDb, client, processAddresses, t.Hash, block.Uid)
			if err != nil {
				fmt.Printf("DbGetBlock() failed in tx traversal. blkcount: %v, txcount: %v\n", blkCounter, txCounter)
				fmt.Printf("Error: %s\n", err.Error())
				fmt.Printf("Tx: %v\n", t)
				break
			}
			txCounter++
			//DbIncrementGlobalTxCount(db)

			if txCounter%5000 == 0 {
				fmt.Printf("%v * 5k TXs done. BlockId: %v, %v\n", txCounter/5000, block.Id, block.Hash)
				fmt.Printf("Block %d processed, tx count: %d\n", block.Id, txCounter)
			}
		}

		blockHash = block.PrevBlock.Hash
		//saveProcessingState(db, blockHash, block.Id-1)
		if blockHash == "0000000000000000000000000000000000000000000000000000000000000000" ||
			block.Id == stoppingBlockId {
			//saveProcessingStateFinished(db)
			break
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

// ProcessNewBlocks process all the new blocks from a given hash down to the block that is already in DB
func ProcessNewBlocks(db *badger.DB,
	client *rpcclient.Client,
	processAddresses bool,
	startingBlockHash string,
	startingBlockId uint64,
	stoppingBlockId uint64) error {

	timerStart := time.Now()
	DbSetStatus(db, DbBlockStatusProcessing)
	err := DbSetUint64(db, DbBlockStopBlockId, stoppingBlockId)
	if err != nil {
		fmt.Printf("Error: failed to save stopBlockID\n%v\n", err)
	}

	blkCounter := 0
	txCounter := 0

	blockHash := startingBlockHash
	startHash, err := chainhash.NewHashFromStr(blockHash)
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
		return err
	}

	var lastBlockHash *chainhash.Hash

	// We will handle CTRL-C and CTRL-Z nicely
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Main loop
mainLoop:
	for {
		select {
		case <-c:
			fmt.Printf("\n### Block processing interrupted\n\nLast processed block #%d - %s\n\n",
				startingBlockId, lastBlockHash)
			break mainLoop
		default:
			// we do nothing
		}
		block := Block{}
		err := DbGetBlock(db, blockHash, &block)
		if err == nil {
			// we already have this block, we need to update the NextBlockHash
			if lastBlockHash != nil {
				block.NextBlockHash = *lastBlockHash
				err = DbSetBlock(db, block)
				if err != nil {
					fmt.Printf("Error saving last known Block with updated NextBlockHash")
				}
			}
			break
		}

		startBlock, err := client.GetBlock(startHash)
		if lastBlockHash == nil {
			lastBlockHash = &chainhash.Hash{} // placeholder for the last block
		}

		if err != nil {
			fmt.Printf("Problem with getBlock() %s\n", err.Error())
			break
		}
		err = ProcessBlock(db, startBlock, *startHash, *lastBlockHash, startingBlockId, &block)
		if err != nil {
			fmt.Printf("Error: we had problem processing the block\n%v\n", block)
			fmt.Printf("Hash: %v, BlockId: %d\n", *startHash, startingBlockId)
			break
		}

		startHash = &block.PrevBlockHash
		if startingBlockId == 0 || block.Id == stoppingBlockId {
			break
		}
		startingBlockId--
		lastBlockHash, _ = chainhash.NewHashFromStr(block.Hash.String())
		blockHash = startHash.String()

		blkCounter++
		if blkCounter%20000 == 0 {
			fmt.Printf("%v * 20k blocks done\n", blkCounter/20000)
		}
	}
	fmt.Printf("Processed in total: %v blocks\n", blkCounter)

	blockHash = startingBlockHash

	for {
		block := Block{}
		err := DbGetBlock(db, blockHash, &block)
		if err != nil {
			fmt.Printf("DbGetBlock() failed. blkcount: %v, txcount: %v\n", blkCounter, txCounter)
			fmt.Printf("Block: %v\n", block)
			break
		}

		txs := block.TxHashes
		for _, t := range txs {
			err = ProcessTx(db, client, processAddresses, t)
			if err != nil {
				fmt.Printf("DbGetBlock() failed in tx traversal. blkcount: %v, txcount: %v\n", blkCounter, txCounter)
				fmt.Printf("Error: %s\n", err.Error())
				fmt.Printf("Tx: %v\n", t)
				break
			}
			txCounter++
			DbIncrementGlobalTxCount(db)

			if txCounter%5000 == 0 {
				fmt.Printf("%v * 5k TXs done. BlockId: %v, %v\n", txCounter/5000, block.Id, block.Hash)
				fmt.Printf("Block %d processed, tx count: %d\n", block.Id, txCounter)
			}
		}

		blockHash = block.PrevBlockHash.String()
		saveProcessingState(db, blockHash, block.Id-1)
		if blockHash == "0000000000000000000000000000000000000000000000000000000000000000" ||
			block.Id == stoppingBlockId {
			saveProcessingStateFinished(db)
			break
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

// resetting the processing state to default values
func saveProcessingStateFinished(db *badger.DB) {
	DbSetStatus(db, DbBlockStatusFinished)
}
