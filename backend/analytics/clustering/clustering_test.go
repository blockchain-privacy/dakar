package clustering

import (
	"backend/testhelper"
	"testing"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	testhelper.RunDgraphTests(m, &dbHandle.DB)
}
