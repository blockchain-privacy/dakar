/*
Package dashrpc is a collection of utilities for processing Dash blockchain data.
The codename for the project is Dakar.

Key-value store

Badger DB (https://github.com/dgraph-io/badger) is a fast key-value store optimised to work as cache system.
After the data is loaded and hashed in memory, the lookups on large datasets are fast.
There is no query language, and the db supports simple store and retrieval operations.
The keys are strings, and the values can be any array of bytes. In Dakar, we have a thin
layer on top of raw badger to provide primitive type-safety and type-enforcement subsystem.
All interactions with the underlying badger db are mediated through two functions, from <pre>database.go</pre>:

	func DbSetItem(db *badger.DB, hash string, data []byte, itemType string) error
	func DbGetItem(db *badger.DB, hash string, item *ChainItem) error

The ChainItem wraps the untyped binary blob in a simple struct and attaches the type information.

	// ChainItem represents a generic blockchain item
	type ChainItem struct {
		ItemType string
		Data     []byte
	}

The <pre>hash</pre> is the key of the item (does not need to be actually a hash).
We use a TX id, address, block id, and so on. See later sections.

Data types

The core datatypes on top of primitive key-value store are as follows:

	// Database core data
	const (
		ChainType_Block		= "blk" // Block hash -> Block
		ChainType_TxDetails	= "txd" // TX hash -> TxDetails
		ChainType_AddrData	= "add" // add -> AddressData
		ChainType_Cluster	= "adc" // cluster ID -> ClusterData
	)

For the actual data types related to structs see the types.go file.

Data access functions

Accessing various datatypes (read and write) through the primitive DbSetItem and DbGetItem would be tedious and error prone. Therefore, any application logic should make use of custom typed DB accessor functions, eg, for primitive types:

	func DbSetString(db *badger.DB, key string, data string) error
	func DbGetString(db *badger.DB, key string, data *string) error
 	func DbSetUint64(db *badger.DB, key string, data uint64) error
	func DbGetUint64(db *badger.DB, key string, item *uint64) error

and for complex data types:

	func DbGetBlock(db *badger.DB, hash string, block *Block) error
	func DbSetBlock(db *badger.DB, block Block) error
	and so on...

NOTE: We also store certain statistics in the DB as primitive data types, and these are available
through utility functions and constants in the file <pre>status.go</pre>

*/
package dashrpc
