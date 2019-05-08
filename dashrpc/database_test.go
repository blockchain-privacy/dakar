package dashrpc

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/dgraph-io/badger"
	"log"
	"testing"
)


func TestDbSetBlock(t *testing.T) {
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

	block := Block{}
	hash1 := "00000f106b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"
	hash2 := "00000f206b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"
	hash3 := "00000f306b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"

	hash, err := chainhash.NewHashFromStr(hash1)
	if err != nil {
		t.Error(err)
	}
	block.Hash = *hash
	block.PrevBlockHash = chainhash.Hash{}
	block.NextBlockHash = chainhash.Hash{}
	block.TxHashes = []string{hash2, hash3}

	err = DbSetBlock(db, block)
	if err != nil {
		t.Error(err)
	}

	block2 := Block{}
	err = DbGetBlock(db, hash1, &block2)
	if err != nil {
		t.Error(err)
	}

	if block2.Hash != block.Hash ||
		block2.NextBlockHash != block.NextBlockHash ||
		block2.PrevBlockHash != block.PrevBlockHash ||
		len(block2.TxHashes) != len(block.TxHashes) ||
		block2.TxHashes[0] != block.TxHashes[0] ||
		block2.TxHashes[1] != block.TxHashes[1] {
		t.Error("Blocks do not match")
	}

	// cleanup
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(hash1))
	})
	if err != nil {
		t.Error(err)
	}
}


func TestDbSetEmptyTxDetails(t *testing.T) {
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

	tx := TxDetails{}
	hash1 := "00000f106b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"

	tx.Hash = hash1
	tx.Timestamp = 42

	err = DbSetTxDetails(db, tx)
	if err != nil {
		t.Error(err)
	}

	tx2 := TxDetails{}
	err = DbGetTxDetails(db, hash1, &tx2)
	if err != nil {
		t.Error(err)
	}

	if tx2.Hash != tx.Hash ||
		len(tx2.Outputs) != len(tx.Outputs) ||
		len(tx2.Inputs) != len(tx.Inputs) ||
		tx2.Timestamp != tx.Timestamp {
		t.Error("TXs do not match")
	}

	// cleanup
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(hash1))
	})
	if err != nil {
		t.Error(err)
	}
}

func TestDbSetTxDetails(t *testing.T) {
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

	tx := TxDetails{}
	hash1 := "00000f106b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"

	tx.Hash = hash1
	tx.Timestamp = 42
	o := TxOutput{}
	o.Index = 0
	o.Amount = 12.54
	o.Addresses = []string{ "exampleAddress" }
	tx.Outputs = append(tx.Outputs, o)

	i := TxOutput{}
	i.Index = 1
	i.Amount = 42.42
	i.Addresses = []string{ "exampleInputAddr", "another one" }
	tx.Inputs = append(tx.Inputs, i)

	err = DbSetTxDetails(db, tx)
	if err != nil {
		t.Error(err)
	}

	tx2 := TxDetails{}
	err = DbGetTxDetails(db, hash1, &tx2)
	if err != nil {
		t.Error(err)
	}

	if tx2.Hash != tx.Hash ||
		len(tx2.Outputs) != len(tx.Outputs) ||
		len(tx2.Inputs) != len(tx.Inputs) ||
		tx2.Timestamp != tx.Timestamp ||
		tx2.Outputs[0].Amount != tx.Outputs[0].Amount ||
		tx2.Inputs[0].Amount != tx.Inputs[0].Amount ||
		tx2.Outputs[0].Index != tx.Outputs[0].Index ||
		tx2.Inputs[0].Index != tx.Inputs[0].Index ||
		tx2.Outputs[0].Addresses[0] != tx.Outputs[0].Addresses[0] ||
		tx2.Inputs[0].Addresses[0] != tx.Inputs[0].Addresses[0] {
		t.Error("TXs do not match")
	}

	// cleanup
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(hash1))
	})
	if err != nil {
		t.Error(err)
	}
}
