package dashrpc

import (
	"dashrpc/btcjson"
	"dashrpc/rpcclient"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/dgraph-io/badger"
)

func CreateTxDetails(client *rpcclient.Client, hash chainhash.Hash) (*TxDetails, error) {
	return nil, nil
}


func GetTxDetails() {

}

func processTxVin(client *rpcclient.Client, vin btcjson.Vin, details *TxDetails) error {
	h, err := chainhash.NewHashFromStr(vin.Txid)
	if err != nil {
		fmt.Printf("Problems with converting str to hash in showTxVinDetails: %s", err.Error())
		return err
	}

	tx, err := client.GetRawTransactionVerbose(h)
	if err != nil {
		fmt.Printf("Problems with getting Tx details: %s", err.Error())
		return err
	}
	//t := TxTransfer{}
	if vin.IsCoinBase() {

		return nil
	}
	value := tx.Vout[vin.Vout].Value
	addrs := tx.Vout[vin.Vout].ScriptPubKey.Addresses
	fmt.Printf("Value: %v | Addrs: %v | Coinbase: %s | IsCoinbase: %v | Sequence: %v | Txid: %v | HasWitness: %v |\n",
		value, addrs, vin.Coinbase, vin.IsCoinBase(), vin.Sequence, vin.Txid, vin.HasWitness())

	return nil
}

func ProcessBlock(db *badger.DB, startBlock *wire.MsgBlock, nextBlock string) (*Block,
	error) {
	txHashes, err := startBlock.TxHashes()
	if err != nil {
		fmt.Printf("we have problem with TxHashes() %s\n", err.Error())
	}

	block := Block{}
	block.Hash = startBlock.BlockHash().String()
	block.PrevBlockHash = startBlock.Header.PrevBlock.String()
	block.NextBlockHash = nextBlock

	var txHashStrings []string

	for _, tx := range txHashes {
		txHashStrings = append(txHashStrings, tx.String())
	}
	block.TxHashes = txHashStrings

	err = DbSetBlock(db, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil

	/*
	for _, v := range txHashes {
		tx, err := client.GetRawTransactionVerbose(&v)
		//tx1, err := client.GetRawTransaction(&v)
		if err != nil {
			fmt.Printf("Problems getting the tx from hash %v,  %s\n", v, err.Error())
		}
		fmt.Printf("DETAILS TXhash: %s | TXid: %v | TX vsize: %v\n", tx.Hash, tx.Txid, tx.Vsize)
		txDetails := TxDetails{}
		for _, d := range tx.Vin {
			err := processTxVin(client, d, &txDetails)
			if err != nil {
				fmt.Printf("Problems with processTxVin() call in ProcessBlock(): %s", err.Error())
			}
		}
	}
*/
}
