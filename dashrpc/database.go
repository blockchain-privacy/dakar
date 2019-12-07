package dashrpc

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

//
// Database is a simple key-value system
//
// hash -> TxDetails
// hash -> Block
// address -> []TxOutput
//
//

// DbGetItem checks if a given hash is stored in the k-v store
func DbGetItem(db *badger.DB, hash string, item *ChainItem) error {
	return db.View(func(txn *badger.Txn) error {
		b, e := txn.Get([]byte(hash))
		if e != nil {
			return e
		}
		return b.Value(func(buf []byte) error {
			dec := gob.NewDecoder(bytes.NewReader(buf))
			e2 := dec.Decode(item)
			return e2
		})
	})
}

// DbWriteItem writes chain item into k-v store
func DbSetItem(db *badger.DB, hash string, data []byte, itemType string) error {
	item := ChainItem{}
	item.ItemType = itemType
	item.Data = data
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(item)
	if err != nil {
		return err
	}

	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(hash), buf.Bytes())
	})
}

// DbGetBlock checks if a given hash is stored in the k-v store as Block
func DbGetBlock(db *badger.DB, hash string, block *Block) error {
	item := ChainItem{}
	err := DbGetItem(db, hash, &item)
	if err != nil {
		return err
	}
	if item.ItemType != ChainType_Block {
		m := fmt.Sprintf("Item's ChainType mismatch. Expected: blk, got: %s", item.ItemType)
		err = errors.New(m)
		return err
	}
	dec := gob.NewDecoder(bytes.NewReader(item.Data))
	return dec.Decode(block)
}

// DbSetBlock sets a given block into the k-v store as blk chainItem type.
func DbSetBlock(db *badger.DB, block Block) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(block)
	if err != nil {
		return err
	}

	return DbSetItem(db, block.Hash.String(), buf.Bytes(), ChainType_Block)
}

// DbGetTxDetails checks if a given hash is stored in the k-v store as TxDetails
func DbGetTxDetails(db *badger.DB, hash string, txDetails *TxDetails) error {
	item := ChainItem{}
	err := DbGetItem(db, hash, &item)
	if err != nil {
		return err
	}
	if item.ItemType != ChainType_TxDetails {
		m := fmt.Sprintf("Item's ChainType mismatch. Expected: txd, got: %s", item.ItemType)
		err = errors.New(m)
		return err
	}
	dec := gob.NewDecoder(bytes.NewReader(item.Data))
	return dec.Decode(txDetails)
}

// DbSetTxDetails sets a given block into the k-v store as txd chainItem type.
func DbSetTxDetails(db *badger.DB, txDetails TxDetails) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(txDetails)
	if err != nil {
		return err
	}
	return DbSetItem(db, txDetails.Hash, buf.Bytes(), ChainType_TxDetails)
}

func DbAddTxToAddress(db *badger.DB, addr string, output TxOutput) error {
	var buf bytes.Buffer
	outSlice := make([]TxOutput, 2)
	err := DbGetTxosForAddress(db, addr, &outSlice)
	if err != nil {
		// we do not have the slice yet
		// we should log the error, but, it might be that we just do not have that addr yet
	}
	outSlice = append(outSlice, output)
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(outSlice)
	if err != nil {
		return err
	}
	return DbSetItem(db, addr, buf.Bytes(), ChainType_AddrOutputs)
}

func DbGetTxosForAddress(db *badger.DB, addr string, txOutputs *[]TxOutput) error {
	item := ChainItem{}
	err := DbGetItem(db, addr, &item)
	if err != nil {
		return err
	}
	if item.ItemType != ChainType_AddrOutputs {
		m := fmt.Sprintf("Item's ChainType mismatch. Expected: tx4a, got: %s", item.ItemType)
		err = errors.New(m)
		return err
	}

	dec := gob.NewDecoder(bytes.NewReader(item.Data))
	err = dec.Decode(&txOutputs)
	if err != nil {
		return err
	}
	return nil
}
