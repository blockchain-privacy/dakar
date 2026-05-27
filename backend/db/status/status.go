// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package status

import (
	"backend/db"
	"backend/external"
	"context"
	"encoding/json"
	"time"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

const uidV = "uid(v)"

// GetCrawlerStatus gets the crawler status from the database
func GetCrawlerStatus(ctx context.Context, c external.Database) (status CrawlerStatus, err error) {
	const query = `{
				 q(func: type(` + CrawlerStatusDType + `)){
					uid
					iscrawling
					lastblockid
				  }
				}`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return
	}

	var r crawlerStatusQuery

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	return r.payload()
}

// GetClassifierStatus gets the classifier status from the database
func GetClassifierStatus(ctx context.Context, c external.Database) (status ClassifierStatus, err error) {
	query := `{
				 q(func: type(` + ClassifierStatusDType + `)){
					uid
					isclassifying
					lastclassifiedid
				  }
				}`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return
	}

	var r classifierStatusQuery

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	return r.payload()
}

// GetClusteringFMIStatus gets the flat multi-input clustering status from the database
func GetClusteringFMIStatus(ctx context.Context, c external.Database) (status ClusteringFlatMultiInputStatus, err error) {
	query := `{
				 q(func: type(` + ClusteringFlatMultiInputDType + `)){
					uid
					isclustering
					lastclusteredid
				  }
				}`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return
	}

	var r clusteringFMIStatusQuery

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	return r.payload()
}

// GetHighestBlockID gets the highest block id.
func GetHighestBlockID(ctx context.Context, c external.Database) (int64, error) {
	query := `{
				var(func: has(id))@filter(eq(dgraph.type, "Block")){
					ids as id
				}
				
				q(){max:max(val(ids))}
			   }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return 0, err
	}

	var r struct {
		TopBlock []struct {
			Max int64 `json:"max,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return 0, serror.New(err)
	}

	switch {
	case len(r.TopBlock) == 0:
		return 0, serror.New(errTopBlockNotFound)
	case len(r.TopBlock) > 1:
		return 0, serror.New(errInvalidNumber)
	case r.TopBlock[0].Max == 0:
		return 0, serror.New(errTopBlockNotFound)
	}

	return r.TopBlock[0].Max, nil
}

// GetFrontendStatus gets verbose status information from the database
func GetFrontendStatus(ctx context.Context, c external.Database) (status FrontendStatus, err error) {
	query := `{
				crawler(func: type(` + CrawlerStatusDType + `)){
					iscrawling
					lastblockid
				}
				classifier(func: type(` + ClassifierStatusDType + `)){
					isclassifying
					lastclassifiedid
				}
				fmi(func: type(` + ClusteringFlatMultiInputDType + `)){
					isclustering
					lastclusteredid
				}
			}`

	resp, err := c.Query(ctx, query, nil)
	if err != nil {
		err = serror.New(err)
		return
	}

	var r struct {
		Crawler    []CrawlerStatus                  `json:"crawler,omitempty"`
		Classifier []ClassifierStatus               `json:"classifier,omitempty"`
		FMI        []ClusteringFlatMultiInputStatus `json:"fmi,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Crawler) == 1 {
		status.IsCrawling = r.Crawler[0].IsCrawling
		status.LastBlockID = r.Crawler[0].LastBlockID
	}

	if len(r.Classifier) == 1 {
		status.IsClassifying = r.Classifier[0].IsClassifying
		status.LastClassifiedBlockID = r.Classifier[0].LastClassifiedBlockID
	}

	if len(r.FMI) == 1 {
		status.IsClusteringFMI = r.FMI[0].IsClustering
		status.LastClusteredFMIBlockID = r.FMI[0].LastClusteredBlockID
	}

	return
}

// GetMeta gets the database metadata
func GetMeta(ctx context.Context, c external.Database) (meta Meta, err error) {
	query := `{
				 q(func: type(` + MetaDType + `)){
					uid
					Meta.creationTime
					Meta.blockchainMode
					Meta.schemaVersion
				  }
				}`

	resp, err := db.QueryVarWithRetry(ctx, c, query, nil)
	if err != nil {
		return
	}

	var r metaQuery

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	return r.payload()
}

// SetCrawlerStatus sets the new crawler status
func SetCrawlerStatus(ctx context.Context, c external.Database, status CrawlerStatus) error {
	status.UID = uidV
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return serror.New(err)
	}

	return db.MutationWithRetry(ctx, c, &api.Request{
		Query:     "{q(func: type(" + CrawlerStatusDType + ")){v as uid}}",
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	})
}

// SetClassifierStatus sets the new classifier status
func SetClassifierStatus(ctx context.Context, c external.Database, status ClassifierStatus) error {
	status.UID = uidV
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return serror.New(err)
	}

	return db.MutationWithRetry(ctx, c, &api.Request{
		Query:     "{q(func:type(" + ClassifierStatusDType + ")){v as uid}}",
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	})
}

// SetClusteringFMIStatus sets the new flat multi-input clustering status
func SetClusteringFMIStatus(ctx context.Context, c external.Database, status ClusteringFlatMultiInputStatus) error {
	status.UID = uidV
	status.SetDType()

	pb, err := json.Marshal(status)
	if err != nil {
		return serror.New(err)
	}

	return db.MutationWithRetry(ctx, c, &api.Request{
		Query:     "{q(func:type(" + ClusteringFlatMultiInputDType + ")){v as uid}}",
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	})
}

// SetCrawling sets the crawling status
func SetCrawling(ctx context.Context, c external.Database, crawling bool) error {
	return SetCrawlerStatus(ctx, c, CrawlerStatus{
		IsCrawling: &crawling,
	})
}

// SetClassifying sets the classifying status
func SetClassifying(ctx context.Context, c external.Database, classifying bool) error {
	return SetClassifierStatus(ctx, c, ClassifierStatus{
		IsClassifying: &classifying,
	})
}

// SetClusteringFMI sets the flat multi-input clustering status
func SetClusteringFMI(ctx context.Context, c external.Database, clustering bool) error {
	return SetClusteringFMIStatus(ctx, c, ClusteringFlatMultiInputStatus{
		IsClustering: &clustering,
	})
}

// SetLastBlockID sets the last block id
func SetLastBlockID(ctx context.Context, c external.Database, id int64) error {
	return SetCrawlerStatus(ctx, c, CrawlerStatus{
		LastBlockID: &id,
	})
}

// SetLastClassifiedBlockID sets the last classified block id
func SetLastClassifiedBlockID(ctx context.Context, c external.Database, id int64) error {
	return SetClassifierStatus(ctx, c, ClassifierStatus{
		LastClassifiedBlockID: &id,
	})
}

// SetLastClusteredFMIBlockID sets the last clustered multi-input block id
func SetLastClusteredFMIBlockID(ctx context.Context, c external.Database, id int64) error {
	return SetClusteringFMIStatus(ctx, c, ClusteringFlatMultiInputStatus{
		LastClusteredBlockID: &id,
	})
}

// SetSchemaVersion sets the database schema version
func SetSchemaVersion(ctx context.Context, c external.Database, version int) error {
	return SetMeta(ctx, c, Meta{SchemaVersion: &version})
}

// SetMeta sets the database metadata
func SetMeta(ctx context.Context, c external.Database, meta Meta) error {
	if meta.SchemaVersion != nil && *meta.SchemaVersion < 0 {
		return serror.FromStr("negative schema version")
	}

	meta.UID = uidV
	meta.SetDType()

	pb, err := json.Marshal(meta)
	if err != nil {
		return serror.New(err)
	}

	return db.MutationWithRetry(ctx, c, &api.Request{
		Query:     "{q(func: type(" + MetaDType + ")){v as uid}}",
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	})
}

// InitializeMeta sets the initial values of the database metadata.
// It should only be called when new database is set up.
func InitializeMeta(ctx context.Context, c external.Database, blockchainMode string) error {
	return SetMeta(ctx, c, Meta{
		CreationTime:   time.Now().UTC().Format(time.RFC3339),
		BlockchainMode: blockchainMode,
		SchemaVersion:  new(db.SchemaVersion),
	})
}
