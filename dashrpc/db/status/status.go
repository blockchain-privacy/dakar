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

// gets status information from the database
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

	return r.payload()
}

// gets verbose status information from the database
func GetVerbose(c *dgo.Dgraph) (status VerboseStatus, err error) {
	query := `{
				  status(func: type(Status)){
						iscrawling
						lastblockid
					}
				  highestblockid(func: type(Block), orderdesc: id,first: 1) @filter(ge(id,0)){
					id
				  }
				  blk(func: type(Block)){
						count: count(uid)
					}
				  tx(func: type(Transaction)){
						count: count(uid)
					}
				  op(func: type(TxOutput)){
						count: count(uid)
					}
				  addr(func: type(Address)){
						count: count(uid)
					}
			}`

	resp, err := c.NewReadOnlyTxn().Query(db.GetContext(), query)

	if err != nil {
		return
	}

	var r struct {
		Status       []Status `json:"status,omitempty"`
		Highestblock []struct {
			Id uint64 `json:"id,omitempty"`
		} `json:"highestblockid,omitempty"`
		Blk []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"blk,omitempty"`
		Tx []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"tx,omitempty"`
		Op []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"op,omitempty"`
		Addr []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"addr,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	// check if all values are set correctly
	if len(r.Status) != 1 || len(r.Highestblock) != 1 || len(r.Blk) != 1 ||
		len(r.Tx) != 1 || len(r.Op) != 1 || len(r.Addr) != 1 {
		err = errors.New(ErrorInvalidNumber)
		return
	}

	if r.Status[0].LastBlockId == nil {
		err = errors.New(ErrorIsCrawlingNotFound)
		return
	}

	if r.Status[0].IsCrawling == nil {
		err = errors.New(ErrorLastBlockIdNotFound)
		return
	}

	status = VerboseStatus{
		Uid:              r.Status[0].Uid,
		IsCrawling:       *r.Status[0].IsCrawling,
		LastBlockId:      *r.Status[0].LastBlockId,
		HighestBlockId:   r.Highestblock[0].Id,
		AddressCount:     r.Addr[0].Count,
		TransactionCount: r.Tx[0].Count,
		BlockCount:       r.Blk[0].Count,
		OutputCount:      r.Op[0].Count,
	}

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
