package main

import (
	"dashrpc/rpcclient"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func main() {
	fmt.Println("Go RPC client")
	fmt.Println("Current block...")

	var conn = rpcclient.ConnConfig{
		Host:       "0.0.0.0:9998",
		User:       "rpc1user",
		Pass:       "1234pass",
		DisableTLS: true,
	}
	var client, err = rpcclient.New(&conn)
	if err != nil {
		fmt.Printf("we have a problem: %s\n", err.Error())
	}

	count, err := client.GetBlockCount()
	if err != nil {
		fmt.Printf("we have problem with count() %s\n", err.Error())
	}
	fmt.Printf("Current block count is: %v\n", count)

	startHash, err := chainhash.NewHashFromStr("00000e2612da1081bbde1fc5202b1f467e90aba904969ce10a4253f0dcec8389")
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
	}

	startBlock, err := client.GetBlock(startHash)
	if err != nil {
		fmt.Printf("we have problem with getBlock() %s\n", err.Error())
	}

	txHashes, err := startBlock.TxHashes()
	if err != nil {
		fmt.Printf("we have problem with TxHashes() %s\n", err.Error())
	}

	fmt.Printf(" The start block has %v txs\n", len(txHashes))
	for _, v := range txHashes {
		tx, err := client.GetTransaction(&v)
		if err != nil { fmt.Printf("Problems getting the tx from hash %v,  %s\n", v, err.Error()) }
		fmt.Printf("DETAILS\n")
		for _, d := range tx.Details {
			fmt.Printf("\nAddress: %s\nAccount: %s\nAmount: %v\nCategory: %s\nFee: %v\nVout: %v\n\n", d.Address,
				d.Account,
				d.Amount,
				d.Category, d.Fee, d.Vout)
		}
	}
}
