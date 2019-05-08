package dashrpc

import (
	"dashrpc/btcjson"
	"dashrpc/rpcclient"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/dgraph-io/badger"
)

// ProcessTx process transaction, and the Vout and Vin records
func ProcessTx(db *badger.DB, client *rpcclient.Client, txHashString string) error {
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
		err := processTxVin(db, client, &txDetails, d, index)
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

func processTxVin(db *badger.DB, client *rpcclient.Client, details *TxDetails, vin btcjson.Vin, index int) error {
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

	out.IsCoinbase = vin.IsCoinBase()
	out.TxHash = vin.Txid
	out.Amount = tx.Vout[vin.Vout].Value
	out.Addresses = tx.Vout[vin.Vout].ScriptPubKey.Addresses

	// Let's associate the address with Tx
	for _, addr := range out.Addresses {
		err = DbAddTxToAddress(db, addr, out)
		if err != nil {
			fmt.Printf("Problem adding Tx to address, %v. Error: %s", out, err.Error())
		}
	}
	out.Index = index

	details.Inputs = append(details.Inputs, out)
	return nil
}

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
func ProcessBlock(db *badger.DB, startBlock *wire.MsgBlock, currentHash chainhash.Hash,
	nextBlock chainhash.Hash, block *Block) error {
	txHashes, err := startBlock.TxHashes()
	if err != nil {
		fmt.Printf("we have problem with TxHashes() %s\n", err.Error())
	}

	block.Hash = currentHash
	block.PrevBlockHash = startBlock.Header.PrevBlock
	block.NextBlockHash = nextBlock

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
func ProcessNewBlocks(db *badger.DB, client *rpcclient.Client, startingBlockHash string) {
	blkCounter := 0
	txCounter := 0

	blockHash := startingBlockHash
	startHash, err := chainhash.NewHashFromStr(blockHash)
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
	}

	var lastBlockHash chainhash.Hash
	// Main loop

	for {
		block := Block{}
		err := DbGetBlock(db, blockHash, &block)
		if err == nil {
			// we have processed the block already, we are done.
			break
		}

		startBlock, err := client.GetBlock(startHash)
		if err != nil {
			fmt.Printf("we have problem with getBlock() %s\n", err.Error())
			break
		}
		err = ProcessBlock(db, startBlock, *startHash, lastBlockHash, &block)
		startHash = &block.PrevBlockHash
		lastBlockHash = block.Hash
		startHash, err = chainhash.NewHashFromStr(blockHash)
		if err != nil {
			fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
		}
		blockHash = startHash.String()
		fmt.Printf("Current %v blockhash: %s\n", blkCounter, blockHash)

		blkCounter++
		if blkCounter % 20000 == 0 {
			fmt.Printf("%v * 20k blocks done\n", (blkCounter / 20000))
		}
	}
	fmt.Printf("Processed in total: %v blocks\n", blkCounter)

	currentBlockHash := startingBlockHash

	for {
		block := Block{}
		err := DbGetBlock(db, currentBlockHash, &block)
		if err != nil {
			fmt.Printf("DbGetBlock() failed in tx traversal. blkcount: %v, txcount: %v\n", blkCounter, txCounter)

			break
		}

		txs := block.TxHashes
		for _, t := range txs {
			err = ProcessTx(db, client, t)
			if err != nil {
				fmt.Printf("DbGetBlock() failed in tx traversal. blkcount: %v, txcount: %v\n", blkCounter, txCounter)
				fmt.Printf("Error: %s\n", err.Error())
				fmt.Println("Continuing....")
			}
			txCounter++
		}

		currentBlockHash = block.PrevBlockHash.String()
	}
}
