package dashrpc

import (
	"dashrpc/rpcclient"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"os"
	"testing"

	"github.com/dgraph-io/badger"
)

func TestDbSetBlock(t *testing.T) {
	db := setupDB(t)
	defer tearDownDB(t, db)

	block := Block{}
	hash1 := "00000f106b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"
	hash2 := "00000f206b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"
	hash3 := "00000f306b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"

	hash, err := chainhash.NewHashFromStr(hash1)
	if err != nil {
		t.Error(err)
		return
	}
	block.Hash = *hash
	block.PrevBlockHash = chainhash.Hash{}
	block.NextBlockHash = chainhash.Hash{}
	block.TxHashes = []string{hash2, hash3}

	err = DbSetBlock(db, block)
	if err != nil {
		t.Error(err)
		return
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
	db := setupDB(t)
	defer tearDownDB(t, db)

	tx := TxDetails{}
	hash1 := "00000f106b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"

	tx.Hash = hash1
	tx.Timestamp = 42

	err := DbSetTxDetails(db, tx)
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
	db := setupDB(t)
	defer tearDownDB(t, db)

	tx := TxDetails{}
	hash1 := "00000f106b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"

	tx.Hash = hash1
	tx.Timestamp = 42
	o := TxOutput{}
	o.Index = 0
	o.Amount = 12.54
	o.Addresses = []string{"exampleAddress"}
	tx.Outputs = append(tx.Outputs, o)

	i := TxOutput{}
	i.Index = 1
	i.Amount = 42.42
	i.Addresses = []string{"exampleInputAddr", "another one"}
	tx.Inputs = append(tx.Inputs, i)

	err := DbSetTxDetails(db, tx)
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

func TestDbSetString(t *testing.T) {
	db := setupDB(t)
	defer tearDownDB(t, db)

	data := map[string]string{
		"key": "value",
		"status": "fantastic",
		DbBlockStatus: DbBlockStatusUnknown,
		"something": DbBlockStatusProcessing,
	}

	for k,v := range data {
		err := DbSetString(db, k, v)
		if err != nil {
			t.Error(err)
		}
		var result string
		err = DbGetString(db, k, &result)
		if err != nil {
			t.Error(err)
		}
		if result != v {
			t.Error("saved and restored values do not match")
		}
	}
}

func TestDbSetUint64(t *testing.T) {
	db := setupDB(t)
	defer tearDownDB(t, db)

	data := map[string]uint64{
		"key": 1,
		"status": 0,
		DbBlockStatus: 676767,
		"something": 42424242,
	}

	for k,v := range data {
		err := DbSetUint64(db, k, v)
		if err != nil {
			t.Error(err)
		}
		var result uint64
		err = DbGetUint64(db, k, &result)
		if err != nil {
			t.Error(err)
		}
		if result != v {
			t.Error("saved and restored values do not match")
		}
	}
}

const TestDbPath = "/tmp/testDb_321"
// setupDB -- set the Badger DB up
func setupDB(t *testing.T) *badger.DB {
	// Setup the Badger DB connection
	opts := badger.DefaultOptions(TestDbPath)
	opts.WithNumVersionsToKeep(0)
	opts.WithSyncWrites(false)
	opts.WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		t.Error(err)
	}
	return db
}

// tearDownDB -- close the DB
func tearDownDB(t *testing.T, db *badger.DB) {
	err := db.Close()
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(TestDbPath)
	if err != nil {
		t.Error(err)
	}
}

// setupRpcClient -- setup client for testing
func setupRpcClient(t *testing.T) *rpcclient.Client {
	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       "0.0.0.0:9998",
		User:       "rpc1user",
		Pass:       "1234pass",
		DisableTLS: true,
	}
	client, err := rpcclient.New(&conn)
	if err != nil {
		t.Error(err)
	}
	return client
}
