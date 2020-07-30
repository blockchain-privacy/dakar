package status

import (
	"dashrpc/db"
	dbaddr "dashrpc/db/address"
	dbblk "dashrpc/db/block"
	dbop "dashrpc/db/output"
	dbtx "dashrpc/db/transaction"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// Database statistics
const (
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

// PrintStatus outputs the stats for the given DB
func PrintStatus(dgraph *dgo.Dgraph) {
	status, _ := Get(dgraph)

	if status.IsCrawling != nil {
		fmt.Println("Currently crawling:", *status.IsCrawling)
	}

	blockCount, _ := dbblk.GetCount(dgraph)
	txCount, _ := dbtx.GetCount(dgraph)
	opCount, _ := dbop.GetCount(dgraph)
	addrCount, _ := dbaddr.GetCount(dgraph)

	fmt.Println("Counts:")
	fmt.Println("\tBlocks:", blockCount)
	fmt.Println("\tTransactions:", txCount)
	fmt.Println("\tOutputs:", opCount)
	fmt.Println("\tAddresses:", addrCount)
	//lastHash := DbGetLastBlockHash(db)
	//lastID := DbGetLastBlockId(db)
	//fmt.Printf("Last hash: %s -- last ID: %v -- ", lastHash, lastID)
	//
	//stopID := DbGetStopBlockId(db)
	//fmt.Printf("Stop ID: %v\n", stopID)
}

// gets block information from the database
func Get(c *dgo.Dgraph) (status Status, err error) {
	query := `{
				 q(func: type(Status)){
					iscrawling
				  }
				}
				`

	resp, err := c.NewReadOnlyTxn().Query(db.GetContext(), query)

	if err != nil {
		return
	}

	var r statusQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.Q) != 1 {
		err = errors.New("wrong number of status objects returned")
		return
	}

	status = r.Q[0]

	return
}

// sets the new status
func Set(c *dgo.Dgraph, status Status) error {
	status.Uid = "uid(v)"
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	query := `{
				q(func: type(Status)){
					v as uid
				  }
				}
				`

	req := &api.Request{
		Query: query,
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	_, err = c.NewTxn().Do(db.GetContext(), req)
	return err
}

// sets the crawling status
func SetCrawling(c *dgo.Dgraph, crawling bool) error {
	return Set(c, Status{
		IsCrawling: &crawling,
	})
}

// gets the number of status instances in the database
// IMPORTANT: Should always be atmost one
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}

///////////////////////////////////////////////////////
// Internal API
///////////////////////////////////////////////////////

//func DbSetLastBlockId(db *badger.DB, id uint64) {
//	DbSetUint64(db, DbBlockLastBlockId, id)
//}
//
//func DbGetLastBlockId(db *badger.DB) uint64 {
//	var value uint64
//	DbGetUint64(db, DbBlockLastBlockId, &value) // ignoring error on purpose
//	return value
//}
//
//func DbSetLastBlockHash(db *badger.DB, lastHash string) {
//	DbSetString(db, DbBlockLastBlockHash, lastHash)
//}
//
//func DbGetLastBlockHash(db *badger.DB) string {
//	var value string
//	DbGetString(db, DbBlockLastBlockHash, &value) // ignoring error on purpose
//	return value
//}
//
//func DbSetStopBlockId(db *badger.DB, id uint64) {
//	DbSetUint64(db, DbBlockStopBlockId, id)
//}
//
//func DbGetStopBlockId(db *badger.DB) uint64 {
//	var value uint64
//	DbGetUint64(db, DbBlockStopBlockId, &value) // ignoring error on purpose
//	return value
//}
