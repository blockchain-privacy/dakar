package dashrpc

import (
	"dashrpc/rpcclient"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/dgraph-io/badger"
	"log"
	"testing"
)

const block49998 = "000000000018692f3cd1e6255d9aa3edc427101e02da940f6e6673823118f016"
const block49999 = "000000000014f796bbd2312686a63cbe17401a1026ab2a8149b74553e8dcb96d"
const block50000 = "00000000000fa6230896498b3cc6f1015456b4512452ead9979f6b43ca0a74dc"

func TestProcessBlock50000(t *testing.T) {
	// Setup the Badger DB connection
	opts := badger.DefaultOptions
	opts.NumVersionsToKeep = 0
	opts.SyncWrites = false
	opts.ValueDir = "/tmp/testDb"
	opts.Dir = "/tmp/testDb"
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       "0.0.0.0:9998",
		User:       "rpc1user",
		Pass:       "1234pass",
		DisableTLS: true,
	}
	client, err := rpcclient.New(&conn)
	if err != nil {
		fmt.Printf("we have a problem: %s\n", err.Error())
	}

	block := Block{}

	defer func(){
		// cleanup
		err = db.Update(func(txn *badger.Txn) error {
			return txn.Delete([]byte(block50000))
		})
		if err != nil {
			t.Error(err)
		}
	}()

	startHash, err := chainhash.NewHashFromStr(block50000)
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
	}
	startBlock, err := client.GetBlock(startHash)
	if err != nil {
		fmt.Printf("we have problem with getBlock() %s\n", err.Error())
		t.Error(t)
	}
	err = ProcessBlock(db, startBlock, *startHash, chainhash.Hash{}, &block)
	if err != nil {
		t.Error(err)
	}

	block2 := Block{}
	err = DbGetBlock(db, block50000, &block2)
	if err != nil {
		t.Error(err)
	}

	if block2.Hash != block.Hash ||
		block2.NextBlockHash != block.NextBlockHash ||
		block2.PrevBlockHash != block.PrevBlockHash ||
		len(block2.TxHashes) != len(block.TxHashes) ||
		block2.TxHashes[0] != block.TxHashes[0] {
		t.Error("Blocks do not match")
	}

	if block2.PrevBlockHash.String() != block49999 ||
		block.PrevBlockHash.String() != block49999 {
		msg := fmt.Sprintf("PrevBlockHash does not match!\n%s\n%s\n%s\n",
			block49999, block.PrevBlockHash, block2.PrevBlockHash)
		t.Error(msg)
	}

	if block2.TxHashes[0] != "c13fc482603f574b7322da10398c20d64a431e14f8e886b054128591abaa66a4" {
		t.Error("Output Transaction hash is WRONG")
	}

}

func TestProcessBlock49999(t *testing.T) {
	// Setup the Badger DB connection
	opts := badger.DefaultOptions
	opts.NumVersionsToKeep = 0
	opts.SyncWrites = false
	opts.ValueDir = "/tmp/testDb"
	opts.Dir = "/tmp/testDb"
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       "0.0.0.0:9998",
		User:       "rpc1user",
		Pass:       "1234pass",
		DisableTLS: true,
	}
	client, err := rpcclient.New(&conn)
	if err != nil {
		fmt.Printf("we have a problem: %s\n", err.Error())
	}

	block := Block{}

	defer func(){
		// cleanup
		err = db.Update(func(txn *badger.Txn) error {
			return txn.Delete([]byte(block49999))
		})
		if err != nil {
			t.Error(err)
		}
	}()

	startHash, err := chainhash.NewHashFromStr(block49999)
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
	}
	block50000hash, err := chainhash.NewHashFromStr(block50000)
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
	}

	startBlock, err := client.GetBlock(startHash)
	if err != nil {
		fmt.Printf("we have problem with getBlock() %s\n", err.Error())
		t.Error(t)
	}
	err = ProcessBlock(db, startBlock, *startHash, *block50000hash, &block)
	if err != nil {
		t.Error(err)
	}

	block2 := Block{}
	err = DbGetBlock(db, block49999, &block2)
	if err != nil {
		t.Error(err)
	}

	if block2.Hash != block.Hash ||
		block2.NextBlockHash != block.NextBlockHash ||
		block2.PrevBlockHash != block.PrevBlockHash ||
		len(block2.TxHashes) != len(block.TxHashes) ||
		len(block2.TxHashes) != 9 ||
		block2.TxHashes[0] != block.TxHashes[0] {
		t.Error("Blocks do not match")
	}

	if block2.PrevBlockHash.String() != block49998 ||
		block.NextBlockHash.String() != block50000 {
		msg := fmt.Sprintf("PrevBlockHash does not match!\n%s\n%s\n%s\n",
			block49999, block.PrevBlockHash, block2.PrevBlockHash)
		t.Error(msg)
	}

	if block2.TxHashes[0] != "106f0dea7bdff518a5db6551dd43210d6639fffad84e56083e73231921c779f9" {
		fmt.Printf("TxHashes %v\n", block2.TxHashes)
		t.Error("Output Transaction hash is WRONG")
	}

}


func TestProcessTxFromBlock50000(t *testing.T) {
	// Setup the Badger DB connection
	opts := badger.DefaultOptions
	opts.NumVersionsToKeep = 0
	opts.SyncWrites = false
	opts.ValueDir = "/tmp/testDb"
	opts.Dir = "/tmp/testDb"
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       "0.0.0.0:9998",
		User:       "rpc1user",
		Pass:       "1234pass",
		DisableTLS: true,
	}
	client, err := rpcclient.New(&conn)
	if err != nil {
		fmt.Printf("we have a problem: %s\n", err.Error())
	}

	block := Block{}
	txHash := "c13fc482603f574b7322da10398c20d64a431e14f8e886b054128591abaa66a4"

	defer func(){
		// cleanup
		err = db.Update(func(txn *badger.Txn) error {
			e1 := txn.Delete([]byte(block50000))
			if e1 != nil {
				return e1
			}
			return txn.Delete([]byte(txHash))
		})

		if err != nil {
			t.Error(err)
		}
	}()

	startHash, err := chainhash.NewHashFromStr(block50000)
	if err != nil {
		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
	}
	startBlock, err := client.GetBlock(startHash)
	if err != nil {
		fmt.Printf("we have problem with getBlock() %s\n", err.Error())
		t.Error(t)
	}
	err = ProcessBlock(db, startBlock, *startHash, chainhash.Hash{}, &block)
	if err != nil {
		t.Fatal(err)
	}

	err = ProcessTx(db, client, txHash)
	if err != nil {
		t.Fatal(err)
	}

	txDetails := TxDetails{}
	err = DbGetTxDetails(db, block.TxHashes[0], &txDetails)
	if err != nil {
		t.Error(err)
	}
	if len(txDetails.Outputs) != 1 ||
		txDetails.Outputs[0].Amount != 16.00 ||
		txDetails.Outputs[0].IsCoinbase != false ||
		len(txDetails.Inputs) != 1 ||
		txDetails.Inputs[0].IsCoinbase != true {
		msg := fmt.Sprintf("Error: TX data does not match for TX c13fc48....\nData:\n%v\n", txDetails)
		t.Error(msg)
	}

	if err != nil {
		t.Error(err)
	}

}

func TestProcessTxFromBlock49999(t *testing.T) {
	// Setup the Badger DB connection
	opts := badger.DefaultOptions
	opts.NumVersionsToKeep = 0
	opts.SyncWrites = false
	opts.ValueDir = "/tmp/testDb"
	opts.Dir = "/tmp/testDb"
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       "0.0.0.0:9998",
		User:       "rpc1user",
		Pass:       "1234pass",
		DisableTLS: true,
	}
	client, err := rpcclient.New(&conn)
	if err != nil {
		fmt.Printf("we have a problem: %s\n", err.Error())
	}

	txHash := "af530c23992d7439107b31d8840facb60d0606d370c9cdd35195eea87113ff1e"

	defer func(){
		// cleanup
		err = db.Update(func(txn *badger.Txn) error {
			return txn.Delete([]byte(txHash))
		})

		if err != nil {
			t.Error(err)
		}
	}()

	err = ProcessTx(db, client, txHash)
	if err != nil {
		t.Fatal(err)
	}

	txDetails := TxDetails{}
	err = DbGetTxDetails(db, txHash, &txDetails)
	if err != nil {
		t.Error(err)
	}
	if len(txDetails.Outputs) != 2 ||
		len(txDetails.Inputs) != 2 ||
		txDetails.Outputs[0].Amount != 9.91547479 ||
		txDetails.Outputs[1].Amount != 6.02335110 ||
		len(txDetails.Outputs[0].Addresses) != 1 ||
		len(txDetails.Outputs[1].Addresses) != 1 ||
		txDetails.Outputs[0].Addresses[0] != "XrHBvi9hxQcrUfXsB9hK6V7hb2625s2kAV" ||
		txDetails.Outputs[1].Addresses[0] != "Xstz9D2DNrrCWhAnsmiu1R144DesKNw22t" ||
		txDetails.Outputs[0].IsCoinbase != false ||
		txDetails.Outputs[1].IsCoinbase != false ||
		txDetails.Outputs[0].Index != 0 ||
		txDetails.Outputs[1].Index != 1 ||
		len(txDetails.Inputs[0].Addresses) != 1 ||
		len(txDetails.Inputs[1].Addresses) != 1 ||
		txDetails.Inputs[0].Amount != 7.73616759 ||
		txDetails.Inputs[1].Amount != 8.20365830 ||
		txDetails.Inputs[0].Addresses[0] != "XooKzX2FFWZekaVg7X8T67oLWE2v1tpX5z" ||
		txDetails.Inputs[1].Addresses[0] != "XnLNnQVYQ9P2zc6uQrY5vypLmXoTiqxrw7" ||
		txDetails.Inputs[0].IsCoinbase != false ||
		txDetails.Inputs[1].IsCoinbase != false {
		msg := fmt.Sprintf("Error: TX data does not match for TX af530c23992d74....\nData:\n%v\n", txDetails)
		t.Error(msg)
	}

	if err != nil {
		t.Error(err)
	}

}