package dashrpc

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)


// DbGetItem checks if a given hash is stored in the k-v store
func DbGetItem(db *badger.DB, hash string) (item *ChainItem, err error) {
	err = db.View(func(txn *badger.Txn) error {
		b, e := txn.Get([]byte(hash))
		if e != nil { return e }
		buf, e := b.Value()
		if e != nil {return e}
		dec := gob.NewDecoder(bytes.NewReader(buf))
		e = dec.Decode(item)
		return e
	})

	return
}

// DbWriteItem writes chain item into k-v store
func DbSetItem(db *badger.DB, hash string, data []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(hash), data)
	})
}

// DbGetBlock checks if a given hash is stored in the k-v store as Block
func DbGetBlock(db *badger.DB, hash string) (block *Block, err error) {
	item, err := DbGetItem(db, hash)
	if err != nil {
		return
	}
	if item.ChainType != ChainType_Block {
		m := fmt.Sprintf("Item's ChainType mismatch. Expected: blk, got: %s", item.ChainType)
		err = errors.New(m)
		return
	}
	dec := gob.NewDecoder(bytes.NewReader(item.Data))
	err = dec.Decode(&block)
	return
}

// DbSetBlock sets a given block into the k-v store as blk chainItem type.
func DbSetBlock(db *badger.DB, block *Block) (err error){
	item := ChainItem{}
	item.ChainType = ChainType_Block
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(*block)
	if err != nil { return }
	err = DbSetItem(db, block.Hash, buf.Bytes())
	return
}