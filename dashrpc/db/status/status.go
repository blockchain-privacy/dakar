package status

import (
	"dashrpc/cmd/cliutil"
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
	status, _ := GetStatus(dgraph)

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
func GetStatus(c *dgo.Dgraph) (status Status, err error) {
	query := `{
				 q(func: type(Status)){
					uid
					iscrawling
					lastblockid
				  }
				}
				`

	resp, err := c.NewReadOnlyTxn().Query(db.GetContext(), query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r statusQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// gets the highest block id.
// on large datasets this is an expensive call.
func GetHighestBlockId(c *dgo.Dgraph) (id uint64, err error) {
	return getTopBlockId(c, false)
}

// gets the lowest block id.
// on large datasets this is an expensive call.
func GetLowestBlockId(c *dgo.Dgraph) (id uint64, err error) {
	return getTopBlockId(c, true)
}

// gets the top block id, either ordered descending (highest or ascending (lowest).
// on large datasets this is an expensive call.
func getTopBlockId(c *dgo.Dgraph, ascending bool) (id uint64, err error) {
	order := "desc"

	if ascending {
		order = "asc"
	}

	query := fmt.Sprintf(`{
				q(func: type(Block), order%s: id,first: 1) @filter(ge(id,0)){
					id
				}
			}`, order)

	resp, err := c.NewReadOnlyTxn().Query(db.GetContext(), query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		TopBlock []struct {
			Id uint64 `json:"id,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	// todo compare errors with error.Is or error.As
	if len(r.TopBlock) == 0 {
		err = errors.New(ErrorTopBlockNotFound)
		return
	} else if len(r.TopBlock) > 1 {
		err = errors.New(ErrorInvalidNumber)
		return
	}
	id = r.TopBlock[0].Id

	return
}

// gets verbose status information from the database
func GetVerbose(c *dgo.Dgraph) (status VerboseStatus, err error) {
	query := `{
				status(func: type(Status)){
					uid
					iscrawling
					lastblockid
					lowestblockid
				}
				blk(func: type(Block)){
					count: count(uid)
				}
				tx(func: type(Transaction)){
					count: count(uid)
				}
				op(func: type(Output)){
					count: count(uid)
				}
				addr(func: type(Address)){
					count: count(uid)
				}
			}`

	resp, err := c.NewReadOnlyTxn().Query(db.GetContext(), query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Status []Status `json:"status,omitempty"`
		Blk    []struct {
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// todo compare errors with error.Is or error.As
	// check if all values are set correctly
	if len(r.Status) != 1 || len(r.Blk) != 1 || len(r.Tx) != 1 ||
		len(r.Op) != 1 || len(r.Addr) != 1 {
		err = errors.New(ErrorInvalidNumber)
		return
	}

	if r.Status[0].IsCrawling == nil {
		err = errors.New(ErrorIsCrawlingNotFound)
		return
	}

	if r.Status[0].LastBlockId == nil {
		err = errors.New(ErrorLastBlockIdNotFound)
		return
	}

	status = VerboseStatus{
		Uid:              r.Status[0].Uid,
		IsCrawling:       *r.Status[0].IsCrawling,
		LastBlockId:      *r.Status[0].LastBlockId,
		LowestBlockId:    *r.Status[0].LowestBlockId,
		AddressCount:     r.Addr[0].Count,
		TransactionCount: r.Tx[0].Count,
		BlockCount:       r.Blk[0].Count,
		OutputCount:      r.Op[0].Count,
	}

	return
}

// sets the new status
func SetStatus(c *dgo.Dgraph, status Status) error {
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
	return SetStatus(c, Status{
		IsCrawling: &crawling,
	})
}

// sets the last block id
func SetLastBlockId(c *dgo.Dgraph, id uint64) error {
	return SetStatus(c, Status{
		LastBlockId: &id,
	})
}

// gets the number of status instances in the database
// IMPORTANT: Should always be at most one
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
