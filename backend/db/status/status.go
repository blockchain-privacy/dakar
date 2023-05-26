package status

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"

	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/dgo/v230/protos/api"
)

const uidV = "uid(v)"

// GetCrawlerStatus gets the crawler status from the database
func GetCrawlerStatus(c external.Database) (status CrawlerStatus, err error) {
	query := `{
				 q(func: type(` + CrawlerStatusDType + `)){
					uid
					iscrawling
					lastblockid
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errTopBlockNotFound)
		return
	} else if len(r.TopBlock) > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidNumber)
		return
	} else if r.TopBlock[0].Max == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errTopBlockNotFound)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errInvalidNumber)
		return
	}

	if r.Crawler[0].IsCrawling == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errIsCrawlingNotFound)
		return
	}

	if r.Crawler[0].LastBlockID == nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errLastBlockIDNotFound)
		return
	}

	if len(r.Classifier) == 1 {
		if r.Classifier[0].IsClassifying == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errIsClassifyingNotFound)
			return
		}

		if r.Classifier[0].LastClassifiedBlockID == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errLastClassifiedBlockIDNotFound)
			return
		}
	}

	if len(r.HMI) == 1 {
		if r.HMI[0].IsClustering == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errIsClusteringMultiInputNotFound)
			return
		}

		if r.HMI[0].LastClusteredBlockID == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errLastClusteringMultiInputBlockIDNotFound)
			return
		}
	}

	if len(r.FMI) == 1 {
		if r.FMI[0].IsClustering == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errIsClusteringMultiInputNotFound)
			return
		}

		if r.FMI[0].LastClusteredBlockID == nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), errLastClusteringMultiInputBlockIDNotFound)
			return
		}
	}

	status = FrontendStatus{
		IsCrawling:  *r.Crawler[0].IsCrawling,
		LastBlockID: *r.Crawler[0].LastBlockID,
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

// GetMeta gets the database metadata
func GetMeta(c external.Database) (meta Meta, err error) {
	query := `{
				 q(func: type(` + MetaDType + `)){
					uid
					Meta.creationTime
					Meta.blockchainMode
					Meta.schemaVersion
				  }
				}`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Second*20, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r metaQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// SetCrawlerStatus sets the new crawler status
func SetCrawlerStatus(c external.Database, status CrawlerStatus) error {
	status.UID = uidV
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
	status.UID = uidV
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
	status.UID = uidV
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
	status.UID = uidV
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

// SetSchemaVersion sets the database schema version
func SetSchemaVersion(c external.Database, version uint64) error {
	return SetMeta(c, Meta{SchemaVersion: &version})
}

// SetMeta sets the database metadata
func SetMeta(c external.Database, meta Meta) error {
	meta.UID = uidV
	meta.SetDType()

	pb, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	req := &api.Request{
		Query:     "{q(func: type(" + MetaDType + ")){v as uid}}",
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	}

	return db.TxWithRetry(c, time.Minute*10, req)
}

// InitializeMeta sets the initial values of the database metadata.
// It should only be called when new database is set up.
func InitializeMeta(c external.Database, blockchainMode string) error {
	dbVersion := db.SchemaVersion
	return SetMeta(c, Meta{
		CreationTime:   time.Now().UTC().Format(time.RFC3339),
		BlockchainMode: blockchainMode,
		SchemaVersion:  &dbVersion,
	})
}
