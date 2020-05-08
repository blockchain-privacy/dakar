package dashrpc

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

const VersionString = "v0.0.1"

// Database statistics
const (
	DbGlobalBlockCount      = "DB_GLOBAL_BLOCK_COUNT"
	DbGlobalTxCount         = "DB_GLOBAL_TX_COUNT"
	DbBlockRangeUp          = "DB_BLOCK_RANGE_UP"
	DbBlockRangeDown        = "DB_BLOCK_RANGE_DOWN"
	DbBlockLastBlockId      = "DB_BLOCK_LAST_ID"
	DbBlockLastBlockHash    = "DB_BLOCK_LAST_HASH"
	DbBlockStopBlockId      = "DB_BLOCK_STOP_BLOCK_ID"
	DbBlockStatus           = "DB_BLOCK_STATUS"
	DbBlockStatusUnknown    = "DB_BLOCK_STATUS_UNKNOWN"
	DbBlockStatusProcessing = "DB_BLOCK_STATUS_PROCESSING"
	DbBlockStatusFinished   = "DB_BLOCK_STATUS_FINISHED"
)



///////////////////////////////////////////////////////
// Utility API
///////////////////////////////////////////////////////

// SetupBadgerDB instantiates and returns the DB
func SetupBadgerDB(badgerDir string) *badger.DB {
	// Setup the Badger DB connection
	opts := badger.DefaultOptions(badgerDir)
	opts.WithNumVersionsToKeep(1)
	opts.WithNumMemtables(50)
	opts.WithMaxTableSize(512 << 20)
	opts.WithSyncWrites(false)
	// not needed opts.WithValueDir(*badgerDir)
	// not needed opts.WithDir(*badgerDir)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// PrintStatus outputs the stats for the given DB
func PrintStatus(db *badger.DB) {
	var status string
	DbGetStatus(db, &status)
	fmt.Printf("Status: %s\n", status)
	var lastID, stopID uint64
	var lastHash string
	err := DbGetUint64(db, DbBlockLastBlockId, &lastID)
	if err != nil {
		fmt.Printf("Error: cannot read LastBlockID: %v\n", err)
	}
	err = DbGetUint64(db, DbBlockStopBlockId, &stopID)
	if err != nil {
		fmt.Printf("Error: cannot read StopBlockID: %v\n", err)
	}
	err = DbGetString(db, DbBlockLastBlockHash, &lastHash)
	if err != nil {
		fmt.Printf("Error: cannot read LastBlockID: %v\n", err)
	}
	fmt.Printf("Last hash: %s -- last ID: %v -- ", lastHash, lastID)
	fmt.Printf("Stop ID: %v\n", stopID)

	rangeUp := DbGetRangeUp(db)
	rangeDown := DbGetRangeDown(db)
	fmt.Printf("DB range: %v - %v\n", rangeDown, rangeUp)
}


///////////////////////////////////////////////////////
// Internal API
///////////////////////////////////////////////////////

// DbSetStatus gets the status
func DbSetStatus(db *badger.DB, status string) {
	err := DbSetString(db, DbBlockStatus, status)
	if err != nil {
		fmt.Printf("Error: saving status failed: %v\n", err)
	}
}

// DbGetStatus get the status
func DbGetStatus(db *badger.DB, status *string) {
	err := DbGetString(db, DbBlockStatus, status)
	if err != nil {
		*status  = DbBlockStatusUnknown
	}
}

func DbGetBlockCount(db *badger.DB) uint64 {
	var count uint64
	DbGetUint64(db, DbGlobalBlockCount, &count)
	return count
}

func DbIncrementBlockCount(db *badger.DB) {
	count := DbGetBlockCount(db)
	count++ // should be 1 on error
	DbSetUint64(db, DbGlobalBlockCount, count) // Ignoring Error on purpose!
}

func DbGetGlobalTxCount(db *badger.DB) uint64 {
	var count uint64
	DbGetUint64(db, DbGlobalTxCount, &count)
	return count
}

func DbIncrementGlobalTxCount(db *badger.DB) {
	count := DbGetGlobalTxCount(db)
	count++ // should be 1 on error
	DbSetUint64(db, DbGlobalTxCount, count)
}

func DbSetRangeUp(db *badger.DB, up uint64) {
	DbSetUint64(db, DbBlockRangeUp, up) // Ignoring errors on purpose
}

func DbGetRangeUp(db *badger.DB) uint64 {
	var value uint64
	DbGetUint64(db, DbBlockRangeUp, &value) // ignoring error on purpose
	return value
}

func DbSetRangeDown(db *badger.DB, down uint64) {
	DbSetUint64(db, DbBlockRangeDown, down)
}

func DbGetRangeDown(db *badger.DB) uint64 {
	var value uint64
	DbGetUint64(db, DbBlockRangeDown, &value) // ignoring error on purpose
	return value
}