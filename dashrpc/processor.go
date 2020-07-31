package dashrpc

import (
	"dashrpc/btcjson"
	dbaddr "dashrpc/db/address"
	dbblk "dashrpc/db/block"
	dbop "dashrpc/db/output"
	dbstat "dashrpc/db/status"
	dbtx "dashrpc/db/transaction"
	"dashrpc/rpcclient"
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

	status, err := dbstat.Get(dgraph)
	if err != nil {
		return
	}

	if status.LastBlockId == nil {
		// last block id is not set -> we start at the beginning of the chain
		startId = 1
		return
	}

	startId = *status.LastBlockId

	return
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
			log.Printf("### Block processing interrupted ###")
			isInterrupt = true
			return
		case <-ticker.C:
			// todo instead of using a ticker use asyncblock call and wait for future channel
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

// processes all the new blocks from a given hash down to the block that is already in DB
func ProcessNewBlocks(dgraph *dgo.Dgraph, client *rpcclient.Client, continuous bool,
	startingBlockId uint64, stoppingBlockId uint64) error {

	currentBlockId, err := getStartingId(dgraph, continuous, startingBlockId)
	if err != nil {
		return err
	}

	currentHashObj, err := client.GetBlockHash(int64(currentBlockId))
	if err != nil {
		log.Println("problem converting startingBlockId", currentBlockId, err)
		return err
	}

	blkCounter := 0
	txCounter := 0
	currentBlockHash := currentHashObj.String()

	log.Println("Starting crawling at Id:", currentBlockId, "Hash:", currentBlockHash)

	// We will handle CTRL-C and CTRL-Z nicely
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	timerStart := time.Now()
	// Main loop

	firstLoop := true
	var currentBlock *btcjson.GetBlockVerboseResult
mainLoop:
	for {
		if continuous && !firstLoop {
			// set values for this round
			if currentBlock.NextHash == "" {
				log.Println("Waiting for next block. Current block id:", currentBlockId)
				var isInterrupt bool
				// can not used short hand declaration, because it would mask currentBlock in the outer scope
				currentBlock, isInterrupt, err = waitForNextBlock(client, c, currentHashObj)
				if err != nil {
					return err
				}

				if isInterrupt {
					break mainLoop
				}

				log.Println("Found next block. Current block id:", currentBlockId)
			}
		}

		// if not the first round set variables for this loop
		if !firstLoop {
			currentHashObj, err = chainhash.NewHashFromStr(currentBlock.NextHash)
			if err != nil {
				return err
			}

			currentBlockHash = currentBlock.NextHash
			currentBlockId++
		}

		// check for stop conditions if not stop
		if !continuous {
			// stoppingBlockId+1 <- +1 because we still need to process this round
			if currentBlockId == stoppingBlockId+1 || (currentBlock != nil && currentBlock.NextHash == "") {
				// finished
				break
			}
		}

		select {
		case <-c:
			log.Printf("### Block processing interrupted ###")
			break mainLoop
		default:
			// we do nothing
		}

		firstLoop = false
		// get block from RPC-Client
		currentBlock, err = client.GetBlockVerbose(currentHashObj)
		if err != nil {
			return err
		}

		if b, err := dbblk.GetBlock(dgraph, currentBlockHash); err == nil && b.IsComplete() {
			// block already in database
			continue
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
			//	log.Printf("%v * 5k TXs done. BlockId: %v, %v\n", txCounter/5000, currentBlockId, currentBlockHash)
			//	log.Printf("Block %s processed, tx count: %d\n", currentBlockId, txCounter)
			//}
		}

		// create new block
		ts := time.Unix(currentBlock.Time, 0).Format(time.RFC3339)
		if err = ProcessBlock(dgraph, transactions, currentBlockHash, currentBlockId, ts, currentBlock.PreviousHash); err != nil {
			log.Println("Error: we had problem processing the block")
			log.Printf("Hash: %s, BlockId: %d\n", currentBlockHash, currentBlockId)
			break
		}

		if err = processAddresses(dgraph, txMapping); err != nil {
			return err
		}

		// save processing state
		if err = dbstat.SetLastBlockId(dgraph, currentBlockId); err != nil {
			log.Printf("error saving CurrentBlockID state: %v\n", err)
			return err
		}

		blkCounter++

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
//func ProcessAddressClustering(db *badger.DB, startingAddr string) error {
//
//	addrData := AddressData{}
//	err := DbGetDataForAddress(db, startingAddr, &addrData)
//	if err != nil {
//		return err
//	}
//	txs := addrData.Txs
//
//	for _, tx := range txs {
//		log.Printf("TX %v -- hash: %v Amount: %f -- %v\n", tx.TxType, tx.TxHash, tx.Amount, tx.Addresses)
//	}
//
//	return nil
//}
