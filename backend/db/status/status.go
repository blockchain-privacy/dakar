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
func PrintStatus(dgraph external.Database) {
	crawlerStatus, _ := GetCrawlerStatus(dgraph)

	if crawlerStatus.IsCrawling != nil {
		fmt.Println("Currently crawling:", *crawlerStatus.IsCrawling)
	}

	if crawlerStatus.LastBlockID != nil {
		fmt.Println("LastBlockID:", *crawlerStatus.LastBlockID)
	}

	classifierStatus, _ := GetClassifierStatus(dgraph)

	if classifierStatus.IsClassifying != nil {
		fmt.Println("Currently classifying:", *classifierStatus.IsClassifying)
	}

	if classifierStatus.LastClassifiedBlockID != nil {
		fmt.Println("LastClassifiedBlockID:", *classifierStatus.LastClassifiedBlockID)
	}

	hmiStatus, _ := GetClusteringHMIStatus(dgraph)

	if hmiStatus.IsClustering != nil {
		fmt.Println("Currently hierarchical multi-input clustering:", *hmiStatus.IsClustering)
	}

	if hmiStatus.LastClusteredBlockID != nil {
		fmt.Println("LastClusteredBlockID (hierarchical multi-input):", *hmiStatus.LastClusteredBlockID)
	}

	fmiStatus, _ := GetClusteringFMIStatus(dgraph)

	if fmiStatus.IsClustering != nil {
		fmt.Println("Currently flat multi-input clustering:", *fmiStatus.IsClustering)
	}

	if fmiStatus.LastClusteredBlockID != nil {
		fmt.Println("LastClusteredBlockID (flat multi-input):", *fmiStatus.LastClusteredBlockID)
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
func GetCrawlerStatus(c external.Database) (status CrawlerStatus, err error) {
	query := `{
				 q(func: type(` + CrawlerStatusDType + `)){
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
func GetClassifierStatus(c external.Database) (status ClassifierStatus, err error) {
	query := `{
				 q(func: type(` + ClassifierStatusDType + `)){
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

// GetClusteringHMIStatus gets the hierarchical multi-input clustering status from the database
func GetClusteringHMIStatus(c external.Database) (status ClusteringHierarchicalMultiInputStatus, err error) {
	query := `{
				 q(func: type(` + ClusteringHierarchicalMultiInputDType + `)){
					uid
					isclustering
					lastclusteredid
				  }
				}`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Second*20, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r clusteringHMIStatusQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// GetClusteringFMIStatus gets the flat multi-input clustering status from the database
func GetClusteringFMIStatus(c external.Database) (status ClusteringFlatMultiInputStatus, err error) {
	query := `{
				 q(func: type(` + ClusteringFlatMultiInputDType + `)){
					uid
					isclustering
					lastclusteredid
				  }
				}`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Second*20, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r clusteringFMIStatusQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// GetHighestBlockID gets the highest block id.
func GetHighestBlockID(c external.Database) (max uint64, err error) {
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorTopBlockNotFound)
		return
	} else if len(r.TopBlock) > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorInvalidNumber)
		return
	} else if r.TopBlock[0].Max == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorTopBlockNotFound)
		return
	}
	max = r.TopBlock[0].Max

	return
}

// GetFrontendStatus gets verbose status information from the database
func GetFrontendStatus(c external.Database) (status FrontendStatus, err error) {
	query := `{
				crawler(func: type(` + CrawlerStatusDType + `)){
					iscrawling
					lastblockid
					lowestblockid
				}
				classifier(func: type(` + ClassifierStatusDType + `)){
					isclassifying
					lastclassifiedid
				}
				hmi(func: type(` + ClusteringHierarchicalMultiInputDType + `)){
					isclustering
					lastclusteredid
				}
				fmi(func: type(` + ClusteringFlatMultiInputDType + `)){
					isclustering
					lastclusteredid
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
		Crawler    []CrawlerStatus                          `json:"crawler,omitempty"`
		Classifier []ClassifierStatus                       `json:"classifier,omitempty"`
		HMI        []ClusteringHierarchicalMultiInputStatus `json:"hmi,omitempty"`
		FMI        []ClusteringFlatMultiInputStatus         `json:"fmi,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// check if all values are set correctly
	if len(r.Crawler) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorInvalidNumber)
		return
	}

	if r.Crawler[0].IsCrawling == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorIsCrawlingNotFound)
		return
	}

	if r.Crawler[0].LastBlockID == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorLastBlockIDNotFound)
		return
	}

	if len(r.Classifier) == 1 {
		if r.Classifier[0].IsClassifying == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorIsClassifyingNotFound)
			return
		}

		if r.Classifier[0].LastClassifiedBlockID == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorLastClassifiedBlockIDNotFound)
			return
		}
	}

	if len(r.HMI) == 1 {
		if r.HMI[0].IsClustering == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorIsClusteringMultiInputNotFound)
			return
		}

		if r.HMI[0].LastClusteredBlockID == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorLastClusteringMultiInputBlockIDNotFound)
			return
		}
	}

	if len(r.FMI) == 1 {
		if r.FMI[0].IsClustering == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorIsClusteringMultiInputNotFound)
			return
		}

		if r.FMI[0].LastClusteredBlockID == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errorLastClusteringMultiInputBlockIDNotFound)
			return
		}
	}

	status = FrontendStatus{
		IsCrawling:    *r.Crawler[0].IsCrawling,
		LastBlockID:   *r.Crawler[0].LastBlockID,
		LowestBlockID: *r.Crawler[0].LowestBlockID,
	}

	if len(r.Classifier) == 1 {
		status.IsClassifying = *r.Classifier[0].IsClassifying
		status.LastClassifiedBlockID = *r.Classifier[0].LastClassifiedBlockID
	}

	if len(r.HMI) == 1 {
		status.IsClusteringHMI = *r.HMI[0].IsClustering
		status.LastClusteredHMIBlockID = *r.HMI[0].LastClusteredBlockID
	}

	if len(r.FMI) == 1 {
		status.IsClusteringFMI = *r.FMI[0].IsClustering
		status.LastClusteredFMIBlockID = *r.FMI[0].LastClusteredBlockID
	}

	return
}

// SetCrawlerStatus sets the new crawler status
func SetCrawlerStatus(c external.Database, status CrawlerStatus) error {
	status.UID = "uid(v)"
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	req := &api.Request{
		Query:     "{q(func: type(" + CrawlerStatusDType + ")){v as uid}}",
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	}

	return db.TxWithRetry(c, time.Minute*10, req)
}

// SetClassifierStatus sets the new classifier status
func SetClassifierStatus(c external.Database, status ClassifierStatus) error {
	status.UID = "uid(v)"
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	req := &api.Request{
		Query:     "{q(func:type(" + ClassifierStatusDType + ")){v as uid}}",
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	}

	return db.TxWithRetry(c, time.Minute*10, req)
}

// SetClusteringHMIStatus sets the new hierarchical multi-input clustering status
func SetClusteringHMIStatus(c external.Database, status ClusteringHierarchicalMultiInputStatus) error {
	status.UID = "uid(v)"
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	req := &api.Request{
		Query:     "{q(func:type(" + ClusteringHierarchicalMultiInputDType + ")){v as uid}}",
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	}

	return db.TxWithRetry(c, time.Minute*10, req)
}

// SetClusteringFMIStatus sets the new flat multi-input clustering status
func SetClusteringFMIStatus(c external.Database, status ClusteringFlatMultiInputStatus) error {
	status.UID = "uid(v)"
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	req := &api.Request{
		Query:     "{q(func:type(" + ClusteringFlatMultiInputDType + ")){v as uid}}",
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	}

	return db.TxWithRetry(c, time.Minute*10, req)
}

// SetCrawling sets the crawling status
func SetCrawling(c external.Database, crawling bool) error {
	return SetCrawlerStatus(c, CrawlerStatus{
		IsCrawling: &crawling,
	})
}

// SetClassifying sets the classifying status
func SetClassifying(c external.Database, classifying bool) error {
	return SetClassifierStatus(c, ClassifierStatus{
		IsClassifying: &classifying,
	})
}

// SetClusteringHMI sets the hierarchical multi-input clustering status
func SetClusteringHMI(c external.Database, clustering bool) error {
	return SetClusteringHMIStatus(c, ClusteringHierarchicalMultiInputStatus{
		IsClustering: &clustering,
	})
}

// SetClusteringFMI sets the flat multi-input clustering status
func SetClusteringFMI(c external.Database, clustering bool) error {
	return SetClusteringFMIStatus(c, ClusteringFlatMultiInputStatus{
		IsClustering: &clustering,
	})
}

// SetLastBlockID sets the last block id
func SetLastBlockID(c external.Database, id uint64) error {
	return SetCrawlerStatus(c, CrawlerStatus{
		LastBlockID: &id,
	})
}

// SetLastClassifiedBlockID sets the last classified block id
func SetLastClassifiedBlockID(c external.Database, id uint64) error {
	return SetClassifierStatus(c, ClassifierStatus{
		LastClassifiedBlockID: &id,
	})
}

// SetLastClusteredHMIBlockID sets the last clustered multi-input block id
func SetLastClusteredHMIBlockID(c external.Database, id uint64) error {
	return SetClusteringHMIStatus(c, ClusteringHierarchicalMultiInputStatus{
		LastClusteredBlockID: &id,
	})
}

// SetLastClusteredFMIBlockID sets the last clustered multi-input block id
func SetLastClusteredFMIBlockID(c external.Database, id uint64) error {
	return SetClusteringFMIStatus(c, ClusteringFlatMultiInputStatus{
		LastClusteredBlockID: &id,
	})
}

// GetCount gets the number of status instances in the database
// IMPORTANT: Should always be at most one
func GetCount(c external.Database) (uint64, error) {
	return db.GetCount(c, CrawlerStatusDType)
}

// IsConnectionEstablished test the database connection
func IsConnectionEstablished(c external.Database) bool {
	ctx, cancel := db.GetBackendContext()
	defer cancel()
	_, err := c.Query(ctx, "{q(func: has(blockhash),first:1){uid}}", nil)

	return err == nil
}
