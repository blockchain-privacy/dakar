package workspace

import (
	"backend/analytics"
	"backend/analytics/graph"
	"backend/db"
	"backend/testhelper"
	"testing"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	InitLogger()
	db.InitLogger()
	graph.InitLogger()
	analytics.InitLogger()

	testhelper.RunDgraphTests(m, &dbHandle.DB)
}
