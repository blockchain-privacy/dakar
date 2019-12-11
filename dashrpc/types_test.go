package dashrpc

import (
	"dashrpc/rpcclient"
	"github.com/dgraph-io/badger"
	"testing"
)

func setupClients(t *testing.T) (*badger.DB, *rpcclient.Client) {
	// Setup Badger
	// Setup the Badger DB connection
	opts := badger.DefaultOptions("/tmp/testDb")
	opts.WithNumVersionsToKeep(0)
	opts.WithSyncWrites(false)
	opts.WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		t.Error(err)
		return db, nil
	}
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
	return db, client
}

func TestMixingTransactions(t *testing.T) {
	block850000 := "000000000000003053911f63ae6c1fe8a8872c43127211977e6efb4a3f621dcd"
	mixingTxs := []string{
		"9bcb21191d918223e0d2428a0201f2d0ed6779693767d0d78f448fcbff7bd789",
	}

	nonMixingTxs := []string{

	}

	db, client := setupClients(t)
	err := ProcessNewBlocks(db, client, block850000, 850000, 849995)
	if err != nil {
		t.Error(err)
		return
	}
	tx := TxDetails{}
	for _, h := range mixingTxs {
		DbGetTxDetails(db, h, &tx)
		if !tx.IsMixingTx() { t.Error("Mixing transaction not recognized\nFalse Negative") }
	}

	for _, h := range nonMixingTxs {
		DbGetTxDetails(db, h, &tx)
		if tx.IsMixingTx() { t.Error("Non-mixing transaction not recognized\nFalse Positive") }
	}
}
