package status

import (
	"backend/cmd/cliutil"
	"backend/db"
	dbaddr "backend/db/address"
	dbblk "backend/db/block"
	dbop "backend/db/output"
	dbtx "backend/db/transaction"
	dbus "backend/db/user"
	"time"

	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// PrintStatus outputs the stats for the given DB
func PrintStatus(dgraph *dgo.Dgraph) {
	crawlerStatus, _ := GetCrawlerStatus(dgraph)

	if crawlerStatus.IsCrawling != nil {
		fmt.Println("Currently crawling:", *crawlerStatus.IsCrawling)
	}

	if crawlerStatus.LastBlockId != nil {
		fmt.Println("LastBlockId:", *crawlerStatus.LastBlockId)
	}

	analyzerStatus, _ := GetAnalyzerStatus(dgraph)

	if analyzerStatus.IsAnalyzing != nil {
		fmt.Println("Currently analyzing:", *analyzerStatus.IsAnalyzing)
	}

	if analyzerStatus.LastAnalysedBlockId != nil {
		fmt.Println("LastAnalysedBlockId:", *analyzerStatus.LastAnalysedBlockId)
	}

	blockCount, _ := dbblk.GetCount(dgraph)
	txCount, _ := dbtx.GetCount(dgraph)
	opCount, _ := dbop.GetCount(dgraph)
	addrCount, _ := dbaddr.GetCount(dgraph)
	userCount, _ := dbus.GetUserCount(dgraph)
	roleCount, _ := dbus.GetRoleCount(dgraph)

	fmt.Print("Counts:")
	fmt.Println("\tBlocks:", blockCount)
	fmt.Println("\tTransactions:", txCount)
	fmt.Println("\tOutputs:", opCount)
	fmt.Println("\tAddresses:", addrCount)
	fmt.Println("\tUsers:", userCount)
	fmt.Println("\tRoles:", roleCount)
}

// gets status information from the database
func GetCrawlerStatus(c *dgo.Dgraph) (status CrawlerStatus, err error) {
	query := `{
				 q(func: type(CrawlerStatus)){
					uid
					iscrawling
					lastblockid
					lowestblockid
				  }
				}`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxWithRetry(c, ctx, query)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r crawlerStatusQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// gets status information from the database
func GetAnalyzerStatus(c *dgo.Dgraph) (status AnalyzerStatus, err error) {
	query := `{
				 q(func: type(AnalyzerStatus)){
					uid
					isanalyzing
					lastanalysedid
				  }
				}`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxWithRetry(c, ctx, query)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r analyzerStatusQuery

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
	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxWithRetry(c, ctx, query)

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

	if len(r.TopBlock) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorTopBlockNotFound)
		return
	} else if len(r.TopBlock) > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidNumber)
		return
	}
	id = r.TopBlock[0].Id

	return
}

// gets verbose status information from the database
func GetFrontendStatus(c *dgo.Dgraph) (status FrontendStatus, err error) {
	query := `{
				crawler(func: type(CrawlerStatus)){
					iscrawling
					lastblockid
					lowestblockid
				}
				analyzer(func: type(AnalyzerStatus)){
					isanalyzing
					lastanalysedid
				}
			}`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().Query(ctx, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Crawler  []CrawlerStatus  `json:"crawler,omitempty"`
		Analyzer []AnalyzerStatus `json:"analyzer,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// check if all values are set correctly
	if len(r.Crawler) != 1 || len(r.Analyzer) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidNumber)
		return
	}

	if r.Crawler[0].IsCrawling == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorIsCrawlingNotFound)
		return
	}

	if r.Analyzer[0].IsAnalyzing == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorIsAnalyzingNotFound)
		return
	}

	if r.Crawler[0].LastBlockId == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorLastBlockIdNotFound)
		return
	}

	if r.Analyzer[0].LastAnalysedBlockId == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorLastAnalysedBlockIdNotFound)
		return
	}

	status = FrontendStatus{
		IsCrawling:          *r.Crawler[0].IsCrawling,
		LastBlockId:         *r.Crawler[0].LastBlockId,
		LowestBlockId:       *r.Crawler[0].LowestBlockId,
		IsAnalyzing:         *r.Analyzer[0].IsAnalyzing,
		LastAnalysedBlockId: *r.Analyzer[0].LastAnalysedBlockId,
	}

	return
}

// sets the new status
func SetCrawlerStatus(c *dgo.Dgraph, status CrawlerStatus) error {
	status.Uid = "uid(v)"
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	query := `{
				q(func: type(CrawlerStatus)){
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

	return db.TxWithRetryAndTimeout(c, time.Second*20, req)
}

// sets the new status
func SetAnalyzerStatus(c *dgo.Dgraph, status AnalyzerStatus) error {
	status.Uid = "uid(v)"
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	query := `{
				q(func: type(AnalyzerStatus)){
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

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	return db.TxWithRetry(c, ctx, req)
}

// sets the crawling status
func SetCrawling(c *dgo.Dgraph, crawling bool) error {
	return SetCrawlerStatus(c, CrawlerStatus{
		IsCrawling: &crawling,
	})
}

// sets the analyzing status
func SetAnalyzing(c *dgo.Dgraph, analyzing bool) error {
	return SetAnalyzerStatus(c, AnalyzerStatus{
		IsAnalyzing: &analyzing,
	})
}

// sets the last block id
func SetLastBlockId(c *dgo.Dgraph, id uint64) error {
	return SetCrawlerStatus(c, CrawlerStatus{
		LastBlockId: &id,
	})
}

// sets the last analysed block id
func SetLastAnalysedBlockId(c *dgo.Dgraph, id uint64) error {
	return SetAnalyzerStatus(c, AnalyzerStatus{
		LastAnalysedBlockId: &id,
	})
}

// gets the number of status instances in the database
// IMPORTANT: Should always be at most one
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, CrawlerStatusDType)
}

// IsConnectionEstablished test the database connection
func IsConnectionEstablished(c *dgo.Dgraph) bool {
	query := `{
				 q(func: has(blockhash), first:1){
					uid
				  }
				}`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	_, err := c.NewReadOnlyTxn().Query(ctx, query)
	if err != nil {
		return false
	}

	return true
}
