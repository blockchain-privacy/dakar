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

	classifierStatus, _ := GetClassifierStatus(dgraph)

	if classifierStatus.IsClassifying != nil {
		fmt.Println("Currently classifying:", *classifierStatus.IsClassifying)
	}

	if classifierStatus.LastClassifiedBlockId != nil {
		fmt.Println("LastClassifiedBlockId:", *classifierStatus.LastClassifiedBlockId)
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

// GetCrawlerStatus gets the crawler status from the database
func GetCrawlerStatus(c *dgo.Dgraph) (status CrawlerStatus, err error) {
	query := `{
				 q(func: type(CrawlerStatus)){
					uid
					iscrawling
					lastblockid
					lowestblockid
				  }
				}`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Second*20, query)
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

// GetAnalyzerStatus gets the analyzer status from the database
func GetAnalyzerStatus(c *dgo.Dgraph) (status AnalyzerStatus, err error) {
	query := `{
				 q(func: type(AnalyzerStatus)){
					uid
					isanalyzing
					lastanalysedid
				  }
				}`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Second*20, query)
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

// GetClassifierStatus gets the classifier status from the database
func GetClassifierStatus(c *dgo.Dgraph) (status ClassifierStatus, err error) {
	query := `{
				 q(func: type(ClassifierStatus)){
					uid
					isclassifying
					lastclassified
				  }
				}`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Second*20, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r classifierStatusQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// GetHighestBlockId gets the highest block id.
func GetHighestBlockId(c *dgo.Dgraph) (max uint64, err error) {
	query := `{
				var(func: has(id))@filter(eq(dgraph.type, "Block")){
					ids as id
				}
				
				q(){max:max(val(ids))}
			   }`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Second*30, query)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		TopBlock []struct {
			Max uint64 `json:"max,omitempty"`
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
	} else if r.TopBlock[0].Max == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorTopBlockNotFound)
		return
	}
	max = r.TopBlock[0].Max

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
	if len(r.Crawler) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidNumber)
		return
	}

	if r.Crawler[0].IsCrawling == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorIsCrawlingNotFound)
		return
	}

	if r.Crawler[0].LastBlockId == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorLastBlockIdNotFound)
		return
	}

	// if analyzer values exist check them
	if len(r.Analyzer) == 1 {
		if r.Analyzer[0].IsAnalyzing == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorIsAnalyzingNotFound)
			return
		}

		if r.Analyzer[0].LastAnalysedBlockId == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorLastAnalysedBlockIdNotFound)
			return
		}
	}

	status = FrontendStatus{
		IsCrawling:    *r.Crawler[0].IsCrawling,
		LastBlockId:   *r.Crawler[0].LastBlockId,
		LowestBlockId: *r.Crawler[0].LowestBlockId,
	}

	if len(r.Analyzer) == 1 {
		status.IsAnalyzing = *r.Analyzer[0].IsAnalyzing
		status.LastAnalysedBlockId = *r.Analyzer[0].LastAnalysedBlockId
	}

	return
}

// SetCrawlerStatus sets the new crawler status
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

	return db.TxWithRetry(c, time.Second*20, req)
}

// SetAnalyzerStatus sets the new analyzer status
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

	return db.TxWithRetry(c, time.Second*20, req)
}

// SetClassifierStatus sets the new classifier status
func SetClassifierStatus(c *dgo.Dgraph, status ClassifierStatus) error {
	status.Uid = "uid(v)"
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	query := `{
				q(func: type(ClassifierStatus)){
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

	return db.TxWithRetry(c, time.Second*20, req)
}

// SetCrawling sets the crawling status
func SetCrawling(c *dgo.Dgraph, crawling bool) error {
	return SetCrawlerStatus(c, CrawlerStatus{
		IsCrawling: &crawling,
	})
}

// SetAnalyzing sets the analyzing status
func SetAnalyzing(c *dgo.Dgraph, analyzing bool) error {
	return SetAnalyzerStatus(c, AnalyzerStatus{
		IsAnalyzing: &analyzing,
	})
}

// SetClassifying sets the classifying status
func SetClassifying(c *dgo.Dgraph, classifying bool) error {
	return SetClassifierStatus(c, ClassifierStatus{
		IsClassifying: &classifying,
	})
}

// SetLastBlockId sets the last block id
func SetLastBlockId(c *dgo.Dgraph, id uint64) error {
	return SetCrawlerStatus(c, CrawlerStatus{
		LastBlockId: &id,
	})
}

// SetLastAnalysedBlockId sets the last analysed block id
func SetLastAnalysedBlockId(c *dgo.Dgraph, id uint64) error {
	return SetAnalyzerStatus(c, AnalyzerStatus{
		LastAnalysedBlockId: &id,
	})
}

// SetLastClassifiedBlockId sets the last classified block id
func SetLastClassifiedBlockId(c *dgo.Dgraph, id uint64) error {
	return SetClassifierStatus(c, ClassifierStatus{
		LastClassifiedBlockId: &id,
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
