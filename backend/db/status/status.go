package status

import (
	"backend/cmd/cliutil"
	"backend/db"
	dbaddr "backend/db/address"
	dbblk "backend/db/block"
	dbop "backend/db/output"
	dbtx "backend/db/transaction"
	dbus "backend/db/user"
	"backend/external"

	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

// PrintStatus outputs the stats for the given DB
func PrintStatus(dgraph *external.GraphDB) {
	crawlerStatus, _ := GetCrawlerStatus(dgraph)

	if crawlerStatus.IsCrawling != nil {
		fmt.Println("Currently crawling:", *crawlerStatus.IsCrawling)
	}

	if crawlerStatus.LastBlockId != nil {
		fmt.Println("LastBlockId:", *crawlerStatus.LastBlockId)
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
func GetCrawlerStatus(c *external.GraphDB) (status CrawlerStatus, err error) {
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

// GetClassifierStatus gets the classifier status from the database
func GetClassifierStatus(c *external.GraphDB) (status ClassifierStatus, err error) {
	query := `{
				 q(func: type(ClassifierStatus)){
					uid
					isclassifying
					lastclassifiedid
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
func GetHighestBlockId(c *external.GraphDB) (max uint64, err error) {
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

// GetFrontendStatus gets verbose status information from the database
func GetFrontendStatus(c *external.GraphDB) (status FrontendStatus, err error) {
	query := `{
				crawler(func: type(CrawlerStatus)){
					iscrawling
					lastblockid
					lowestblockid
				}
				classifier(func: type(ClassifierStatus)){
					isclassifying
					lastclassifiedid
				}
			}`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.Query(ctx, query, nil)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Crawler    []CrawlerStatus    `json:"crawler,omitempty"`
		Classifier []ClassifierStatus `json:"classifier,omitempty"`
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

	if len(r.Classifier) == 1 {
		if r.Classifier[0].IsClassifying == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorIsClassifyingNotFound)
			return
		}

		if r.Classifier[0].LastClassifiedBlockId == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorLastClassifiedBlockIdNotFound)
			return
		}
	}

	status = FrontendStatus{
		IsCrawling:    *r.Crawler[0].IsCrawling,
		LastBlockId:   *r.Crawler[0].LastBlockId,
		LowestBlockId: *r.Crawler[0].LowestBlockId,
	}

	if len(r.Classifier) == 1 {
		status.IsClassifying = *r.Classifier[0].IsClassifying
		status.LastClassifiedBlockId = *r.Classifier[0].LastClassifiedBlockId
	}

	return
}

// SetCrawlerStatus sets the new crawler status
func SetCrawlerStatus(c *external.GraphDB, status CrawlerStatus) error {
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

	return db.TxWithRetry(c, time.Minute*10, req)
}

// SetClassifierStatus sets the new classifier status
func SetClassifierStatus(c *external.GraphDB, status ClassifierStatus) error {
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

	return db.TxWithRetry(c, time.Minute*10, req)
}

// SetCrawling sets the crawling status
func SetCrawling(c *external.GraphDB, crawling bool) error {
	return SetCrawlerStatus(c, CrawlerStatus{
		IsCrawling: &crawling,
	})
}

// SetClassifying sets the classifying status
func SetClassifying(c *external.GraphDB, classifying bool) error {
	return SetClassifierStatus(c, ClassifierStatus{
		IsClassifying: &classifying,
	})
}

// SetLastBlockId sets the last block id
func SetLastBlockId(c *external.GraphDB, id uint64) error {
	return SetCrawlerStatus(c, CrawlerStatus{
		LastBlockId: &id,
	})
}

// SetLastClassifiedBlockId sets the last classified block id
func SetLastClassifiedBlockId(c *external.GraphDB, id uint64) error {
	return SetClassifierStatus(c, ClassifierStatus{
		LastClassifiedBlockId: &id,
	})
}

// GetCount gets the number of status instances in the database
// IMPORTANT: Should always be at most one
func GetCount(c *external.GraphDB) (uint64, error) {
	return db.GetCount(c, CrawlerStatusDType)
}

// IsConnectionEstablished test the database connection
func IsConnectionEstablished(c *external.GraphDB) bool {
	query := `{
				 q(func: has(blockhash), first:1){
					uid
				  }
				}`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	_, err := c.Query(ctx, query, nil)
	if err != nil {
		return false
	}

	return true
}
