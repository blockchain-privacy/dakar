package dashrpc

import (
	"dashrpc/rpcclient"
	"fmt"
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
		"093d3a6b58027ad6c421a917da14686ee0132279a3693d1e82bd39f64d20357f",
		"5f5d145e89eb6e61060d20c606955e721601444ab09638eaacebb2208f10e279",
		"6ebd6c0895bbda1c9c6f5ccdb157fb334bf3b18b0df66ca2d5c1236584c1f065",
		"14c95bd2290b1d3b5010bfe71c2e462b5c2d704f90345fb5fee645f2215d7f45",
	}

	nonMixingTxs := []string{
		"39ff330b04d5ff92621263af5f2ffce197f6acba92beb5bf36d145da60891e53",
		"5ecd1e968d05b1d459ab660638043195c0f226c0dde2a87c0a86889890afa952",
		"ddb64c54bc090bcac7f7caa524691bcd390ef1814339d6ca9e95e117c3342cfe",
		"e40683aad63bc5ee70a99644ce98a836f2a32eb77f64b107232f06a018d8253a",
		"559a2be1df28485dbc733362806f63920e5feb0209cac7018e77f85ba5a5327d",
		"fd946e63e4fad17da110a54c0d2287b56142d2f56cf1ac264e773345d06b0491",
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
		if !tx.IsMixing() {
			fmt.Println(tx)
			t.Error("Mixing transaction not recognized\nFalse Negative")
		}
	}

	for _, h := range nonMixingTxs {
		DbGetTxDetails(db, h, &tx)
		if tx.IsMixing() {
			fmt.Println(tx)
			t.Error("Non-mixing transaction not recognized\nFalse Positive")
		}
	}
}
