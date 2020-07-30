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

///////////////////////////////////////////////////////
// Utility API
///////////////////////////////////////////////////////

// PrintStatus outputs the stats for the given DB
func PrintStatus(dgraph *dgo.Dgraph) {
	status, _ := Get(dgraph)

	if status.IsCrawling != nil {
		fmt.Println("Currently crawling:", *status.IsCrawling)
	}
	if status.LastBlockId != nil {
		fmt.Println("LastBlockId:", *status.LastBlockId)
	}

	blockCount, _ := dbblk.GetCount(dgraph)
	txCount, _ := dbtx.GetCount(dgraph)
	opCount, _ := dbop.GetCount(dgraph)
	addrCount, _ := dbaddr.GetCount(dgraph)

	fmt.Print("Counts:")
	fmt.Println("\tBlocks:", blockCount)
	fmt.Println("\tTransactions:", txCount)
	fmt.Println("\tOutputs:", opCount)
	fmt.Println("\tAddresses:", addrCount)
}

// gets block information from the database
func Get(c *dgo.Dgraph) (status Status, err error) {
	query := `{
				 q(func: type(Status)){
					iscrawling
					lastblockid
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

// sets the last block id
func SetLastBlockId(c *dgo.Dgraph, id uint64) error {
	return Set(c, Status{
		LastBlockId: &id,
	})
}

// gets the number of status instances in the database
// IMPORTANT: Should always be at most one
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
