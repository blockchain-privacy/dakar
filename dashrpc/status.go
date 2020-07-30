package dashrpc

import (
	dbblk "dashrpc/db/block"
	dbtx "dashrpc/db/transaction"
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/dgo/v2"
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
//func PrintStatus(db *badger.DB) {
//	status := DbGetStatus(db)
//	fmt.Printf("Status: %s\n", status)
//
//	dbBlockCount := DbGetBlockCount(db)
//	dbTxCount := DbGetGlobalTxCount(db)
//	fmt.Printf("DB block count: %v  TX count: %v\n", dbBlockCount, dbTxCount)
//
//	lastHash := DbGetLastBlockHash(db)
//	lastID := DbGetLastBlockId(db)
//	fmt.Printf("Last hash: %s -- last ID: %v -- ", lastHash, lastID)
//
//	stopID := DbGetStopBlockId(db)
//	fmt.Printf("Stop ID: %v\n", stopID)
//}

// PrintStatus outputs the stats for the given DB
func PrintStatus(dgraph *dgo.Dgraph) {
	//status := DbGetStatus(db)
	//fmt.Printf("Status: %s\n", status)

	blockCount, _ := dbblk.GetCount(dgraph)
	txCount, _ := dbtx.GetCount(dgraph)

	fmt.Printf("DB block count: %v  TX count: %v\n", blockCount, txCount)

	//lastHash := DbGetLastBlockHash(db)
	//lastID := DbGetLastBlockId(db)
	//fmt.Printf("Last hash: %s -- last ID: %v -- ", lastHash, lastID)
	//
	//stopID := DbGetStopBlockId(db)
	//fmt.Printf("Stop ID: %v\n", stopID)
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
func DbGetStatus(db *badger.DB) string {
	var status string
	err := DbGetString(db, DbBlockStatus, &status)
	if err != nil {
		status = DbBlockStatusUnknown
	}
	return status
}

func DbSetLastBlockId(db *badger.DB, id uint64) {
	DbSetUint64(db, DbBlockLastBlockId, id)
}

func DbGetLastBlockId(db *badger.DB) uint64 {
	var value uint64
	DbGetUint64(db, DbBlockLastBlockId, &value) // ignoring error on purpose
	return value
}

func DbSetLastBlockHash(db *badger.DB, lastHash string) {
	DbSetString(db, DbBlockLastBlockHash, lastHash)
}

func DbGetLastBlockHash(db *badger.DB) string {
	var value string
	DbGetString(db, DbBlockLastBlockHash, &value) // ignoring error on purpose
	return value
}

func DbSetStopBlockId(db *badger.DB, id uint64) {
	DbSetUint64(db, DbBlockStopBlockId, id)
}

func DbGetStopBlockId(db *badger.DB) uint64 {
	var value uint64
	DbGetUint64(db, DbBlockStopBlockId, &value) // ignoring error on purpose
	return value
}

// Information about the state of the database
type Meta struct {
	LastBlockId      uint64 `json:"lastblockid"`
	StopBlockId      uint64 `json:"stopblockid"`
	LastBlockHash    string `json:"lastblockhash"`
	Status           string `json:"status"`
	RangeUp          uint64 `json:"rangeup"`
	RangeDown        uint64 `json:"rangedown"`
	GlobalBlockCount uint64 `json:"globalblockcount"`
	GlobalTxCount    uint64 `json:"globaltxcount"`
}

func (m Meta) String() string {
	return fmt.Sprintf(
		"LastBlockId: %v\n"+
			"StopBlockId: %v\n"+
			"LastBlockHash: %v\n"+
			"Status: %v\n"+
			"RangeUp: %v\n"+
			"RangeDown: %v\n"+
			"GlobalBlockCount: %v\n"+
			"GlobalTxCount: %v\n",
		m.LastBlockId, m.StopBlockId, m.LastBlockHash, m.Status, m.RangeUp, m.RangeDown, m.GlobalBlockCount, m.GlobalTxCount)
}
